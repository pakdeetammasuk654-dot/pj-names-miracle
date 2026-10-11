import SwiftUI

struct RankingResultsView: View {
    enum Filter { case all, good, caution }
    @ObservedObject var viewModel: SearchViewModel
    @State private var selectedFilter: Filter = .all
    @State private var showsTapHint = false
    @State private var hasShownTapHint = false
    @State private var previousRanks: [Int: Int] = [:]
    @State private var rankMovements: [Int: Int] = [:]
    @State private var newlyRankedIDs: Set<Int> = []
    @State private var highlightedTopID: Int?
    @State private var rankingEffectRevision = 0
    @State private var selectedCardItem: NameResultItem?
    @State private var showsBlessingCard = false
    @State private var showUpButton = false
    @State private var showsChatSheet = false
    @StateObject private var purchaseManager = PurchaseManager.shared

    private var filteredResults: [NameResultItem] {
        switch selectedFilter {
        case .all: return viewModel.results
        case .good: return viewModel.results.filter { !$0.hasBadPair }
        case .caution: return viewModel.results.filter(\.hasBadPair)
        }
    }
    private var goodCount: Int { viewModel.results.count { !$0.hasBadPair } }
    private var cautionCount: Int { viewModel.results.count(where: \.hasBadPair) }

    var body: some View {
        ZStack(alignment: .bottom) {
            Color.brandBackground.ignoresSafeArea()
            ScrollViewReader { scrollProxy in
                ScrollView {
                    LazyVStack(spacing: 0, pinnedViews: [.sectionHeaders]) {
                        Color.clear.frame(height: 1).id("rankingTopAnchor")
                            .onAppear {
                                withAnimation(.easeInOut(duration: 0.2)) { showUpButton = false }
                            }
                            .onDisappear {
                                withAnimation(.easeInOut(duration: 0.2)) { showUpButton = true }
                            }

                        formulaCard.padding(.horizontal, 3).padding(.vertical, 7)
                        filterBar
                        Section {
                            if !purchaseManager.effectiveIsUnlocked, !viewModel.results.isEmpty {
                                unlockCard
                                Divider().overlay(Color.white.opacity(0.1))
                            }
                            ForEach(Array(filteredResults.enumerated()), id: \.element.id) { index, item in
                                resultRow(item, isEvenRow: index % 2 == 0)
                                    .transition(.asymmetric(
                                        insertion: .opacity.combined(with: .scale(scale: 0.96)),
                                        removal: .opacity
                                    ))
                                Divider().overlay(Color.white.opacity(0.08))
                            }
                        } header: { tableHeader }
                    }
                }
                .overlay(alignment: .bottomTrailing) {
                    VStack(spacing: 12) {
                        if showUpButton {
                            Button {
                                withAnimation(.easeInOut(duration: 0.35)) {
                                    scrollProxy.scrollTo("rankingTopAnchor", anchor: .top)
                                }
                            } label: {
                                Image(systemName: "chevron.up")
                                    .font(.system(size: 20, weight: .bold))
                                    .foregroundStyle(Color.brandGreen)
                                    .frame(width: 46, height: 46)
                                    .background(Color.brandGold)
                                    .clipShape(Circle())
                                    .shadow(color: Color.black.opacity(0.45), radius: 6, x: 0, y: 3)
                            }
                            .transition(.scale.combined(with: .opacity))
                        }

                        FloatingChatButton {
                            showsChatSheet = true
                        }
                    }
                    .padding(.trailing, 16)
                    .padding(.bottom, showsTapHint ? 66 : 20)
                    .zIndex(12)
                }
            }

            if showsTapHint {
                tapHintToast
                    .padding(.horizontal, 16)
                    .padding(.bottom, 18)
                    .transition(.move(edge: .bottom).combined(with: .opacity))
                    .zIndex(10)
            }
        }
        .font(.sarabun(14)).preferredColorScheme(.dark).tint(.brandGold)
        .navigationBarTitleDisplayMode(.inline)
        .toolbarBackground(Color.brandGreen, for: .navigationBar)
        .toolbarBackground(.visible, for: .navigationBar)
        .toolbar {
            ToolbarItem(placement: .principal) {
                HStack(spacing: 8) {
                    Image("AppLogo").resizable().scaledToFit().frame(width: 32, height: 32).clipShape(Circle()).overlay(Circle().stroke(Color.brandGold.opacity(0.65)))
                    Text(cautionCount > 0 ? "ผลจัดอันดับตามน้ำหนัก 4 มิติ" : "ผลจัดอันดับชื่อดี 4 มิติ")
                        .font(.sarabun(17, weight: .bold)).foregroundStyle(.white)
                }
            }
        }
        .task(id: viewModel.results.isEmpty) {
            guard !viewModel.results.isEmpty, !hasShownTapHint else { return }
            hasShownTapHint = true
            try? await Task.sleep(for: .milliseconds(550))
            guard !Task.isCancelled else { return }
            withAnimation(.spring(response: 0.42, dampingFraction: 0.86)) {
                showsTapHint = true
            }
            try? await Task.sleep(for: .seconds(3))
            guard !Task.isCancelled else { return }
            withAnimation(.easeInOut(duration: 0.28)) {
                showsTapHint = false
            }
        }
        .onAppear {
            if previousRanks.isEmpty {
                previousRanks = rankMap(for: viewModel.results)
            }
        }
        .onChange(of: viewModel.results.map(\.nameID)) {
            applyRankingChangeEffects()
        }
        .navigationDestination(isPresented: $showsBlessingCard) {
            if let selectedCardItem {
                BlessingCardView(item: selectedCardItem, viewModel: viewModel)
            }
        }
        .alert("การซื้อ", isPresented: Binding(
            get: { purchaseManager.message != nil },
            set: { if !$0 { purchaseManager.message = nil } }
        )) {
            Button("ตกลง", role: .cancel) { purchaseManager.message = nil }
        } message: {
            Text(purchaseManager.message ?? "")
        }
        .fullScreenCover(isPresented: $showsChatSheet) {
            SupportChatView()
                .onDisappear {
                    SupportChatViewModel.shared.onDisappear()
                    Task {
                        await SupportChatViewModel.shared.checkUnread()
                    }
                }
        }
    }

