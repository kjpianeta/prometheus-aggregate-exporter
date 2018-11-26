# prometheus-aggregate-exporter
Prometheus exporter is a Go lang application that aggregates other exporters based on the target URL.

### Requires
[goreleaser](https://goreleaser.com/introduction/)

### Create a release and publish it to Github
```
goreleaser release
```
### Engineering RPM build
```
goreleaser release --rm-dist --snapshot
```