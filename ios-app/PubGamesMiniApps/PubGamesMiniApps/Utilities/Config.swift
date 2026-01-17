//
//  Config.swift
//  PubGamesMiniApps
//
//  Configuration management for server URLs and app settings
//

import Foundation

struct Config {
    // MARK: - Server Configuration

    /// Base URL for the Identity Service backend
    static var identityBackendURL: String {
        #if DEBUG
        return UserDefaults.standard.string(forKey: "serverURL") ?? "http://localhost:3001"
        #else
        return UserDefaults.standard.string(forKey: "serverURL") ?? "https://your-production-server.com:3001"
        #endif
    }

    /// Base URL for the Identity Service frontend
    static var identityFrontendURL: String {
        let backend = identityBackendURL
        let host = backend.components(separatedBy: ":").dropLast().joined(separator: ":")
        return "\(host):30000"
    }

    // MARK: - API Endpoints

    enum APIEndpoint {
        case register
        case login
        case validateToken
        case apps
        case serverInfo

        var path: String {
            switch self {
            case .register: return "/api/register"
            case .login: return "/api/login"
            case .validateToken: return "/api/validate-token"
            case .apps: return "/api/apps"
            case .serverInfo: return "/api/server-info"
            }
        }

        var url: URL? {
            URL(string: Config.identityBackendURL + path)
        }
    }

    // MARK: - Cache Configuration

    /// Maximum cache size in bytes (50 MB)
    static let maxCacheSize: Int = 50 * 1024 * 1024

    /// Cache expiration time in seconds (1 hour)
    static let cacheExpirationTime: TimeInterval = 3600

    // MARK: - Security Configuration

    /// Enable biometric authentication
    static var biometricsEnabled: Bool {
        get { UserDefaults.standard.bool(forKey: "biometricsEnabled") }
        set { UserDefaults.standard.set(newValue, forKey: "biometricsEnabled") }
    }

    // MARK: - Helper Methods

    /// Update the server URL (useful for development/testing)
    static func updateServerURL(_ url: String) {
        UserDefaults.standard.set(url, forKey: "serverURL")
    }

    /// Get URL for a mini app
    static func miniAppURL(port: Int) -> String {
        let backend = identityBackendURL
        let host = backend.components(separatedBy: ":").dropLast().joined(separator: ":")
        return "\(host):\(port)"
    }
}
