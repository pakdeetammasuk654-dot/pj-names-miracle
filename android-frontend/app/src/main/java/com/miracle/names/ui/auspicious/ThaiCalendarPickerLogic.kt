package com.miracle.names.ui.auspicious

internal const val BuddhistEraOffset = 543

internal fun parseBuddhistYear(value: String, minimumGregorianYear: Int, maximumGregorianYear: Int): Int? {
    val buddhistYear = value.trim().toIntOrNull() ?: return null
    val gregorianYear = buddhistYear - BuddhistEraOffset
    return gregorianYear.takeIf { it in minimumGregorianYear..maximumGregorianYear }
}
