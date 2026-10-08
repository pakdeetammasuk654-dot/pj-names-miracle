import SwiftUI

struct RemoteAvatarView<Placeholder: View>: View {
    let url: URL?
    let placeholder: () -> Placeholder

    @State private var phase: AsyncImagePhase = .empty

    init(url: URL?, @ViewBuilder placeholder: @escaping () -> Placeholder) {
        self.url = url
        self.placeholder = placeholder
    }

    var body: some View {
        if let url = url {
            AsyncImage(url: url) { imagePhase in
                switch imagePhase {
                case .empty:
                    placeholder()
                case .success(let img):
                    img.resizable().scaledToFill()
                case .failure:
                    placeholder()
                @unknown default:
                    placeholder()
                }
            }
        } else {
            placeholder()
        }
    }
}

struct CustomerAvatarIcon: View {
    let avatarKey: String?
    let size: CGFloat

    var body: some View {
        let localName = localAssetName(avatarKey)
        let avatarURL = AdminAPIClient.shared.avatarURL(for: normalizedAvatarPath(avatarKey))

        RemoteAvatarView(url: avatarURL) {
            if let local = localName {
                Image(local)
                    .resizable()
                    .scaledToFill()
            } else {
                Circle()
                    .fill(Color.brandGold.opacity(0.3))
                    .overlay(
                        Image(systemName: "person.fill")
                            .font(.system(size: size * 0.5))
                            .foregroundStyle(Color.brandGold)
                    )
            }
        }
        .frame(width: size, height: size)
        .clipShape(Circle())
        .overlay(Circle().stroke(Color.brandGold.opacity(0.6), lineWidth: 1))
        .shadow(color: Color.black.opacity(0.3), radius: 2)
    }

    private func localAssetName(_ key: String?) -> String? {
        guard let k = key?.lowercased(), !k.isEmpty else { return "chat_avatar_cat" }
        if k.contains("dog") || (k.contains("male") && !k.contains("female")) { return "chat_avatar_dog" }
        if k.contains("rabbit") { return "chat_avatar_rabbit" }
        if k.contains("bear") { return "chat_avatar_bear" }
        if k.contains("panda") { return "chat_avatar_panda" }
        return nil
    }

    private func normalizedAvatarPath(_ key: String?) -> String {
        guard let k = key?.lowercased(), !k.isEmpty else { return "/images/chat_avatar_cat.png" }
        if k.contains("dog") || (k.contains("male") && !k.contains("female")) { return "/images/chat_avatar_dog.png" }
        if k.contains("rabbit") { return "/images/chat_avatar_rabbit.png" }
        if k.contains("bear") { return "/images/chat_avatar_bear.png" }
        if k.contains("panda") { return "/images/chat_avatar_panda.png" }
        if k.contains("taya") { return "/images/taya_avatar.jpg?v=child_hd" }
        if k.hasPrefix("/") { return k }
        return "/images/chat_avatar_cat.png"
    }
}
