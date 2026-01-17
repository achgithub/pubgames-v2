//
//  PubGamesMiniAppsApp.swift
//  PubGamesMiniApps
//
//  Main app entry point with authentication flow
//

import SwiftUI

@main
struct PubGamesMiniAppsApp: App {
    @StateObject private var authService = AuthService.shared

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environmentObject(authService)
        }
    }
}

struct ContentView: View {
    @EnvironmentObject var authService: AuthService

    var body: some View {
        Group {
            if authService.isAuthenticated {
                LauncherView()
            } else {
                LoginView()
            }
        }
        .animation(.easeInOut, value: authService.isAuthenticated)
    }
}

#Preview {
    ContentView()
        .environmentObject(AuthService.shared)
}
