package com.miracle.names.data.model

import com.google.gson.annotations.SerializedName

data class AuspiciousActivity(
    @SerializedName("code") val code: String = "",
    @SerializedName("name_th") val nameTh: String = "",
    @SerializedName("icon") val icon: String = "✦",
    @SerializedName("description") val description: String = "",
    @SerializedName("rule_scope_note") val ruleScopeNote: String = "",
    @SerializedName("safety_note") val safetyNote: String = "",
    @SerializedName("sort_order") val sortOrder: Int = 0,
    @SerializedName("category_code") val categoryCode: String = "other",
    @SerializedName("category_name_th") val categoryNameTh: String = "กิจกรรมอื่น",
    @SerializedName("category_icon") val categoryIcon: String = "✦",
    @SerializedName("category_sort_order") val categorySortOrder: Int = 0
)

data class AuspiciousActivitiesResponse(
    @SerializedName("activities") val activities: List<AuspiciousActivity> = emptyList()
)

data class AuspiciousSearchRequest(
    @SerializedName("activity_code") val activityCode: String,
    @SerializedName("start_date") val startDate: String,
    @SerializedName("end_date") val endDate: String,
    @SerializedName("birth_date") val birthDate: String = "",
    @SerializedName("birth_time") val birthTime: String = "",
    @SerializedName("prefer_weekend") val preferWeekend: Boolean = false,
    @SerializedName("avoid_wan_phra") val avoidWanPhra: Boolean = false,
    @SerializedName("monk_ceremony") val monkCeremony: Boolean = false
)

data class AuspiciousReason(
    @SerializedName("rule_code") val ruleCode: String = "",
    @SerializedName("name") val name: String = "",
    @SerializedName("explanation") val explanation: String = "",
    @SerializedName("meaning") val meaning: String = "",
    @SerializedName("impact") val impact: String = "neutral",
    @SerializedName("score") val score: Int = 0
)

data class OutfitColor(
    @SerializedName("name_th") val nameTh: String = "",
    @SerializedName("hex") val hex: String = ""
)

data class OutfitColorSet(
    @SerializedName("context") val context: String = "",
    @SerializedName("label_th") val labelTh: String = "",
    @SerializedName("role_th") val roleTh: String = "",
    @SerializedName("role_meaning_th") val roleMeaningTh: String = "",
    @SerializedName("planet_name_th") val planetNameTh: String = "",
    @SerializedName("number") val number: Int = 0,
    @SerializedName("colors") val colors: List<OutfitColor> = emptyList()
)

data class OutfitColorAdvice(
    @SerializedName("auspicious") val auspicious: List<OutfitColorSet> = emptyList(),
    @SerializedName("inauspicious") val inauspicious: List<OutfitColorSet> = emptyList()
)

data class AuspiciousDateResult(
    @SerializedName("date") val date: String = "",
    @SerializedName("thai_date") val thaiDate: String = "",
    @SerializedName("weekday") val weekday: String = "",
    @SerializedName("lunar_label") val lunarLabel: String = "",
    @SerializedName("is_wan_phra") val isWanPhra: Boolean = false,
    @SerializedName("important_day") val importantDay: String = "",
    @SerializedName("is_weekend") val isWeekend: Boolean = false,
    @SerializedName("score") val score: Int = 0,
    @SerializedName("rating") val rating: String = "",
    @SerializedName("reasons") val reasons: List<AuspiciousReason> = emptyList(),
    @SerializedName("outfit_advice") val outfitAdvice: OutfitColorAdvice? = null
)

data class AuspiciousSearchResponse(
    @SerializedName("activity") val activity: AuspiciousActivity = AuspiciousActivity(),
    @SerializedName("start_date") val startDate: String = "",
    @SerializedName("end_date") val endDate: String = "",
    @SerializedName("evaluated_days") val evaluatedDays: Int = 0,
    @SerializedName("excluded_days") val excludedDays: Int = 0,
    @SerializedName("recommendations") val recommendations: List<AuspiciousDateResult> = emptyList(),
    @SerializedName("rule_sets") val ruleSets: List<String> = emptyList(),
    @SerializedName("disclaimer") val disclaimer: String = ""
)

data class TaksaRole(
    @SerializedName("role_name") val roleName: String = "",
    @SerializedName("day_name") val dayName: String = "",
    @SerializedName("day_code") val dayCode: String = "",
    @SerializedName("planet_num") val planetNum: Int = 0,
    @SerializedName("symbol") val symbol: String = "✦",
    @SerializedName("score") val score: Int = 0,
    @SerializedName("impact") val impact: String = "support",
    @SerializedName("description") val description: String = "",
    @SerializedName("meaning") val meaning: String? = null,
    @SerializedName("auspicious_guidance") val auspiciousGuidance: String? = null
)

data class TaksaCalculationResponse(
    @SerializedName("is_today_default") val isTodayDefault: Boolean = true,
    @SerializedName("birth_day_code") val birthDayCode: String = "",
    @SerializedName("birth_day_name") val birthDayName: String = "",
    @SerializedName("birth_planet_num") val birthPlanetNum: Int = 0,
    @SerializedName("birth_time_shifted") val birthTimeShifted: Boolean = false,
    @SerializedName("full_age") val fullAge: Int = 0,
    @SerializedName("progressed_age") val progressedAge: Int = 0,
    @SerializedName("is_center_cell") val isCenterCell: Boolean = false,
    @SerializedName("landed_planet_num") val landedPlanetNum: Int = 0,
    @SerializedName("landed_planet_name") val landedPlanetName: String = "",
    @SerializedName("landed_planet_symbol") val landedPlanetSymbol: String = "✦",
    @SerializedName("landed_planet_code") val landedPlanetCode: String = "",
    @SerializedName("wheel_rotation_degrees") val wheelRotationDegrees: Float = 0f,
    @SerializedName("hero_roles") val heroRoles: Map<String, TaksaRole> = emptyMap(),
    @SerializedName("all_roles") val allRoles: List<TaksaRole> = emptyList()
)
