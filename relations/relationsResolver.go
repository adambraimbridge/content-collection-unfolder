package relations

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Financial-Times/transactionid-utils-go"
)

type RelationsResolver interface {
	Resolve(contentCollectionUUID string, tid string) (*CCRelations, error)
}

type defaultRelationsResolver struct {
	httpClient                 *http.Client
	relationsApiPlaceholderUri string
}

func NewDefaultRelationsResolver(httpClient *http.Client, relationsApiPlaceholderUri string) *defaultRelationsResolver {
	return &defaultRelationsResolver{httpClient: httpClient, relationsApiPlaceholderUri: relationsApiPlaceholderUri}
}

func (drr *defaultRelationsResolver) Resolve(contentCollectionUUID string, tid string) (*CCRelations, error) {
	completeUri := strings.Replace(drr.relationsApiPlaceholderUri, "{uuid}", contentCollectionUUID, 1)

	resp, err := drr.callRelationsResolverApp(contentCollectionUUID, completeUri, tid)
	if err != nil {
		return nil, fmt.Errorf("Error calling on url [%v] for relations, error was: [%v]", completeUri, err.Error())
	}
	defer resp.Body.Close()

	bodyAsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Could not read response after calling [%v], transaction_id=[%v], error was: [%v]", completeUri, tid, err.Error())
	}

	if resp.StatusCode == http.StatusNotFound {
		return &CCRelations{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Call to [%v] for transaction_id=[%v], responded with error statusCode [%d], error was: [%v]", completeUri, tid, resp.StatusCode, string(bodyAsBytes))
	}

	var rel CCRelations
	err = json.Unmarshal(bodyAsBytes, &rel)
	if err != nil {
		return nil, fmt.Errorf("Could not read response body from call to [%v], transaction_id=[%v], error was: [%v]", completeUri, tid, err.Error())
	}

	return &rel, nil
}

func (drr *defaultRelationsResolver) callRelationsResolverApp(contentCollectionUUID string, completeUri string, tid string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, completeUri, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request to uri=[%v], transaction_id=[%v].", completeUri, tid)
	}

	req.Header.Set(transactionidutils.TransactionIDHeader, tid)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := drr.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error doing request to uri=[%v], transaction_id=[%v].", completeUri, tid)
	}

	return resp, nil
}

type CCRelations struct {
	ContainedIn string
	Contains    []string
}
