import Foundation

struct AuspiciousActivity: Codable, Identifiable, Hashable {
    let code: String
    let nameTH: String
    let icon: String
    let description: String
    let ruleScopeNote: String
    let safetyNote: String?
    let sortOrder: Int
    let categoryCode: String
    let categoryNameTH: String
    let categoryIcon: String
    let categorySortOrder: Int

    var id: String { code }

    enum CodingKeys: String, CodingKey {
        case code, icon, description
        case nameTH = "name_th"
        case ruleScopeNote = "rule_scope_note"
        case safetyNote = "safety_note"
        case sortOrder = "sort_order"
        case categoryCode = "category_code"
        case categoryNameTH = "category_name_th"
        case categoryIcon = "category_icon"
        case categorySortOrder = "category_sort_order"
    }
}

struct AuspiciousActivitiesResponse: Decodable {
    let activities: [AuspiciousActivity]
}

struct AuspiciousSearchRequest: Encodable {
    let activityCode: String
    let startDate: String
    let endDate: String
    let birthDate: String
    let birthTime: String
    let preferWeekend: Bool
    let avoidWanPhra: Bool
    let monkCeremony: Bool

    enum CodingKeys: String, CodingKey {
        case activityCode = "activity_code"
        case startDate = "start_date"
        case endDate = "end_date"
        case birthDate = "birth_date"
        case birthTime = "birth_time"
        case preferWeekend = "prefer_weekend"
        case avoidWanPhra = "avoid_wan_phra"
        case monkCeremony = "monk_ceremony"
    }
}

struct AuspiciousReason: Decodable, Hashable {
    let ruleCode: String
    let name: String
    let explanation: String
    let meaning: String
    let impact: String
    let score: Int

    enum CodingKeys: String, CodingKey {
        case name, explanation, meaning, impact, score
        case ruleCode = "rule_code"
    }
}

struct OutfitColor: Codable, Hashable, Identifiable {
    let nameTH: String
    let hex: String

    var id: String { "\(nameTH)-\(hex)" }

    enum CodingKeys: String, CodingKey {
        case hex
        case nameTH = "name_th"
    }
}

struct OutfitColorSet: Codable, Hashable, Identifiable {
    let context: String
    let labelTH: String
    let roleTH: String
    let roleMeaningTH: String
    let planetNameTH: String
    let number: Int
    let colors: [OutfitColor]

    var id: String { "\(context)-\(roleTH)-\(number)" }

    enum CodingKeys: String, CodingKey {
        case context, number, colors
        case labelTH = "label_th"
        case roleTH = "role_th"
        case roleMeaningTH = "role_meaning_th"
        case planetNameTH = "planet_name_th"
    }
}

struct OutfitColorAdvice: Codable, Hashable {
    let auspicious: [OutfitColorSet]
    let inauspicious: [OutfitColorSet]
}

struct AuspiciousDateResult: Decodable, Identifiable, Hashable {
    let date: String
    let thaiDate: String
    let weekday: String
    let lunarLabel: String
    let isWanPhra: Bool
    let importantDay: String?
    let isWeekend: Bool
    let score: Int
    let rating: String
    let reasons: [AuspiciousReason]
    let outfitAdvice: OutfitColorAdvice?

    var id: String { date }

    enum CodingKeys: String, CodingKey {
        case date, weekday, score, rating, reasons
        case thaiDate = "thai_date"
        case lunarLabel = "lunar_label"
        case isWanPhra = "is_wan_phra"
        case importantDay = "important_day"
        case isWeekend = "is_weekend"
        case outfitAdvice = "outfit_advice"
    }
}

struct AuspiciousSearchResponse: Decodable {
    let activity: AuspiciousActivity
    let startDate: String
    let endDate: String
    let evaluatedDays: Int
    let excludedDays: Int
    let recommendations: [AuspiciousDateResult]
    let ruleSets: [String]
    let disclaimer: String

    enum CodingKeys: String, CodingKey {
        case activity, recommendations, disclaimer
        case startDate = "start_date"
        case endDate = "end_date"
        case evaluatedDays = "evaluated_days"
        case excludedDays = "excluded_days"
        case ruleSets = "rule_sets"
    }
}

struct TaksaRole: Decodable, Identifiable, Hashable {
    let roleName: String
    let dayName: String
    let dayCode: String
    let planetNum: Int
    let symbol: String
    let score: Int
    let impact: String
    let description: String
    let meaning: String?
    let auspiciousGuidance: String?

    var id: String { "\(roleName)-\(dayCode)" }

    enum CodingKeys: String, CodingKey {
        case symbol, score, impact, description, meaning
        case auspiciousGuidance = "auspicious_guidance"
        case roleName = "role_name"
        case dayName = "day_name"
        case dayCode = "day_code"
        case planetNum = "planet_num"
    }
}

struct TaksaCalculationResponse: Decodable {
    let isTodayDefault: Bool
    let birthDayCode: String
    let birthDayName: String
    let birthPlanetNum: Int
    let birthTimeShifted: Bool
    let fullAge: Int
    let progressedAge: Int
    let isCenterCell: Bool
    let landedPlanetNum: Int
    let landedPlanetName: String
    let landedPlanetSymbol: String
    let landedPlanetCode: String
    let wheelRotationDegrees: Double
    let heroRoles: [String: TaksaRole]
    let allRoles: [TaksaRole]

    enum CodingKeys: String, CodingKey {
        case isTodayDefault = "is_today_default"
        case birthDayCode = "birth_day_code"
        case birthDayName = "birth_day_name"
        case birthPlanetNum = "birth_planet_num"
        case birthTimeShifted = "birth_time_shifted"
        case fullAge = "full_age"
        case progressedAge = "progressed_age"
        case isCenterCell = "is_center_cell"
        case landedPlanetNum = "landed_planet_num"
        case landedPlanetName = "landed_planet_name"
        case landedPlanetSymbol = "landed_planet_symbol"
        case landedPlanetCode = "landed_planet_code"
        case wheelRotationDegrees = "wheel_rotation_degrees"
        case heroRoles = "hero_roles"
        case allRoles = "all_roles"
    }
}

struct TaksaRequest: Encodable {
    let birthDate: String
    let birthTime: String

    enum CodingKeys: String, CodingKey {
        case birthDate = "birth_date"
        case birthTime = "birth_time"
    }
}
