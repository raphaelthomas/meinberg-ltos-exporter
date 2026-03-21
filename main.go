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
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/ltosapi"
)

// Config holds the exporter configuration
type Config struct {
	ListenAddr      string
	ListenPort      string
	MetricsPath     string
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
	app.Version(Version)
	app.HelpFlag.Short('h')

	cfg := &Config{}

	EnvPrefix := "MEINBERG_LTOS_EXPORTER_"

	app.Flag("web.listen-address", "Address to listen on").
		Default("localhost").
		Envar(EnvPrefix + "LISTEN_ADDR").
		StringVar(&cfg.ListenAddr)

	app.Flag("web.listen-port", "Port to listen on").
		Default("10123").
		Envar(EnvPrefix + "LISTEN_PORT").
		StringVar(&cfg.ListenPort)

	app.Flag("web.telemetry-path", "Path under which to expose metrics").
		Default("/metrics").
		Envar(EnvPrefix + "METRICS_PATH").
		StringVar(&cfg.MetricsPath)

	app.Flag("ltos-api-url", "URL of the Meinberg LTOS API").
		Required().
		Envar(EnvPrefix + "LTOS_API_URL").
		StringVar(&cfg.LTOSAPIURL)

	app.Flag("auth-user", "Basic auth username").
		Envar(EnvPrefix + "AUTH_USER").
		StringVar(&cfg.AuthBasicUser)

	app.Flag("auth-pass", "Basic auth password").
		Envar(EnvPrefix + "AUTH_PASS").
		StringVar(&cfg.AuthBasicPass)

	app.Flag("timeout", "Timeout for HTTP requests to Meinberg device").
		Default("5s").
		Envar(EnvPrefix + "TIMEOUT").
		DurationVar(&cfg.Timeout)

	app.Flag("ignore-ssl-verify", "Ignore SSL certificate verification").
		Default("false").
		Envar(EnvPrefix + "IGNORE_SSL_VERIFY").
		BoolVar(&cfg.IgnoreSSLVerify)

	logLevelFlag := app.Flag("log-level", "Log level (debug, info, warn, error)").
		Default("info").
		Enum("debug", "info", "warn", "error")

	kingpin.MustParse(app.Parse(os.Args[1:]))

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

	return cfg
}

// registerMetrics registers Prometheus metrics
func registerMetrics(client *ltosapi.Client, logger *slog.Logger) error {
	collector := NewCollector(client, logger)
	return collector.Register()
}

func main() {
	cfg := parseFlags()

	logLevel := &slog.LevelVar{}
	logLevel.Set(cfg.LogLevel)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	logger.Info("Starting Meinberg LTOS Exporter",
		"version", Version,
		"listen_addr", cfg.ListenAddr,
		"listen_port", cfg.ListenPort,
	)

	prometheus.MustRegister(versioncollector.NewCollector("meinberg_exporter"))

	client := ltosapi.NewClient(cfg.LTOSAPIURL, cfg.Timeout, cfg.AuthBasicUser, cfg.AuthBasicPass, cfg.IgnoreSSLVerify, logger)

	if err := registerMetrics(client, logger); err != nil {
		logger.Error("Failed to register metrics", "error", err)
		os.Exit(1)
	}

	http.Handle(cfg.MetricsPath, promhttp.Handler())

	if cfg.MetricsPath != "/" && cfg.MetricsPath != "" {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			if _, err := fmt.Fprintf(w, `
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
		`, cfg.LTOSAPIURL); err != nil {
				logger.Error("Failed to write response", slog.String("error", err.Error()))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		})
	}

	listenAddr := fmt.Sprintf("%s:%s", cfg.ListenAddr, cfg.ListenPort)
	logger.Info("HTTP server listening", "address", listenAddr)

	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		logger.Error("HTTP server error", "error", err)
		os.Exit(1)
	}
}
