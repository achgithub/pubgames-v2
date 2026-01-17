//
//  SettingsView.swift
//  PubGamesMiniApps
//
//  App settings and configuration
//

import SwiftUI

struct SettingsView: View {
    @Binding var isPresented: Bool
    @StateObject private var authService = AuthService.shared
    @State private var serverURL = Config.identityBackendURL
    @State private var biometricsEnabled = Config.biometricsEnabled

    var body: some View {
        NavigationView {
            Form {
                // User Info Section
                Section(header: Text("User Information")) {
                    if let user = authService.currentUser {
                        HStack {
                            Text("Username")
                            Spacer()
                            Text(user.username)
                                .foregroundColor(.secondary)
                        }

                        if let email = user.email {
                            HStack {
                                Text("Email")
                                Spacer()
                                Text(email)
                                    .foregroundColor(.secondary)
                            }
                        }

                        HStack {
                            Text("User ID")
                            Spacer()
                            Text("\(user.id)")
                                .foregroundColor(.secondary)
                        }

                        if user.isAdmin {
                            HStack {
                                Text("Role")
                                Spacer()
                                Text("Administrator")
                                    .foregroundColor(.orange)
                            }
                        }
                    }
                }

                // Server Configuration
                Section(header: Text("Server Configuration")) {
                    HStack {
                        Text("Server URL")
                        Spacer()
                        Text(serverURL)
                            .font(.caption)
                            .foregroundColor(.secondary)
                            .lineLimit(1)
                    }

                    // Future: Allow changing server URL
                    // TextField("Server URL", text: $serverURL)
                    //     .autocapitalization(.none)
                    //     .keyboardType(.URL)
                }

                // Security Settings
                Section(header: Text("Security")) {
                    Toggle("Face ID / Touch ID", isOn: $biometricsEnabled)
                        .onChange(of: biometricsEnabled) { newValue in
                            Config.biometricsEnabled = newValue
                        }
                        .disabled(true) // Disabled until implemented
                        .opacity(0.5)
                }
                .footer(Text("Biometric authentication coming in a future update"))

                // App Information
                Section(header: Text("About")) {
                    HStack {
                        Text("Version")
                        Spacer()
                        Text("1.0.0")
                            .foregroundColor(.secondary)
                    }

                    HStack {
                        Text("Build")
                        Spacer()
                        Text("1")
                            .foregroundColor(.secondary)
                    }
                }

                // Logout
                Section {
                    Button(role: .destructive, action: {
                        authService.logout()
                        isPresented = false
                    }) {
                        HStack {
                            Spacer()
                            Text("Logout")
                            Spacer()
                        }
                    }
                }
            }
            .navigationTitle("Settings")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("Done") {
                        isPresented = false
                    }
                }
            }
        }
    }
}

#Preview {
    SettingsView(isPresented: .constant(true))
}
