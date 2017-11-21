package resolver

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Financial-Times/uuid-utils-go"
)

const dateTimeFormat = "2006-01-02T15:04:05.000Z0700"

type UuidsAndDateResolver interface {
	Resolve(reqData []byte) (UuidsAndDate, error)
}

type UuidsAndDate struct {
	UuidArr      []string
	LastModified string
}

type fromRequestResolver struct {
}

func NewUuidResolver() UuidsAndDateResolver {
	return &fromRequestResolver{}
}

func (r *fromRequestResolver) Resolve(reqData []byte) (UuidsAndDate, error) {
	cc := contentCollection{}
	err := json.Unmarshal(reqData, &cc)
	if err != nil {
		return UuidsAndDate{}, fmt.Errorf("Unmarshalling error: %v", err)
	}

	uuidArr, err := resolveUuids(cc)
	if err != nil {
		return UuidsAndDate{}, err
	}

	lastModified, err := resolveLastModified(cc)
	if err != nil {
		return UuidsAndDate{}, err
	}

	return UuidsAndDate{uuidArr, lastModified}, nil
}

func resolveUuids(cc contentCollection) ([]string, error) {
	var uuidArr []string
	for _, item := range cc.Items {
		err := uuidutils.ValidateUUID(item.Uuid)
		if err != nil {
			return nil, fmt.Errorf("UUID validation error: %v", err)
		}

		uuidArr = append(uuidArr, item.Uuid)
	}

	return uuidArr, nil
}

func resolveLastModified(cc contentCollection) (string, error) {
	if _, err := time.Parse(dateTimeFormat, cc.LastModified); err != nil {
		return "", fmt.Errorf("Invalid lastModified value. Error was: %v", err)
	}

	return cc.LastModified, nil
}

type contentCollection struct {
	LastModified string                  `json:"lastModified"`
	Items        []contentCollectionItem `json:"items"`
}

type contentCollectionItem struct {
	Uuid string `json:"uuid"`
}
