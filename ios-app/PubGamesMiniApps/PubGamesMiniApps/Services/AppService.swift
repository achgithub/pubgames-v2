//
//  AppService.swift
//  PubGamesMiniApps
//
//  Service for fetching and managing mini apps from the Identity Service
//

import Foundation

enum AppServiceError: LocalizedError {
    case invalidURL
    case networkError(Error)
    case invalidResponse
    case unauthorized

    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "Invalid server URL"
        case .networkError(let error):
            return "Network error: \(error.localizedDescription)"
        case .invalidResponse:
            return "Invalid response from server"
        case .unauthorized:
            return "Unauthorized - please log in again"
        }
    }
}

@MainActor
class AppService: ObservableObject {
    static let shared = AppService()

    @Published var apps: [MiniApp] = []
    @Published var isLoading = false

    private let keychain = KeychainHelper.shared

    private init() {}

    // MARK: - Fetch Apps

    func fetchApps() async throws {
        guard let url = Config.APIEndpoint.apps.url else {
            throw AppServiceError.invalidURL
        }

        guard let token = keychain.retrieveString(for: KeychainHelper.Key.authToken) else {
            throw AppServiceError.unauthorized
        }

        isLoading = true
        defer { isLoading = false }

        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = "GET"
        urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        do {
            let (data, response) = try await URLSession.shared.data(for: urlRequest)

            guard let httpResponse = response as? HTTPURLResponse else {
                throw AppServiceError.invalidResponse
            }

            guard httpResponse.statusCode == 200 else {
                if httpResponse.statusCode == 401 {
                    throw AppServiceError.unauthorized
                }
                throw AppServiceError.invalidResponse
            }

            let appsResponse = try JSONDecoder().decode(AppsResponse.self, from: data)
            self.apps = appsResponse.apps.filter { $0.isActive }.sorted {
                ($0.order ?? Int.max) < ($1.order ?? Int.max)
            }
        } catch let error as AppServiceError {
            throw error
        } catch {
            throw AppServiceError.networkError(error)
        }
    }

    // MARK: - Refresh Apps

    func refreshApps() async {
        do {
            try await fetchApps()
        } catch {
            print("Error refreshing apps: \(error.localizedDescription)")
        }
    }
}
