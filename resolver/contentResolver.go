package resolver

import (
	"net/http"
	"io/ioutil"
	log "github.com/Sirupsen/logrus"
	"encoding/json"
	"github.com/Financial-Times/transactionid-utils-go"
	"fmt"
)

type contentResolver struct {
	contentResolverAppName string
	contentResolverAppURI string
	httpClient    *http.Client
}

func (cr contentResolver) ResolveContents(uuids []string, tid string) ([]map[string]interface{}, error) {
	resp, err := cr.callContentResolverApp(uuids, tid)
	if err != nil {
		log.Warnf("Error calling [%v] for content, error was: [%v]", cr.contentResolverAppName, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	bodyAsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("Could not read response from [%v], transaction_id=[%v], error was: [%v]", cr.contentResolverAppName, tid, err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Warnf("Call to [%v] for transaction_id=[%v], responded with error statusCode [%d], error was: [%v]", cr.contentResolverAppName, tid, resp.StatusCode, string(bodyAsBytes))
		return nil, err
	}

	var contents []map[string]interface{}
	err = json.Unmarshal(bodyAsBytes, contents)
	if err != nil {
		log.Warnf("Could not read response body from [%v], transaction_id=[%v], error was: [%v]", cr.contentResolverAppName, tid, err.Error())
		return nil, err
	}

	return contents, nil
}

func (cr contentResolver) callContentResolverApp(uuids []string, tid string) (*http.Response, error) {
	req, err := http.NewRequest("GET", cr.contentResolverAppURI, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request to [%v] with uri=[%v], transaction_id=[%v].", cr.contentResolverAppName, cr.contentResolverAppURI, tid)
	}

	req.Header.Set(transactionidutils.TransactionIDHeader, tid)
	req.Header.Set("Content-Type", "application/json")

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