    private var unlockCard: some View {
        VStack(spacing: 8) {
            HStack(spacing: 12) {
                Image("AppLogo")
                    .resizable().scaledToFit().frame(width: 46, height: 46)
                    .clipShape(Circle()).overlay(Circle().stroke(Color.brandGold, lineWidth: 1.5))
                VStack(alignment: .leading, spacing: 3) {
                    Text("100 อันดับชื่อที่ดีที่สุด")
                        .font(.sarabun(14, weight: .bold)).foregroundStyle(Color.brandCream)
                    Text("ปลดล็อกเพื่อดู 100 ชื่อ")
                        .font(.sarabun(11)).foregroundStyle(.white.opacity(0.78))
                    Text(purchaseManager.displayPrice)
                        .font(.sarabun(10, weight: .bold)).foregroundStyle(Color.brandCream)
                }
                Spacer(minLength: 4)
                VStack(spacing: 6) {
                    Button {
                        Task { await purchaseManager.purchase() }
                    } label: {
                        Group {
                            if purchaseManager.isBusy { ProgressView().tint(Color(hex: "#1F120B")) }
                            else { Text("เปิดชื่อทั้งหมด") }
                        }
                        .font(.sarabun(11, weight: .bold))
                        .frame(minWidth: 92).padding(.vertical, 8)
                    }
                    .buttonStyle(.borderedProminent).tint(.brandGold).foregroundStyle(Color(hex: "#1F120B"))
                    .disabled(purchaseManager.isBusy)
                    Button("คืนสิทธิ์การซื้อ") { Task { await purchaseManager.restore() } }
                        .font(.sarabun(10, weight: .medium)).foregroundStyle(Color.brandCream.opacity(0.82))
                        .padding(.top, 2)
                        .disabled(purchaseManager.isBusy)
                }
            }

            HStack(spacing: 12) {
                Link("ข้อตกลงการใช้งาน (Terms)", destination: URL(string: "https://xn--b3cu8e7ah6h.com/terms")!)
                Text("•").foregroundStyle(.white.opacity(0.4))
                Link("นโยบายความเป็นส่วนตัว (Privacy)", destination: URL(string: "https://xn--b3cu8e7ah6h.com/privacy")!)
            }
            .font(.sarabun(9))
            .foregroundStyle(Color.brandCream.opacity(0.68))
        }
        .padding(14)
        .background(Color(hex: "#1F120B"))
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .overlay(RoundedRectangle(cornerRadius: 14).stroke(Color.brandGold, lineWidth: 1.5))
        .padding(.horizontal, 10).padding(.vertical, 9)
    }

