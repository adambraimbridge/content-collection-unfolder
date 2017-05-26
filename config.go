package main

import (
	"github.com/jawher/mow.cli"
)

type serviceConfig struct {
	appSystemCode                string
	appName                      string
	appPort                      string
	contentResolverAppURI        string
	contentResolverAppHealthURI  string
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

	contentResolverAppURI := app.String(cli.StringOpt{
		Name:   "content-resolver-app-uri",
		Value:  "http://localhost:8080/__document-store-api/content/",
		Desc:   "Content Resolver APP URI",
		EnvVar: "CONTENT_RESOLVER_APP_URI",
	})

	contentResolverAppHealthURI := app.String(cli.StringOpt{
		Name:   "content-resolver-app-health-uri",
		Value:  "http://localhost:8080/__document-store-api/__health",
		Desc:   "URI of the Content Resolver APP health endpoint",
		EnvVar: "CONTENT_RESOLVER_APP_HEALTH_URI",
	})

	return &serviceConfig{appSystemCode: *appSystemCode, appName: *appName, appPort: *appPort, contentResolverAppURI: *contentResolverAppURI,
		contentResolverAppHealthURI: *contentResolverAppHealthURI}
}
