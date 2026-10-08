import Foundation

enum APIError: LocalizedError {
    case invalidConfiguration
    case invalidResponse
    case server(Int)
    case contractMismatch

    var errorDescription: String? {
        switch self {
        case .invalidConfiguration: return "ตั้งค่าที่อยู่เซิร์ฟเวอร์ไม่ถูกต้อง"
        case .invalidResponse: return "ข้อมูลตอบกลับไม่สมบูรณ์"
        case .server: return "เซิร์ฟเวอร์ยังไม่พร้อม กรุณาลองใหม่"
        case .contractMismatch: return "แอปเวอร์ชันนี้ไม่รองรับข้อมูลจากเซิร์ฟเวอร์"
        }
    }
}

@MainActor
final class APIClient {
    static let shared = APIClient()
    static let supportedAPIVersion = "1"
    static let supportedRankingContract = "adaptive_pairs_v1"

    private let session: URLSession
    private let decoder = JSONDecoder()
    private let encoder = JSONEncoder()

    init(session: URLSession = .shared) {
        self.session = session
    }

    private var baseURL: URL? {
        guard let value = Bundle.main.object(forInfoDictionaryKey: "API_BASE_URL") as? String else {
            return nil
        }
        return URL(string: value)
    }

    func readiness() async throws -> ReadinessResponse {
        let request = try makeRequest(path: "readiness", method: "GET")
        return try await send(request, as: ReadinessResponse.self)
    }

    func search(_ payload: NameSearchRequest) async throws -> NameSearchResponse {
        var request = try makeRequest(path: "names/search", method: "POST")
        request.httpBody = try encoder.encode(payload)
        return try await send(request, as: NameSearchResponse.self)
    }
    func rerank(_ payload: NameRerankRequest) async throws -> NameSearchResponse {
        var request = try makeRequest(path: "names/rerank", method: "POST")
        request.httpBody = try encoder.encode(payload)
        return try await send(request, as: NameSearchResponse.self)
    }

    func verifyPurchase(transactionID: String) async throws -> EntitlementResponse {
        var request = try makeRequest(path: "entitlements/verify", method: "POST")
        request.httpBody = try encoder.encode(
            StorePurchaseRequest(
                platform: "ios",
                productID: PurchaseManager.productID,
                transactionID: transactionID
            )
        )
        return try await send(request, as: EntitlementResponse.self)
    }

    func decode(name: String, dayOfBirth: String = "") async throws -> DecodeResponse {
        guard var components = URLComponents(url: try endpoint(path: "decode"), resolvingAgainstBaseURL: false) else { throw APIError.invalidConfiguration }
        components.queryItems = [
            URLQueryItem(name: "name", value: name),
            URLQueryItem(name: "day_of_birth", value: dayOfBirth)
        ]
        guard let url = components.url else { throw APIError.invalidConfiguration }
        var request = URLRequest(url: url); request.httpMethod = "GET"
        return try await send(request, as: DecodeResponse.self)
    }

    func nameSuggestions(query: String, limit: Int = 8) async throws -> NameSuggestionsResponse {
        guard var components = URLComponents(url: try endpoint(path: "names/suggestions"), resolvingAgainstBaseURL: false) else { throw APIError.invalidConfiguration }
        components.queryItems = [
            URLQueryItem(name: "query", value: query),
            URLQueryItem(name: "limit", value: String(limit))
        ]
        guard let url = components.url else { throw APIError.invalidConfiguration }
        var request = URLRequest(url: url); request.httpMethod = "GET"
        if let entitlement = KeychainHelper.shared.get(key: "premium_entitlement_token"), !entitlement.isEmpty {
            request.setValue(entitlement, forHTTPHeaderField: "X-Entitlement-Token")
        }
        return try await send(request, as: NameSuggestionsResponse.self)
    }

    func randomMeanings(limit: Int = 108) async throws -> RandomMeaningsResponse {
        guard var components = URLComponents(url: try endpoint(path: "meanings/random"), resolvingAgainstBaseURL: false) else { throw APIError.invalidConfiguration }
        components.queryItems = [URLQueryItem(name: "limit", value: String(limit))]
        guard let url = components.url else { throw APIError.invalidConfiguration }
        var request = URLRequest(url: url); request.httpMethod = "GET"
        return try await send(request, as: RandomMeaningsResponse.self)
    }

    func pairDetail(number: String) async throws -> PairDetailResponse {
        guard var components = URLComponents(url: try endpoint(path: "pair-detail"), resolvingAgainstBaseURL: false) else { throw APIError.invalidConfiguration }
        components.queryItems = [URLQueryItem(name: "number", value: number)]
        guard let url = components.url else { throw APIError.invalidConfiguration }
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        return try await send(request, as: PairDetailResponse.self)
    }

