package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Financial-Times/content-collection-unfolder/forwarder"
	"github.com/Financial-Times/content-collection-unfolder/relations"
	"github.com/Financial-Times/content-collection-unfolder/resolver"
	"github.com/Financial-Times/transactionid-utils-go"
	"github.com/Workiva/go-datastructures/set"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	ignoredCollection = "story-package"
	invalidUuid       = "1234"
	errorJson         = "{\"msg\":\"error\"}"
	requestTimeout    = time.Second * 2
)

func TestInvalidUuid(t *testing.T) {
	mur, mrr, mcd, mf, mcr, mcp, u := newUnfolderWithMocks()

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	req := buildRequest(t, server.URL, whitelistedCollection, invalidUuid, readTestFile(t, inputFile), tid)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	verifyResponse(t, http.StatusBadRequest, tid, resp)

	mur.AssertNotCalled(t, "Resolve", mock.Anything)
	mrr.AssertNotCalled(t, "Resolve", mock.Anything, mock.Anything)
	mcd.AssertNotCalled(t, "SymmetricDifference", mock.Anything, mock.Anything)
	mf.AssertNotCalled(t, "Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mcr.AssertNotCalled(t, "ResolveContentsNew", mock.Anything, mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestUuidResolverError(t *testing.T) {
	mur, mrr, mcd, mf, mcr, mcp, u := newUnfolderWithMocks()

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, collectionUuid, body, tid)

	mur.On("Resolve", mock.MatchedBy(expectByteSlice(t, body))).
		Return(resolver.UuidsAndDate{}, errors.New("uuid resolver error"))

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusBadRequest, tid, resp)

	mur.AssertExpectations(t)
	mrr.AssertNotCalled(t, "Resolve", mock.Anything, mock.Anything)
	mcd.AssertNotCalled(t, "SymmetricDifference", mock.Anything, mock.Anything)
	mf.AssertNotCalled(t, "Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mcr.AssertNotCalled(t, "ResolveContentsNew", mock.Anything, mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestRelationsResolverError(t *testing.T) {
	mur, mrr, mcd, mf, mcr, mcp, u := newUnfolderWithMocks()

	uuidsAndDate := resolver.UuidsAndDate{
		UuidArr:      []string{firstExistingItemUuid, secondExistingItemUuid, addedItemUuid},
		LastModified: lastModified,
	}

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, collectionUuid, body, tid)

	mur.On("Resolve", mock.MatchedBy(expectByteSlice(t, body))).Return(uuidsAndDate, nil)
	mrr.On("Resolve",
		mock.MatchedBy(expectString(t, collectionUuid)),
		mock.MatchedBy(expectString(t, tid))).
		Return(&relations.CCRelations{}, errors.New("relations resolver error"))

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusInternalServerError, tid, resp)

	mock.AssertExpectationsForObjects(t, mur, mrr)
	mcd.AssertNotCalled(t, "SymmetricDifference", mock.Anything, mock.Anything)
	mf.AssertNotCalled(t, "Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mcr.AssertNotCalled(t, "ResolveContentsNew", mock.Anything, mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestForwarderError(t *testing.T) {
	mur, mrr, mcd, mf, mcr, mcp, u := newUnfolderWithMocks()

	uuidsAndDate := resolver.UuidsAndDate{
		UuidArr:      []string{firstExistingItemUuid, secondExistingItemUuid, addedItemUuid},
		LastModified: lastModified,
	}
	oldRelations := relations.CCRelations{
		ContainedIn: leadArticleUuid,
		Contains:    []string{firstExistingItemUuid, secondExistingItemUuid, deletedItemUuid},
	}
	diffUuidsSet := set.New()
	diffUuidsSet.Add(addedItemUuid)
	diffUuidsSet.Add(deletedItemUuid)

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, collectionUuid, body, tid)

	mur.On("Resolve", mock.MatchedBy(expectByteSlice(t, body))).Return(uuidsAndDate, nil)
	mrr.On("Resolve",
		mock.MatchedBy(expectString(t, collectionUuid)),
		mock.MatchedBy(expectString(t, tid))).
		Return(&oldRelations, nil)
	mcd.On("SymmetricDifference",
		mock.MatchedBy(expectStringSlice(t, uuidsAndDate.UuidArr)),
		mock.MatchedBy(expectStringSlice(t, oldRelations.Contains))).
		Return(diffUuidsSet)
	mf.On("Forward",
		mock.MatchedBy(expectString(t, tid)),
		mock.MatchedBy(expectString(t, collectionUuid)),
		mock.MatchedBy(expectString(t, whitelistedCollection)),
		mock.MatchedBy(expectByteSlice(t, body))).
		Return(forwarder.ForwarderResponse{}, errors.New("forwarder error"))

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusInternalServerError, tid, resp)

	mock.AssertExpectationsForObjects(t, mur, mrr, mcd, mf)
	mcr.AssertNotCalled(t, "ResolveContentsNew", mock.Anything, mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestForwarderNon200Response(t *testing.T) {
	mur, mrr, mcd, mf, mcr, mcp, u := newUnfolderWithMocks()

	uuidsAndDate := resolver.UuidsAndDate{
		UuidArr:      []string{firstExistingItemUuid, secondExistingItemUuid, addedItemUuid},
		LastModified: lastModified,
	}
	oldRelations := relations.CCRelations{
		ContainedIn: leadArticleUuid,
		Contains:    []string{firstExistingItemUuid, secondExistingItemUuid, deletedItemUuid},
	}
	diffUuidsSet := set.New()
	diffUuidsSet.Add(addedItemUuid)
	diffUuidsSet.Add(deletedItemUuid)

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, collectionUuid, body, tid)

	mur.On("Resolve", mock.MatchedBy(expectByteSlice(t, body))).Return(uuidsAndDate, nil)
	mrr.On("Resolve",
		mock.MatchedBy(expectString(t, collectionUuid)),
		mock.MatchedBy(expectString(t, tid))).
		Return(&oldRelations, nil)
	mcd.On("SymmetricDifference",
		mock.MatchedBy(expectStringSlice(t, uuidsAndDate.UuidArr)),
		mock.MatchedBy(expectStringSlice(t, oldRelations.Contains))).
		Return(diffUuidsSet)
	fwResp := forwarder.ForwarderResponse{Status: http.StatusUnprocessableEntity, ResponseBody: []byte(errorJson)}
	mf.On("Forward",
		mock.MatchedBy(expectString(t, tid)),
		mock.MatchedBy(expectString(t, collectionUuid)),
		mock.MatchedBy(expectString(t, whitelistedCollection)),
		mock.MatchedBy(expectByteSlice(t, body))).
		Return(fwResp, nil)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, fwResp.Status, tid, resp)

	respBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, fwResp.ResponseBody, respBody)

	mock.AssertExpectationsForObjects(t, mur, mrr, mcd, mf)
	mcr.AssertNotCalled(t, "ResolveContentsNew", mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestNotWhitelistedCollectionType(t *testing.T) {
	mur, mrr, mcd, mf, mcr, mcp, u := newUnfolderWithMocks()

	uuidsAndDate := resolver.UuidsAndDate{
		UuidArr:      []string{firstExistingItemUuid, secondExistingItemUuid, addedItemUuid},
		LastModified: lastModified,
	}
	oldRelations := relations.CCRelations{
		ContainedIn: leadArticleUuid,
		Contains:    []string{firstExistingItemUuid, secondExistingItemUuid, deletedItemUuid},
	}
	diffUuidsSet := set.New()
	diffUuidsSet.Add(addedItemUuid)
	diffUuidsSet.Add(deletedItemUuid)

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, ignoredCollection, collectionUuid, body, tid)

	mur.On("Resolve", mock.MatchedBy(expectByteSlice(t, body))).Return(uuidsAndDate, nil)
	mrr.On("Resolve",
		mock.MatchedBy(expectString(t, collectionUuid)),
		mock.MatchedBy(expectString(t, tid))).
		Return(&oldRelations, nil)
	mcd.On("SymmetricDifference",
		mock.MatchedBy(expectStringSlice(t, uuidsAndDate.UuidArr)),
		mock.MatchedBy(expectStringSlice(t, oldRelations.Contains))).
		Return(diffUuidsSet)
	mf.On("Forward",
		mock.MatchedBy(expectString(t, tid)),
		mock.MatchedBy(expectString(t, collectionUuid)),
		mock.MatchedBy(expectString(t, ignoredCollection)),
		mock.MatchedBy(expectByteSlice(t, body))).
		Return(forwarder.ForwarderResponse{http.StatusOK, []byte{}}, nil)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusOK, tid, resp)

	mur.AssertCalled(t, "Resolve",
		mock.MatchedBy(func(actualReqBody []byte) bool {
			assert.Equal(t, body, actualReqBody)
			return true
		}))

	mock.AssertExpectationsForObjects(t, mur, mrr, mcd, mf)
	mcr.AssertNotCalled(t, "ResolveContentsNew", mock.Anything, mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestContentResolverError(t *testing.T) {
	mur, mrr, mcd, mf, mcr, mcp, u := newUnfolderWithMocks()

	uuidsAndDate := resolver.UuidsAndDate{
		UuidArr:      []string{firstExistingItemUuid, secondExistingItemUuid, addedItemUuid},
		LastModified: lastModified,
	}
	oldRelations := relations.CCRelations{
		ContainedIn: leadArticleUuid,
		Contains:    []string{firstExistingItemUuid, secondExistingItemUuid, deletedItemUuid},
	}
	diffUuidsSet := set.New()
	diffUuidsSet.Add(addedItemUuid)
	diffUuidsSet.Add(deletedItemUuid)

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, collectionUuid, body, tid)

	mur.On("Resolve", mock.MatchedBy(expectByteSlice(t, body))).Return(uuidsAndDate, nil)
	mrr.On("Resolve",
		mock.MatchedBy(expectString(t, collectionUuid)),
		mock.MatchedBy(expectString(t, tid))).
		Return(&oldRelations, nil)
	mcd.On("SymmetricDifference",
		mock.MatchedBy(expectStringSlice(t, uuidsAndDate.UuidArr)),
		mock.MatchedBy(expectStringSlice(t, oldRelations.Contains))).
		Return(diffUuidsSet)
	mf.On("Forward",
		mock.MatchedBy(expectString(t, tid)),
		mock.MatchedBy(expectString(t, collectionUuid)),
		mock.MatchedBy(expectString(t, whitelistedCollection)),
		mock.MatchedBy(expectByteSlice(t, body))).
		Return(forwarder.ForwarderResponse{http.StatusOK, []byte{}}, nil)

	mcr.On("ResolveContentsNew",
		mock.MatchedBy(expectSet(t, diffUuidsSet)),
		mock.MatchedBy(expectString(t, tid)),
		mock.MatchedBy(expectTimeDuration(t, requestTimeout))).
		Return([]map[string]interface{}{}, errors.New("content resolver error"))

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusInternalServerError, tid, resp)

	mock.AssertExpectationsForObjects(t, mur, mrr, mcd, mf, mcr)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestAllOk(t *testing.T) {
	mur, mrr, mcd, mf, mcr, mcp, u := newUnfolderWithMocks()

	uuidsAndDate := resolver.UuidsAndDate{
		UuidArr:      []string{firstExistingItemUuid, secondExistingItemUuid, addedItemUuid},
		LastModified: lastModified,
	}
	oldRelations := relations.CCRelations{
		ContainedIn: leadArticleUuid,
		Contains:    []string{firstExistingItemUuid, secondExistingItemUuid, deletedItemUuid},
	}
	diffUuidsSet := set.New()
	diffUuidsSet.Add(addedItemUuid)
	diffUuidsSet.Add(deletedItemUuid)

	contentArr := []map[string]interface{}{
		{addedItemUuid: addedItemUuid},
		{deletedItemUuid: deletedItemUuid},
		{leadArticleUuid: leadArticleUuid},
	}

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, collectionUuid, body, tid)

	mur.On("Resolve", mock.MatchedBy(expectByteSlice(t, body))).Return(uuidsAndDate, nil)
	mrr.On("Resolve",
		mock.MatchedBy(expectString(t, collectionUuid)),
		mock.MatchedBy(expectString(t, tid))).
		Return(&oldRelations, nil)
	mcd.On("SymmetricDifference",
		mock.MatchedBy(expectStringSlice(t, uuidsAndDate.UuidArr)),
		mock.MatchedBy(expectStringSlice(t, oldRelations.Contains))).
		Return(diffUuidsSet)
	mf.On("Forward",
		mock.MatchedBy(expectString(t, tid)),
		mock.MatchedBy(expectString(t, collectionUuid)),
		mock.MatchedBy(expectString(t, whitelistedCollection)),
		mock.MatchedBy(expectByteSlice(t, body))).
		Return(forwarder.ForwarderResponse{http.StatusOK, []byte{}}, nil)
	mcr.On("ResolveContentsNew",
		mock.MatchedBy(expectSet(t, diffUuidsSet)),
		mock.MatchedBy(expectString(t, tid)),
		mock.MatchedBy(expectTimeDuration(t, requestTimeout))).
		Return(contentArr, nil)
	mcp.On("Send",
		mock.MatchedBy(expectString(t, tid)),
		mock.MatchedBy(expectString(t, uuidsAndDate.LastModified)),
		mock.MatchedBy(expectMap(t, contentArr)))

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusOK, tid, resp)

	mock.AssertExpectationsForObjects(t, mur, mrr, mcd, mf, mcr, mcp)
}

