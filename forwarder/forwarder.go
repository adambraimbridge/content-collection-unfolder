package forwarder

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Financial-Times/transactionid-utils-go"
)

type Forwarder interface {
	Forward(tid string, uuid string, collectionType string, reqBody []byte) (ForwarderResponse, error)
}

type ForwarderResponse struct {
	Status       int
	ResponseBody []byte
}

type defaultForwarder struct {
	client    *http.Client
	writerUri string
}

func NewForwarder(client *http.Client, writerUri string) Forwarder {
	return &defaultForwarder{
		client:    client,
		writerUri: strings.TrimSuffix(writerUri, "/"),
	}
}

func (f *defaultForwarder) Forward(tid string, uuid string, collectionType string, reqBody []byte) (ForwarderResponse, error) {
	req, err := http.NewRequest(http.MethodPut, f.buildUrl(collectionType, uuid), bytes.NewBuffer(reqBody))
	if err != nil {
		return ForwarderResponse{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", "UPP content-collection-unfolder")
	req.Header.Add(transactionidutils.TransactionIDHeader, tid)

	resp, err := f.client.Do(req)
	if err != nil {
		return ForwarderResponse{}, err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ForwarderResponse{}, err
	}

	return ForwarderResponse{resp.StatusCode, respBody}, nil
}

func (f *defaultForwarder) buildUrl(collectionType string, uuid string) string {
	return fmt.Sprintf("%s/%s/%s", f.writerUri, collectionType, uuid)
}
