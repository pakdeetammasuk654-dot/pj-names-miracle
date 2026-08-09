package com.miracle.names.viewmodel

import com.miracle.names.data.model.RankingWeightsPayload
import com.miracle.names.data.model.NameSearchPolicy
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class RankingStateTest {

    @Test
    fun normalizeRankingWeights_producesOneHundredPercent() {
        val weights = normalizeRankingWeights(0.15f, 0.15f, 0.70f)

        assertEquals(0.15f, weights.numerology, 0.0001f)
        assertEquals(0.15f, weights.semantic, 0.0001f)
        assertEquals(0.70f, weights.sound, 0.0001f)
        assertEquals(1f, weights.numerology + weights.semantic + weights.sound, 0.0001f)
    }

    @Test
    fun resolveAppliedSurname_neverLeaksIntoStandaloneMode() {
        assertEquals("", resolveAppliedSurname(MODE_STANDALONE, "  สมบูรณ์  "))
        assertEquals("สมบูรณ์", resolveAppliedSurname(MODE_COMBINED, "  สมบูรณ์  "))
    }

    @Test
    fun responseWeights_useBackendAppliedFormulaWhenPresent() {
        val selected = RankingWeights(0.15f, 0.15f, 0.70f)
        val applied = resolveResponseWeights(
            payload = RankingWeightsPayload(0.20f, 0f, 0.80f),
            fallback = selected
        )

        assertEquals(0.20f, applied.numerology, 0.0001f)
        assertEquals(0f, applied.semantic, 0.0001f)
        assertEquals(0.80f, applied.sound, 0.0001f)
    }

    @Test
    fun responseWeights_fallBackForMissingOrInvalidMetadata() {
        val selected = RankingWeights(0.45f, 0.35f, 0.20f)

        assertEquals(selected, resolveResponseWeights(null, selected))
        assertEquals(
            selected,
            resolveResponseWeights(RankingWeightsPayload(0f, 0f, 0f), selected)
        )
    }

    @Test
    fun snapshotFullName_obeysAppliedModeNotHiddenFormState() {
        val standalone = snapshot(mode = MODE_STANDALONE, surname = "")
        val combined = snapshot(mode = MODE_COMBINED, surname = "สมบูรณ์")

        assertEquals("วรชัย", standalone.fullName("วรชัย"))
        assertEquals("วรชัย สมบูรณ์", combined.fullName("วรชัย"))
    }

    @Test
    fun countsPreferActualPayloadAndKeepEligiblePoolAtLeastAsLarge() {
        assertEquals(42, resolveActualCount(reportedCount = 100, dataSize = 42))
        assertEquals(42, resolveActualCount(reportedCount = 42, dataSize = 42))
        assertEquals(250, resolveEligibleCandidateCount(reportedCount = 250, resultCount = 42))
        assertEquals(42, resolveEligibleCandidateCount(reportedCount = 0, resultCount = 42))
    }

    @Test
    fun approximateWeightComparison_toleratesSerializationNoiseOnly() {
        val first = RankingWeights(0.15f, 0.15f, 0.70f)

        assertTrue(first.approximatelyEquals(RankingWeights(0.15001f, 0.14999f, 0.70f)))
    }

    @Test
    fun requestRevisionGate_acceptsOnlyNewestSearch() {
        val gate = RequestRevisionGate()
        val firstSearch = gate.begin()
        val secondSearch = gate.begin()

        assertTrue(!gate.accepts(firstSearch))
        assertTrue(gate.accepts(secondSearch))
    }

    @Test
    fun requestRevisionGate_resetInvalidatesInFlightSearch() {
        val gate = RequestRevisionGate()
        val inFlightSearch = gate.begin()

        gate.invalidate()

        assertTrue(!gate.accepts(inFlightSearch))
    }

    @Test
    fun semanticProof_requiresApprovedCorpusAndVerifiedCandidateVectors() {
        val approved = snapshot(mode = MODE_STANDALONE, surname = "").copy(
            semanticAvailable = true,
            semanticSource = "embedding",
            semanticModel = "text-embedding-3-small",
            semanticVectorHash = "query-vector-hash",
            candidateVectorProvenanceVerified = true,
            candidateVectorProvenanceSource = "approved_snapshot",
            policy = NameSearchPolicy(
                semanticQueryModel = "text-embedding-3-small",
                semanticInputTemplateVersion = "wish_query_v1",
                semanticVectorDimensions = 1536,
                semanticCandidateCorpusApproved = true,
                semanticCandidateCorpusSource = "approved_snapshot",
                semanticCandidateCorpusSha256 = "corpus-vector-hash",
                semanticEvaluationVersion = "semantic_eval_v1"
            )
        )

        assertTrue(approved.hasApprovedSemanticProof)
        assertTrue(approved.semanticStatusThai().contains("ผ่านการยืนยัน"))
        assertTrue(approved.semanticProofThai().contains("1536 มิติ"))
        assertTrue(approved.semanticProofThai().contains("semantic_eval_v1"))
    }

    @Test
    fun semanticStatus_explainsFailClosedStaleCorpus() {
        val unavailable = snapshot(mode = MODE_STANDALONE, surname = "").copy(
            semanticAvailable = false,
            semanticSource = "unavailable_stale_vector_corpus",
            candidateVectorProvenanceVerified = false,
            candidateVectorProvenanceSource = "unavailable_stale_vector_corpus",
            policy = NameSearchPolicy(
                semanticCandidateCorpusApproved = false,
                semanticCandidateCorpusSource = "unavailable_stale_vector_corpus"
            )
        )

        assertTrue(!unavailable.hasApprovedSemanticProof)
        assertTrue(unavailable.semanticStatusThai().contains("ล้าสมัย"))
        assertEquals("", unavailable.semanticProofThai())
    }

    private fun snapshot(mode: String, surname: String): AppliedSearchSnapshot {
        return AppliedSearchSnapshot(
            mode = mode,
            surname = surname,
            dayOfBirth = "monday",
            wishQuery = "มงคล",
            weights = RankingWeights.Default,
            count = 1
        )
    }
}
