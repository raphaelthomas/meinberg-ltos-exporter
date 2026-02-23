// Copyright 2026 Raphael Seebacher
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Mock server utility for testing the Meinberg LTOS exporter.
// This program reads api-status-response.json and serves it on a local HTTP server.
// Usage: go run mock-server.go [-addr localhost] [-port 8080]

//go:build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	addr := flag.String("addr", "localhost", "Address to listen on")
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()

	// Read the JSON file
	jsonFile, err := os.Open("tests/mock-api-status-response.json")
	if err != nil {
		log.Fatalf("Failed to open api-status-response.json: %v", err)
	}
	defer jsonFile.Close()

	// Read and parse JSON to validate it
	var response any
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&response); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	// Reset file pointer for serving
	if _, err := jsonFile.Seek(0, io.SeekStart); err != nil {
		log.Fatalf("Failed to seek: %v", err)
	}

	// Read the JSON data into memory
	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Fatalf("Failed to read JSON: %v", err)
	}

	// Create a handler for /api/status
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonData)
	})

	// Create a simple index page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
  <title>Meinberg LTOS Mock Server</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 20px; }
    pre { background-color: #f0f0f0; padding: 10px; overflow-x: auto; }
  </style>
</head>
<body>
  <h1>Meinberg LTOS Mock Server</h1>
  <p>This server simulates the Meinberg LTOS API for testing purposes.</p>
  <h2>Available Endpoints</h2>
  <ul>
    <li><a href="/api/status">/api/status</a> - Returns the device status JSON</li>
  </ul>
  <h2>Example Usage</h2>
  <pre>
# Start the exporter pointing to this mock server:
./meinberg_ltos_exporter --ltos-api-url http://localhost:%s

# In another terminal, fetch metrics:
curl http://localhost:10123/metrics
  </pre>
</body>
</html>
`, *port)
	})

	listenAddr := fmt.Sprintf("%s:%s", *addr, *port)
	log.Printf("Mock server listening on http://%s", listenAddr)
	log.Printf("API endpoint available at http://%s/api/status", listenAddr)

	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
