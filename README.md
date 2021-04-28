# avalanche

Avalanche serves a text-based [Prometheus metrics](https://prometheus.io/docs/instrumenting/exposition_formats/) endpoint for load testing [Prometheus](https://prometheus.io/) and possibly other [OpenMetrics](https://github.com/OpenObservability/OpenMetrics) consumers.

- Metric names and unique series change over time to simulate series churn.
- Support for generating Topology with multiple related entities [see example avalanche.yaml for more details]
- Topology can also change over time to simulate churn

## configuration flags 
```bash 
avalanche --help
```

## Endpoints

Two endpoints are available :
* `/metrics` - metrics endpoint
* `/health` - healthcheck endpoint

## build and run go binary
```bash
go get github.com/open-fresh/avalanche/cmd/...
avalanche --help
```