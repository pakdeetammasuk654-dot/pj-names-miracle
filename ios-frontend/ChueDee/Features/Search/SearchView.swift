import SwiftUI
import UIKit

// MARK: - Scroll offset observer (UIKit KVO — fires every frame)
private struct ScrollOffsetObserver: UIViewRepresentable {
    var onScroll: (CGFloat) -> Void

    func makeCoordinator() -> Coordinator { Coordinator(onScroll: onScroll) }

    func makeUIView(context: Context) -> UIView {
        let v = UIView(frame: .zero)
        v.backgroundColor = .clear
        v.isUserInteractionEnabled = false
        DispatchQueue.main.async { context.coordinator.attach(to: v) }
        return v
    }

    func updateUIView(_ uiView: UIView, context: Context) {
        context.coordinator.onScroll = onScroll
    }

    final class Coordinator {
        var onScroll: (CGFloat) -> Void
        private var observation: NSKeyValueObservation?
        private weak var scrollView: UIScrollView?

        init(onScroll: @escaping (CGFloat) -> Void) { self.onScroll = onScroll }

        func attach(to view: UIView) {
            guard scrollView == nil else { return }
            var current: UIView? = view
            while let parent = current?.superview {
                if let sv = parent as? UIScrollView {
                    scrollView = sv
                    observation = sv.observe(\.contentOffset, options: [.new]) { [weak self] sv, _ in
                        DispatchQueue.main.async { self?.onScroll(sv.contentOffset.y) }
                    }
                    return
                }
                current = parent
            }
        }
    }
}

struct SearchView: View {
    @StateObject var viewModel: SearchViewModel
    @StateObject private var chatViewModel = SupportChatViewModel.shared
    @FocusState private var isNameFieldFocused: Bool
    @State private var showsRanking = false
    @State private var showsGuideSheet = false
    @State private var showsChatSheet = false
    @State private var showsAuspicious = false
    @State private var showsDebugPurchase = false
    @State private var showsUpButton = false
    @State private var isTopHeaderCollapsed = false
    @State private var showsResetAlert = false
    @State private var scrollToTopTrigger = 0
    @State private var targetScrollStep = 1
    @State private var scrollToStepTrigger = 0
    @StateObject private var purchaseManager = PurchaseManager.shared
    private let days = [("sunday","วันอาทิตย์"),("monday","วันจันทร์"),("tuesday","วันอังคาร"),("wednesday1","วันพุธ (กลางวัน)"),("wednesday2","วันพุธ (กลางคืน)"),("thursday","วันพฤหัสบดี"),("friday","วันศุกร์"),("saturday","วันเสาร์")]

