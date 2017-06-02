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
	unfolderPath = "/unfold/{collectionType}/{uuid}"
)

type unfolder struct {
	forwarder       fw.Forwarder
	uuidsAndDateRes res.UuidsAndDateResolver
	contentRes      res.ContentResolver
	producer        prod.ContentProducer
}

func newUnfolder(forwarder fw.Forwarder, uuidsAndDateRes res.UuidsAndDateResolver, contentRes res.ContentResolver, producer prod.ContentProducer) *unfolder {
	return &unfolder{
		forwarder:       forwarder,
		uuidsAndDateRes: uuidsAndDateRes,
		contentRes:      contentRes,
		producer:        producer,
	}
}

func (u *unfolder) handle(writer http.ResponseWriter, req *http.Request) {
	tid := transactionidutils.GetTransactionIDFromRequest(req)
	uuid, collectionType := u.extractPathVariables(req)

	logEntry := log.WithFields(log.Fields{
		"tid":            tid,
		"uuid":           uuid,
		"collectionType": collectionType,
	})

	writer.Header().Add(transactionidutils.TransactionIDHeader, tid)
	writer.Header().Add("Content-Type", "application/json;charset=utf-8")

	if err := uuidutils.ValidateUUID(uuid); err != nil {
		logEntry.Errorf("Invalid uuid in request path: %v", err)
		u.writeError(writer, http.StatusBadRequest, err)
		return
	}

	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logEntry.Errorf("Unable to extract request body: %v", err)
		u.writeError(writer, http.StatusUnprocessableEntity, err)
		return
	}

	logEntry.Info("Forwarding request to writer")
	fwResp, err := u.forwarder.Forward(tid, uuid, collectionType, body)
	if err != nil {
		logEntry.Errorf("Error during forwarding: %v", err)
		u.writeError(writer, http.StatusInternalServerError, err)
		return
	}

	if fwResp.Status != http.StatusOK {
		logEntry.Warnf("Forwarder received status %v", fwResp.Status)
		u.writeResponse(writer, fwResp.Status, fwResp.ResponseBody)
		return
	}

	uuidsAndDate, err := u.uuidsAndDateRes.Resolve(body, fwResp.ResponseBody)
	if err != nil {
		logEntry.Errorf("Error while resolving UUIDs: %v", err)
		u.writeError(writer, http.StatusBadRequest, err)
		return
	}

	contentArr, err := u.contentRes.ResolveContents(uuidsAndDate.UuidArr, tid)
	if err != nil {
		logEntry.Errorf("Error while resolving Contents: %v", err)
		u.writeError(writer, http.StatusInternalServerError, err)
		return
	}

	u.producer.Send(tid, uuidsAndDate.LastModified, contentArr)
}

func (u *unfolder) extractPathVariables(req *http.Request) (string, string) {
	vars := mux.Vars(req)
	uuid := vars["uuid"]
	collectionType := vars["collectionType"]
	return uuid, collectionType
}

func (u *unfolder) writeError(writer http.ResponseWriter, status int, err error) {
	u.writeMap(writer, status, map[string]interface{}{"message": err.Error()})
}

func (u *unfolder) writeMap(writer http.ResponseWriter, status int, resp map[string]interface{}) {
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Errorf("Error during json marshalling of response: %v", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	u.writeResponse(writer, status, jsonResp)
}

func (u *unfolder) writeResponse(writer http.ResponseWriter, status int, data []byte) {
	writer.WriteHeader(status)
	_, err := writer.Write(data)
	if err != nil {
		log.Errorf("Error writing response: %v", err)
	}
}
