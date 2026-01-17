//
//  User.swift
//  PubGamesMiniApps
//
//  User model matching the backend Identity Service schema
//

import Foundation

struct User: Codable, Identifiable {
    let id: Int
    let username: String
    let email: String?
    let isAdmin: Bool
    let createdAt: String?

    enum CodingKeys: String, CodingKey {
        case id
        case username
        case email
        case isAdmin = "is_admin"
        case createdAt = "created_at"
    }
}

// MARK: - Authentication Request/Response Models

struct LoginRequest: Codable {
    let username: String
    let password: String
}

struct RegisterRequest: Codable {
    let username: String
    let password: String
    let email: String?
}

struct AuthResponse: Codable {
    let message: String
    let token: String?
    let user: User?
}

struct TokenValidationResponse: Codable {
    let valid: Bool
    let user: User?
}
