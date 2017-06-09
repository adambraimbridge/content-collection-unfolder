package main

import (
	"bytes"
	"errors"
	"github.com/Financial-Times/content-collection-unfolder/forwarder"
	"github.com/Financial-Times/content-collection-unfolder/resolver"
	"github.com/Financial-Times/transactionid-utils-go"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const (
	whitelistedCollection = "content-package"
	ignoredCollection     = "story-package"

	invalidUuid = "1234"
	validUuid   = "45163790-eec9-11e6-abbc-ee7d9c5b3b90"

	inputFile = "content-collection.json"

	errorJson = "{\"msg\":\"error\"}"

	firstItemUuid  = "d4986a58-de3b-11e6-86ac-f253db7791c6"
	secondItemUuid = "d4986a58-de3b-11e6-86ac-f253db7791c6"
	lastModified   = "2017-01-31T15:33:21.687Z"
)

func TestInvalidUuid(t *testing.T) {
	mf, mur, mcr, mcp, u := newUnfolderWithMocks()

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	req := buildRequest(t, server.URL, whitelistedCollection, invalidUuid, readTestFile(t, inputFile), tid)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	verifyResponse(t, http.StatusBadRequest, tid, resp)

	mf.AssertNotCalled(t, "Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mur.AssertNotCalled(t, "Resolve", mock.Anything, mock.Anything)
	mcr.AssertNotCalled(t, "ResolveContents", mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestForwarderError(t *testing.T) {
	mf, mur, mcr, mcp, u := newUnfolderWithMocks()

	mf.On("Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&forwarder.ForwarderResponse{}, errors.New("Forwarder error"))

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, validUuid, body, tid)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusInternalServerError, tid, resp)

	mf.AssertCalled(t, "Forward",
		mock.MatchedBy(func(actualTid string) bool {
			assert.Equal(t, tid, actualTid)
			return true
		}),
		mock.MatchedBy(func(actualUuid string) bool {
			assert.Equal(t, validUuid, actualUuid)
			return true
		}),
		mock.MatchedBy(func(actualCollectionType string) bool {
			assert.Equal(t, whitelistedCollection, actualCollectionType)
			return true
		}),
		mock.MatchedBy(func(actualBody []byte) bool {
			assert.Equal(t, body, actualBody)
			return true
		}))
	mur.AssertNotCalled(t, "Resolve", mock.Anything, mock.Anything)
	mcr.AssertNotCalled(t, "ResolveContents", mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestForwarderNon200Response(t *testing.T) {
	mf, mur, mcr, mcp, u := newUnfolderWithMocks()

	fwResp := &forwarder.ForwarderResponse{http.StatusUnprocessableEntity, []byte(errorJson)}
	mf.On("Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(fwResp, nil)

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, validUuid, body, tid)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, fwResp.Status, tid, resp)

	respBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, fwResp.ResponseBody, respBody)

	mf.AssertCalled(t, "Forward",
		mock.MatchedBy(func(actualTid string) bool {
			assert.Equal(t, tid, actualTid)
			return true
		}),
		mock.MatchedBy(func(actualUuid string) bool {
			assert.Equal(t, validUuid, actualUuid)
			return true
		}),
		mock.MatchedBy(func(actualCollectionType string) bool {
			assert.Equal(t, whitelistedCollection, actualCollectionType)
			return true
		}),
		mock.MatchedBy(func(actualBody []byte) bool {
			assert.Equal(t, body, actualBody)
			return true
		}))
	mur.AssertNotCalled(t, "Resolve", mock.Anything, mock.Anything)
	mcr.AssertNotCalled(t, "ResolveContents", mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestNotWhitelistedCollectionType(t *testing.T) {
	mf, mur, mcr, mcp, u := newUnfolderWithMocks()

	mf.On("Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&forwarder.ForwarderResponse{http.StatusOK, []byte{}}, nil)

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, ignoredCollection, validUuid, body, tid)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusOK, tid, resp)

	mf.AssertExpectations(t)
	mur.AssertNotCalled(t, "Resolve", mock.Anything, mock.Anything)
	mcr.AssertNotCalled(t, "ResolveContents", mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestUuidResolverError(t *testing.T) {
	mf, mur, mcr, mcp, u := newUnfolderWithMocks()

	fwResp := &forwarder.ForwarderResponse{http.StatusOK, []byte{}}
	mf.On("Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(fwResp, nil)
	mur.On("Resolve", mock.Anything, mock.Anything).
		Return(&resolver.UuidsAndDate{}, errors.New("Uuid resolver error"))

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, validUuid, body, tid)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusBadRequest, tid, resp)

	mf.AssertExpectations(t)
	mur.AssertCalled(t, "Resolve",
		mock.MatchedBy(func(actualReqBody []byte) bool {
			assert.Equal(t, body, actualReqBody)
			return true
		}),
		mock.MatchedBy(func(actualRespBody []byte) bool {
			assert.Equal(t, fwResp.ResponseBody, actualRespBody)
			return true
		}))
	mcr.AssertNotCalled(t, "ResolveContents", mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestContentResolverError(t *testing.T) {
	mf, mur, mcr, mcp, u := newUnfolderWithMocks()

	fwResp := &forwarder.ForwarderResponse{http.StatusOK, []byte{}}
	mf.On("Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(fwResp, nil)
	uuidsAndDate := &resolver.UuidsAndDate{[]string{firstItemUuid, secondItemUuid}, lastModified}
	mur.On("Resolve", mock.Anything, mock.Anything).
		Return(uuidsAndDate, nil)
	mcr.On("ResolveContents", mock.Anything, mock.Anything).
		Return([]map[string]interface{}{}, errors.New("Content resolver error"))

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, validUuid, body, tid)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusInternalServerError, tid, resp)

	mf.AssertExpectations(t)
	mur.AssertExpectations(t)
	mcr.AssertCalled(t, "ResolveContents",
		mock.MatchedBy(func(actualUuidArr []string) bool {
			assert.Equal(t, uuidsAndDate.UuidArr, actualUuidArr)
			return true
		}),
		mock.MatchedBy(func(actualTid string) bool {
			assert.Equal(t, actualTid, tid)
			return true
		}))
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestAllOk(t *testing.T) {
	mf, mur, mcr, mcp, u := newUnfolderWithMocks()

	fwResp := &forwarder.ForwarderResponse{http.StatusOK, []byte{}}
	mf.On("Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(fwResp, nil)
	uuidsAndDate := &resolver.UuidsAndDate{[]string{firstItemUuid, secondItemUuid}, lastModified}
	mur.On("Resolve", mock.Anything, mock.Anything).
		Return(uuidsAndDate, nil)
	contentArr := []map[string]interface{}{
		{"uuid": firstItemUuid},
		{"uuid": secondItemUuid},
	}
	mcr.On("ResolveContents", mock.Anything, mock.Anything).
		Return(contentArr, nil)
	mcp.On("Send", mock.Anything, mock.Anything, mock.Anything)

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, validUuid, body, tid)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusOK, tid, resp)

	mf.AssertExpectations(t)
	mur.AssertExpectations(t)
	mcr.AssertExpectations(t)
	mcp.AssertCalled(t, "Send",
		mock.MatchedBy(func(actualTid string) bool {
			assert.Equal(t, tid, actualTid)
			return true
		}),
		mock.MatchedBy(func(actualLastModified string) bool {
			assert.Equal(t, uuidsAndDate.LastModified, actualLastModified)
			return true
		}),
		mock.MatchedBy(func(actualContentArr []map[string]interface{}) bool {
			assert.Equal(t, contentArr, actualContentArr)
			return true
		}))
}

func TestMarshallingErrorIs500(t *testing.T) {
	recorder := httptest.NewRecorder()

	u := &unfolder{}
	u.writeMap(recorder, http.StatusOK, map[string]interface{}{"dude, what?": func() {}})

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func newUnfolderWithMocks() (*mockForwarder, *mockUuidResolver, *mockContentResolver, *mockContentProducer, *unfolder) {
	mf := new(mockForwarder)
	mur := new(mockUuidResolver)
	mcr := new(mockContentResolver)
	mcp := new(mockContentProducer)
	u := newUnfolder(mf, mur, mcr, mcp, []string{whitelistedCollection})
	return mf, mur, mcr, mcp, u
}

func startTestServer(u *unfolder) *httptest.Server {
	router := mux.NewRouter()
	router.HandleFunc(unfolderPath, u.handle).Methods(http.MethodPut)

	return httptest.NewServer(router)
}

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

func verifyResponse(t *testing.T, expectedStatus int, expectedTid string, resp *http.Response) {
	assert.Equal(t, expectedStatus, resp.StatusCode)
	assert.Equal(t, expectedTid, resp.Header.Get(transactionidutils.TransactionIDHeader))
	assert.Equal(t, "application/json;charset=utf-8", resp.Header.Get("Content-Type"))
}

type mockForwarder struct {
	mock.Mock
}

func (mf *mockForwarder) Forward(tid string, uuid string, collectionType string, reqBody []byte) (*forwarder.ForwarderResponse, error) {
	args := mf.Called(tid, uuid, collectionType, reqBody)
	return args.Get(0).(*forwarder.ForwarderResponse), args.Error(1)
}

type mockUuidResolver struct {
	mock.Mock
}

func (mur *mockUuidResolver) Resolve(reqData []byte, respData []byte) (*resolver.UuidsAndDate, error) {
	args := mur.Called(reqData, respData)
	return args.Get(0).(*resolver.UuidsAndDate), args.Error(1)
}

type mockContentResolver struct {
	mock.Mock
}

func (mcr *mockContentResolver) ResolveContents(uuids []string, tid string) ([]map[string]interface{}, error) {
	args := mcr.Called(uuids, tid)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

type mockContentProducer struct {
	mock.Mock
}

func (mcp *mockContentProducer) Send(tid string, lastModified string, contentArr []map[string]interface{}) {
	mcp.Called(tid, lastModified, contentArr)
}
