package main

// auth.go
//
// This file uses the shared authentication library from pubgames/shared/auth
//
// The shared library provides:
// - AuthMiddleware: Validates JWT tokens with Identity Service
// - AdminMiddleware: Ensures user has admin privileges
//
// No need to reimplement authentication logic here.
// Identity Service handles token generation.
// Shared library handles token validation.
//
// Usage examples:
//
//   Protected route (requires valid token):
//   api.HandleFunc("/data", authMw(handler)).Methods("GET")
//
//   Admin route (requires valid token + admin flag):
//   api.HandleFunc("/admin", authMw(adminMw(handler))).Methods("POST")
//
// See main.go for implementation examples.
