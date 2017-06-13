package main

import (
	"encoding/json"
	"errors"
	fw "github.com/Financial-Times/content-collection-unfolder/forwarder"
	prod "github.com/Financial-Times/content-collection-unfolder/producer"
	res "github.com/Financial-Times/content-collection-unfolder/resolver"
	health "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/transactionid-utils-go"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	writerPath                = "/content-collection/{collectionType}/{uuid}"
	writerHealthPath          = "/__health"
	contentResolverPath       = "/content"
	contentResolverHealthPath = "/__health"
	tid                       = "tid_test123456"
)

func TestAllHealthChecksBad(t *testing.T) {
	writerServer := startWriterServer(t, notFoundHandler)
	defer writerServer.Close()

	contentResolverServer := startContentResolverServer(t, notFoundHandler)
	defer contentResolverServer.Close()

	messageProducer := &testProducer{t, false, []string{}}

	routing := startRouting(writerServer, contentResolverServer, messageProducer)

	unfolderServer := httptest.NewServer(routing.router)
	defer unfolderServer.Close()

	req, err := http.NewRequest(http.MethodGet, unfolderServer.URL+healthPath, nil)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, http.StatusOK)

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	healthResult := health.HealthResult{}
	json.Unmarshal(body, &healthResult)

	assert.False(t, healthResult.Ok)
	assert.Equal(t, 3, len(healthResult.Checks))

	for _, checkResult := range healthResult.Checks {
		assert.False(t, checkResult.Ok)
	}
}

func TestAllHealthChecksGood(t *testing.T) {
	writerServer := startWriterServer(t, okHandler)
	defer writerServer.Close()

	contentResolverServer := startContentResolverServer(t, okHandler)
	defer contentResolverServer.Close()

	messageProducer := &testProducer{t, true, []string{}}

	routing := startRouting(writerServer, contentResolverServer, messageProducer)

	unfolderServer := httptest.NewServer(routing.router)
	defer unfolderServer.Close()

	req, err := http.NewRequest(http.MethodGet, unfolderServer.URL+healthPath, nil)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, http.StatusOK)

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	healthResult := health.HealthResult{}
	json.Unmarshal(body, &healthResult)

	assert.True(t, healthResult.Ok)
	assert.Equal(t, 3, len(healthResult.Checks))

	for _, checkResult := range healthResult.Checks {
		assert.True(t, checkResult.Ok)
	}
}

func TestEndToEndFlow(t *testing.T) {
	writerServer := startWriterServer(t, okHandler)
	defer writerServer.Close()

	contentResolverServer := startContentResolverServer(t, okHandler)
	defer contentResolverServer.Close()

	messageProducer := &testProducer{t, true, []string{}}

	routing := startRouting(writerServer, contentResolverServer, messageProducer)

	unfolderServer := httptest.NewServer(routing.router)
	defer unfolderServer.Close()

	req := buildRequest(t, unfolderServer.URL, whitelistedCollection, collectionUuid, readTestFile(t, inputFile), tid)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, 2, len(messageProducer.received))
	allMessages := strings.Join(messageProducer.received, "")
	assert.Equal(t, 2, strings.Count(allMessages, firstItemUuid))
	assert.Equal(t, 2, strings.Count(allMessages, secondItemUuid))
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func startWriterServer(t *testing.T, healthHandler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc(writerPath, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)

		assert.Equal(t, tid, transactionidutils.GetTransactionIDFromRequest(r))
		assert.Equal(t, readTestFile(t, inputFile), body)

		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodPut)

	router.HandleFunc(writerHealthPath, healthHandler).Methods(http.MethodGet)

	return httptest.NewServer(router)
}

func startContentResolverServer(t *testing.T, healthHandler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc(contentResolverPath, func(w http.ResponseWriter, r *http.Request) {
		uuidArr := r.URL.Query()["uuid"]

		assert.Equal(t, tid, transactionidutils.GetTransactionIDFromRequest(r))
		assert.Contains(t, uuidArr, firstItemUuid)
		assert.Contains(t, uuidArr, secondItemUuid)

		contentArr := []map[string]string{}
		for _, uuid := range uuidArr {
			contentArr = append(contentArr, map[string]string{"uuid": uuid})
		}

		body, err := json.Marshal(contentArr)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(body)

		assert.NoError(t, err)
	}).Methods(http.MethodGet)

	router.HandleFunc(contentResolverHealthPath, healthHandler).Methods(http.MethodGet)

	return httptest.NewServer(router)
}

type testProducer struct {
	t        *testing.T
	healthy  bool
	received []string
}

func (tp *testProducer) SendMessage(key string, msg producer.Message) error {
	assert.Equal(tp.t, tid, msg.Headers["X-Request-Id"])
	assert.Equal(tp.t, lastModified, msg.Headers["Message-Timestamp"])

	tp.received = append(tp.received, msg.Body)

	return nil
}

func (tp *testProducer) ConnectivityCheck() (string, error) {
	if tp.healthy {
		return "Ok", nil
	} else {
		return "", errors.New("Not healthy")
	}
}

func startRouting(
	writerServer *httptest.Server,
	contentResolverServer *httptest.Server,
	messageProducer *testProducer) *routing {

	client := setupHttpClient()
	hc := &healthConfig{
		appDesc:                  appDescription,
		port:                     "8080",
		appSystemCode:            "content-collection-unfolder",
		appName:                  "Content Collection Unfolder",
		writerHealthURI:          writerServer.URL + writerHealthPath,
		contentResolverHealthURI: contentResolverServer.URL + contentResolverHealthPath,
		producer:                 messageProducer,
		client:                   client,
	}
	routing := newRouting(
		newUnfolder(
			fw.NewForwarder(client, writerServer.URL+strings.Split(writerPath, "/{")[0]),
			res.NewUuidResolver(),
			res.NewContentResolver(client, contentResolverServer.URL+contentResolverPath),
			prod.NewContentProducer(messageProducer),
			[]string{whitelistedCollection},
		),
		newHealthService(hc),
	)

	return routing
}
