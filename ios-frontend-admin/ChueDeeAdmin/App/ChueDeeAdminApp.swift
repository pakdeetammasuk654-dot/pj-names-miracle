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
        AdminBadgeManager.clearBadge()
        return true
    }

    func applicationDidBecomeActive(_ application: UIApplication) {
        AdminBadgeManager.clearBadge()
    }

    func application(_ application: UIApplication, didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data) {
        let tokenParts = deviceToken.map { data in String(format: "%02.2hhx", data) }
        let token = tokenParts.joined()
        print("[Admin APNs] Successfully registered device token: \(token)")
        AdminKeychainHelper.shared.save(key: "admin_device_token", value: token)
        Task {
            try? await AdminAPIClient.shared.registerDeviceToken(token: token)
        }
    }

    func application(_ application: UIApplication, didFailToRegisterForRemoteNotificationsWithError error: Error) {
        print("[Admin APNs] Failed to register remote notifications: \(error.localizedDescription)")
    }

    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        willPresent notification: UNNotification,
        withCompletionHandler completionHandler: @escaping (UNNotificationPresentationOptions) -> Void
    ) {
        completionHandler([.banner, .sound, .badge])
    }

    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse,
        withCompletionHandler completionHandler: @escaping () -> Void
    ) {
        let userInfo = response.notification.request.content.userInfo
        if let convID = userInfo["conversation_id"] as? String, !convID.isEmpty {
            NotificationCenter.default.post(
                name: NSNotification.Name("OpenAdminConversation"),
                object: nil,
                userInfo: ["conversation_id": convID]
            )
        }
        completionHandler()
    }
}

enum AdminBadgeManager {
    static func clearBadge() {
        if #available(iOS 16.0, *) {
            UNUserNotificationCenter.current().setBadgeCount(0) { error in
                if let error = error {
                    print("[Admin Badge] Failed to clear badge count: \(error)")
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
                    print("[Admin Badge] Failed to set badge count to \(safeCount): \(error)")
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
struct ChueDeeAdminApp: App {
    @UIApplicationDelegateAdaptor(AppDelegate.self) var appDelegate: AppDelegate
    @StateObject private var authVM = AuthViewModel.shared

    var body: some Scene {
        WindowGroup {
            if authVM.isAuthenticated {
                InboxView()
            } else {
                LoginView()
            }
        }
    }
}