    private var step1Done: Bool {
        !viewModel.referenceName.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    private var step2Done: Bool {
        step1Done && !viewModel.dayOfBirth.isEmpty
    }

    private var step3Done: Bool {
        step2Done && viewModel.hasUserSelectedMeaning && !viewModel.pinnedMeaning.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    private var step4Done: Bool {
        // Step 4 represents the actual search, not merely completing earlier fields.
        step3Done && viewModel.state == .loaded
    }

    private var currentStep: Int {
        if !step1Done { return 1 }
        if !step2Done { return 2 }
        if !step3Done { return 3 }
        return 4
    }

    private var stepProgressBar: some View {
        StepProgressBar(
            currentStep: currentStep,
            step1Done: step1Done,
            step2Done: step2Done,
            step3Done: step3Done,
            step4Done: step4Done,
            onReset: {
                purchaseManager.message = nil
                showsResetAlert = true
            },
            onStepClick: { step in
                targetScrollStep = step
                scrollToStepTrigger += 1
            }
        )
        .alert("รีเซ็ตข้อมูลทั้งหมด?", isPresented: $showsResetAlert) {
            Button("ยกเลิก", role: .cancel) {}
            Button("รีเซ็ต", role: .destructive) {
                purchaseManager.message = nil
                viewModel.resetAll()
                withAnimation(.easeInOut(duration: 0.35)) {
                    isTopHeaderCollapsed = false
                    showsUpButton = false
                    scrollToTopTrigger += 1
                }
            }
        } message: {
            Text("ข้อมูลชื่อ วันเกิด ความหมาย และผลลัพธ์จะถูกล้างเพื่อเริ่มต้นใหม่")
        }
    }

    var body: some View {
        NavigationStack {
            ZStack(alignment: .bottomTrailing) {
                Color.brandBackground.ignoresSafeArea()
                VStack(spacing: 0) {
                    header

                    ScrollViewReader { scrollProxy in
                        ScrollView {
                            VStack(spacing: 18) {
                                Color.clear
                                    .frame(height: 1)
                                    .id("searchTopAnchor")
                                    .background(ScrollOffsetObserver { offsetY in
                                        if offsetY > 40 {
                                            if !isTopHeaderCollapsed {
                                                withAnimation(.spring(response: 0.35, dampingFraction: 0.8)) {
                                                    isTopHeaderCollapsed = true
                                                    showsUpButton = true
                                                }
                                            }
                                        } else if offsetY < 5 {
                                            if isTopHeaderCollapsed {
                                                withAnimation(.spring(response: 0.35, dampingFraction: 0.8)) {
                                                    isTopHeaderCollapsed = false
                                                    showsUpButton = false
                                                }
                                            }
                                        }
                                    })

                                form
                                step2BirthdaySection
                                    .id("step2BirthdaySection")
                                wishSection
                                    .id("wishSection")
                                formula
                                    .id("step4FormulaSection")
                                searchStatus
                                footerLinks
                            }
                            // Use one shared page gutter so every content frame has
                            // the same full-width alignment.
                            .padding(.horizontal, 3)
                            .padding(.vertical, 14)
                        }
                        .overlay(alignment: .bottomTrailing) {
                            VStack(spacing: 12) {
                                if showsUpButton {
                                    Button {
                                        withAnimation(.easeInOut(duration: 0.35)) {
                                            isTopHeaderCollapsed = false
                                            scrollProxy.scrollTo("searchTopAnchor", anchor: .top)
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

                                Button {
                                    showsChatSheet = true
                                } label: {
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
                                .accessibilityLabel("ปรึกษาคุณทญา")
                            }
                            .padding(.trailing, 16)
                            .padding(.bottom, 20)
                            .zIndex(12)
                        }
                        .onChange(of: scrollToTopTrigger) {
                            withAnimation(.easeInOut(duration: 0.35)) {
                                isTopHeaderCollapsed = false
                                showsUpButton = false
                                scrollProxy.scrollTo("searchTopAnchor", anchor: .top)
                            }
                        }
                        .onChange(of: scrollToStepTrigger) {
                            withAnimation(.easeInOut(duration: 0.35)) {
                                switch targetScrollStep {
                                case 1:
                                    showsUpButton = false
                                    isTopHeaderCollapsed = false
                                    scrollProxy.scrollTo("searchTopAnchor", anchor: .top)
                                case 2:
                                    scrollProxy.scrollTo("step2BirthdaySection", anchor: .top)
                                case 3:
                                    scrollProxy.scrollTo("wishSection", anchor: .top)
                                case 4:
                                    scrollProxy.scrollTo("step4FormulaSection", anchor: .top)
                                default:
                                    break
                                }
                            }
                        }
                    }
                }
                if viewModel.state == .loading {
                    MagicalLoadingView().transition(.opacity.combined(with: .scale(scale: 1.02))).zIndex(10)
                }
            }
                .navigationDestination(isPresented: $showsRanking) {
                    RankingResultsView(viewModel: viewModel)
                }
                .navigationDestination(isPresented: $showsAuspicious) {
                    AuspiciousCalendarView()
                }
                .font(.sarabun(14)).tint(.brandGold).preferredColorScheme(.dark)
                .task {
                    if !viewModel.hasUserSelectedMeaning || viewModel.suggestions.isEmpty {
                        await viewModel.refreshMeanings()
                    }
                    await chatViewModel.checkUnread()
                }
                .task(id: viewModel.isMeaningAutoPlaying) {
                    guard viewModel.isMeaningAutoPlaying else { return }
                    while !Task.isCancelled {
                        try? await Task.sleep(nanoseconds: 5_000_000_000)
                        guard !Task.isCancelled && viewModel.isMeaningAutoPlaying else { break }
                        await MainActor.run {
                            viewModel.nextMeaningPage()
                        }
                    }
                }
                .task {
                    while !Task.isCancelled {
                        try? await Task.sleep(nanoseconds: 8_000_000_000)
                        guard !Task.isCancelled else { break }
                        await chatViewModel.checkUnread()
                    }
                }
                .onChange(of: viewModel.state) { _, state in
                    if state == .loaded { showsRanking = true }
                }
                .onChange(of: purchaseManager.isUnlocked) {
                    viewModel.refreshSuggestionsAfterPurchase()
                }
                .onChange(of: chatViewModel.openChatRequested) { _, requested in
                    if requested {
                        showsChatSheet = true
                        chatViewModel.openChatRequested = false
                    }
                }
                .sheet(isPresented: $showsGuideSheet) {
                    MethodologyGuideSheet(articles: viewModel.articles)
                }
                .fullScreenCover(isPresented: $showsChatSheet) {
                    SupportChatView()
                        .onDisappear {
                            chatViewModel.onDisappear()
                            Task {
                                await chatViewModel.checkUnread()
                            }
                        }
                }
#if DEBUG
                .sheet(isPresented: $showsDebugPurchase, onDismiss: {
                    viewModel.refreshSuggestionsAfterPurchase()
                }) {
                    DebugPurchaseView(purchaseManager: purchaseManager)
                }
#endif
                .sheet(isPresented: isPairDetailPresented) {
                    pairDetailSheetContent
                }
                .alert("การซื้อ", isPresented: Binding(
                    get: { purchaseManager.message != nil },
                    set: { if !$0 { purchaseManager.message = nil } }
                )) {
                    Button("ตกลง", role: .cancel) { purchaseManager.message = nil }
                } message: {
                    Text(purchaseManager.message ?? "")
                }
        }
    }
    private var isPairDetailPresented: Binding<Bool> {
        Binding(
            get: { viewModel.isPairDetailLoading || viewModel.selectedPairDetail != nil || viewModel.pairDetailError != nil },
            set: { if !$0 { viewModel.dismissPairDetail() } }
        )
    }

    private var pairDetailSheetContent: some View {
        PairDetailSheet(
            detail: viewModel.selectedPairDetail,
            isLoading: viewModel.isPairDetailLoading,
            errorMessage: viewModel.pairDetailError,
            onDismiss: viewModel.dismissPairDetail
        )
        .presentationDetents([.medium, .large])
        .presentationDragIndicator(.visible)
    }

    @ViewBuilder private var celebrityStrip: some View {
        if viewModel.mode == .standalone && !viewModel.celebrities.isEmpty {
            InteractiveCelebrityMarquee(
                celebrities: viewModel.celebrities,
                selectedName: viewModel.referenceName,
                avatarURL: viewModel.avatarURL(for:),
                onSelect: viewModel.selectCelebrity
            )
            .transition(.opacity)
        }
    }

    private var header: some View {
        VStack(spacing: 0) {
            if !isTopHeaderCollapsed {
                topHeaderRow
                    .transition(.move(edge: .top).combined(with: .opacity))
            }
            stepProgressBar
                .padding(.horizontal, 4)
                .padding(.top, 4)
                .padding(.bottom, 6)

            Divider().overlay(Color.brandGold.opacity(0.35))
        }
        .background(Color.brandGreen.ignoresSafeArea(edges: .top))
    }

    private var topHeaderRow: some View {
        HStack(spacing: 10) {
            Image("AppLogo").resizable().scaledToFit().frame(width: 42, height: 42).clipShape(Circle()).overlay(Circle().stroke(Color.brandGold.opacity(0.7)))
            VStack(alignment: .leading, spacing: 2) {
                Text("ชื่อดี 4 มิติ").font(.sarabun(16, weight: .bold)).foregroundStyle(Color.brandCream)
                Text("ศาสตร์ตั้งชื่อมงคลอัจฉริยะ").font(.sarabun(11)).foregroundStyle(Color.brandGold.opacity(0.85))
            }
            Spacer()
            Button {
                showsAuspicious = true
            } label: {
                Image(systemName: "calendar.badge.clock")
                    .font(.title2)
                    .foregroundStyle(Color.brandGold)
            }
            .accessibilityLabel("ดูฤกษ์ยาม")
            Button {
                showsChatSheet = true
            } label: {
                ZStack(alignment: .topTrailing) {
                    Image(systemName: "bubble.left.and.bubble.right.fill")
                        .font(.system(size: 20))
                        .foregroundStyle(Color.brandGold)
                        .frame(width: 28, height: 28)

                    if chatViewModel.unreadCount > 0 {
                        Text(chatViewModel.unreadCount > 99 ? "99+" : "\(chatViewModel.unreadCount)")
                            .font(.system(size: 10, weight: .bold))
                            .foregroundColor(.white)
                            .padding(.horizontal, 4)
                            .padding(.vertical, 1)
                            .background(Color.red)
                            .clipShape(Capsule())
                            .offset(x: 8, y: -6)
                    }
                }
            }
            .accessibilityLabel("ปรึกษาคุณทญา")
            Button {
                showsGuideSheet = true
            } label: {
                Image(systemName: "books.vertical")
                    .font(.title2)
                    .foregroundStyle(Color.brandGold)
            }
            .accessibilityLabel("แนะนำ อ.ทญา")
#if DEBUG
            Button {
                Task {
                    await purchaseManager.setMockPurchase(enabled: true)
                    viewModel.refreshSuggestionsAfterPurchase()
                }
                showsDebugPurchase = true
            } label: {
                Image(systemName: "ladybug.fill")
                    .font(.title2)
                    .foregroundStyle(Color.orange)
            }
            .accessibilityLabel("เมนู Debug purchase")
#endif
        }
        .padding(.horizontal, 18)
        .padding(.vertical, 10)
    }

    private var form: some View {
        VStack(alignment: .leading, spacing: 10) {
            StepHeadingView(
                number: "1",
                title: viewModel.mode == .combined ? "ระบุนามสกุล (เพื่อหาชื่อคู่มงคล)" : "ระบุชื่อต้นแบบที่คุณชอบ",
                subtitle: viewModel.mode == .combined ? "ถอดรหัสคู่เลขและค้นหาชื่อที่เข้ากับนามสกุล" : "ถอดรหัสคู่เลขศาสตร์และค้นหาชื่อใกล้เคียง",
                isCompleted: step1Done,
                isActive: currentStep == 1,
                integratedWithSection: true
            )
            celebrityStrip
            VStack(spacing: 12) {
                HStack(spacing: 6) {
                    modeButton(.standalone, "เฉพาะชื่อตัว", "person.fill")
                    modeButton(.combined, "ชื่อ + นามสกุล", "book.closed.fill")
                }
                .padding(5)
                .background(Color(hex: "#1A0F0A").opacity(0.85))
                .clipShape(RoundedRectangle(cornerRadius: 12))
                .overlay(
                    RoundedRectangle(cornerRadius: 12)
                        .stroke(Color.white.opacity(0.12), lineWidth: 1)
                )

                if viewModel.mode == .combined {
                    HStack {
                        Spacer()
                        Button(action: viewModel.switchTarget) {
                            HStack(spacing: 6) {
                                Text(viewModel.target == .surname ? "กำลังหา: นามสกุล แตะเพื่อสลับหาชื่อตัว" : "กำลังหา: ชื่อตัว แตะเพื่อสลับหาสกุล")
                                    .font(.sarabun(11, weight: .semibold))
                                Text("⇄")
                                    .font(.sarabun(12, weight: .bold))
                            }
                            .foregroundStyle(Color.brandGold)
                            .padding(.horizontal, 12)
                            .padding(.vertical, 6)
                            .overlay(Capsule().stroke(Color.brandGold.opacity(0.55)))
                        }
                    }
                }

                // Outlined Name Field matching Android OutlinedTextField
                ZStack(alignment: .topLeading) {
                    HStack(spacing: 10) {
                        Image(systemName: "square.and.pencil")
                            .font(.system(size: 19, weight: .semibold))
                            .foregroundStyle(Color.brandGold)
                            .frame(width: 26)

                        TextField("", text: $viewModel.referenceName)
                            .focused($isNameFieldFocused)
                            .font(.sarabun(17, weight: .semibold))
                            .foregroundStyle(.white)
                            .textContentType(nil)
                            .textInputAutocapitalization(.never)
                            .autocorrectionDisabled()
                            .submitLabel(.done)
                            .onSubmit { isNameFieldFocused = false }
                            .onChange(of: viewModel.referenceName) { viewModel.nameChanged() }

                        if viewModel.isDecoding {
                            ProgressView().tint(.brandGold).controlSize(.small)
                                .accessibilityLabel("กำลังตรวจสอบชื่อ")
                        } else if !viewModel.referenceName.isEmpty {
                            Button {
                                withAnimation(.easeInOut(duration: 0.2)) {
                                    viewModel.referenceName = ""
                                    viewModel.nameChanged()
                                    isNameFieldFocused = true
                                }
                            } label: {
                                Image(systemName: "xmark")
                                    .font(.system(size: 14, weight: .bold))
                                    .foregroundStyle(Color.brandCream.opacity(0.72))
                                    .frame(width: 26, height: 26)
                            }
                            .accessibilityLabel("ล้างชื่อ")
                        }
                    }
                    .padding(.horizontal, 14)
                    .padding(.top, 16)
                    .padding(.bottom, 14)
                    .background(Color(hex: "#26160E"))
                    .clipShape(RoundedRectangle(cornerRadius: 14))
                    .overlay(
                        RoundedRectangle(cornerRadius: 14)
                            .stroke(isNameFieldFocused ? Color.brandGold : Color.brandGold.opacity(0.55), lineWidth: isNameFieldFocused ? 1.5 : 1)
                    )

                    // Floating label on the top border
                    Text(viewModel.fieldPrompt)
                        .font(.sarabun(12, weight: .medium))
                        .foregroundStyle(isNameFieldFocused ? Color.brandGold : Color.brandCream.opacity(0.78))
                        .padding(.horizontal, 6)
                        .background(Color(hex: "#1F120B"))
                        .offset(x: 14, y: -7)
                }
                .padding(.top, 4)

                if !viewModel.decodedMeaning.isEmpty {
                    VStack(spacing: 0) {
                        Divider().overlay(Color.brandGold.opacity(0.28)).padding(.horizontal, 12)
                        VStack(alignment: .leading, spacing: 7) {
                            let refName = viewModel.referenceName.trimmingCharacters(in: .whitespacesAndNewlines)
                            let titleText = !refName.isEmpty ? "💡 ความหมายของ “\(refName)”" : "💡 ความหมายของชื่อต้นแบบ:"
                            Text(titleText).font(.sarabun(12, weight: .semibold)).foregroundStyle(Color.brandGold)
                            let trimmedMeaning = viewModel.decodedMeaning.trimmingCharacters(in: .whitespacesAndNewlines)
                            let displayMeaning = (trimmedMeaning.hasPrefix("“") && trimmedMeaning.hasSuffix("”")) || (trimmedMeaning.hasPrefix("\"") && trimmedMeaning.hasSuffix("\""))
                                ? trimmedMeaning
                                : "“\(trimmedMeaning.trimmingCharacters(in: CharacterSet(charactersIn: "\"“”'")))”"
                            Text(displayMeaning)
                                .id(viewModel.decodedMeaning)
                                .font(.sarabun(13)).foregroundStyle(.primary)
                                .multilineTextAlignment(.center)
                                .frame(maxWidth: .infinity, alignment: .center)
                                .fixedSize(horizontal: false, vertical: true)
                                .transition(verticalReplacement)
                        }
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(.horizontal, 15).padding(.vertical, 13)
                    }
                    .transition(verticalReplacement)
                }

                decodedAnalysisSection
            }
            .padding(10)
            .background(Color(hex: "#1F120B"))
            .clipShape(RoundedRectangle(cornerRadius: 16))
            .animation(.easeInOut(duration: 0.2), value: isNameFieldFocused)
            .animation(.easeInOut(duration: 0.36), value: viewModel.decodedMeaning)
        }
        .padding(.horizontal, 8)
        .padding(.bottom, 8)
        .background(Color(hex: "#160C08"))
        .clipShape(RoundedRectangle(cornerRadius: 20))
        .overlay(
            RoundedRectangle(cornerRadius: 20)
                .stroke(Color.brandGold.opacity(0.30), lineWidth: 1)
        )
    }

    private var step2BirthdaySection: some View {
        VStack(alignment: .leading, spacing: 10) {
            StepHeadingView(
                number: "2",
                title: "ระบุวันเกิดประจำตัว",
                subtitle: "คัดกรองอักษรกาลกิณีให้ถูกต้องตามวันเกิด",
                isCompleted: step2Done,
                isActive: currentStep == 2,
                integratedWithSection: true
            )
            birthdayMenu
                .padding(.horizontal, 10)
                .padding(.bottom, 2)

            if !viewModel.dayOfBirth.isEmpty, let dayInfo = kalakiniDayInfo(for: viewModel.dayOfBirth) {
                kalakiniSummaryCard(dayInfo: dayInfo)
                    .padding(.horizontal, 10)
                    .padding(.bottom, 6)
            }
        }
        .padding(.horizontal, 8)
        .padding(.bottom, 8)
        .background(Color(hex: "#160C08"))
        .clipShape(RoundedRectangle(cornerRadius: 20))
        .overlay(RoundedRectangle(cornerRadius: 20).stroke(Color.brandGold.opacity(0.30), lineWidth: 1))
    }

    private struct KalakiniDayDisplay {
        let dayName: String
        let letters: String
    }

    private func kalakiniDayInfo(for dayKey: String) -> KalakiniDayDisplay? {
        switch dayKey {
        case "sunday":
            return KalakiniDayDisplay(dayName: "วันอาทิตย์", letters: "ศษสหฬฮ")
        case "monday":
            return KalakiniDayDisplay(dayName: "วันจันทร์", letters: "ะาิีึืุูเแโใไ")
        case "tuesday":
            return KalakiniDayDisplay(dayName: "วันอังคาร", letters: "กขคฆง")
        case "wednesday1":
            return KalakiniDayDisplay(dayName: "วันพุธ (กลางวัน)", letters: "จฉชซฌญ")
        case "wednesday2":
            return KalakiniDayDisplay(dayName: "วันพุธ (กลางคืน)", letters: "บปผฝพฟภม")
        case "thursday":
            return KalakiniDayDisplay(dayName: "วันพฤหัสบดี", letters: "ดตถทธน")
        case "friday":
            return KalakiniDayDisplay(dayName: "วันศุกร์", letters: "ยรลว")
        case "saturday":
            return KalakiniDayDisplay(dayName: "วันเสาร์", letters: "ฎฏฐฑฒณ")
        default:
            return nil
        }
    }

    private func kalakiniSummaryCard(dayInfo: KalakiniDayDisplay) -> some View {
        HStack(spacing: 8) {
            Image(systemName: "checkmark.shield.fill")
                .font(.system(size: 15, weight: .semibold))
                .foregroundStyle(Color(hex: "#FFDF85"))
            Text("คัดเฉพาะชื่อมงคลปลอดกาลกิณี 100%")
                .font(.sarabun(12.5, weight: .medium))
                .foregroundStyle(Color.brandCream.opacity(0.92))
            Spacer(minLength: 0)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 10)
        .background(Color(hex: "#26160E"))
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .overlay(
            RoundedRectangle(cornerRadius: 12)
                .stroke(Color.brandGold.opacity(0.28), lineWidth: 1)
        )
    }

    private var birthdayMenu: some View {
        Menu {
            ForEach(days, id: \.0) { day in
                Button {
                    viewModel.dayOfBirth = day.0
                    viewModel.refreshSelectedNameForBirthday()
                } label: {
                    if viewModel.dayOfBirth == day.0 {
                        Label(day.1, systemImage: "checkmark")
                    } else {
                        Text(day.1)
                    }
                }
            }
        } label: {
            HStack(spacing: 13) {
                Image(systemName: "calendar")
                    .font(.system(size: 21, weight: .medium))
                    .foregroundStyle(Color.brandGold)
                    .frame(width: 30)
                VStack(alignment: .leading, spacing: 3) {
                    Text("วันเกิดประจำตัว")
                        .font(.sarabun(12, weight: .medium))
                        .foregroundStyle(Color.brandCream.opacity(0.72))
                    Text(selectedDayLabel)
                        .font(.sarabun(16, weight: .semibold))
                        .foregroundStyle(viewModel.dayOfBirth.isEmpty ? Color.brandCream.opacity(0.58) : .white)
                }
                Spacer()
                Image(systemName: "chevron.up.chevron.down")
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundStyle(Color.brandCream.opacity(0.7))
            }
            .contentShape(Rectangle())
            .padding(.horizontal, 14)
            .padding(.vertical, 11)
            .background(Color(hex: "#26160E"))
            .clipShape(RoundedRectangle(cornerRadius: 14))
        }
        .buttonStyle(.plain)
        .accessibilityLabel("วันเกิดประจำตัว")
        .accessibilityValue(selectedDayLabel)
        .accessibilityHint("แตะเพื่อเลือกวันเกิด")
    }

    private var selectedDayLabel: String {
        days.first(where: { $0.0 == viewModel.dayOfBirth })?.1 ?? "เลือกวันเกิด"
    }

    private var verticalReplacement: AnyTransition {
        .asymmetric(
            insertion: .move(edge: .bottom).combined(with: .opacity),
            removal: .move(edge: .top).combined(with: .opacity)
        )
    }

    @ViewBuilder
    private var decodedAnalysisSection: some View {
        if !viewModel.decodedPairs.isEmpty {
            let refName = viewModel.referenceName.trimmingCharacters(in: .whitespacesAndNewlines)
            VStack(spacing: 0) {
                VStack(spacing: 10) {
                    HStack {
                        Text("“\(refName)”")
                            .font(.sarabun(15, weight: .bold))
                            .foregroundStyle(Color.brandGold)
                            .padding(.horizontal, 10)
                            .padding(.vertical, 3)
                            .background(Color.brandGold.opacity(0.18))
                            .clipShape(RoundedRectangle(cornerRadius: 6))
                            .overlay(
                                RoundedRectangle(cornerRadius: 6)
                                    .stroke(Color.brandGold.opacity(0.6), lineWidth: 1)
                            )
                    }
                    .frame(maxWidth: .infinity, alignment: .center)

                    HStack {
                        Spacer()
                        PairCircles(
                            pairs: decodedPairPresentation,
                            onPairTap: viewModel.fetchPairDetail
                        )
                        Spacer()
                    }

                    if let kalakini = viewModel.decodedKalakini, kalakini.applied {
                        let hasMatches = !kalakini.matches.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
                        let formattedMatches = kalakini.matches.filter { !$0.isWhitespace }.map(String.init).joined(separator: " ")
                        let message = hasMatches ? "พบกาลกิณีในชื่อ: \(formattedMatches)" : "ไม่พบกาลกิณีในชื่อ"
                        let textColor = hasMatches ? Color.orange.opacity(0.9) : Color.green.opacity(0.85)
                        Text(message)
                            .font(.sarabun(11, weight: .medium))
                            .foregroundStyle(textColor)
                            .frame(maxWidth: .infinity, alignment: .center)
                    }

                    if let profile = viewModel.decodedProfile4D {
                        NameProfile4DCard(
                            profile: profile,
                            onWeightsChanged: { n, sem, snd, sim in
                                viewModel.weights = RankingWeights(numerology: n, semantic: sem, sound: snd, similarity: sim)
                            }
                        )
                    }
                }
                .id(viewModel.decodedPairs.joined(separator: "-"))
                .padding(.horizontal, 6).padding(.vertical, 12)
                .background(
                    RoundedRectangle(cornerRadius: 12)
                        .fill(Color(hex: "#1F120B").opacity(0.92))
                )
                .padding(.horizontal, 0).padding(.vertical, 8)
                .transition(verticalReplacement)
            }
            .transition(verticalReplacement)
        }
    }

    private var decodedPairPresentation: [(value: String, isBad: Bool)] {
        viewModel.decodedPairs.enumerated().compactMap { index, pair in
            let trimmed = pair.trimmingCharacters(in: .whitespacesAndNewlines)
            guard !trimmed.isEmpty else { return nil }
            let type = index < viewModel.decodedPairTypes.count ? viewModel.decodedPairTypes[index].uppercased() : ""
            return (trimmed, type.hasPrefix("R"))
        }
    }

    private func modeButton(_ mode: SearchViewModel.SearchMode, _ title: String, _ icon: String) -> some View {
        let isSelected = viewModel.mode == mode
        return Button {
            withAnimation(.easeInOut(duration: 0.22)) { viewModel.setMode(mode) }
        } label: {
            HStack(spacing: 6) {
                Image(systemName: icon)
                    .font(.system(size: 13, weight: .bold))
                Text(title)
                    .font(.sarabun(13, weight: isSelected ? .bold : .medium))
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 10)
            .foregroundStyle(isSelected ? Color(hex: "#1F120B") : Color.brandCream.opacity(0.85))
            .background(
                Group {
                    if isSelected {
                        LinearGradient(
                            colors: [Color(hex: "#D0A12B"), Color(hex: "#E6C85A"), Color(hex: "#D0A12B")],
                            startPoint: .leading,
                            endPoint: .trailing
                        )
                    } else {
                        Color(hex: "#26160E").opacity(0.72)
                    }
                }
            )
            .clipShape(RoundedRectangle(cornerRadius: 10))
            .overlay(
                RoundedRectangle(cornerRadius: 10)
                    .stroke(isSelected ? Color.clear : Color.brandGold.opacity(0.18), lineWidth: 1)
            )
        }
        .buttonStyle(.plain)
    }

    @ViewBuilder private var wishSection: some View {
        VStack(alignment: .leading, spacing: 10) {
            StepHeadingView(
                number: "3",
                title: "เสริมความหมายมงคลที่ต้องการ",
                subtitle: "เลือกความหมายที่ต้องการหนุนนำ หรือสุ่มใหม่",
                isCompleted: step3Done,
                isActive: currentStep == 3,
                integratedWithSection: true
            )
            VStack(spacing: 0) {
                if !viewModel.decodedMeaning.isEmpty {
                    let isPrototypeSelected = viewModel.hasUserSelectedMeaning && viewModel.pinnedMeaning == viewModel.decodedMeaning
                    Button {
                        viewModel.pin(viewModel.decodedMeaning)
                    } label: {
                        HStack(alignment: .center, spacing: 9) {
                            Text(isPrototypeSelected ? "✅" : "🏷️")
                                .font(.system(size: 14))
                            VStack(alignment: .leading, spacing: 2) {
                                Text("ความหมายจากชื่อต้นแบบ\(viewModel.referenceName.isEmpty ? "" : " (\(viewModel.referenceName))")")
                                    .font(.sarabun(11, weight: .bold))
                                    .foregroundStyle(isPrototypeSelected ? Color(hex: "#FFDF85") : Color.brandGold)
                                let cleanProto = viewModel.decodedMeaning.trimmingCharacters(in: CharacterSet(charactersIn: "\"“”' "))
                                Text("“\(cleanProto)”")
                                    .font(.sarabun(isPrototypeSelected ? 14.5 : 13.5, weight: isPrototypeSelected ? .bold : .regular))
                                    .foregroundStyle(isPrototypeSelected ? Color(hex: "#FFDF85") : .white)
                                    .multilineTextAlignment(.leading)
                                    .lineLimit(2)
                            }
                            .frame(maxWidth: .infinity, minHeight: 44, alignment: .leading)
                        }
                        .frame(maxWidth: .infinity, minHeight: 48, alignment: .leading)
                        .padding(.horizontal, 14).padding(.vertical, 8)
                        .background(
                            isPrototypeSelected
                                ? LinearGradient(
                                    colors: [Color(hex: "#522E1A"), Color(hex: "#381F12"), Color(hex: "#26150C")],
                                    startPoint: .leading,
                                    endPoint: .trailing
                                )
                                : LinearGradient(
                                    colors: [Color(hex: "#26160E"), Color(hex: "#1F120B")],
                                    startPoint: .leading,
                                    endPoint: .trailing
                                )
                        )
                        .overlay(alignment: .leading) {
                            if isPrototypeSelected {
                                Rectangle()
                                    .fill(LinearGradient(colors: [Color(hex: "#FFDF85"), Color.brandGold], startPoint: .top, endPoint: .bottom))
                                    .frame(width: 4)
                            }
                        }
                        .contentShape(Rectangle())
                    }
                    .buttonStyle(.plain)
                    .accessibilityLabel("เลือกความหมายจากชื่อต้นแบบ \(viewModel.decodedMeaning)")
                }

                ForEach(Array(viewModel.suggestions.prefix(3).enumerated()), id: \.offset) { index, meaning in
                    let isSelected = viewModel.hasUserSelectedMeaning && viewModel.pinnedMeaning == meaning
                    let cleanMeaning = meaning.trimmingCharacters(in: CharacterSet(charactersIn: "\"“”' "))
                    Button { viewModel.pin(meaning) } label: {
                        HStack(alignment: .center, spacing: 9) {
                            Text(isSelected ? "✅" : "✨")
                                .font(.system(size: 14))
                            Text("“\(cleanMeaning)”")
                                .id("\(index)-\(meaning)")
                                .font(.sarabun(isSelected ? 14.5 : 13.5, weight: isSelected ? .bold : .regular))
                                .foregroundStyle(isSelected ? Color(hex: "#FFDF85") : .white)
                                .multilineTextAlignment(.leading)
                                .lineLimit(2)
                                .lineSpacing(1.5)
                                .truncationMode(.tail)
                                .frame(maxWidth: .infinity, minHeight: 44, alignment: .leading)
                                .transition(
                                    .asymmetric(
                                        insertion: .move(edge: .bottom).combined(with: .opacity),
                                        removal: .move(edge: .top).combined(with: .opacity)
                                    )
                                )
                        }
                        .frame(maxWidth: .infinity, minHeight: 46, alignment: .leading)
                        .padding(.horizontal, 14).padding(.vertical, 6)
                        .background(
                            isSelected
                                ? LinearGradient(
                                    colors: [Color(hex: "#522E1A"), Color(hex: "#381F12"), Color(hex: "#26150C")],
                                    startPoint: .leading,
                                    endPoint: .trailing
                                )
                                : LinearGradient(
                                    colors: index.isMultiple(of: 2)
                                        ? [Color(hex: "#26160E"), Color(hex: "#1F120B")]
                                        : [Color(hex: "#2C1A11"), Color(hex: "#22140D")],
                                    startPoint: .leading,
                                    endPoint: .trailing
                                )
                        )
                        .overlay(alignment: .leading) {
                            if isSelected {
                                Rectangle()
                                    .fill(LinearGradient(colors: [Color(hex: "#FFDF85"), Color.brandGold], startPoint: .top, endPoint: .bottom))
                                    .frame(width: 4)
                            }
                        }
                        .contentShape(Rectangle())
                    }
                    .buttonStyle(.plain)
                    .accessibilityLabel("เลือกความหมาย \(meaning)")
                }

                // Media-style Control Bar (Back, Play/Pause, Forward, Counter)
                HStack(spacing: 12) {
                    // Back button
                    Button(action: viewModel.previousMeaningPage) {
                        Image(systemName: "backward.fill")
                            .font(.system(size: 13, weight: .semibold))
                            .foregroundStyle(Color.brandGold)
                            .frame(width: 32, height: 32)
                            .background(Color(hex: "#2C1A11"))
                            .clipShape(Circle())
                            .overlay(Circle().stroke(Color.brandGold.opacity(0.35), lineWidth: 1))
                    }
                    .buttonStyle(.plain)
                    .accessibilityLabel("ชุดก่อนหน้า")

                    // Play / Pause button
                    Button(action: viewModel.toggleMeaningAutoPlay) {
                        HStack(spacing: 4) {
                            Image(systemName: viewModel.isMeaningAutoPlaying ? "pause.fill" : "play.fill")
                                .font(.system(size: 11, weight: .bold))
                            Text(viewModel.isMeaningAutoPlaying ? "หยุดเลื่อน" : "เลื่อนต่อ")
                                .font(.sarabun(11, weight: .semibold))
                        }
                        .foregroundStyle(Color.brandGold)
                        .padding(.horizontal, 10)
                        .padding(.vertical, 6)
                        .background(Color(hex: "#2C1A11"))
                        .clipShape(Capsule())
                        .overlay(Capsule().stroke(Color.brandGold.opacity(0.40), lineWidth: 1))
                    }
                    .buttonStyle(.plain)
                    .accessibilityLabel(viewModel.isMeaningAutoPlaying ? "หยุดเลื่อน" : "เลื่อนต่อ")

                    // Forward button
                    Button(action: viewModel.nextMeaningPage) {
                        Image(systemName: "forward.fill")
                            .font(.system(size: 13, weight: .semibold))
                            .foregroundStyle(Color.brandGold)
                            .frame(width: 32, height: 32)
                            .background(Color(hex: "#2C1A11"))
                            .clipShape(Circle())
                            .overlay(Circle().stroke(Color.brandGold.opacity(0.35), lineWidth: 1))
                    }
                    .buttonStyle(.plain)
                    .accessibilityLabel("ชุดถัดไป")

                    Spacer(minLength: 4)

                    // Page counter indicator
                    Text("ชุดที่ \(viewModel.meaningPageIndex + 1)/\(max(1, viewModel.totalMeaningPages))")
                        .font(.sarabun(11, weight: .medium))
                        .foregroundStyle(Color.brandGold.opacity(0.85))
                        .lineLimit(1)
                        .fixedSize(horizontal: true, vertical: false)
                }
                .padding(.horizontal, 12)
                .padding(.vertical, 8)
                .background(Color(hex: "#1F120B"))
            }
            .background(Color(hex: "#1F120B"))
            .clipShape(RoundedRectangle(cornerRadius: 14))
            .animation(.easeInOut(duration: 0.36), value: viewModel.pinnedMeaning)
            .animation(.easeInOut(duration: 0.42), value: viewModel.suggestions[safe: 0])
            .animation(.easeInOut(duration: 0.42).delay(0.08), value: viewModel.suggestions[safe: 1])
            .animation(.easeInOut(duration: 0.42).delay(0.16), value: viewModel.suggestions[safe: 2])
        }
        .padding(.horizontal, 8)
        .padding(.bottom, 8)
        .background(Color(hex: "#160C08"))
        .clipShape(RoundedRectangle(cornerRadius: 20))
        .overlay(RoundedRectangle(cornerRadius: 20).stroke(Color.brandGold.opacity(0.30), lineWidth: 1))
    }

    private var formula: some View {
        let refName = viewModel.referenceName.trimmingCharacters(in: .whitespacesAndNewlines)
        return VStack(alignment: .leading, spacing: 10) {
            StepHeadingView(
                number: "4",
                title: "ยืนยันโปรไฟล์ 4 มิติ & ค้นหาชื่อมงคล",
                isCompleted: step4Done,
                isActive: currentStep == 4,
                integratedWithSection: true
            )
            VStack(alignment: .leading, spacing: 8) {
                HStack {
                    Text("⚖️ ยืนยันโปรไฟล์ชื่อดี 4 มิติ").font(.sarabun(15, weight: .bold))
                    Spacer()
                    Button(action: viewModel.resetToBalance) {
                        HStack(spacing: 4) {
                            Text("↺").font(.sarabun(12, weight: .bold))
                            Text("คืนค่าจุดสมดุล").font(.sarabun(10, weight: .medium))
                        }
                        .foregroundStyle(Color.brandGold)
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .overlay(Capsule().stroke(Color.brandGold.opacity(0.45)))
                    }
                }
                if !refName.isEmpty {
                    Text("✨ อ้างอิงจากต้นแบบ “\(refName)”")
                        .font(.sarabun(11.5, weight: .semibold))
                        .foregroundStyle(Color.brandGold)
                }
                Text("แสดงโปรไฟล์เดียวกับผลวิเคราะห์ชื่อที่กรอก").font(.sarabun(11)).foregroundStyle(Color.brandCream.opacity(0.8))
                HStack {
                    let (numPct, semPct, resPct, sndPct) = dynamicWeightPercents
                    VStack(alignment: .leading, spacing: 10) {
                        weight("⭐", "เลขศาสตร์", numPct, .brandGold)
                        weight("🎯", "ความหมาย", semPct, .orange)
                        weight("🔍", "เอกลักษณ์", resPct, .cyan)
                        weight("🎶", "ไพเราะ", sndPct, .green)
                    }
                    .frame(width: 145)
                    RadarWeightSelector(weights: Binding(
                        get: { viewModel.weights },
                        set: { value in viewModel.weights = value }
                    ))
                }
                Divider().overlay(Color.brandGold.opacity(0.32)).padding(.top, 3)
                Text("พร้อมใช้ข้อมูลและสูตรนี้เพื่อจัดอันดับ")
                    .font(.sarabun(11, weight: .medium))
                    .foregroundStyle(Color.brandCream.opacity(0.76))
                    .frame(maxWidth: .infinity, alignment: .center)
                Button(action: viewModel.search) {
                    Group {
                        if viewModel.state == .loading {
                            HStack(spacing: 9) { ProgressView().tint(Color.brandBackground); Text("กำลังจัดอันดับจากข้อมูลทั้งหมด…") }
                        } else if viewModel.referenceName.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                            Label(viewModel.mode == .standalone ? "โปรดพิมพ์ชื่อ" : (viewModel.target == .surname ? "โปรดพิมพ์ชื่อตัว" : "โปรดพิมพ์นามสกุล"), systemImage: "pencil")
                        } else if viewModel.dayOfBirth.isEmpty {
                            Label("โปรดเลือกวันเกิด", systemImage: "calendar")
                        } else if !viewModel.hasUserSelectedMeaning {
                            Label("โปรดเลือกความหมาย", systemImage: "text.quote")
                        } else {
                            Label("ดูชื่อที่ดีที่สุดของคุณ", systemImage: "sparkles")
                        }
                    }
                    .font(.sarabun(17, weight: .bold)).frame(maxWidth: .infinity).padding(.vertical, 12)
                }
                .foregroundStyle(Color.brandBackground).background(Color.brandGold).clipShape(RoundedRectangle(cornerRadius: 18)).disabled(!viewModel.canSearch).opacity(viewModel.canSearch || viewModel.state == .loading ? 1 : 0.42)
            }
            .padding(10)
            .background(Color(hex: "#1F120B"))
            .clipShape(RoundedRectangle(cornerRadius: 14))
        }
        .padding(.horizontal, 8)
        .padding(.bottom, 8)
        .background(Color(hex: "#160C08"))
        .clipShape(RoundedRectangle(cornerRadius: 20))
        .overlay(RoundedRectangle(cornerRadius: 20).stroke(Color.brandGold.opacity(0.30), lineWidth: 1))
    }

    private func weight(_ icon: String, _ label: String, _ pct: Int, _ color: Color) -> some View {
        HStack {
            Text(icon)
            Text(label).font(.sarabun(13, weight: .semibold))
            Spacer()
            Text("\(pct)%").font(.sarabun(15, weight: .bold)).foregroundStyle(color)
        }
    }

    private var dynamicWeightPercents: (Int, Int, Int, Int) {
        let raw = [
            viewModel.weights.numerology,
            viewModel.weights.semantic,
            viewModel.weights.similarity,
            viewModel.weights.sound
        ]
        let total = max(0.0001, raw.reduce(0, +))
        let scaled = raw.map { ($0 / total) * 100.0 }
        var ints = scaled.map { Int(floor($0)) }
        var remainder = 100 - ints.reduce(0, +)
        let order = scaled.enumerated().sorted { lhs, rhs in
            (lhs.element - floor(lhs.element)) > (rhs.element - floor(rhs.element))
        }
        for entry in order where remainder > 0 {
            ints[entry.offset] += 1
            remainder -= 1
        }
        return (ints[0], ints[1], ints[2], ints[3])
    }

    @ViewBuilder private var searchStatus: some View {
        switch viewModel.state {
        case .idle: EmptyView()
        case .loading: EmptyView()
        case .failed(let message): Label(message, systemImage: "wifi.exclamationmark").foregroundStyle(.orange).padding()
        case .loaded: EmptyView()
        }
    }

    private var footerLinks: some View {
        HStack(spacing: 14) {
            Link("ข้อตกลงการใช้งาน", destination: URL(string: "https://xn--b3cu8e7ah6h.com/terms")!)
            Text("•").foregroundStyle(.white.opacity(0.3))
            Link("นโยบายความเป็นส่วนตัว", destination: URL(string: "https://xn--b3cu8e7ah6h.com/privacy")!)
        }
        .font(.sarabun(11))
        .foregroundStyle(Color.brandCream.opacity(0.65))
        .padding(.top, 12)
    }
}

private struct InteractiveCelebrityMarquee: View {
    let celebrities: [Celebrity]
    let selectedName: String
    let avatarURL: (Celebrity) -> URL?
    let onSelect: (Celebrity) -> Void

    private let itemWidth: CGFloat = 80
    private let repeatCount = 4
    @State private var isAutoScrolling = true

    var body: some View {
        ContinuousHorizontalScrollView(
            cycleWidth: itemWidth * CGFloat(celebrities.count),
            cycleCount: repeatCount,
            isAutoScrolling: isAutoScrolling
        ) {
            HStack(spacing: 0) {
                ForEach(0..<repeatCount, id: \.self) { copy in
                    celebrityTrack(copy: copy)
                }
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 4)
        }
        .frame(height: 92)
        .accessibilityHint("ปัดไปทางซ้ายหรือขวาเพื่อดูรายชื่อเพิ่มเติม")
    }

    private func celebrityTrack(copy: Int) -> some View {
        HStack(spacing: 0) {
            ForEach(Array(celebrities.enumerated()), id: \.element.id) { index, celebrity in
                Button {
                    isAutoScrolling = false
                    onSelect(celebrity)
                } label: {
                    VStack(spacing: 6) {
                        RemoteAvatarImage(url: avatarURL(celebrity)) {
                            ZStack {
                                avatarColor(index).opacity(0.22)
                                Text(String(celebrity.name.prefix(1)))
                                    .font(.sarabun(14, weight: .bold))
                                    .foregroundStyle(.white)
                            }
                        }
                        .frame(width: 60, height: 60)
                        .clipShape(Circle())
                        .overlay(
                            Circle()
                                .stroke(
                                    selectedName == celebrity.name ? Color.brandGold : avatarColor(index),
                                    lineWidth: selectedName == celebrity.name ? 3.5 : 2.5
                                )
                        )
                        Text(celebrity.name.split(separator: " ").first.map(String.init) ?? celebrity.name)
                            .font(.sarabun(14, weight: .semibold))
                            .foregroundStyle(selectedName == celebrity.name ? Color.brandGold : Color.brandCream)
                            .lineLimit(1)
                    }
                    .frame(width: itemWidth)
                }
                .buttonStyle(.plain)
                .id(copy * celebrities.count + index)
            }
        }
    }

    private func avatarColor(_ index: Int) -> Color {
        let colors: [Color] = [
            Color(red: 0.90, green: 0.45, blue: 0.45),
            Color(red: 0.35, green: 0.78, blue: 0.45),
            Color(red: 0.25, green: 0.65, blue: 0.96),
            Color(red: 1.00, green: 0.65, blue: 0.20),
            Color(red: 0.78, green: 0.35, blue: 0.85)
        ]
        return colors[index % colors.count]
    }
}

private struct ContinuousHorizontalScrollView<Content: View>: UIViewRepresentable {
    let cycleWidth: CGFloat
    let cycleCount: Int
    let isAutoScrolling: Bool
    let content: Content

    init(cycleWidth: CGFloat, cycleCount: Int, isAutoScrolling: Bool, @ViewBuilder content: () -> Content) {
        self.cycleWidth = cycleWidth
        self.cycleCount = cycleCount
        self.isAutoScrolling = isAutoScrolling
        self.content = content()
    }

    func makeCoordinator() -> Coordinator { Coordinator() }

    func makeUIView(context: Context) -> UIScrollView {
        let scrollView = UIScrollView()
        scrollView.showsHorizontalScrollIndicator = false
        scrollView.showsVerticalScrollIndicator = false
        scrollView.alwaysBounceHorizontal = true
        scrollView.alwaysBounceVertical = false
        scrollView.delegate = context.coordinator

        let hostingController = UIHostingController(rootView: content)
        hostingController.view.backgroundColor = .clear
        scrollView.addSubview(hostingController.view)
        context.coordinator.hostingController = hostingController
        context.coordinator.configure(scrollView: scrollView, cycleWidth: cycleWidth, cycleCount: cycleCount, isAutoScrolling: isAutoScrolling)
        return scrollView
    }

    func updateUIView(_ scrollView: UIScrollView, context: Context) {
        context.coordinator.hostingController?.rootView = content
        context.coordinator.configure(scrollView: scrollView, cycleWidth: cycleWidth, cycleCount: cycleCount, isAutoScrolling: isAutoScrolling)
    }

    static func dismantleUIView(_ uiView: UIScrollView, coordinator: Coordinator) { coordinator.stop() }

    final class Coordinator: NSObject, UIScrollViewDelegate {
        var hostingController: UIHostingController<Content>?
        private weak var scrollView: UIScrollView?
        private var displayLink: CADisplayLink?
        private var cycleWidth: CGFloat = 0
        private var cycleCount = 0
        private var isAutoScrolling = true
        private var lastTimestamp: CFTimeInterval = 0
        private var resumeAt: CFTimeInterval = 0

        func configure(scrollView: UIScrollView, cycleWidth: CGFloat, cycleCount: Int, isAutoScrolling: Bool) {
            guard cycleWidth > 0 else { return }
            self.scrollView = scrollView
            self.cycleWidth = cycleWidth
            self.cycleCount = cycleCount
            self.isAutoScrolling = isAutoScrolling
            let contentWidth = cycleWidth * CGFloat(cycleCount) + 32
            hostingController?.view.frame = CGRect(x: 0, y: 0, width: contentWidth, height: 92)
            scrollView.contentSize = CGSize(width: contentWidth, height: 92)
            if scrollView.contentOffset.x == 0 { scrollView.contentOffset.x = cycleWidth + 16 }
            start()
        }

        func start() {
            guard displayLink == nil else { return }
            let displayLink = CADisplayLink(target: self, selector: #selector(advance))
            displayLink.add(to: .main, forMode: .common)
            self.displayLink = displayLink
        }

        func stop() { displayLink?.invalidate(); displayLink = nil }

        @objc private func advance(_ displayLink: CADisplayLink) {
            guard let scrollView, cycleWidth > 0, isAutoScrolling,
                  !scrollView.isDragging, !scrollView.isDecelerating,
                  displayLink.timestamp >= resumeAt else {
                lastTimestamp = displayLink.timestamp
                return
            }
            let elapsed = lastTimestamp == 0 ? 0 : displayLink.timestamp - lastTimestamp
            lastTimestamp = displayLink.timestamp
            scrollView.contentOffset.x += CGFloat(elapsed * 30)
            let wrapPoint = cycleWidth * 2 + 16
            if scrollView.contentOffset.x >= wrapPoint { scrollView.contentOffset.x -= cycleWidth }
        }

        func scrollViewWillBeginDragging(_ scrollView: UIScrollView) { resumeAt = .greatestFiniteMagnitude }
        func scrollViewDidEndDragging(_ scrollView: UIScrollView, willDecelerate decelerate: Bool) {
            if !decelerate { resumeAfterPause() }
        }
        func scrollViewDidEndDecelerating(_ scrollView: UIScrollView) { resumeAfterPause() }
        private func resumeAfterPause() { resumeAt = CACurrentMediaTime() + 4 }
    }
}

private extension Collection {
    subscript(safe index: Index) -> Element? {
        indices.contains(index) ? self[index] : nil
    }
}

private struct NameProfile4DCard: View {
    let profile: NameProfile4D
    var onWeightsChanged: ((Double, Double, Double, Double) -> Void)? = nil

    @State private var activeWeights: (Double, Double, Double, Double)? = nil

    var body: some View {
        let numRaw = Double(max(0, min(100, profile.numerologyScore)))
        let semRaw = Double(max(0, min(100, profile.meaningScore)))
        let resRaw = Double(max(0, min(100, profile.resonanceScore)))
        let sndRaw = Double(max(0, min(100, profile.soundScore)))

        let total = max(0.001, numRaw + semRaw + resRaw + sndRaw)
        let cur = activeWeights ?? (numRaw / total, semRaw / total, sndRaw / total, resRaw / total)

        let numPct = Int((cur.0 * 100.0).rounded())
        let semPct = Int((cur.1 * 100.0).rounded())
        let sndPct = Int((cur.2 * 100.0).rounded())
        let resPct = max(0, 100 - numPct - semPct - sndPct)

        return VStack(alignment: .leading, spacing: 8) {
            Divider().overlay(Color.white.opacity(0.22))

            HStack {
                HStack(spacing: 6) {
                    Text("📊").font(.system(size: 13))
                    Text("โปรไฟล์ชื่อดี 4D")
                        .font(.sarabun(12.5, weight: .semibold))
                        .foregroundStyle(Color.white)
                }
                Spacer()
                Text(onWeightsChanged != nil ? "แตะ/ลากเพื่อเลือกความสำคัญ ✨" : "ค่าประเมินเบื้องต้น")
                    .font(.sarabun(11, weight: .regular))
                    .foregroundStyle(onWeightsChanged != nil ? Color.brandGold : Color.white.opacity(0.80))
            }

            HStack(alignment: .center, spacing: 6) {
                VStack(alignment: .leading, spacing: 5) {
                    NameProfileDimensionRow(
                        icon: "⭐",
                        name: "เลขศาสตร์",
                        score: numPct,
                        accentColor: Color(red: 0.99, green: 0.88, blue: 0.28)
                    )

                    NameProfileDimensionRow(
                        icon: "🎯",
                        name: "ความหมาย",
                        score: semPct,
                        accentColor: Color(red: 0.95, green: 0.73, blue: 0.29)
                    )

                    NameProfileDimensionRow(
                        icon: "🔍",
                        name: "เอกลักษณ์",
                        score: resPct,
                        accentColor: Color(red: 0.30, green: 0.71, blue: 0.67)
                    )

                    NameProfileDimensionRow(
                        icon: "🎶",
                        name: "ไพเราะ",
                        score: sndPct,
                        accentColor: Color(red: 0.49, green: 0.83, blue: 0.66)
                    )
                }

                // Right Side: Interactive 4D Diamond Radar Shape (Fills available space)
                NameProfileRadarShapeView(
                    wNum: CGFloat(cur.0),
                    wSem: CGFloat(cur.1),
                    wRes: CGFloat(cur.3),
                    wSnd: CGFloat(cur.2),
                    onWeightsChanged: { n, sem, snd, sim in
                        activeWeights = (n, sem, snd, sim)
                        onWeightsChanged?(n, sem, snd, sim)
                    }
                )
                .frame(height: 145)
                .frame(maxWidth: .infinity)
            }
            .padding(.vertical, 2)
        }
        .padding(.top, 4)
        .onChange(of: profile) {
            activeWeights = nil
        }
    }
}

private struct NameProfileDimensionRow: View {
    let icon: String
    let name: String
    let score: Int
    let accentColor: Color

    var body: some View {
        HStack(spacing: 6) {
            Text(icon)
                .font(.system(size: 14))

            Text(name)
                .font(.sarabun(12, weight: .medium))
                .foregroundStyle(Color.white)
                .frame(width: 58, alignment: .leading)

            Text("\(score)%")
                .font(.sarabun(12.5, weight: .bold))
                .foregroundStyle(accentColor)
                .frame(width: 36, alignment: .leading)
        }
    }
}

private struct NameProfileRadarShapeView: View {
    let wNum: CGFloat
    let wSem: CGFloat
    let wRes: CGFloat
    let wSnd: CGFloat
    var onWeightsChanged: ((Double, Double, Double, Double) -> Void)? = nil

    var body: some View {
        GeometryReader { geometry in
            let size = geometry.size
            let w = size.width
            let h = size.height
            let cx = w / 2.0
            let cy = h / 2.0
            let rx = max(1.0, (w / 2.0) - 18.0)
            let ry = max(1.0, (h / 2.0) - 18.0)

            let topPt = CGPoint(x: cx, y: cy - ry)
            let bottomPt = CGPoint(x: cx, y: cy + ry)
            let leftPt = CGPoint(x: cx - rx, y: cy)
            let rightPt = CGPoint(x: cx + rx, y: cy)

            let sparkX = (topPt.x * wNum) + (bottomPt.x * wSnd) + (leftPt.x * wSem) + (rightPt.x * wRes)
            let sparkY = (topPt.y * wNum) + (bottomPt.y * wSnd) + (leftPt.y * wSem) + (rightPt.y * wRes)
            let sparkPoint = CGPoint(x: sparkX, y: sparkY)

            ZStack {
                // 1. Diamond Background Fill
                Path { path in
                    path.move(to: topPt)
                    path.addLine(to: rightPt)
                    path.addLine(to: bottomPt)
                    path.addLine(to: leftPt)
                    path.closeSubpath()
                }
                .fill(
                    LinearGradient(
                        colors: [
                            Color(red: 0.99, green: 0.88, blue: 0.28).opacity(0.32),
                            Color(hex: "#382014").opacity(0.50)
                        ],
                        startPoint: .top,
                        endPoint: .bottom
                    )
                )

                // 2. Diamond Frame Stroke
                Path { path in
                    path.move(to: topPt)
                    path.addLine(to: rightPt)
                    path.addLine(to: bottomPt)
                    path.addLine(to: leftPt)
                    path.closeSubpath()
                }
                .stroke(
                    LinearGradient(
                        colors: [
                            Color(red: 0.99, green: 0.88, blue: 0.28),
                            Color(red: 0.52, green: 0.80, blue: 0.09)
                        ],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    ),
                    lineWidth: 2.0
                )

                // 3. Crosshair Lines
                Path { path in
                    path.move(to: topPt); path.addLine(to: bottomPt)
                    path.move(to: leftPt); path.addLine(to: rightPt)
                }
                .stroke(Color.white.opacity(0.22), lineWidth: 1.2)

                // 4. Corner Icons
                Text("⭐")
                    .font(.system(size: 15))
                    .position(x: topPt.x, y: topPt.y - 10)
                Text("🎯")
                    .font(.system(size: 15))
                    .position(x: leftPt.x - 10, y: leftPt.y)
                Text("🔍")
                    .font(.system(size: 15))
                    .position(x: rightPt.x + 10, y: rightPt.y)
                Text("🎶")
                    .font(.system(size: 15))
                    .position(x: bottomPt.x, y: bottomPt.y + 10)

                // 5. Corner Dots
                Circle().fill(Color(red: 0.99, green: 0.88, blue: 0.28)).frame(width: 5.6, height: 5.6).position(topPt)
                Circle().fill(Color(red: 0.30, green: 0.71, blue: 0.67)).frame(width: 5.6, height: 5.6).position(rightPt)
                Circle().fill(Color(red: 0.49, green: 0.83, blue: 0.66)).frame(width: 5.6, height: 5.6).position(bottomPt)
                Circle().fill(Color(red: 0.95, green: 0.73, blue: 0.29)).frame(width: 5.6, height: 5.6).position(leftPt)

                // 6. Moving Center Spark Indicator Dot
                Circle()
                    .fill(
                        RadialGradient(
                            colors: [
                                Color(red: 1.0, green: 0.97, blue: 0.78).opacity(0.90),
                                Color(red: 0.99, green: 0.88, blue: 0.28).opacity(0.45),
                                .clear
                            ],
                            center: .center,
                            startRadius: 2,
                            endRadius: 18
                        )
                    )
                    .frame(width: 36, height: 36)
                    .position(sparkPoint)

                Circle()
                    .stroke(Color(red: 0.99, green: 0.88, blue: 0.28).opacity(0.55), lineWidth: 1.5)
                    .frame(width: 22, height: 22)
                    .position(sparkPoint)

                Circle()
                    .fill(Color.white)
                    .frame(width: 12, height: 12)
                    .position(sparkPoint)

                Circle()
                    .fill(Color(red: 0.96, green: 0.62, blue: 0.04))
                    .frame(width: 6.4, height: 6.4)
                    .position(sparkPoint)
            }
            .contentShape(Rectangle())
            .gesture(
                DragGesture(minimumDistance: 0)
                    .onChanged { value in
                        guard let onWeightsChanged else { return }
                        let (num, sem, snd, sim) = diamondWeights(at: value.location, in: size)
                        onWeightsChanged(num, sem, snd, sim)
                    }
            )
        }
    }

    private func diamondWeights(at point: CGPoint, in size: CGSize) -> (Double, Double, Double, Double) {
        let cx = size.width / 2.0
        let cy = size.height / 2.0
        let rx = max(1.0, cx - 18.0)
        let ry = max(1.0, cy - 18.0)

        var nx = Double((point.x - cx) / rx)
        var ny = Double((point.y - cy) / ry)

        let scale = abs(nx) + abs(ny)
        if scale > 1.0 {
            nx /= scale
            ny /= scale
        }

        let u = (nx + ny + 1.0) / 2.0
        let v = (nx - ny + 1.0) / 2.0

        let sound = max(0.0, min(1.0, u * (1.0 - v)))
        let numerology = max(0.0, min(1.0, (1.0 - u) * v))
        let semantic = max(0.0, min(1.0, (1.0 - u) * (1.0 - v)))
        let similarity = max(0.0, min(1.0, u * v))

        let sum = numerology + semantic + sound + similarity
        guard sum > 0 else { return (0.35, 0.25, 0.20, 0.20) }
        return (numerology / sum, semantic / sum, sound / sum, similarity / sum)
    }
}

struct StepProgressBar: View {
    let currentStep: Int
    let step1Done: Bool
    let step2Done: Bool
    let step3Done: Bool
    let step4Done: Bool
    var onReset: (() -> Void)? = nil
    var onStepClick: ((Int) -> Void)? = nil
    @State private var isPulsing = false

    private var steps: [(Int, String, Bool)] {
        [
            (1, "ระบุชื่อ", step1Done),
            (2, "วันเกิด", step2Done),
            (3, "ความหมาย", step3Done),
            (4, "ค้นหา 4 มิติ", step4Done)
        ]
    }

    var body: some View {
        VStack(spacing: 8) {
            HStack {
                HStack(spacing: 6) {
                    Image(systemName: "sparkles")
                        .font(.system(size: 13, weight: .bold))
                        .foregroundStyle(Color.brandGold)
                    Text("4 ขั้นตอนการตั้งชื่อมงคล")
                        .font(.sarabun(12.5, weight: .bold))
                        .foregroundStyle(Color.brandCream)
                }
                Spacer()
                HStack(spacing: 7) {
                    Text("ขั้นตอนที่ \(currentStep) / 4")
                        .font(.sarabun(11, weight: .semibold))
                        .foregroundStyle(Color.brandGold)

                    if let onReset = onReset {
                        Button(action: onReset) {
                            HStack(spacing: 3.5) {
                                Image(systemName: "arrow.counterclockwise")
                                    .font(.system(size: 9.5, weight: .bold))
                                Text("รีเซ็ต")
                                    .font(.sarabun(10.5, weight: .semibold))
                            }
                            .foregroundStyle(Color.brandGold)
                            .padding(.horizontal, 7)
                            .padding(.vertical, 2.5)
                            .background(Color.white.opacity(0.08))
                            .clipShape(Capsule())
                            .overlay(
                                Capsule()
                                    .stroke(Color.brandGold.opacity(0.32), lineWidth: 0.8)
                            )
                        }
                        .buttonStyle(.plain)
                    }
                }
            }

            let completedGoldGradient = LinearGradient(
                colors: [Color(hex: "#FFF0A3"), Color(hex: "#D4AF37"), Color(hex: "#A37C10")],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
            let completedLineGradient = LinearGradient(
                colors: [Color.brandGold, Color(hex: "#FFF0A3"), Color.brandGold],
                startPoint: .leading,
                endPoint: .trailing
            )
            let activeGold = Color.brandGold
            let inactiveLineColor = Color.white.opacity(0.12)

            HStack(spacing: 0) {
                ForEach(Array(steps.enumerated()), id: \.offset) { index, item in
                    let stepNum = item.0
                    let title = item.1
                    let isDone = item.2
                    let isActive = currentStep == stepNum

                    VStack(spacing: 4) {
                        ZStack {
                            // Left and Right connector lines
                            HStack(spacing: 0) {
                                if index > 0 {
                                    let prevDone = steps[index - 1].2
                                    ZStack(alignment: .leading) {
                                        Rectangle()
                                            .fill(inactiveLineColor)
                                            .frame(height: 2.5)
                                        Rectangle()
                                            .fill(completedLineGradient)
                                            .frame(height: 2.5)
                                            .shadow(color: Color.brandGold.opacity(0.55), radius: 2, x: 0, y: 0)
                                            .opacity(prevDone ? 1 : 0)
                                    }
                                } else {
                                    Spacer()
                                }

                                if index < steps.count - 1 {
                                    ZStack(alignment: .leading) {
                                        Rectangle()
                                            .fill(inactiveLineColor)
                                            .frame(height: 2.5)
                                        Rectangle()
                                            .fill(completedLineGradient)
                                            .frame(height: 2.5)
                                            .shadow(color: Color.brandGold.opacity(0.55), radius: 2, x: 0, y: 0)
                                            .opacity(isDone ? 1 : 0)
                                    }
                                } else {
                                    Spacer()
                                }
                            }

                            // Sparkling aura halo (fixed frame, soft light intensity breathing)
                            if isActive {
                                Circle()
                                    .fill(
                                        RadialGradient(
                                            colors: [
                                                activeGold.opacity(isPulsing ? 0.70 : 0.30),
                                                Color(hex: "#FFDF85").opacity(isPulsing ? 0.35 : 0.10),
                                                .clear
                                            ],
                                            center: .center,
                                            startRadius: 5,
                                            endRadius: 17
                                        )
                                    )
                                    .frame(width: 34, height: 34)
                            } else if isDone {
                                Circle()
                                    .fill(
                                        RadialGradient(colors: [Color.brandGold.opacity(0.42), Color(hex: "#FFE082").opacity(0.18), .clear], center: .center, startRadius: 6, endRadius: 18)
                                    )
                                    .frame(width: 34, height: 34)
                            }

                            Group {
                                if isDone {
                                    Circle()
                                        .fill(completedGoldGradient)
                                        .frame(width: 24, height: 24)
                                        .overlay(
                                            Circle()
                                                .stroke(Color(hex: "#FFF7D6"), lineWidth: 1.5)
                                        )
                                        .shadow(color: Color.brandGold.opacity(0.55), radius: 3, x: 0, y: 1)
                                } else if isActive {
                                    Circle()
                                        .fill(
                                            RadialGradient(
                                                colors: [Color(hex: "#4D3318"), Color(hex: "#2B1A0E")],
                                                center: .center,
                                                startRadius: 2,
                                                endRadius: 12
                                            )
                                        )
                                        .frame(width: 24, height: 24)
                                        .overlay(
                                            Circle()
                                                .stroke(activeGold, lineWidth: 1.8)
                                        )
                                        .shadow(color: activeGold.opacity(isPulsing ? 0.85 : 0.35), radius: 4, x: 0, y: 0)
                                } else {
                                    Circle()
                                        .fill(Color(hex: "#26160E"))
                                        .frame(width: 24, height: 24)
                                        .overlay(
                                            Circle()
                                                .stroke(Color.white.opacity(0.18), lineWidth: 1)
                                        )
                                }
                            }

                            if isDone {
                                Image(systemName: "checkmark")
                                    .font(.system(size: 11, weight: .black))
                                    .foregroundStyle(Color(hex: "#26160E"))
                                    .transition(.scale.combined(with: .opacity))
                            } else {
                                Text("\(stepNum)")
                                    .font(.sarabun(11, weight: isActive ? .black : .bold))
                                    .foregroundStyle(isActive ? Color(hex: "#FFF7D6") : Color.white.opacity(0.52))
                                    .transition(.scale.combined(with: .opacity))
                            }
                        }
                        // Reserve the halo height for every step, so connector
                        // segments always share exactly the same centerline.
                        .frame(height: 32)

                        Text(title)
                            .font(.sarabun(10, weight: isActive || isDone ? .bold : .regular))
                            .foregroundStyle(isDone ? Color(hex: "#FFDF85") : (isActive ? activeGold : Color.white.opacity(0.45)))
                            .lineLimit(1)
                            .animation(.easeInOut(duration: 0.3), value: isActive)
                            .animation(.easeInOut(duration: 0.3), value: isDone)
                    }
                    .frame(maxWidth: .infinity)
                    .contentShape(Rectangle())
                    .onTapGesture {
                        onStepClick?(stepNum)
                    }
                    .accessibilityElement(children: .combine)
                    .accessibilityAddTraits(.isButton)
                    .accessibilityLabel("ขั้นตอนที่ \(stepNum): \(title)")
                }
            }
            .overlay {
                TimelineView(.animation) { timeline in
                    GeometryReader { proxy in
                        let completedSegments = [step1Done, step2Done, step3Done].prefix { $0 }.count
                        if completedSegments > 0 {
                            let centers = (0..<4).map { proxy.size.width * (CGFloat($0) + 0.5) / 4 }
                            let trackStart = centers[completedSegments - 1] + 14
                            let trackEnd = centers[completedSegments] - 14
                            let cycle = timeline.date.timeIntervalSinceReferenceDate
                                .truncatingRemainder(dividingBy: 1.25) / 1.25
                            // Reset outside the segment so one comet always travels left to right.
                            let travel = -0.16 + (cycle * 1.32)
                            let cometX = trackStart + ((trackEnd - trackStart) * travel)

                            ZStack {
                                Capsule()
                                    .fill(LinearGradient(colors: [.clear, Color.brandGold.opacity(0.30)], startPoint: .leading, endPoint: .trailing))
                                    .frame(width: 22, height: 3)
                                    .position(x: cometX - 11, y: 16)
                                Circle()
                                    .fill(Color.brandCream)
                                    .frame(width: 7, height: 7)
                                    .shadow(color: Color.brandGold, radius: 5)
                                    .position(x: cometX, y: 16)
                            }
                            .opacity((0...1).contains(travel) ? 1 : 0)
                        }
                    }
                }
                .allowsHitTesting(false)
            }
        }
        .padding(.horizontal, 14)
        .padding(.vertical, 10)
        .background(Color(hex: "#1F120B"))
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .overlay(
            RoundedRectangle(cornerRadius: 14)
                .stroke(Color.brandGold.opacity(0.35), lineWidth: 1)
        )
        .onAppear {
            withAnimation(Animation.easeInOut(duration: 1.0).repeatForever(autoreverses: true)) {
                isPulsing = true
            }
        }
    }
}

struct MagicSuggestionsSkeletonView: View {
    @State private var phase: CGFloat = -1.0

    private let shimmerColors = [
        Color(hex: "#1F120B"),
        Color(hex: "#382014"),
        Color.brandGold.opacity(0.42),
        Color(hex: "#382014"),
        Color(hex: "#1F120B")
    ]

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 6) {
                Image(systemName: "sparkles")
                    .font(.system(size: 12, weight: .bold))
                    .foregroundStyle(Color.brandGold)
                Text("กำลังคำนวณและตรวจคู่เลขชื่อใกล้เคียง...")
                    .font(.sarabun(12, weight: .semibold))
                    .foregroundStyle(Color.brandGold.opacity(0.85))
            }
            .padding(.horizontal, 2)
            .padding(.vertical, 2)

            ForEach([86, 64, 98], id: \.self) { nameWidth in
                HStack(spacing: 10) {
                    RoundedRectangle(cornerRadius: 3)
                        .fill(shimmerGradient)
                        .frame(width: 14, height: 14)

                    RoundedRectangle(cornerRadius: 4)
                        .fill(shimmerGradient)
                        .frame(width: CGFloat(nameWidth), height: 13)

                    Spacer(minLength: 0)

                    HStack(spacing: 6) {
                        Circle()
                            .fill(shimmerGradient)
                            .frame(width: 28, height: 28)
                            .overlay(Circle().stroke(Color.brandGold.opacity(0.35), lineWidth: 1))

                        Circle()
                            .fill(shimmerGradient)
                            .frame(width: 28, height: 28)
                            .overlay(Circle().stroke(Color.brandGold.opacity(0.35), lineWidth: 1))
                    }
                }
                .padding(.horizontal, 12)
                .padding(.vertical, 8)
                .background(Color(hex: "#26160E"))
                .clipShape(RoundedRectangle(cornerRadius: 10))
                .overlay(
                    RoundedRectangle(cornerRadius: 10)
                        .stroke(Color.brandGold.opacity(0.22), lineWidth: 0.7)
                )
            }
        }
        .onAppear {
            withAnimation(.linear(duration: 1.35).repeatForever(autoreverses: false)) {
                phase = 2.0
            }
        }
    }

    private var shimmerGradient: LinearGradient {
        LinearGradient(
            colors: shimmerColors,
            startPoint: UnitPoint(x: phase - 0.5, y: 0.0),
            endPoint: UnitPoint(x: phase + 0.5, y: 1.0)
        )
    }
}

struct StepHeaderPatternView: View {
    var isCompleted: Bool = false

