package com.miracle.names.viewmodel

import com.miracle.names.data.model.NameSearchPolicy
import com.miracle.names.data.model.RankingWeightsPayload
import kotlin.math.abs

data class RankingWeights(
    val numerology: Float,
    val semantic: Float,
    val sound: Float
) {
    fun approximatelyEquals(other: RankingWeights, tolerance: Float = 0.0001f): Boolean {
        return abs(numerology - other.numerology) <= tolerance &&
            abs(semantic - other.semantic) <= tolerance &&
            abs(sound - other.sound) <= tolerance
    }

    companion object {
        val Default = RankingWeights(
            numerology = 0.45f,
            semantic = 0.35f,
            sound = 0.20f
        )
    }
}

data class AppliedSearchSnapshot(
    val mode: String,
    val surname: String,
    val isFindSurnameMode: Boolean = false,
    val dayOfBirth: String,
    val wishQuery: String,
    /** The normalized formula the backend actually applied to FinalScore. */
    val weights: RankingWeights,
    val count: Int,
    /** The user's selected formula before backend availability redistribution. */
    val requestedWeights: RankingWeights = weights,
    val eligibleCandidateCount: Int = count,
    val semanticAvailable: Boolean = false,
    val semanticSource: String = "",
    /** Query embedding model; parsed from semantic_query_model or its legacy alias. */
    val semanticModel: String = "",
    val semanticVectorHash: String = "",
    val candidateVectorProvenanceVerified: Boolean = false,
    val candidateVectorProvenanceSource: String = "",
    val policy: NameSearchPolicy = NameSearchPolicy(),
    val appliedWeightsConfirmed: Boolean = false
) {
    val isCombined: Boolean
        get() = mode == MODE_COMBINED

    fun fullName(firstName: String): String {
        return if (isCombined && surname.isNotBlank()) {
            if (isFindSurnameMode) "$surname $firstName" else "$firstName $surname"
        } else firstName
    }

    val hasApprovedSemanticProof: Boolean
        get() = semanticAvailable &&
            candidateVectorProvenanceVerified &&
            policy.semanticCandidateCorpusApproved

    val hasOwnerConfirmedSemanticProof: Boolean
        get() = semanticAvailable &&
            candidateVectorProvenanceVerified &&
            (
                candidateVectorProvenanceSource == "owner_confirmed_legacy_embedding" ||
                    policy.semanticCandidateCorpusSource == "owner_confirmed_legacy_embedding"
                )

    fun semanticStatusThai(): String {
        if (semanticSource == "not_requested") {
            return "ไม่ได้ใช้คะแนนความหมายในการจัดอันดับครั้งนี้"
        }
        if (semanticSource == "trigram" || candidateVectorProvenanceSource == "not_applicable_trigram") {
            return "ใช้ความคล้ายชื่อแบบข้อความ (ไม่ใช้เวกเตอร์คลังชื่อ)"
        }
        if (hasApprovedSemanticProof) {
            return "คะแนนความหมายผ่านการยืนยัน snapshot และที่มาเวกเตอร์"
        }
        if (hasOwnerConfirmedSemanticProof) {
            return "ใช้คะแนนความหมายจากเวกเตอร์ที่เจ้าของข้อมูลยืนยัน"
        }
        if (semanticAvailable) {
            return "หลักฐานคะแนนความหมายยังไม่ผ่านการอนุมัติหรือมาจาก API รุ่นเดิม"
        }
        return when {
            semanticSource.contains("stale_vector_corpus") ||
                policy.semanticCandidateCorpusSource.contains("stale_vector_corpus") ->
                "ระบบปิดคะแนนความหมาย: snapshot เวกเตอร์คลังชื่อล้าสมัย"
            semanticSource.contains("query_embedding_error") ->
                "ระบบปิดคะแนนความหมาย: สร้างเวกเตอร์คำค้นไม่สำเร็จ"
            semanticSource.contains("embedding_missing") ->
                "ระบบปิดคะแนนความหมาย: เวกเตอร์รายชื่อไม่ครบ"
            !policy.semanticCandidateCorpusApproved ||
                semanticSource.contains("unapproved_vector_corpus") ->
                "ระบบปิดคะแนนความหมาย: ยังไม่มี snapshot เวกเตอร์คลังชื่อที่อนุมัติ"
            !candidateVectorProvenanceVerified ->
                "ระบบปิดคะแนนความหมาย: ยืนยันที่มาเวกเตอร์รายชื่อไม่ได้"
            else -> "ระบบปิดคะแนนความหมาย: ไม่มีคะแนนที่ตรวจสอบได้"
        }
    }

    fun semanticProofThai(): String {
        if (!hasApprovedSemanticProof && !hasOwnerConfirmedSemanticProof) return ""
        val queryModel = semanticModel.ifBlank { policy.semanticQueryModel }
        val provenanceSource = candidateVectorProvenanceSource
            .ifBlank { policy.semanticCandidateCorpusSource }
        return buildString {
            if (provenanceSource.isNotBlank()) append(provenanceSource)
            if (queryModel.isNotBlank()) {
                if (isNotEmpty()) append(" · ")
                append(queryModel)
            }
            if (policy.semanticInputTemplateVersion.isNotBlank()) {
                if (isNotEmpty()) append(" · ")
                append("template ${policy.semanticInputTemplateVersion}")
            }
            if (policy.semanticVectorDimensions > 0) {
                if (isNotEmpty()) append(" · ")
                append("${policy.semanticVectorDimensions} มิติ")
            }
            if (policy.semanticCandidateCorpusSha256.isNotBlank()) {
                if (isNotEmpty()) append(" · ")
                append("คลัง SHA-256 ${policy.semanticCandidateCorpusSha256.take(12)}…")
            }
            if (semanticVectorHash.isNotBlank()) {
                if (isNotEmpty()) append(" · ")
                append("คำค้น ${semanticVectorHash.take(12)}…")
            }
            if (policy.semanticEvaluationVersion.isNotBlank()) {
                if (isNotEmpty()) append(" · ")
                append("eval ${policy.semanticEvaluationVersion}")
            }
            if (policy.semanticCandidateCorpusAuditAt.isNotBlank()) {
                if (isNotEmpty()) append(" · ")
                append("audit ${policy.semanticCandidateCorpusAuditAt}")
            }
        }
    }
}

