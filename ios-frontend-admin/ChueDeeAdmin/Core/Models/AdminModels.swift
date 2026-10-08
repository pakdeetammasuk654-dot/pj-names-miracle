import Foundation

struct AdminConversation: Identifiable, Codable, Hashable {
    let id: String
    let customerToken: String?
    let customerName: String?
    let customerAvatar: String?
    let status: String
    let adminUnreadCount: Int
    let customerUnreadCount: Int
    let lastMessagePreview: String?
    let lastSender: String?
    let createdAt: String
    let updatedAt: String

    enum CodingKeys: String, CodingKey {
        case id
        case customerToken = "customer_token"
        case customerName = "customer_name"
        case customerAvatar = "customer_avatar"
        case status
        case adminUnreadCount = "admin_unread_count"
        case customerUnreadCount = "customer_unread_count"
        case lastMessagePreview = "last_message_preview"
        case lastSender = "last_sender"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    var displayName: String {
        let n = customerName?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        return n.isEmpty ? "ลูกค้า" : n
    }

    var customerChatLabel: String {
        let n = customerName?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        if n.isEmpty || n == "ลูกค้า" {
            return "ลูกค้า"
        }
        if n.hasSuffix("-ลูกค้า") {
            return n
        }
        return "\(n)-ลูกค้า"
    }

    var formattedTime: String {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        var date = formatter.date(from: updatedAt)
        if date == nil {
            formatter.formatOptions = [.withInternetDateTime]
            date = formatter.date(from: updatedAt)
        }
        guard let d = date else { return "" }
        let out = DateFormatter()
        out.locale = Locale(identifier: "th_TH")
        out.dateFormat = "dd MMM HH:mm"
        return out.string(from: d)
    }
}

struct AdminMessage: Identifiable, Codable, Hashable {
    let id: String
    let conversationId: String
    let sender: String
    let body: String?
    let imageUrl: String?
    let isRead: Bool?
    let createdAt: String

    enum CodingKeys: String, CodingKey {
        case id
        case conversationId = "conversation_id"
        case sender
        case body
        case imageUrl = "image_url"
        case isRead = "is_read"
        case createdAt = "created_at"
    }

    var isFromAdmin: Bool {
        sender.lowercased() == "admin"
    }

    var formattedTime: String {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        var date = formatter.date(from: createdAt)
        if date == nil {
            formatter.formatOptions = [.withInternetDateTime]
            date = formatter.date(from: createdAt)
        }
        guard let d = date else { return "" }
        let out = DateFormatter()
        out.locale = Locale(identifier: "th_TH")
        out.dateFormat = "HH:mm"
        return out.string(from: d)
    }
}

struct AdminInboxStats: Codable {
    let needsReply: Int
    let active: Int
    let closed: Int

    enum CodingKeys: String, CodingKey {
        case needsReply = "needs_reply"
        case active
        case closed
    }
}

struct AdminConversationsResponse: Codable {
    let inbox: String
    let conversations: [AdminConversation]
    let stats: AdminInboxStats
}

struct AdminConversationDetailResponse: Codable {
    let id: String
    let customerName: String?
    let customerAvatar: String?
    let status: String
    let messages: [AdminMessage]

    enum CodingKeys: String, CodingKey {
        case id
        case customerName = "customer_name"
        case customerAvatar = "customer_avatar"
        case status
        case messages
    }
}

struct AdminAuthResponse: Codable {
    let status: String
    let role: String
    let username: String
    let name: String
}
