package main

import (
	"github.com/jawher/mow.cli"
)

type serviceConfig struct {
	appSystemCode            *string
	appName                  *string
	appPort                  *string
	unfoldingWhitelist       *[]string
	writerURI                *string
	writerHealthURI          *string
	contentResolverURI       *string
	contentResolverHealthURI *string
	queueAddress             *string
	writeTopic               *string
	writeQueue               *string
	authorization            *string
}

func createServiceConfiguration(app *cli.Cli) *serviceConfig {
	appSystemCode := app.String(cli.StringOpt{
		Name:   "app-system-code",
		Value:  "content-collection-unfolder",
		Desc:   "System Code of the application",
		EnvVar: "APP_SYSTEM_CODE",
	})

	appName := app.String(cli.StringOpt{
		Name:   "app-name",
		Value:  "Content Collection Unfolder",
		Desc:   "Application name",
		EnvVar: "APP_NAME",
	})

	appPort := app.String(cli.StringOpt{
		Name:   "app-port",
		Value:  "8080",
		Desc:   "Port to listen on",
		EnvVar: "APP_PORT",
	})

	unfoldingWhitelist := app.Strings(cli.StringsOpt{
		Name:   "unfolding-whitelist",
		Value:  []string{"content-package"},
		Desc:   "Collection types for which the unfolding process should be performed",
		EnvVar: "UNFOLDING_WHITELIST",
	})

	writerURI := app.String(cli.StringOpt{
		Name:   "writer-uri",
		Value:  "http://localhost:8080/__content-collection-rw-neo4j/content-collection/",
		Desc:   "URI of the Writer",
		EnvVar: "WRITER_URI",
	})

	writerHealthURI := app.String(cli.StringOpt{
		Name:   "writer-health-uri",
		Value:  "http://localhost:8080/__content-collection-rw-neo4j/__health",
		Desc:   "URI of the Writer health endpoint",
		EnvVar: "WRITER_HEALTH_URI",
	})

	contentResolverURI := app.String(cli.StringOpt{
		Name:   "content-resolver-uri",
		Value:  "http://localhost:8080/__document-store-api/content/",
		Desc:   "URI of the Content Resolver",
		EnvVar: "CONTENT_RESOLVER_URI",
	})

	contentResolverHealthURI := app.String(cli.StringOpt{
		Name:   "content-resolver-health-uri",
		Value:  "http://localhost:8080/__document-store-api/__health",
		Desc:   "URI of the Content Resolver health endpoint",
		EnvVar: "CONTENT_RESOLVER_HEALTH_URI",
	})

	queueAddress := app.String(cli.StringOpt{
		Name:   "queue-address",
		Value:  "http://localhost:8080",
		Desc:   "Addresses to connect to the queue (hostnames).",
		EnvVar: "Q_ADDR",
	})

	writeTopic := app.String(cli.StringOpt{
		Name:   "write-topic",
		Value:  "PostPublicationEvents",
		Desc:   "The topic to write the meassages to.",
		EnvVar: "Q_WRITE_TOPIC",
	})

	writeQueue := app.String(cli.StringOpt{
		Name:   "write-queue",
		Value:  "kafka",
		Desc:   "The queue to write the meassages to.",
		EnvVar: "Q_WRITE_QUEUE",
	})

	authorization := app.String(cli.StringOpt{
		Name:   "authorization",
		Desc:   "Authorization key to access the queue.",
		EnvVar: "Q_AUTHORIZATION",
	})

	return &serviceConfig{
		appSystemCode:            appSystemCode,
		appName:                  appName,
		appPort:                  appPort,
		unfoldingWhitelist:       unfoldingWhitelist,
		writerURI:                writerURI,
		writerHealthURI:          writerHealthURI,
		contentResolverURI:       contentResolverURI,
		contentResolverHealthURI: contentResolverHealthURI,
		queueAddress:             queueAddress,
		writeTopic:               writeTopic,
		writeQueue:               writeQueue,
		authorization:            authorization,
	}
}

func (sc *serviceConfig) toMap() map[string]interface{} {
	return map[string]interface{}{
		"appSystemCode":            *sc.appSystemCode,
		"appName":                  *sc.appName,
		"appPort":                  *sc.appPort,
		"unfoldingWhitelist":       *sc.unfoldingWhitelist,
		"writerURI":                *sc.writerURI,
		"writerHealthURI":          *sc.writerHealthURI,
		"contentResolverURI":       *sc.contentResolverURI,
		"contentResolverHealthURI": *sc.contentResolverHealthURI,
		"queueAddress":             *sc.queueAddress,
		"writeTopic":               *sc.writeTopic,
		"writeQueue":               *sc.writeQueue,
		"authorization":            *sc.authorization,
	}
}
