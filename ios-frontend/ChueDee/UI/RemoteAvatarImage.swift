import SwiftUI

final class AvatarImageCache {
    static let shared = AvatarImageCache()
    private let cache = NSCache<NSURL, UIImage>()

    init() {
        cache.countLimit = 300
        cache.totalCostLimit = 60 * 1024 * 1024 // 60MB RAM Cache
    }

    func image(for url: URL) -> UIImage? {
        cache.object(forKey: url as NSURL)
    }

    func insert(_ image: UIImage, for url: URL) {
        cache.setObject(image, forKey: url as NSURL)
    }
}

@MainActor
final class AvatarImageLoader: ObservableObject {
    @Published var image: UIImage?
    @Published var isLoading = false
    private var currentURL: URL?

    func load(url: URL?) {
        guard let url = url else {
            self.image = nil
            return
        }
        if let cached = AvatarImageCache.shared.image(for: url) {
            self.image = cached
            self.isLoading = false
            return
        }
        if currentURL == url && isLoading { return }
        currentURL = url
        isLoading = true

        Task {
            do {
                var request = URLRequest(url: url)
                request.cachePolicy = .useProtocolCachePolicy
                let (data, response) = try await URLSession.shared.data(for: request)
                guard let http = response as? HTTPURLResponse, (200...299).contains(http.statusCode),
                      let uiImage = UIImage(data: data) else {
                    await MainActor.run { self.isLoading = false }
                    return
                }
                AvatarImageCache.shared.insert(uiImage, for: url)
                await MainActor.run {
                    if self.currentURL == url {
                        self.image = uiImage
                        self.isLoading = false
                    }
                }
            } catch {
                await MainActor.run {
                    if self.currentURL == url {
                        self.isLoading = false
                    }
                }
            }
        }
    }
}

struct RemoteAvatarImage<Fallback: View>: View {
    let url: URL?
    @ViewBuilder let fallback: () -> Fallback
    @StateObject private var loader = AvatarImageLoader()

    init(url: URL?, @ViewBuilder fallback: @escaping () -> Fallback) {
        self.url = url
        self.fallback = fallback
    }

    var body: some View {
        Group {
            if let url = url, let cached = AvatarImageCache.shared.image(for: url) {
                Image(uiImage: cached)
                    .resizable()
                    .scaledToFill()
            } else if let img = loader.image {
                Image(uiImage: img)
                    .resizable()
                    .scaledToFill()
            } else {
                fallback()
            }
        }
        .onAppear {
            loader.load(url: url)
        }
        .onChange(of: url) { _, newURL in
            loader.load(url: newURL)
        }
    }
}
