//
//  MiniApp.swift
//  PubGamesMiniApps
//
//  Mini app model from Identity Service /api/apps endpoint
//

import Foundation

struct MiniApp: Codable, Identifiable {
    let id: Int
    let name: String
    let url: String
    let description: String?
    let iconName: String?
    let isActive: Bool
    let order: Int?

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case url
        case description
        case iconName = "icon_name"
        case isActive = "is_active"
        case order
    }

    // Computed property for full URL
    var fullURL: String {
        // If URL is relative, prepend base URL
        if url.starts(with: "http") {
            return url
        } else {
            // URL is like "http://192.168.1.100:30040"
            return url
        }
    }
}

struct AppsResponse: Codable {
    let apps: [MiniApp]
}
