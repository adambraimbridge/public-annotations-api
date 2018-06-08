package annotations

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	knownUUID       = "12345"
	platformVersion = "v1"
)

type test struct {
	name         string
	req          *http.Request
	dummyService driver
	statusCode   int
	contentType  string // Contents of the Content-Type header
	body         string
}

func TestGetHandler(t *testing.T) {
	tests := []test{
		{"Success", newRequest("GET", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", nil), dummyService{contentUUID: knownUUID}, http.StatusOK, "", "[]"},
		{"NotFound", newRequest("GET", fmt.Sprintf("/content/%s/annotations", "99999"), "application/json", nil), dummyService{contentUUID: knownUUID}, http.StatusNotFound, "", message("No annotations found for content with uuid 99999.")},
		{"ReadError", newRequest("GET", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", nil), dummyService{contentUUID: knownUUID, failRead: true}, http.StatusServiceUnavailable, "", message("Error getting annotations for content with uuid 12345, err=TEST failing to READ")},
	}

	for _, test := range tests {
		AnnotationsDriver = test.dummyService
		rec := httptest.NewRecorder()
		r := mux.NewRouter()
		r.HandleFunc("/content/{uuid}/annotations", GetAnnotations).Methods("GET")
		r.ServeHTTP(rec, test.req)
		assert.True(t, test.statusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.statusCode))
		assert.JSONEq(t, test.body, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
	}
}

func TestMethodeNotFound(t *testing.T) {
	tests := []test{
		{"NotFound", newRequest("GET", fmt.Sprintf("/content/%s/annotations/", knownUUID), "application/json", nil), dummyService{contentUUID: knownUUID}, http.StatusNotFound, "", "404 page not found\n"},
	}

	for _, test := range tests {
		AnnotationsDriver = test.dummyService
		rec := httptest.NewRecorder()
		r := mux.NewRouter()
		r.HandleFunc("/content/{uuid}/annotations", GetAnnotations).Methods("GET")
		r.ServeHTTP(rec, test.req)
		assert.True(t, test.statusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.statusCode))
		assert.Equal(t, test.body, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
	}
}

func newRequest(method, url, contentType string, body []byte) *http.Request {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", contentType)
	return req
}

func message(errMsg string) string {
	return fmt.Sprintf("{\"message\": \"%s\"}\n", errMsg)
}

type dummyService struct {
	contentUUID       string
	failRead          bool
	connectivityError error
}

func (dS dummyService) read(contentUUID string) (annotations, bool, error) {
	if dS.failRead {
		return nil, false, errors.New("TEST failing to READ")
	}
	if contentUUID == dS.contentUUID {
		return []annotation{}, true, nil
	}
	return nil, false, nil
}

func (dS dummyService) checkConnectivity() error {
	return dS.connectivityError
}
