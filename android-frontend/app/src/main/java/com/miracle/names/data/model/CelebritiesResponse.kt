package com.miracle.names.data.model

data class Celebrity(
    val name: String,
    val avatar_url: String?
)

data class CelebritiesResponse(
    val error: Boolean,
    val data: List<Celebrity>
)
