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

	//mapper uri bases (not needed)
	//articleUriBase = "http://methode-article-mapper.svc.ft.com/content/"
	//placeholderUriBase = "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/content/"
	//videoUriBase = "http://next-video-mapper.svc.ft.com/video/model/"

)

type ContentProducer struct {
	msgProducer producer.MessageProducer
}

func (p ContentProducer) Send(tid string, lastModified string, contentArr []map[string]interface{}) {
	for _, content := range contentArr {
		p.sendSingleMessage(tid, content, lastModified)
	}
}

func (p ContentProducer) sendSingleMessage(tid string, content map[string]interface{}, lastModified string) {
	logEntry := log.WithField("tid", tid)
	uuid, err := p.extractUuid(content)
	if err != nil {
		logEntry.Warnf("Skip creation of kafka message. Reason: %v", err)
		return
	}

	logEntry = logEntry.WithField("uuid", uuid)
	msg, err := p.buildMessage(tid, uuid, lastModified, content)
	if err != nil {
		logEntry.Warnf("Skip creation of kafka message. Reason: %v", err)
		return
	}

	err = p.msgProducer.SendMessage("", *msg)
	if err != nil {
		logEntry.Warnf("Unable to send message to Kafka. Reason: %v", err)
	}
}

func (p ContentProducer) extractUuid(content map[string]interface{}) (string, error) {
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

func (p ContentProducer) buildMessage(tid string, uuid string, lastModified string, content map[string]interface{}) (*producer.Message, error) {
	body := publicationMessageBody{
		ContentURI:   uriBase + uuid,
		LastModified: lastModified,
		Payload:      content,
	}
	bodyAsString, err := p.marshallToString(&body)
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

func (p ContentProducer) marshallToString(body *publicationMessageBody) (*string, error) {
	binary, err := json.Marshal(*body)
	if err != nil {
		return nil, err
	}

	//not sure if needed
	//binary = bytes.Replace(binary, []byte("\\u003c"), []byte("<"), -1)
	//binary = bytes.Replace(binary, []byte("\\u003e"), []byte(">"), -1)

	binaryStr := string(binary)
	return &binaryStr, nil
}

type publicationMessageBody struct {
	ContentURI   string                 `json:"contentUri"`
	LastModified string                 `json:"lastModified"`
	Payload      map[string]interface{} `json:"payload"`
}
