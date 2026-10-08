import Foundation

enum AdminAPIError: LocalizedError {
    case unauthorized
    case invalidURL
    case invalidResponse
    case server(Int)
    case custom(String)

    var errorDescription: String? {
        switch self {
        case .unauthorized:
            return "ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง"
        case .invalidURL:
            return "URL ไม่ถูกต้อง"
        case .invalidResponse:
            return "การตอบกลับจากเซิร์ฟเวอร์ไม่ถูกต้อง"
        case .server(let code):
            return "เซิร์ฟเวอร์ตอบกลับผิดพลาด (\(code))"
        case .custom(let msg):
            return msg
        }
    }
}

final class AdminAPIClient: @unchecked Sendable {
    static let shared = AdminAPIClient()

    private let baseURL: URL
    private let session: URLSession

    private init() {
        let urlString = Bundle.main.object(forInfoDictionaryKey: "API_BASE_URL") as? String ?? "https://xn--b3cu8e7ah6h.com/api/v1"
        self.baseURL = URL(string: urlString) ?? URL(string: "https://xn--b3cu8e7ah6h.com/api/v1")!
        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = 20
        config.timeoutIntervalForResource = 45
        self.session = URLSession(configuration: config)
    }

    private func authHeader() -> String? {
        guard let user = AdminKeychainHelper.shared.get(key: "admin_username"),
              let pass = AdminKeychainHelper.shared.get(key: "admin_password"),
              !user.isEmpty, !pass.isEmpty else {
            return nil
        }
        let creds = "\(user):\(pass)"
        guard let data = creds.data(using: .utf8) else { return nil }
        return "Basic " + data.base64EncodedString()
    }

    private func urlForPath(_ path: String) throws -> URL {
        let clean = path.hasPrefix("/") ? String(path.dropFirst()) : path
        var baseStr = baseURL.absoluteString
        if !baseStr.hasSuffix("/") {
            baseStr += "/"
        }
        guard let base = URL(string: baseStr), let url = URL(string: clean, relativeTo: base) else {
            throw AdminAPIError.invalidURL
        }
        return url.absoluteURL
    }

    private func makeRequest(path: String, method: String = "GET") throws -> URLRequest {
        let url = try urlForPath(path)
        var req = URLRequest(url: url)
        req.httpMethod = method
        if let auth = authHeader() {
            req.setValue(auth, forHTTPHeaderField: "Authorization")
        }
        req.setValue("application/json", forHTTPHeaderField: "Accept")
        return req
    }

    func login(username: String, password: String) async throws -> AdminAuthResponse {
        let url = try urlForPath("admin/auth/login")
        var req = URLRequest(url: url)
        req.httpMethod = "POST"
        let creds = "\(username):\(password)"
        guard let data = creds.data(using: .utf8) else { throw AdminAPIError.unauthorized }
        req.setValue("Basic " + data.base64EncodedString(), forHTTPHeaderField: "Authorization")
        req.setValue("application/json", forHTTPHeaderField: "Accept")

        let (respData, response) = try await session.data(for: req)
        guard let http = response as? HTTPURLResponse else { throw AdminAPIError.invalidResponse }
        if http.statusCode == 401 { throw AdminAPIError.unauthorized }
        guard (200...299).contains(http.statusCode) else { throw AdminAPIError.server(http.statusCode) }

        return try JSONDecoder().decode(AdminAuthResponse.self, from: respData)
    }

    func getConversations(inbox: String = "needs_reply") async throws -> AdminConversationsResponse {
        let req = try makeRequest(path: "admin/chat/conversations?view=\(inbox)")
        let (data, response) = try await session.data(for: req)
        guard let http = response as? HTTPURLResponse else { throw AdminAPIError.invalidResponse }
        if http.statusCode == 401 { throw AdminAPIError.unauthorized }
        guard (200...299).contains(http.statusCode) else { throw AdminAPIError.server(http.statusCode) }
        return try JSONDecoder().decode(AdminConversationsResponse.self, from: data)
    }

    func getMessages(conversationId: String) async throws -> AdminConversationDetailResponse {
        let req = try makeRequest(path: "admin/chat/conversations/\(conversationId)/messages")
        let (data, response) = try await session.data(for: req)
        guard let http = response as? HTTPURLResponse else { throw AdminAPIError.invalidResponse }
        if http.statusCode == 401 { throw AdminAPIError.unauthorized }
        guard (200...299).contains(http.statusCode) else { throw AdminAPIError.server(http.statusCode) }
        return try JSONDecoder().decode(AdminConversationDetailResponse.self, from: data)
    }

