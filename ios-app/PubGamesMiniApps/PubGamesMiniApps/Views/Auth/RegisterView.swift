//
//  RegisterView.swift
//  PubGamesMiniApps
//
//  Native registration screen
//

import SwiftUI

struct RegisterView: View {
    @Binding var isPresented: Bool
    @StateObject private var authService = AuthService.shared

    @State private var username = ""
    @State private var email = ""
    @State private var password = ""
    @State private var confirmPassword = ""
    @State private var isLoading = false
    @State private var errorMessage: String?

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

                ScrollView {
                    VStack(spacing: 20) {
                        // Header
                        VStack(spacing: 8) {
                            Image(systemName: "person.badge.plus")
                                .font(.system(size: 50))
                                .foregroundColor(.white)

                            Text("Create Account")
                                .font(.largeTitle)
                                .fontWeight(.bold)
                                .foregroundColor(.white)
                        }
                        .padding(.top, 40)
                        .padding(.bottom, 20)

                        // Registration Form
                        VStack(spacing: 16) {
                            TextField("Username", text: $username)
                                .textFieldStyle(RoundedTextFieldStyle())
                                .autocapitalization(.none)
                                .autocorrectionDisabled()

                            TextField("Email (optional)", text: $email)
                                .textFieldStyle(RoundedTextFieldStyle())
                                .autocapitalization(.none)
                                .keyboardType(.emailAddress)

                            SecureField("Password", text: $password)
                                .textFieldStyle(RoundedTextFieldStyle())

                            SecureField("Confirm Password", text: $confirmPassword)
                                .textFieldStyle(RoundedTextFieldStyle())

                            if let errorMessage = errorMessage {
                                Text(errorMessage)
                                    .font(.caption)
                                    .foregroundColor(.red)
                                    .padding(.horizontal)
                            }

                            Button(action: handleRegister) {
                                HStack {
                                    if isLoading {
                                        ProgressView()
                                            .progressViewStyle(CircularProgressViewStyle(tint: .white))
                                    } else {
                                        Text("Register")
                                            .fontWeight(.semibold)
                                    }
                                }
                                .frame(maxWidth: .infinity)
                                .padding()
                                .background(Color.white.opacity(0.2))
                                .foregroundColor(.white)
                                .cornerRadius(10)
                            }
                            .disabled(isLoading || !isFormValid)

                            Button(action: { isPresented = false }) {
                                Text("Already have an account? Login")
                                    .foregroundColor(.white)
                                    .fontWeight(.medium)
                            }
                            .padding(.top, 8)
                        }
                        .padding(.horizontal, 40)
                    }
                }
            }
            .navigationBarHidden(true)
        }
    }

    private var isFormValid: Bool {
        !username.isEmpty &&
        !password.isEmpty &&
        password == confirmPassword &&
        password.count >= 6
    }

    private func handleRegister() {
        errorMessage = nil

        // Validate passwords match
        guard password == confirmPassword else {
            errorMessage = "Passwords do not match"
            return
        }

        // Validate password length
        guard password.count >= 6 else {
            errorMessage = "Password must be at least 6 characters"
            return
        }

        isLoading = true

        Task {
            do {
                let emailValue = email.isEmpty ? nil : email
                _ = try await authService.register(
                    username: username,
                    password: password,
                    email: emailValue
                )
                // Close sheet and navigate to launcher
                isPresented = false
            } catch {
                errorMessage = error.localizedDescription
                isLoading = false
            }
        }
    }
}

#Preview {
    RegisterView(isPresented: .constant(true))
}
