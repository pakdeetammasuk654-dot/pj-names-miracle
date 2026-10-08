import SwiftUI
import UserNotifications
import UIKit

final class AppDelegate: NSObject, UIApplicationDelegate, UNUserNotificationCenterDelegate {
    func application(
        _ application: UIApplication,
        didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]? = nil
    ) -> Bool {
        UNUserNotificationCenter.current().delegate = self
        UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .badge, .sound]) { granted, error in
            if granted {
                DispatchQueue.main.async {
                    UIApplication.shared.registerForRemoteNotifications()
                }
            }
        }
        BadgeManager.clearBadge()
        if let splashBoardPath = NSSearchPathForDirectoriesInDomains(.libraryDirectory, .userDomainMask, true).first?.appending("/SplashBoard") {
            try? FileManager.default.removeItem(atPath: splashBoardPath)
        }
        return true
    }

    func applicationDidBecomeActive(_ application: UIApplication) {
        BadgeManager.clearBadge()
    }

    func application(_ application: UIApplication, didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data) {
        let tokenParts = deviceToken.map { data in String(format: "%02.2hhx", data) }
        let token = tokenParts.joined()
        print("[APNs] Successfully registered device token: \(token)")
        Task { @MainActor in
            SupportChatViewModel.shared.setDeviceToken(token)
        }
    }

    func application(_ application: UIApplication, didFailToRegisterForRemoteNotificationsWithError error: Error) {
        print("[APNs] Failed to register remote notifications: \(error.localizedDescription)")
    }

    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        willPresent notification: UNNotification,
        withCompletionHandler completionHandler: @escaping (UNNotificationPresentationOptions) -> Void
    ) {
        Task { @MainActor in
            await SupportChatViewModel.shared.refresh()
        }
        completionHandler([.banner, .sound, .badge])
    }

    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse,
        withCompletionHandler completionHandler: @escaping () -> Void
    ) {
        BadgeManager.clearBadge()
        Task { @MainActor in
            SupportChatViewModel.shared.openChatRequested = true
            await SupportChatViewModel.shared.refresh()
        }
        completionHandler()
    }
}

enum BadgeManager {
    static func clearBadge() {
        if #available(iOS 16.0, *) {
            UNUserNotificationCenter.current().setBadgeCount(0) { error in
                if let error = error {
                    print("[Badge] Failed to clear badge count: \(error)")
                }
            }
        } else {
            DispatchQueue.main.async {
                UIApplication.shared.applicationIconBadgeNumber = 0
            }
        }
    }

    static func setBadge(_ count: Int) {
        let safeCount = max(0, count)
        if #available(iOS 16.0, *) {
            UNUserNotificationCenter.current().setBadgeCount(safeCount) { error in
                if let error = error {
                    print("[Badge] Failed to set badge count: \(error)")
                }
            }
        } else {
            DispatchQueue.main.async {
                UIApplication.shared.applicationIconBadgeNumber = safeCount
            }
        }
    }
}

@main
struct ChueDeeApp: App {
    @UIApplicationDelegateAdaptor(AppDelegate.self) var appDelegate: AppDelegate

    var body: some Scene {
        WindowGroup {
            AppLaunchContainer()
        }
    }
}

private struct AppLaunchContainer: View {
    @StateObject private var viewModel = SearchViewModel()
    @State private var showsSplash = true

    var body: some View {
        ZStack {
            SearchView(viewModel: viewModel)

            if showsSplash {
                BrandedLaunchView()
                    .transition(.opacity)
                    .zIndex(100)
            }
        }
        .task {
            try? await Task.sleep(for: .milliseconds(1_200))
            withAnimation(.easeOut(duration: 0.34)) {
                showsSplash = false
            }
        }
    }
}

private struct BrandedLaunchView: View {
    @State private var isVisible = false

    var body: some View {
        ZStack {
            Color.brandBackground.ignoresSafeArea()
            VStack(spacing: 16) {
                Image("LaunchLogo")
                    .resizable()
                    .scaledToFit()
                    .frame(width: 120, height: 120)
                    .scaleEffect(isVisible ? 1 : 0.82)
                    .opacity(isVisible ? 1 : 0)
                    .shadow(color: Color.brandGold.opacity(0.30), radius: 18)
                Text("ชื่อดี 4 มิติ")
                    .font(.sarabun(18, weight: .bold))
                    .foregroundStyle(Color.brandCream)
                    .opacity(isVisible ? 1 : 0)
            }
        }
        .onAppear {
            withAnimation(.spring(response: 0.55, dampingFraction: 0.78)) {
                isVisible = true
            }
        }
        .accessibilityElement(children: .combine)
        .accessibilityLabel("ชื่อดี 4 มิติ กำลังเปิดแอป")
    }
}
