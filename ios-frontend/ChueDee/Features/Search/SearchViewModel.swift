import Foundation
import SwiftUI

@MainActor
final class SearchViewModel: ObservableObject {
    enum SearchMode: String { case standalone, combined }
    enum Target: String { case firstName = "first_name", surname }
    enum LoadState: Equatable { case idle, loading, loaded, failed(String) }

    @Published var mode: SearchMode = .standalone
    @Published var target: Target = .firstName
    @Published var referenceName = ""
    @Published var nameSuggestions: [NameResultItem] = []
    @Published var suggestionPreviewPairs: [SuggestionPairPreview] = []
    @Published var suggestionCount = 0
    @Published var suggestionsUnlocked = false
    @Published var isLoadingNameSuggestions = false
    @Published var selectedSuggestionName = ""
    @Published var dayOfBirth = ""
    @Published var decodedMeaning = ""
    @Published var decodedPairs: [String] = []
    @Published var decodedPairTypes: [String] = []
    @Published var decodedPairDescs: [String] = []
    @Published var decodedProfile4D: NameProfile4D?
    @Published var decodedKalakini: KalakiniInfo?
    static let defaultAuspiciousMeanings: [String] = [
        "ความสำเร็จ ความเจริญก้าวหน้าในหน้าที่การงาน",
        "มีเสน่ห์ เมตตามหานิยม เป็นที่รักใคร่ของผู้คน",
        "ร่ำรวย มั่งคั่ง มั่นคง มีโชคลาภวาสนาบารมี",
        "สติปัญญาเฉลียวฉลาด รอบรู้ มีไหวพริบปฏิภาณ",
        "สุขภาพแข็งแรง อายุยืนยาว แคล้วคลาดปลอดภัย",
        "ชีวิตราบรื่น มีความสุข สมหวังในสิ่งที่ปรารถนา"
    ]
    @Published var pinnedMeaning = ""
    @Published var hasUserSelectedMeaning = false
    @Published var meaningPool: [String] = []
    @Published var meaningPageIndex = 0
    @Published var isMeaningAutoPlaying = true
    @Published var suggestions: [String] = Array(defaultAuspiciousMeanings.shuffled().prefix(3))
    @Published var weights = RankingWeights(numerology: 0.35, semantic: 0.25, sound: 0.20, similarity: 0.20)
    @Published private(set) var appliedWeights = RankingWeights(numerology: 0.35, semantic: 0.25, sound: 0.20, similarity: 0.20)
    @Published var results: [NameResultItem] = []
    @Published var state: LoadState = .idle
    @Published var isDecoding = false
    @Published var celebrities: [Celebrity] = []
    @Published var articles: [ArticleItem] = []
    @Published var isLoadingArticles = false
    @Published var isReranking = false
    @Published var rerankError: String?
    @Published var selectedPairDetail: PairDetailResponse?
    @Published var isPairDetailLoading = false
    @Published var pairDetailError: String?

    private let api: APIClient
    private var decodeTask: Task<Void, Never>?
    private var suggestionTask: Task<Void, Never>?
    private var suggestionIsImmediate = false
    private var rerankTask: Task<Void, Never>?
    private var isRefreshingMeanings = false
    private var candidateToken = ""
    private var rerankRevision = 0
    init(api: APIClient = .shared) {
        self.api = api
#if DEBUG
        if let queryIndex = ProcessInfo.processInfo.arguments.firstIndex(of: "-debugQuery"),
           queryIndex + 1 < ProcessInfo.processInfo.arguments.count {
            let name = ProcessInfo.processInfo.arguments[queryIndex + 1]
            referenceName = name
            selectedSuggestionName = ""
        }
        if let bdayIndex = ProcessInfo.processInfo.arguments.firstIndex(of: "-debugBirthday"),
           bdayIndex + 1 < ProcessInfo.processInfo.arguments.count {
            let day = ProcessInfo.processInfo.arguments[bdayIndex + 1]
            dayOfBirth = day
        }
#endif
        Task {
#if DEBUG
            if ProcessInfo.processInfo.arguments.contains("-debugMockPurchase") {
                await PurchaseManager.shared.setMockPurchase(enabled: true)
            }
#endif
            await loadCelebrities()
            await loadArticles()
            nameChanged()
            await refreshMeanings()
        }
    }

    func avatarURL(for celebrity: Celebrity) -> URL? { api.avatarURL(for: celebrity.avatarURL) }

    func selectCelebrity(_ celebrity: Celebrity) {
        referenceName = celebrity.name
        nameChanged()
    }

    private func loadCelebrities() async {
        do { celebrities = try await api.celebrities().data } catch { celebrities = [] }
    }

