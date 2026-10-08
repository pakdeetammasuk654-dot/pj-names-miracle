import Foundation
import SwiftUI
import Combine

struct ChatAvatarOption: Identifiable, Hashable {
    let id: String
    let label: String
    let imagePath: String

    static let all: [ChatAvatarOption] = [
        ChatAvatarOption(id: "cat", label: "แมว", imagePath: "/images/chat_avatar_cat.png"),
        ChatAvatarOption(id: "dog", label: "สุนัข", imagePath: "/images/chat_avatar_dog.png"),
        ChatAvatarOption(id: "rabbit", label: "กระต่าย", imagePath: "/images/chat_avatar_rabbit.png"),
        ChatAvatarOption(id: "bear", label: "หมี", imagePath: "/images/chat_avatar_bear.png"),
        ChatAvatarOption(id: "panda", label: "แพนด้า", imagePath: "/images/chat_avatar_panda.png")
    ]

    static func path(for key: String) -> String {
        let clean = key.lowercased().trimmingCharacters(in: .whitespacesAndNewlines)
        switch clean {
        case "cat": return "/images/chat_avatar_cat.png"
        case "dog": return "/images/chat_avatar_dog.png"
        case "rabbit": return "/images/chat_avatar_rabbit.png"
        case "bear": return "/images/chat_avatar_bear.png"
        case "panda": return "/images/chat_avatar_panda.png"
        case "male": return "/images/chat_avatar_dog.png"
        case "female": return "/images/chat_avatar_cat.png"
        default:
            if clean.hasPrefix("/images/") { return clean }
            return "/images/chat_avatar_cat.png"
        }
    }

    static func key(for keyOrPath: String) -> String {
        let clean = keyOrPath.lowercased()
        if clean.contains("dog") || (clean.contains("male") && !clean.contains("female")) { return "dog" }
        if clean.contains("rabbit") { return "rabbit" }
        if clean.contains("bear") { return "bear" }
        if clean.contains("panda") { return "panda" }
        return "cat"
    }

    static func localAssetName(for keyOrPath: String?) -> String? {
        guard let s = keyOrPath?.lowercased(), !s.isEmpty else { return "chat_avatar_cat" }
        if s.contains("app_logo") { return "AppLogo" }
        if s.contains("cat") { return "chat_avatar_cat" }
        if s.contains("dog") || (s.contains("male") && !s.contains("female")) { return "chat_avatar_dog" }
        if s.contains("rabbit") { return "chat_avatar_rabbit" }
        if s.contains("bear") { return "chat_avatar_bear" }
        if s.contains("panda") { return "chat_avatar_panda" }
        return "chat_avatar_cat"
    }
}

@MainActor
final class SupportChatViewModel: ObservableObject {
    static let shared = SupportChatViewModel()

    @Published var conversationId: String = ""
    @Published var token: String = ""
    @Published var customerAvatar: String = "cat"
    @Published var customerName: String = ""
    @Published var messages: [SupportChatMessage] = []
    @Published var unreadCount: Int = 0
    @Published var isLoading: Bool = false
    @Published var isDeleting: Bool = false
    @Published var isSending: Bool = false
    @Published var errorMessage: String?
    @Published var openChatRequested: Bool = false
    @Published var apnsDeviceToken: String = ""
    @Published var selectedImageData: Data?
    @Published var selectedImageMimeType: String?

    func attachImage(data: Data?, mimeType: String? = "image/jpeg") {
        self.selectedImageData = data
        self.selectedImageMimeType = mimeType
    }

    func clearAttachedImage() {
        self.selectedImageData = nil
        self.selectedImageMimeType = nil
    }

    private var pollTask: Task<Void, Never>?
    private let apiClient: APIClient
    private let keychain: KeychainHelper

    var hasConversation: Bool {
        !conversationId.isEmpty && !token.isEmpty
    }

    var customerAvatarPath: String {
        ChatAvatarOption.path(for: customerAvatar)
    }

    init(apiClient: APIClient = .shared, keychain: KeychainHelper = .shared) {
        self.apiClient = apiClient
        self.keychain = keychain

        self.conversationId = keychain.get(key: "conversation_id") ?? ""
        self.token = keychain.get(key: "customer_token") ?? ""
        self.customerAvatar = keychain.get(key: "customer_avatar") ?? "cat"
        self.apnsDeviceToken = keychain.get(key: "apns_device_token") ?? ""
    }

    func setDeviceToken(_ token: String) {
        let clean = token.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !clean.isEmpty else { return }
        self.apnsDeviceToken = clean
        keychain.save(key: "apns_device_token", value: clean)
        if hasConversation {
            Task {
                await registerDeviceToken(clean)
            }
        }
    }

    func onAppear() {
        self.unreadCount = 0
        BadgeManager.clearBadge()
        if hasConversation {
            Task {
                if !apnsDeviceToken.isEmpty {
                    await registerDeviceToken(apnsDeviceToken)
                }
                await refresh()
                startPolling()
            }
        }
    }

    func onDisappear() {
        stopPolling()
    }

