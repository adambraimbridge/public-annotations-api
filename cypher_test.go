package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	annrw "github.com/Financial-Times/annotations-rw-neo4j/annotations"
	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/content-rw-neo4j/content"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/Financial-Times/organisations-rw-neo4j/organisations"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
)

const (
	contentUUID                  = "d6c9c76e-a625-11e3-8a2a-00144feab7de"
	contentUUIDWithNoAnnotations = "1d76cb3c-9d18-43a8-ade4-57d99f88eac5"
)

func TestRetrieveMultipleAnnotations(t *testing.T) {
	assert := assert.New(t)
	expectedAnnotations := []annotation{getExpectedFacebookAnnotation(), getExpectedWallStreetJournalAnnotation()}
	db := getDatabaseConnectionAndCheckClean(t, assert)
	batchRunner := neoutils.NewBatchCypherRunner(neoutils.StringerDb{db}, 1)

	contentRW := writeContent(assert, db, &batchRunner)
	organisationRW := writeOrganisations(assert, db, &batchRunner)
	annotationsRW := writeAnnotations(assert, db, &batchRunner)

	defer cleanDB(db, t, assert)
	defer deleteContent(contentRW)
	defer deleteOrganisations(organisationRW)
	defer deleteAnnotations(annotationsRW)

	annotationsDriver := newCypherDriver(db, "prod")
	anns, found, err := annotationsDriver.read(contentUUID)
	assert.NoError(err, "Unexpected error for content %s", contentUUID)
	assert.True(found, "Found no annotations for content %s", contentUUID)
	assert.Equal(len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	//TODO add specific tests for both annotations
}

func TestRetrieveNoAnnotationsWhenThereAreNonePresent(t *testing.T) {
	assert := assert.New(t)
	expectedAnnotations := []annotation{}
	db := getDatabaseConnectionAndCheckClean(t, assert)
	batchRunner := neoutils.NewBatchCypherRunner(neoutils.StringerDb{db}, 1)

	contentRW := writeContent(assert, db, &batchRunner)
	organisationRW := writeOrganisations(assert, db, &batchRunner)

	defer cleanDB(db, t, assert)
	defer deleteContent(contentRW)
	defer deleteOrganisations(organisationRW)

	annotationsDriver := newCypherDriver(db, "prod")
	anns, found, err := annotationsDriver.read(contentUUIDWithNoAnnotations)
	assert.NoError(err, "Unexpected error for content %s", contentUUID)
	assert.False(found, "Found annotations for content %s", contentUUID)
	assert.Equal(len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
}

func TestRetrieveNoAnnotationsWhenThereAreNoConceptsPresent(t *testing.T) {
	assert := assert.New(t)
	expectedAnnotations := []annotation{}
	db := getDatabaseConnectionAndCheckClean(t, assert)
	batchRunner := neoutils.NewBatchCypherRunner(neoutils.StringerDb{db}, 1)

	contentRW := writeContent(assert, db, &batchRunner)
	annotationsRW := writeAnnotations(assert, db, &batchRunner)

	defer cleanDB(db, t, assert)
	defer deleteContent(contentRW)
	defer deleteAnnotations(annotationsRW)

	annotationsDriver := newCypherDriver(db, "prod")
	anns, found, err := annotationsDriver.read(contentUUIDWithNoAnnotations)
	assert.NoError(err, "Unexpected error for content %s", contentUUID)
	assert.False(found, "Found annotations for content %s", contentUUID)
	assert.Equal(len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
}

func writeContent(assert *assert.Assertions, db *neoism.Database, batchRunner *neoutils.CypherRunner) baseftrwapp.Service {
	contentRW := content.NewCypherDriver(*batchRunner, db)
	assert.NoError(contentRW.Initialise())
	writeJSONToService(contentRW, "./fixtures/Content-d6c9c76e-a625-11e3-8a2a-00144feab7de.json", assert)
	return contentRW
}

func deleteContent(contentRW baseftrwapp.Service) {
	contentRW.Delete("d6c9c76e-a625-11e3-8a2a-00144feab7de")
}

func writeOrganisations(assert *assert.Assertions, db *neoism.Database, batchRunner *neoutils.CypherRunner) baseftrwapp.Service {
	organisationRW := organisations.NewCypherOrganisationService(*batchRunner, db)
	assert.NoError(organisationRW.Initialise())
	writeJSONToService(organisationRW, "./fixtures/Organisation-WSJ-b1d71698-41b7-3754-b50e-fff60ca341b8.json", assert)
	writeJSONToService(organisationRW, "./fixtures/Organisation-Facebook-b252f5b8-e55f-343b-82a8-f23ce9cf0ee7.json", assert)
	return organisationRW
}

func deleteOrganisations(organisationRW baseftrwapp.Service) {
	organisationRW.Delete("b1d71698-41b7-3754-b50e-fff60ca341b8")
	organisationRW.Delete("b252f5b8-e55f-343b-82a8-f23ce9cf0ee7")
}

func writeAnnotations(assert *assert.Assertions, db *neoism.Database, batchRunner *neoutils.CypherRunner) annrw.Service {
	annotationsRW := annrw.NewAnnotationsService(*batchRunner, db, "v2")
	assert.NoError(annotationsRW.Initialise())
	writeJSONToAnnotationsService(annotationsRW, contentUUID, "./fixtures/Annotations-d6c9c76e-a625-11e3-8a2a-00144feab7de.json", assert)
	return annotationsRW
}

func deleteAnnotations(annotationsRW annrw.Service) {
	annotationsRW.Delete("d6c9c76e-a625-11e3-8a2a-00144feab7de")
}

func writeJSONToService(service baseftrwapp.Service, pathToJSONFile string, assert *assert.Assertions) {
	f, err := os.Open(pathToJSONFile)
	assert.NoError(err)
	dec := json.NewDecoder(f)
	inst, _, errr := service.DecodeJSON(dec)
	assert.NoError(errr)
	errrr := service.Write(inst)
	assert.NoError(errrr)
}

func writeJSONToAnnotationsService(service annrw.Service, contentUUID string, pathToJSONFile string, assert *assert.Assertions) {
	f, err := os.Open(pathToJSONFile)
	assert.NoError(err)
	dec := json.NewDecoder(f)
	inst, errr := service.DecodeJSON(dec)
	assert.NoError(errr)
	errrr := service.Write(contentUUID, inst)
	assert.NoError(errrr)
}

func assertListContainsAll(assert *assert.Assertions, list interface{}, items ...interface{}) {
	assert.Len(list, len(items))
	for _, item := range items {
		assert.Contains(list, item)
	}
}

func getDatabaseConnectionAndCheckClean(t *testing.T, assert *assert.Assertions) *neoism.Database {
	db := getDatabaseConnection(t, assert)
	cleanDB(db, t, assert)
	return db
}

func getDatabaseConnection(t *testing.T, assert *assert.Assertions) *neoism.Database {
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}

	db, err := neoism.Connect(url)
	assert.NoError(err, "Failed to connect to Neo4j")
	return db
}

func cleanDB(db *neoism.Database, t *testing.T, assert *assert.Assertions) {
	uuids := []string{
		"d6c9c76e-a625-11e3-8a2a-00144feab7de",
		"b1d71698-41b7-3754-b50e-fff60ca341b8",
		"b252f5b8-e55f-343b-82a8-f23ce9cf0ee7",
	}

	qs := make([]*neoism.CypherQuery, len(uuids))
	for i, uuid := range uuids {
		qs[i] = &neoism.CypherQuery{
			Statement: fmt.Sprintf("MATCH (a:Thing {uuid: '%s'}) DETACH DELETE a", uuid)}
	}
	err := db.CypherBatch(qs)
	assert.NoError(err)
}

func getExpectedFacebookAnnotation() annotation {
	return annotation{
		Predicate: "http://www.ft.com/ontology/annotation/mentions",
		ID:        "http://api.ft.com/things/b252f5b8-e55f-343b-82a8-f23ce9cf0ee7",
		APIURL:    "http://api.ft.com/organisations/b252f5b8-e55f-343b-82a8-f23ce9cf0ee7",
		Types: []string{
			"http://www.ft.com/ontology/organisation/Organisation",
			"http://www.ft.com/ontology/company/PublicCompany",
			"http://www.ft.com/ontology/company/Company",
		},
		LeiCode:   "BQ4BKCS1HXDV9HN80Z93",
		PrefLabel: "Facebook, Inc.",
	}
}

func getExpectedWallStreetJournalAnnotation() annotation {
	return annotation{
		Predicate: "http://www.ft.com/ontology/annotation/mentions",
		ID:        "http://api.ft.com/things/b1d71698-41b7-3754-b50e-fff60ca341b8",
		APIURL:    "http://api.ft.com/organisations/b1d71698-41b7-3754-b50e-fff60ca341b8",
		Types: []string{
			"http://www.ft.com/ontology/organisation/Organisation",
		},
		PrefLabel: "The Wall Street Journal",
	}
}
