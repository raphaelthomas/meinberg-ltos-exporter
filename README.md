# Prometheus Exporter for Meinberg LTOS

[![Go Version](https://img.shields.io/github/go-mod/go-version/raphaelthomas/meinberg-ltos-exporter)](https://github.com/raphaelthomas/meinberg-ltos-exporter/blob/main/go.mod)
[![Latest Release](https://img.shields.io/github/v/release/raphaelthomas/meinberg-ltos-exporter)](https://github.com/raphaelthomas/meinberg-ltos-exporter/releases/latest)
[![License](https://img.shields.io/github/license/raphaelthomas/meinberg-ltos-exporter)](https://github.com/raphaelthomas/meinberg-ltos-exporter/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/raphaelthomas/meinberg-ltos-exporter)](https://goreportcard.com/report/github.com/raphaelthomas/meinberg-ltos-exporter)

Prometheus exporter `meinberg_ltos_exporter` is designed for Meinberg devices
running LTOS. It retrieves the status of a device via its REST API and makes
the data available as scrape-able Prometheus metrics.

> [!WARNING]
> This exporter is **experimental** and has only been tested against a very
> limited number of devices. It may not work correctly with all Meinberg
> devices and LTOS versions.

> [!IMPORTANT]
> Please **provide feedback** through GitHub issues, include the
> anonymized/obfuscated JSON output of `/api/status` to facilitate extending or
> fixing the exporter.

## Supported Meinberg LTOS Devices

The exporter has been tested with the following Meinberg LTOS devices:

| Model | Receiver | LTOS Version |
| ----- | -------- | ------------ |
| M600  | grc180   | 7.10.008     |
| M300  | pzf511   | 7.06.014-light |

## Build

To build the exporter, run the following command, which will create an
executable named `meinberg_ltos_exporter`:

```sh
make build
```

or run `goreleaser` directly:

```sh
goreleaser build --snapshot --clean
```

## Local Development

Run the following in three separate terminal windows:

```sh
make run
go run tests/mock_server.go
curl -s http://localhost:10123/metrics | grep mbg_ltos
```

## Configuration

The exporter can be configured via the following parameters:

``` sh
usage: meinberg_ltos_exporter --ltos-api-url=LTOS-API-URL [<flags>]

Prometheus exporter for Meinberg LTOS devices


Flags:
  -h, --[no-]help                Show context-sensitive help (also try --help-long and --help-man).
      --[no-]version             Show application version.
      --web.listen-address="localhost"
                                 Address to listen on ($MEINBERG_LTOS_EXPORTER_LISTEN_ADDR)
      --web.listen-port="10123"  Port to listen on ($MEINBERG_LTOS_EXPORTER_LISTEN_PORT)
      --web.telemetry-path="/metrics"
                                 Path under which to expose metrics ($MEINBERG_LTOS_EXPORTER_METRICS_PATH)
      --ltos-api-url=LTOS-API-URL
                                 URL of the Meinberg LTOS API ($MEINBERG_LTOS_EXPORTER_LTOS_API_URL)
      --auth-user=AUTH-USER      Basic auth username ($MEINBERG_LTOS_EXPORTER_AUTH_USER)
      --auth-pass=AUTH-PASS      Basic auth password ($MEINBERG_LTOS_EXPORTER_AUTH_PASS)
      --timeout=5s               Timeout for HTTP requests to Meinberg device ($MEINBERG_LTOS_EXPORTER_TIMEOUT)
      --[no-]ignore-ssl-verify   Ignore SSL certificate verification ($MEINBERG_LTOS_EXPORTER_IGNORE_SSL_VERIFY)
      --log-level=info           Log level (debug, info, warn, error)
      --[no-]collector.system    Enable system collector. ($MEINBERG_LTOS_EXPORTER_COLLECTOR_SYSTEM)
      --[no-]collector.notification
                                 Enable notification collector. ($MEINBERG_LTOS_EXPORTER_COLLECTOR_NOTIFICATION)
      --[no-]collector.storage   Enable storage collector. ($MEINBERG_LTOS_EXPORTER_COLLECTOR_STORAGE)
      --[no-]collector.clock     Enable clock collector. ($MEINBERG_LTOS_EXPORTER_COLLECTOR_CLOCK)
      --[no-]collector.receiver  Enable receiver collectors (GNSS + DCF77). ($MEINBERG_LTOS_EXPORTER_COLLECTOR_RECEIVER)
      --[no-]collector.ntp       Enable NTP collector. ($MEINBERG_LTOS_EXPORTER_COLLECTOR_NTP)
```

These parameters can be provided as environment variables or command-line
arguments.

### Authentication

The exporter supports Basic Authentication. Ensure the user has the "info"
access level (lowest permission level) configured on the LTOS device.
