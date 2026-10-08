package com.miracle.names.data.billing

import android.app.Activity
import android.content.Context
import android.util.Log
import com.miracle.names.BuildConfig
import com.android.billingclient.api.*
import com.miracle.names.data.model.StorePurchaseRequest
import com.miracle.names.data.remote.RetrofitClient
import com.miracle.names.data.security.EntitlementTokenProvider
import com.miracle.names.data.security.SecureEntitlementStore
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

class BillingManager(private val context: Context) : PurchasesUpdatedListener {

    companion object {
        private const val TAG = "BillingManager"
        const val PRODUCT_ID_UNLOCK = "premium_lifetime"
    }

    private val secureStore = SecureEntitlementStore(context)
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    private val _isUnlocked = MutableStateFlow(false)
    val isUnlocked: StateFlow<Boolean> = _isUnlocked.asStateFlow()

    private val _mockPurchaseEnabled = MutableStateFlow(false)
    val mockPurchaseEnabled: StateFlow<Boolean> = _mockPurchaseEnabled.asStateFlow()
    val effectiveIsUnlocked: StateFlow<Boolean> = combine(_isUnlocked, _mockPurchaseEnabled) { real, mock ->
        real || (BuildConfig.DEBUG && mock)
    }.stateIn(scope, SharingStarted.Eagerly, false)

    fun setMockPurchaseEnabled(enabled: Boolean) {
        if (BuildConfig.DEBUG) {
            _mockPurchaseEnabled.value = enabled
            if (enabled) {
                scope.launch {
                    runCatching {
                        RetrofitClient.apiService.verifyPurchase(
                            StorePurchaseRequest(
                                platform = "android",
                                productId = PRODUCT_ID_UNLOCK,
                                purchaseToken = "debug-mock-purchase-token"
                            )
                        )
                    }.onSuccess { response ->
                        val entitlement = response.body()
                        if (response.isSuccessful && entitlement?.entitlementToken?.isNotBlank() == true) {
                            secureStore.save(entitlement.entitlementToken)
                            setUnlockedState(true)
                            _billingMessage.value = "เปิดโหมดจำลองการซื้อ VIP สำเร็จ"
                        }
                    }
                }
            } else {
                secureStore.clear()
                setUnlockedState(false)
                _billingMessage.value = "ปิดโหมดจำลอง กลับสู่สถานะผู้ใช้ฟรี"
            }
        }
    }

    private val _productPrice = MutableStateFlow("ดูราคาใน Google Play")
    val productPrice: StateFlow<String> = _productPrice.asStateFlow()

    private val _billingMessage = MutableStateFlow<String?>(null)
    val billingMessage: StateFlow<String?> = _billingMessage.asStateFlow()

    private var billingClient: BillingClient = BillingClient.newBuilder(context)
        .setListener(this)
        .enablePendingPurchases(
            PendingPurchasesParams.newBuilder()
                .enableOneTimeProducts()
                .build()
        )
        .build()

    private var productDetails: ProductDetails? = null

    fun hasProductDetails(): Boolean = productDetails != null

    init {
        EntitlementTokenProvider.update(secureStore.load())
        _isUnlocked.value = EntitlementTokenProvider.token.isNotBlank()
        startConnection()
    }

    fun startConnection() {
        billingClient.startConnection(object : BillingClientStateListener {
            override fun onBillingSetupFinished(billingResult: BillingResult) {
                if (billingResult.responseCode == BillingClient.BillingResponseCode.OK) {
                    Log.d(TAG, "Billing Setup Finished Successfully")
                    queryProductDetails()
                    queryExistingPurchases()
                } else {
                    Log.e(TAG, "Billing Setup Failed: ${billingResult.debugMessage}")
                }
            }

            override fun onBillingServiceDisconnected() {
                Log.w(TAG, "Billing Service Disconnected. Reconnecting...")
            }
        })
    }

    private fun queryProductDetails() {
        val productList = listOf(
            QueryProductDetailsParams.Product.newBuilder()
                .setProductId(PRODUCT_ID_UNLOCK)
                .setProductType(BillingClient.ProductType.INAPP)
                .build()
        )

        val params = QueryProductDetailsParams.newBuilder()
            .setProductList(productList)
            .build()

        billingClient.queryProductDetailsAsync(params) { billingResult, queryResult ->
            if (billingResult.responseCode == BillingClient.BillingResponseCode.OK) {
                val productDetailsList = queryResult.productDetailsList
                if (!productDetailsList.isNullOrEmpty()) {
                    productDetails = productDetailsList.firstOrNull { it.productId == PRODUCT_ID_UNLOCK }
                    val price = productDetails?.oneTimePurchaseOfferDetails?.formattedPrice
                    if (!price.isNullOrEmpty()) {
                        _productPrice.value = price
                        Log.d(TAG, "Loaded Product Price from Play Store: $price")
                    }
                }
            } else {
                Log.e(TAG, "Query Product Details Failed: ${billingResult.debugMessage}")
            }
        }
    }

    fun queryExistingPurchases() {
        if (!billingClient.isReady) return

        val params = QueryPurchasesParams.newBuilder()
            .setProductType(BillingClient.ProductType.INAPP)
            .build()

        billingClient.queryPurchasesAsync(params) { billingResult, purchases ->
            if (billingResult.responseCode == BillingClient.BillingResponseCode.OK) {
                val owned = purchases.firstOrNull { purchase ->
                    purchase.products.contains(PRODUCT_ID_UNLOCK) &&
                        purchase.purchaseState == Purchase.PurchaseState.PURCHASED
                }
                if (owned != null) {
                    if (!owned.isAcknowledged) {
                        acknowledgePurchase(owned)
                    }
                    verifyWithServer(owned)
                } else {
                    setUnlockedState(false)
                }
            }
        }
    }

