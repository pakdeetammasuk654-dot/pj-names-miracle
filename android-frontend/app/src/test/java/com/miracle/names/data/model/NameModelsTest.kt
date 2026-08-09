package com.miracle.names.data.model

import com.google.gson.Gson
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class NameModelsTest {

    private val gson = Gson()

    @Test
    fun responseMetadata_deserializesAppliedWeightsAndScopes() {
        val response = gson.fromJson(
            """
            {
              "error": false,
              "mode": "combined",
              "applied_mode": "combined",
              "count": 1,
              "returned_count": 1,
              "eligible_candidate_count": 240,
              "requested_weights": {"numerology": 0.15, "semantic": 0.15, "sound": 0.70},
              "applied_weights": {"numerology": 0.20, "semantic": 0.00, "sound": 0.80},
              "semantic_available": false,
              "semantic_source": "unavailable",
              "semantic_model": "legacy-should-not-win",
              "semantic_query_model": "text-embedding-3-small",
              "semantic_vector_hash": "abc123",
              "candidate_vector_provenance_verified": true,
              "candidate_vector_provenance_source": "approved_snapshot",
              "policy": {
                "pair_scope": "first_name_and_combined",
                "kalakini_scope": "first_name",
                "sound_scope": "first_name",
                "numerology_rule_version": "thai_sum_v1_dictionary",
                "ranking_precision_version": "ranking_precision_12dp_v1",
                "sound_formula_version": "sound_45_35_20_v1",
                "sound_calibration_source": "live_repository_distribution_provisional",
                "sound_calibration_reference_version": "typhoon_v1_promptA",
                "sound_calibration_approved": false,
                "sound_profile_loaded_at": "2026-08-07T02:00:00Z",
                "sound_evaluation_version": "sound_eval_v1",
                "sound_corpus_fingerprint_sha256": "sound-corpus-hash",
                "corpus_filter_version": "title_prefix_v1_child_titles",
                "semantic_query_model": "text-embedding-3-small",
                "semantic_input_template_version": "wish_query_v1",
                "semantic_vector_dimensions": 1536,
                "semantic_candidate_corpus_approved": true,
                "semantic_candidate_corpus_source": "approved_snapshot",
                "semantic_candidate_corpus_sha256": "corpus-hash",
                "semantic_candidate_membership_sha256": "membership-hash",
                "semantic_candidate_corpus_audit_at": "2026-08-07T01:00:00Z",
                "semantic_evaluation_version": "semantic_eval_v1"
              },
              "data": [{
                "name_id": 7,
                "thname": "วรชัย",
                "rank": 1,
                "semantic_available": false,
                "sound_scope": "first_name",
                "raw_sound_score": 88.5,
                "phonetic_version": "v2",
                "sound_calibration": "reference-v1"
              }]
            }
            """.trimIndent(),
            NameSearchResponse::class.java
        )

        assertEquals(240, response.eligibleCandidateCount)
        assertEquals("combined", response.appliedMode)
        assertEquals(1, response.returnedCount)
        assertEquals(0.20f, response.appliedWeights!!.numerology, 0.0001f)
        assertEquals("first_name", response.policy.soundScope)
        assertEquals("thai_sum_v1_dictionary", response.policy.numerologyRuleVersion)
        assertEquals("ranking_precision_12dp_v1", response.policy.rankingPrecisionVersion)
        assertEquals("sound_45_35_20_v1", response.policy.soundFormulaVersion)
        assertEquals("live_repository_distribution_provisional", response.policy.soundCalibrationSource)
        assertEquals("typhoon_v1_promptA", response.policy.soundCalibrationReferenceVersion)
        assertEquals("title_prefix_v1_child_titles", response.policy.corpusFilterVersion)
        assertEquals("2026-08-07T02:00:00Z", response.policy.soundProfileLoadedAt)
        assertEquals("sound_eval_v1", response.policy.soundEvaluationVersion)
        assertEquals("sound-corpus-hash", response.policy.soundCorpusFingerprintSha256)
        assertEquals("text-embedding-3-small", response.policy.semanticQueryModel)
        assertEquals("wish_query_v1", response.policy.semanticInputTemplateVersion)
        assertEquals(1536, response.policy.semanticVectorDimensions)
        assertTrue(response.policy.semanticCandidateCorpusApproved)
        assertEquals("approved_snapshot", response.policy.semanticCandidateCorpusSource)
        assertEquals("corpus-hash", response.policy.semanticCandidateCorpusSha256)
        assertEquals("membership-hash", response.policy.semanticCandidateMembershipSha256)
        assertEquals("2026-08-07T01:00:00Z", response.policy.semanticCandidateCorpusAuditAt)
        assertEquals("semantic_eval_v1", response.policy.semanticEvaluationVersion)
        assertFalse(response.policy.soundCalibrationApproved)
        assertFalse(response.semanticAvailable)
        assertEquals("text-embedding-3-small", response.semanticQueryModel)
        assertEquals("legacy-should-not-win", response.legacySemanticModel)
        assertEquals("text-embedding-3-small", response.resolvedSemanticQueryModel())
        assertEquals("abc123", response.semanticVectorHash)
        assertTrue(response.candidateVectorProvenanceVerified)
        assertEquals("approved_snapshot", response.candidateVectorProvenanceSource)
        assertEquals(88.5f, response.data.single().rawSoundScore, 0.0001f)
        assertEquals("v2", response.data.single().phoneticVersion)
    }

    @Test
    fun legacyResponse_keepsBackwardCompatibleSafeDefaults() {
        val response = gson.fromJson(
            """{"error":false,"mode":"standalone","count":0,"semantic_model":"legacy-model","data":[]}""",
            NameSearchResponse::class.java
        )

        assertNull(response.appliedWeights)
        assertNull(response.requestedWeights)
        assertFalse(response.semanticAvailable)
        assertEquals("first_name", response.policy.kalakiniScope)
        assertEquals("first_name", response.policy.soundScope)
        assertEquals("legacy-model", response.legacySemanticModel)
        assertEquals("legacy-model", response.resolvedSemanticQueryModel())
        assertFalse(response.candidateVectorProvenanceVerified)
        assertTrue(response.data.isEmpty())
        assertNotNull(response.policy)
    }

    @Test
    fun decodeMetadata_keepsContradictoryDPairFailClosed() {
        val response = gson.fromJson(
            """
            {
              "error": false,
              "pairs": ["40"],
              "pair_types": ["D5"],
              "pair_points": [-10],
              "pair_is_good": [false]
            }
            """.trimIndent(),
            DecodeResponse::class.java
        )

        assertEquals(-10, response.pairPoints!!.single())
        assertFalse(response.pairIsGood!!.single())
    }
}
