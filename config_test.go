package main

import (
	"testing"

	cli "github.com/jawher/mow.cli"
	"github.com/stretchr/testify/assert"
)

const (
	emptyString           = ""
	requestTimeoutSeconds = 2
)

func TestDefaults(t *testing.T) {
	app := cli.App("content-collection-unfolder", serviceDescription)

	sc := createServiceConfiguration(app)

	app.Action = func() {
		configMap := sc.toMap()
		assert.NotEmpty(t, configMap["appSystemCode"])
		assert.NotEmpty(t, configMap["appName"])
		assert.NotEmpty(t, configMap["appPort"])
		assert.NotEmpty(t, configMap["unfoldingWhitelist"])
		assert.NotEmpty(t, configMap["writerURI"])
		assert.NotEmpty(t, configMap["writerHealthURI"])
		assert.NotEmpty(t, configMap["contentResolverURI"])
		assert.NotEmpty(t, configMap["contentResolverHealthURI"])
		assert.NotEmpty(t, configMap["relationsResolverURI"])
		assert.NotEmpty(t, configMap["relationsResolverHealthURI"])
		assert.NotEmpty(t, configMap["writeTopic"])
		assert.NotEmpty(t, configMap["kafkaAddr"])
		assert.NotEmpty(t, configMap["kafkaHostname"])
		assert.Equal(t, emptyString, configMap["kafkaAuth"])
		assert.Equal(t, requestTimeoutSeconds, configMap["requestTimeout"])
	}

	app.Run([]string{"content-collection-unfolder"})
}
