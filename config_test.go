package main

import (
	"github.com/jawher/mow.cli"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaults(t *testing.T) {
	app := cli.App("content-collection-unfolder", appDescription)

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
		assert.NotEmpty(t, configMap["writeTopic"])
		assert.NotEmpty(t, configMap["kafkaAddr"])
		assert.NotEmpty(t, configMap["kafkaHostname"])
		assert.Equal(t, "", configMap["kafkaAuth"])

	}

	app.Run([]string{"content-collection-unfolder"})
}
