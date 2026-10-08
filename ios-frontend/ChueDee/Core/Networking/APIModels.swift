import Foundation

struct RankingWeights: Codable, Equatable {
    let numerology: Double
    let semantic: Double
    let sound: Double
    let similarity: Double

    init(numerology: Double, semantic: Double, sound: Double, similarity: Double = 0) {
        self.numerology = numerology
        self.semantic = semantic
        self.sound = sound
        self.similarity = similarity
    }

    enum CodingKeys: String, CodingKey {
        case numerology, semantic, sound, similarity
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        numerology = try container.decode(Double.self, forKey: .numerology)
        semantic = try container.decode(Double.self, forKey: .semantic)
        sound = try container.decode(Double.self, forKey: .sound)
        similarity = try container.decodeIfPresent(Double.self, forKey: .similarity) ?? 0
    }
}

struct NameProfile4D: Decodable, Equatable {
    let numerologyScore: Double
    let numerologyLabel: String
    let meaningScore: Double
    let meaningLabel: String
    let soundScore: Double
    let soundLabel: String
    let resonanceScore: Double
    let resonanceLabel: String

    enum CodingKeys: String, CodingKey {
        case numerologyScore = "numerology_score"
        case numerologyLabel = "numerology_label"
        case meaningScore = "meaning_score"
        case meaningLabel = "meaning_label"
        case soundScore = "sound_score"
        case soundLabel = "sound_label"
        case resonanceScore = "resonance_score"
        case resonanceLabel = "resonance_label"
    }
}

struct DecodeResponse: Decodable {
    let error: Bool
    let meaning: String?
    let pairs: [String]?
    let pairTypes: [String]?
    let pairDescs: [String]?
    let profile4D: NameProfile4D?
    let kalakini: KalakiniInfo?
    enum CodingKeys: String, CodingKey {
        case error, meaning, pairs
        case pairTypes = "pair_types"
        case pairDescs = "pair_descs"
        case profile4D = "profile_4d"
        case kalakini
    }
}

struct KalakiniInfo: Decodable {
    let applied: Bool
    let letters: String
    let matches: String
}

struct NameSuggestionsResponse: Decodable {
    let error: Bool
    let data: [NameResultItem]
    let previewPairs: [SuggestionPairPreview]
    let count: Int
    let premiumUnlocked: Bool

    enum CodingKeys: String, CodingKey {
        case error, data, count
        case previewPairs = "preview_pairs"
        case premiumUnlocked = "premium_unlocked"
    }
}

struct SuggestionPairPreview: Decodable {
    let maskedName: String?
    let satPairs: [PairPresentation]
    let shaPairs: [PairPresentation]

    enum CodingKeys: String, CodingKey {
        case maskedName = "masked_name"
        case satPairs = "sat_pairs"
        case shaPairs = "sha_pairs"
    }

    var pairs: [(value: String, isBad: Bool)] {
        (satPairs + shaPairs).map { ($0.value, !$0.isGood) }
    }
}

struct RandomMeaningsResponse: Decodable { let error: Bool; let meanings: [String] }

struct PairDetailResponse: Decodable {
    let error: Bool
    let pairNumber: String
    let pairType: String
    let pairPoint: Int
    let miracleDesc: String
    let miracleDetail: String
    let isGood: Bool

    enum CodingKeys: String, CodingKey {
        case error
        case pairNumber = "pair_number"
        case pairType = "pair_type"
        case pairPoint = "pair_point"
        case miracleDesc = "miracle_desc"
        case miracleDetail = "miracle_detail"
        case isGood = "is_good"
    }
}

struct Celebrity: Decodable, Identifiable {
    let name: String
    let avatarURL: String?
    var id: String { name }
    enum CodingKeys: String, CodingKey { case name; case avatarURL = "avatar_url" }
}

struct CelebritiesResponse: Decodable { let error: Bool; let data: [Celebrity] }

struct NameSearchRequest: Encodable {
    let mode: String
    let searchTarget: String
    let surname: String
    let prototypeName: String
    let dayOfBirth: String
    let wishQuery: String
    let isNameIntent: Bool
    let limit: Int
    let numerologyWeight: Double
    let semanticWeight: Double
    let soundWeight: Double
    let similarityWeight: Double