func TestAllOk_NoLeadArticleRelation(t *testing.T) {
	mur, mrr, mcd, mf, mcr, mcp, u := newUnfolderWithMocks()

	uuidsAndDate := resolver.UuidsAndDate{
		UuidArr:      []string{firstExistingItemUuid, secondExistingItemUuid, addedItemUuid},
		LastModified: lastModified,
	}
	oldRelations := relations.CCRelations{}
	diffUuidsSet := set.New()
	diffUuidsSet.Add(firstExistingItemUuid)
	diffUuidsSet.Add(addedItemUuid)
	diffUuidsSet.Add(deletedItemUuid)
	contentArr := []map[string]interface{}{
		{firstExistingItemUuid: firstExistingItemUuid},
		{secondExistingItemUuid: secondExistingItemUuid},
		{addedItemUuid: addedItemUuid},
	}

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, collectionUuid, body, tid)

	mur.On("Resolve", mock.MatchedBy(expectByteSlice(t, body))).Return(uuidsAndDate, nil)
	mrr.On("Resolve",
		mock.MatchedBy(expectString(t, collectionUuid)),
		mock.MatchedBy(expectString(t, tid))).
		Return(&oldRelations, nil)
	mcd.On("SymmetricDifference",
		mock.MatchedBy(expectStringSlice(t, uuidsAndDate.UuidArr)),
		mock.MatchedBy(expectStringSlice(t, oldRelations.Contains))).
		Return(diffUuidsSet)
	mf.On("Forward",
		mock.MatchedBy(expectString(t, tid)),
		mock.MatchedBy(expectString(t, collectionUuid)),
		mock.MatchedBy(expectString(t, whitelistedCollection)),
		mock.MatchedBy(expectByteSlice(t, body))).
		Return(forwarder.ForwarderResponse{http.StatusOK, []byte{}}, nil)
	mcr.On("ResolveContentsNew",
		mock.MatchedBy(func(actualDiffUuids []string) bool {
			for _, uuid := range actualDiffUuids {
				assert.True(t, diffUuidsSet.Exists(uuid))
			}
			assert.False(t, contains(actualDiffUuids, leadArticleUuid))
			return true
		}),
		mock.MatchedBy(expectString(t, tid)),
		mock.MatchedBy(expectTimeDuration(t, requestTimeout))).
		Return(contentArr, nil)
	mcp.On("Send",
		mock.MatchedBy(expectString(t, tid)),
		mock.MatchedBy(expectString(t, uuidsAndDate.LastModified)),
		mock.MatchedBy(expectMap(t, contentArr)))

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusOK, tid, resp)

	mock.AssertExpectationsForObjects(t, mur, mrr, mcd, mf, mcr, mcp)
}

