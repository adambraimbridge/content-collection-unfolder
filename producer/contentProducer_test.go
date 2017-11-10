package producer

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/transactionid-utils-go"
	"github.com/Financial-Times/uuid-utils-go"
	gouuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const timeFormat = "2006-01-02T15:04:05.000Z0700"

func TestHeadersAndBodyAreOk(t *testing.T) {
	mp := new(mockProducer)
	mp.On("SendMessage", mock.AnythingOfType("string"), mock.AnythingOfType("producer.Message")).Return(nil)

	cp := NewContentProducer(mp)

	tid := transactionidutils.NewTransactionID()
	lastModified := time.Now().Format(timeFormat)
	uuid := gouuid.NewV4().String()
	contentArr := map[string]interface{}{"uuid": uuid}

	cp.Send(tid, lastModified, []map[string]interface{}{contentArr}, map[string]bool{uuid: false})

	mp.AssertCalled(t, "SendMessage",
		mock.MatchedBy(func(key string) bool {
			assert.Equal(t, "", key)
			return true
		}),
		mock.MatchedBy(func(msg producer.Message) bool {
			//validate headers
			assert.Equal(t, tid, msg.Headers["X-Request-Id"])
			assert.Equal(t, lastModified, msg.Headers["Message-Timestamp"])
			assert.Equal(t, cmsContentPublished, msg.Headers["Message-Type"])
			assert.Equal(t, methodeSystemOrigin, msg.Headers["Origin-System-Id"])
			assert.Equal(t, "application/json", msg.Headers["Content-Type"])
			assert.NoError(t, uuidutils.ValidateUUID(msg.Headers["Message-Id"]))

			//validate body
			body := unmarshall(msg.Body)
			assert.Equal(t, uriBase+uuid, body["contentUri"].(string))
			assert.Equal(t, lastModified, body["lastModified"])
			assert.Equal(t, contentArr, body["payload"])

			return true
		}),
	)
	mp.AssertNumberOfCalls(t, "SendMessage", 1)
}

func TestMultipleMessagesHaveDifferentIds(t *testing.T) {
	headerIds := []string{}

	mp := new(mockProducer)
	mp.On("SendMessage",
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(msg producer.Message) bool {
			headerIds = append(headerIds, msg.Headers["Message-Id"])
			return true
		}),
	).Times(2).Return(nil)

	cp := NewContentProducer(mp)
	uuid1 := gouuid.NewV4().String()
	uuid2 := gouuid.NewV4().String()

	cp.Send(transactionidutils.NewTransactionID(),
		time.Now().Format(timeFormat),
		[]map[string]interface{}{{"uuid": uuid1}, {"uuid": uuid2}}, map[string]bool{uuid1: false, uuid2: false})

	mp.AssertNumberOfCalls(t, "SendMessage", 2)

	assert.Equal(t, 2, len(headerIds))
	assert.NotEqual(t, headerIds[0], headerIds[1])
}

func TestFailedUuidExtractionCausesSkip(t *testing.T) {
	mp := new(mockProducer)

	cp := NewContentProducer(mp)

	cp.Send(transactionidutils.NewTransactionID(),
		time.Now().Format(timeFormat),
		[]map[string]interface{}{{}, {"uuid": 123}, {"uuid": "1234"}}, map[string]bool{"123": false, "1234": false})

	mp.AssertNotCalled(t, "SendMessage", mock.AnythingOfType("string"), mock.AnythingOfType("producer.Message"))
}

func TestSendFailureDoesNotStopProducer(t *testing.T) {
	mp := new(mockProducer)
	mp.On("SendMessage", mock.AnythingOfType("string"), mock.AnythingOfType("producer.Message")).Times(4).Return(errors.New("Test error"))

	cp := NewContentProducer(mp)

	uuid1 := gouuid.NewV4().String()
	uuid2 := gouuid.NewV4().String()

	cp.Send(transactionidutils.NewTransactionID(),
		time.Now().Format(timeFormat),
		[]map[string]interface{}{{"uuid": uuid1}, {"uuid": uuid2}},
		map[string]bool{uuid1: false, uuid2: false})
	cp.Send(transactionidutils.NewTransactionID(),
		time.Now().Format(timeFormat),
		[]map[string]interface{}{{"uuid": uuid1}, {"uuid": uuid2}},
		map[string]bool{uuid1: false, uuid2: false})

	mp.AssertNumberOfCalls(t, "SendMessage", 4)
}

func TestMarshallErrorsCauseSkip(t *testing.T) {
	mp := new(mockProducer)

	cp := NewContentProducer(mp)

	uuid1 := gouuid.NewV4().String()

	cp.Send(transactionidutils.NewTransactionID(),
		time.Now().Format(timeFormat),
		[]map[string]interface{}{{"uuid": uuid1, "dude, what?": func() {}}},
		map[string]bool{uuid1: false, "dude, what?": false})

	mp.AssertNotCalled(t, "SendMessage", mock.AnythingOfType("string"), mock.AnythingOfType("producer.Message"))
}

func unmarshall(jsonString string) map[string]interface{} {
	var u map[string]interface{}
	json.Unmarshal([]byte(jsonString), &u)
	return u
}

type mockProducer struct {
	mock.Mock
}

func (mp *mockProducer) SendMessage(key string, msg producer.Message) error {
	args := mp.Called(key, msg)
	return args.Error(0)
}

func (mp *mockProducer) ConnectivityCheck() (string, error) {
	args := mp.Called()
	return args.String(0), args.Error(1)
}