internal const val MODE_STANDALONE = "standalone"
internal const val MODE_COMBINED = "combined"

internal fun normalizeRankingWeights(
    numerology: Float,
    semantic: Float,
    sound: Float
): RankingWeights {
    val safeNumerology = numerology.coerceIn(0f, 1f)
    val safeSemantic = semantic.coerceIn(0f, 1f)
    val safeSound = sound.coerceIn(0f, 1f)
    val total = safeNumerology + safeSemantic + safeSound
    if (total <= 0f) return RankingWeights.Default

    return RankingWeights(
        numerology = safeNumerology / total,
        semantic = safeSemantic / total,
        sound = safeSound / total
    )
}

internal fun resolveAppliedMode(responseMode: String, requestedMode: String): String {
    return when (responseMode) {
        MODE_STANDALONE, MODE_COMBINED -> responseMode
        else -> when (requestedMode) {
            MODE_COMBINED -> MODE_COMBINED
            else -> MODE_STANDALONE
        }
    }
}

internal fun resolveAppliedSurname(mode: String, surname: String): String {
    return if (mode == MODE_COMBINED) surname.trim() else ""
}

internal fun resolveActualCount(reportedCount: Int, dataSize: Int): Int {
    return if (reportedCount >= 0 && reportedCount == dataSize) reportedCount else dataSize
}

internal fun resolveResponseWeights(
    payload: RankingWeightsPayload?,
    fallback: RankingWeights
): RankingWeights {
    if (payload == null) return fallback
    val total = payload.numerology + payload.semantic + payload.sound
    if (!total.isFinite() || total <= 0f) return fallback
    return normalizeRankingWeights(payload.numerology, payload.semantic, payload.sound)
}

internal fun resolveEligibleCandidateCount(reportedCount: Int, resultCount: Int): Int {
    return reportedCount.takeIf { it >= resultCount && it > 0 } ?: resultCount
}

/**
 * Small request reducer used to ensure that only the newest asynchronous
 * operation may publish state. ViewModel calls are serialized on the main
 * dispatcher, so no additional synchronization is required here.
 */
internal class RequestRevisionGate {
    private var activeRevision = 0L

    fun begin(): Long {
        activeRevision += 1
        return activeRevision
    }

    fun invalidate() {
        activeRevision += 1
    }

    fun accepts(revision: Long): Boolean = revision == activeRevision
}
