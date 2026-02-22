# Prometheus Exporter for Meinberg LTOS

This exporter is experimental and has only been tested against a Meinberg M600 device running LTOS 7.10.008.

This Prometheus exporter, `meinberg_ltos_exporter`, is designed for Meinberg devices running LTOS. It retrieves the status of a device via its REST API and makes the data available as scrape-able [Prometheus metrics](./metrics.md).

## Building

To build the exporter, run the following command:

```sh
make build
```

This will create an executable named `meinberg_ltos_exporter`.

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
      --listen-addr="localhost"  Address to listen on
      --listen-port="10123"      Port to listen on
      --ltos-api-url=LTOS-API-URL
                                 URL of the Meinberg LTOS API
      --timeout=10s              Timeout for HTTP requests to Meinberg device
      --log-level=info           Log level (debug, info, warn, error)
      --auth-user=AUTH-USER      Basic auth username
      --auth-pass=AUTH-PASS      Basic auth password
```

These parameters can be provided as environment variables or command-line arguments.

### Authentication

The exporter supports Basic Authentication. Ensure the user has the "info" access level (lowest permission level) configured on the LTOS device.