    enum CodingKeys: String, CodingKey {
        case mode, surname, limit
        case searchTarget = "search_target"
        case prototypeName = "prototype_name"
        case dayOfBirth = "day_of_birth"
        case wishQuery = "wish_query"
        case isNameIntent = "is_name_intent"
        case numerologyWeight = "numerology_weight"
        case semanticWeight = "semantic_weight"
        case soundWeight = "sound_weight"
        case similarityWeight = "similarity_weight"
    }
}

struct NameRerankRequest: Encodable {
    let candidateToken: String
    let limit: Int
    let numerologyWeight: Double
    let semanticWeight: Double
    let soundWeight: Double
    let similarityWeight: Double
    enum CodingKeys: String, CodingKey {
        case limit
        case candidateToken = "candidate_token"
        case numerologyWeight = "numerology_weight"
        case semanticWeight = "semantic_weight"
        case soundWeight = "sound_weight"
        case similarityWeight = "similarity_weight"
    }
}

struct StorePurchaseRequest: Encodable {
    let platform: String
    let productID: String
    let transactionID: String

    enum CodingKeys: String, CodingKey {
        case platform
        case productID = "product_id"
        case transactionID = "transaction_id"
    }
}

struct EntitlementResponse: Decodable {
    let entitlementToken: String
    let productID: String
    let expiresAt: String

    enum CodingKeys: String, CodingKey {
        case entitlementToken = "entitlement_token"
        case productID = "product_id"
        case expiresAt = "expires_at"
    }
}

struct NameResultItem: Decodable, Identifiable {
    let nameID: Int
    let thname: String
    let meaning: String?
    let rank: Int
    let finalScore: Double
    let semanticScore: Double?
    let soundScore: Double?
    let satSum: Int?
    let shaSum: Int?
    let totalSatSum: Int?
    let totalShaSum: Int?
    let satPairsDisplay: String?
    let shaPairsDisplay: String?
    let satPairs: [PairPresentation]?
    let shaPairs: [PairPresentation]?
    let candidateSatPairs: [PairPresentation]?
    let candidateShaPairs: [PairPresentation]?
    let satPairtype: String?
    let shaPairtype: String?
    let satMiracleDesc: String?
    let shaMiracleDesc: String?
    let hasBadPair: Bool
    let badPairCount: Int
    let pairRiskLevel: String?
    let rankingReasons: [String]?

    var id: Int { nameID }

    enum CodingKeys: String, CodingKey {
        case thname, meaning, rank
        case nameID = "name_id"
        case satSum = "sat_sum"
        case shaSum = "sha_sum"
        case totalSatSum = "total_sat_sum"
        case totalShaSum = "total_sha_sum"
        case finalScore = "final_score"
        case semanticScore = "semantic_score"
        case soundScore = "sound_score"
        case satPairsDisplay = "sat_pairs_display"
        case shaPairsDisplay = "sha_pairs_display"
        case satPairs = "sat_pairs"
        case shaPairs = "sha_pairs"
        case candidateSatPairs = "candidate_sat_pairs"
        case candidateShaPairs = "candidate_sha_pairs"
        case satPairtype = "sat_pairtype"
        case shaPairtype = "sha_pairtype"
        case satMiracleDesc = "sat_miracledesc"
        case shaMiracleDesc = "sha_miracledesc"
        case hasBadPair = "has_bad_pair"
        case badPairCount = "bad_pair_count"
        case pairRiskLevel = "pair_risk_level"
        case rankingReasons = "ranking_reasons"
    }

    var candidatePairs: [(value: String, isBad: Bool)] {
        let structured = (candidateSatPairs ?? []) + (candidateShaPairs ?? [])
        if !structured.isEmpty {
            return structured.map { ($0.value, !$0.isGood) }
        }
        var pairs: [String] = []
        if let s = satSum, s > 0 {
            pairs.append(contentsOf: formatTotalValue(s))
        }
        if let sh = shaSum, sh > 0 {
            pairs.append(contentsOf: formatTotalValue(sh))
        }
        return pairs.map { ($0, false) }
    }

