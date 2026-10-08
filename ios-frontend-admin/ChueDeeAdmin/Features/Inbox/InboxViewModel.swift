import SwiftUI
import Combine

enum InboxTab: String, CaseIterable, Identifiable {
    case needsReply = "needs_reply"
    case active = "active"
    case closed = "closed"

    var id: String { rawValue }

    var title: String {
        switch self {
        case .needsReply: return "รอตอบ"
        case .active: return "กำลังคุย"
        case .closed: return "ปิดแล้ว"
        }
    }
}

@MainActor
final class InboxViewModel: ObservableObject {
    @Published var selectedTab: InboxTab = .needsReply
    @Published var conversations: [AdminConversation] = []
    @Published var stats: AdminInboxStats = AdminInboxStats(needsReply: 0, active: 0, closed: 0)
    @Published var isLoading: Bool = false
    @Published var errorMessage: String?
    @Published var selectedConversationId: String?

    private var pollTask: Task<Void, Never>?
    private let apiClient = AdminAPIClient.shared

    init() {
        startPolling()
    }

    func load() async {
        do {
            let res = try await apiClient.getConversations(inbox: selectedTab.rawValue)
            self.conversations = res.conversations
            self.stats = res.stats
            AdminBadgeManager.setBadge(res.stats.needsReply)
            self.errorMessage = nil
        } catch let err as AdminAPIError {
            if case .unauthorized = err {
                AuthViewModel.shared.logout()
            } else {
                self.errorMessage = err.localizedDescription
            }
        } catch {
            self.errorMessage = "ไม่สามารถโหลดข้อมูลได้"
        }
    }

    func changeTab(_ tab: InboxTab) {
        selectedTab = tab
        Task {
            isLoading = true
            await load()
            isLoading = false
        }
    }

    func startPolling() {
        stopPolling()
        pollTask = Task { [weak self] in
            while !Task.isCancelled {
                try? await Task.sleep(for: .seconds(4))
                guard !Task.isCancelled, let self = self else { break }
                await self.load()
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
