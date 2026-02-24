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

package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config holds the exporter configuration
type Config struct {
	ListenAddr      string
	ListenPort      string
	LTOSAPIURL      string
	Timeout         time.Duration
	LogLevel        slog.Level
	AuthBasicUser   string
	AuthBasicPass   string
	IgnoreSSLVerify bool
}

// parseFlags parses command-line flags using kingpin
func parseFlags() *Config {
	app := kingpin.New("meinberg_ltos_exporter", "Prometheus exporter for Meinberg LTOS devices")
	app.Version("0.1.0")
	app.HelpFlag.Short('h')

	cfg := &Config{}

	// Command-line flags with kingpin
	app.Flag("listen-addr", "Address to listen on").
		Default("localhost").
		StringVar(&cfg.ListenAddr)

	app.Flag("listen-port", "Port to listen on").
		Default("10123").
		StringVar(&cfg.ListenPort)

	app.Flag("ltos-api-url", "URL of the Meinberg LTOS API").
		Required().
		StringVar(&cfg.LTOSAPIURL)

	app.Flag("timeout", "Timeout for HTTP requests to Meinberg device").
		Default("10s").
		DurationVar(&cfg.Timeout)

	logLevelFlag := app.Flag("log-level", "Log level (debug, info, warn, error)").
		Default("info").
		Enum("debug", "info", "warn", "error")

	app.Flag("auth-user", "Basic auth username").
		StringVar(&cfg.AuthBasicUser)

	app.Flag("auth-pass", "Basic auth password").
		StringVar(&cfg.AuthBasicPass)

	app.Flag("ignore-ssl-verify", "Ignore SSL certificate verification").
		BoolVar(&cfg.IgnoreSSLVerify)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Convert log level string to slog.Level
	switch *logLevelFlag {
	case "debug":
		cfg.LogLevel = slog.LevelDebug
	case "info":
		cfg.LogLevel = slog.LevelInfo
	case "warn":
		cfg.LogLevel = slog.LevelWarn
	case "error":
		cfg.LogLevel = slog.LevelError
	default:
		cfg.LogLevel = slog.LevelInfo
	}

	// Override with environment variables if set
	if url := os.Getenv("LTOS_API_URL"); url != "" {
		cfg.LTOSAPIURL = url
	}
	if addr := os.Getenv("LISTEN_ADDR"); addr != "" {
		cfg.ListenAddr = addr
	}
	if port := os.Getenv("LISTEN_PORT"); port != "" {
		cfg.ListenPort = port
	}
	if timeout := os.Getenv("TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			cfg.Timeout = d
		}
	}
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		switch level {
		case "debug":
			cfg.LogLevel = slog.LevelDebug
		case "info":
			cfg.LogLevel = slog.LevelInfo
		case "warn":
			cfg.LogLevel = slog.LevelWarn
		case "error":
			cfg.LogLevel = slog.LevelError
		}
	}
	if user := os.Getenv("AUTH_USER"); user != "" {
		cfg.AuthBasicUser = user
	}
	if pass := os.Getenv("AUTH_PASS"); pass != "" {
		cfg.AuthBasicPass = pass
	}
	if ignoreSSL := os.Getenv("IGNORE_SSL_VERIFY"); ignoreSSL != "" {
		if value, err := strconv.ParseBool(ignoreSSL); err == nil {
			cfg.IgnoreSSLVerify = value
		}
	}

	return cfg
}

// registerMetrics registers Prometheus metrics
func registerMetrics(client *Client, logger *slog.Logger) error {
	collector := NewCollector(client, logger)
	return collector.Register()
}

func main() {
	cfg := parseFlags()

	// Initialize structured logger
	logLevel := &slog.LevelVar{}
	logLevel.Set(cfg.LogLevel)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	logger.Info("Starting Meinberg LTOS Exporter",
		"version", "0.1.0",
		"listen_addr", cfg.ListenAddr,
		"listen_port", cfg.ListenPort,
	)

	// Create Meinberg API client
	client := NewClient(cfg.LTOSAPIURL, cfg.Timeout, cfg.AuthBasicUser, cfg.AuthBasicPass, cfg.IgnoreSSLVerify)

	// Register metrics
	if err := registerMetrics(client, logger); err != nil {
		logger.Error("Failed to register metrics", "error", err)
		os.Exit(1)
	}

	// Register the /metrics handler
	http.Handle("/metrics", promhttp.Handler())

	// Create a simple index page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
  <title>Meinberg LTOS Exporter</title>
</head>
<body>
  <h1>Meinberg LTOS Exporter</h1>
  <p>Prometheus exporter for Meinberg LTOS devices.</p>
	<p>Check <a href="/metrics">/metrics</a> for the Prometheus metrics in text exposition format scraped from %s.</p>
</body>
</html>
`, cfg.LTOSAPIURL)
	})

	listenAddr := fmt.Sprintf("%s:%s", cfg.ListenAddr, cfg.ListenPort)
	logger.Info("HTTP server listening", "address", listenAddr)

	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		logger.Error("HTTP server error", "error", err)
		os.Exit(1)
	}
}
