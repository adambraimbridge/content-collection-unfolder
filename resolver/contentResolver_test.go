package resolver

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const (
	statusWorking    = "working"
	statusNotWorking = "notWorking"
	tid              = "tid_qkeqptjwji"
	requestTimeout   = time.Second * 2
)

var contentResolver ContentResolver
var dsAPIMock *httptest.Server

func mockDSAPI(t *testing.T, appStatus string, outputFileName string) {
	file, err := os.Open("../test-resources/" + outputFileName)
	assert.NoError(t, err, "Opening file shouldn't throw error.")
	defer file.Close()
	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	assert.NoError(t, err, "Reading from file to []byte buffer shouldn't throw error.")
	mockDSAPIBytes(appStatus, buf.Bytes())
}

func mockDSAPIBytes(appStatus string, output []byte) {
	router := mux.NewRouter()
	var contentResolverEndpointHandler http.HandlerFunc

	if appStatus == statusWorking {
		contentResolverEndpointHandler = func(w http.ResponseWriter, r *http.Request) {
			w.Write(output)
		}
	} else if appStatus == statusNotWorking {
		contentResolverEndpointHandler = func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	router.Path("/content").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(contentResolverEndpointHandler)})
	dsAPIMock = httptest.NewServer(router)

	contentResolver = NewContentResolver(http.DefaultClient, dsAPIMock.URL+"/content", requestTimeout)
}

func Test_callContentResolverApp_1_Content(t *testing.T) {
	mockDSAPI(t, statusWorking, "document-store-api-1-content-output.json")

	diffUuids := []string{"ab43b1a6-1f47-11e7-b7d3-163f5a7f229c"}
	contents, err := contentResolver.ResolveContentsNew(diffUuids, tid, requestTimeout)
	if err != nil {
		assert.FailNow(t, "Failed retrieving contents.", err.Error())
	}

	assert.Equal(t, 1, len(contents), "There should be 1 content retrieved.")
}

func Test_callContentResolverApp_2_Content(t *testing.T) {
	mockDSAPI(t, statusWorking, "document-store-api-2-content-output.json")

	diffUuids := []string{"ab43b1a6-1f47-11e7-b7d3-163f5a7f229c", "70c800d8-b3e3-11e6-ba85-95d1533d9a62"}
	contents, err := contentResolver.ResolveContentsNew(diffUuids, tid, requestTimeout)
	if err != nil {
		assert.FailNow(t, "Failed retrieving contents.", err.Error())
	}

	assert.Equal(t, 2, len(contents), "There should be 2 contents retrieved.")
}

func Test_callContentResolverApp_3_Content(t *testing.T) {
	mockDSAPI(t, statusWorking, "document-store-api-3-content-output.json")

	diffUuids := []string{"ab43b1a6-1f47-11e7-b7d3-163f5a7f229c", "70c800d8-b3e3-11e6-ba85-95d1533d9a62", "70c800d8-b3e3-11e6-ba85-95d1533d9a63"}
	contents, err := contentResolver.ResolveContentsNew(diffUuids, tid, requestTimeout)
	if err != nil {
		assert.FailNow(t, "Failed retrieving contents.", err.Error())
	}

	assert.Equal(t, 3, len(contents), "There should be 3 contents retrieved.")
}

func Test_callContentResolverApp_Empty_Content(t *testing.T) {
	mockDSAPIBytes(statusWorking, []byte("[]"))

	var diffUuids []string
	contents, err := contentResolver.ResolveContentsNew(diffUuids, tid, requestTimeout)
	if err != nil {
		assert.FailNow(t, "Failed retrieving contents.", err.Error())
	}

	assert.Equal(t, 0, len(contents), "There should be no contents retrieved.")
}

func Test_callContentResolverApp_NotWorking(t *testing.T) {
	mockDSAPIBytes(statusNotWorking, []byte("[]"))

	var diffUuids []string
	_, err := contentResolver.ResolveContentsNew(diffUuids, tid, requestTimeout)
	if err == nil {
		assert.FailNow(t, "Should have thrown error for failing to reach service.", err.Error())
	}
}
