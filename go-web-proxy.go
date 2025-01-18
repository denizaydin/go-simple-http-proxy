package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	defaultTargetPort        = "8080"
	defaultTargetDestination = "localhost"
	defaultRequestTimeout    = 3 * time.Second // Timeout for requests to the target server
)

var (
	targetPort        string
	targetDestination string
	proxyNodeName     string
	nodeName          string
	podName           string
	hostname          string
)

func init() {
	// Initialize environment variables with defaults
	targetPort = os.Getenv("TARGET_PORT")
	if targetPort == "" {
		targetPort = defaultTargetPort
	}

	targetDestination = os.Getenv("TARGET_DESTINATION")
	if targetDestination == "" {
		targetDestination = defaultTargetDestination
	}

	proxyNodeName = os.Getenv("PROXY_NODE_NAME")
	if proxyNodeName == "" {
		proxyNodeName = "default-proxy-node"
	}

	nodeName = os.Getenv("NODE_NAME")
	podName = os.Getenv("POD_NAME")

	var err error
	hostname, err = os.Hostname()
	if err != nil {
		hostname = ""
	}
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// Detect protocol
	protocol := "HTTP"
	if r.TLS != nil {
		protocol = "HTTPS"
	}
	fmt.Printf("Received a %s request: %s %s\n", protocol, r.Method, r.URL.String())

	// Construct target URL
	targetURL := fmt.Sprintf("http://%s:%s%s", targetDestination, targetPort, r.URL.Path)

	// Prepare a new HTTP request for the target server
	ctx, cancel := context.WithTimeout(r.Context(), defaultRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, r.Method, targetURL, r.Body)
	if err != nil {
		handleError(w, "Failed to create request", targetURL, http.StatusInternalServerError)
		return
	}

	// Copy headers and add custom proxy headers
	copyHeaders(req.Header, r.Header)
	addCustomHeaders(req.Header, r.UserAgent(), r.RemoteAddr, r.Context().Value(http.LocalAddrContextKey))

	// Perform the request to the target server
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			handleError(w, "Request to target server timed out", targetURL, http.StatusGatewayTimeout)
			return
		}
		handleError(w, "Failed to forward request: "+err.Error(), targetURL, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy the response from the target server
	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func copyHeaders(dest, src http.Header) {
	for name, values := range src {
		for _, value := range values {
			dest.Add(name, value)
		}
	}
}

func addCustomHeaders(headers http.Header, userAgent, sourceAddress string, localAddr interface{}) {
	headers.Set("X-Proxied-Client-Agent", userAgent)

	if addr, ok := localAddr.(net.Addr); ok {
		headers.Set("X-Proxied-Client-Destination", addr.String())
	}

	headers.Set("X-Proxied-Client-Source", sourceAddress)

	if nodeName != "" {
		headers.Set("X-Proxy-Node", nodeName)
	}
	if podName != "" {
		headers.Set("X-Proxy-Pod", podName)
	}
	if hostname != "" {
		headers.Set("X-Proxy-Host", hostname)
	}
}

func handleError(w http.ResponseWriter, message, targetURL string, statusCode int) {
	w.WriteHeader(statusCode)

	// Construct the error response
	response := fmt.Sprintf(
		"Destination Address    : %s\nFull URL               : %s\n",
		targetDestination+":"+targetPort, targetURL,
	)

	if nodeName != "" {
		response += fmt.Sprintf("Node Name              : %s\n", nodeName)
	}
	if podName != "" {
		response += fmt.Sprintf("Pod Name               : %s\n", podName)
	}
	if hostname != "" {
		response += fmt.Sprintf("Hostname               : %s\n", hostname)
	}

	response += fmt.Sprintf("\n%s\n", message)
	w.Write([]byte(response))
}

func main() {
	fmt.Println("Starting HTTP proxy server on port 80...")
	if err := http.ListenAndServe(":80", http.HandlerFunc(proxyHandler)); err != nil {
		fmt.Println("HTTP server failed:", err)
	}
}
