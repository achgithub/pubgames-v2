//
//  WebViewContainer.swift
//  PubGamesMiniApps
//
//  WebView container for loading mini apps with token injection
//

import SwiftUI
import WebKit

struct WebViewContainer: View {
    let app: MiniApp
    @Binding var isPresented: MiniApp?

    @StateObject private var viewModel = WebViewModel()
    @StateObject private var authService = AuthService.shared

    var body: some View {
        NavigationView {
            ZStack {
                WebView(
                    url: buildAppURL(),
                    viewModel: viewModel
                )
                .ignoresSafeArea()

                if viewModel.isLoading {
                    ProgressView()
                }

                if let error = viewModel.errorMessage {
                    VStack {
                        Spacer()
                        HStack {
                            Image(systemName: "exclamationmark.triangle")
                            Text(error)
                        }
                        .padding()
                        .background(Color.red.opacity(0.8))
                        .foregroundColor(.white)
                        .cornerRadius(8)
                        .padding()
                    }
                }
            }
            .navigationTitle(app.name)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarLeading) {
                    Button(action: {
                        isPresented = nil
                    }) {
                        Image(systemName: "xmark.circle.fill")
                            .foregroundColor(.gray)
                    }
                }

                ToolbarItem(placement: .navigationBarTrailing) {
                    HStack(spacing: 16) {
                        Button(action: {
                            viewModel.webView?.reload()
                        }) {
                            Image(systemName: "arrow.clockwise")
                        }

                        if viewModel.canGoBack {
                            Button(action: {
                                viewModel.webView?.goBack()
                            }) {
                                Image(systemName: "chevron.left")
                            }
                        }

                        if viewModel.canGoForward {
                            Button(action: {
                                viewModel.webView?.goForward()
                            }) {
                                Image(systemName: "chevron.right")
                            }
                        }
                    }
                }
            }
        }
    }

    private func buildAppURL() -> URL {
        guard let token = authService.getToken() else {
            return URL(string: app.fullURL)!
        }

        // Append token as query parameter for SSO
        var components = URLComponents(string: app.fullURL)
        components?.queryItems = [URLQueryItem(name: "token", value: token)]

        return components?.url ?? URL(string: app.fullURL)!
    }
}

// MARK: - WebView (UIViewRepresentable)

struct WebView: UIViewRepresentable {
    let url: URL
    @ObservedObject var viewModel: WebViewModel

    func makeUIView(context: Context) -> WKWebView {
        let config = WKWebViewConfiguration()

        // Enable JavaScript
        config.preferences.javaScriptEnabled = true

        // Allow inline media playback
        config.allowsInlineMediaPlayback = true

        // Future: Add message handlers for Apple Pay, Face ID
        // let contentController = WKUserContentController()
        // contentController.add(context.coordinator, name: "applePay")
        // contentController.add(context.coordinator, name: "faceID")
        // config.userContentController = contentController

        let webView = WKWebView(frame: .zero, configuration: config)
        webView.navigationDelegate = context.coordinator
        webView.allowsBackForwardNavigationGestures = true

        // Store reference in view model
        viewModel.webView = webView

        return webView
    }

    func updateUIView(_ webView: WKWebView, context: Context) {
        // Only load if not already loaded
        if webView.url == nil {
            let request = URLRequest(url: url)
            webView.load(request)
        }
    }

    func makeCoordinator() -> Coordinator {
        Coordinator(viewModel: viewModel)
    }

    // MARK: - Coordinator

    class Coordinator: NSObject, WKNavigationDelegate {
        let viewModel: WebViewModel

        init(viewModel: WebViewModel) {
            self.viewModel = viewModel
        }

        func webView(_ webView: WKWebView, didStartProvisionalNavigation navigation: WKNavigation!) {
            viewModel.isLoading = true
            viewModel.errorMessage = nil
        }

        func webView(_ webView: WKWebView, didFinish navigation: WKNavigation!) {
            viewModel.isLoading = false
            viewModel.canGoBack = webView.canGoBack
            viewModel.canGoForward = webView.canGoForward

            // Inject native bridge script after page loads
            injectNativeBridge(webView)
        }

        func webView(_ webView: WKWebView, didFail navigation: WKNavigation!, withError error: Error) {
            viewModel.isLoading = false
            viewModel.errorMessage = "Failed to load: \(error.localizedDescription)"
        }

        func webView(_ webView: WKWebView, didFailProvisionalNavigation navigation: WKNavigation!, withError error: Error) {
            viewModel.isLoading = false
            viewModel.errorMessage = "Failed to load: \(error.localizedDescription)"
        }

        private func injectNativeBridge(_ webView: WKWebView) {
            let script = """
            // Native app indicator
            window.NATIVE_APP = true;
            window.NATIVE_PLATFORM = 'ios';

            // Future: Native feature bridges
            // window.nativeApplePay = function(amount, description) {
            //     window.webkit.messageHandlers.applePay.postMessage({amount, description});
            // };
            //
            // window.nativeFaceID = function() {
            //     return new Promise((resolve, reject) => {
            //         window.webkit.messageHandlers.faceID.postMessage({});
            //         // Handle response via another message handler
            //     });
            // };

            console.log('PubGames iOS Native Bridge loaded');
            """

            webView.evaluateJavaScript(script) { result, error in
                if let error = error {
                    print("Error injecting native bridge: \(error.localizedDescription)")
                }
            }
        }
    }
}

// MARK: - WebViewModel

class WebViewModel: ObservableObject {
    @Published var isLoading = false
    @Published var canGoBack = false
    @Published var canGoForward = false
    @Published var errorMessage: String?

    weak var webView: WKWebView?
}

#Preview {
    WebViewContainer(
        app: MiniApp(
            id: 1,
            name: "Test App",
            url: "https://www.apple.com",
            description: "Test description",
            iconName: nil,
            isActive: true,
            order: 1
        ),
        isPresented: .constant(nil)
    )
}