func TestAllOk_NewEmptyCollection_NoRelations(t *testing.T) {
	mur, mrr, mcd, mf, mcr, mcp, u := newUnfolderWithMocks()

	uuidsAndDate := resolver.UuidsAndDate{
		UuidArr:      []string{},
		LastModified: lastModified,
	}
	oldRelations := relations.CCRelations{}
	diffUuidsSet := set.New()

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, collectionUuid, body, tid)

	mur.On("Resolve", mock.MatchedBy(expectByteSlice(t, body))).Return(uuidsAndDate, nil)
	mrr.On("Resolve",
		mock.MatchedBy(expectString(t, collectionUuid)),
		mock.MatchedBy(expectString(t, tid))).
		Return(&oldRelations, nil)
	mcd.On("SymmetricDifference",
		mock.MatchedBy(expectStringSlice(t, uuidsAndDate.UuidArr)),
		mock.MatchedBy(expectStringSlice(t, oldRelations.Contains))).
		Return(diffUuidsSet)
	mf.On("Forward",
		mock.MatchedBy(expectString(t, tid)),
		mock.MatchedBy(expectString(t, collectionUuid)),
		mock.MatchedBy(expectString(t, whitelistedCollection)),
		mock.MatchedBy(expectByteSlice(t, body))).
		Return(forwarder.ForwarderResponse{http.StatusOK, []byte{}}, nil)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusOK, tid, resp)

	mock.AssertExpectationsForObjects(t, mur, mrr, mcd, mf)
	mcr.AssertNotCalled(t, "ResolveContentsNew", mock.Anything, mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestMarshallingErrorIs500(t *testing.T) {
	recorder := httptest.NewRecorder()

	writeMap(recorder, http.StatusOK, map[string]interface{}{"dude, what?": func() {}})

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func newUnfolderWithMocks() (*mockUuidResolver, *mockRelationsResolver, *mockCollectionsDiffer, *mockForwarder, *mockContentResolver, *mockContentProducer, *unfolder) {
	mur := new(mockUuidResolver)
	mrr := new(mockRelationsResolver)
	mcd := new(mockCollectionsDiffer)
	mf := new(mockForwarder)
	mcr := new(mockContentResolver)
	mcp := new(mockContentProducer)
	u := newUnfolder(mur, mrr, mcd, mf, mcr, mcp, []string{whitelistedCollection})
	return mur, mrr, mcd, mf, mcr, mcp, u
}

func startTestServer(u *unfolder) *httptest.Server {
	router := mux.NewRouter()
	router.HandleFunc(unfolderPath, u.handle).Methods(http.MethodPut)

	return httptest.NewServer(router)
}

func verifyResponse(t *testing.T, expectedStatus int, expectedTid string, resp *http.Response) {
	assert.Equal(t, expectedStatus, resp.StatusCode)
	assert.Equal(t, expectedTid, resp.Header.Get(transactionidutils.TransactionIDHeader))
	assert.Equal(t, "application/json;charset=utf-8", resp.Header.Get("Content-Type"))
}

type mockForwarder struct {
	mock.Mock
}

func (mf *mockForwarder) Forward(tid string, uuid string, collectionType string, reqBody []byte) (forwarder.ForwarderResponse, error) {
	args := mf.Called(tid, uuid, collectionType, reqBody)
	return args.Get(0).(forwarder.ForwarderResponse), args.Error(1)
}

type mockUuidResolver struct {
	mock.Mock
}

func (mur *mockUuidResolver) Resolve(reqData []byte) (resolver.UuidsAndDate, error) {
	args := mur.Called(reqData)
	return args.Get(0).(resolver.UuidsAndDate), args.Error(1)
}

type mockContentResolver struct {
	mock.Mock
}

func (mcr *mockContentResolver) ResolveContents(diffUuids []string, tid string) ([]map[string]interface{}, error) {
	args := mcr.Called(diffUuids, tid)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (mcr *mockContentResolver) ResolveContentsNew(diffUuids []string, tid string, requestTimeout time.Duration) ([]map[string]interface{}, error) {
	args := mcr.Called(diffUuids, tid, requestTimeout)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (mcr *mockContentResolver) GetRequestTimeout() time.Duration {
	return requestTimeout
}

type mockContentProducer struct {
	mock.Mock
}

func (mcp *mockContentProducer) Send(tid string, lastModified string, contents []map[string]interface{}) {
	mcp.Called(tid, lastModified, contents)
	return
}

type mockRelationsResolver struct {
	mock.Mock
}

func (mrr *mockRelationsResolver) Resolve(contentCollectionUUID string, tid string) (*relations.CCRelations, error) {
	args := mrr.Called(contentCollectionUUID, tid)
	return args.Get(0).(*relations.CCRelations), args.Error(1)
}

type mockCollectionsDiffer struct {
	mock.Mock
}

func (mcd *mockCollectionsDiffer) SymmetricDifference(incomingCollectionUuids []string, oldCollectionUuids []string) *set.Set {
	args := mcd.Called(incomingCollectionUuids, oldCollectionUuids)
	return args.Get(0).(*set.Set)
}

func contains(values []string, valueToCheck string) bool {
	for _, value := range values {
		if valueToCheck == value {
			return true
		}
	}
	return false
}

func expectString(t *testing.T, expected string) func(string) bool {
	return func(actual string) bool {
		assert.Equal(t, expected, actual)
		return true
	}
}

func expectByteSlice(t *testing.T, expected []byte) func([]byte) bool {
	return func(actual []byte) bool {
		assert.Equal(t, expected, actual)
		return true
	}
}

func expectStringSlice(t *testing.T, expected []string) func([]string) bool {
	return func(actual []string) bool {
		assert.Equal(t, expected, actual)
		return true
	}
}

func expectSet(t *testing.T, expected *set.Set) func([]string) bool {
	return func(actual []string) bool {
		for _, uuid := range actual {
			assert.True(t, expected.Exists(uuid))
		}
		return true
	}
}

func expectMap(t *testing.T, expected []map[string]interface{}) func([]map[string]interface{}) bool {
	return func(actual []map[string]interface{}) bool {
		assert.Equal(t, expected, actual)
		return true
	}
}

func expectTimeDuration(t *testing.T, expected time.Duration) func(time.Duration) bool {
	return func(actual time.Duration) bool {
		assert.Equal(t, expected, actual)
		return true
	}
}
