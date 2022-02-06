# tc4400_exporter 

[![Go Reference](https://pkg.go.dev/badge/github.com/blainsmith/tc4400_exporter.svg)](https://pkg.go.dev/github.com/blainsmith/tc4400_exporter)

Command `tc4400_exporter` provides a Prometheus exporter for the [Technicolor](https://www.technicolor.com/) TC4400 Cable Modem. MIT Licensed.

## Config

```yaml
listen:
  address: ":9130"
  metricspath: "/metrics"
modems:
  -
    address: "http://192.168.100.1"
    username: "user"
    password: "pass"
    timeout: "5s"
  -
    address: "http://10.0.100.1"
    username: "admin"
    password: "secret"
    timeout: "5s"
```

The `tc4400_exporter`'s Prometheus scrape configuration (`prometheus.yml`) is configured in a similar way to the official Prometheus [`blackbox_exporter`](https://github.com/prometheus/blackbox_exporter).

The `targets` list under `static_configs` should specify the addresses of any TC4400 modems which should be monitored by the exporter. The address of the `tc4400_exporter` itself must be specified in `relabel_configs` as well.

```yaml
scrape_configs:
  - job_name: 'tc4400'
    static_configs:
      - targets:
        - 'http://192.168.100.1' # TC4400 modem.
        - 'http://10.0.100.1' # TC4400 modem.
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: '127.0.0.1:9393' # tc4400_exporter.
```

## Usage

```
$ ./tc4400_exporter -config.file ./config.yaml
2022/02/05 16:18:06 starting TC4400 exporter on ":9130"
```

The exporter is now running at http://localhost:9130/metrics?target=http://192.168.100.1 and notice the `target` query param matches one of the `modems` in the config to support running a single `tc4400_exporter` for multiple TC4400 modems.

