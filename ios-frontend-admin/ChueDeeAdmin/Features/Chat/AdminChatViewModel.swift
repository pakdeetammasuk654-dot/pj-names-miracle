import SwiftUI
import Combine
import PhotosUI

@MainActor
final class AdminChatViewModel: ObservableObject {
    @Published var conversation: AdminConversation
    @Published var messages: [AdminMessage] = []
    @Published var isLoading: Bool = false
    @Published var isSending: Bool = false
    @Published var deletingMessageID: String?
    @Published var errorMessage: String?

    private var pollTask: Task<Void, Never>?
    private let apiClient = AdminAPIClient.shared

    init(conversation: AdminConversation) {
        self.conversation = conversation
    }

    func loadMessages() async {
        do {
            let res = try await apiClient.getMessages(conversationId: conversation.id)
            self.messages = res.messages
            let updatedName = (res.customerName?.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty == false) ? res.customerName : self.conversation.customerName
            let updatedAvatar = (res.customerAvatar?.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty == false) ? res.customerAvatar : self.conversation.customerAvatar
            self.conversation = AdminConversation(
                id: res.id,
                customerToken: self.conversation.customerToken,
                customerName: updatedName,
                customerAvatar: updatedAvatar,
                status: res.status.isEmpty ? self.conversation.status : res.status,
                adminUnreadCount: self.conversation.adminUnreadCount,
                customerUnreadCount: self.conversation.customerUnreadCount,
                lastMessagePreview: self.conversation.lastMessagePreview,
                lastSender: self.conversation.lastSender,
                createdAt: self.conversation.createdAt,
                updatedAt: self.conversation.updatedAt
            )
            self.errorMessage = nil
        } catch let err as AdminAPIError {
            self.errorMessage = err.localizedDescription
        } catch {
            self.errorMessage = "ไม่สามารถโหลดข้อความได้"
        }
    }

    func sendMessage(body: String, imageData: Data?, mimeType: String?) async {
        let trimmed = body.trimmingCharacters(in: .whitespacesAndNewlines)
        guard (!trimmed.isEmpty || imageData != nil), !isSending else { return }

        isSending = true
        errorMessage = nil

        do {
            try await apiClient.sendMessage(
                conversationId: conversation.id,
                body: trimmed,
                imageData: imageData,
                mimeType: mimeType
            )
            self.isSending = false
            await loadMessages()
        } catch let err as AdminAPIError {
            self.isSending = false
            self.errorMessage = err.localizedDescription
        } catch {
            self.isSending = false
            self.errorMessage = "ส่งข้อความไม่สำเร็จ"
        }
    }

    func toggleCaseStatus() async {
        do {
            if conversation.status == "open" {
                try await apiClient.closeConversation(conversationId: conversation.id)
            }
            await loadMessages()
        } catch {
            self.errorMessage = "ไม่สามารถเปลี่ยนสถานะเคสได้"
        }
    }

    func deleteMessage(_ message: AdminMessage) async {
        guard deletingMessageID == nil else { return }
        deletingMessageID = message.id
        errorMessage = nil
        do {
            try await apiClient.deleteMessage(conversationId: conversation.id, messageId: message.id)
            messages.removeAll { $0.id == message.id }
        } catch let err as AdminAPIError {
            errorMessage = err.localizedDescription
        } catch {
            errorMessage = "ไม่สามารถลบข้อความได้"
        }
        deletingMessageID = nil
    }

    func startPolling() {
        stopPolling()
        pollTask = Task { [weak self] in
            while !Task.isCancelled {
                try? await Task.sleep(for: .seconds(3))
                guard !Task.isCancelled, let self = self else { break }
                await self.loadMessages()
            }
        }
    }

    func stopPolling() {
        pollTask?.cancel()
        pollTask = nil
    }

    deinit {
        pollTask?.cancel()
    }
}