    func celebrities() async throws -> CelebritiesResponse {
        let request = try makeRequest(path: "celebrities", method: "GET")
        return try await send(request, as: CelebritiesResponse.self)
    }

    func articles() async throws -> [ArticleItem] {
        let request = try makeRequest(path: "articles", method: "GET")
        let response = try await send(request, as: ArticleListResponse.self)
        return response.data
    }

    func auspiciousActivities() async throws -> AuspiciousActivitiesResponse {
        let request = try makeRequest(path: "auspicious/activities", method: "GET")
        return try await send(request, as: AuspiciousActivitiesResponse.self)
    }

    func searchAuspiciousDates(_ payload: AuspiciousSearchRequest) async throws -> AuspiciousSearchResponse {
        var request = try makeRequest(path: "auspicious/search", method: "POST")
        request.httpBody = try encoder.encode(payload)
        return try await send(request, as: AuspiciousSearchResponse.self)
    }

    func taksaCalculation(_ payload: TaksaRequest) async throws -> TaksaCalculationResponse {
        var request = try makeRequest(path: "auspicious/taksa", method: "POST")
        request.httpBody = try encoder.encode(payload)
        return try await send(request, as: TaksaCalculationResponse.self)
    }

    func createChatConversation(customerName: String, customerAvatar: String) async throws -> CreateChatConversationResponse {
        let boundary = "Boundary-\(UUID().uuidString)"
        let url = try endpoint(path: "chat/conversations")
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.timeoutInterval = 30
        request.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        request.setValue(Self.supportedAPIVersion, forHTTPHeaderField: "X-API-Version")

        request.httpBody = createMultipartBody(
            parameters: [
                "customer_name": customerName,
                "customer_avatar": customerAvatar
            ],
            fileData: nil,
            fileParamName: nil,
            fileName: nil,
            mimeType: nil,
            boundary: boundary
        )

        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else { throw APIError.invalidResponse }
        guard (200...299).contains(http.statusCode) else { throw APIError.server(http.statusCode) }
        return try decoder.decode(CreateChatConversationResponse.self, from: data)
    }

    func getChatUnreadCount(conversationId: String, token: String) async throws -> SupportChatUnreadResponse {
        let url = try endpoint(path: "chat/conversations/\(conversationId)/unread")
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        request.timeoutInterval = 15
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        request.setValue(token, forHTTPHeaderField: "X-Chat-Token")
        request.setValue(Self.supportedAPIVersion, forHTTPHeaderField: "X-API-Version")

        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else { throw APIError.invalidResponse }
        guard (200...299).contains(http.statusCode) else { throw APIError.server(http.statusCode) }
        return try decoder.decode(SupportChatUnreadResponse.self, from: data)
    }

    func getChatMessages(conversationId: String, token: String) async throws -> SupportChatMessagesResponse {
        let url = try endpoint(path: "chat/conversations/\(conversationId)/messages")
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        request.timeoutInterval = 30
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        request.setValue(token, forHTTPHeaderField: "X-Chat-Token")
        request.setValue(Self.supportedAPIVersion, forHTTPHeaderField: "X-API-Version")

        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else { throw APIError.invalidResponse }
        guard (200...299).contains(http.statusCode) else { throw APIError.server(http.statusCode) }
        return try decoder.decode(SupportChatMessagesResponse.self, from: data)
    }