    func loadArticles() async {
        isLoadingArticles = true
        defer { isLoadingArticles = false }
        do {
            articles = try await api.articles()
        } catch {
            articles = []
        }
    }

    var canSearch: Bool {
        !referenceName.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty &&
        !dayOfBirth.isEmpty &&
        hasUserSelectedMeaning &&
        !pinnedMeaning.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty &&
        state != .loading
    }
    var fieldPrompt: String {
        guard mode == .combined else { return "พิมพ์ชื่อเพื่อค้นหาชื่อใกล้เคียง" }
        return target == .surname ? "พิมพ์ชื่อตัวที่ต้องการใช้ร่วมกัน" : "พิมพ์นามสกุลที่ต้องการใช้ร่วมกัน"
    }
    var targetLabel: String { target == .surname ? "นามสกุล" : "ชื่อตัว" }

    func setMode(_ value: SearchMode) {
        guard mode != value else { return }
        mode = value
        if value == .standalone { target = .firstName }
        referenceName = ""
        selectedSuggestionName = ""
        clearMeaning()
        nameChanged()
    }
    func switchTarget() { guard mode == .combined else { return }; target = target == .firstName ? .surname : .firstName; referenceName = ""; selectedSuggestionName = ""; clearMeaning() }

    func nameChanged() {
        if mode == .standalone, selectedSuggestionName == referenceName, !referenceName.isEmpty { return }
        selectedSuggestionName = ""
        if mode == .standalone {
            requestNameSuggestions()
            decodeSelectedName()
        } else {
            decodeSelectedName()
        }
    }

    func selectSuggestedName(_ name: String) {
        let cleanName = name.trimmingCharacters(in: .whitespacesAndNewlines)
        referenceName = cleanName
        selectedSuggestionName = cleanName
        decodeSelectedName()
        requestNameSuggestions()
    }

    func refreshSelectedNameForBirthday() {
        let name = referenceName.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !name.isEmpty else { return }
        let day = dayOfBirth
        Task {
            do {
                let decoded = try await api.decode(name: name, dayOfBirth: day)
                guard name == referenceName.trimmingCharacters(in: .whitespacesAndNewlines) else { return }
                decodedKalakini = decoded.kalakini
            } catch { }
        }
    }

    func refreshSuggestionsAfterPurchase() {
        guard mode == .standalone, !referenceName.isEmpty else { return }
        requestNameSuggestions(immediate: true)
    }

    private func requestNameSuggestions(immediate: Bool = false) {
        // Don't let a debounced request cancel an in-flight immediate (purchase refresh) request
        if !immediate && suggestionIsImmediate { return }
        suggestionTask?.cancel()
        suggestionIsImmediate = immediate
        nameSuggestions = []
        let query = referenceName.trimmingCharacters(in: .whitespacesAndNewlines)
        guard query.count >= 2, mode == .standalone else {
            nameSuggestions = []
            suggestionPreviewPairs = []
            suggestionCount = 0
            suggestionsUnlocked = false
            isLoadingNameSuggestions = false
            suggestionIsImmediate = false
            return
        }
        isLoadingNameSuggestions = true
        suggestionTask = Task {
            if !immediate {
                try? await Task.sleep(for: .milliseconds(280))
            }
            guard !Task.isCancelled else { return }
            do {
                let response = try await api.nameSuggestions(query: query)
                guard !Task.isCancelled, query == referenceName.trimmingCharacters(in: .whitespacesAndNewlines) else { return }
                nameSuggestions = response.data
                suggestionPreviewPairs = response.previewPairs
                suggestionCount = response.count
                suggestionsUnlocked = response.premiumUnlocked
            } catch {
                nameSuggestions = []
            }
            suggestionIsImmediate = false
            if query == referenceName.trimmingCharacters(in: .whitespacesAndNewlines) { isLoadingNameSuggestions = false }
        }
    }

