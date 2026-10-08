import Foundation
import Combine

@MainActor
final class AuspiciousViewModel: ObservableObject {
    @Published var activities: [AuspiciousActivity] = []
    @Published var selectedActivityCode = ""
    @Published var isLoadingActivities = false
    @Published var isSearching = false
    @Published var isLoadingTaksa = false
    @Published var result: AuspiciousSearchResponse?
    @Published var taksa: TaksaCalculationResponse?
    @Published var errorMessage: String?
    @Published var taksaErrorMessage: String?

    private let apiClient: APIClient
    private var taksaTask: Task<Void, Never>?

    init(apiClient: APIClient = .shared) {
        self.apiClient = apiClient
    }

    var selectedActivity: AuspiciousActivity? {
        activities.first { $0.code == selectedActivityCode }
    }

    func loadActivities() async {
        guard activities.isEmpty, !isLoadingActivities else { return }
        isLoadingActivities = true
        errorMessage = nil
        do {
            let response = try await apiClient.auspiciousActivities()
            guard !response.activities.isEmpty else {
                throw NSError(domain: "Auspicious", code: 1, userInfo: [NSLocalizedDescriptionKey: "ยังไม่พบกิจกรรมสำหรับดูฤกษ์"])
            }
            activities = response.activities.sorted {
                if $0.categorySortOrder == $1.categorySortOrder { return $0.sortOrder < $1.sortOrder }
                return $0.categorySortOrder < $1.categorySortOrder
            }
            if selectedActivityCode.isEmpty {
                selectedActivityCode = activities.first(where: { $0.code == "name_change" })?.code ?? activities[0].code
            }
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoadingActivities = false
    }

    func selectActivity(_ code: String) {
        selectedActivityCode = code
        result = nil
        errorMessage = nil
    }

    func search(_ request: AuspiciousSearchRequest) async {
        guard !isSearching else { return }
        isSearching = true
        result = nil
        errorMessage = nil
        do {
            result = try await apiClient.searchAuspiciousDates(request)
        } catch {
            errorMessage = error.localizedDescription
        }
        isSearching = false
    }

    func loadTaksa(birthDate: String, birthTime: String) {
        taksaTask?.cancel()
        guard !birthDate.isEmpty, !birthTime.isEmpty else {
            taksa = nil
            taksaErrorMessage = nil
            isLoadingTaksa = false
            return
        }
        taksaTask = Task { [weak self] in
            guard let self else { return }
            isLoadingTaksa = true
            taksaErrorMessage = nil
            do {
                try await Task.sleep(for: .milliseconds(250))
                try Task.checkCancellation()
                taksa = try await apiClient.taksaCalculation(
                    TaksaRequest(birthDate: birthDate, birthTime: birthTime)
                )
            } catch is CancellationError {
                return
            } catch {
                taksaErrorMessage = error.localizedDescription
            }
            isLoadingTaksa = false
        }
    }
}
