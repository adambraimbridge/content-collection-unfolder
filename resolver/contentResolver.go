package resolver

import (
	"encoding/json"
	"fmt"
	"github.com/Financial-Times/transactionid-utils-go"
	"io/ioutil"
	"net/http"
	"regexp"
)

var serviceNameRegex = regexp.MustCompile("__(.*?)/")

type contentResolver struct {
	contentResolverAppURI  string
	contentResolverAppName string
	httpClient             *http.Client
}

func NewContentResolver(contentResolverAppURI string) *contentResolver {
	return &contentResolver{contentResolverAppURI: contentResolverAppURI, contentResolverAppName: serviceNameRegex.FindStringSubmatch(contentResolverAppURI)[1], httpClient: http.DefaultClient}
}

func (cr *contentResolver) ResolveContents(uuids []string, tid string) ([]map[string]interface{}, error) {
	resp, err := cr.callContentResolverApp(uuids, tid)
	if err != nil {
		return nil, fmt.Errorf("Error calling [%v] for content, error was: [%v]", cr.contentResolverAppName, err.Error())
	}
	defer resp.Body.Close()

	bodyAsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Could not read response from [%v], transaction_id=[%v], error was: [%v]", cr.contentResolverAppName, tid, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Call to [%v] for transaction_id=[%v], responded with error statusCode [%d], error was: [%v]", cr.contentResolverAppName, tid, resp.StatusCode, string(bodyAsBytes))
	}

	var contents []map[string]interface{}
	err = json.Unmarshal(bodyAsBytes, contents)
	if err != nil {
		return nil, fmt.Errorf("Could not read response body from [%v], transaction_id=[%v], error was: [%v]", cr.contentResolverAppName, tid, err.Error())
	}

	return contents, nil
}

func (cr *contentResolver) callContentResolverApp(uuids []string, tid string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, cr.contentResolverAppURI, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request to [%v] with uri=[%v], transaction_id=[%v].", cr.contentResolverAppName, cr.contentResolverAppURI, tid)
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
		return nil, fmt.Errorf("Error doing request to [%v] with uri=[%v], transaction_id=[%v].", cr.contentResolverAppName, cr.contentResolverAppURI, tid)
	}

	return resp, nil
}
