//
//  KeychainHelper.swift
//  PubGamesMiniApps
//
//  Secure storage for authentication tokens using iOS Keychain
//

import Foundation
import Security

class KeychainHelper {
    static let shared = KeychainHelper()

    private let service = "com.pubgames.miniapps"

    private init() {}

    // MARK: - Save

    func save(_ data: Data, for key: String) -> Bool {
        // Delete any existing item
        delete(key)

        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecValueData as String: data,
            kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        ]

        let status = SecItemAdd(query as CFDictionary, nil)
        return status == errSecSuccess
    }

    func save(_ string: String, for key: String) -> Bool {
        guard let data = string.data(using: .utf8) else { return false }
        return save(data, for: key)
    }

    // MARK: - Retrieve

    func retrieve(for key: String) -> Data? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        guard status == errSecSuccess else { return nil }
        return result as? Data
    }

    func retrieveString(for key: String) -> String? {
        guard let data = retrieve(for: key) else { return nil }
        return String(data: data, encoding: .utf8)
    }

    // MARK: - Delete

    func delete(_ key: String) -> Bool {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key
        ]

        let status = SecItemDelete(query as CFDictionary)
        return status == errSecSuccess || status == errSecItemNotFound
    }

    // MARK: - Clear All

    func clearAll() -> Bool {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service
        ]

        let status = SecItemDelete(query as CFDictionary)
        return status == errSecSuccess || status == errSecItemNotFound
    }
}

// MARK: - Convenience Extensions

extension KeychainHelper {
    enum Key {
        static let authToken = "authToken"
        static let refreshToken = "refreshToken"
        static let userID = "userID"
        static let username = "username"
    }
}