    var body: some View {
        Canvas { context, size in
            let step: CGFloat = 22
            let halfStep: CGFloat = step / 2
            let totalW = size.width
            let totalH = size.height
            guard totalW > 0, totalH > 0 else { return }

            let gold = Color(hex: "#D4AF37")
            let lineAlpha = isCompleted ? 0.13 : 0.08
            let starAlpha = isCompleted ? 0.22 : 0.14
            let dotAlpha  = isCompleted ? 0.16 : 0.10

            // 1. Diagonal grid lines anchored to origin (0, 0)
            var gridPath = Path()

            // Slope +1 (angle 45°): starts at (k*step, 0), ends at (k*step + totalH, totalH)
            let minK = Int(floor(-totalH / step)) - 1
            let maxK = Int(ceil(totalW / step)) + 1
            for k in minK...maxK {
                let x0 = CGFloat(k) * step
                gridPath.move(to: CGPoint(x: x0, y: 0))
                gridPath.addLine(to: CGPoint(x: x0 + totalH, y: totalH))
            }

            // Slope -1 (angle -45°): starts at (m*step, 0), ends at (m*step - totalH, totalH)
            let minM = -1
            let maxM = Int(ceil((totalW + totalH) / step)) + 1
            for m in minM...maxM {
                let x0 = CGFloat(m) * step
                gridPath.move(to: CGPoint(x: x0, y: 0))
                gridPath.addLine(to: CGPoint(x: x0 - totalH, y: totalH))
            }
            context.stroke(gridPath, with: .color(gold.opacity(lineAlpha)), lineWidth: 1.0)

            // 2. Intersections (Thai 4-pointed Pra-jam-yam star) & cell centers (dot)
            var starPath = Path()
            var dotPath = Path()
            let starR: CGFloat = 3.2
            let innerR: CGFloat = 1.2
            let dotR: CGFloat = 1.1

            let maxRow = Int(ceil(totalH / halfStep)) + 1
            let minCol = -1
            let maxCol = Int(ceil(totalW / step)) + 1

            for row in 0...maxRow {
                let cy = CGFloat(row) * halfStep
                let isOdd = (row % 2 != 0)

                // Intersections at (cx, cy)
                let xOffset = isOdd ? halfStep : 0
                for col in minCol...maxCol {
                    let cx = CGFloat(col) * step + xOffset
                    if cx >= -starR && cx <= totalW + starR {
                        starPath.move(to: CGPoint(x: cx, y: cy - starR))
                        starPath.addLine(to: CGPoint(x: cx + innerR, y: cy - innerR))
                        starPath.addLine(to: CGPoint(x: cx + starR, y: cy))
                        starPath.addLine(to: CGPoint(x: cx + innerR, y: cy + innerR))
                        starPath.addLine(to: CGPoint(x: cx, y: cy + starR))
                        starPath.addLine(to: CGPoint(x: cx - innerR, y: cy + innerR))
                        starPath.addLine(to: CGPoint(x: cx - starR, y: cy))
                        starPath.addLine(to: CGPoint(x: cx - innerR, y: cy - innerR))
                        starPath.closeSubpath()
                    }
                }

                // Cell centers at (dotX, cy)
                let dotOffset = isOdd ? 0 : halfStep
                for col in minCol...maxCol {
                    let dotX = CGFloat(col) * step + dotOffset
                    if dotX >= -dotR && dotX <= totalW + dotR {
                        dotPath.addEllipse(in: CGRect(x: dotX - dotR, y: cy - dotR, width: dotR * 2, height: dotR * 2))
                    }
                }
            }

            context.fill(starPath, with: .color(gold.opacity(starAlpha)))
            context.fill(dotPath, with: .color(gold.opacity(dotAlpha)))
        }
        .allowsHitTesting(false)
    }
}

struct StepHeadingView: View {
    let number: String
    let title: String
    var subtitle: String? = nil
    var isCompleted: Bool = false
    var isActive: Bool = false
    var integratedWithSection: Bool = false
    @State private var isPulsing = false