    func sendMessage(conversationId: String, body: String, imageData: Data?, mimeType: String?) async throws {
        let url = try urlForPath("admin/chat/conversations/\(conversationId)/messages")
        var req = URLRequest(url: url)
        req.httpMethod = "POST"
        if let auth = authHeader() {
            req.setValue(auth, forHTTPHeaderField: "Authorization")
        }

        let boundary = "Boundary-\(UUID().uuidString)"
        req.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")

        var bodyData = Data()
        if !body.isEmpty {
            bodyData.append("--\(boundary)\r\n".data(using: .utf8)!)
            bodyData.append("Content-Disposition: form-data; name=\"body\"\r\n\r\n".data(using: .utf8)!)
            bodyData.append("\(body)\r\n".data(using: .utf8)!)
        }
        if let img = imageData {
            let mime = mimeType ?? "image/jpeg"
            let ext = mime.contains("png") ? "png" : "jpg"
            bodyData.append("--\(boundary)\r\n".data(using: .utf8)!)
            bodyData.append("Content-Disposition: form-data; name=\"image\"; filename=\"reply.\(ext)\"\r\n".data(using: .utf8)!)
            bodyData.append("Content-Type: \(mime)\r\n\r\n".data(using: .utf8)!)
            bodyData.append(img)
            bodyData.append("\r\n".data(using: .utf8)!)
        }
        bodyData.append("--\(boundary)--\r\n".data(using: .utf8)!)
        req.httpBody = bodyData

        let (_, response) = try await session.data(for: req)
        guard let http = response as? HTTPURLResponse else { throw AdminAPIError.invalidResponse }
        if http.statusCode == 401 { throw AdminAPIError.unauthorized }
        guard (200...299).contains(http.statusCode) else { throw AdminAPIError.server(http.statusCode) }
    }

    func closeConversation(conversationId: String) async throws {
        var req = try makeRequest(path: "admin/chat/conversations/\(conversationId)/close", method: "POST")
        let (_, response) = try await session.data(for: req)
        guard let http = response as? HTTPURLResponse else { throw AdminAPIError.invalidResponse }
        if http.statusCode == 401 { throw AdminAPIError.unauthorized }
        guard (200...299).contains(http.statusCode) else { throw AdminAPIError.server(http.statusCode) }
    }

    func deleteMessage(conversationId: String, messageId: String) async throws {
        let req = try makeRequest(
            path: "admin/chat/conversations/\(conversationId)/messages/\(messageId)",
            method: "DELETE"
        )
        let (_, response) = try await session.data(for: req)
        guard let http = response as? HTTPURLResponse else { throw AdminAPIError.invalidResponse }
        if http.statusCode == 401 { throw AdminAPIError.unauthorized }
        guard (200...299).contains(http.statusCode) else { throw AdminAPIError.server(http.statusCode) }
    }

    func registerDeviceToken(token: String, appVersion: String = "1.0.0") async throws {
        var req = try makeRequest(path: "admin/chat/device-token", method: "POST")
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        let payload: [String: String] = [
            "platform": "ios_admin",
            "device_token": token,
            "app_version": appVersion
        ]
        req.httpBody = try JSONSerialization.data(withJSONObject: payload)
        let (_, response) = try await session.data(for: req)
        guard let http = response as? HTTPURLResponse else { throw AdminAPIError.invalidResponse }
        guard (200...299).contains(http.statusCode) else { throw AdminAPIError.server(http.statusCode) }
    }

    func avatarURL(for value: String?) -> URL? {
        guard let value, !value.isEmpty else { return nil }
        if let absolute = URL(string: value), absolute.scheme != nil { return absolute }
        var cleaned = value.hasPrefix("/") ? String(value.dropFirst()) : value
        if !cleaned.hasPrefix("images/") && !cleaned.hasPrefix("avatars/") && !cleaned.hasPrefix("chat-uploads/") && !cleaned.hasPrefix("static/") {
            cleaned = "images/" + cleaned
        }
        let base = baseURL
        if let scheme = base.scheme, let host = base.host {
            let portString = base.port != nil ? ":\(base.port!)" : ""
            return URL(string: "\(scheme)://\(host)\(portString)/\(cleaned)")
        }
        return URL(string: "https://xn--b3cu8e7ah6h.com/\(cleaned)")
    }
}
