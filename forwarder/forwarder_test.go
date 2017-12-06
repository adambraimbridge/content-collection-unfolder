package forwarder

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Financial-Times/transactionid-utils-go"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const (
	testPath       = "/test"
	mockTestPath   = testPath + "/{collectionType}/{uuid}"
	testTid        = "tid_1234567890"
	testUuid       = "ab43b1a6-1f47-11e7-b7d3-163f5a7f229c"
	testCollection = "test-collection"
	testReqBody    = "{\"reqField\":\"testValue\"}"
	testRespBody   = "{\"respField\":\"testValue\"}"
)

func mockWriter(t *testing.T, respStatus int) *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc(mockTestPath, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, testTid, transactionidutils.GetTransactionIDFromRequest(r))
		assert.Equal(t, r.Header.Get("Content-Type"), "application/json;charset=utf-8")

		vars := mux.Vars(r)
		assert.Equal(t, testUuid, vars["uuid"])
		assert.Equal(t, testCollection, vars["collectionType"])

		defer r.Body.Close()
		reqBody, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, []byte(testReqBody), reqBody)

		w.WriteHeader(respStatus)
		_, err = w.Write([]byte(testRespBody))
		assert.NoError(t, err)
	}).Methods(http.MethodPut)

	return httptest.NewServer(router)
}

func TestForwardingNoSchema(t *testing.T) {
	f := NewForwarder(http.DefaultClient, "some-bad-url")
	_, err := f.Forward(testTid, testUuid, testCollection, []byte(testReqBody))

	assert.Error(t, err)
}

func TestForwardingNotFound(t *testing.T) {
	mockServer := mockWriter(t, http.StatusOK)
	defer mockServer.Close()

	f := NewForwarder(http.DefaultClient, mockServer.URL)
	resp, err := f.Forward(testTid, testUuid, testCollection, []byte(testReqBody))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.Status)
}

func TestForwardingOkResponse(t *testing.T) {
	mockServer := mockWriter(t, http.StatusOK)
	defer mockServer.Close()

	f := NewForwarder(http.DefaultClient, mockServer.URL+testPath)
	resp, err := f.Forward(testTid, testUuid, testCollection, []byte(testReqBody))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Status)
	assert.Equal(t, []byte(testRespBody), resp.ResponseBody)
}

func TestForwardingErrorResponse(t *testing.T) {
	mockServer := mockWriter(t, http.StatusBadRequest)
	defer mockServer.Close()

	f := NewForwarder(http.DefaultClient, mockServer.URL+testPath)
	resp, err := f.Forward(testTid, testUuid, testCollection, []byte(testReqBody))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, []byte(testRespBody), resp.ResponseBody)
}
