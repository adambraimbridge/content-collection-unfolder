package main

import (
	"github.com/jawher/mow.cli"
)

type serviceConfig struct {
	appSystemCode            *string
	appName                  *string
	appPort                  *string
	writerURI                *string
	writerHealthURI          *string
	contentResolverURI       *string
	contentResolverHealthURI *string
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

	return &serviceConfig{
		appSystemCode:            appSystemCode,
		appName:                  appName,
		appPort:                  appPort,
		writerURI:                writerURI,
		writerHealthURI:          writerHealthURI,
		contentResolverURI:       contentResolverURI,
		contentResolverHealthURI: contentResolverHealthURI,
	}
}
