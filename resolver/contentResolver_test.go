package resolver

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const (
	statusWorking    = "working"
	statusNotWorking = "notWorking"

	outputFile1Content     = "document-store-api-1-content-output.json"
	outputFile2Content     = "document-store-api-2-content-output.json"
	outputFileEmptyContent = "document-store-api-empty-content-output.json"

	tid = "tid_qkeqptjwji"
)

var contentResolver ContentResolver
var dsAPIMock *httptest.Server

func mockDSAPI(appStatus string, outputFileName string) {
	router := mux.NewRouter()
	var contentResolverEndpointHandler http.HandlerFunc

	if appStatus == statusWorking {
		contentResolverEndpointHandler = func(w http.ResponseWriter, r *http.Request) {
			file, err := os.Open("../test-resources/" + outputFileName)
			if err != nil {
				fmt.Println(err)
				return
			}

			defer file.Close()
			io.Copy(w, file)
		}
	} else if appStatus == statusNotWorking {
		contentResolverEndpointHandler = func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	router.Path("/content").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(contentResolverEndpointHandler)})
	dsAPIMock = httptest.NewServer(router)

	contentResolver = NewContentResolver(http.DefaultClient, dsAPIMock.URL+"/content")
}

func Test_callContentResolverApp_1_Content(t *testing.T) {
	mockDSAPI(statusWorking, outputFile1Content)

	uuids := []string{"ab43b1a6-1f47-11e7-b7d3-163f5a7f229c"}
	contents, err := contentResolver.ResolveContents(uuids, tid)
	if err != nil {
		assert.FailNow(t, "Failed retrieving contents.", err.Error())
	}

	assert.Equal(t, 1, len(contents), "There should be 1 content retrieved.")
}

func Test_callContentResolverApp_2_Content(t *testing.T) {
	mockDSAPI(statusWorking, outputFile2Content)

	uuids := []string{"ab43b1a6-1f47-11e7-b7d3-163f5a7f229c", "70c800d8-b3e3-11e6-ba85-95d1533d9a62"}
	contents, err := contentResolver.ResolveContents(uuids, tid)
	if err != nil {
		assert.FailNow(t, "Failed retrieving contents.", err.Error())
	}

	assert.Equal(t, 2, len(contents), "There should be 2 contents retrieved.")
}

func Test_callContentResolverApp_Empty_Content(t *testing.T) {
	mockDSAPI(statusWorking, outputFileEmptyContent)

	uuids := []string{}
	contents, err := contentResolver.ResolveContents(uuids, tid)
	if err != nil {
		assert.FailNow(t, "Failed retrieving contents.", err.Error())
	}

	assert.Equal(t, 0, len(contents), "There should be no contents retrieved.")
}

func Test_callContentResolverApp_NotWorking(t *testing.T) {
	mockDSAPI(statusNotWorking, outputFileEmptyContent)

	uuids := []string{}
	_, err := contentResolver.ResolveContents(uuids, tid)
	if err == nil {
		assert.FailNow(t, "Should have thrown error for failing to reach service.", err.Error())
	}
}
