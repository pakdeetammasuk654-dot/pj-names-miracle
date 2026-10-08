package com.miracle.names.ui.components

import androidx.compose.ui.geometry.Offset
import org.junit.Assert.assertEquals
import org.junit.Test

class RadarWeightSelectorTest {
    @Test
    fun diamondCornersMapToTheExpectedFourDimensions() {
        val width = 400f
        val height = 200f
        val hInset = 28f
        val vInset = 18f
        val rx = (width / 2f) - hInset
        val ry = (height / 2f) - vInset

        assertWeights(listOf(1f, 0f, 0f, 0f), diamondWeightsAt(Offset(200f, 100f - ry), width, height, hInset, vInset))
        assertWeights(listOf(0f, 1f, 0f, 0f), diamondWeightsAt(Offset(200f - rx, 100f), width, height, hInset, vInset))
        assertWeights(listOf(0f, 0f, 0f, 1f), diamondWeightsAt(Offset(200f + rx, 100f), width, height, hInset, vInset))
        assertWeights(listOf(0f, 0f, 1f, 0f), diamondWeightsAt(Offset(200f, 100f + ry), width, height, hInset, vInset))
    }

    @Test
    fun diamondCenterBalancesAllDimensions() {
        assertWeights(
            listOf(0.25f, 0.25f, 0.25f, 0.25f),
            diamondWeightsAt(Offset(200f, 100f), width = 400f, height = 200f)
        )
    }

    private fun assertWeights(expected: List<Float>, actual: List<Float>) {
        assertEquals(1f, actual.sum(), 0.0001f)
        expected.zip(actual).forEach { (expectedValue, actualValue) ->
            assertEquals(expectedValue, actualValue, 0.0001f)
        }
    }
}
