package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jawher/mow.cli"
	"os"
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

	log.SetLevel(log.InfoLevel)
	log.Infof("[Startup] content-collection-unfolder is starting ")

	app.Action = func() {
		log.Infof("System code: %s, App Name: %s, Port: %s", *appSystemCode, *appName, *port)

		config := &healthConfig{
			port:          *port,
			appSystemCode: *appSystemCode,
			appName:       *appName,
			appDesc:       appDescription,
		}

		newRouting(newUnfolder(), newHealthService(config)).listenAndServe(*port)
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Errorf("App could not start, error=[%v]\n", err)
		return
	}
}
