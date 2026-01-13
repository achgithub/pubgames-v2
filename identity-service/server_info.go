package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

// getServerInfoHandler returns server information including local IP
func getServerInfoHandler(w http.ResponseWriter, r *http.Request) {
	localIP := getLocalIP()
	
	response := map[string]interface{}{
		"local_ip":      localIP,
		"frontend_port": FRONTEND_PORT,
		"backend_port":  BACKEND_PORT,
		"qr_url":        fmt.Sprintf("http://%s:%s", localIP, FRONTEND_PORT),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getLocalIP returns the local IP address of the server
func getLocalIP() string {
	// Try to find local IP by connecting to a remote address
	// This doesn't actually make a connection, just determines which interface would be used
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Fallback: try to find first non-loopback interface
		return getLocalIPFallback()
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// getLocalIPFallback finds local IP by checking all interfaces
func getLocalIPFallback() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "localhost"
}
