#!/bin/bash

set -e

if [ -z "$(docker ps | grep work-redistribution-example_kafka)" ]; then
  echo "Kafka is not running, starting up zk+kafka"
  docker-compose up -d
  echo
  echo "Kafka sometimes does not start up properly when started soon after ZooKeeper."
  echo "Sleeping for 30s then retrying startup"
  sleep 30
  echo
  docker-compose up -d
  echo
  echo "Sleeping for 10s to wait for Kafka broker to start up"
  sleep 10
fi

docker run --network=host confluentinc/cp-kafka:latest bash -c "
  kafka-configs --bootstrap-server localhost:9092 --alter --entity-type topics --add-config retention.ms=1000 --entity-name example-topic;
  sleep 2;
  kafka-topics --bootstrap-server localhost:9092 --delete --topic example-topic;
  kafka-topics --bootstrap-server localhost:9092 --create --partitions 4 --topic example-topic;
  kafka-configs --bootstrap-server localhost:9092 --alter --entity-type topics --add-config retention.ms=100000 --entity-name example-topic;
"

function cleanup {
  killall kafka_work_redistribute
}
trap cleanup EXIT

go run . consumer 1 &
go run . consumer 2 &
go run . consumer 3 &
go run . consumer 4 &
go run . producer
