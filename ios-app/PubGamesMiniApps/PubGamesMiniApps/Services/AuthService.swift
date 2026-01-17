//
//  AuthService.swift
//  PubGamesMiniApps
//
//  Authentication service managing login, registration, and token validation
//

import Foundation

enum AuthError: LocalizedError {
    case invalidURL
    case networkError(Error)
    case invalidResponse
    case serverError(String)
    case invalidCredentials
    case unknown

    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "Invalid server URL"
        case .networkError(let error):
            return "Network error: \(error.localizedDescription)"
        case .invalidResponse:
            return "Invalid response from server"
        case .serverError(let message):
            return message
        case .invalidCredentials:
            return "Invalid username or password"
        case .unknown:
            return "An unknown error occurred"
        }
    }
}

@MainActor
class AuthService: ObservableObject {
    static let shared = AuthService()

    @Published var currentUser: User?
    @Published var isAuthenticated = false

    private let keychain = KeychainHelper.shared

    private init() {
        // Check if we have a stored token on init
        Task {
            await validateStoredToken()
        }
    }

    // MARK: - Authentication Methods

    func login(username: String, password: String) async throws -> User {
        guard let url = Config.APIEndpoint.login.url else {
            throw AuthError.invalidURL
        }

        let request = LoginRequest(username: username, password: password)

        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = "POST"
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
        urlRequest.httpBody = try JSONEncoder().encode(request)

        let (data, response) = try await URLSession.shared.data(for: urlRequest)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw AuthError.invalidResponse
        }

        guard httpResponse.statusCode == 200 else {
            if let errorResponse = try? JSONDecoder().decode(AuthResponse.self, from: data) {
                throw AuthError.serverError(errorResponse.message)
            }
            throw AuthError.invalidCredentials
        }

        let authResponse = try JSONDecoder().decode(AuthResponse.self, from: data)

        guard let token = authResponse.token, let user = authResponse.user else {
            throw AuthError.invalidResponse
        }

        // Store token in keychain
        _ = keychain.save(token, for: KeychainHelper.Key.authToken)
        _ = keychain.save(String(user.id), for: KeychainHelper.Key.userID)
        _ = keychain.save(user.username, for: KeychainHelper.Key.username)

        self.currentUser = user
        self.isAuthenticated = true

        return user
    }

    func register(username: String, password: String, email: String?) async throws -> User {
        guard let url = Config.APIEndpoint.register.url else {
            throw AuthError.invalidURL
        }

        let request = RegisterRequest(username: username, password: password, email: email)

        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = "POST"
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
        urlRequest.httpBody = try JSONEncoder().encode(request)

        let (data, response) = try await URLSession.shared.data(for: urlRequest)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw AuthError.invalidResponse
        }

        guard httpResponse.statusCode == 200 || httpResponse.statusCode == 201 else {
            if let errorResponse = try? JSONDecoder().decode(AuthResponse.self, from: data) {
                throw AuthError.serverError(errorResponse.message)
            }
            throw AuthError.serverError("Registration failed")
        }

        let authResponse = try JSONDecoder().decode(AuthResponse.self, from: data)

        guard let token = authResponse.token, let user = authResponse.user else {
            throw AuthError.invalidResponse
        }

        // Store token in keychain
        _ = keychain.save(token, for: KeychainHelper.Key.authToken)
        _ = keychain.save(String(user.id), for: KeychainHelper.Key.userID)
        _ = keychain.save(user.username, for: KeychainHelper.Key.username)

        self.currentUser = user
        self.isAuthenticated = true

        return user
    }

    func logout() {
        _ = keychain.clearAll()
        self.currentUser = nil
        self.isAuthenticated = false
    }

    func validateStoredToken() async {
        guard let token = keychain.retrieveString(for: KeychainHelper.Key.authToken) else {
            self.isAuthenticated = false
            return
        }

        do {
            let user = try await validateToken(token)
            self.currentUser = user
            self.isAuthenticated = true
        } catch {
            // Token is invalid, clear it
            logout()
        }
    }

    private func validateToken(_ token: String) async throws -> User {
        guard let url = Config.APIEndpoint.validateToken.url else {
            throw AuthError.invalidURL
        }

        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = "GET"
        urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        let (data, response) = try await URLSession.shared.data(for: urlRequest)

        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw AuthError.invalidResponse
        }

        let validationResponse = try JSONDecoder().decode(TokenValidationResponse.self, from: data)

        guard validationResponse.valid, let user = validationResponse.user else {
            throw AuthError.invalidCredentials
        }

        return user
    }

    // MARK: - Token Access

    func getToken() -> String? {
        return keychain.retrieveString(for: KeychainHelper.Key.authToken)
    }
}
