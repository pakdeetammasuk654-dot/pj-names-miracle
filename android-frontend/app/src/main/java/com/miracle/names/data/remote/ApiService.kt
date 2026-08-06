package com.miracle.names.data.remote

import com.miracle.names.data.model.NameSearchRequest
import com.miracle.names.data.model.NameSearchResponse
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

interface ApiService {
    @POST("api/v1/names/search")
    suspend fun searchTopNames(@Body request: NameSearchRequest): Response<NameSearchResponse>
    
    @GET("api/v1/celebrities")
    suspend fun getCelebrities(): Response<CelebritiesResponse>

    @GET("api/v1/decode")
    suspend fun decodeName(@retrofit2.http.Query("name") name: String): Response<com.miracle.names.data.model.DecodeResponse>
}

object RetrofitClient {
    // Default emulator loopback IP (Change to local server IP if testing on physical device)
    private const val BASE_URL = "http://10.0.2.2:8080/"

    private val loggingInterceptor = HttpLoggingInterceptor().apply {
        level = HttpLoggingInterceptor.Level.BODY
    }

    private val okHttpClient = OkHttpClient.Builder()
        .addInterceptor(loggingInterceptor)
        .connectTimeout(15, TimeUnit.SECONDS)
        .readTimeout(15, TimeUnit.SECONDS)
        .build()

    val apiService: ApiService by lazy {
        Retrofit.Builder()
            .baseUrl(BASE_URL)
            .client(okHttpClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
            .create(ApiService::class.java)
    }
}
