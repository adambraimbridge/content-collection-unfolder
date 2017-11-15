package relations

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const tid = "tid_qkeqptjwji"

var relationsResolver RelationsResolver
var relationsAPIMock *httptest.Server

func mockRelationsAPI(t *testing.T, appStatusCode int, outputFileName string) {
	file, err := os.Open("../test-resources/" + outputFileName)
	assert.NoError(t, err, "Opening file shouldn't throw error.")
	defer file.Close()
	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	assert.NoError(t, err, "Reading from file to []byte buffer shouldn't throw error.")
	mockRelationsAPIBytes(appStatusCode, buf.Bytes())
}

func mockRelationsAPIBytes(appStatus int, output []byte) {
	router := mux.NewRouter()
	var relationsResolverEndpointHandler http.HandlerFunc

	if appStatus == http.StatusOK {
		relationsResolverEndpointHandler = func(w http.ResponseWriter, r *http.Request) {
			w.Write(output)
		}
	} else {
		relationsResolverEndpointHandler = func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(appStatus)
		}
	}

	router.Path("/contentcollection/{uuid}/relations").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(relationsResolverEndpointHandler)})
	relationsAPIMock = httptest.NewServer(router)

	relationsResolver = NewDefaultRelationsResolver(http.DefaultClient, relationsAPIMock.URL+"/contentcollection/{uuid}/relations")
}

func TestRelationsResolver_Resolve_Ok(t *testing.T) {
	mockRelationsAPI(t, http.StatusOK, "relations-api-good-response.json")

	rel, err := relationsResolver.Resolve("3dd42508-a9ab-11e7-8e2d-6debe43a48b4", tid)
	if err != nil {
		assert.FailNow(t, "Normal resolve should not throw error.", err.Error())
	}

	assert.Equal(t, 2, len(rel.Contains), "There should be 2 uuids for the contains relation.")
	assert.Equal(t, "25f7b70e-a98a-11e7-8e2d-6debe43a48b4", rel.Contains[0], "Wrong first uuid for the contains relation.")
	assert.Equal(t, "267dfacf-62e5-3a5c-a645-4d9beb4de0be", rel.Contains[1], "Wrong second uuid for the contains relation.")
	assert.Equal(t, "ddda0e1c-a9b2-11e7-8e2d-6debe43a48b4", rel.ContainedIn, "The lead article's uuid is missing or wrong.")
}

func TestRelationsResolver_Resolve_NoContainsRelations(t *testing.T) {
	mockRelationsAPI(t, http.StatusOK, "relations-api-no-contains-response.json")

	rel, err := relationsResolver.Resolve("3dd42508-a9ab-11e7-8e2d-6debe43a48b4", tid)
	if err != nil {
		assert.FailNow(t, "Normal resolve should not throw error.", err.Error())
	}

	assert.Equal(t, 0, len(rel.Contains), "There should be no uuids for the contains relation.")
	assert.Equal(t, "ddda0e1c-a9b2-11e7-8e2d-6debe43a48b4", rel.ContainedIn, "The lead article's uuid is missing or wrong.")
}

func TestRelationsResolver_Resolve_RelationsApiNotWorking(t *testing.T) {
	mockRelationsAPI(t, http.StatusInternalServerError, "relations-api-good-response.json")

	_, err := relationsResolver.Resolve("3dd42508-a9ab-11e7-8e2d-6debe43a48b4", tid)
	if err == nil {
		assert.FailNow(t, "Relations API not working should throw error.", err.Error())
	}
}

func TestRelationsResolver_Resolve_RelationsNotFound(t *testing.T) {
	mockRelationsAPI(t, http.StatusNotFound, "relations-api-good-response.json")

	_, err := relationsResolver.Resolve("3dd42508-a9ab-11e7-8e2d-6debe43a48b4", tid)
	if err != nil {
		assert.FailNow(t, "Relations API no relations found shouldn't throw error.", err.Error())
	}
}

func TestRelationsResolver_Resolve_NotOkStatusThrowsError(t *testing.T) {
	mockRelationsAPI(t, http.StatusBadRequest, "relations-api-good-response.json")

	_, err := relationsResolver.Resolve("3dd42508-a9ab-11e7-8e2d-6debe43a48b4", tid)
	if err == nil {
		assert.FailNow(t, "Relations API not 200 OK status should throw error.", err.Error())
	}
}

func TestRelationsResolver_Resolve_WrongRelationsResponseThrowsError(t *testing.T) {
	mockRelationsAPI(t, http.StatusBadRequest, "relations-api-wrong-json-response.json")

	_, err := relationsResolver.Resolve("3dd42508-a9ab-11e7-8e2d-6debe43a48b4", tid)
	if err == nil {
		assert.FailNow(t, "Relations API wrong json format response should throw error.", err.Error())
	}
}

type mockHttpClient struct {
	mock.Mock
}

func (c *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{}, nil
}
