package annotations

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const (
	knownUUID = "12345"
)

func TestGetHandler(t *testing.T) {
	tests := []struct {
		name               string
		req                *http.Request
		annotationsDriver  mockDriver
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "Success",
			req:  newRequest("GET", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", nil),
			annotationsDriver: mockDriver{
				readFunc: func(string) (anns annotations, found bool, err error) {
					return []annotation{}, true, nil
				},
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "null",
		},
		{
			name: "NotFound",
			req:  newRequest("GET", fmt.Sprintf("/content/%s/annotations", "99999"), "application/json", nil),
			annotationsDriver: mockDriver{
				readFunc: func(string) (anns annotations, found bool, err error) {
					return []annotation{}, false, nil
				},
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       message("No annotations found for content with uuid 99999."),
		},
		{
			name: "ReadError",
			req:  newRequest("GET", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", nil),
			annotationsDriver: mockDriver{
				readFunc: func(string) (anns annotations, found bool, err error) {
					return nil, false, errors.New("TEST failing to READ")
				},
			},
			expectedStatusCode: http.StatusServiceUnavailable,
			expectedBody:       message("Error getting annotations for content with uuid 12345, err=TEST failing to READ"),
		},
	}

	for _, test := range tests {
		AnnotationsDriver = test.annotationsDriver
		rec := httptest.NewRecorder()
		r := mux.NewRouter()
		r.HandleFunc("/content/{uuid}/annotations", GetAnnotations).Methods("GET")
		r.ServeHTTP(rec, test.req)
		assert.True(t, test.expectedStatusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.expectedStatusCode))
		assert.JSONEq(t, test.expectedBody, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
	}
}

func TestMethodeNotFound(t *testing.T) {
	tests := []struct {
		name               string
		req                *http.Request
		annotationsDriver  mockDriver
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "NotFound",
			req:  newRequest("GET", fmt.Sprintf("/content/%s/annotations/", knownUUID), "application/json", nil),
			annotationsDriver: mockDriver{
				readFunc: func(string) (anns annotations, found bool, err error) {
					return []annotation{}, true, nil
				},
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "404 page not found\n",
		},
	}

	for _, test := range tests {
		AnnotationsDriver = test.annotationsDriver
		rec := httptest.NewRecorder()
		r := mux.NewRouter()
		r.HandleFunc("/content/{uuid}/annotations", GetAnnotations).Methods("GET")
		r.ServeHTTP(rec, test.req)
		assert.True(t, test.expectedStatusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.expectedStatusCode))
		assert.Equal(t, test.expectedBody, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
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

type mockDriver struct {
	readFunc              func(string) (annotations, bool, error)
	checkConnectivityFunc func() error
}

func (md mockDriver) read(contentUUID string) (annotations, bool, error) {
	if md.readFunc == nil {
		return nil, false, errors.New("not implemented")
	}

	return md.readFunc(contentUUID)
}

func (md mockDriver) checkConnectivity() error {
	if md.checkConnectivityFunc == nil {
		return errors.New("not implemented")
	}

	return md.checkConnectivityFunc()
}
