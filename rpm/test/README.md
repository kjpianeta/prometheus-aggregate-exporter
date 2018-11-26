### Will start, build and test the container with the RPM(node_exporter-0.16.0-2.el6.x86_64.rpm) installed.
```
run.sh rpm/test/prometheus-aggregate-exporter-v0.0.0-next.el6.x86_64.rpm
```
### Creates inspec as a function
```
docker pull chef/inspec
function inspec { docker run -it --rm -v $(pwd):/share -v /var/run/docker.sock:/var/run/docker.sock chef/inspec "$@"; }
```
### Runs inspec test file against container test_rpmtest_1
```
inspec exec centos6/test/node_exporter_spec.rb -t docker://test_rpmtest_1
```

### Cleanup environment
```
docker-compose down --remove-orphans -v --rmi all
```