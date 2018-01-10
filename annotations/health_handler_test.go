package annotations

import (
	"errors"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGTGUnhealthyCluster(t *testing.T) {
	//create a request to pass to our handler
	req := httptest.NewRequest("GET", "/__gtg", nil)

	AnnotationsDriver = dummyService{connectivityError: errors.New("test error")}

	//create a responseRecorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(httphandlers.NewGoodToGoHandler(GoodToGo))

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)
	actual := rr.Result()

	// Series of verifications:
	assert.Equal(t, http.StatusServiceUnavailable, actual.StatusCode, "status code")
	assert.Equal(t, "no-cache", actual.Header.Get("Cache-Control"), "cache-control header")
	assert.Equal(t, "test error", rr.Body.String(), "GTG response body")
}

func TestGTGHealthyCluster(t *testing.T) {
	//create a request to pass to our handler
	req := httptest.NewRequest("GET", "/__gtg", nil)
	AnnotationsDriver = dummyService{}
	//create a responseRecorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(httphandlers.NewGoodToGoHandler(GoodToGo))

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)
	actual := rr.Result()

	// Series of verifications:
	assert.Equal(t, http.StatusOK, actual.StatusCode, "status code")
	assert.Equal(t, "no-cache", actual.Header.Get("Cache-Control"), "cache-control header")
	assert.Equal(t, "OK", rr.Body.String(), "GTG response body")
}

func TestNeo4jConnectivityChecker_Healthy(t *testing.T) {
	AnnotationsDriver = dummyService{}
	message, err := Neo4jChecker()
	assert.Equal(t, "Connectivity to neo4j is ok", message)
	assert.Equal(t, nil, err)
}

func TestNeo4jConnectivityChecker_Unhealthy(t *testing.T) {
	AnnotationsDriver = dummyService{connectivityError: errors.New("test error")}
	message, err := Neo4jChecker()
	assert.Equal(t, "Error connecting to neo4j", message)
	assert.Equal(t, "test error", err.Error())
}
