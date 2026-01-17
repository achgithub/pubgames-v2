//
//  LoginView.swift
//  PubGamesMiniApps
//
//  Native login screen with future Face ID integration point
//

import SwiftUI

struct LoginView: View {
    @StateObject private var authService = AuthService.shared
    @State private var username = ""
    @State private var password = ""
    @State private var isLoading = false
    @State private var errorMessage: String?
    @State private var showRegister = false

    var body: some View {
        NavigationView {
            ZStack {
                // Background gradient
                LinearGradient(
                    colors: [Color.blue.opacity(0.6), Color.purple.opacity(0.6)],
                    startPoint: .topLeading,
                    endPoint: .bottomTrailing
                )
                .ignoresSafeArea()

                VStack(spacing: 20) {
                    Spacer()

                    // Logo/Title
                    VStack(spacing: 8) {
                        Image(systemName: "gamecontroller.fill")
                            .font(.system(size: 60))
                            .foregroundColor(.white)

                        Text("PubGames")
                            .font(.largeTitle)
                            .fontWeight(.bold)
                            .foregroundColor(.white)

                        Text("Mini Apps")
                            .font(.subheadline)
                            .foregroundColor(.white.opacity(0.8))
                    }
                    .padding(.bottom, 40)

                    // Login Form
                    VStack(spacing: 16) {
                        TextField("Username", text: $username)
                            .textFieldStyle(RoundedTextFieldStyle())
                            .autocapitalization(.none)
                            .autocorrectionDisabled()

                        SecureField("Password", text: $password)
                            .textFieldStyle(RoundedTextFieldStyle())

                        if let errorMessage = errorMessage {
                            Text(errorMessage)
                                .font(.caption)
                                .foregroundColor(.red)
                                .padding(.horizontal)
                        }

                        Button(action: handleLogin) {
                            HStack {
                                if isLoading {
                                    ProgressView()
                                        .progressViewStyle(CircularProgressViewStyle(tint: .white))
                                } else {
                                    Text("Login")
                                        .fontWeight(.semibold)
                                }
                            }
                            .frame(maxWidth: .infinity)
                            .padding()
                            .background(Color.white.opacity(0.2))
                            .foregroundColor(.white)
                            .cornerRadius(10)
                        }
                        .disabled(isLoading || username.isEmpty || password.isEmpty)

                        // Future: Face ID Button
                        // Button(action: handleFaceID) {
                        //     HStack {
                        //         Image(systemName: "faceid")
                        //         Text("Login with Face ID")
                        //     }
                        // }

                        Divider()
                            .background(Color.white.opacity(0.5))
                            .padding(.vertical, 8)

                        Button(action: { showRegister = true }) {
                            Text("Don't have an account? Register")
                                .foregroundColor(.white)
                                .fontWeight(.medium)
                        }
                    }
                    .padding(.horizontal, 40)

                    Spacer()

                    // Server URL (for debugging)
                    Text("Server: \(Config.identityBackendURL)")
                        .font(.caption2)
                        .foregroundColor(.white.opacity(0.6))
                        .padding(.bottom, 8)
                }
            }
            .navigationBarHidden(true)
            .sheet(isPresented: $showRegister) {
                RegisterView(isPresented: $showRegister)
            }
        }
    }

    private func handleLogin() {
        errorMessage = nil
        isLoading = true

        Task {
            do {
                _ = try await authService.login(username: username, password: password)
                // Navigation happens automatically via authService.isAuthenticated
            } catch {
                errorMessage = error.localizedDescription
                isLoading = false
            }
        }
    }
}

// MARK: - Custom TextField Style

struct RoundedTextFieldStyle: TextFieldStyle {
    func _body(configuration: TextField<Self._Label>) -> some View {
        configuration
            .padding()
            .background(Color.white.opacity(0.2))
            .cornerRadius(10)
            .foregroundColor(.white)
            .overlay(
                RoundedRectangle(cornerRadius: 10)
                    .stroke(Color.white.opacity(0.5), lineWidth: 1)
            )
    }
}

#Preview {
    LoginView()
}