    private var tapHintToast: some View {
        ShimmeringGoldToastView()
    }

    private var formulaCard: some View {
        VStack(alignment: .leading, spacing: 5) {
            HStack(spacing: 8) {
                Text("📌 \(searchContext)").font(.sarabun(12, weight: .bold)).foregroundStyle(Color.brandCream).lineLimit(2)
                Spacer()
                if viewModel.isReranking {
                    HStack(spacing: 5) { ProgressView().controlSize(.mini).tint(.brandGold); Text("กำลังคำนวณสูตรใหม่") }
                        .font(.sarabun(9, weight: .semibold)).foregroundStyle(Color.brandCream)
                } else {
                    Text(viewModel.rerankError == nil ? "ลากจุดปรับ % • แตะชื่อเพื่อดูต่อ" : "⚠️ ปรับสูตรไม่สำเร็จ")
                        .font(.sarabun(9, weight: .semibold)).foregroundStyle(viewModel.rerankError == nil ? Color.brandCream.opacity(0.86) : .orange).lineLimit(1)
                }
            }
            HStack(spacing: 9) {
                Text("🎯")
                    .font(.system(size: 14))
                    .frame(width: 28, height: 28)
                    .background(Color.brandGold.opacity(0.18), in: Circle())
                    .overlay(Circle().stroke(Color.brandGold.opacity(0.55), lineWidth: 1))
                Text(viewModel.pinnedMeaning)
                    .font(.sarabun(11.5, weight: .medium))
                    .foregroundStyle(.white)
                    .lineSpacing(2)
                    .lineLimit(2)
                Spacer(minLength: 0)
            }
            .padding(.horizontal, 10)
            .padding(.vertical, 8)
            .background(
                LinearGradient(
                    colors: [Color.brandGold.opacity(0.16), Color(hex: "#382014").opacity(0.92)],
                    startPoint: .leading,
                    endPoint: .trailing
                ),
                in: RoundedRectangle(cornerRadius: 10)
            )
            .overlay(
                RoundedRectangle(cornerRadius: 10)
                    .stroke(Color.brandGold.opacity(0.45), lineWidth: 0.8)
            )
            HStack(spacing: 12) {
                VStack(alignment: .leading, spacing: 11) {
                    weightLine("⭐", "เลขศาสตร์", viewModel.appliedWeights.numerology, .brandGold)
                    weightLine("🎯", "ความหมาย", viewModel.appliedWeights.semantic, .pink)
                    weightLine("🔍", "เอกลักษณ์", viewModel.appliedWeights.similarity, Color(red: 0.30, green: 0.78, blue: 0.72))
                    weightLine("🎶", "ไพเราะ", viewModel.appliedWeights.sound, .cyan)
                }.frame(width: 135)
                RadarWeightSelector(weights: Binding(
                    get: { viewModel.weights },
                    set: { viewModel.updateWeightsAndRerank($0) }
                ))
            }
        }
        .padding(.horizontal, 10).padding(.vertical, 9)
        .background(
            ZStack {
                Color(hex: "#26160E")
                StepHeaderPatternView(isCompleted: false)
                    .opacity(0.35)
            }
        )
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .overlay(RoundedRectangle(cornerRadius: 14).stroke(Color.brandGold.opacity(0.30), lineWidth: 0.8))
    }

    private func weightLine(_ icon: String, _ title: String, _ value: Double, _ color: Color) -> some View {
        HStack { Text(icon); Text(title).font(.sarabun(11, weight: .semibold)); Spacer(); Text("\(Int((value * 100).rounded()))%").font(.sarabun(12, weight: .bold)).foregroundStyle(color) }
    }

