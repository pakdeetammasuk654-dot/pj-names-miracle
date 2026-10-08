import SwiftUI

struct AdminMessageBubble: View {
    let message: AdminMessage
    let customerAvatarKey: String?
    var customerName: String? = nil
    let onImageTap: (URL) -> Void

    private var customerSenderTitle: String {
        let name = customerName?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        if name.isEmpty || name == "ลูกค้า" {
            return "ลูกค้า"
        }
        if name.hasSuffix("-ลูกค้า") {
            return name
        }
        return "\(name)-ลูกค้า"
    }

    var body: some View {
        let isMine = message.isFromAdmin

        HStack(alignment: .top, spacing: 8) {
            if !isMine {
                CustomerAvatarIcon(avatarKey: customerAvatarKey, size: 36)
            } else {
                Spacer(minLength: 40)
            }

            VStack(alignment: isMine ? .trailing : .leading, spacing: 4) {
                HStack(spacing: 6) {
                    if isMine && (message.isRead == true) {
                        Text("อ่านแล้ว")
                            .font(.sarabun(10))
                            .foregroundStyle(Color.brandCream.opacity(0.55))
                    }
                    Text(isMine ? "อ.ทญา ว่องไว" : customerSenderTitle)
                        .font(.sarabun(12, weight: .semibold))
                        .foregroundStyle(Color.brandGoldLight)
                    Text("· \(message.formattedTime)")
                        .font(.sarabun(11))
                        .foregroundStyle(Color.brandCream.opacity(0.5))
                }

                VStack(alignment: .leading, spacing: 8) {
                    if let body = message.body, !body.isEmpty {
                        Text(body)
                            .font(.sarabun(15))
                            .foregroundStyle(Color.white)
                            .textSelection(.enabled)
                    }

                    if let imgStr = message.imageUrl, !imgStr.isEmpty,
                       let imgURL = AdminAPIClient.shared.avatarURL(for: imgStr) {
                        Button(action: {
                            onImageTap(imgURL)
                        }) {
                            AsyncImage(url: imgURL) { phase in
                                switch phase {
                                case .empty:
                                    Rectangle()
                                        .fill(Color.brandGreen.opacity(0.4))
                                        .frame(height: 180)
                                        .overlay(ProgressView().tint(Color.brandGold))
                                case .success(let image):
                                    image
                                        .resizable()
                                        .scaledToFit()
                                        .frame(maxHeight: 280)
                                        .clipShape(RoundedRectangle(cornerRadius: 8))
                                case .failure:
                                    HStack {
                                        Image(systemName: "photo.badge.exclamationmark")
                                        Text("โหลดรูปภาพไม่สำเร็จ")
                                    }
                                    .font(.sarabun(13))
                                    .foregroundStyle(Color.brandCream.opacity(0.7))
                                    .padding()
                                @unknown default:
                                    EmptyView()
                                }
                            }
                        }
                        .buttonStyle(.plain)
                    }
                }
                .padding(.horizontal, 14)
                .padding(.vertical, 10)
                .background(isMine ? Color.brandBubbleAdmin : Color.brandBubbleCustomer)
                .clipShape(RoundedRectangle(cornerRadius: 14))
                .overlay(
                    RoundedRectangle(cornerRadius: 14)
                        .stroke(isMine ? Color.brandGold.opacity(0.4) : Color.brandGold.opacity(0.2), lineWidth: 1)
                )
            }

            if isMine {
                let tayaAvatarURL = AdminAPIClient.shared.avatarURL(for: "/images/taya_avatar.jpg?v=child_hd")
                RemoteAvatarView(url: tayaAvatarURL) {
                    Circle().fill(Color.brandGold.opacity(0.4))
                }
                .frame(width: 36, height: 36)
                .clipShape(Circle())
                .overlay(Circle().stroke(Color.brandGold.opacity(0.7), lineWidth: 1.2))
                .shadow(color: Color.black.opacity(0.3), radius: 2)
            } else {
                Spacer(minLength: 40)
            }
        }
    }
}
