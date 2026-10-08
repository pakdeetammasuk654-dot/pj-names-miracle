package com.miracle.names.ui.results

import org.junit.Assert.assertEquals
import org.junit.Test

class NameResultsDisplayTest {
    @Test
    fun combinedRankingShowsTheRankedFirstNameTogetherWithSurname() {
        assertEquals(
            "วิจารา วงศ์",
            buildRankingDisplayName(candidate = " วิจารา ", fixedPart = " วงศ์ ", isCombined = true)
        )
    }

    @Test
    fun standaloneRankingDoesNotAppendHiddenSurname() {
        assertEquals(
            "วิจารา",
            buildRankingDisplayName(candidate = "วิจารา", fixedPart = "วงศ์", isCombined = false)
        )
    }

    @Test
    fun surnameSearchShowsFixedFirstNameBeforeCandidateSurname() {
        assertEquals(
            "วิจารา วงศ์",
            buildRankingDisplayName(
                candidate = "วงศ์",
                fixedPart = "วิจารา",
                isCombined = true,
                isFindSurnameMode = true
            )
        )
    }
}
