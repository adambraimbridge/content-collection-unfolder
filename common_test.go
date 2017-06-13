package main

import (
	"bytes"
	"github.com/Financial-Times/transactionid-utils-go"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

const (
	whitelistedCollection = "content-package"
	inputFile             = "content-collection.json"
	collectionUuid        = "45163790-eec9-11e6-abbc-ee7d9c5b3b90"
	firstItemUuid         = "d4986a58-de3b-11e6-86ac-f253db7791c6"
	secondItemUuid        = "d9b4c4c6-dcc6-11e6-86ac-f253db7791c6"
	lastModified          = "2017-01-31T15:33:21.687Z"
)

func buildRequest(t *testing.T, serverUrl string, collection string, uuid string, body []byte, tid string) *http.Request {
	req, err := http.NewRequest(http.MethodPut, serverUrl+buildPath(t, collection, uuid), bytes.NewBuffer(body))
	assert.NoError(t, err)

	req.Header.Add(transactionidutils.TransactionIDHeader, tid)

	return req
}

func buildPath(t *testing.T, collectionType string, uuid string) string {
	pathWithCollection := strings.Replace(unfolderPath, "{collectionType}", collectionType, 1)
	assert.NotEqual(t, unfolderPath, pathWithCollection)

	pathWithCollectionAndUuid := strings.Replace(pathWithCollection, "{uuid}", uuid, 1)
	assert.NotEqual(t, pathWithCollection, pathWithCollectionAndUuid)

	return pathWithCollectionAndUuid
}

func readTestFile(t *testing.T, fileName string) []byte {
	file, err := os.Open("test-resources/" + fileName)
	assert.NoError(t, err)

	defer file.Close()
	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	assert.NoError(t, err)

	return buf.Bytes()
}
