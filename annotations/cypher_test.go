package annotations

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/Financial-Times/alphaville-series-rw-neo4j/alphavilleseries"
	annrw "github.com/Financial-Times/annotations-rw-neo4j/annotations"
	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/content-rw-neo4j/content"
	"github.com/Financial-Times/people-rw-neo4j/people"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/Financial-Times/organisations-rw-neo4j/organisations"
	"github.com/Financial-Times/subjects-rw-neo4j/subjects"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
)

const (
	//Generate uuids so there's no clash with real data
	contentUUID            = "3fc9fe3e-af8c-4f7f-961a-e5065392bb31"
	MSJConceptUUID         = "5d1510f8-2779-4b74-adab-0a5eb138fca6"
	FakebookConceptUUID    = "eac853f5-3859-4c08-8540-55e043719400"
	MetalMickeyConceptUUID = "0483bef8-5797-40b8-9b25-b12e492f63c6"
	alphavilleSeriesUUID   = "747894f8-a231-4efb-805d-753635eca712"
	JohnSmithConceptUUID = "75e2f7e9-cb5e-40a5-a074-86d69fe09f69"
)

func TestRetrieveMultipleAnnotations(t *testing.T) {
	assert := assert.New(t)
	expectedAnnotations := annotations{getExpectedFakebookAnnotation(),
		getExpectedMallStreetJournalAnnotation(),
		getExpectedMetalMickeyAnnotation(),
		getExpectedAlphavilleSeriesAnnotation(),
		getExpectedJohnSmithAnnotation()}
	db := getDatabaseConnectionAndCheckClean(t, assert)

	writeContent(assert, db)
	writeOrganisations(assert, db)
	writePerson(assert, db)
	writeSubjects(assert, db)
	writeAlphavilleSeries(assert, db)
	writeV1Annotations(assert, db)
	writeV2Annotations(assert, db)

	defer cleanDB(db, contentUUID,
		[]string{MSJConceptUUID, FakebookConceptUUID, MetalMickeyConceptUUID, alphavilleSeriesUUID, JohnSmithConceptUUID},
		t, assert)
	defer cleanUpBrandsUppIdentifier(db, t, assert)

	annotationsDriver := NewCypherDriver(db, "prod")
	anns, found, err := annotationsDriver.read(contentUUID)
	assert.NoError(err, "Unexpected error for content %s", contentUUID)
	assert.True(found, "Found no annotations for content %s", contentUUID)
	assert.Equal(len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(assert, anns, expectedAnnotations)
}

func TestRetrieveNoAnnotationsWhenThereAreNonePresent(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnectionAndCheckClean(t, assert)

	writeContent(assert, db)
	writeOrganisations(assert, db)
	writePerson(assert, db)
	writeSubjects(assert, db)

	defer cleanDB(db, contentUUID,
		[]string{MSJConceptUUID, FakebookConceptUUID, MetalMickeyConceptUUID, alphavilleSeriesUUID, JohnSmithConceptUUID},
		t, assert)
	defer cleanUpBrandsUppIdentifier(db, t, assert)

	annotationsDriver := NewCypherDriver(db, "prod")
	anns, found, err := annotationsDriver.read(contentUUID)
	assert.NoError(err, "Unexpected error for content %s", contentUUID)
	assert.False(found, "Found annotations for content %s", contentUUID)
	assert.Equal(0, len(anns), "Didn't get the same number of annotations")
}

func TestRetrieveNoAnnotationsWhenThereAreNoConceptsPresent(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnectionAndCheckClean(t, assert)

	writeContent(assert, db)
	writeV1Annotations(assert, db)
	writeV2Annotations(assert, db)

	defer cleanDB(db, contentUUID,
		[]string{MSJConceptUUID, FakebookConceptUUID, MetalMickeyConceptUUID, alphavilleSeriesUUID, JohnSmithConceptUUID},
		t, assert)
	defer cleanUpBrandsUppIdentifier(db, t, assert)

	annotationsDriver := NewCypherDriver(db, "prod")
	anns, found, err := annotationsDriver.read(contentUUID)
	assert.NoError(err, "Unexp"+
		""+
		""+
		""+
		""+
		""+
		""+
		""+
		""+
		""+
		"ected error for content %s", contentUUID)
	assert.False(found, "Found annotations for content %s", contentUUID)
	assert.Equal(0, len(anns), "Didn't get the same number of annotations, anns=%s", anns)
}

func writeContent(assert *assert.Assertions, db neoutils.NeoConnection) baseftrwapp.Service {
	contentRW := content.NewCypherContentService(db)
	assert.NoError(contentRW.Initialise())
	writeJSONToService(contentRW, "./fixtures/Content-3fc9fe3e-af8c-4f7f-961a-e5065392bb31.json", assert)
	return contentRW
}

func writePerson(assert *assert.Assertions, db neoutils.NeoConnection) baseftrwapp.Service {
	personRW := people.NewCypherPeopleService(db)
	assert.NoError(personRW.Initialise())
	writeJSONToService(personRW, "./fixtures/People-75e2f7e9-cb5e-40a5-a074-86d69fe09f69.json", assert)
	return personRW
}

func writeOrganisations(assert *assert.Assertions, db neoutils.NeoConnection) baseftrwapp.Service {
	organisationRW := organisations.NewCypherOrganisationService(db)
	assert.NoError(organisationRW.Initialise())
	writeJSONToService(organisationRW, "./fixtures/Organisation-MSJ-5d1510f8-2779-4b74-adab-0a5eb138fca6.json", assert)
	writeJSONToService(organisationRW, "./fixtures/Organisation-Fakebook-eac853f5-3859-4c08-8540-55e043719400.json", assert)
	return organisationRW
}

func writeSubjects(assert *assert.Assertions, db neoutils.NeoConnection) baseftrwapp.Service {
	subjectsRW := subjects.NewCypherSubjectsService(db)
	assert.NoError(subjectsRW.Initialise())
	writeJSONToService(subjectsRW, "./fixtures/Subject-MetalMickey-0483bef8-5797-40b8-9b25-b12e492f63c6.json", assert)
	return subjectsRW
}

func writeAlphavilleSeries(assert *assert.Assertions, db neoutils.NeoConnection) baseftrwapp.Service {
	alphavilleSeriesRW := alphavilleseries.NewCypherAlphavilleSeriesService(db)
	assert.NoError(alphavilleSeriesRW.Initialise())
	writeJSONToService(alphavilleSeriesRW, "./fixtures/TestAlphavilleSeries.json", assert)
	return alphavilleSeriesRW
}

func writeV1Annotations(assert *assert.Assertions, db neoutils.NeoConnection) annrw.Service {
	annotationsRW := annrw.NewCypherAnnotationsService(db, "v1")
	assert.NoError(annotationsRW.Initialise())
	writeJSONToAnnotationsService(annotationsRW, contentUUID, "./fixtures/Annotations-3fc9fe3e-af8c-4f7f-961a-e5065392bb31-v1.json", assert)
	return annotationsRW
}

func writeV2Annotations(assert *assert.Assertions, db neoutils.NeoConnection) annrw.Service {
	annotationsRW := annrw.NewCypherAnnotationsService(db, "v2")
	assert.NoError(annotationsRW.Initialise())
	writeJSONToAnnotationsService(annotationsRW, contentUUID, "./fixtures/Annotations-3fc9fe3e-af8c-4f7f-961a-e5065392bb31-v2.json", assert)
	return annotationsRW
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
	assert.NoError(errr, "Error parsing file %s", pathToJSONFile)
	errrr := service.Write(contentUUID, inst)
	assert.NoError(errrr)
}

func assertListContainsAll(assert *assert.Assertions, list interface{}, items ...interface{}) {
	if reflect.TypeOf(items[0]).Kind().String() == "slice" {
		expected := reflect.ValueOf(items[0])
		expectedLength := expected.Len()
		assert.Len(list, expectedLength)
		for i := 0; i < expectedLength; i++ {
			assert.Contains(list, expected.Index(i).Interface())
		}
	} else {
		assert.Len(list, len(items))
		for _, item := range items {
			assert.Contains(list, item)
		}
	}
}

func getDatabaseConnectionAndCheckClean(t *testing.T, assert *assert.Assertions) neoutils.NeoConnection {
	db := getDatabaseConnection(t, assert)
	cleanDB(db, contentUUID, []string{}, t, assert)
	return db
}

func getDatabaseConnection(t *testing.T, assert *assert.Assertions) neoutils.NeoConnection {
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}

	conf := neoutils.DefaultConnectionConfig()
	conf.Transactional = false
	db, err := neoutils.Connect(url, conf)
	assert.NoError(err, "Failed to connect to Neo4j")
	return db
}

