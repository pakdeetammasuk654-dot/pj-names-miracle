import SwiftUI
import Combine

@MainActor
final class AuthViewModel: ObservableObject {
    static let shared = AuthViewModel()

    @Published var isAuthenticated: Bool = false
    @Published var usernameInput: String = "admin"
    @Published var passwordInput: String = ""
    @Published var isLoading: Bool = false
    @Published var errorMessage: String?
    @Published var adminName: String = "อ.ทญา ว่องไว"

    private let apiClient = AdminAPIClient.shared
    private let keychain = AdminKeychainHelper.shared

    init() {
        if let user = keychain.get(key: "admin_username"),
           let pass = keychain.get(key: "admin_password"),
           !user.isEmpty, !pass.isEmpty {
            self.usernameInput = user
            self.passwordInput = pass
            self.isAuthenticated = true
        }
    }

    func login() async {
        let u = usernameInput.trimmingCharacters(in: .whitespacesAndNewlines)
        let p = passwordInput.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !u.isEmpty && !p.isEmpty else {
            errorMessage = "กรุณากรอกชื่อผู้ใช้และรหัสผ่าน"
            return
        }

        isLoading = true
        errorMessage = nil

        do {
            let res = try await apiClient.login(username: u, password: p)
            self.adminName = res.name
            keychain.save(key: "admin_username", value: u)
            keychain.save(key: "admin_password", value: p)
            self.isAuthenticated = true
            self.isLoading = false

            // Auto register device token if cached
            if let token = keychain.get(key: "admin_device_token"), !token.isEmpty {
                try? await apiClient.registerDeviceToken(token: token)
            }
        } catch let err as AdminAPIError {
            self.isLoading = false
            self.errorMessage = err.localizedDescription
        } catch {
            self.isLoading = false
            self.errorMessage = "เข้าสู่ระบบไม่สำเร็จ กรุณาตรวจสอบรหัสผ่าน"
        }
    }

    func logout() {
        keychain.delete(key: "admin_username")
        keychain.delete(key: "admin_password")
        self.passwordInput = ""
        self.isAuthenticated = false
    }
}
