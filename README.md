# content-collection-unfolder

[![Circle CI](https://circleci.com/gh/Financial-Times/content-collection-unfolder/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/content-collection-unfolder/tree/master)[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/content-collection-unfolder)](https://goreportcard.com/report/github.com/Financial-Times/content-collection-unfolder) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/content-collection-unfolder/badge.svg)](https://coveralls.io/github/Financial-Times/content-collection-unfolder)

## Introduction

UPP Service that finds added/deleted collection members and lead article through relations-api.
Then it forwards mapped content collections to the content-collection-rw-neo4j to be written in Neo4j database.
If a 200 answer is received from the writer, 
it retrieves the added/deleted contents and the lead article in the collection from document-store-api 
and then places them in Kafka on the Post Publication topic so that notifications will be created for them.

Dependencies are:
1. relations-api
2. content-collection-rw-neo4j
3. document-store-api
4. kafka

## Installation
      
Download the source code, dependencies and test dependencies:

        go get -u github.com/kardianos/govendor
        go get -u github.com/Financial-Times/content-collection-unfolder
        cd $GOPATH/src/github.com/Financial-Times/content-collection-unfolder
        govendor sync
        go build .

## Running locally

1. Run the tests and install the binary:

        govendor sync
        govendor test -v -race
        go install

2. Run the binary (using the `help` flag to see the available optional arguments):

        $GOPATH/bin/content-collection-unfolder [--help]

Options:

        --app-system-code="content-collection-unfolder"                                                         System Code of the application ($APP_SYSTEM_CODE)
        --app-name="Content Collection Unfolder"                                                                Application name ($APP_NAME)
        --app-port="8080"                                                                                       Port to listen on ($APP_PORT)
        --unfolding-whitelist=["content-package"]                                                               Collection types for which the unfolding process should be performed ($UNFOLDING_WHITELIST)
        --writer-uri="http://localhost:8080/__content-collection-rw-neo4j/content-collection/"                  URI of the Writer ($WRITER_URI)
        --writer-health-uri="http://localhost:8080/__content-collection-rw-neo4j/__health"                      URI of the Writer health endpoint ($WRITER_HEALTH_URI)
        --content-resolver-uri="http://localhost:8080/__document-store-api/content/"                            URI of the Content Resolver ($CONTENT_RESOLVER_URI)
        --content-resolver-health-uri="http://localhost:8080/__document-store-api/__health"                     URI of the Content Resolver health endpoint ($CONTENT_RESOLVER_HEALTH_URI)
        --relations-resolver-uri="http://localhost:8080/__relations-api/contentcollection/{uuid}/relations" \   URI of the Relations Resolver ($RELATIONS_RESOLVER_URI)
        --relations-resolver-health-uri="http://localhost:8080/__relations-api/__health" \                      URI of the Relations Resolver health endpoint ($RELATIONS_RESOLVER_HEALTH_URI)
        --kafka-write-topic="PostPublicationEvents"                                                             The topic to write the messages to ($Q_WRITE_TOPIC)
        --kafka-proxy-address="http://localhost:8080"                                                           Addresses of the kafka proxy ($Q_ADDR)
        --kafka-proxy-hostname="kafka"                                                                          The hostname of the kafka proxy (for hostname based routing) ($Q_HOSTNAME)
        --kafka-authorization=""                                                                                Authorization for kafka ($Q_AUTHORIZATION)
        
        
3. Test:

Create a file with the following content collection contents:

      {
        "uuid": "45163790-eec9-11e6-abbc-ee7d9c5b3b90",
        "items": [
          {
            "uuid": "d4986a58-de3b-11e6-86ac-f253db7791c6"
          },
          {
            "uuid": "d9b4c4c6-dcc6-11e6-86ac-f253db7791c6"
          }
        ],
        "publishReference": "tdi23377744",
        "lastModified": "2017-01-31T15:33:21.687Z"
      }


Assuming that the file you just is `cc.json`, you can run the following curl command to test the unfolder

      curl -X PUT --data "@cc.json"  localhost:8080/content-collection/content-package/45163790-eec9-11e6-abbc-ee7d9c5b3b90

If you've setup everything correctly, you should receive a `200` response. You can also watch the app logs for errors as the
unfolder will return a `200` even if kafka related processing fails.

## Build and deployment
_How can I build and deploy it (lots of this will be links out as the steps will be common)_

* CI provided by CircleCI: [content-collection-unfolder](https://circleci.com/gh/Financial-Times/content-collection-unfolder)
* Built by Docker Hub on merge to master or on any pushed tag: [coco/content-collection-unfolder](https://hub.docker.com/r/coco/content-collection-unfolder/)

## Service endpoints

### PUT

Using curl:

    curl -X PUT --data "@cc.json"  localhost:8080/content-collection/content-package/45163790-eec9-11e6-abbc-ee7d9c5b3b90

The expected response is a simple `200` with no response body. In case a error takes place, a json response body will be provided,
similar to the following example:

    {"msg\":"Something bad happened"}
    
As a rule of thumb, the unfolder will return the exact response status code and body received from the **content-collection-neo4j-rw** app in
case a non `200` response is received.

The flow of the unfolder is the following:

1. the PUT request is received and a call is made to **relations-api** /contentcollection/{uuid}/relations to get added/deleted members and lead article
1. the PUT request is forwarded to the **content-collection-neo4j-rw** to write data in Neo4j
2. the response from the RW app is evaluated, if it is not a `200` response, the writer response is sent to the unfolder client
3. in case the collection type is not one that needs unfolding (at the moment only `content-package` is unfolded), a `200` response is returned
4. the content of UUIDs of added/deleted content and lead article are resolved using the **document-store-api**
5. for each piece of content retrieved from the DSAPI, a new message is created and placed on the configured **kafka** topic

## Healthchecks
Admin endpoints are:

`/__gtg`

`/__build-info`

`/__ping`

`/__health`

There are following checks are performed when the `/__health` is called:
1. **relations-api** connectivity check
2. **content-collection-neo4j-rw** connectivity check
3. **document-store-api** connectivity check
4. **kafka** connectivity check

The `/__gtg` endpoint will return a `200` in case all above health checks are successful. 

test
