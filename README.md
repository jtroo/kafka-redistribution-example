# Kafka work re-distributing example

In a system using kafka, consumers are tied to specific partitions. When a key
does not exist for a message, the partition is chosen based on a round robin
approach. The processing required for each message might not be equal, so a
consumer could get unlucky where the round robin distribution just so happens
to give that consumer more higher-processing-requirement messages than others.
The situation described only makes sense if processing times for messages are a
significant amount of time, which is not the case for many kafka use cases.

This repository contains example code that showcases one approach to
re-distribute work.

In this example, the consumer republishes the message in the hopes that another
consumer has the capacity to handle it so that the message can be processed
sooner.

Another approach might be to do manual commit management, but this could have
worse time-to-complete for the overloading jobs since the broker would need to
wait for the timeout the commit for messages before sending them to another
consumer.

## Prerequisites

- docker
- docker-compose
- Go (minimum compiler version unknown - tested with 1.18)

## How to run the example

    ./run.sh

The code will be working as expected if you see a printed message once in a while saying:

    consumer <id> got unlucky, re-producing work

## Example run

    2022/03/31 22:59:52 consumer 4 got work: 2173ms
    2022/03/31 22:59:52 consumer 2 got work: 1235ms
    2022/03/31 22:59:53 consumer 4 got work: 1443ms
    2022/03/31 22:59:53 consumer 4 got unlucky, re-producing work: 1443ms
    2022/03/31 22:59:53 consumer 1 got work: 1443ms
    2022/03/31 22:59:53 consumer 1 got work: 1421ms
    2022/03/31 22:59:53 consumer 2 got work: 1390ms
    2022/03/31 22:59:53 consumer 1 got work: 1574ms
    2022/03/31 22:59:53 consumer 1 got work: 1869ms
    2022/03/31 22:59:54 consumer 1 got unlucky, re-producing work: 1869ms
    2022/03/31 22:59:54 consumer 1 got work: 1869ms
    2022/03/31 22:59:54 consumer 4 got work: 1070ms
    2022/03/31 22:59:54 consumer 3 got work: 1336ms