    private var filterBar: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: 7) {
                Text("สูตร: \(formulaLabel)").font(.sarabun(10, weight: .bold)).foregroundStyle(Color.brandCream)
                filterChip("ทั้งหมด \(viewModel.results.count)", .all)
                filterChip("คู่ดี \(goodCount)", .good)
                if cautionCount > 0 { filterChip("🔴 ระวัง \(cautionCount)", .caution) }
            }.padding(.horizontal, 10).padding(.vertical, 8)
        }.background(Color.brandGreen.opacity(0.48))
    }

    private func filterChip(_ title: String, _ filter: Filter) -> some View {
        Button { withAnimation(.easeInOut(duration: 0.2)) { selectedFilter = filter } } label: {
            Text(title).font(.sarabun(10, weight: .semibold)).foregroundStyle(selectedFilter == filter ? Color.brandCream : .secondary).padding(.horizontal, 11).padding(.vertical, 6)
                .background(selectedFilter == filter ? Color.brandGold.opacity(0.15) : Color.clear).clipShape(Capsule()).overlay(Capsule().stroke(selectedFilter == filter ? Color.brandGold.opacity(0.65) : Color.white.opacity(0.16)))
        }.buttonStyle(.plain)
    }

    private var formulaLabel: String {
        let values = [
            ("เลขศาสตร์", viewModel.weights.numerology),
            ("ความหมาย", viewModel.weights.semantic),
            ("เอกลักษณ์", viewModel.weights.similarity),
            ("ความไพเราะ", viewModel.weights.sound)
        ]
        return "เน้น" + (values.max(by: { $0.1 < $1.1 })?.0 ?? "สมดุล")
    }

    private var tableHeader: some View {
        VStack(spacing: 0) {
            if viewModel.mode == .combined && !viewModel.referenceName.isEmpty {
                HStack(spacing: 6) {
                    Text("🏷️ " + (viewModel.target == .surname ? "ค้นหานามสกุลสำหรับชื่อ:" : "ค้นหาชื่อสำหรับนามสกุล:"))
                        .font(.sarabun(11, weight: .semibold))
                        .foregroundStyle(Color.brandGold)
                    Text(viewModel.referenceName)
                        .font(.sarabun(12, weight: .bold))
                        .foregroundStyle(Color.brandCream)
                    Spacer()
                }
                .padding(.horizontal, 10)
                .padding(.vertical, 6)
                .background(Color(hex: "#160C08"))

                Rectangle()
                    .fill(Color.brandGold.opacity(0.25))
                    .frame(height: 0.5)
            }

            HStack(spacing: 0) {
                headerLabel("#", color: .brandGold)
                    .frame(width: 54)
                headerDivider
                headerLabel(viewModel.mode == .combined ? combinedColumnTitle : "ชื่อ")
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.leading, 10)
                headerDivider
                headerLabel("คู่เลข", color: .brandGold)
                    .frame(width: 130)
                headerDivider
                headerLabel("คะแนน", color: .brandGold)
                    .frame(width: 70, alignment: .trailing)
            }
            .padding(.horizontal, 8)
            .frame(height: 44)
            .background(
                LinearGradient(
                    colors: [Color(hex: "#1B0F09"), Color(hex: "#26160E"), Color(hex: "#1B0F09")],
                    startPoint: .leading,
                    endPoint: .trailing
                )
            )
        }
        .frame(maxWidth: .infinity)
        .background(Color(hex: "#140B07"))
        .overlay(alignment: .top) {
            Rectangle().fill(Color.brandGold.opacity(0.34)).frame(height: 0.75)
        }
        .overlay(alignment: .bottom) {
            Rectangle().fill(Color.brandGold.opacity(0.72)).frame(height: 1)
        }
    }

    private func headerLabel(_ title: String, color: Color = .brandCream) -> some View {
        Text(title)
            .font(.sarabun(12, weight: .bold))
            .tracking(0.25)
            .foregroundStyle(color)
            .lineLimit(1)
            .shadow(color: .black.opacity(0.35), radius: 1, y: 1)
    }

    private var headerDivider: some View {
        Rectangle()
            .fill(Color.brandGold.opacity(0.42))
            .frame(width: 1, height: 30)
    }

    private var combinedColumnTitle: String { viewModel.target == .surname ? "นามสกุลมงคล" : "ชื่อมงคล" }

    private func resultRow(_ item: NameResultItem, isEvenRow: Bool = true) -> some View {
        HStack(spacing: 0) {
            Text("#\(item.rank)").font(.sarabun(14, weight: .bold)).foregroundStyle(item.rank <= 3 ? Color.brandGold : .secondary).frame(width: 54)
            Text(maskedCandidate(item.thname, rank: item.rank))
                .font(.sarabun(17, weight: .bold))
                .lineLimit(1)
                .frame(maxWidth: .infinity, alignment: .leading)
            VStack(alignment: .center, spacing: 2) {
                if viewModel.mode == .combined && !item.candidatePairs.isEmpty {
                    // Tier 1: label above circles
                    Text(viewModel.target == .surname ? "สกุล" : "ชื่อ")
                        .font(.sarabun(8, weight: .bold))
                        .foregroundStyle(Color.brandGold)
                    PairCircles(pairs: item.candidatePairs, size: 24, fontSize: 10, lineWidth: 1.2, spacing: 4, onPairTap: viewModel.fetchPairDetail)
                    Divider()
                        .background(Color.white.opacity(0.15))
                        .padding(.vertical, 1)
                    // Tier 2: label above circles
                    Text("รวม")
                        .font(.sarabun(8, weight: .bold))
                        .foregroundStyle(Color.brandCream.opacity(0.85))
                    PairCircles(pairs: item.displayedPairs, size: 24, fontSize: 10, lineWidth: 1.2, spacing: 4, onPairTap: viewModel.fetchPairDetail)
                } else {
                    PairCircles(pairs: item.displayedPairs, size: 32, fontSize: 12, lineWidth: 1.5, spacing: 5, onPairTap: viewModel.fetchPairDetail)
                }
            }
            .frame(width: 130)

            HStack(spacing: 6) {
                VStack(alignment: .trailing, spacing: 1) {
                    Text(item.finalScore, format: .number.precision(.fractionLength(2)))
                        .font(.system(size: 12, weight: .semibold, design: .rounded).monospacedDigit())
                        .foregroundStyle(Color.brandCream)
                    if let movement = rankMovements[item.nameID], movement != 0 {
                        Text(movement > 0 ? "↑\(movement)" : "↓\(-movement)")
                            .font(.system(size: 9, weight: .bold, design: .rounded))
                            .foregroundStyle(movement > 0 ? Color.green : Color.red)
                            .transition(.opacity.combined(with: .scale))
                    } else if newlyRankedIDs.contains(item.nameID) {
                        Text("ใหม่")
                            .font(.sarabun(8, weight: .bold))
                            .foregroundStyle(Color.green)
                            .transition(.opacity.combined(with: .scale))
                    }
                }
                Image(systemName: "chevron.right").font(.system(size: 12, weight: .bold)).foregroundStyle(.secondary)
            }.frame(width: 70, alignment: .trailing)
        }
        .padding(.horizontal, 8)
        .padding(.vertical, 16)
        .background(
            highlightedTopID == item.nameID
                ? Color(hex: "#382014")
                : (isEvenRow ? Color(hex: "#1F120B") : Color(hex: "#160C08"))
        )
        .contentShape(Rectangle())
        .animation(.easeInOut(duration: 0.45), value: highlightedTopID)
        .onTapGesture {
            selectedCardItem = item
            showsBlessingCard = true
        }
    }

    private func rankMap(for results: [NameResultItem]) -> [Int: Int] {
        Dictionary(uniqueKeysWithValues: results.enumerated().map { ($0.element.nameID, $0.offset + 1) })
    }

    private func applyRankingChangeEffects() {
        let currentRanks = rankMap(for: viewModel.results)
        guard !previousRanks.isEmpty else {
            previousRanks = currentRanks
            return
        }

        var movements: [Int: Int] = [:]
        var newIDs: Set<Int> = []
        for (id, newRank) in currentRanks {
            if let oldRank = previousRanks[id] {
                let movement = oldRank - newRank
                if movement != 0 { movements[id] = movement }
            } else {
                newIDs.insert(id)
            }
        }

        previousRanks = currentRanks
        rankingEffectRevision += 1
        let revision = rankingEffectRevision
        withAnimation(.spring(response: 0.48, dampingFraction: 0.82)) {
            rankMovements = movements
            newlyRankedIDs = newIDs
            highlightedTopID = viewModel.results.first?.nameID
        }

        Task { @MainActor in
            try? await Task.sleep(for: .seconds(1.8))
            guard revision == rankingEffectRevision else { return }
            withAnimation(.easeOut(duration: 0.3)) {
                rankMovements = [:]
                newlyRankedIDs = []
                highlightedTopID = nil
            }
        }
    }

    private func displayName(_ candidate: String) -> String {
        guard viewModel.mode == .combined else { return candidate }
        return viewModel.target == .surname ? "\(viewModel.referenceName) \(candidate)" : "\(candidate) \(viewModel.referenceName)"
    }

    private func maskedCandidate(_ candidate: String, rank: Int) -> String {
        guard !purchaseManager.effectiveIsUnlocked, !candidate.isEmpty else { return candidate }
        if rank <= 3 { return candidate }
        let visibleCount = max(2, candidate.count / 3)
        return String(candidate.prefix(visibleCount)) + "***"
    }

    private var searchContext: String {
        if viewModel.mode == .standalone { return "ค้นหาเฉพาะชื่อตัว · \(dayLabel)" }
        return viewModel.target == .surname
            ? "ค้นหานามสกุล สำหรับชื่อ: \(viewModel.referenceName) · \(dayLabel)"
            : "ค้นหาชื่อมงคล สำหรับนามสกุล: \(viewModel.referenceName) · \(dayLabel)"
    }

    private var dayLabel: String {
        ["sunday":"วันอาทิตย์", "monday":"วันจันทร์", "tuesday":"วันอังคาร", "wednesday1":"วันพุธ (กลางวัน)", "wednesday2":"วันพุธ (กลางคืน)", "thursday":"วันพฤหัสบดี", "friday":"วันศุกร์", "saturday":"วันเสาร์"][viewModel.dayOfBirth] ?? viewModel.dayOfBirth
    }
}