func cleanDB(db neoutils.NeoConnection, contentUUID string, conceptUUIDs []string, t *testing.T, assert *assert.Assertions) {
	size := len(conceptUUIDs)
	if size == 0 && contentUUID == "" {
		return
	}

	uuids := make([]string, size+1)
	copy(uuids, conceptUUIDs)
	if contentUUID != "" {
		uuids[size] = contentUUID
	}

	qs := make([]*neoism.CypherQuery, len(uuids))
	for i, uuid := range uuids {
		qs[i] = &neoism.CypherQuery{
			Statement: fmt.Sprintf(`
			MATCH (a:Thing {uuid: "%s"})
			OPTIONAL MATCH (a)<-[iden:IDENTIFIES]-(i:Identifier)
			DELETE iden, i
			DETACH DELETE a`, uuid)}
	}
	err := db.CypherBatch(qs)
	assert.NoError(err)
}

func cleanUpBrandsUppIdentifier(db neoutils.NeoConnection, t *testing.T, assert *assert.Assertions) {
	qs := []*neoism.CypherQuery{
		{
			//deletes parent 'org' which only has type Thing
			Statement: fmt.Sprintf("MATCH (a:Thing {uuid: '%v'}) DETACH DELETE a", "dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"),
		},
		{
			//deletes upp identifier for the above parent 'org'
			Statement: fmt.Sprintf("MATCH (b:Identifier {value: '%v'}) DETACH DELETE b", "dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"),
		},
	}

	err := db.CypherBatch(qs)
	assert.NoError(err)
}

