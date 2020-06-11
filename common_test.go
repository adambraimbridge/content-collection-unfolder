package main

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	transactionidutils "github.com/Financial-Times/transactionid-utils-go"
	"github.com/stretchr/testify/assert"
)

const (
	whitelistedCollection   = "content-package"
	inputFile               = "content-collection.json"
	collectionUUID          = "45163790-eec9-11e6-abbc-ee7d9c5b3b90"
	leadArticleUUID         = "ddda0e1c-a9b2-11e7-8e2d-6debe43a48b4"
	firstExistingItemUUID   = "aaaac4c6-dcc6-11e6-86ac-f253db7791c6"
	secondExistingItemUUUID = "bbbbc4c6-dcc6-11e6-86ac-f253db7791c6"
	deletedItemUUID         = "d9b4c4c6-dcc6-11e6-86ac-f253db7791c6"
	addedItemUUUID          = "d4986a58-de3b-11e6-86ac-f253db7791c6"
	lastModified            = "2017-01-31T15:33:21.687Z"
)

func buildRequest(t *testing.T, serverURL string, collection string, uuid string, body []byte, tid string) *http.Request {
	req, err := http.NewRequest(http.MethodPut, serverURL+buildPath(t, collection, uuid), bytes.NewBuffer(body))
	assert.NoError(t, err)

	req.Header.Add(transactionidutils.TransactionIDHeader, tid)

	return req
}

func buildPath(t *testing.T, collectionType string, uuid string) string {
	pathWithCollection := strings.Replace(unfolderPath, "{collectionType}", collectionType, 1)
	assert.NotEqual(t, unfolderPath, pathWithCollection)

	pathWithCollectionAndUUID := strings.Replace(pathWithCollection, "{uuid}", uuid, 1)
	assert.NotEqual(t, pathWithCollection, pathWithCollectionAndUUID)

	return pathWithCollectionAndUUID
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
