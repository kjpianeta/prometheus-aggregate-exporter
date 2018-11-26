#!/usr/bin/env bash
set -ex
TEST_SERVICE_NAME="rpmtest_prometheus_aggregate_exporter"
RPM_FILE=$1

function cleanup(){
    docker-compose exec ${TEST_SERVICE_NAME} service prometheus-aggregate-exporter status
    echo "test/run.sh: Cleaning up..."
    unset inspec
    docker-compose down --remove-orphans -v --rmi all || true
}
trap cleanup EXIT

: "${RPM_FILE?Not set. Need copy RPM file to test folder}"
cp -p "${RPM_FILE}" .

export RPM_FILE_NAME=$(find . -name "*.rpm" -type f)
echo "Building container with RPM ${RPM_FILE_NAME}..."

docker-compose up -d
export SERVICE_NAME=$(docker-compose ps --services | grep "prometheus_aggregate_exporter")
docker-compose exec ${SERVICE_NAME} service prometheus-aggregate-exporter start

echo "Building inspec environment..."
inspec --version || docker pull chef/inspec && function inspec { docker run -it --rm -v $(pwd):/share -v /var/run/docker.sock:/var/run/docker.sock chef/inspec "$@"; }

echo "Running inspec tests"
CONTAINER_UNDER_TEST_ID=$(docker-compose ps -q)
for TEST_SPEC in `find . -name "*_spec.rb"`
do
    inspec exec "${TEST_SPEC}" -t docker://${CONTAINER_UNDER_TEST_ID}
done