    func deleteChatConversation(conversationId: String, token: String) async throws {
        let url = try endpoint(path: "chat/conversations/\(conversationId)")
        var request = URLRequest(url: url)
        request.httpMethod = "DELETE"
        request.timeoutInterval = 30
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        request.setValue(token, forHTTPHeaderField: "X-Chat-Token")
        request.setValue(Self.supportedAPIVersion, forHTTPHeaderField: "X-API-Version")

        let (_, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else { throw APIError.invalidResponse }
        guard http.statusCode == 204 else { throw APIError.server(http.statusCode) }
    }

    func sendChatMessage(conversationId: String, token: String, body: String, imageData: Data?, mimeType: String?) async throws {
        let boundary = "Boundary-\(UUID().uuidString)"
        let url = try endpoint(path: "chat/conversations/\(conversationId)/messages")
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.timeoutInterval = 30
        request.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        request.setValue(token, forHTTPHeaderField: "X-Chat-Token")
        request.setValue(Self.supportedAPIVersion, forHTTPHeaderField: "X-API-Version")

        var params: [String: String] = [:]
        if !body.isEmpty {
            params["body"] = body
        }
        let ext = mimeType == "image/png" ? "png" : (mimeType == "image/webp" ? "webp" : "jpg")
        let fileName = imageData != nil ? "image.\(ext)" : nil

        request.httpBody = createMultipartBody(
            parameters: params,
            fileData: imageData,
            fileParamName: imageData != nil ? "image" : nil,
            fileName: fileName,
            mimeType: mimeType ?? "image/jpeg",
            boundary: boundary
        )

        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else { throw APIError.invalidResponse }
        guard (200...299).contains(http.statusCode) else {
            if let errObj = try? JSONDecoder().decode([String: String].self, from: data), let msg = errObj["error"] {
                throw NSError(domain: "ChatAPI", code: http.statusCode, userInfo: [NSLocalizedDescriptionKey: msg])
            }
            throw APIError.server(http.statusCode)
        }
    }

    func registerChatDeviceToken(conversationId: String, token: String, deviceToken: String, appVersion: String = "1.3.1") async throws {
        let url = try endpoint(path: "chat/conversations/\(conversationId)/device-token")
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.timeoutInterval = 30
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        request.setValue(token, forHTTPHeaderField: "X-Chat-Token")
        request.setValue(Self.supportedAPIVersion, forHTTPHeaderField: "X-API-Version")

        let bodyObj: [String: String] = [
            "platform": "ios",
            "device_token": deviceToken,
            "app_version": appVersion
        ]
        request.httpBody = try JSONSerialization.data(withJSONObject: bodyObj)

        let (_, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else { throw APIError.invalidResponse }
        guard (200...299).contains(http.statusCode) else { throw APIError.server(http.statusCode) }
    }

    func avatarURL(for value: String?) -> URL? {
        guard let value, !value.isEmpty else { return nil }
        if let absolute = URL(string: value), absolute.scheme != nil { return absolute }
        var cleaned = value.hasPrefix("/") ? String(value.dropFirst()) : value
        if !cleaned.hasPrefix("images/") && !cleaned.hasPrefix("avatars/") && !cleaned.hasPrefix("chat-uploads/") && !cleaned.hasPrefix("static/") {
            cleaned = "avatars/" + cleaned
        }
        if let base = baseURL, let scheme = base.scheme, let host = base.host {
            let portString = base.port != nil ? ":\(base.port!)" : ""
            return URL(string: "\(scheme)://\(host)\(portString)/\(cleaned)")
        }
        return URL(string: "https://xn--b3cu8e7ah6h.com/\(cleaned)")
    }

    private func createMultipartBody(
        parameters: [String: String],
        fileData: Data?,
        fileParamName: String?,
        fileName: String?,
        mimeType: String?,
        boundary: String
    ) -> Data {
        var data = Data()
        for (key, value) in parameters {
            data.append("--\(boundary)\r\n".data(using: .utf8)!)
            data.append("Content-Disposition: form-data; name=\"\(key)\"\r\n\r\n".data(using: .utf8)!)
            data.append("\(value)\r\n".data(using: .utf8)!)
        }
        if let fileData = fileData, let fileParamName = fileParamName, let fileName = fileName, let mimeType = mimeType {
            data.append("--\(boundary)\r\n".data(using: .utf8)!)
            data.append("Content-Disposition: form-data; name=\"\(fileParamName)\"; filename=\"\(fileName)\"\r\n".data(using: .utf8)!)
            data.append("Content-Type: \(mimeType)\r\n\r\n".data(using: .utf8)!)
            data.append(fileData)
            data.append("\r\n".data(using: .utf8)!)
        }
        data.append("--\(boundary)--\r\n".data(using: .utf8)!)
        return data
    }

    private func endpoint(path: String) throws -> URL {
        guard let url = baseURL?.appendingPathComponent(path) else { throw APIError.invalidConfiguration }
        return url
    }

    private func makeRequest(path: String, method: String) throws -> URLRequest {
        guard let url = baseURL?.appendingPathComponent(path) else {
            throw APIError.invalidConfiguration
        }
        var request = URLRequest(url: url)
        request.httpMethod = method
        request.timeoutInterval = 120
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue(Self.supportedAPIVersion, forHTTPHeaderField: "X-API-Version")
        if let entitlement = KeychainHelper.shared.get(key: "premium_entitlement_token"), !entitlement.isEmpty {
            request.setValue(entitlement, forHTTPHeaderField: "X-Entitlement-Token")
        }
        return request
    }

    private func send<T: Decodable>(_ request: URLRequest, as type: T.Type) async throws -> T {
        do {
            let (data, response) = try await session.data(for: request)
            guard let http = response as? HTTPURLResponse else { throw APIError.invalidResponse }
            guard (200...299).contains(http.statusCode) else { throw APIError.server(http.statusCode) }
            if let version = http.value(forHTTPHeaderField: "X-API-Version"),
               version != Self.supportedAPIVersion {
                throw APIError.contractMismatch
            }
            if let contract = http.value(forHTTPHeaderField: "X-Ranking-Contract-Version"),
               contract != Self.supportedRankingContract {
                throw APIError.contractMismatch
            }
            return try decoder.decode(type, from: data)
        } catch let urlError as URLError where urlError.code == .timedOut {
            throw APIError.server(504)
        }
    }
}