    private func decodeSelectedName() {
        decodeTask?.cancel(); clearMeaning()
        let name = referenceName.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !name.isEmpty else { return }
        decodeTask = Task {
            try? await Task.sleep(for: .milliseconds(450)); guard !Task.isCancelled else { return }
            isDecoding = true; defer { isDecoding = false }
            do {
                let decoded = try await api.decode(name: name, dayOfBirth: dayOfBirth)
                guard !Task.isCancelled, name == referenceName.trimmingCharacters(in: .whitespacesAndNewlines) else { return }
                withAnimation(.easeInOut(duration: 0.36)) {
                    decodedPairs = decoded.pairs ?? []
                    decodedPairTypes = decoded.pairTypes ?? []
                    decodedPairDescs = decoded.pairDescs ?? []
                    decodedProfile4D = decoded.profile4D
                    decodedKalakini = decoded.kalakini
                    if let profile = decoded.profile4D {
                        let total = max(0.001, profile.numerologyScore + profile.meaningScore + profile.soundScore + profile.resonanceScore)
                        weights = RankingWeights(numerology: profile.numerologyScore / total,
                                                 semantic: profile.meaningScore / total,
                                                 sound: profile.soundScore / total,
                                                 similarity: profile.resonanceScore / total)
                    }
                }
                if let meaning = decoded.meaning?.trimmingCharacters(in: .whitespacesAndNewlines), !meaning.isEmpty {
                    withAnimation(.easeInOut(duration: 0.36)) {
                        decodedMeaning = meaning
                        pinnedMeaning = ""
                        hasUserSelectedMeaning = false
                    }
                    await refreshMeanings()
                } else {
                    withAnimation(.easeInOut(duration: 0.36)) {
                        decodedMeaning = ""
                        pinnedMeaning = ""
                        hasUserSelectedMeaning = false
                    }
                    await refreshMeanings()
                }
            } catch { await refreshMeanings() }
        }
    }

    func resetToBalance() {
        withAnimation(.easeInOut(duration: 0.3)) {
            weights = RankingWeights(numerology: 0.25, semantic: 0.25, sound: 0.25, similarity: 0.25)
        }
    }

    var totalMeaningPages: Int {
        let count = meaningPool.count
        guard count > 0 else { return 36 }
        return count / 3
    }

    func refreshMeanings() async {
        guard !isRefreshingMeanings else { return }
        isRefreshingMeanings = true
        defer { isRefreshingMeanings = false }
        do {
            let response = try await api.randomMeanings(limit: 108)
            let values = response.meanings.filter { !$0.isEmpty }
            if !values.isEmpty {
                // Ensure exactly up to 108 meanings (36 sets of 3)
                let exactCount = min(values.count, 108) - (min(values.count, 108) % 3)
                let trimmed = Array(values.prefix(exactCount))
                withAnimation(.easeInOut(duration: 0.34)) {
                    meaningPool = trimmed
                    meaningPageIndex = 0
                    updatePageSuggestions()
                }
            } else if meaningPool.isEmpty {
                withAnimation(.easeInOut(duration: 0.34)) {
                    meaningPool = Self.defaultAuspiciousMeanings.shuffled()
                    meaningPageIndex = 0
                    updatePageSuggestions()
                }
            }
        } catch {
            if meaningPool.isEmpty {
                withAnimation(.easeInOut(duration: 0.34)) {
                    meaningPool = Self.defaultAuspiciousMeanings.shuffled()
                    meaningPageIndex = 0
                    updatePageSuggestions()
                }
            }
        }
    }

    private func updatePageSuggestions() {
        guard !meaningPool.isEmpty else {
            suggestions = Array(Self.defaultAuspiciousMeanings.shuffled().prefix(3))
            return
        }
        let total = totalMeaningPages
        if meaningPageIndex >= total {
            meaningPageIndex = 0
        } else if meaningPageIndex < 0 {
            meaningPageIndex = max(0, total - 1)
        }
        let start = meaningPageIndex * 3
        let end = min(start + 3, meaningPool.count)
        if start < end {
            suggestions = Array(meaningPool[start..<end])
        }
    }

    func nextMeaningPage() {
        withAnimation(.easeInOut(duration: 0.35)) {
            let total = totalMeaningPages
            if total > 0 {
                meaningPageIndex = (meaningPageIndex + 1) % total
                updatePageSuggestions()
            }
        }
    }

    func previousMeaningPage() {
        withAnimation(.easeInOut(duration: 0.35)) {
            let total = totalMeaningPages
            if total > 0 {
                meaningPageIndex = (meaningPageIndex - 1 + total) % total
                updatePageSuggestions()
            }
        }
    }

    func toggleMeaningAutoPlay() {
        withAnimation(.easeInOut(duration: 0.2)) {
            isMeaningAutoPlaying.toggle()
        }
    }

    func runAutomaticMeaningRefresh() async {
        while !Task.isCancelled {
            try? await Task.sleep(for: .seconds(8))
            guard !Task.isCancelled else { return }
            if isMeaningAutoPlaying {
                nextMeaningPage()
            }
        }
    }

    func pin(_ meaning: String) {
        withAnimation(.easeInOut(duration: 0.36)) {
            pinnedMeaning = meaning
            hasUserSelectedMeaning = true
            isMeaningAutoPlaying = false
        }
    }
    private func clearMeaning() {
        withAnimation(.easeInOut(duration: 0.26)) {
            decodedMeaning = ""
            pinnedMeaning = ""
            hasUserSelectedMeaning = false
            isMeaningAutoPlaying = true
            if suggestions.isEmpty {
                suggestions = Array(Self.defaultAuspiciousMeanings.shuffled().prefix(3))
            }
            decodedPairs = []
            decodedPairTypes = []
            decodedPairDescs = []
            decodedProfile4D = nil
            decodedKalakini = nil
        }
    }