    var displayedPairs: [(value: String, isBad: Bool)] {
        let structured = (satPairs ?? []) + (shaPairs ?? [])
        if !structured.isEmpty {
            return structured.map { ($0.value, !$0.isGood) }
        }
        let values = [satPairsDisplay, shaPairsDisplay]
            .compactMap { $0 }
            .flatMap { $0.components(separatedBy: CharacterSet.decimalDigits.inverted) }
            .filter { !$0.isEmpty }
        let types = [satPairtype, shaPairtype].compactMap { $0?.uppercased() }
        return values.enumerated().map { index, value in
            let isBad = index < types.count && types[index].hasPrefix("R")
            return (value, isBad)
        }
    }

    private func formatTotalValue(_ total: Int) -> [String] {
        guard total > 0 else { return [] }
        let s = String(total)
        if s.count == 1 { return [String(format: "%02d", total)] }
        if s.count == 2 { return [s] }
        var result: [String] = []
        for i in 0..<(s.count - 1) {
            let start = s.index(s.startIndex, offsetBy: i)
            let end = s.index(start, offsetBy: 2)
            result.append(String(s[start..<end]))
        }
        return result
    }
}

struct PairPresentation: Decodable {
    let value: String
    let type: String
    let isGood: Bool

    enum CodingKeys: String, CodingKey {
        case value, type
        case isGood = "is_good"
    }
}

struct RankingPolicy: Decodable {
    let rankingContractVersion: String
    let pairScope: String

    enum CodingKeys: String, CodingKey {
        case rankingContractVersion = "ranking_contract_version"
        case pairScope = "pair_scope"
    }
}

struct NameSearchResponse: Decodable {
    let error: Bool
    let data: [NameResultItem]
    let candidateToken: String
    let requestedWeights: RankingWeights
    let appliedWeights: RankingWeights
    let policy: RankingPolicy

    enum CodingKeys: String, CodingKey {
        case error, data, policy
        case candidateToken = "candidate_token"
        case requestedWeights = "requested_weights"
        case appliedWeights = "applied_weights"
    }
}

struct ReadinessResponse: Decodable {
    struct Status: Decodable {
        let status: String
        let apiVersion: String
        let rankingContractVersion: String

        enum CodingKeys: String, CodingKey {
            case status
            case apiVersion = "api_version"
            case rankingContractVersion = "ranking_contract_version"
        }
    }

    let error: Bool
    let data: Status
}

struct SupportChatMessage: Decodable, Identifiable {
    let id: String
    let sender: String
    let body: String
    let imageURL: String
    let isRead: Bool?
    let createdAt: String

    var isFromCustomer: Bool { sender == "customer" }

    enum CodingKeys: String, CodingKey {
        case id, sender, body
        case imageURL = "image_url"
        case isRead = "is_read"
        case createdAt = "created_at"
    }
}

struct SupportChatMessagesResponse: Decodable {
    let messages: [SupportChatMessage]
    let customerAvatar: String?
    let customerName: String?
    let customerUnreadCount: Int?
    let status: String?

    enum CodingKeys: String, CodingKey {
        case messages, status
        case customerAvatar = "customer_avatar"
        case customerName = "customer_name"
        case customerUnreadCount = "customer_unread_count"
    }
}

struct SupportChatUnreadResponse: Decodable {
    let unreadCount: Int
    let status: String?
    let customerAvatar: String?

    enum CodingKeys: String, CodingKey {
        case status
        case unreadCount = "unread_count"
        case customerAvatar = "customer_avatar"
    }
}

struct CreateChatConversationResponse: Decodable {
    let conversationID: String
    let customerToken: String
    let customerAvatar: String?

    enum CodingKeys: String, CodingKey {
        case conversationID = "conversation_id"
        case customerToken = "customer_token"
        case customerAvatar = "customer_avatar"
    }
}

struct ArticleItem: Identifiable, Codable, Equatable {
    var id: String { slug }
    let slug: String
    let title: String
    let titleShort: String?
    let excerpt: String
    let category: String
    let imageUrl: String

    enum CodingKeys: String, CodingKey {
        case slug, title
        case titleShort = "title_short"
        case excerpt, category
        case imageUrl = "image_url"
    }

    var fullImageUrl: URL? {
        if imageUrl.starts(with: "http") {
            return URL(string: imageUrl)
        }
        return URL(string: "https://xn--b3cu8e7ah6h.com\(imageUrl)")
    }

    var fullArticleUrl: URL? {
        return URL(string: "https://xn--b3cu8e7ah6h.com/articles/\(slug)")
    }
}

struct ArticleListResponse: Decodable {
    let data: [ArticleItem]
}

