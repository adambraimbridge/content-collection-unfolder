package resolver

import (
	"encoding/json"
	"fmt"
	"github.com/Financial-Times/transactionid-utils-go"
	"io/ioutil"
	"net/http"
)

type contentResolver struct {
	contentResolverAppURI string
	httpClient            *http.Client
}

func NewContentResolver(contentResolverAppURI string) *contentResolver {
	return &contentResolver{contentResolverAppURI: contentResolverAppURI, httpClient: http.DefaultClient}
}

func (cr *contentResolver) ResolveContents(uuids []string, tid string) ([]map[string]interface{}, error) {
	resp, err := cr.callContentResolverApp(uuids, tid)
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
	err = json.Unmarshal(bodyAsBytes, contents)
	if err != nil {
		return nil, fmt.Errorf("Could not read response body from call to [%v], transaction_id=[%v], error was: [%v]", cr.contentResolverAppURI, tid, err.Error())
	}

	return contents, nil
}

func (cr *contentResolver) callContentResolverApp(uuids []string, tid string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, cr.contentResolverAppURI, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request to uri=[%v], transaction_id=[%v].", cr.contentResolverAppURI, tid)
	}

	req.Header.Set(transactionidutils.TransactionIDHeader, tid)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	httpQuery := req.URL.Query()
	for _, currentUUID := range uuids {
		httpQuery.Add("uuid", currentUUID)
	}
	req.URL.RawQuery = httpQuery.Encode()

	resp, err := cr.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error doing request to uri=[%v], transaction_id=[%v].", cr.contentResolverAppURI, tid)
	}

	return resp, nil
}
