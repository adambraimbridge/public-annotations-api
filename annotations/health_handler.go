package annotations

import (
	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/service-status-go/gtg"
)

const (
	deweyUrl = "https://dewey.ft.com/up-csa.html"
)

func HealthCheck() fthealth.Check {
	return fthealth.Check{
		ID:               "neo4j-cluster-health",
		BusinessImpact:   "Unable to respond to Public Annotations api requests",
		Name:             "Check connectivity to Neo4j",
		PanicGuide:       deweyUrl,
		Severity:         1,
		TechnicalSummary: `Cannot connect to Neo4j. If this check fails, check that Neo4j instance is up and running. You can find the neoUrl as a parameter for this service.`,
		Checker:          Neo4jChecker,
	}
}

func Neo4jChecker() (string, error) {
	err := AnnotationsDriver.checkConnectivity()
	if err == nil {
		return "Connectivity to neo4j is ok", err
	}
	return "Error connecting to neo4j", err
}

func GoodToGo() gtg.Status {
	statusCheck := func() gtg.Status {
		return gtgCheck(Neo4jChecker)
	}

	return gtg.FailFastParallelCheck([]gtg.StatusChecker{statusCheck})()
}

func gtgCheck(handler func() (string, error)) gtg.Status {
	if _, err := handler(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	return gtg.Status{GoodToGo: true}
}
