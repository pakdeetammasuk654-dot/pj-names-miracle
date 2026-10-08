package com.miracle.names.ui.auspicious

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class ThaiCalendarPickerLogicTest {
    @Test
    fun `converts typed Buddhist Era birth year`() {
        assertEquals(1982, parseBuddhistYear("2525", 1900, 2026))
    }

    @Test
    fun `accepts surrounding whitespace`() {
        assertEquals(1982, parseBuddhistYear(" 2525 ", 1900, 2026))
    }

    @Test
    fun `rejects malformed and out of range years`() {
        assertNull(parseBuddhistYear("สองพันห้าร้อย", 1900, 2026))
        assertNull(parseBuddhistYear("2442", 1900, 2026))
        assertNull(parseBuddhistYear("2570", 1900, 2026))
    }
}
