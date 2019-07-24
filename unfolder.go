package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Financial-Times/content-collection-unfolder/differ"
	fw "github.com/Financial-Times/content-collection-unfolder/forwarder"
	prod "github.com/Financial-Times/content-collection-unfolder/producer"
	"github.com/Financial-Times/content-collection-unfolder/relations"
	res "github.com/Financial-Times/content-collection-unfolder/resolver"
	logger "github.com/Financial-Times/go-logger"
	"github.com/Financial-Times/transactionid-utils-go"
	"github.com/Financial-Times/uuid-utils-go"
	"github.com/Workiva/go-datastructures/set"
	"github.com/gorilla/mux"
)

const (
	unfolderPath = "/content-collection/{collectionType}/{uuid}"
)

type unfolder struct {
	uuidsAndDateRes   res.UuidsAndDateResolver
	relationsResolver relations.RelationsResolver
	collectionsDiffer differ.CollectionsDiffer
	forwarder         fw.Forwarder
	contentRes        res.ContentResolver
	producer          prod.ContentProducer
	whitelist         map[string]struct{}
}

func newUnfolder(uuidsAndDateRes res.UuidsAndDateResolver,
	relationsResolver relations.RelationsResolver,
	collectionsDiffer differ.CollectionsDiffer,
	forwarder fw.Forwarder,
	contentRes res.ContentResolver,
	producer prod.ContentProducer,
	whitelist []string) *unfolder {

	u := unfolder{
		uuidsAndDateRes:   uuidsAndDateRes,
		relationsResolver: relationsResolver,
		collectionsDiffer: collectionsDiffer,
		forwarder:         forwarder,
		contentRes:        contentRes,
		producer:          producer,
		whitelist:         map[string]struct{}{},
	}

	for _, val := range whitelist {
		u.whitelist[val] = struct{}{}
	}

	return &u
}

func (u *unfolder) handle(writer http.ResponseWriter, req *http.Request) {
	tid := transactionidutils.GetTransactionIDFromRequest(req)
	uuid, collectionType := extractPathVariables(req)

	writer.Header().Add(transactionidutils.TransactionIDHeader, tid)
	writer.Header().Add("Content-Type", "application/json;charset=utf-8")

	if err := uuidutils.ValidateUUID(uuid); err != nil {
		logger.Errorf("Message with tid=%v contentCollectionUuid=%v collectionType=%v Invalid uuid in request path: %v", tid, uuid, collectionType, err)
		writeError(writer, http.StatusBadRequest, err)
		return
	}

	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Errorf("Message with tid=%v contentCollectionUuid=%v collectionType=%v Unable to extract request body: %v", tid, uuid, collectionType, err)
		writeError(writer, http.StatusUnprocessableEntity, err)
		return
	}

	uuidsAndDate, err := u.uuidsAndDateRes.Resolve(body)
	if err != nil {
		logger.Errorf("Message with tid=%v contentCollectionUuid=%v collectionType=%v Error while resolving UUIDs: %v", tid, uuid, collectionType, err)
		writeError(writer, http.StatusBadRequest, err)
		return
	}

	oldCollectionRelations, err := u.relationsResolver.Resolve(uuid, tid)
	if err != nil {
		logger.Errorf("Message with tid=%v contentCollectionUuid=%v collectionType=%v Error while fetching old collection relations: %v", tid, uuid, collectionType, err)
		writeError(writer, http.StatusInternalServerError, err)
		return
	}

	diffUuidsSet := u.collectionsDiffer.SymmetricDifference(uuidsAndDate.UuidArr, oldCollectionRelations.Contains)

	fwResp, err := u.forwarder.Forward(tid, uuid, collectionType, body)
	if err != nil {
		logger.Errorf("Message with tid=%v contentCollectionUuid=%v collectionType=%v Error during forwarding: %v", tid, uuid, collectionType, err)
		writeError(writer, http.StatusInternalServerError, err)
		return
	}

	if fwResp.Status != http.StatusOK {
		logger.Warnf("Message with tid=%v contentCollectionUuid=%v collectionType=%v Skip unfolding. Writer returned status [%v]", tid, uuid, collectionType, fwResp.Status)
		writeResponse(writer, fwResp.Status, fwResp.ResponseBody)
		return
	}

	if _, ok := u.whitelist[collectionType]; !ok {
		logger.Infof("Message with tid=%v contentCollectionUuid=%v collectionType=%v Skip unfolding. Collection type [%v] not in unfolding whitelist", tid, uuid, collectionType, collectionType)
		writeResponse(writer, fwResp.Status, fwResp.ResponseBody)
		return
	}

	if oldCollectionRelations.ContainedIn != "" {
		diffUuidsSet.Add(oldCollectionRelations.ContainedIn)
	}

	if diffUuidsSet.Len() == 0 {
		logger.Infof("Message with tid=%v contentCollectionUuid=%v collectionType=%v Skip unfolding. No uuids to resolve after diff was done.", tid, uuid, collectionType)
		writeResponse(writer, http.StatusOK, fwResp.ResponseBody)
		return
	}

	requestTimeout := u.contentRes.GetRequestTimeout()
	resolvedContentArr, err := u.contentRes.ResolveContentsNew(flattenToStringSlice(diffUuidsSet), tid, requestTimeout)

	if err != nil {
		logger.Errorf("Message with tid=%v contentCollectionUuid=%v collectionType=%v Error while resolving contents: %v", tid, uuid, collectionType, err)
		writeError(writer, http.StatusInternalServerError, err)
		return
	}

	logger.Infof("Message with tid=%v contentCollectionUuid=%v collectionType=%v Done unfolding. Preparing to send messages.", tid, uuid, collectionType)

	u.producer.Send(tid, uuidsAndDate.LastModified, resolvedContentArr)
}

func flattenToStringSlice(set *set.Set) []string {
	stringSlice := make([]string, set.Len())
	for i, v := range set.Flatten() {
		stringSlice[i] = fmt.Sprint(v)
	}
	return stringSlice
}

func extractPathVariables(req *http.Request) (string, string) {
	vars := mux.Vars(req)
	uuid := vars["uuid"]
	collectionType := vars["collectionType"]
	return uuid, collectionType
}

func writeError(writer http.ResponseWriter, status int, err error) {
	writeMap(writer, status, map[string]interface{}{"message": err.Error()})
}

func writeMap(writer http.ResponseWriter, status int, resp map[string]interface{}) {
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		logger.Errorf("Error during json marshalling of response: %v", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeResponse(writer, status, jsonResp)
}

func writeResponse(writer http.ResponseWriter, status int, data []byte) {
	writer.WriteHeader(status)
	_, err := writer.Write(data)
	if err != nil {
		logger.Errorf("Error writing response: %v", err)
	}
}
