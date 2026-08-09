package com.miracle.names.data.model

import com.google.gson.annotations.SerializedName

data class NameSearchRequest(
    @SerializedName("mode") val mode: String = "standalone",
    @SerializedName("surname") val surname: String = "",
    @SerializedName("day_of_birth") val dayOfBirth: String = "",
    @SerializedName("wish_query") val wishQuery: String = "",
    @SerializedName("limit") val limit: Int = 100,
    @SerializedName("numerology_weight") val numerologyWeight: Float? = null,
    @SerializedName("semantic_weight") val semanticWeight: Float? = null,
    @SerializedName("sound_weight") val soundWeight: Float? = null,
    @SerializedName("is_name_intent") val isNameIntent: Boolean = false
)

data class RankingWeightsPayload(
    @SerializedName("numerology") val numerology: Float = 0f,
    @SerializedName("semantic") val semantic: Float = 0f,
    @SerializedName("sound") val sound: Float = 0f
)

data class NameSearchPolicy(
    @SerializedName("pair_scope") val pairScope: String = "",
    @SerializedName("kalakini_scope") val kalakiniScope: String = "first_name",
    @SerializedName("sound_scope") val soundScope: String = "first_name",
    @SerializedName("numerology_rule_version") val numerologyRuleVersion: String = "",
    @SerializedName("ranking_precision_version") val rankingPrecisionVersion: String = "",
    @SerializedName("sound_formula_version") val soundFormulaVersion: String = "",
    @SerializedName("sound_calibration_source") val soundCalibrationSource: String = "",
    @SerializedName("sound_calibration_reference_version") val soundCalibrationReferenceVersion: String = "",
    @SerializedName("sound_calibration_approved") val soundCalibrationApproved: Boolean = false,
    @SerializedName("sound_calibration_audit_at") val soundCalibrationAuditAt: String = "",
    @SerializedName("sound_profile_loaded_at") val soundProfileLoadedAt: String = "",
    @SerializedName("sound_evaluation_version") val soundEvaluationVersion: String = "",
    @SerializedName("sound_corpus_fingerprint_sha256") val soundCorpusFingerprintSha256: String = "",
    @SerializedName("corpus_filter_version") val corpusFilterVersion: String = "",
    @SerializedName("semantic_query_model") val semanticQueryModel: String = "",
    @SerializedName("semantic_input_template_version") val semanticInputTemplateVersion: String = "",
    @SerializedName("semantic_vector_dimensions") val semanticVectorDimensions: Int = 0,
    @SerializedName("semantic_candidate_corpus_approved") val semanticCandidateCorpusApproved: Boolean = false,
    @SerializedName("semantic_candidate_corpus_source") val semanticCandidateCorpusSource: String = "",
    @SerializedName("semantic_candidate_corpus_sha256") val semanticCandidateCorpusSha256: String = "",
    @SerializedName("semantic_candidate_membership_sha256") val semanticCandidateMembershipSha256: String = "",
    @SerializedName("semantic_candidate_corpus_audit_at") val semanticCandidateCorpusAuditAt: String = "",
    @SerializedName("semantic_evaluation_version") val semanticEvaluationVersion: String = ""
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
    @SerializedName("semantic_available") val semanticAvailable: Boolean = false,
    @SerializedName("semantic_source") val semanticSource: String = "",
    @SerializedName("sound_score") val soundScore: Float = 0f,
    @SerializedName("raw_sound_score") val rawSoundScore: Float = 0f,
    @SerializedName("sound_scope") val soundScope: String = "first_name",
    @SerializedName("phonetic_version") val phoneticVersion: String = "",
    @SerializedName("sound_calibration") val soundCalibration: String = "",
    @SerializedName("pronunciation_ease") val pronunciationEase: Int = 0,
    @SerializedName("euphony_score") val euphonyScore: Int = 0,
    @SerializedName("rhythm_score") val rhythmScore: Int = 0,
    @SerializedName("final_score") val finalScore: Float = 0f,
    @SerializedName("ranking_reasons") val rankingReasons: List<String> = emptyList(),
    @SerializedName("is_perfect_combined") val isPerfectCombined: Boolean = false,
    @SerializedName("vector_distance") val vectorDistance: Float = 0f,
    @SerializedName("rank") val rank: Int = 0
)

data class NameSearchResponse(
    @SerializedName("error") val error: Boolean = false,
    @SerializedName("mode") val mode: String = "",
    @SerializedName("applied_mode") val appliedMode: String = "",
    @SerializedName("count") val count: Int = 0,
    @SerializedName("returned_count") val returnedCount: Int = 0,
    @SerializedName("data") val data: List<NameResultItem> = emptyList(),
    @SerializedName("candidate_token") val candidateToken: String = "",
    @SerializedName("eligible_candidate_count") val eligibleCandidateCount: Int = 0,
    @SerializedName("requested_weights") val requestedWeights: RankingWeightsPayload? = null,
    @SerializedName("applied_weights") val appliedWeights: RankingWeightsPayload? = null,
    @SerializedName("semantic_available") val semanticAvailable: Boolean = false,
    @SerializedName("semantic_source") val semanticSource: String = "",
    @SerializedName("semantic_query_model") val semanticQueryModel: String = "",
    @SerializedName("semantic_model") val legacySemanticModel: String = "",
    @SerializedName("semantic_vector_hash") val semanticVectorHash: String = "",
    @SerializedName("candidate_vector_provenance_verified")
    val candidateVectorProvenanceVerified: Boolean = false,
    @SerializedName("candidate_vector_provenance_source")
    val candidateVectorProvenanceSource: String = "",
    @SerializedName("policy") val policy: NameSearchPolicy = NameSearchPolicy(),
    @SerializedName("message") val message: String = ""
) {
    /** Prefer the explicit query-model contract while accepting older servers. */
    fun resolvedSemanticQueryModel(): String =
        semanticQueryModel.ifBlank { legacySemanticModel }
}

data class NameRerankRequest(
    @SerializedName("candidate_token") val candidateToken: String,
    @SerializedName("numerology_weight") val numerologyWeight: Float,
    @SerializedName("semantic_weight") val semanticWeight: Float,
    @SerializedName("sound_weight") val soundWeight: Float,
    @SerializedName("limit") val limit: Int = 100
)

data class DecodeResponse(
    @SerializedName("error") val error: Boolean = false,
    @SerializedName("pairs") val pairs: List<String>? = emptyList(),
    @SerializedName("pair_types") val pairTypes: List<String>? = emptyList(),
    @SerializedName("pair_points") val pairPoints: List<Int>? = emptyList(),
    @SerializedName("pair_is_good") val pairIsGood: List<Boolean>? = emptyList(),
    @SerializedName("meaning") val meaning: String? = null,
    @SerializedName("message") val message: String = ""
)

data class PairDetailResponse(
    @SerializedName("error") val error: Boolean = false,
    @SerializedName("pair_number") val pairNumber: String = "",
    @SerializedName("pair_type") val pairType: String = "",
    @SerializedName("pair_point") val pairPoint: Int = 0,
    @SerializedName("miracle_desc") val miracleDesc: String = "",
    @SerializedName("miracle_detail") val miracleDetail: String = "",
    @SerializedName("is_good") val isGood: Boolean = true,
    @SerializedName("message") val message: String? = null
)
