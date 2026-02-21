# Prometheus Exporter for Meinberg LTOS

This is Prometheus exporter `meinberg_ltos_exporter` for Meinberg devices
running LTOS. It fetches the status of a LTOS device via its REST API and
exposes the status as scrape-able [metrics](./metrics.md).

## Building

FIXME

## Configuration

- location of the meinberg device
- timeout for fetching data from the REST API
- logging: level and format
- listen address and port where `/metrics` is exposed: `localhost:10123`
- environment variables

### Authentication

- Basic Authentication or Bearer token
- User with level "info" (lowest permission level) to be configured on the LTOS
device.
