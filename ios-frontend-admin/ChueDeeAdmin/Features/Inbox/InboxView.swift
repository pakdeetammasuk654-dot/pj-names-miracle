import SwiftUI

struct InboxView: View {
    @StateObject private var viewModel = InboxViewModel()
    @ObservedObject var authVM: AuthViewModel = .shared
    @State private var navPath = NavigationPath()

    var body: some View {
        NavigationStack(path: $navPath) {
            ZStack {
                Color.brandBackground.ignoresSafeArea()

                VStack(spacing: 0) {
                    // Filter Tabs
                    HStack(spacing: 8) {
                        ForEach(InboxTab.allCases) { tab in
                            let isSelected = viewModel.selectedTab == tab
                            let count: Int = {
                                switch tab {
                                case .needsReply: return viewModel.stats.needsReply
                                case .active: return viewModel.stats.active
                                case .closed: return viewModel.stats.closed
                                }
                            }()

                            Button(action: {
                                viewModel.changeTab(tab)
                            }) {
                                HStack(spacing: 6) {
                                    Text(tab.title)
                                        .font(.sarabun(14, weight: isSelected ? .bold : .medium))
                                    if count > 0 {
                                        Text("\(count)")
                                            .font(.sarabun(11, weight: .bold))
                                            .padding(.horizontal, 6)
                                            .padding(.vertical, 2)
                                            .background(isSelected ? Color.black.opacity(0.25) : Color.white.opacity(0.15))
                                            .clipShape(Capsule())
                                    }
                                }
                                .padding(.vertical, 8)
                                .padding(.horizontal, 14)
                                .background(isSelected ? Color.brandGold : Color.brandGreen.opacity(0.4))
                                .foregroundStyle(isSelected ? Color.black : Color.brandCream.opacity(0.8))
                                .clipShape(RoundedRectangle(cornerRadius: 10))
                                .overlay(
                                    RoundedRectangle(cornerRadius: 10)
                                        .stroke(Color.brandGold.opacity(isSelected ? 0.8 : 0.2), lineWidth: 1)
                                )
                            }
                            .buttonStyle(.plain)
                        }
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 12)
                    .background(Color.brandGreen.opacity(0.2))

                    // Conversation List
                    if viewModel.conversations.isEmpty {
                        VStack(spacing: 12) {
                            Spacer()
                            Image(systemName: "tray")
                                .font(.system(size: 48))
                                .foregroundStyle(Color.brandGold.opacity(0.5))
                            Text("ไม่มีบทสนทนาในหมวดนี้")
                                .font(.sarabun(16))
                                .foregroundStyle(Color.brandCream.opacity(0.7))
                            Spacer()
                        }
                    } else {
                        ScrollView {
                            LazyVStack(spacing: 10) {
                                ForEach(viewModel.conversations) { item in
                                    NavigationLink(value: item) {
                                        ThreadRowView(conversation: item)
                                    }
                                    .buttonStyle(.plain)
                                }
                            }
                            .padding(.horizontal, 16)
                            .padding(.vertical, 12)
                        }
                        .refreshable {
                            await viewModel.load()
                        }
                    }
                }
            }
            .navigationTitle("แชตลูกค้า — อ.ทญา")
            .navigationBarTitleDisplayMode(.inline)
            .toolbarBackground(Color.brandGreen.opacity(0.95), for: .navigationBar)
            .toolbarBackground(.visible, for: .navigationBar)
            .toolbarColorScheme(.dark, for: .navigationBar)
            .toolbar {
                ToolbarItem(placement: .topBarLeading) {
                    Image("AppLogo")
                        .resizable()
                        .scaledToFit()
                        .frame(width: 32, height: 32)
                        .clipShape(RoundedRectangle(cornerRadius: 7))
                }

                ToolbarItem(placement: .topBarTrailing) {
                    Button(action: {
                        authVM.logout()
                    }) {
                        Image(systemName: "rectangle.portrait.and.arrow.right")
                            .font(.system(size: 14, weight: .semibold))
                            .foregroundStyle(Color.brandCream.opacity(0.8))
                    }
                }
            }
            .navigationDestination(for: AdminConversation.self) { conv in
                AdminChatView(conversation: conv)
            }
            .onReceive(NotificationCenter.default.publisher(for: NSNotification.Name("OpenAdminConversation"))) { notif in
                if let convID = notif.userInfo?["conversation_id"] as? String {
                    Task {
                        await viewModel.load()
                        if let found = viewModel.conversations.first(where: { $0.id == convID }) {
                            navPath.append(found)
                        } else if let res = try? await AdminAPIClient.shared.getMessages(conversationId: convID) {
                            let conv = AdminConversation(
                                id: res.id,
                                customerToken: nil,
                                customerName: res.customerName,
                                customerAvatar: res.customerAvatar,
                                status: res.status,
                                adminUnreadCount: 0,
                                customerUnreadCount: 0,
                                lastMessagePreview: res.messages.last?.body,
                                lastSender: res.messages.last?.sender,
                                createdAt: "",
                                updatedAt: ""
                            )
                            navPath.append(conv)
                        }
                    }
                }
            }
            .task {
                await viewModel.load()
            }
        }
    }
}
