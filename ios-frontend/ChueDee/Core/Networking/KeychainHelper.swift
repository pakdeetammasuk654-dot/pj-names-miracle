import Foundation
import Security

final class KeychainHelper {
    static let shared = KeychainHelper()
    private init() {}

    private let service = "com.numberniceic.app.support_chat"

    func save(key: String, value: String) {
        guard let data = value.data(using: .utf8) else { return }
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key
        ]
        let update = SecItemUpdate(query as CFDictionary, [kSecValueData as String: data] as CFDictionary)
        let status = update == errSecItemNotFound
            ? SecItemAdd(query.merging([
                kSecValueData as String: data,
                kSecAttrAccessible as String: kSecAttrAccessibleAfterFirstUnlock
            ]) { _, new in new } as CFDictionary, nil)
            : update
        guard status == errSecSuccess else { return }

        // Remove values written by older builds only after they have a Keychain copy.
        UserDefaults.standard.removeObject(forKey: legacyDefaultsKey(for: key))
    }

    func get(key: String) -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]
        var item: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &item)
        if status == errSecSuccess, let data = item as? Data, let str = String(data: data, encoding: .utf8), !str.isEmpty {
            return str
        }
        // One-time migration for values stored by older builds as a simulator fallback.
        guard let legacyValue = UserDefaults.standard.string(forKey: legacyDefaultsKey(for: key)),
              !legacyValue.isEmpty else { return nil }
        save(key: key, value: legacyValue)
        return legacyValue
    }

    func delete(key: String) {
        UserDefaults.standard.removeObject(forKey: legacyDefaultsKey(for: key))
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key
        ]
        SecItemDelete(query as CFDictionary)
    }

    func clearAllChatSession() {
        delete(key: "conversation_id")
        delete(key: "customer_token")
        delete(key: "customer_avatar")
        delete(key: "apns_device_token")
    }

    private func legacyDefaultsKey(for key: String) -> String {
        "keychain_fallback_" + key
    }
}
