# Kafka "work stealing" example

In a system using kafka, consumers are tied to a specific partition. When a key
does not exist for a message, the partition is chosen based on a round robin
approach. The processing required for each message might not be equal, so a
consumer could get unlucky where the round robin distribution just so happens
to give that consumer more higher-processing-requirement messages than others.
If that's the case, the consumer can choose to not do the work and republish
the message in the hopes that another consumer has the capacity to handle it so
that the message can be processed sooner.

This repository contains example code that showcases one approach to do this.

## Prerequisites

- docker
- docker-compose
- Go (minimum compiler version unknown - tested with 1.18)

## How to run the example

```
./run.sh
```
