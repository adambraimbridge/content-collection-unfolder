package main

import (
	"encoding/json"
	fw "github.com/Financial-Times/content-collection-unfolder/forwarder"
	prod "github.com/Financial-Times/content-collection-unfolder/producer"
	res "github.com/Financial-Times/content-collection-unfolder/resolver"
	"github.com/Financial-Times/transactionid-utils-go"
	"github.com/Financial-Times/uuid-utils-go"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

const (
	unfolderPath = "/content-collection/{collectionType}/{uuid}"
)

type unfolder struct {
	forwarder       fw.Forwarder
	uuidsAndDateRes res.UuidsAndDateResolver
	contentRes      res.ContentResolver
	producer        prod.ContentProducer
	whitelist       map[string]struct{}
}

func newUnfolder(forwarder fw.Forwarder,
	uuidsAndDateRes res.UuidsAndDateResolver,
	contentRes res.ContentResolver,
	producer prod.ContentProducer,
	whitelist []string) *unfolder {

	u := unfolder{
		forwarder:       forwarder,
		uuidsAndDateRes: uuidsAndDateRes,
		contentRes:      contentRes,
		producer:        producer,
		whitelist:       map[string]struct{}{},
	}

	for _, val := range whitelist {
		u.whitelist[val] = struct{}{}
	}

	return &u
}

func (u *unfolder) handle(writer http.ResponseWriter, req *http.Request) {
	tid := transactionidutils.GetTransactionIDFromRequest(req)
	uuid, collectionType := extractPathVariables(req)

	logEntry := log.WithFields(log.Fields{
		"tid":            tid,
		"uuid":           uuid,
		"collectionType": collectionType,
	})

	writer.Header().Add(transactionidutils.TransactionIDHeader, tid)
	writer.Header().Add("Content-Type", "application/json;charset=utf-8")

	if err := uuidutils.ValidateUUID(uuid); err != nil {
		logEntry.Errorf("Invalid uuid in request path: %v", err)
		writeError(writer, http.StatusBadRequest, err)
		return
	}

	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logEntry.Errorf("Unable to extract request body: %v", err)
		writeError(writer, http.StatusUnprocessableEntity, err)
		return
	}

	logEntry.Info("Forwarding request to writer")
	fwResp, err := u.forwarder.Forward(tid, uuid, collectionType, body)
	if err != nil {
		logEntry.Errorf("Error during forwarding: %v", err)
		writeError(writer, http.StatusInternalServerError, err)
		return
	}

	if fwResp.Status != http.StatusOK {
		logEntry.Warnf("Skip unfolding. Writer returned status [%v]", fwResp.Status)
		writeResponse(writer, fwResp.Status, fwResp.ResponseBody)
		return
	}

	if _, ok := u.whitelist[collectionType]; !ok {
		logEntry.Infof("Skip unfolding. Collection type [%v] not in unfolding whitelist", collectionType)
		writeResponse(writer, fwResp.Status, fwResp.ResponseBody)
		return
	}

	uuidsAndDate, err := u.uuidsAndDateRes.Resolve(body, fwResp.ResponseBody)
	if err != nil {
		logEntry.Errorf("Error while resolving UUIDs: %v", err)
		writeError(writer, http.StatusBadRequest, err)
		return
	}

	logEntry.Infof("Resolving contents for following UUIDs: %v", uuidsAndDate.UuidArr)
	contentArr, err := u.contentRes.ResolveContents(uuidsAndDate.UuidArr, tid)
	if err != nil {
		logEntry.Errorf("Error while resolving Contents: %v", err)
		writeError(writer, http.StatusInternalServerError, err)
		return
	}

	logEntry.Info("Producing Kafka messages for resolved contents")
	u.producer.Send(tid, uuidsAndDate.LastModified, contentArr)
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
		log.Errorf("Error during json marshalling of response: %v", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeResponse(writer, status, jsonResp)
}

func writeResponse(writer http.ResponseWriter, status int, data []byte) {
	writer.WriteHeader(status)
	_, err := writer.Write(data)
	if err != nil {
		log.Errorf("Error writing response: %v", err)
	}
}
