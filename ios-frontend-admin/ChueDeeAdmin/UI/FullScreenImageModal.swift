import SwiftUI

struct FullScreenImageModal: View {
    let url: URL
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        ZStack {
            Color.black.ignoresSafeArea()

            AsyncImage(url: url) { phase in
                switch phase {
                case .empty:
                    ProgressView().tint(.white)
                case .success(let image):
                    image
                        .resizable()
                        .scaledToFit()
                case .failure:
                    VStack(spacing: 8) {
                        Image(systemName: "exclamationmark.triangle")
                            .font(.system(size: 36))
                            .foregroundStyle(.white.opacity(0.7))
                        Text("ไม่สามารถโหลดรูปภาพได้")
                            .font(.sarabun(14))
                            .foregroundStyle(.white.opacity(0.7))
                    }
                @unknown default:
                    EmptyView()
                }
            }

            VStack {
                HStack {
                    Spacer()
                    Button(action: { dismiss() }) {
                        Image(systemName: "xmark.circle.fill")
                            .font(.system(size: 30))
                            .foregroundStyle(.white.opacity(0.85))
                            .padding()
                    }
                }
                Spacer()
            }
        }
    }
}