    fun launchPurchaseFlow(activity: Activity) {
        val details = productDetails
        if (details == null) {
            Log.w(TAG, "Product details not loaded yet. Retrying connection...")
            if (!billingClient.isReady) {
                startConnection()
            }
            _billingMessage.value = "กำลังเชื่อมต่อกับ Google Play Store กรุณาลองใหม่อีกครั้ง"
            return
        }

        val productDetailsParamsList = listOf(
            BillingFlowParams.ProductDetailsParams.newBuilder()
                .setProductDetails(details)
                .build()
        )

        val billingFlowParams = BillingFlowParams.newBuilder()
            .setProductDetailsParamsList(productDetailsParamsList)
            .build()

        val responseCode = billingClient.launchBillingFlow(activity, billingFlowParams).responseCode
        if (responseCode != BillingClient.BillingResponseCode.OK) {
            Log.e(TAG, "Launch Billing Flow Failed with code: $responseCode")
            _billingMessage.value = "ไม่สามารถเปิดหน้าชำระเงินได้ กรุณาลองใหม่"
        }
    }

    override fun onPurchasesUpdated(billingResult: BillingResult, purchases: MutableList<Purchase>?) {
        when (billingResult.responseCode) {
            BillingClient.BillingResponseCode.OK -> {
                if (!purchases.isNullOrEmpty()) {
                    for (purchase in purchases) {
                        handlePurchase(purchase)
                    }
                }
            }
            BillingClient.BillingResponseCode.USER_CANCELED -> {
                Log.d(TAG, "User canceled billing flow")
            }
            BillingClient.BillingResponseCode.ITEM_ALREADY_OWNED -> {
                Log.d(TAG, "Item already owned, restoring purchase state")
                queryExistingPurchases()
                _billingMessage.value = "กำลังตรวจสอบสิทธิ์จาก Google Play"
            }
            else -> {
                Log.e(TAG, "Purchases Updated Error: ${billingResult.debugMessage}")
                _billingMessage.value = "การชำระเงินไม่สำเร็จ กรุณาลองใหม่"
            }
        }
    }

    private fun handlePurchase(purchase: Purchase) {
        if (purchase.purchaseState == Purchase.PurchaseState.PURCHASED &&
            purchase.products.contains(PRODUCT_ID_UNLOCK)
        ) {
            verifyWithServer(purchase)
        }
    }

    private fun verifyWithServer(purchase: Purchase) {
        scope.launch {
            runCatching {
                RetrofitClient.apiService.verifyPurchase(
                    StorePurchaseRequest(
                        productId = PRODUCT_ID_UNLOCK,
                        purchaseToken = purchase.purchaseToken
                    )
                )
            }.onSuccess { response ->
                val entitlement = response.body()
                if (response.isSuccessful && entitlement?.entitlementToken?.isNotBlank() == true &&
                    entitlement.productId == PRODUCT_ID_UNLOCK
                ) {
                    secureStore.save(entitlement.entitlementToken)
                    setUnlockedState(true)
                    if (!purchase.isAcknowledged) acknowledgePurchase(purchase)
                    else _billingMessage.value = "ปลดล็อกรายชื่อมงคลทั้งหมดเรียบร้อยแล้ว!"
                } else {
                    if (secureStore.load().isBlank()) {
                        setUnlockedState(false)
                    }
                    _billingMessage.value = "ยังตรวจสอบสิทธิ์การซื้อไม่ได้ กรุณาลองกู้คืนการซื้ออีกครั้ง"
                }
            }.onFailure {
                Log.e(TAG, "Server purchase verification failed", it)
                if (!purchase.isAcknowledged) {
                    acknowledgePurchase(purchase)
                }
                if (secureStore.load().isBlank()) {
                    setUnlockedState(false)
                }
                _billingMessage.value = "เชื่อมต่อเพื่อตรวจสอบสิทธิ์ไม่ได้ กรุณาลองอีกครั้ง"
            }
        }
    }

    private fun acknowledgePurchase(purchase: Purchase) {
        val acknowledgePurchaseParams = AcknowledgePurchaseParams.newBuilder()
            .setPurchaseToken(purchase.purchaseToken)
            .build()

        billingClient.acknowledgePurchase(acknowledgePurchaseParams) { billingResult ->
            if (billingResult.responseCode == BillingClient.BillingResponseCode.OK) {
                Log.d(TAG, "Purchase Acknowledged Successfully")
                _billingMessage.value = "ปลดล็อกรายชื่อมงคลทั้งหมดเรียบร้อยแล้ว!"
            } else {
                Log.e(TAG, "Acknowledge Purchase Failed: ${billingResult.debugMessage}")
            }
        }
    }

    fun restorePurchases() {
        queryExistingPurchases()
        _billingMessage.value = "กำลังตรวจสอบสิทธิ์จาก Google Play"
    }

    fun clearBillingMessage() {
        _billingMessage.value = null
    }

    private fun setUnlockedState(unlocked: Boolean) {
        if (!unlocked) secureStore.clear()
        _isUnlocked.value = unlocked
    }

}
