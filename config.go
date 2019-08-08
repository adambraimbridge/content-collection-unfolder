package main

import (
	"github.com/jawher/mow.cli"
)

type serviceConfig struct {
	appSystemCode              *string
	appName                    *string
	appPort                    *string
	unfoldingWhitelist         *[]string
	writerURI                  *string
	writerHealthURI            *string
	contentResolverURI         *string
	contentResolverHealthURI   *string
	relationsResolverURI       *string
	relationsResolverHealthURI *string
	writeTopic                 *string
	kafkaAddr                  *string
	kafkaHostname              *string
	kafkaAuth                  *string
	requestTimeout             *int
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

	relationsResolverURI := app.String(cli.StringOpt{
		Name:   "relations-resolver-uri",
		Value:  "http://localhost:8080/__relations-api/contentcollection/{uuid}/relations",
		Desc:   "URI of the relations resolver endpoint",
		EnvVar: "RELATIONS_RESOLVER_URI",
	})

	relationsResolverHealthURI := app.String(cli.StringOpt{
		Name:   "relations-resolver-health-uri",
		Value:  "http://localhost:8080/__relations-api/__health",
		Desc:   "URI of the relations resolver health endpoint",
		EnvVar: "RELATIONS_RESOLVER_HEALTH_URI",
	})

	writeTopic := app.String(cli.StringOpt{
		Name:   "kafka-write-topic",
		Value:  "PostPublicationEvents",
		Desc:   "The topic to write the messages to",
		EnvVar: "Q_WRITE_TOPIC",
	})

	kafkaAddr := app.String(cli.StringOpt{
		Name:   "kafka-proxy-address",
		Value:  "http://localhost:8080",
		Desc:   "Addresses of the kafka proxy",
		EnvVar: "Q_ADDR",
	})

	kafkaHostname := app.String(cli.StringOpt{
		Name:   "kafka-proxy-hostname",
		Value:  "kafka",
		Desc:   "The hostname of the kafka proxy (for hostname based routing)",
		EnvVar: "Q_HOSTNAME",
	})

	kafkaAuth := app.String(cli.StringOpt{
		Name:   "kafka-authorization",
		Desc:   "Authorization for kafka",
		EnvVar: "Q_AUTHORIZATION",
	})

	requestTimeout := app.Int(cli.IntOpt{
		Name:   "request_timeout",
		Value:  2,
		Desc:   "timeout per request for taking contents from document store",
		EnvVar: "REQUEST_TIMEOUT",
	})

	return &serviceConfig{
		appSystemCode:              appSystemCode,
		appName:                    appName,
		appPort:                    appPort,
		unfoldingWhitelist:         unfoldingWhitelist,
		writerURI:                  writerURI,
		writerHealthURI:            writerHealthURI,
		contentResolverURI:         contentResolverURI,
		contentResolverHealthURI:   contentResolverHealthURI,
		relationsResolverURI:       relationsResolverURI,
		relationsResolverHealthURI: relationsResolverHealthURI,
		writeTopic:                 writeTopic,
		kafkaAddr:                  kafkaAddr,
		kafkaHostname:              kafkaHostname,
		kafkaAuth:                  kafkaAuth,
		requestTimeout:             requestTimeout,
	}
}

func (sc *serviceConfig) toMap() map[string]interface{} {
	return map[string]interface{}{
		"appSystemCode":              *sc.appSystemCode,
		"appName":                    *sc.appName,
		"appPort":                    *sc.appPort,
		"unfoldingWhitelist":         *sc.unfoldingWhitelist,
		"writerURI":                  *sc.writerURI,
		"writerHealthURI":            *sc.writerHealthURI,
		"contentResolverURI":         *sc.contentResolverURI,
		"contentResolverHealthURI":   *sc.contentResolverHealthURI,
		"relationsResolverURI":       *sc.relationsResolverURI,
		"relationsResolverHealthURI": *sc.relationsResolverHealthURI,
		"writeTopic":                 *sc.writeTopic,
		"kafkaAddr":                  *sc.kafkaAddr,
		"kafkaHostname":              *sc.kafkaHostname,
		"kafkaAuth":                  *sc.kafkaAuth,
		"requestTimeout":             *sc.requestTimeout,
	}
}
