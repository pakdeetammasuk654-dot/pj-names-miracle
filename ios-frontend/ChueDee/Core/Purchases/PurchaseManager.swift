import Foundation
import StoreKit
import SwiftUI

@MainActor
final class PurchaseManager: ObservableObject {
    static let shared = PurchaseManager()
    static let productID = "premium_lifetime"

    @Published private(set) var isUnlocked = false
    @Published private(set) var product: Product?
    @Published private(set) var isBusy = false
    @Published var message: String?

#if DEBUG
    @Published var mockPurchaseEnabled = false

    func setMockPurchase(enabled: Bool) async {
        mockPurchaseEnabled = enabled
        if enabled {
            do {
                let entitlement = try await APIClient.shared.verifyPurchase(transactionID: "debug-mock-12345")
                KeychainHelper.shared.save(key: "premium_entitlement_token", value: entitlement.entitlementToken)
                isUnlocked = true
            } catch {
                isUnlocked = true
            }
        } else {
            KeychainHelper.shared.delete(key: "premium_entitlement_token")
            isUnlocked = false
        }
    }
#endif

    var effectiveIsUnlocked: Bool {
#if DEBUG
        isUnlocked || mockPurchaseEnabled
#else
        isUnlocked
#endif
    }

    private var updatesTask: Task<Void, Never>?

    private init() {
        updatesTask = observeTransactions()
        Task { await prepare() }
    }

    deinit { updatesTask?.cancel() }

    var displayPrice: String { product?.displayPrice ?? "ดูราคาใน App Store" }

    func prepare() async {
        await refreshEntitlement()
        do {
            let products = try await Product.products(for: [Self.productID])
            product = products.first(where: { $0.id == Self.productID })
        } catch {
            message = "ยังโหลดข้อมูลการซื้อไม่ได้ กรุณาลองใหม่"
        }
    }

    func purchase() async {
        guard !isBusy else { return }
        guard !effectiveIsUnlocked else {
            message = "คุณได้ปลดล็อกสิทธิ์พรีเมียมเรียบร้อยแล้ว"
            return
        }
        isBusy = true
        defer { isBusy = false }
        do {
            if product == nil {
                let products = try await Product.products(for: [Self.productID])
                product = products.first(where: { $0.id == Self.productID })
            }
            guard let product else {
                message = "ไม่พบสินค้าใน App Store Connect"
                return
            }
            switch try await product.purchase() {
            case .success(let verification):
                let transaction = try verified(verification)
                if await verifyWithServer(transaction) {
                    await transaction.finish()
                }
            case .pending:
                message = "รายการซื้อกำลังรอการอนุมัติ"
            case .userCancelled:
                break
            @unknown default:
                break
            }
        } catch {
            message = "ซื้อไม่สำเร็จ กรุณาลองใหม่"
        }
    }

    func restore() async {
        guard !isBusy else { return }
        isBusy = true
        defer { isBusy = false }
        do {
            try await AppStore.sync()
            await refreshEntitlement()
            message = isUnlocked ? "คืนสิทธิ์การซื้อเรียบร้อยแล้ว" : "ไม่พบรายการซื้อเดิม"
        } catch {
            message = "คืนสิทธิ์การซื้อไม่สำเร็จ"
        }
    }

    private func refreshEntitlement() async {
        var activeTransaction: StoreKit.Transaction?
        for await result in Transaction.currentEntitlements {
            guard let transaction = try? verified(result) else { continue }
            if transaction.productID == Self.productID,
               transaction.revocationDate == nil {
                activeTransaction = transaction
                break
            }
        }
        guard let activeTransaction else {
            KeychainHelper.shared.delete(key: "premium_entitlement_token")
            isUnlocked = false
            return
        }
        _ = await verifyWithServer(activeTransaction)
    }

    private func verifyWithServer(_ transaction: StoreKit.Transaction) async -> Bool {
        do {
            let entitlement = try await APIClient.shared.verifyPurchase(transactionID: String(transaction.id))
            guard entitlement.productID == Self.productID, !entitlement.entitlementToken.isEmpty else {
                throw PurchaseError.failedVerification
            }
            KeychainHelper.shared.save(key: "premium_entitlement_token", value: entitlement.entitlementToken)
            isUnlocked = true
            message = "ปลดล็อกรายชื่อมงคลทั้งหมดเรียบร้อยแล้ว"
            return true
        } catch {
            KeychainHelper.shared.delete(key: "premium_entitlement_token")
            isUnlocked = false
            message = "ยังตรวจสอบสิทธิ์กับ App Store ไม่ได้ กรุณาลองอีกครั้ง"
            return false
        }
    }

    private func observeTransactions() -> Task<Void, Never> {
        Task { [weak self] in
            for await result in Transaction.updates {
                guard let self, let transaction = try? self.verified(result) else { continue }
                if await self.verifyWithServer(transaction) {
                    await transaction.finish()
                }
            }
        }
    }

    nonisolated private func verified<T>(_ result: VerificationResult<T>) throws -> T {
        switch result {
        case .verified(let value): return value
        case .unverified: throw PurchaseError.failedVerification
        }
    }
}

private enum PurchaseError: Error { case failedVerification }

#if DEBUG
struct DebugPurchaseView: View {
    @Environment(\.dismiss) private var dismiss
    @ObservedObject var purchaseManager: PurchaseManager

    var body: some View {
        NavigationStack {
            Form {
                Section("สถานะการซื้อจำลอง") {
                    Toggle("Mock purchase สำเร็จ", isOn: Binding(
                        get: { purchaseManager.mockPurchaseEnabled },
                        set: { val in
                            Task { await purchaseManager.setMockPurchase(enabled: val) }
                        }
                    ))
                    Text(purchaseManager.mockPurchaseEnabled
                         ? "แอปจะแสดงชื่อเต็มโดยไม่เรียก StoreKit หรือ Google Play"
                         : "แอปจะแสดงชื่อแบบล็อกตามสถานะการซื้อจริง")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                }
                Section {
                    Button(role: .destructive) {
                        Task {
                            await purchaseManager.setMockPurchase(enabled: false)
                            dismiss()
                        }
                    } label: {
                        Text("RESET: ปิด Mock และกลับไปใช้สถานะจริง")
                            .frame(maxWidth: .infinity, alignment: .center)
                    }
                }
            }
            .navigationTitle("Debug purchase")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("ปิด") {
                        dismiss()
                    }
                }
            }
        }
    }
}
#endif