    func search() {
        guard canSearch else { return }; state = .loading
        Task {
            do {
                let response = try await api.search(NameSearchRequest(
                    mode: mode.rawValue, searchTarget: target.rawValue,
                    surname: referenceName.trimmingCharacters(in: .whitespacesAndNewlines),
                    prototypeName: referenceName.trimmingCharacters(in: .whitespacesAndNewlines),
                    dayOfBirth: dayOfBirth, wishQuery: pinnedMeaning, isNameIntent: true, limit: 100,
                    numerologyWeight: weights.numerology, semanticWeight: weights.semantic,
                    soundWeight: weights.sound, similarityWeight: weights.similarity))
                guard response.policy.rankingContractVersion == APIClient.supportedRankingContract else { throw APIError.contractMismatch }
                candidateToken = response.candidateToken
                // Keep the selector on the user's formula while exposing the exact
                // formula that produced the returned ordering separately.
                weights = response.requestedWeights
                appliedWeights = response.appliedWeights
                results = response.data; state = .loaded
            } catch { results = []; state = .failed((error as? LocalizedError)?.errorDescription ?? "เกิดข้อผิดพลาด กรุณาลองใหม่") }
        }
    }

    func updateWeightsAndRerank(_ newWeights: RankingWeights) {
        weights = newWeights
        rerankError = nil
        guard !candidateToken.isEmpty else { return }
        rerankTask?.cancel()
        rerankRevision += 1
        let revision = rerankRevision
        rerankTask = Task {
            try? await Task.sleep(for: .milliseconds(220))
            guard !Task.isCancelled else { return }
            isReranking = true
            defer { if revision == rerankRevision { isReranking = false } }
            do {
                let response = try await api.rerank(NameRerankRequest(
                    candidateToken: candidateToken, limit: 100,
                    numerologyWeight: newWeights.numerology,
                    semanticWeight: newWeights.semantic,
                    soundWeight: newWeights.sound,
                    similarityWeight: newWeights.similarity
                ))
                guard !Task.isCancelled, revision == rerankRevision else { return }
                withAnimation(.spring(response: 0.52, dampingFraction: 0.84)) {
                    weights = response.requestedWeights
                    appliedWeights = response.appliedWeights
                    results = response.data
                }
            } catch {
                guard !Task.isCancelled, revision == rerankRevision else { return }
                rerankError = (error as? LocalizedError)?.errorDescription ?? "ปรับสูตรไม่สำเร็จ"
            }
        }
    }

    func fetchPairDetail(_ pairNumber: String) {
        let cleanPair = pairNumber.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !cleanPair.isEmpty else { return }
        selectedPairDetail = nil
        pairDetailError = nil
        isPairDetailLoading = true
        Task {
            do {
                selectedPairDetail = try await api.pairDetail(number: cleanPair)
            } catch {
                pairDetailError = (error as? LocalizedError)?.errorDescription ?? "ไม่สามารถโหลดความหมายคู่เลขได้"
            }
            isPairDetailLoading = false
        }
    }

    func dismissPairDetail() {
        selectedPairDetail = nil
        pairDetailError = nil
        isPairDetailLoading = false
    }

    func resetAll() {
        decodeTask?.cancel()
        suggestionTask?.cancel()
        rerankTask?.cancel()

        withAnimation(.easeInOut(duration: 0.3)) {
            referenceName = ""
            selectedSuggestionName = ""
            nameSuggestions = []
            suggestionPreviewPairs = []
            suggestionCount = 0
            suggestionsUnlocked = false
            isLoadingNameSuggestions = false

            dayOfBirth = ""

            decodedMeaning = ""
            pinnedMeaning = ""
            hasUserSelectedMeaning = false
            isMeaningAutoPlaying = true
            meaningPageIndex = 0
            suggestions = Array(Self.defaultAuspiciousMeanings.shuffled().prefix(3))
            decodedPairs = []
            decodedPairTypes = []
            decodedPairDescs = []
            decodedProfile4D = nil
            decodedKalakini = nil

            weights = RankingWeights(numerology: 0.35, semantic: 0.25, sound: 0.20, similarity: 0.20)
            appliedWeights = weights
            results = []
            candidateToken = ""
            state = .idle
            isDecoding = false
            isReranking = false
            rerankError = nil
            selectedPairDetail = nil
            pairDetailError = nil
            isPairDetailLoading = false
        }
    }
}
