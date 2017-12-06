package producer

import (
	"encoding/json"
	"errors"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/uuid-utils-go"
	log "github.com/Sirupsen/logrus"
	gouuid "github.com/satori/go.uuid"
)

const (
	uriBase             = "http://content-collection-unfolder.svc.ft.com/content/"
	cmsContentPublished = "cms-content-published"
	methodeSystemOrigin = "http://cmdb.ft.com/systems/methode-web-pub"
)

type ContentProducer interface {
	Send(tid string, lastModified string, contents []map[string]interface{}, isDeleted map[string]bool)
}

type defaultContentProducer struct {
	msgProducer producer.MessageProducer
}

func NewContentProducer(msgProducer producer.MessageProducer) ContentProducer {
	return &defaultContentProducer{
		msgProducer: msgProducer,
	}
}

func (p *defaultContentProducer) Send(tid string, lastModified string, contents []map[string]interface{}, isDeleted map[string]bool) {
	for _, content := range contents {
		logEntry := log.WithField("tid", tid)
		uuid, err := extractUuid(content)
		if err != nil {
			logEntry.Warnf("Skip creation of kafka message. Reason: %v", err)
		} else {
			p.sendSingleMessage(tid, uuid, content, isDeleted[uuid], lastModified)
		}
	}
}

func (p *defaultContentProducer) sendSingleMessage(tid string, uuid string, content map[string]interface{}, deleted bool, lastModified string) {
	logEntry := log.WithField("tid", tid).WithField("uuid", uuid)
	msg, err := buildMessage(tid, uuid, lastModified, content, deleted)
	if err != nil {
		logEntry.Warnf("Skip creation of kafka message. Reason: %v", err)
		return
	}

	err = p.msgProducer.SendMessage("", *msg)
	if err != nil {
		logEntry.Warnf("Unable to send message to Kafka. Reason: %v", err)
	}
}

func extractUuid(content map[string]interface{}) (string, error) {
	val, ok := content["uuid"]
	if !ok {
		return "", errors.New("No UUID found in content")
	}

	uuid, ok := val.(string)
	if !ok {
		return "", errors.New("Found UUID was not a string")
	}

	err := uuidutils.ValidateUUID(uuid)
	if err != nil {
		return "", err
	}

	return uuid, nil
}

func buildMessage(tid string, uuid string, lastModified string, content map[string]interface{}, deleted bool) (*producer.Message, error) {
	body := publicationMessageBody{
		ContentURI:   uriBase + uuid,
		LastModified: lastModified,
	}

	if !deleted {
		body.Payload = content
	}

	bodyAsString, err := body.toJson()
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"X-Request-Id":      tid,
		"Message-Timestamp": lastModified,
		"Message-Id":        gouuid.NewV4().String(),
		"Message-Type":      cmsContentPublished,
		"Origin-System-Id":  methodeSystemOrigin,
		"Content-Type":      "application/json",
	}

	return &producer.Message{Headers: headers, Body: *bodyAsString}, nil

}

type publicationMessageBody struct {
	ContentURI   string                 `json:"contentUri"`
	LastModified string                 `json:"lastModified"`
	Payload      map[string]interface{} `json:"payload"`
}

func (body publicationMessageBody) toJson() (*string, error) {
	binary, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	binaryStr := string(binary)
	return &binaryStr, nil
}
