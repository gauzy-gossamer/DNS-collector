<p align="center">
  <img src="https://goreportcard.com/badge/github.com/dmachard/DNS-collector" alt="Go Report"/>
  <img src="https://img.shields.io/badge/go%20version-min%201.23-green" alt="Go version"/>
  <img src="https://img.shields.io/badge/go%20tests-534-green" alt="Go tests"/>
  <img src="https://img.shields.io/badge/go%20bench-21-green" alt="Go bench"/>
  <img src="https://img.shields.io/badge/go%20lines-33707-green" alt="Go lines"/>
</p>

<p align="center">
  <img src="https://img.shields.io/github/v/release/dmachard/DNS-collector?logo=github&sort=semver" alt="release"/>
  <img src="https://img.shields.io/docker/pulls/dmachard/go-dnscollector.svg" alt="docker"/>
</p>

<p align="center">
  <img src="docs/dns-collector_logo.png" alt="DNS-collector"/>
</p>

`DNS-collector` acts as a passive high speed **ingestor** with **pipelining** support for your DNS logs, written in **Golang**. It allows enhancing your DNS logs by adding metadata, extracting usage patterns, and facilitating security analysis.

> Additionally, DNS-collector also support
>
> - [Extended](https://github.com/dmachard/DNS-collector/blob/main/docs/extended_dnstap.md) DNStap with TLS encryption, compression, and more metadata capabilities
> - DNS protocol conversions to [Plain text, Key/Value JSON, Jinja, PCAP and more](https://github.com/dmachard/DNS-collector/blob/main/docs/dnsconversions.md)
> - DNS parser with [Extension Mechanisms for DNS (EDNS)](https://github.com/dmachard/DNS-collector/blob/main/docs/dnsparser.md) support
> - Live capture on a network interface
> - IPv4/v6 defragmentation and TCP reassembly
> - Nanoseconds in timestamps

> The following DNS servers are automatically tested in CI to verify `DNS-collector` compatibility 
> with various DNS servers using [dnstap](https://dnstap.info/).
> | DNS Server     | Versions Tested     | Modes Tested        |
> |----------------|---------------------|---------------------|
> | ✅ **Unbound**     | 1.22.x, 1.21.x     | TCP                 |
> | ✅ **CoreDNS**     | 1.12.1, 1.11.1  | TCP, TLS            |
> | ✅ **PowerDNS DNSdist**     | 2.0.x, 1.9.x, 1.8.x, 1.7.x       | TCP, Unix           |
> | ✅ **Knot Resolver** | 6.0.11           | Unix                |
> | ✅ **Bind** | 9.18.33          | Unix                |


## 🔧 Features

- **[Pipelining](./docs/running_mode.md)**

   The DNS traffic can be collected and aggregated from simultaneously [sources](./docs/workers.md) like DNStap streams, network interface or log files and relays it to multiple other [listeners](./docs/workers.md) 

  [![overview](./docs/_images/overview.png)](./docs/running_mode.md)

  You can also applied  [transformations](./docs/transformers.md) on it like ([traffic filtering](./docs/transformers.md#dns-filtering), [user privacy](./docs/transformers.md#user-privacy), ...).

  [![config](./docs/_images/config.png)](./docs/configuration.md)

- **[Collectors & Loggers](./docs/workers.md)**

  - *Listen for logging traffic with streaming network protocols*
    - [`DNStap`](docs/collectors/collector_dnstap.md#dns-tap) with `tls`|`tcp`|`unix` transports support and [`proxifier`](docs/collectors/collector_dnstap.md#dns-tap-proxifier)
    - [`PowerDNS`](docs/collectors/collector_powerdns.md) streams with full  support
    - [`DNSMessage`](docs/collectors/collector_dnsmessage.md) to route DNS messages based on specific dns fields
    - [`TZSP`](docs/collectors/collector_tzsp.md) protocol support
  - *Live capture on a network interface*
    - [`AF_PACKET`](docs/collectors/collector_afpacket.md) socket with BPF filter and GRE tunnel support
    - [`eBPF XDP`](docs/collectors/collector_xdp.md) ingress traffic
  - *Read text or binary files as input*
    - Read and tail on [`Plain text`](docs/collectors/collector_tail.md) files
    - Ingest [`PCAP`](docs/collectors/collector_fileingestor.md) or [`DNSTap`](docs/collectors/collector_fileingestor.md) files by watching a directory
  - *Local storage of your DNS logs in text or binary formats*
    - [`Stdout`](docs/loggers/logger_stdout.md) console in text or binary output
    - [`File`](docs/loggers/logger_file.md) with automatic rotation and compression
  - *Provide metrics and API*
    - [`Prometheus`](docs/loggers/logger_prometheus.md) exporter
    - [`OpenTelemetry`](docs/loggers/logger_opentelemetry.md) tracing dns
    - [`Statsd`](docs/loggers/logger_statsd.md) support
    - [`REST API`](docs/loggers/logger_restapi.md) with [swagger](https://generator.swagger.io/?url=https://raw.githubusercontent.com/dmachard/DNS-collector/main/docs/swagger.yml) to search DNS domains
  - *Send to remote host with generic transport protocol*
    - Raw [`TCP`](docs/loggers/logger_tcp.md) client
    - [`Syslog`](docs/loggers/logger_syslog.md) with TLS support
    - [`DNSTap`](docs/loggers/logger_dnstap.md) protobuf client
  - *Send to various sinks*
    - [`Fluentd`](docs/loggers/logger_fluentd.md)
    - [`InfluxDB`](docs/loggers/logger_influxdb.md)
    - [`Loki`](docs/loggers/logger_loki.md) client
    - [`ElasticSearch`](docs/loggers/logger_elasticsearch.md)
    - [`Scalyr`](docs/loggers/logger_scalyr.md)
    - [`Redis`](docs/loggers/logger_redis.md) publisher
    - [`Kafka`](docs/loggers/logger_kafka.md) producer
    - [`ClickHouse`](docs/loggers/logger_clickhouse.md) client
  - *Send to security tools*
    - [`Falco`](docs/loggers/logger_falco.md)

- **[Transformers](./docs/transformers.md)**

  - Detect [Newly Observed Domains](docs/transformers/transform_newdomaintracker.md)
  - [Rewrite](docs/transformers/transform_rewrite.md) DNS messages 
  - Custom JSON output [Relabeling](docs/transformers/transform_relabeling.md)
  - Add additionnal [Tags](docs/transformers/transform_atags.md) in DNS messages
  - Traffic [Filtering](docs/transformers/transform_trafficfiltering.md) 
  - Merge similar DNS logs with the [Reducer](docs/transformers/transform_trafficreducer.md)
  - Latency [Computing](docs/transformers/transform_latency.md)
  - Apply [User Privacy](docs/transformers/transform_userprivacy.md)
  - [Normalize](docs/transformers/transform_normalize.md) DNS messages
  - Add [Geographical](docs/transformers/transform_geoip.md) metadata
  - Various data [Extractor](docs/transformers/transform_dataextractor.md)
  - Suspicious traffic [Detector](docs/transformers/transform_suspiciousdetector.md) 
  - Help to train your machine learning models with the [Prediction](docs/transformers/transform_trafficprediction.md) transformer
  - [Reordering](docs/transformers/transform_reordering.md) DNS messages based on timestamps

## 🚀 Get Started

Download the latest [`release`](https://github.com/dmachard/DNS-collector/releases) binary and start the DNS-collector with the provided configuration file. The default configuration listens on `tcp/6000` for a DNSTap stream and DNS logs are printed on standard output.

```bash
./dnscollector -config config.yml
```

![run](docs/_images/terminal.gif)

If you prefer run it from docker, follow this [guide](./docs/docker.md).

## ⚙️ Configuration

The configuration of DNS-collector is done through a file named [`config.yml`](config.yml). 
When the DNS-collector starts, it will look for the config.yml from the current working directory.
A typical [configuration in pipeline](./running_mode.md) mode includes one or more collectors to receive DNS traffic and several loggers to process the incoming data. 

To get started quickly, you can use this default [`config.yml`](config.yml). You can also see  the `_examples` folder from documentation witch contains a number of [various configurations](./docs/examples.md) to get you started with the DNS-collector in different ways.

For advanced settings, see the [advanced configuration guide](./docs/advanced_config.md).

Additionally, the [`_integration`](./docs/_integration) folder contains preconfigured files and `docker compose` examples
for integrating DNS-collector with popular tools:

- [Fluentd](./docs/_integration/fluentd/README.md)
- [Elasticsearch](./docs/_integration/elasticsearch/README.md)
- [Kafka](./docs/_integration/kafka/README.md)
- [InfluxDB](./docs/_integration/influxdb/README.md)
- [Prometheus](./docs/_integration/prometheus/README.md)
- [Loki](./docs/_integration/loki/README.md)

## 📊 DNS Telemetry

`DNS-collector` provides telemetry capabilities with the Prometheus logger, 
you can easily monitor key performance indicators and detect anomalies in real-time.

![dashboard](docs/_images/dashboard_prometheus.png)

## ⚡ Performance

Tuning may be necessary to deal with a large traffic loads.
Please refer to the [performance tuning](./docs/performance.md) guide if needed.

Performance metrics are available to evaluate the efficiency of your pipelines. These metrics allow you to track:
- The number of incoming and outgoing packets processed by each worker
- The number of packets matching the policies applied (forwarded, dropped)
- The number of "discarded" packets
- Memory consumption
- CPU consumption

A [build-in](./docs/dashboards/grafana_exporter.json) dashboard is available for monitoring these metrics.

![dashboard](docs/_images/dashboard_global.png)

## ❤️ Contributing

See the [development guide](./docs/development.md) for more information on how to build it yourself.

## 🧰 More DNS tools ?

| | |
|:--:|------------|
| <a href="https://github.com/dmachard/DNS-collector" target="_blank"><img src="https://github.com/dmachard/DNS-collector/blob/main/docs/dns-collector_logo.png?raw=true" alt="DNS-collector" width="200"/></a> | Ingesting, pipelining, and enhancing your DNS logs with usage indicators, security analysis, and additional metadata. |
| <a href="https://github.com/dmachard/DNS-tester" target="_blank"><img src="https://github.com/dmachard/DNS-tester/blob/main/docs/logo-dns-tester.png?raw=true" alt="DNS-collector" width="200"/></a> | Monitoring DNS server availability and comparing response times across multiple DNS providers. |