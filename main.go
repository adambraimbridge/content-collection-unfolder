package main

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Financial-Times/content-collection-unfolder/differ"
	fw "github.com/Financial-Times/content-collection-unfolder/forwarder"
	prod "github.com/Financial-Times/content-collection-unfolder/producer"
	"github.com/Financial-Times/content-collection-unfolder/relations"
	res "github.com/Financial-Times/content-collection-unfolder/resolver"
	logger "github.com/Financial-Times/go-logger"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	cli "github.com/jawher/mow.cli"
)

const (
	serviceName        = "content-collection-unfolder"
	serviceDescription = "UPP Service that forwards mapped content collections to the content-collection-rw-neo4j. If a 200 answer is received from the writer, it retrieves the elements in the collection from the document-store-api and places them in Kafka on the Post Publication topic so that notifications will be created for them."
)

func main() {
	app := cli.App(serviceName, serviceDescription)
	sc := createServiceConfiguration(app)

	logger.InitDefaultLogger(serviceName)

	app.Action = func() {
		logger.Infof("[Startup] content-collection-unfolder is starting with service config %v", sc.toMap())

		client := setupHTTPClient()
		producer := setupMessageProducer(sc, client)

		unfolder := newUnfolder(
			res.NewUuidResolver(),
			relations.NewDefaultRelationsResolver(client, *sc.relationsResolverURI),
			differ.NewDefaultCollectionsDiffer(),
			fw.NewForwarder(client, *sc.writerURI),
			res.NewContentResolver(client, *sc.contentResolverURI, time.Duration(*sc.requestTimeout)*time.Second),
			prod.NewContentProducer(producer),
			*sc.unfoldingWhitelist,
		)
		healthService := newHealthService(&healthConfig{
			appDesc:                    serviceDescription,
			port:                       *sc.appPort,
			appSystemCode:              *sc.appSystemCode,
			appName:                    *sc.appName,
			writerHealthURI:            *sc.writerHealthURI,
			contentResolverHealthURI:   *sc.contentResolverHealthURI,
			relationsResolverHealthURI: *sc.relationsResolverHealthURI,
			producer:                   producer,
			client:                     client,
		})

		routing := newRouting(unfolder, healthService)
		routing.listenAndServe(*sc.appPort)
	}
	err := app.Run(os.Args)
	if err != nil {
		logger.Fatalf("App could not start, error=[%v]", err)
	}
}

func setupHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConnsPerHost:   20,
			TLSHandshakeTimeout:   3 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func setupMessageProducer(sc *serviceConfig, client *http.Client) producer.MessageProducer {
	config := producer.MessageProducerConfig{
		Addr:          *sc.kafkaAddr,
		Topic:         *sc.writeTopic,
		Queue:         *sc.kafkaHostname,
		Authorization: *sc.kafkaAuth,
	}

	return producer.NewMessageProducerWithHTTPClient(config, client)
}
