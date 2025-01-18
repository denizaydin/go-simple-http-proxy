
```

# go-simple-http-proxy

## Overview

This project implements a simple HTTP proxy server in Go. The proxy server forwards incoming HTTP requests to a target server and returns the response to the client. It also adds custom headers to the forwarded requests for debugging and tracing purposes.

## Environment Variables

The proxy server can be configured using the following environment variables:

- `TARGET_PORT`: The port of the target server to which requests will be forwarded. Default is `8080`.
- `TARGET_DESTINATION`: The address of the target server to which requests will be forwarded. Default is `localhost`.
- `PROXY_NODE_NAME`: The name of the proxy node. Default is `default-proxy-node`.
- `NODE_NAME`: The name of the node where the proxy is running. Optional.
- `POD_NAME`: The name of the pod where the proxy is running. Optional.

## How the Code Works

1. **Initialization**: The `init` function initializes the environment variables with their default values if they are not set.

2. **Handling Requests**: The `proxyHandler` function handles incoming HTTP requests:
   - It detects the protocol (HTTP or HTTPS) and prints the request details.
   - It constructs the target URL using the `TARGET_DESTINATION` and `TARGET_PORT` environment variables.
   - It prepares a new HTTP request for the target server, copying headers from the original request and adding custom proxy headers.
   - It performs the request to the target server using an HTTP client.
   - It copies the response from the target server back to the client.

3. **Custom Headers**: The `addCustomHeaders` function adds custom headers to the forwarded request, including:
   - `X-Proxied-Client-Agent`: The user agent of the client.
   - `X-Proxied-Client-Destination`: The destination address of the client.
   - `X-Proxied-Client-Source`: The source address of the client.
   - `X-Proxy-Node`: The name of the proxy node.
   - `X-Proxy-Pod`: The name of the pod where the proxy is running.
   - `X-Proxy-Host`: The hostname of the machine where the proxy is running.

4. **Error Handling**: The `handleError` function constructs and sends an error response to the client if an error occurs during the proxying process. The error response includes details such as the destination address, full URL, node name, pod name, hostname, and the error message.

5. **Starting the Server**: The `main` function starts the HTTP proxy server on port 80 and uses the `proxyHandler` function to handle incoming requests.

## Running the Proxy Server

To run the proxy server, set the necessary environment variables and execute the Go program:

```sh
export TARGET_PORT=8080
export TARGET_DESTINATION=localhost
export PROXY_NODE_NAME=my-proxy-node
export NODE_NAME=my-node
export POD_NAME=my-pod

go run go-web-proxy.go