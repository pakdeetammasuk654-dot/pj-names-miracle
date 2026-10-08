import SwiftUI
import UIKit
import Photos

struct BlessingCardView: View {
    let item: NameResultItem
    @ObservedObject var viewModel: SearchViewModel
    @State private var sharedImage: SharedBlessingImage?
    @State private var isPreparingShare = false
    @State private var showsExportOptions = false
    @State private var exportNotice = ""
    @State private var selectedTheme: WallpaperTheme = .emerald
    @State private var backgroundImage: UIImage?
    @State private var buddhaImage: UIImage?
    @State private var frogImage: UIImage?
    @State private var goldImage: UIImage?
    @State private var showsChatSheet = false
    @StateObject private var purchaseManager = PurchaseManager.shared

    var body: some View {
        ZStack {
            Color.brandBackground.ignoresSafeArea()
            ScrollView {
                VStack(spacing: 14) {
                    if !purchaseManager.effectiveIsUnlocked {
                        HStack(spacing: 8) {
                            Image(systemName: "lock.fill")
                            Text("ตัวอย่างการ์ด · เปิดชื่อทั้งหมดเพื่อสร้างแรงดึงดูด")
                                .font(.sarabun(11, weight: .semibold))
                            Spacer(minLength: 0)
                            Button {
                                Task { await purchaseManager.purchase() }
                            } label: {
                                if purchaseManager.isBusy {
                                    ProgressView().tint(Color.brandGreen)
                                } else {
                                    Text("เปิดชื่อ")
                                }
                            }
                            .font(.sarabun(11, weight: .bold))
                            .disabled(purchaseManager.isBusy)
                        }
                        .foregroundStyle(Color.brandGreen)
                        .padding(.horizontal, 13).padding(.vertical, 10)
                        .background(Color.brandGold)
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                    }

                    // Theme selector pills
                    HStack(spacing: 8) {
                        ForEach(WallpaperTheme.allCases) { theme in
                            let isSelected = theme == selectedTheme
                            Button {
                                selectedTheme = theme
                                Task { await loadBackgroundImage(for: theme) }
                            } label: {
                                Text(theme.title)
                                    .font(.sarabun(12, weight: isSelected ? .bold : .medium))
                                    .foregroundStyle(isSelected ? Color(red: 0.11, green: 0.08, blue: 0.02) : Color.brandGold)
                                    .frame(maxWidth: .infinity)
                                    .padding(.vertical, 7)
                                    .background(isSelected ? Color.brandGold : Color.brandGreen.opacity(0.6))
                                    .clipShape(RoundedRectangle(cornerRadius: 16))
                                    .overlay(
                                        RoundedRectangle(cornerRadius: 16)
                                            .stroke(isSelected ? Color.white.opacity(0.8) : Color.brandGold.opacity(0.35), lineWidth: 1)
                                    )
                            }
                            .buttonStyle(.plain)
                        }
                    }

                    GeometryReader { proxy in
                        blessingCard(width: proxy.size.width)
                    }
                    .aspectRatio(9.0 / 19.5, contentMode: .fit)

                    Button { showsExportOptions = true } label: {
                        HStack(spacing: 12) {
                            if isPreparingShare {
                                ProgressView().tint(.black)
                            } else {
                                Image(systemName: "square.and.arrow.up")
                                    .font(.system(size: 21, weight: .bold))
                            }
                            Text(isPreparingShare ? "กำลังสร้างภาพ…" : "บันทึกภาพ/แชร์")
                                .font(.sarabun(17, weight: .bold))
                        }
                        .foregroundStyle(.black)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 15)
                        .background(
                            LinearGradient(
                                colors: [Color(red: 0.89, green: 0.69, blue: 0.16), Color(red: 1.0, green: 0.65, blue: 0.04)],
                                startPoint: .leading,
                                endPoint: .trailing
                            )
                        )
                        .clipShape(RoundedRectangle(cornerRadius: 17))
                    }
                    .buttonStyle(.plain)
                    .disabled(isPreparingShare)
                    .accessibilityHint("สร้างภาพการ์ดมงคลเพื่อบันทึกหรือแชร์")
                    .padding(.top, 8)
                    .padding(.bottom, 24)
                }
                .padding(16)
            }
        }
        .navigationTitle("ได้ชื่อดีเป็นมงคล")
        .navigationBarTitleDisplayMode(.inline)
        .toolbarBackground(Color.brandGreen, for: .navigationBar)
        .toolbarBackground(.visible, for: .navigationBar)
        .toolbar {
            ToolbarItem(placement: .principal) {
                HStack(spacing: 8) {
                    Image("AppLogo")
                        .resizable().scaledToFit()
                        .frame(width: 30, height: 30)
                        .clipShape(Circle())
                        .overlay(Circle().stroke(Color.brandGold.opacity(0.65)))
                    Text("ได้ชื่อดีเป็นมงคล")
                        .font(.sarabun(18, weight: .bold))
                        .foregroundStyle(Color.brandCream)
                }
            }
        }
        .preferredColorScheme(.dark)
        .task { await loadDecorationImages() }
        .sheet(item: $sharedImage) { payload in
            ActivityShareSheet(items: [payload.image])
        }
        .confirmationDialog("เลือกวิธีใช้การ์ดมงคล", isPresented: $showsExportOptions, titleVisibility: .visible) {
            Button("บันทึกเป็นรูปภาพ") { saveToPhotos() }
            Button("แชร์ไปยังโซเชียล") { shareToSocial() }
            Button("ยกเลิก", role: .cancel) {}
        } message: {
            Text("คุณสามารถเก็บภาพไว้ใน Photos หรือแชร์ผ่านแอปที่ติดตั้งอยู่")
        }
        .alert("การ์ดมงคล", isPresented: Binding(
            get: { !exportNotice.isEmpty },
            set: { if !$0 { exportNotice = "" } }
        )) {
            Button("ตกลง", role: .cancel) { exportNotice = "" }
        } message: {
            Text(exportNotice)
        }
        .alert("การชำระเงิน", isPresented: Binding(
            get: { purchaseManager.message != nil },
            set: { if !$0 { purchaseManager.message = nil } }
        )) {
            Button("ตกลง", role: .cancel) { purchaseManager.message = nil }
        } message: {
            Text(purchaseManager.message ?? "")
        }
        .floatingChatOverlay(showsChatSheet: $showsChatSheet)
    }

    private func blessingCard(width: CGFloat) -> some View {
        let height = width * 19.5 / 9.0
        return ZStack(alignment: .top) {
            if let backgroundImage {
                Image(uiImage: backgroundImage)
                    .resizable()
                    .scaledToFill()
                    .frame(width: width, height: height)
                    .clipped()
            } else {
                LinearGradient(
                    colors: [
                        Color(red: 1.0, green: 0.99, blue: 0.94),
                        Color(red: 0.98, green: 0.94, blue: 0.86),
                        Color(red: 0.94, green: 0.87, blue: 0.74)
                    ],
                    startPoint: .top,
                    endPoint: .bottom
                )
            }

            sparkleLayer

            let isTuesday = viewModel.dayOfBirth.lowercased() == "tuesday"
            let isStandingBuddha = ["sunday", "monday", "friday"].contains(viewModel.dayOfBirth.lowercased())
            let tuesdayBoxSize = width * 0.90
            let buddhaWidth = isTuesday ? tuesdayBoxSize : (isStandingBuddha ? width * 0.54 : width * 0.52)
            let buddhaHeight = isTuesday ? tuesdayBoxSize : (isStandingBuddha ? height * 0.28 : height * 0.27)
            let buddhaTopPadding = isTuesday ? (height * 0.32 - (tuesdayBoxSize * 0.655)) : (isStandingBuddha ? height * 0.035 : height * 0.05)

            Group {
                if let buddhaImage {
                    Image(uiImage: buddhaImage)
                        .resizable()
                        .scaledToFit()
                        .frame(
                            width: buddhaWidth,
                            height: buddhaHeight
                        )
                } else {
                    Circle()
                        .fill(Color.white.opacity(0.22))
                        .overlay(Image(systemName: "sparkles").foregroundStyle(Color.brandGold))
                        .frame(width: buddhaWidth, height: buddhaHeight)
                }
            }
            .frame(
                width: buddhaWidth,
                height: buddhaHeight
            )
            .padding(.top, buddhaTopPadding)

            VStack(spacing: 0) {
                Text(fullName)
                    .font(.sarabun(nameFontSize, weight: .bold))
                    .foregroundStyle(selectedTheme.nameTextColor)
                    .lineLimit(1)
                    .minimumScaleFactor(0.6)

                Text(item.meaning?.nonEmpty ?? "ยังไม่มีข้อมูลความหมายของชื่อนี้ในฐานข้อมูล")
                    .font(.sarabun(13, weight: .medium))
                    .foregroundStyle(selectedTheme.meaningTextColor)
                    .multilineTextAlignment(.center)
                    .lineLimit(3)
                    .padding(.top, 7)

                cardDivider

                Text("ผู้นี้เกิด\(dayLabel) ไร้กาลกิณี")
                    .font(.sarabun(12, weight: .semibold))
                    .foregroundStyle(selectedTheme.kalakiniTextColor)
                Text("✦ เลขศาสตร์นำพา พลังเงาเกื้อหนุน ✦")
                    .font(.sarabun(11, weight: .bold))
                    .foregroundStyle(selectedTheme.quoteTextColor)
                    .padding(.top, 3)

                cardDivider

                pairMeaning(
                    value: item.satPairsDisplay,
                    type: item.satPairtype,
                    description: item.satMiracleDesc
                )
                .padding(.bottom, 6)
                pairMeaning(
                    value: item.shaPairsDisplay,
                    type: item.shaPairtype,
                    description: item.shaMiracleDesc
                )

                Spacer(minLength: 4)

                HStack {
                    Group {
                        if let frogImage { Image(uiImage: frogImage).resizable().scaledToFit() }
                        else { Color.clear }
                    }
                    .frame(width: 82, height: 58)
                    .opacity(0.95)
                    Spacer()
                    Group {
                        if let goldImage { Image(uiImage: goldImage).resizable().scaledToFit() }
                        else { Color.clear }
                    }
                    .frame(width: 82, height: 58)
                    .opacity(0.95)
                }
            }
            .padding(.horizontal, 14)
            .padding(.vertical, 14)
            .frame(width: width * 0.88, height: height * 0.56)
            .background(selectedTheme.cardBackgroundBrush)
            .clipShape(RoundedRectangle(cornerRadius: 22))
            .overlay(
                RoundedRectangle(cornerRadius: 22)
                    .stroke(
                        LinearGradient(
                            colors: selectedTheme.cardBorderColors,
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        ),
                        lineWidth: 1.2
                    )
            )
            .shadow(color: selectedTheme == .ivory ? Color(hex: "#3D270C").opacity(0.18) : Color.black.opacity(0.25), radius: 10, x: 0, y: 4)
            .padding(.top, height * 0.32)

            Text("ชื่อดี.com")
                .font(.sarabun(11, weight: .bold))
                .foregroundStyle(selectedTheme.footerTextColor)
                .frame(maxHeight: .infinity, alignment: .bottom)
                .padding(.bottom, height * 0.035)
        }
        .frame(width: width, height: height)
        .clipped()
    }

    private var sparkleLayer: some View {
        GeometryReader { proxy in
            ForEach(0..<7, id: \.self) { index in
                Circle()
                    .fill(Color.white.opacity(index.isMultiple(of: 2) ? 0.55 : 0.30))
                    .frame(width: index.isMultiple(of: 2) ? 8 : 5)
                    .position(
                        x: proxy.size.width * [0.18, 0.82, 0.12, 0.88, 0.28, 0.72, 0.50][index],
                        y: proxy.size.height * [0.14, 0.19, 0.37, 0.42, 0.64, 0.70, 0.10][index]
                    )
            }
        }
    }

    @ViewBuilder
    private func pairMeaning(
        value: String?,
        type: String?,
        description: String?,
        circleSize: CGFloat = 36,
        fontSize: CGFloat = 13,
        lineWidth: CGFloat = 1.8
    ) -> some View {
        let values = parsedPairs(value)
        if !values.isEmpty {
            VStack(spacing: 3) {
                PairCircles(
                    pairs: values.map { ($0, type?.uppercased().hasPrefix("R") == true) },
                    size: circleSize,
                    fontSize: fontSize,
                    lineWidth: lineWidth,
                    spacing: 6,
                    onPairTap: viewModel.fetchPairDetail
                )
                Text(description?.nonEmpty ?? "พลังคู่เลขมงคลตามตำรา")
                    .font(.sarabun(11.5, weight: .medium))
                    .foregroundStyle(selectedTheme.pairDescTextColor)
                    .multilineTextAlignment(.center)
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
    }

    private var cardDivider: some View {
        Rectangle()
            .fill(selectedTheme.dividerColor)
            .frame(height: 1)
            .padding(.vertical, 10)
    }

    private var fullName: String {
        let candidate = maskedCandidate
        guard viewModel.mode == .combined else { return candidate }
        return viewModel.target == .surname
            ? "\(viewModel.referenceName) \(candidate)"
            : "\(candidate) \(viewModel.referenceName)"
    }

    private var maskedCandidate: String {
        guard !purchaseManager.effectiveIsUnlocked, !item.thname.isEmpty else { return item.thname }
        let visibleCount = max(2, item.thname.count / 3)
        return String(item.thname.prefix(visibleCount)) + "***"
    }

    private var nameFontSize: CGFloat {
        switch fullName.count {
        case 26...: return 16
        case 21...: return 18
        case 16...: return 21
        case 11...: return 23
        default: return 26
        }
    }

    private var dayLabel: String {
        ["sunday":"วันอาทิตย์", "monday":"วันจันทร์", "tuesday":"วันอังคาร", "wednesday1":"วันพุธ (กลางวัน)", "wednesday2":"วันพุธ (กลางคืน)", "thursday":"วันพฤหัสบดี", "friday":"วันศุกร์", "saturday":"วันเสาร์"][viewModel.dayOfBirth] ?? viewModel.dayOfBirth
    }

    private var buddhaURL: URL? {
        URL(string: "https://xn--b3cu8e7ah6h.com/buddhas/\(viewModel.dayOfBirth)_buddha.png?v=2")
    }

    private func parsedPairs(_ value: String?) -> [String] {
        (value ?? "")
            .components(separatedBy: CharacterSet.decimalDigits.inverted)
            .filter { !$0.isEmpty }
    }

    private func shareToSocial() {
        guard !isPreparingShare else { return }
        Task { @MainActor in
            if let image = await renderSocialSquareImage() {
                sharedImage = SharedBlessingImage(image: image)
            } else {
                exportNotice = "ไม่สามารถสร้างภาพได้ กรุณาลองอีกครั้ง"
            }
        }
    }

    private func saveToPhotos() {
        guard !isPreparingShare else { return }
        Task { @MainActor in
            guard let image = await renderWallpaperImage(), let pngData = image.pngData() else {
                exportNotice = "ไม่สามารถสร้างภาพได้ กรุณาลองอีกครั้ง"
                return
            }

            let authorization = await PhotoLibraryWriter.requestAddAuthorization()
            guard authorization == .authorized || authorization == .limited else {
                exportNotice = "กรุณาอนุญาตให้แอปเพิ่มรูปภาพใน Photos ผ่าน Settings"
                return
            }

            let saved = await PhotoLibraryWriter.savePNGData(pngData)
            exportNotice = saved ? "บันทึกการ์ดมงคลลง Photos แล้ว" : "บันทึกรูปภาพไม่สำเร็จ กรุณาลองอีกครั้ง"
        }
    }

    @MainActor
    private func renderWallpaperImage() async -> UIImage? {
        isPreparingShare = true
        defer { isPreparingShare = false }
        // Render at standard 390pt width @3x (1170x2535 px) for modern 9:19.5 phone lockscreen wallpaper
        let renderer = ImageRenderer(content: blessingCard(width: 390))
        renderer.scale = 3
        renderer.isOpaque = true
        return renderer.uiImage
    }

    @MainActor
    private func renderSocialSquareImage() async -> UIImage? {
        isPreparingShare = true
        defer { isPreparingShare = false }
        // Render at 1:1 square (420pt @3x = 1260x1260 px) matching Android's balanced social card layout
        let renderer = ImageRenderer(content: socialShareCard(size: 420))
        renderer.scale = 3
        renderer.isOpaque = true
        return renderer.uiImage
    }

    private func socialShareCard(size: CGFloat) -> some View {
        ZStack(alignment: .top) {
            if let backgroundImage {
                Image(uiImage: backgroundImage)
                    .resizable()
                    .scaledToFill()
                    .frame(width: size, height: size)
                    .clipped()
            } else {
                LinearGradient(
                    colors: [
                        Color(red: 1.0, green: 0.99, blue: 0.94),
                        Color(red: 0.98, green: 0.94, blue: 0.86),
                        Color(red: 0.94, green: 0.87, blue: 0.74)
                    ],
                    startPoint: .top,
                    endPoint: .bottom
                )
            }

            let isTuesday = viewModel.dayOfBirth.lowercased() == "tuesday"
            let isStandingBuddha = ["sunday", "monday", "friday"].contains(viewModel.dayOfBirth.lowercased())
            let tuesdayBoxSize = size * 0.58
            let buddhaWidth = isTuesday ? tuesdayBoxSize : (isStandingBuddha ? size * 0.35 : size * 0.33)
            let buddhaHeight = isTuesday ? tuesdayBoxSize : (isStandingBuddha ? size * 0.25 : size * 0.23)
            let buddhaTopPadding = isTuesday ? (size * 0.25 - (tuesdayBoxSize * 0.655)) : (isStandingBuddha ? size * 0.015 : size * 0.022)

            Group {
                if let buddhaImage {
                    Image(uiImage: buddhaImage)
                        .resizable()
                        .scaledToFit()
                        .frame(
                            width: buddhaWidth,
                            height: buddhaHeight
                        )
                } else {
                    Circle()
                        .fill(Color.white.opacity(0.22))
                        .overlay(Image(systemName: "sparkles").foregroundStyle(Color.brandGold))
                        .frame(width: buddhaWidth, height: buddhaHeight)
                }
            }
            .frame(width: buddhaWidth, height: buddhaHeight)
            .padding(.top, buddhaTopPadding)

            VStack(spacing: 0) {
                Text(fullName)
                    .font(.sarabun(nameFontSize, weight: .bold))
                    .foregroundStyle(selectedTheme.nameTextColor)
                    .lineLimit(1)
                    .minimumScaleFactor(0.6)
                    .padding(.top, 14)

                Text(item.meaning?.nonEmpty ?? "ยังไม่มีข้อมูลความหมายของชื่อนี้ในฐานข้อมูล")
                    .font(.sarabun(12, weight: .medium))
                    .foregroundStyle(selectedTheme.meaningTextColor)
                    .multilineTextAlignment(.center)
                    .lineLimit(2)
                    .lineSpacing(2)
                    .padding(.top, 4)
                    .padding(.horizontal, 6)

                Rectangle()
                    .fill(selectedTheme.dividerColor)
                    .frame(height: 1)
                    .padding(.vertical, 5)

                Text("ผู้นี้เกิด\(dayLabel) ไร้กาลกิณี")
                    .font(.sarabun(11.5, weight: .semibold))
                    .foregroundStyle(selectedTheme.kalakiniTextColor)
                Text("✦ เลขศาสตร์นำพา พลังเงาเกื้อหนุน ✦")
                    .font(.sarabun(10.5, weight: .bold))
                    .foregroundStyle(selectedTheme.quoteTextColor)
                    .padding(.top, 2)

                Rectangle()
                    .fill(selectedTheme.dividerColor)
                    .frame(height: 1)
                    .padding(.vertical, 5)

                VStack(spacing: 5) {
                    pairMeaning(
                        value: item.satPairsDisplay,
                        type: item.satPairtype,
                        description: item.satMiracleDesc,
                        circleSize: 30,
                        fontSize: 13,
                        lineWidth: 1.5
                    )

                    pairMeaning(
                        value: item.shaPairsDisplay,
                        type: item.shaPairtype,
                        description: item.shaMiracleDesc,
                        circleSize: 30,
                        fontSize: 13,
                        lineWidth: 1.5
                    )
                }
                .padding(.horizontal, 2)

                Spacer(minLength: 4)

                HStack {
                    Group {
                        if let frogImage { Image(uiImage: frogImage).resizable().scaledToFit() }
                        else { Color.clear }
                    }
                    .frame(width: 68, height: 44)
                    .opacity(0.95)

                    Spacer()

                    Text("ชื่อดี.com")
                        .font(.sarabun(11, weight: .bold))
                        .foregroundStyle(selectedTheme.footerTextColor.opacity(0.95))

                    Spacer()

                    Group {
                        if let goldImage { Image(uiImage: goldImage).resizable().scaledToFit() }
                        else { Color.clear }
                    }
                    .frame(width: 68, height: 44)
                    .opacity(0.95)
                }
                .padding(.horizontal, 4)
                .padding(.bottom, 4)
            }
            .padding(.horizontal, 14)
            .padding(.vertical, 2)
            .frame(width: size * 0.91, height: size * 0.73)
            .background(selectedTheme.cardBackgroundBrush)
            .clipShape(RoundedRectangle(cornerRadius: 18))
            .overlay(
                RoundedRectangle(cornerRadius: 18)
                    .stroke(
                        LinearGradient(
                            colors: selectedTheme.cardBorderColors,
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        ),
                        lineWidth: 1.2
                    )
            )
            .shadow(color: selectedTheme == .ivory ? Color(hex: "#3D270C").opacity(0.18) : Color.black.opacity(0.25), radius: 8, x: 0, y: 3)
            .padding(.top, size * 0.245)
        }
        .frame(width: size, height: size)
        .clipped()
    }

    private func loadDecorationImages() async {
        await loadBackgroundImage(for: selectedTheme)
        async let loadedBuddha = loadImage(from: buddhaURL)
        async let loadedFrog = loadImage(from: URL(string: "https://xn--b3cu8e7ah6h.com/images/frog_left.png"))
        async let loadedGold = loadImage(from: URL(string: "https://xn--b3cu8e7ah6h.com/images/gold.png"))
        let images = await (loadedBuddha, loadedFrog, loadedGold)
        buddhaImage = images.0
        frogImage = images.1
        goldImage = images.2
    }

    private func loadBackgroundImage(for theme: WallpaperTheme) async {
        let loadedBg = await loadImage(from: URL(string: "https://xn--b3cu8e7ah6h.com/images/\(theme.rawValue)?v=2"))
        if let loadedBg {
            backgroundImage = loadedBg
        }
    }

    private func loadImage(from url: URL?) async -> UIImage? {
        guard let url else { return nil }
        do {
            var request = URLRequest(url: url)
            request.cachePolicy = .returnCacheDataElseLoad
            request.timeoutInterval = 20
            let (data, response) = try await URLSession.shared.data(for: request)
            guard let http = response as? HTTPURLResponse,
                  (200...299).contains(http.statusCode) else { return nil }
            return UIImage(data: data)
        } catch {
            return nil
        }
    }
}

enum WallpaperTheme: String, CaseIterable, Identifiable {
    case ivory = "blessing_wallpaper_bg_ivory.jpg"
    case emerald = "blessing_wallpaper_bg_emerald.jpg"
    case midnight = "blessing_wallpaper_bg_midnight.jpg"

    var id: String { rawValue }
    var title: String {
        switch self {
        case .ivory: return "✨ ทองคำ"
        case .emerald: return "🌿 มรกต"
        case .midnight: return "🪐 จักรวาล"
        }
    }

    var cardBackgroundBrush: LinearGradient {
        switch self {
        case .ivory:
            return LinearGradient(
                colors: [Color(hex: "#F9F0D9").opacity(0.95), Color(hex: "#EBD8AA").opacity(0.95)],
                startPoint: .top,
                endPoint: .bottom
            )
        case .emerald:
            return LinearGradient(
                colors: [Color(red: 0.03, green: 0.15, blue: 0.09).opacity(0.88), Color(red: 0.02, green: 0.11, blue: 0.07).opacity(0.88)],
                startPoint: .top,
                endPoint: .bottom
            )
        case .midnight:
            return LinearGradient(
                colors: [Color(red: 0.04, green: 0.06, blue: 0.12).opacity(0.88), Color(red: 0.02, green: 0.04, blue: 0.08).opacity(0.88)],
                startPoint: .top,
                endPoint: .bottom
            )
        }
    }

    var cardBackground: Color {
        switch self {
        case .ivory: return Color(hex: "#F3E4C2").opacity(0.95)
        case .emerald: return Color(red: 0.03, green: 0.15, blue: 0.09).opacity(0.88)
        case .midnight: return Color(red: 0.04, green: 0.06, blue: 0.12).opacity(0.88)
        }
    }

    var nameTextColor: Color {
        switch self {
        case .ivory: return Color(hex: "#24170A")
        case .emerald, .midnight: return Color(red: 1.0, green: 0.87, blue: 0.49)
        }
    }

    var meaningTextColor: Color {
        switch self {
        case .ivory: return Color(hex: "#503D2A")
        case .emerald: return Color(red: 0.88, green: 0.95, blue: 0.91)
        case .midnight: return Color(red: 0.94, green: 0.92, blue: 0.88)
        }
    }

    var footerTextColor: Color {
        switch self {
        case .ivory: return Color(hex: "#6B4918")
        case .emerald, .midnight: return Color(red: 1.0, green: 0.87, blue: 0.49)
        }
    }

    var kalakiniTextColor: Color {
        switch self {
        case .ivory: return Color(hex: "#1E582D")
        case .emerald, .midnight: return Color(red: 0.51, green: 0.78, blue: 0.52)
        }
    }

    var quoteTextColor: Color {
        switch self {
        case .ivory: return Color(hex: "#7A5012")
        case .emerald, .midnight: return Color(red: 1.0, green: 0.84, blue: 0.31)
        }
    }

    var dividerColor: Color {
        switch self {
        case .ivory: return Color(hex: "#C4A255").opacity(0.45)
        case .emerald, .midnight: return Color(red: 0.83, green: 0.69, blue: 0.22).opacity(0.35)
        }
    }

    var pairDescTextColor: Color {
        switch self {
        case .ivory: return Color(hex: "#503D2A")
        case .emerald: return Color(red: 0.88, green: 0.95, blue: 0.91)
        case .midnight: return Color(red: 0.94, green: 0.92, blue: 0.88)
        }
    }

    var cardBorderColors: [Color] {
        switch self {
        case .ivory:
            return [Color(hex: "#F7E5BC"), Color(hex: "#D4AF37").opacity(0.85), Color(hex: "#E8D3A2")]
        case .emerald:
            return [Color(red: 1.0, green: 0.88, blue: 0.51), Color(red: 0.51, green: 0.78, blue: 0.52).opacity(0.6), Color(red: 1.0, green: 0.84, blue: 0.31)]
        case .midnight:
            return [Color(red: 1.0, green: 0.87, blue: 0.49), Color(red: 0.83, green: 0.69, blue: 0.22).opacity(0.65), Color(red: 1.0, green: 0.87, blue: 0.49)]
        }
    }
}

private struct SharedBlessingImage: Identifiable {
    let id = UUID()
    let image: UIImage
}

private struct ActivityShareSheet: UIViewControllerRepresentable {
    let items: [Any]

    func makeUIViewController(context: Context) -> UIActivityViewController {
        UIActivityViewController(activityItems: items, applicationActivities: nil)
    }

    func updateUIViewController(_ uiViewController: UIActivityViewController, context: Context) {}
}

/// Photos invokes its completion handlers on private background queues. Keeping
/// those closures outside MainActor isolation avoids Swift 6 queue assertions.
private enum PhotoLibraryWriter {
    nonisolated static func requestAddAuthorization() async -> PHAuthorizationStatus {
        await withCheckedContinuation { continuation in
            PHPhotoLibrary.requestAuthorization(for: .addOnly) { status in
                continuation.resume(returning: status)
            }
        }
    }

    nonisolated static func savePNGData(_ data: Data) async -> Bool {
        await withCheckedContinuation { continuation in
            PHPhotoLibrary.shared().performChanges {
                let request = PHAssetCreationRequest.forAsset()
                request.addResource(with: .photo, data: data, options: nil)
            } completionHandler: { success, _ in
                continuation.resume(returning: success)
            }
        }
    }
}

private extension String {
    var nonEmpty: String? {
        let value = trimmingCharacters(in: .whitespacesAndNewlines)
        return value.isEmpty ? nil : value
    }
}
