package resolver

import (
	"encoding/json"
	"fmt"
	"github.com/Financial-Times/uuid-utils-go"
)

type UuidsAndDateResolver interface {
	Resolve(reqData []byte, respData []byte) (*UuidsAndDate, error)
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

func (r *fromRequestResolver) Resolve(reqData []byte, respData []byte) (*UuidsAndDate, error) {
	reqMap := map[string]interface{}{}
	err := json.Unmarshal(reqData, reqMap)
	if err != nil {
		return nil, fmt.Errorf("Unmarshalling error: %v", err)
	}

	uuidArr, err := r.resolveUuids(reqMap)
	if err != nil {
		return nil, err
	}

	lastModified, err := r.resolveLastModified(reqMap)
	if err != nil {
		return nil, err
	}

	return &UuidsAndDate{uuidArr, lastModified}, nil
}

func (*fromRequestResolver) resolveUuids(reqMap map[string]interface{}) ([]string, error) {
	items, ok := reqMap["items"]
	if !ok {
		return nil, fmt.Errorf("Found request with no items. Request was: %v", reqMap)
	}

	itemArr, ok := items.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Found malformed item array. Request was: %v", reqMap)
	}

	uuidArr := []string{}
	for _, item := range itemArr {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Found malformed item. Request was: %v", reqMap)
		}

		uuid, ok := itemMap["uuid"]
		if !ok {
			return nil, fmt.Errorf("Found item with missing UUID. Request was: %v", reqMap)
		}

		castUuid, ok := uuid.(string)
		if !ok {
			return nil, fmt.Errorf("Found item with malformed UUID. Request was: %v", reqMap)
		}

		err := uuidutils.ValidateUUID(castUuid)
		if err != nil {
			return nil, fmt.Errorf("UUID validation error: %v. Request was: %v", err, reqMap)
		}

		uuidArr = append(uuidArr, castUuid)
	}

	return uuidArr, nil
}

func (*fromRequestResolver) resolveLastModified(reqMap map[string]interface{}) (string, error) {
	lastModified, ok := reqMap["lastModified"]
	if !ok {
		return "", fmt.Errorf("Found request with no lastModified field. Request was: %v", reqMap)
	}

	castLastModified, ok := lastModified.(string)
	if !ok {
		return "", fmt.Errorf("Found request with malforemd lastModified field. Request was: %v", reqMap)
	}

	return castLastModified, nil
}
