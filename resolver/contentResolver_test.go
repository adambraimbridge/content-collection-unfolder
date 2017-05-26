package resolver

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
)

var contentResolverAppMock *httptest.Server

func workingResolverAppHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("test-resources/document-store-api-output.json")
	if err != nil {
		return
	}
	defer file.Close()
	io.Copy(w, file)
}

func notWorkingResolverAppHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func mockContentResolverApp(appStatus string) {
	router := mux.NewRouter()
	var contentResolverEndpointHandler http.HandlerFunc

	if appStatus == "working" {
		contentResolverEndpointHandler = workingResolverAppHandler
	} else if appStatus == "notWorking" {
		contentResolverEndpointHandler = notWorkingResolverAppHandler
	}

	router.Path("/content").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(contentResolverEndpointHandler)})

	contentResolverAppMock = httptest.NewServer(router)
}

func getUUIDsParamsEncoded(uuids []string) string {
	httpQuery := url.Values{}
	for _, currentUUID := range uuids {
		httpQuery.Add("uuid", currentUUID)
	}
	return httpQuery.Encode()
}

//func Test_callContentResolverApp_Successful(t *testing.T) {
//	mockContentResolverApp("healthy")
//
//	resp, err := http.Get(contentResolverAppMock.URL + "/content" + getUUIDsParamsEncoded([]string{"ab43b1a6-1f47-11e7-b7d3-163f5a7f229c", "70c800d8-b3e3-11e6-ba85-95d1533d9a62"}))
//	if err != nil {
//		assert.FailNow(t, "Cannot make request to content resolver.", err.Error())
//	}
//	defer resp.Body.Close()
//
//	assert.Equal(t, http.StatusOK, resp.StatusCode, "Response status should be 200")
//	var contents []map[string]interface{}
//	json.NewDecoder(resp.Body).Decode(&contents)
//
//	assert.Equal(t, len(contents), 2, "There should be 2 contents retrieved.")
//}
