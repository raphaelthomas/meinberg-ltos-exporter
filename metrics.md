# Proposed Prometheus Metrics for Meinberg LTOS (M600 Model)

This document details the Prometheus metrics designed for the Meinberg LTOS M600 device based on the provided API response. These metrics are prefixed with `mbg_ltos_` and adhere to the best practices detailed in the Prometheus documentation.

---

## Build Information Metrics

| Metric Name                     | Description                                      | Type    | API Path                                          |
|---------------------------------|--------------------------------------------------|---------|--------------------------------------------------|
| `mbg_ltos_build_info`           | Build information (e.g., API, firmware version). Labels: `api_version`, `firmware_version`, `serial_number`, `hostname` | Gauge   | `system-information`                            |

---

## System Information Metrics

| Metric Name                         | Description                                                | Type    | API Path             |
|-------------------------------------|------------------------------------------------------------|---------|----------------------|
| `mbg_ltos_system_uptime_seconds`    | System uptime in seconds                                   | Gauge   | `data.system.uptime` |
| `mbg_ltos_system_cpu_load_avg`      | CPU load averaged over 1, 5, and 15 minutes.              | Gauge   | `data.system.cpuload`|
| `mbg_ltos_system_memory_bytes`      | Total memory in bytes.                                    | Gauge   | `data.system.memory` |
| `mbg_ltos_system_memory_free_bytes` | Free memory in bytes.                                     | Gauge   | `data.system.memory` |

---

## Storage Metrics

| Metric Name                            | Description                                             | Type    | API Path              |
|----------------------------------------|---------------------------------------------------------|---------|-----------------------|
| `mbg_ltos_storage_capacity_bytes{device}`  | Total size of the storage device in bytes.             | Gauge   | `data.system.storage` |
| `mbg_ltos_storage_used_bytes{device}`      | Used storage on the device in bytes.                   | Gauge   | `data.system.storage` |
| `mbg_ltos_storage_free_bytes{device}`      | Free storage available on the device in bytes.         | Gauge   | `data.system.storage` |
| `mbg_ltos_storage_usage_ratio{device}`     | Ratio of space used on the storage device (0-1).       | Gauge   | `data.system.storage` |

---

## GPS Positioning Metrics

| Metric Name                          | Description                              | Type    | API Path                     |
|--------------------------------------|------------------------------------------|---------|------------------------------|
| `mbg_ltos_gps_position_x_meters`     | GPS X-coordinate in meters.             | Gauge   | `data.system.chassis0.clk1.satellites.position-x` |
| `mbg_ltos_gps_position_y_meters`     | GPS Y-coordinate in meters.             | Gauge   | `data.system.chassis0.clk1.satellites.position-y` |
| `mbg_ltos_gps_position_z_meters`     | GPS Z-coordinate in meters.             | Gauge   | `data.system.chassis0.clk1.satellites.position-z` |
| `mbg_ltos_gps_latitude_degrees`      | Latitude in degrees.                    | Gauge   | `data.system.chassis0.clk1.satellites.latitude`   |
| `mbg_ltos_gps_longitude_degrees`     | Longitude in degrees.                   | Gauge   | `data.system.chassis0.clk1.satellites.longitude`  |
| `mbg_ltos_gps_altitude_meters`       | Altitude above sea level in meters.     | Gauge   | `data.system.chassis0.clk1.satellites.altitude`   |

---

## Synchronization Status Metrics

| Metric Name                                  | Description                                           | Type    | API Path                                   |
|----------------------------------------------|-------------------------------------------------------|---------|--------------------------------------------|
| `mbg_ltos_sync_clock_status`                 | Status of clock synchronization.                     | Gauge   | `data.system.sync-status.clock-status`     |
| `mbg_ltos_sync_leapsecond_announcement`      | If a leapsecond is announced.                        | Boolean | `data.system.sync-status.leapsecond-announced` |
| `mbg_ltos_sync_time_quality_nanoseconds`     | Estimated quality of time in nanoseconds.            | Gauge   | `data.system.sync-status.est-time-quality` |

---

## Front Panel LED Metrics (Optional)

| Metric Name                                | Description                                           | Type    | API Path                          |
|--------------------------------------------|-------------------------------------------------------|---------|-----------------------------------|
| `mbg_ltos_led_active{led}`                 | Current active state (on/off) of LEDs on the front panel.  | Boolean | `data.system.front-leds`          |
| `mbg_ltos_led_color_status{led}`           | Current LED color (green/red/none).                 | Gauge   | `data.system.front-leds`          |

---

## Network Metrics

| Metric Name                                | Description                                           | Type    | API Path                            |
|--------------------------------------------|-------------------------------------------------------|---------|-------------------------------------|
| `mbg_ltos_network_receive_bytes{interface}`| Received bytes on a network interface.              | Counter | `data.system.network.ports`        |
| `mbg_ltos_network_transmit_bytes{interface}`| Transmitted bytes on a network interface.           | Counter | `data.system.network.ports`        |
| `mbg_ltos_network_receive_errors{interface}`| Errors encountered on receive.                      | Counter | `data.system.network.ports`        |
| `mbg_ltos_network_transmit_errors{interface}`| Errors encountered on transmit.                     | Counter | `data.system.network.ports`        |
| `mbg_ltos_network_port_duplex_mode{port}`   | Duplex mode of the network port.                   | Gauge   | `data.system.network.ports`        |
| `mbg_ltos_network_port_speed_mbps{port}`    | Operating speed of the network port in Mbps.        | Gauge   | `data.system.network.ports`        |

---

## NTP Metrics

| Metric Name                                         | Description                          | Type    | API Path                                |
|-----------------------------------------------------|--------------------------------------|---------|----------------------------------------|
| `mbg_ltos_ntp_stratum_level{clock}`                 | Stratum level of the NTP clock.      | Gauge   | `ntp[].stratum`                        |
| `mbg_ltos_ntp_leap_indicator_status{clock}`         | Leap indicator of the NTP clock.     | Gauge   | `ntp[].leap`                           |
| `mbg_ltos_ntp_clock_precision_seconds{clock}`       | Clock precision for the NTP instance.| Gauge   | `ntp[].precision`                      |
| `mbg_ltos_ntp_reference_id{clock}`                  | Reference source ID for the NTP instance.| String | `ntp[].refid`                      |
| `mbg_ltos_ntp_root_dispersion_seconds{clock}`       | Root clock dispersion error.         | Gauge   | `ntp[].rootdisp`                       |
| `mbg_ltos_ntp_root_delay_seconds{clock}`            | Delay introduced by the root clock.  | Gauge   | `ntp[].rootdelay`                      |
| `mbg_ltos_ntp_offset_seconds{clock}`                | Clock offset in synchronization.     | Gauge   | `ntp[].offset`                         |
| `mbg_ltos_ntp_clock_wander_seconds{clock}`          | Clock wander measurement.            | Gauge   | `ntp[].clk-wander`                     |

---
