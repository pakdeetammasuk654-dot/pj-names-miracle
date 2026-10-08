import SwiftUI

extension Color {
    // Modern Auspicious Luxury: espresso surfaces, restrained gold, and semantic status colors.
    nonisolated static let brandBackground = Color(hex: "#160D09")
    nonisolated static let brandGreen = Color(hex: "#2A1710")
    nonisolated static let brandGold = Color(hex: "#D4AF37")
    nonisolated static let brandCream = Color(hex: "#F5E7B6")
    nonisolated static let statusPositive = Color(hex: "#0E6B4D")
    nonisolated static let statusCaution = Color(hex: "#B93643")

    nonisolated init(hex: String) {
        let hex = hex.trimmingCharacters(in: CharacterSet.alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        let a, r, g, b: UInt64
        switch hex.count {
        case 3: // RGB (12-bit)
            (a, r, g, b) = (255, (int >> 8) * 17, (int >> 4 & 0xF) * 17, (int & 0xF) * 17)
        case 6: // RGB (24-bit)
            (a, r, g, b) = (255, int >> 16, int >> 8 & 0xFF, int & 0xFF)
        case 8: // ARGB (32-bit)
            (a, r, g, b) = (int >> 24, int >> 16 & 0xFF, int >> 8 & 0xFF, int & 0xFF)
        default:
            (a, r, g, b) = (255, 128, 128, 128)
        }
        self.init(
            .sRGB,
            red: Double(r) / 255,
            green: Double(g) / 255,
            blue: Double(b) / 255,
            opacity: Double(a) / 255
        )
    }
}

extension Font {
    nonisolated static func sarabun(_ size: CGFloat, weight: Weight = .regular) -> Font {
        let name: String
        switch weight {
        case .bold: name = "Sarabun-Bold"
        case .semibold: name = "Sarabun-SemiBold"
        case .medium: name = "Sarabun-Medium"
        default: name = "Sarabun-Regular"
        }
        return .custom(name, size: size)
    }

    nonisolated static func prompt(_ size: CGFloat, weight: Weight = .regular) -> Font {
        let name: String
        switch weight {
        case .bold: name = "Prompt-Bold"
        case .semibold: name = "Prompt-SemiBold"
        case .medium: name = "Prompt-Medium"
        default: name = "Prompt-Bold"
        }
        return .custom(name, size: size)
    }
}

struct LuxuryCard: ViewModifier {
    func body(content: Content) -> some View {
        content
            .padding(16)
            .background(Color.brandGreen.opacity(0.58))
            .clipShape(RoundedRectangle(cornerRadius: 22))
            .overlay(RoundedRectangle(cornerRadius: 22).stroke(Color.brandGold.opacity(0.48), lineWidth: 1.2))
    }
}

extension View {
    func luxuryCard() -> some View { modifier(LuxuryCard()) }
}

struct FloatingChatButton: View {
    @ObservedObject var chatViewModel: SupportChatViewModel = .shared
    let onTap: () -> Void

    var body: some View {
        Button(action: onTap) {
            ZStack(alignment: .topTrailing) {
                Image(systemName: "bubble.left.and.bubble.right.fill")
                    .font(.system(size: 20, weight: .bold))
                    .foregroundStyle(Color.brandGreen)
                    .frame(width: 48, height: 48)
                    .background(Color.brandGold)
                    .clipShape(Circle())
                    .shadow(color: Color.black.opacity(0.45), radius: 6, x: 0, y: 3)

                if chatViewModel.unreadCount > 0 {
                    Text(chatViewModel.unreadCount > 99 ? "99+" : "\(chatViewModel.unreadCount)")
                        .font(.system(size: 11, weight: .bold))
                        .foregroundColor(.white)
                        .padding(.horizontal, 5)
                        .padding(.vertical, 2)
                        .background(Color.red)
                        .clipShape(Capsule())
                        .overlay(Capsule().stroke(Color.brandGreen, lineWidth: 1.5))
                        .shadow(color: Color.black.opacity(0.35), radius: 3, x: 0, y: 1)
                        .offset(x: 4, y: -4)
                }
            }
        }
        .buttonStyle(.plain)
        .accessibilityLabel("ปรึกษาคุณทญา")
    }
}

extension View {
    func floatingChatOverlay(showsChatSheet: Binding<Bool>) -> some View {
        self
            .overlay(alignment: .bottomTrailing) {
                FloatingChatButton {
                    showsChatSheet.wrappedValue = true
                }
                .padding(.trailing, 16)
                .padding(.bottom, 20)
                .zIndex(12)
            }
            .fullScreenCover(isPresented: showsChatSheet) {
                SupportChatView()
                    .onDisappear {
                        SupportChatViewModel.shared.onDisappear()
                        Task {
                            await SupportChatViewModel.shared.checkUnread()
                        }
                    }
            }
    }
}
