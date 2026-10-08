import SwiftUI

struct ThreadRowView: View {
    let conversation: AdminConversation

    var body: some View {
        HStack(spacing: 12) {
            CustomerAvatarIcon(avatarKey: conversation.customerAvatar, size: 48)

            VStack(alignment: .leading, spacing: 4) {
                HStack {
                    Text(conversation.displayName)
                        .font(.sarabun(16, weight: .bold))
                        .foregroundStyle(Color.brandGoldLight)

                    if conversation.adminUnreadCount > 0 {
                        Text("รอตอบ \(conversation.adminUnreadCount)")
                            .font(.sarabun(11, weight: .bold))
                            .foregroundStyle(Color.black)
                            .padding(.horizontal, 6)
                            .padding(.vertical, 2)
                            .background(Color.brandGold)
                            .clipShape(Capsule())
                    }

                    Spacer()

                    Text(conversation.formattedTime)
                        .font(.sarabun(12))
                        .foregroundStyle(Color.brandCream.opacity(0.6))
                }

                Text(conversation.lastMessagePreview?.isEmpty == false ? conversation.lastMessagePreview! : "เริ่มการสนทนา")
                    .font(.sarabun(14))
                    .foregroundStyle(Color.brandCream.opacity(0.85))
                    .lineLimit(2)
            }
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
        .background(conversation.adminUnreadCount > 0 ? Color.brandCardHover : Color.brandCard)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .overlay(
            RoundedRectangle(cornerRadius: 14)
                .stroke(conversation.adminUnreadCount > 0 ? Color.brandGold.opacity(0.5) : Color.brandGold.opacity(0.15), lineWidth: 1)
        )
    }
}