func getExpectedFakebookAnnotation() annotation {
	return annotation{
		Predicate: "http://www.ft.com/ontology/annotation/mentions",
		ID:        "http://api.ft.com/things/eac853f5-3859-4c08-8540-55e043719400",
		APIURL:    "http://api.ft.com/organisations/eac853f5-3859-4c08-8540-55e043719400",
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/organisation/Organisation",
			"http://www.ft.com/ontology/company/Company",
			"http://www.ft.com/ontology/company/PublicCompany",
		},
		LeiCode:   "BQ4BKCS1HXDV9TTTTTTTT",
		PrefLabel: "Fakebook, Inc.",
	}
}

func getExpectedJohnSmithAnnotation() annotation {
	return annotation{
		Predicate: "http://www.ft.com/ontology/annotation/hasAuthor",
		ID:        "http://api.ft.com/things/75e2f7e9-cb5e-40a5-a074-86d69fe09f69",
		APIURL:    "http://api.ft.com/people/75e2f7e9-cb5e-40a5-a074-86d69fe09f69",
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/person/Person",
		},
		PrefLabel: "John Smith",
	}
}

func getExpectedMallStreetJournalAnnotation() annotation {
	return annotation{
		Predicate: "http://www.ft.com/ontology/annotation/mentions",
		ID:        "http://api.ft.com/things/5d1510f8-2779-4b74-adab-0a5eb138fca6",
		APIURL:    "http://api.ft.com/organisations/5d1510f8-2779-4b74-adab-0a5eb138fca6",
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/organisation/Organisation",
		},
		PrefLabel: "The Mall Street Journal",
	}
}

func getExpectedMetalMickeyAnnotation() annotation {
	return annotation{
		Predicate: "http://www.ft.com/ontology/classification/isClassifiedBy",
		ID:        "http://api.ft.com/things/0483bef8-5797-40b8-9b25-b12e492f63c6",
		APIURL:    "http://api.ft.com/things/0483bef8-5797-40b8-9b25-b12e492f63c6",
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/classification/Classification",
			"http://www.ft.com/ontology/Subject",
		},
		PrefLabel: "Metal Mickey",
	}
}

func getExpectedAlphavilleSeriesAnnotation() annotation {
	return annotation{
		Predicate: "http://www.ft.com/ontology/classification/isClassifiedBy",
		ID:        "http://api.ft.com/things/" + alphavilleSeriesUUID,
		APIURL:    "http://api.ft.com/things/" + alphavilleSeriesUUID,
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/classification/Classification",
			"http://www.ft.com/ontology/AlphavilleSeries",
		},
		PrefLabel: "Test Alphaville Series",
	}
}
