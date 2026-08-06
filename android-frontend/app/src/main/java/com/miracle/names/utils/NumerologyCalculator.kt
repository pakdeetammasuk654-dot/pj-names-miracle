package com.miracle.names.utils

object NumerologyCalculator {

    // Standard Thai Numerology Map (เลขศาสตร์)
    // 0 points: วรรณยุกต์ (่ ้ ๊ ๋) ไม้ไต่คู้ (็)
    private val charValueMap = mapOf(
        // 1
        'ก' to 1, 'ด' to 1, 'ถ' to 1, 'ท' to 1, 'ภ' to 1, 'ฤ' to 1, 'า' to 1, 'ำ' to 1, 'ุ' to 1,
        // 2
        'ข' to 2, 'ช' to 2, 'บ' to 2, 'ป' to 2, 'ู' to 2, 'เ' to 2,
        // 3
        'ฆ' to 3, 'ฑ' to 3, 'ฒ' to 3, 'ต' to 3, 'แ' to 3,
        // 4
        'ค' to 4, 'ธ' to 4, 'ร' to 4, 'ญ' to 4, 'ษ' to 4, 'โ' to 4, 'ะ' to 4, 'ั' to 4,
        // 5
        'ฉ' to 5, 'ณ' to 5, 'ฌ' to 5, 'น' to 5, 'ม' to 5, 'ห' to 5, 'ฮ' to 5, 'ฬ' to 5, 'ึ' to 5,
        // 6
        'จ' to 6, 'ล' to 6, 'ว' to 6, 'อ' to 6, 'ใ' to 6,
        // 7
        'ซ' to 7, 'ศ' to 7, 'ส' to 7, 'ี' to 7, 'ื' to 7,
        // 8
        'ย' to 8, 'ผ' to 8, 'ฝ' to 8, 'พ' to 8, 'ฟ' to 8,
        // 9
        'ฏ' to 9, 'ฐ' to 9, 'ไ' to 9, '์' to 9
    )

    fun getCharValue(c: Char): Int {
        return charValueMap[c] ?: 0
    }

    /**
     * Decode a string into a list of numbers (excluding zeros)
     */
    fun decodeString(text: String): List<Int> {
        val numbers = mutableListOf<Int>()
        for (char in text) {
            val value = getCharValue(char)
            if (value > 0) {
                numbers.add(value)
            }
        }
        return numbers
    }

    /**
     * Calculate total score for a string
     */
    fun calculateSum(text: String): Int {
        return decodeString(text).sum()
    }

    /**
     * Get astrological pairs (e.g., [4, 5, 1] -> ["45", "51"])
     */
    fun getPairs(numbers: List<Int>): List<String> {
        val pairs = mutableListOf<String>()
        if (numbers.size < 2) return pairs
        
        for (i in 0 until numbers.size - 1) {
            pairs.add("${numbers[i]}${numbers[i+1]}")
        }
        return pairs
    }

    private val goodPairs = setOf(
        "14", "41", "15", "51", "19", "91", "24", "42", "26", "62", "28", "82", "29", "92",
        "36", "63", "39", "93", "44", "45", "54", "46", "64", "49", "94", "55", "56", "65",
        "59", "95", "66", "69", "96", "78", "87", "79", "97", "89", "98", "99"
    )

    fun getPairType(pair: String): String {
        return if (goodPairs.contains(pair)) "D" else "R"
    }
}