    func selectAvatar(_ avatarKey: String) {
        let key = ChatAvatarOption.key(for: avatarKey)
        self.customerAvatar = key
        keychain.save(key: "customer_avatar", value: key)
    }

    func startNewConversation() {
        stopPolling()
        keychain.clearAllChatSession()
        self.conversationId = ""
        self.token = ""
        self.messages = []
        self.unreadCount = 0
        self.errorMessage = nil
        BadgeManager.clearBadge()
    }

    func deleteConversation() {
        guard hasConversation, !isDeleting else { return }
        let id = conversationId
        let chatToken = token
        isDeleting = true
        errorMessage = nil
        Task {
            do {
                try await apiClient.deleteChatConversation(conversationId: id, token: chatToken)
                startNewConversation()
            } catch {
                errorMessage = "ลบข้อมูลแชตไม่สำเร็จ กรุณาลองใหม่"
            }
            isDeleting = false
        }
    }

    func checkUnread() async {
        guard hasConversation else {
            if unreadCount != 0 { unreadCount = 0 }
            BadgeManager.clearBadge()
            return
        }
        do {
            let res = try await apiClient.getChatUnreadCount(conversationId: conversationId, token: token)
            self.unreadCount = res.unreadCount
            if res.unreadCount == 0 {
                BadgeManager.clearBadge()
            } else {
                BadgeManager.setBadge(res.unreadCount)
            }
        } catch {
            // silent fail on network glitch
        }
    }

    func startConversation(name: String) async {
        guard !isLoading else { return }
        isLoading = true
        errorMessage = nil

        let chosenKey = ChatAvatarOption.key(for: customerAvatar)

        do {
            let res = try await apiClient.createChatConversation(
                customerName: name.trimmingCharacters(in: .whitespacesAndNewlines),
                customerAvatar: chosenKey
            )
            self.conversationId = res.conversationID
            self.token = res.customerToken
            let serverAvatar = res.customerAvatar ?? chosenKey
            self.customerAvatar = serverAvatar
            self.unreadCount = 0

            keychain.save(key: "conversation_id", value: res.conversationID)
            keychain.save(key: "customer_token", value: res.customerToken)
            keychain.save(key: "customer_avatar", value: serverAvatar)

            self.isLoading = false
            if !self.apnsDeviceToken.isEmpty {
                await self.registerDeviceToken(self.apnsDeviceToken)
            }
            await refresh()
            startPolling()
        } catch {
            self.isLoading = false
            self.errorMessage = "ไม่สามารถเริ่มแชตได้ กรุณาลองใหม่"
        }
    }

    func refresh() async {
        guard hasConversation else { return }
        do {
            let res = try await apiClient.getChatMessages(conversationId: conversationId, token: token)
            self.messages = res.messages
            self.unreadCount = 0
            BadgeManager.clearBadge()
            if let avatar = res.customerAvatar, !avatar.isEmpty {
                let key = ChatAvatarOption.key(for: avatar)
                self.customerAvatar = key
                keychain.save(key: "customer_avatar", value: key)
            }
            if let name = res.customerName {
                self.customerName = name
            }
            self.errorMessage = nil
        } catch {
            // silent fail on polling or network glitch
        }
    }

    func sendCurrentMessage(body: String) async {
        let img = selectedImageData
        let mime = selectedImageMimeType
        clearAttachedImage()
        await sendMessage(body: body, imageData: img, mimeType: mime)
    }

    func sendMessage(body: String, imageData: Data?, mimeType: String?) async {
        let trimmed = body.trimmingCharacters(in: .whitespacesAndNewlines)
        guard hasConversation, (!trimmed.isEmpty || imageData != nil), !isSending else { return }

        isSending = true
        errorMessage = nil

        do {
            try await apiClient.sendChatMessage(
                conversationId: conversationId,
                token: token,
                body: trimmed,
                imageData: imageData,
                mimeType: mimeType
            )
            self.isSending = false
            self.unreadCount = 0
            await refresh()
        } catch let err as NSError {
            self.isSending = false
            self.errorMessage = err.localizedDescription.isEmpty ? "ส่งข้อความไม่สำเร็จ" : err.localizedDescription
        } catch {
            self.isSending = false
            self.errorMessage = "ส่งข้อความไม่สำเร็จ"
        }
    }

    func registerDeviceToken(_ deviceToken: String) async {
        guard hasConversation, !deviceToken.isEmpty else { return }
        try? await apiClient.registerChatDeviceToken(
            conversationId: conversationId,
            token: token,
            deviceToken: deviceToken,
            appVersion: "1.3.1"
        )
    }

    private func startPolling() {
        stopPolling()
        pollTask = Task { [weak self] in
            while !Task.isCancelled {
                try? await Task.sleep(for: .seconds(15))
                guard !Task.isCancelled, let self = self else { break }
                await self.refresh()
            }
        }
    }

    private func stopPolling() {
        pollTask?.cancel()
        pollTask = nil
    }

    deinit {
        pollTask?.cancel()
    }
}
