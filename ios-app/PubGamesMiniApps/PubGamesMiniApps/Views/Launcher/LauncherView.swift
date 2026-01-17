//
//  LauncherView.swift
//  PubGamesMiniApps
//
//  App launcher grid showing all available mini apps
//

import SwiftUI

struct LauncherView: View {
    @StateObject private var authService = AuthService.shared
    @StateObject private var appService = AppService.shared
    @State private var selectedApp: MiniApp?
    @State private var showSettings = false
    @State private var errorMessage: String?

    let columns = [
        GridItem(.flexible()),
        GridItem(.flexible())
    ]

    var body: some View {
        NavigationView {
            ZStack {
                // Background
                Color(UIColor.systemGroupedBackground)
                    .ignoresSafeArea()

                if appService.isLoading {
                    ProgressView("Loading apps...")
                } else if appService.apps.isEmpty {
                    VStack(spacing: 16) {
                        Image(systemName: "tray")
                            .font(.system(size: 60))
                            .foregroundColor(.gray)

                        Text("No apps available")
                            .font(.headline)
                            .foregroundColor(.gray)

                        Button("Refresh") {
                            Task {
                                await appService.refreshApps()
                            }
                        }
                    }
                } else {
                    ScrollView {
                        VStack(alignment: .leading, spacing: 16) {
                            // Welcome header
                            if let user = authService.currentUser {
                                VStack(alignment: .leading, spacing: 4) {
                                    Text("Welcome back,")
                                        .font(.subheadline)
                                        .foregroundColor(.secondary)

                                    Text(user.username)
                                        .font(.title)
                                        .fontWeight(.bold)
                                }
                                .padding(.horizontal)
                                .padding(.top)
                            }

                            // Apps grid
                            LazyVGrid(columns: columns, spacing: 16) {
                                ForEach(appService.apps) { app in
                                    AppTileView(app: app)
                                        .onTapGesture {
                                            selectedApp = app
                                        }
                                }
                            }
                            .padding()
                        }
                    }
                }

                if let errorMessage = errorMessage {
                    VStack {
                        Spacer()
                        Text(errorMessage)
                            .padding()
                            .background(Color.red.opacity(0.8))
                            .foregroundColor(.white)
                            .cornerRadius(8)
                            .padding()
                    }
                }
            }
            .navigationTitle("PubGames")
            .navigationBarTitleDisplayMode(.large)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Menu {
                        Button(action: {
                            Task {
                                await appService.refreshApps()
                            }
                        }) {
                            Label("Refresh Apps", systemImage: "arrow.clockwise")
                        }

                        Button(action: { showSettings = true }) {
                            Label("Settings", systemImage: "gear")
                        }

                        Divider()

                        Button(role: .destructive, action: {
                            authService.logout()
                        }) {
                            Label("Logout", systemImage: "rectangle.portrait.and.arrow.right")
                        }
                    } label: {
                        Image(systemName: "ellipsis.circle")
                    }
                }
            }
            .task {
                await loadApps()
            }
            .sheet(item: $selectedApp) { app in
                WebViewContainer(app: app, isPresented: $selectedApp)
            }
            .sheet(isPresented: $showSettings) {
                SettingsView(isPresented: $showSettings)
            }
        }
    }

    private func loadApps() async {
        do {
            try await appService.fetchApps()
        } catch {
            errorMessage = error.localizedDescription
            // Auto-hide error after 3 seconds
            DispatchQueue.main.asyncAfter(deadline: .now() + 3) {
                errorMessage = nil
            }
        }
    }
}

// MARK: - App Tile View

struct AppTileView: View {
    let app: MiniApp

    var body: some View {
        VStack(spacing: 12) {
            // Icon
            ZStack {
                RoundedRectangle(cornerRadius: 16)
                    .fill(
                        LinearGradient(
                            colors: [iconColor.opacity(0.7), iconColor],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        )
                    )
                    .frame(width: 80, height: 80)

                Image(systemName: iconName)
                    .font(.system(size: 36))
                    .foregroundColor(.white)
            }

            // Name
            Text(app.name)
                .font(.headline)
                .multilineTextAlignment(.center)
                .lineLimit(2)

            // Description (optional)
            if let description = app.description {
                Text(description)
                    .font(.caption)
                    .foregroundColor(.secondary)
                    .multilineTextAlignment(.center)
                    .lineLimit(2)
            }
        }
        .padding()
        .frame(maxWidth: .infinity)
        .background(Color(UIColor.secondarySystemGroupedBackground))
        .cornerRadius(12)
    }

    // Generate icon based on app name
    private var iconName: String {
        switch app.name.lowercased() {
        case let name where name.contains("tic") || name.contains("tac"):
            return "number.square"
        case let name where name.contains("sweep"):
            return "gift"
        case let name where name.contains("last") || name.contains("standing"):
            return "sportscourt"
        default:
            return "app.fill"
        }
    }

    // Generate color based on app ID
    private var iconColor: Color {
        let colors: [Color] = [.blue, .purple, .green, .orange, .pink, .red]
        return colors[app.id % colors.count]
    }
}

#Preview {
    LauncherView()
}
