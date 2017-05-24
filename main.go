package main

import (
	fw "github.com/Financial-Times/content-collection-unfolder/forwarder"
	log "github.com/Sirupsen/logrus"
	"github.com/jawher/mow.cli"
	"net"
	"net/http"
	"os"
	"time"
)

const appDescription = "UPP Service that forwards mapped content collections to the content-collection-rw-neo4j. If a 200 answer is received from the writer, it retrieves the elements in the collection from the document-store-api and places them in Kafka on the Post Publication topic so that notifications will be created for them."

func main() {
	app := cli.App("content-collection-unfolder", appDescription)

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

	port := app.String(cli.StringOpt{
		Name:   "port",
		Value:  "8080",
		Desc:   "Port to listen on",
		EnvVar: "APP_PORT",
	})

	writerUri := app.String(cli.StringOpt{
		Name:   "writer-uri",
		Value:  "http://localhost:8080/__content-collection-rw-neo4j/content-collection/",
		Desc:   "URI of the writer",
		EnvVar: "WRITER_URI",
	})

	writerHealthUri := app.String(cli.StringOpt{
		Name:   "writer-health-uri",
		Value:  "http://localhost:8080/__content-collection-rw-neo4j/__health",
		Desc:   "URI of the writer health endpoint",
		EnvVar: "WRITER_HEALTH_URI",
	})

	log.SetLevel(log.InfoLevel)
	log.Infof("[Startup] content-collection-unfolder is starting ")

	app.Action = func() {
		log.Infof("System code: %s, App Name: %s, Port: %s", *appSystemCode, *appName, *port)

		client := setupHttpClient()
		config := &healthConfig{
			client:          client,
			port:            *port,
			appSystemCode:   *appSystemCode,
			appName:         *appName,
			appDesc:         appDescription,
			writerHealthUri: *writerHealthUri,
		}

		newRouting(
			newUnfolder(
				fw.NewForwarder(client, *writerUri),
			),
			newHealthService(config),
		).listenAndServe(*port)
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Errorf("App could not start, error=[%v]\n", err)
		return
	}
}

func setupHttpClient() *http.Client {
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