    var body: some View {
        HStack(alignment: .center, spacing: 10) {
            ZStack {
                if isActive && !isCompleted {
                    // Glowing beacon aura (fixed frame, soft golden light intensity)
                    Circle()
                        .fill(
                            RadialGradient(
                                colors: [Color.brandGold.opacity(isPulsing ? 0.65 : 0.25), Color(hex: "#FFE082").opacity(isPulsing ? 0.25 : 0.05), .clear],
                                center: .center,
                                startRadius: 5,
                                endRadius: 18
                            )
                        )
                        .frame(width: 36, height: 36)
                }
                if isCompleted {
                    Circle()
                        .fill(
                            LinearGradient(
                                colors: [Color(hex: "#FFF0A3"), Color(hex: "#D4AF37"), Color(hex: "#A37C10")],
                                startPoint: .topLeading,
                                endPoint: .bottomTrailing
                            )
                        )
                        .frame(width: 28, height: 28)
                        .overlay(
                            Circle()
                                .stroke(Color(hex: "#FFF7D6"), lineWidth: 1.2)
                        )
                        .shadow(color: Color.brandGold.opacity(0.50), radius: 3, x: 0, y: 1)
                } else {
                    Circle()
                        .fill(isActive ? Color(hex: "#4D2D1B") : Color(hex: "#3D2418"))
                        .frame(width: 28, height: 28)
                        .overlay(
                            Circle()
                                .stroke(isActive ? Color.brandGold : Color.brandGold.opacity(0.6), lineWidth: isActive ? 1.8 : 1.2)
                        )
                        .shadow(color: isActive ? Color.brandGold.opacity(isPulsing ? 0.8 : 0.2) : .clear, radius: 4)
                }

                if isCompleted {
                    Image(systemName: "checkmark")
                        .font(.system(size: 13, weight: .black))
                        .foregroundStyle(Color(hex: "#26160E"))
                } else {
                    Text(number)
                        .font(.sarabun(13, weight: isActive ? .black : .bold))
                        .foregroundStyle(isActive ? Color(hex: "#FFF7D6") : Color.brandCream)
                }
            }

            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .font(.prompt(14.5, weight: .bold))
                    .foregroundStyle(.white)
                if let subtitle, !subtitle.isEmpty {
                    Text(subtitle)
                        .font(.sarabun(11.5))
                        .foregroundStyle(Color.brandCream.opacity(0.75))
                        .lineLimit(1)
                        .minimumScaleFactor(0.85)
                }
            }
            Spacer(minLength: 0)
        }
        .padding(.horizontal, integratedWithSection ? 16 : 10)
        .padding(.vertical, integratedWithSection ? 12 : 8)
        .background(
            ZStack {
                (isCompleted ? Color(hex: "#382014") : Color(hex: "#26160E"))
                StepHeaderPatternView(isCompleted: isCompleted)
            }
        )
        .clipShape(RoundedRectangle(cornerRadius: integratedWithSection ? 0 : 14))
        .overlay(
            RoundedRectangle(cornerRadius: 14)
                .stroke(
                    integratedWithSection ? .clear : (isCompleted ? Color.brandGold.opacity(0.55) : Color.brandGold.opacity(0.28)),
                    lineWidth: 1
                )
        )
        .overlay(alignment: .bottom) {
            if integratedWithSection {
                Rectangle()
                    .fill(Color.brandGold.opacity(0.18))
                    .frame(height: 0.75)
            }
        }
        .onAppear {
            withAnimation(Animation.easeInOut(duration: 1.0).repeatForever(autoreverses: true)) {
                isPulsing = true
            }
        }
    }
}
