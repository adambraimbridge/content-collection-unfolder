package resolver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Financial-Times/transactionid-utils-go"
)

type ContentResolver interface {
	ResolveContents(diffUuids []string, tid string) ([]map[string]interface{}, error)
	ResolveContentsNew(diffUuids []string, tid string, requestTimeout time.Duration) ([]map[string]interface{}, error)
	GetRequestTimeout() time.Duration
}

type defaultContentResolver struct {
	contentResolverAppURI string
	requestTimeout        time.Duration
	httpClient            *http.Client
}

func NewContentResolver(client *http.Client, contentResolverAppURI string, requestTimeoutArg time.Duration) ContentResolver {
	return &defaultContentResolver{contentResolverAppURI: contentResolverAppURI, requestTimeout: requestTimeoutArg, httpClient: client}
}

func (cr *defaultContentResolver) ResolveContents(diffUuids []string, tid string) ([]map[string]interface{}, error) {
	resp, err := cr.callContentResolverApp(diffUuids, tid)
	if err != nil {
		return nil, fmt.Errorf("Error calling on url [%v] for content, error was: [%v]", cr.contentResolverAppURI, err.Error())
	}
	defer resp.Body.Close()

	bodyAsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Could not read response after calling [%v], transaction_id=[%v], error was: [%v]", cr.contentResolverAppURI, tid, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Call to [%v] for transaction_id=[%v], responded with error statusCode [%d], error was: [%v]", cr.contentResolverAppURI, tid, resp.StatusCode, string(bodyAsBytes))
	}

	var contents []map[string]interface{}
	err = json.Unmarshal(bodyAsBytes, &contents)
	if err != nil {
		return nil, fmt.Errorf("Could not read response body from call to [%v], transaction_id=[%v], error was: [%v]", cr.contentResolverAppURI, tid, err.Error())
	}

	return contents, nil
}

func (cr *defaultContentResolver) callContentResolverApp(diffUuids []string, tid string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, cr.contentResolverAppURI, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request to uri=[%v], transaction_id=[%v].", cr.contentResolverAppURI, tid)
	}

	req.Header.Set(transactionidutils.TransactionIDHeader, tid)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "UPP content-collection-unfolder")

	httpQuery := req.URL.Query()
	for _, key := range diffUuids {
		httpQuery.Add("uuid", key)
	}
	req.URL.RawQuery = httpQuery.Encode()

	resp, err := cr.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error doing request to uri=[%v], transaction_id=[%v].", cr.contentResolverAppURI, tid)
	}

	return resp, nil
}

func (cr *defaultContentResolver) GetRequestTimeout() time.Duration {
	return cr.requestTimeout
}

func (cr *defaultContentResolver) ResolveContentsNew(diffUuids []string, tid string, requestTimeout time.Duration) ([]map[string]interface{}, error) {
	return cr.callContentResolverAppNew(diffUuids, tid, requestTimeout)
}

func (cr *defaultContentResolver) callContentResolverAppNew(diffUuids []string, tid string, requestTimeout time.Duration) ([]map[string]interface{}, error) {
	jsonResponses := make(chan []map[string]interface{}, 1)

	req, err := cr.createRequest(tid)
	if err != nil {
		return nil, fmt.Errorf("Error calling on url [%v] for content, error was: [%v]", cr.contentResolverAppURI, err.Error())
	}
	httpQuery := req.URL.Query()
	for _, diffUuid := range diffUuids {
		httpQuery.Add("uuid", diffUuid)
		time.Sleep(requestTimeout)
	}

	req.URL.RawQuery = httpQuery.Encode()
	resp, err := cr.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error doing request to uri=[%v], transaction_id=[%v].", cr.contentResolverAppURI, tid)
	}

	bodyAsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Could not read response after calling [%v], transaction_id=[%v], error was: [%v]", cr.contentResolverAppURI, tid, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Call to [%v] for transaction_id=[%v], responded with error statusCode [%d], error was: [%v]", cr.contentResolverAppURI, tid, resp.StatusCode, string(bodyAsBytes))
	}
	defer resp.Body.Close()
	var content []map[string]interface{}
	err = json.Unmarshal(bodyAsBytes, &content)
	if err != nil {
		return nil, fmt.Errorf("Could not read response body from call to [%v], transaction_id=[%v], error was: [%v]", cr.contentResolverAppURI, tid, err.Error())
	}
	jsonResponses <- content

	return getFromResponsesChannel(jsonResponses)
}

func (cr *defaultContentResolver) createRequest(tid string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, cr.contentResolverAppURI, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request to uri=[%v], transaction_id=[%v].", cr.contentResolverAppURI, tid)
	}

	req.Header.Set(transactionidutils.TransactionIDHeader, tid)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "UPP content-collection-unfolder")
	return req, nil
}

func getFromResponsesChannel(jsonResponses chan []map[string]interface{}) ([]map[string]interface{}, error) {
	responsesCount := len(jsonResponses)
	var jsonResponsesArr []map[string]interface{}

	for i := 0; i < responsesCount; i++ {
		jsonResponsesArr = <-jsonResponses
	}

	close(jsonResponses)

	if len(jsonResponsesArr) != 0 {
		return jsonResponsesArr, nil
	} else {
		return nil, nil
	}
}