private struct ShimmeringGoldToastView: View {
    @State private var shimmerPhase: CGFloat = -1.0
    @State private var glowPulse: Bool = false

    var body: some View {
        HStack(spacing: 8) {
            Text("✨")
                .font(.system(size: 16))
                .scaleEffect(glowPulse ? 1.25 : 0.92)
                .animation(.easeInOut(duration: 0.8).repeatForever(autoreverses: true), value: glowPulse)
            Text("แตะชื่อใดก็ได้เพื่อเปิดการ์ดมงคล")
                .font(.sarabun(14, weight: .bold))
                .foregroundStyle(Color(hex: "#1F120B"))
        }
        .frame(maxWidth: .infinity)
        .padding(.horizontal, 18)
        .padding(.vertical, 16)
        .background(
            LinearGradient(
                colors: [
                    Color(red: 1.0, green: 0.88, blue: 0.35),
                    Color(red: 0.98, green: 0.76, blue: 0.12),
                    Color(red: 1.0, green: 0.84, blue: 0.30)
                ],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
        )
        .clipShape(RoundedRectangle(cornerRadius: 16))
        .overlay(
            GeometryReader { geo in
                let width = geo.size.width
                Rectangle()
                    .fill(
                        LinearGradient(
                            colors: [
                                .clear,
                                Color.white.opacity(0.12),
                                Color.white.opacity(0.80),
                                Color.white.opacity(0.12),
                                .clear
                            ],
                            startPoint: .leading,
                            endPoint: .trailing
                        )
                    )
                    .rotationEffect(.degrees(25))
                    .offset(x: shimmerPhase * width * 1.5)
            }
            .clipShape(RoundedRectangle(cornerRadius: 16))
        )
        .overlay(
            RoundedRectangle(cornerRadius: 16)
                .stroke(
                    LinearGradient(
                        colors: [
                            Color.white,
                            Color(red: 1.0, green: 0.95, blue: 0.7),
                            Color(red: 0.95, green: 0.74, blue: 0.15),
                            Color.white
                        ],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    ),
                    lineWidth: 1.8
                )
        )
        .shadow(color: Color(red: 1.0, green: 0.82, blue: 0.2).opacity(glowPulse ? 0.90 : 0.40), radius: glowPulse ? 16 : 8, x: 0, y: 0)
        .shadow(color: Color.black.opacity(0.35), radius: 10, x: 0, y: 5)
        .scaleEffect(glowPulse ? 1.015 : 1.0)
        .onAppear {
            withAnimation(.easeInOut(duration: 1.2)) {
                shimmerPhase = 1.2
            }
            withAnimation(.easeInOut(duration: 1.1).repeatForever(autoreverses: true)) {
                glowPulse = true
            }
        }
    }
}
