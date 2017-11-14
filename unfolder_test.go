package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Financial-Times/content-collection-unfolder/forwarder"
	"github.com/Financial-Times/content-collection-unfolder/relations"
	"github.com/Financial-Times/content-collection-unfolder/resolver"
	"github.com/Financial-Times/transactionid-utils-go"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	ignoredCollection = "story-package"
	invalidUuid       = "1234"
	errorJson         = "{\"msg\":\"error\"}"
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
	mcd.AssertNotCalled(t, "Diff", mock.Anything, mock.Anything)
	mf.AssertNotCalled(t, "Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mcr.AssertNotCalled(t, "ResolveContents", mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestUuidResolverError(t *testing.T) {
	mur, mrr, mcd, mf, mcr, mcp, u := newUnfolderWithMocks()

	mur.On("Resolve", mock.Anything).
		Return(resolver.UuidsAndDate{}, errors.New("Uuid resolver error"))

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, collectionUuid, body, tid)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusBadRequest, tid, resp)

	mur.AssertCalled(t, "Resolve",
		mock.MatchedBy(func(actualReqBody []byte) bool {
			assert.Equal(t, body, actualReqBody)
			return true
		}))

	mrr.AssertNotCalled(t, "Resolve", mock.Anything, mock.Anything)
	mcd.AssertNotCalled(t, "Diff", mock.Anything, mock.Anything)
	mf.AssertNotCalled(t, "Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mcr.AssertNotCalled(t, "ResolveContents", mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestRelationsResolverError(t *testing.T) {
	mur, mrr, mcd, mf, mcr, mcp, u := newUnfolderWithMocks()

	uuidsAndDate := resolver.UuidsAndDate{
		UuidArr:      []string{firstExistingItemUuid, secondExistingItemUuid, addedItemUuid},
		LastModified: lastModified,
	}
	mur.On("Resolve", mock.Anything).
		Return(uuidsAndDate, nil)

	mrr.On("Resolve", mock.Anything, mock.Anything).
		Return(&relations.CCRelations{}, errors.New("Relations resolver error"))

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, collectionUuid, body, tid)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusInternalServerError, tid, resp)

	mur.AssertCalled(t, "Resolve",
		mock.MatchedBy(func(actualReqBody []byte) bool {
			assert.Equal(t, body, actualReqBody)
			return true
		}))

	mrr.AssertCalled(t, "Resolve",
		mock.MatchedBy(func(actualContentCollectionUUID string) bool {
			assert.Equal(t, collectionUuid, actualContentCollectionUUID)
			return true
		}),
		mock.MatchedBy(func(actualTid string) bool {
			assert.Equal(t, tid, actualTid)
			return true
		}))

	mcd.AssertNotCalled(t, "Diff", mock.Anything, mock.Anything)
	mf.AssertNotCalled(t, "Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mcr.AssertNotCalled(t, "ResolveContents", mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestForwarderError(t *testing.T) {
	mur, mrr, mcd, mf, mcr, mcp, u := newUnfolderWithMocks()

	uuidsAndDate := resolver.UuidsAndDate{
		UuidArr:      []string{firstExistingItemUuid, secondExistingItemUuid, addedItemUuid},
		LastModified: lastModified,
	}
	mur.On("Resolve", mock.Anything, mock.Anything).
		Return(uuidsAndDate, nil)

	oldRelations := relations.CCRelations{
		ContainedIn: leadArticleUuid,
		Contains:    []string{firstExistingItemUuid, secondExistingItemUuid, deletedItemUuid},
	}
	mrr.On("Resolve", mock.Anything, mock.Anything).
		Return(&oldRelations, nil)

	diffUuids := []string{addedItemUuid, deletedItemUuid}
	isDeletedMap := map[string]bool{addedItemUuid: false, deletedItemUuid: true}
	mcd.On("Diff", mock.Anything, mock.Anything).
		Return(diffUuids, isDeletedMap)

	mf.On("Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(forwarder.ForwarderResponse{}, errors.New("Forwarder error"))

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, collectionUuid, body, tid)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusInternalServerError, tid, resp)

	mur.AssertCalled(t, "Resolve",
		mock.MatchedBy(func(actualReqBody []byte) bool {
			assert.Equal(t, body, actualReqBody)
			return true
		}))

	mrr.AssertCalled(t, "Resolve",
		mock.MatchedBy(func(actualContentCollectionUUID string) bool {
			assert.Equal(t, collectionUuid, actualContentCollectionUUID)
			return true
		}),
		mock.MatchedBy(func(actualTid string) bool {
			assert.Equal(t, tid, actualTid)
			return true
		}))

	mcd.AssertCalled(t, "Diff",
		mock.MatchedBy(func(incomingCollectionUuids []string) bool {
			assert.Equal(t, uuidsAndDate.UuidArr, incomingCollectionUuids)
			return true
		}),
		mock.MatchedBy(func(oldCollectionUuids []string) bool {
			assert.Equal(t, oldRelations.Contains, oldCollectionUuids)
			return true
		}))

	mf.AssertCalled(t, "Forward",
		mock.MatchedBy(func(actualTid string) bool {
			assert.Equal(t, tid, actualTid)
			return true
		}),
		mock.MatchedBy(func(actualUuid string) bool {
			assert.Equal(t, collectionUuid, actualUuid)
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

	mcr.AssertNotCalled(t, "ResolveContents", mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestForwarderNon200Response(t *testing.T) {
	mur, mrr, mcd, mf, mcr, mcp, u := newUnfolderWithMocks()

	uuidsAndDate := resolver.UuidsAndDate{
		UuidArr:      []string{firstExistingItemUuid, secondExistingItemUuid, addedItemUuid},
		LastModified: lastModified,
	}
	mur.On("Resolve", mock.Anything, mock.Anything).
		Return(uuidsAndDate, nil)

	oldRelations := relations.CCRelations{
		ContainedIn: leadArticleUuid,
		Contains:    []string{firstExistingItemUuid, secondExistingItemUuid, deletedItemUuid},
	}
	mrr.On("Resolve", mock.Anything, mock.Anything).
		Return(&oldRelations, nil)

	diffUuids := []string{addedItemUuid, deletedItemUuid}
	isDeletedMap := map[string]bool{addedItemUuid: false, deletedItemUuid: true}
	mcd.On("Diff", mock.Anything, mock.Anything).
		Return(diffUuids, isDeletedMap)

	fwResp := forwarder.ForwarderResponse{http.StatusUnprocessableEntity, []byte(errorJson)}
	mf.On("Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(fwResp, nil)

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, collectionUuid, body, tid)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, fwResp.Status, tid, resp)

	respBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, fwResp.ResponseBody, respBody)

	mur.AssertCalled(t, "Resolve",
		mock.MatchedBy(func(actualReqBody []byte) bool {
			assert.Equal(t, body, actualReqBody)
			return true
		}))

	mrr.AssertCalled(t, "Resolve",
		mock.MatchedBy(func(actualContentCollectionUUID string) bool {
			assert.Equal(t, collectionUuid, actualContentCollectionUUID)
			return true
		}),
		mock.MatchedBy(func(actualTid string) bool {
			assert.Equal(t, tid, actualTid)
			return true
		}))

	mcd.AssertCalled(t, "Diff",
		mock.MatchedBy(func(incomingCollectionUuids []string) bool {
			assert.Equal(t, uuidsAndDate.UuidArr, incomingCollectionUuids)
			return true
		}),
		mock.MatchedBy(func(oldCollectionUuids []string) bool {
			assert.Equal(t, oldRelations.Contains, oldCollectionUuids)
			return true
		}))

	mf.AssertCalled(t, "Forward",
		mock.MatchedBy(func(actualTid string) bool {
			assert.Equal(t, tid, actualTid)
			return true
		}),
		mock.MatchedBy(func(actualUuid string) bool {
			assert.Equal(t, collectionUuid, actualUuid)
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

	mcr.AssertNotCalled(t, "ResolveContents", mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestNotWhitelistedCollectionType(t *testing.T) {
	mur, mrr, mcd, mf, mcr, mcp, u := newUnfolderWithMocks()

	uuidsAndDate := resolver.UuidsAndDate{
		UuidArr:      []string{firstExistingItemUuid, secondExistingItemUuid, addedItemUuid},
		LastModified: lastModified,
	}
	mur.On("Resolve", mock.Anything, mock.Anything).
		Return(uuidsAndDate, nil)

	oldRelations := relations.CCRelations{
		ContainedIn: leadArticleUuid,
		Contains:    []string{firstExistingItemUuid, secondExistingItemUuid, deletedItemUuid},
	}
	mrr.On("Resolve", mock.Anything, mock.Anything).
		Return(&oldRelations, nil)

	diffUuids := []string{addedItemUuid, deletedItemUuid}
	isDeletedMap := map[string]bool{addedItemUuid: false, deletedItemUuid: true}
	mcd.On("Diff", mock.Anything, mock.Anything).
		Return(diffUuids, isDeletedMap)

	mf.On("Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(forwarder.ForwarderResponse{http.StatusOK, []byte{}}, nil)

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, ignoredCollection, collectionUuid, body, tid)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusOK, tid, resp)

	mur.AssertCalled(t, "Resolve",
		mock.MatchedBy(func(actualReqBody []byte) bool {
			assert.Equal(t, body, actualReqBody)
			return true
		}))

	mrr.AssertCalled(t, "Resolve",
		mock.MatchedBy(func(actualContentCollectionUUID string) bool {
			assert.Equal(t, collectionUuid, actualContentCollectionUUID)
			return true
		}),
		mock.MatchedBy(func(actualTid string) bool {
			assert.Equal(t, tid, actualTid)
			return true
		}))

	mcd.AssertCalled(t, "Diff",
		mock.MatchedBy(func(incomingCollectionUuids []string) bool {
			assert.Equal(t, uuidsAndDate.UuidArr, incomingCollectionUuids)
			return true
		}),
		mock.MatchedBy(func(oldCollectionUuids []string) bool {
			assert.Equal(t, oldRelations.Contains, oldCollectionUuids)
			return true
		}))

	mf.AssertCalled(t, "Forward",
		mock.MatchedBy(func(actualTid string) bool {
			assert.Equal(t, tid, actualTid)
			return true
		}),
		mock.MatchedBy(func(actualUuid string) bool {
			assert.Equal(t, collectionUuid, actualUuid)
			return true
		}),
		mock.MatchedBy(func(actualCollectionType string) bool {
			assert.Equal(t, ignoredCollection, actualCollectionType)
			return true
		}),
		mock.MatchedBy(func(actualBody []byte) bool {
			assert.Equal(t, body, actualBody)
			return true
		}))

	mcr.AssertNotCalled(t, "ResolveContents", mock.Anything, mock.Anything)
	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestContentResolverError(t *testing.T) {
	mur, mrr, mcd, mf, mcr, mcp, u := newUnfolderWithMocks()

	uuidsAndDate := resolver.UuidsAndDate{
		UuidArr:      []string{firstExistingItemUuid, secondExistingItemUuid, addedItemUuid},
		LastModified: lastModified,
	}
	mur.On("Resolve", mock.Anything, mock.Anything).
		Return(uuidsAndDate, nil)

	oldRelations := relations.CCRelations{
		ContainedIn: leadArticleUuid,
		Contains:    []string{firstExistingItemUuid, secondExistingItemUuid, deletedItemUuid},
	}
	mrr.On("Resolve", mock.Anything, mock.Anything).
		Return(&oldRelations, nil)

	diffUuids := []string{addedItemUuid, deletedItemUuid}
	isDeletedMap := map[string]bool{addedItemUuid: false, deletedItemUuid: true}
	mcd.On("Diff", mock.Anything, mock.Anything).
		Return(diffUuids, isDeletedMap)

	mf.On("Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(forwarder.ForwarderResponse{http.StatusOK, []byte{}}, nil)

	mcr.On("ResolveContents", mock.Anything, mock.Anything).
		Return([]map[string]interface{}{}, errors.New("Content resolver error"))

	server := startTestServer(u)
	defer server.Close()

	tid := transactionidutils.NewTransactionID()
	body := readTestFile(t, inputFile)
	req := buildRequest(t, server.URL, whitelistedCollection, collectionUuid, body, tid)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	verifyResponse(t, http.StatusInternalServerError, tid, resp)

	mur.AssertCalled(t, "Resolve",
		mock.MatchedBy(func(actualReqBody []byte) bool {
			assert.Equal(t, body, actualReqBody)
			return true
		}))

	mrr.AssertCalled(t, "Resolve",
		mock.MatchedBy(func(actualContentCollectionUUID string) bool {
			assert.Equal(t, collectionUuid, actualContentCollectionUUID)
			return true
		}),
		mock.MatchedBy(func(actualTid string) bool {
			assert.Equal(t, tid, actualTid)
			return true
		}))

	mcd.AssertCalled(t, "Diff",
		mock.MatchedBy(func(incomingCollectionUuids []string) bool {
			assert.Equal(t, uuidsAndDate.UuidArr, incomingCollectionUuids)
			return true
		}),
		mock.MatchedBy(func(oldCollectionUuids []string) bool {
			assert.Equal(t, oldRelations.Contains, oldCollectionUuids)
			return true
		}))

	mf.AssertCalled(t, "Forward",
		mock.MatchedBy(func(actualTid string) bool {
			assert.Equal(t, tid, actualTid)
			return true
		}),
		mock.MatchedBy(func(actualUuid string) bool {
			assert.Equal(t, collectionUuid, actualUuid)
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

	mcr.AssertCalled(t, "ResolveContents",
		mock.MatchedBy(func(actualUuids []string) bool {
			expectedUuids := append(diffUuids, leadArticleUuid)
			assert.Equal(t, expectedUuids, actualUuids)
			return true
		}),
		mock.MatchedBy(func(actualTid string) bool {
			assert.Equal(t, tid, actualTid)
			return true
		}))

	mcp.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

//func TestAllOk(t *testing.T) {
//	mur, mrr, mcd, mf, mcr, mcp, u := newUnfolderWithMocks()
//
//	uuidsAndDate := resolver.UuidsAndDate{
//		UuidArr:      []string{firstExistingItemUuid, secondExistingItemUuid, addedItemUuid},
//		LastModified: lastModified,
//	}
//	mur.On("Resolve", mock.Anything, mock.Anything).
//		Return(uuidsAndDate, nil)
//
//	oldRelations := relations.CCRelations{
//		ContainedIn: leadArticleUuid,
//		Contains:    []string{firstExistingItemUuid, secondExistingItemUuid, deletedItemUuid},
//	}
//	mrr.On("Resolve", mock.Anything, mock.Anything).
//		Return(&oldRelations, nil)
//
//	diffUuids := []string{addedItemUuid, deletedItemUuid}
//	isDeletedMap := map[string]bool{addedItemUuid: false, deletedItemUuid: true}
//	mcd.On("Diff", mock.Anything, mock.Anything).
//		Return(diffUuids, isDeletedMap)
//
//	mf.On("Forward", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//		Return(forwarder.ForwarderResponse{http.StatusOK, []byte{}}, nil)
//
//	contentArr := []map[string]interface{}{
//		{addedItemUuid: addedItemUuid},
//		{deletedItemUuid: deletedItemUuid},
//		{leadArticleUuid: leadArticleUuid},
//	}
//	mcr.On("ResolveContents", mock.Anything, mock.Anything).
//		Return(contentArr, nil)
//
//	mcp.On("Send", mock.Anything, mock.Anything, mock.Anything)
//
//	server := startTestServer(u)
//	defer server.Close()
//
//	tid := transactionidutils.NewTransactionID()
//	body := readTestFile(t, inputFile)
//	req := buildRequest(t, server.URL, whitelistedCollection, collectionUuid, body, tid)
//
//	resp, err := http.DefaultClient.Do(req)
//	assert.NoError(t, err)
//	defer resp.Body.Close()
//
//	verifyResponse(t, http.StatusOK, tid, resp)
//
//	mur.AssertCalled(t, "Resolve",
//		mock.MatchedBy(func(actualReqBody []byte) bool {
//			assert.Equal(t, body, actualReqBody)
//			return true
//		}))
//
//	mrr.AssertCalled(t, "Resolve",
//		mock.MatchedBy(func(actualContentCollectionUUID string) bool {
//			assert.Equal(t, collectionUuid, actualContentCollectionUUID)
//			return true
//		}),
//		mock.MatchedBy(func(actualTid string) bool {
//			assert.Equal(t, tid, actualTid)
//			return true
//		}))
//
//	mcd.AssertCalled(t, "Diff",
//		mock.MatchedBy(func(incomingCollectionUuids []string) bool {
//			assert.Equal(t, uuidsAndDate.UuidArr, incomingCollectionUuids)
//			return true
//		}),
//		mock.MatchedBy(func(oldCollectionUuids []string) bool {
//			assert.Equal(t, oldRelations.Contains, oldCollectionUuids)
//			return true
//		}))
//
//	mf.AssertCalled(t, "Forward",
//		mock.MatchedBy(func(actualTid string) bool {
//			assert.Equal(t, tid, actualTid)
//			return true
//		}),
//		mock.MatchedBy(func(actualUuid string) bool {
//			assert.Equal(t, collectionUuid, actualUuid)
//			return true
//		}),
//		mock.MatchedBy(func(actualCollectionType string) bool {
//			assert.Equal(t, whitelistedCollection, actualCollectionType)
//			return true
//		}),
//		mock.MatchedBy(func(actualBody []byte) bool {
//			assert.Equal(t, body, actualBody)
//			return true
//		}))
//
//	mcr.AssertCalled(t, "ResolveContents",
//		mock.MatchedBy(func(actualUuids []string) bool {
//			expectedUuids := append(diffUuids, leadArticleUuid)
//			assert.Equal(t, expectedUuids, actualUuids)
//			return true
//		}),
//		mock.MatchedBy(func(actualTid string) bool {
//			assert.Equal(t, tid, actualTid)
//			return true
//		}))
//
//	mcp.AssertCalled(t, "Send",
//		mock.MatchedBy(func(actualTid string) bool {
//			assert.Equal(t, tid, actualTid)
//			return true
//		}),
//		mock.MatchedBy(func(actualLastModified string) bool {
//			assert.Equal(t, uuidsAndDate.LastModified, actualLastModified)
//			return true
//		}),
//		mock.MatchedBy(func(actualContentArr []map[string]interface{}) bool {
//			assert.Equal(t, contentArr, actualContentArr)
//			return true
//		}),
//		mock.MatchedBy(func(actualIsDeletedMap map[string]bool) bool {
//			isDeletedMap[leadArticleUuid] = false
//			assert.Equal(t, isDeletedMap, actualIsDeletedMap)
//			return true
//		}))
//}

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

func (mcr *mockContentResolver) ResolveContents(uuids []string, tid string) ([]map[string]interface{}, error) {
	args := mcr.Called(uuids, tid)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

type mockContentProducer struct {
	mock.Mock
}

func (mcp *mockContentProducer) Send(tid string, lastModified string, contents []map[string]interface{}, isDeleted map[string]bool) {
	mcp.Called(tid, lastModified, contents, isDeleted)
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

func (mcd *mockCollectionsDiffer) Diff(incomingCollectionUuids []string, oldCollectionUuids []string) ([]string, map[string]bool) {
	args := mcd.Called(incomingCollectionUuids, oldCollectionUuids)
	return args.Get(0).([]string), args.Get(1).(map[string]bool)
}
