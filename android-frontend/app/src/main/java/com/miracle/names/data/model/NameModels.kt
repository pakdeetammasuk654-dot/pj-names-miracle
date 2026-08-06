package com.miracle.names.data.model

import com.google.gson.annotations.SerializedName

data class NameSearchRequest(
    @SerializedName("mode") val mode: String = "standalone",
    @SerializedName("surname") val surname: String = "",
    @SerializedName("day_of_birth") val dayOfBirth: String = "",
    @SerializedName("wish_query") val wishQuery: String = "",
    @SerializedName("limit") val limit: Int = 100,
    @SerializedName("semantic_weight") val semanticWeight: Float? = null,
    @SerializedName("is_name_intent") val isNameIntent: Boolean = false
)

data class NameResultItem(
    @SerializedName("name_id") val nameId: Int = 0,
    @SerializedName("thname") val thName: String = "",
    @SerializedName("meaning") val meaning: String = "",
    @SerializedName("sat_sum") val satSum: Int = 0,
    @SerializedName("sha_sum") val shaSum: Int = 0,
    @SerializedName("total_sat_sum") val totalSatSum: Int = 0,
    @SerializedName("total_sha_sum") val totalShaSum: Int = 0,
    @SerializedName("sat_pairs_display") val satPairsDisplay: String = "",
    @SerializedName("sha_pairs_display") val shaPairsDisplay: String = "",
    @SerializedName("sat_pairtype") val satPairType: String = "",
    @SerializedName("sat_miracledesc") val satMiracleDesc: String = "",
    @SerializedName("sha_pairtype") val shaPairType: String = "",
    @SerializedName("sha_miracledesc") val shaMiracleDesc: String = "",
    @SerializedName("sat_grade_display") val satGradeDisplay: String = "",
    @SerializedName("total_score") val totalScore: Int = 0,
    @SerializedName("numerology_score") val numerologyScore: Float = 0f,
    @SerializedName("semantic_score") val semanticScore: Float = 0f,
    @SerializedName("perfect_bonus") val perfectBonus: Float = 0f,
    @SerializedName("final_score") val finalScore: Float = 0f,
    @SerializedName("ranking_reasons") val rankingReasons: List<String> = emptyList(),
    @SerializedName("is_perfect_combined") val isPerfectCombined: Boolean = false,
    @SerializedName("vector_distance") val vectorDistance: Float = 0f,
    @SerializedName("rank") val rank: Int = 0
)

data class NameSearchResponse(
    @SerializedName("error") val error: Boolean = false,
    @SerializedName("mode") val mode: String = "",
    @SerializedName("count") val count: Int = 0,
    @SerializedName("data") val data: List<NameResultItem> = emptyList(),
    @SerializedName("message") val message: String = ""
)

data class DecodeResponse(
    @SerializedName("error") val error: Boolean = false,
    @SerializedName("pairs") val pairs: List<String>? = emptyList(),
    @SerializedName("pair_types") val pairTypes: List<String>? = emptyList(),
    @SerializedName("meaning") val meaning: String? = null
)
