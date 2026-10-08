package com.miracle.names.data.remote

import com.miracle.names.BuildConfig
import com.miracle.names.data.model.NameSearchRequest
import com.miracle.names.data.model.NameSearchResponse
import com.miracle.names.data.model.NameRerankRequest
import com.miracle.names.data.model.EntitlementResponse
import com.miracle.names.data.model.StorePurchaseRequest
import com.miracle.names.data.security.EntitlementTokenProvider
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Response
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import retrofit2.http.Body
import retrofit2.http.POST
import java.util.concurrent.TimeUnit

import com.miracle.names.data.model.CelebritiesResponse
import retrofit2.http.GET
import retrofit2.http.Header
import retrofit2.http.Multipart
import retrofit2.http.Part
import com.miracle.names.data.model.AuspiciousActivitiesResponse
import com.miracle.names.data.model.AuspiciousSearchRequest
import com.miracle.names.data.model.AuspiciousSearchResponse
import com.miracle.names.data.model.TaksaCalculationResponse
import retrofit2.http.Query

interface ApiService {
    @GET("api/v1/auspicious-activities")
    suspend fun getAuspiciousActivities(): Response<AuspiciousActivitiesResponse>

    @POST("api/v1/auspicious-search")
    suspend fun searchAuspiciousDates(
        @Body request: AuspiciousSearchRequest
    ): Response<AuspiciousSearchResponse>

    @GET("api/v1/auspicious/taksa")
    suspend fun getTaksaCalculation(
        @Query("birth_date") birthDate: String = "",
        @Query("birth_time") birthTime: String = "",
        @Query("target_date") targetDate: String = ""
    ): Response<TaksaCalculationResponse>

    @POST("api/v1/names/search")
    suspend fun searchTopNames(@Body request: NameSearchRequest): Response<NameSearchResponse>

    @GET("api/v1/names/suggestions")
    suspend fun suggestNames(
        @Query("query") query: String,
        @Query("limit") limit: Int = 8
    ): Response<com.miracle.names.data.model.NameSuggestionsResponse>

    @POST("api/v1/names/rerank")
    suspend fun rerankNames(@Body request: NameRerankRequest): Response<NameSearchResponse>

    @POST("api/v1/entitlements/verify")
    suspend fun verifyPurchase(@Body request: StorePurchaseRequest): Response<EntitlementResponse>
    
    @GET("api/v1/celebrities")
    suspend fun getCelebrities(): Response<CelebritiesResponse>

    @GET("api/v1/articles")
    suspend fun getArticles(): Response<com.miracle.names.data.model.ArticleListResponse>

    @GET("api/v1/decode")
    suspend fun decodeName(
        @retrofit2.http.Query("name") name: String,
        @retrofit2.http.Query("day_of_birth") dayOfBirth: String = ""
    ): Response<com.miracle.names.data.model.DecodeResponse>

    @GET("api/v1/pair-detail")
    suspend fun getPairDetail(@retrofit2.http.Query("number") pairNumber: String): Response<com.miracle.names.data.model.PairDetailResponse>

    @GET("api/v1/meanings/random")
    suspend fun getRandomMeanings(
        @retrofit2.http.Query("limit") limit: Int = 108
    ): Response<com.miracle.names.data.model.RandomMeaningsResponse>

    @Multipart
    @POST("api/v1/chat/conversations")
    suspend fun createChatConversation(
        @Part("customer_name") customerName: okhttp3.RequestBody,
        @Part("customer_avatar") customerAvatar: okhttp3.RequestBody
    ): Response<com.miracle.names.data.model.ChatConversationResponse>

    @GET("api/v1/chat/conversations/{id}/unread")
    suspend fun getChatUnreadCount(
        @retrofit2.http.Path("id") conversationId: String,
        @Header("X-Chat-Token") customerToken: String
    ): Response<com.miracle.names.data.model.ChatUnreadResponse>

    @GET("api/v1/chat/conversations/{id}/messages")
    suspend fun getChatMessages(
        @retrofit2.http.Path("id") conversationId: String,
        @Header("X-Chat-Token") customerToken: String
    ): Response<com.miracle.names.data.model.ChatMessagesResponse>

    @retrofit2.http.DELETE("api/v1/chat/conversations/{id}")
    suspend fun deleteChatConversation(
        @retrofit2.http.Path("id") conversationId: String,
        @Header("X-Chat-Token") customerToken: String
    ): Response<Unit>

    @Multipart
    @POST("api/v1/chat/conversations/{id}/messages")
    suspend fun sendChatMessage(
        @retrofit2.http.Path("id") conversationId: String,
        @Header("X-Chat-Token") customerToken: String,
        @Part("body") body: okhttp3.RequestBody,
        @Part image: okhttp3.MultipartBody.Part?
    ): Response<Unit>

    @retrofit2.http.POST("api/v1/chat/conversations/{id}/device-token")
    suspend fun registerChatDeviceToken(
        @retrofit2.http.Path("id") conversationId: String,
        @Header("X-Chat-Token") customerToken: String,
        @retrofit2.http.Body request: com.miracle.names.data.model.RegisterChatDeviceTokenRequest
    ): Response<Unit>
}

object MiracleApiConfig {
    val BaseUrl: String = BuildConfig.API_BASE_URL
    const val DomainAssetUrl = "https://xn--b3cu8e7ah6h.com/"

    fun assetUrl(path: String): String {
        if (path.isBlank()) return ""
        if (path.startsWith("http://") || path.startsWith("https://")) {
            return path
        }
        return DomainAssetUrl + path.trimStart('/')
    }
}

object RetrofitClient {
    private val loggingInterceptor = HttpLoggingInterceptor().apply {
        level = if (BuildConfig.DEBUG) {
            // Purchase tokens, entitlement tokens, and chat data share this
            // client. Request/response bodies must not be copied into Logcat.
            HttpLoggingInterceptor.Level.BASIC
        } else {
            HttpLoggingInterceptor.Level.NONE
        }
    }

    private fun createOkHttpClient(): OkHttpClient {
        // Use Android's platform trust store and hostname verification. Never
        // install a trust-all manager here: chat tokens and premium sessions
        // travel through this shared client.
        return OkHttpClient.Builder()
            .addInterceptor { chain ->
                val request = chain.request().newBuilder().apply {
                    EntitlementTokenProvider.token.takeIf { it.isNotBlank() }?.let {
                        header("X-Entitlement-Token", it)
                    }
                }.build()
                chain.proceed(request)
            }
            .addInterceptor(loggingInterceptor)
            .connectTimeout(15, TimeUnit.SECONDS)
            .readTimeout(60, TimeUnit.SECONDS)
            .callTimeout(60, TimeUnit.SECONDS)
            .build()
    }

    val okHttpClient: OkHttpClient by lazy {
        createOkHttpClient()
    }

    val apiService: ApiService by lazy {
        Retrofit.Builder()
            .baseUrl(MiracleApiConfig.BaseUrl)
            .client(okHttpClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
            .create(ApiService::class.java)
    }
}
