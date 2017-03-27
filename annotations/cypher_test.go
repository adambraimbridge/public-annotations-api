package annotations

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Financial-Times/alphaville-series-rw-neo4j/alphavilleseries"
	annrw "github.com/Financial-Times/annotations-rw-neo4j/annotations"
	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/brands-rw-neo4j/brands"
	"github.com/Financial-Times/content-rw-neo4j/content"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/Financial-Times/organisations-rw-neo4j/organisations"
	"github.com/Financial-Times/people-rw-neo4j/people"
	"github.com/Financial-Times/subjects-rw-neo4j/subjects"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
)

const (
	//Generate uuids so there's no clash with real data
	contentUUID                        = "3fc9fe3e-af8c-4f7f-961a-e5065392bb31"
	contentWithNoAnnotationsUUID       = "3fc9fe3e-af8c-1a1a-961a-e5065392bb31"
	contentWithParentAndChildBrandUUID = "3fc9fe3e-af8c-2a2a-961a-e5065392bb31"
	contentWithThreeLevelsOfBrandUUID  = "3fc9fe3e-af8c-3a3a-961a-e5065392bb31"
	contentWithCircularBrandUUID       = "3fc9fe3e-af8c-4a4a-961a-e5065392bb31"
	contentWithOnlyFTUUID              = "3fc9fe3e-af8c-5a5a-961a-e5065392bb31"
	MSJConceptUUID                     = "5d1510f8-2779-4b74-adab-0a5eb138fca6"
	FakebookConceptUUID                = "eac853f5-3859-4c08-8540-55e043719400"
	MetalMickeyConceptUUID             = "0483bef8-5797-40b8-9b25-b12e492f63c6"
	alphavilleSeriesUUID               = "747894f8-a231-4efb-805d-753635eca712"
	JohnSmithConceptUUID               = "75e2f7e9-cb5e-40a5-a074-86d69fe09f69"
	brandParentUUID                    = "dbb0bdae-1f0c-1a1a-b0cb-b2227cce2b54"
	brandChildUUID                     = "ff691bf8-8d92-1a1a-8326-c273400bff0b"
	brandGrandChildUUID                = "ff691bf8-8d92-2a2a-8326-c273400bff0b"
	brandCircularAUUID                 = "ff691bf8-8d92-3a3a-8326-c273400bff0b"
	brandCircularBUUID                 = "ff691bf8-8d92-4a4a-8326-c273400bff0b"
)

func TestRetrieveMultipleAnnotations(t *testing.T) {
	expectedAnnotations := annotations{getExpectedFakebookAnnotation(),
		getExpectedMallStreetJournalAnnotation(),
		getExpectedMetalMickeyAnnotation(),
		getExpectedAlphavilleSeriesAnnotation(),
		getExpectedJohnSmithAnnotation(),
		getExpectedBrandChildAnnotation(),
		getExpectedBrandParentAnnotation(),
		getExpectedBrandGrandChildAnnotation()}
	db := getDatabaseConnectionAndCheckClean(t)

	writeAllDataToDB(t, db)

	defer cleanAll(db, t)

	driver := NewCypherDriver(db, "prod")
	anns := getAndCheckAnnotations(driver, contentUUID, t)
	assert.Equal(t, len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(t, anns, expectedAnnotations)
}

func TestRetrieveMultipleV1Annotations(t *testing.T) {
	expectedAnnotations := getExpectedV1Annotations()
	db := getDatabaseConnectionAndCheckClean(t)

	writeAllDataToDB(t, db)

	defer cleanAll(db, t)

	driver := NewCypherDriver(db, "prod")
	anns := getAndCheckFilteredAnnotations(driver, contentUUID, "v1", t)
	assert.Equal(t, len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(t, anns, expectedAnnotations)

	for _, ann := range anns {
		assert.Equal(t, "v1", ann.PlatformVersion)
	}
}

func TestRetrieveMultipleV2Annotations(t *testing.T) {
	expectedAnnotations := getExpectedV2Annotations()
	db := getDatabaseConnectionAndCheckClean(t)

	writeAllDataToDB(t, db)

	defer cleanAll(db, t)

	driver := NewCypherDriver(db, "prod")
	anns := getAndCheckFilteredAnnotations(driver, contentUUID, "v2", t)
	assert.Equal(t, len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(t, anns, expectedAnnotations)

	for _, ann := range anns {
		assert.Equal(t, "v2", ann.PlatformVersion)
	}
}

func TestRetrieveContentWithParentBrand(t *testing.T) {
	expectedAnnotations := annotations{getExpectedBrandChildAnnotation(),
		getExpectedBrandParentAnnotation(),
		getExpectedBrandGrandChildAnnotation()}
	db := getDatabaseConnectionAndCheckClean(t)

	writeAllDataToDB(t, db)

	defer cleanAll(db, t)

	driver := NewCypherDriver(db, "prod")
	anns := getAndCheckAnnotations(driver, contentWithParentAndChildBrandUUID, t)
	assert.Equal(t, len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(t, anns, expectedAnnotations)
}

func TestRetrieveContentWithGrandParentBrand(t *testing.T) {
	expectedAnnotations := annotations{getExpectedBrandChildAnnotation(),
		getExpectedBrandParentAnnotation(),
		getExpectedBrandGrandChildAnnotation()}
	db := getDatabaseConnectionAndCheckClean(t)

	writeAllDataToDB(t, db)

	defer cleanAll(db, t)

	driver := NewCypherDriver(db, "prod")
	anns := getAndCheckAnnotations(driver, contentWithThreeLevelsOfBrandUUID, t)
	assert.Equal(t, len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(t, anns, expectedAnnotations)
}

func TestRetrieveContentWithCircularBrand(t *testing.T) {
	expectedAnnotations := annotations{getExpectedBrandCircularAAnnotation(),
		getExpectedBrandCircularBAnnotation()}
	db := getDatabaseConnectionAndCheckClean(t)

	writeAllDataToDB(t, db)

	defer cleanAll(db, t)

	driver := NewCypherDriver(db, "prod")
	anns := getAndCheckAnnotations(driver, contentWithCircularBrandUUID, t)
	assert.Equal(t, len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(t, anns, expectedAnnotations)
}

func TestRetrieveContentWithJustParentBrand(t *testing.T) {
	expectedAnnotations := annotations{getExpectedBrandParentAnnotation()}

	db := getDatabaseConnectionAndCheckClean(t)

	writeAllDataToDB(t, db)

	defer cleanAll(db, t)

	driver := NewCypherDriver(db, "prod")
	anns := getAndCheckAnnotations(driver, contentWithOnlyFTUUID, t)
	assert.Equal(t, len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(t, anns, expectedAnnotations)
}

func getAndCheckAnnotations(driver cypherDriver, contentUUID string, t *testing.T) annotations {
	anns, found, err := driver.read(contentUUID)
	assert.NoError(t, err, "Unexpected error for content %s", contentUUID)
	assert.True(t, found, "Found no annotations for content %s", contentUUID)
	return anns
}

func getAndCheckFilteredAnnotations(driver cypherDriver, contentUUID string, platformVersion string, t *testing.T) annotations {
	anns, found, err := driver.filteredRead(contentUUID, platformVersion)
	assert.NoError(t, err, "Unexpected error for content %s", contentUUID)
	assert.True(t, found, "Found no annotations for content %s", contentUUID)
	return anns
}

func writeAllDataToDB(t testing.TB, db neoutils.NeoConnection) {
	writeBrands(t, db)
	writeContent(t, db)
	writeOrganisations(t, db)
	writePerson(t, db)
	writeSubjects(t, db)
	writeAlphavilleSeries(t, db)
	writeV1Annotations(t, db)
	writeV2Annotations(t, db)
}

func BenchmarkRetrieveNoAnnotationsWhenThereAreNonePresent(b *testing.B) {
	db := getDatabaseConnectionAndCheckClean(b)

	writeAllDataToDB(b, db)
	defer cleanAll(db, b)

	driver := NewCypherDriver(db, "prod")
	log.Info("Running benchmark...")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		anns, found, err := driver.read(contentUUID)
		assert.NoError(b, err, "Unexpected error for content %s", contentUUID)
		assert.True(b, found, "Found no annotations for content %s", contentUUID)
		assert.Equal(b, 8, len(anns), "Didn't get the same number of annotations")
	}

	b.StopTimer()
	log.Info("... Done.")
}

func TestRetrieveNoAnnotationsWhenThereAreNonePresentExceptBrands(t *testing.T) {
	db := getDatabaseConnectionAndCheckClean(t)

	writeAllDataToDB(t, db)

	defer cleanAll(db, t)

	driver := NewCypherDriver(db, "prod")
	anns, found, err := driver.read(contentWithNoAnnotationsUUID)
	assert.NoError(t, err, "Unexpected error for content %s", contentWithNoAnnotationsUUID)
	assert.False(t, found, "Found annotations for content %s", contentWithNoAnnotationsUUID)
	assert.Equal(t, 0, len(anns), "Didn't get the same number of annotations") // Two brands, child and parent
}

func TestRetrieveNoAnnotationsWhenThereAreNoConceptsPresent(t *testing.T) {
	db := getDatabaseConnectionAndCheckClean(t)

	writeContent(t, db)
	writeV1Annotations(t, db)
	writeV2Annotations(t, db)

	defer cleanAll(db, t)

	driver := NewCypherDriver(db, "prod")
	anns, found, err := driver.read(contentUUID)
	assert.NoError(t, err, "Unexp"+
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
	assert.False(t, found, "Found annotations for content %s", contentUUID)
	assert.Equal(t, 0, len(anns), "Didn't get the same number of annotations, anns=%s", anns)
}

func writeBrands(t testing.TB, db neoutils.NeoConnection) baseftrwapp.Service {
	brandRW := brands.NewCypherBrandsService(db)
	assert.NoError(t, brandRW.Initialise())
	writeJSONToService(brandRW, "./fixtures/Brand-dbb0bdae-1f0c-1a1a-b0cb-b2227cce2b54-parent.json", t)
	writeJSONToService(brandRW, "./fixtures/Brand-ff691bf8-8d92-1a1a-8326-c273400bff0b-child.json", t)
	writeJSONToService(brandRW, "./fixtures/Brand-ff691bf8-8d92-2a2a-8326-c273400bff0b-grand_child.json", t)
	writeJSONToService(brandRW, "./fixtures/Brand-ff691bf8-8d92-3a3a-8326-c273400bff0b-circular_a.json", t)
	writeJSONToService(brandRW, "./fixtures/Brand-ff691bf8-8d92-4a4a-8326-c273400bff0b-circular_b.json", t)
	return brandRW
}

func writeContent(t testing.TB, db neoutils.NeoConnection) baseftrwapp.Service {
	contentRW := content.NewCypherContentService(db)
	assert.NoError(t, contentRW.Initialise())
	writeJSONToService(contentRW, "./fixtures/Content-3fc9fe3e-af8c-4f7f-961a-e5065392bb31.json", t)
	writeJSONToService(contentRW, "./fixtures/Content-3fc9fe3e-af8c-1a1a-961a-e5065392bb31.json", t)
	writeJSONToService(contentRW, "./fixtures/Content-3fc9fe3e-af8c-2a2a-961a-e5065392bb31.json", t)
	writeJSONToService(contentRW, "./fixtures/Content-3fc9fe3e-af8c-3a3a-961a-e5065392bb31.json", t)
	writeJSONToService(contentRW, "./fixtures/Content-3fc9fe3e-af8c-4a4a-961a-e5065392bb31.json", t)
	writeJSONToService(contentRW, "./fixtures/Content-3fc9fe3e-af8c-5a5a-961a-e5065392bb31.json", t)
	return contentRW
}

func writePerson(t testing.TB, db neoutils.NeoConnection) baseftrwapp.Service {
	personRW := people.NewCypherPeopleService(db)
	assert.NoError(t, personRW.Initialise())
	writeJSONToService(personRW, "./fixtures/People-75e2f7e9-cb5e-40a5-a074-86d69fe09f69.json", t)
	return personRW
}

func writeOrganisations(t testing.TB, db neoutils.NeoConnection) baseftrwapp.Service {
	organisationRW := organisations.NewCypherOrganisationService(db)
	assert.NoError(t, organisationRW.Initialise())
	writeJSONToService(organisationRW, "./fixtures/Organisation-MSJ-5d1510f8-2779-4b74-adab-0a5eb138fca6.json", t)
	writeJSONToService(organisationRW, "./fixtures/Organisation-Fakebook-eac853f5-3859-4c08-8540-55e043719400.json", t)
	return organisationRW
}

func writeSubjects(t testing.TB, db neoutils.NeoConnection) baseftrwapp.Service {
	subjectsRW := subjects.NewCypherSubjectsService(db)
	assert.NoError(t, subjectsRW.Initialise())
	writeJSONToService(subjectsRW, "./fixtures/Subject-MetalMickey-0483bef8-5797-40b8-9b25-b12e492f63c6.json", t)
	return subjectsRW
}

func writeAlphavilleSeries(t testing.TB, db neoutils.NeoConnection) baseftrwapp.Service {
	alphavilleSeriesRW := alphavilleseries.NewCypherAlphavilleSeriesService(db)
	assert.NoError(t, alphavilleSeriesRW.Initialise())
	writeJSONToService(alphavilleSeriesRW, "./fixtures/TestAlphavilleSeries.json", t)
	return alphavilleSeriesRW
}

func writeV1Annotations(t testing.TB, db neoutils.NeoConnection) annrw.Service {
	service := annrw.NewCypherAnnotationsService(db, "v1")
	assert.NoError(t, service.Initialise())
	writeJSONToAnnotationsService(service, contentUUID, "./fixtures/Annotations-3fc9fe3e-af8c-4f7f-961a-e5065392bb31-v1.json", t)
	writeJSONToAnnotationsService(service, contentWithParentAndChildBrandUUID, "./fixtures/Annotations-3fc9fe3e-af8c-2a2a-961a-e5065392bb31-v1.json", t)
	writeJSONToAnnotationsService(service, contentWithThreeLevelsOfBrandUUID, "./fixtures/Annotations-3fc9fe3e-af8c-3a3a-961a-e5065392bb31-v1.json", t)
	writeJSONToAnnotationsService(service, contentWithCircularBrandUUID, "./fixtures/Annotations-3fc9fe3e-af8c-4a4a-961a-e5065392bb31-v1.json", t)
	writeJSONToAnnotationsService(service, contentWithOnlyFTUUID, "./fixtures/Annotations-3fc9fe3e-af8c-5a5a-961a-e5065392bb31-v1.json", t)
	return service
}

func writeV2Annotations(t testing.TB, db neoutils.NeoConnection) annrw.Service {
	service := annrw.NewCypherAnnotationsService(db, "v2")
	assert.NoError(t, service.Initialise())
	writeJSONToAnnotationsService(service, contentUUID, "./fixtures/Annotations-3fc9fe3e-af8c-4f7f-961a-e5065392bb31-v2.json", t)
	return service
}

func writeJSONToService(service baseftrwapp.Service, pathToJSONFile string, t testing.TB) {
	absPath, _ := filepath.Abs(pathToJSONFile)
	f, err := os.Open(absPath)
	assert.NoError(t, err)
	dec := json.NewDecoder(f)
	inst, _, errr := service.DecodeJSON(dec)
	assert.NoError(t, errr)
	errrr := service.Write(inst)
	assert.NoError(t, errrr)
}

func writeJSONToAnnotationsService(service annrw.Service, contentUUID string, pathToJSONFile string, t testing.TB) {
	absPath, _ := filepath.Abs(pathToJSONFile)
	f, err := os.Open(absPath)
	assert.NoError(t, err)
	dec := json.NewDecoder(f)
	inst, errr := service.DecodeJSON(dec)
	assert.NoError(t, errr, "Error parsing file %s", pathToJSONFile)
	errrr := service.Write(contentUUID, inst)
	assert.NoError(t, errrr)
}

func assertListContainsAll(t *testing.T, list interface{}, items ...interface{}) {
	if reflect.TypeOf(items[0]).Kind().String() == "slice" {
		expected := reflect.ValueOf(items[0])
		expectedLength := expected.Len()
		assert.Len(t, list, expectedLength)
		for i := 0; i < expectedLength; i++ {
			assert.Contains(t, list, expected.Index(i).Interface())
		}
	} else {
		assert.Len(t, list, len(items))
		for _, item := range items {
			assert.Contains(t, list, item)
		}
	}
}

func getDatabaseConnectionAndCheckClean(t testing.TB) neoutils.NeoConnection {
	db := getDatabaseConnection(t)
	cleanAll(db, t)
	return db
}

func cleanAll(db neoutils.NeoConnection, t testing.TB) {
	cleanUpBrandsUppIdentifier(db, t)
	cleanDB(db, contentUUID,
		[]string{MSJConceptUUID, FakebookConceptUUID, MetalMickeyConceptUUID, alphavilleSeriesUUID, JohnSmithConceptUUID, brandGrandChildUUID, brandChildUUID, brandParentUUID, brandCircularAUUID, brandCircularBUUID, contentWithNoAnnotationsUUID, contentWithParentAndChildBrandUUID, contentWithThreeLevelsOfBrandUUID, contentWithCircularBrandUUID, contentWithOnlyFTUUID}, t)
}

func getDatabaseConnection(t testing.TB) neoutils.NeoConnection {
	db, err := getDBConn()
	assert.NoError(t, err, "Failed to connect to Neo4j")
	return db
}

func getDBConn() (neoutils.NeoConnection, error) {
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}

	conf := neoutils.DefaultConnectionConfig()
	conf.Transactional = false
	return neoutils.Connect(url, conf)
}

func cleanDB(db neoutils.NeoConnection, contentUUID string, conceptUUIDs []string, t testing.TB) {
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
	assert.NoError(t, err)
}

func cleanUpBrandsUppIdentifier(db neoutils.NeoConnection, t testing.TB) {
	qs := []*neoism.CypherQuery{
		{
			//deletes upp identifier for the above parent 'org'
			Statement: fmt.Sprintf("MATCH (i:Identifier)-[:IDENTIFIES]->(a:Thing {uuid: '%v'}) DETACH DELETE i", brandParentUUID),
		},
		{
			//deletes parent 'org' which only has type Thing
			Statement: fmt.Sprintf("MATCH (a:Thing {uuid: '%v'}) DETACH DELETE a", brandParentUUID),
		},
		{
			//deletes upp identifier for the above parent 'org'
			Statement: fmt.Sprintf("MATCH (i:Identifier)-[:IDENTIFIES]->(a:Thing {uuid: '%v'}) DETACH DELETE i", brandChildUUID),
		},
		{
			//deletes parent 'org' which only has type Thing
			Statement: fmt.Sprintf("MATCH (a:Thing {uuid: '%v'}) DETACH DELETE a", brandChildUUID),
		},
		{
			//deletes upp identifier for the above parent 'org'
			Statement: fmt.Sprintf("MATCH (i:Identifier)-[:IDENTIFIES]->(a:Thing {uuid: '%v'}) DETACH DELETE i", brandGrandChildUUID),
		},
		{
			//deletes parent 'org' which only has type Thing
			Statement: fmt.Sprintf("MATCH (a:Thing {uuid: '%v'}) DETACH DELETE a", brandGrandChildUUID),
		},
		{
			//deletes upp identifier for the above parent 'org'
			Statement: fmt.Sprintf("MATCH (i:Identifier)-[:IDENTIFIES]->(a:Thing {uuid: '%v'}) DETACH DELETE i", brandCircularAUUID),
		},
		{
			//deletes parent 'org' which only has type Thing
			Statement: fmt.Sprintf("MATCH (a:Thing {uuid: '%v'}) DETACH DELETE a", brandCircularAUUID),
		},
		{
			//deletes upp identifier for the above parent 'org'
			Statement: fmt.Sprintf("MATCH (i:Identifier)-[:IDENTIFIES]->(a:Thing {uuid: '%v'}) DETACH DELETE i", brandCircularBUUID),
		},
		{
			//deletes parent 'org' which only has type Thing
			Statement: fmt.Sprintf("MATCH (a:Thing {uuid: '%v'}) DETACH DELETE a", brandCircularBUUID),
		},
	}

	assert.NoError(t, db.CypherBatch(qs))
}

func getExpectedV1Annotations() annotations {

	js := getExpectedJohnSmithAnnotation()
	js.FactsetID = "000ZZZ-E"
	js.TmeIDs = []string{}
	js.UUIDs = []string{"75e2f7e9-cb5e-40a5-a074-86d69fe09f69"}
	js.PlatformVersion = "v1"

	av := getExpectedAlphavilleSeriesAnnotation()
	av.TmeIDs = []string{"FOOBAR"}
	av.UUIDs = []string{"747894f8-a231-4efb-805d-753635eca712"}
	av.PlatformVersion = "v1"

	mm := getExpectedMetalMickeyAnnotation()
	mm.TmeIDs = []string{"AAAA-U3ViamVjdHM="}
	mm.UUIDs = []string{"0483bef8-5797-40b8-9b25-b12e492f63c6"}
	mm.PlatformVersion = "v1"

	b := getExpectedBrandGrandChildAnnotation()
	b.TmeIDs = []string{"MTVkNjNmNzctOTA3Mi00GrandChildUtMmI4MGIyODRiNmI0-QnJhbmRz"}
	b.UUIDs = []string{"ff691bf8-8d92-2a2a-8326-c273400bff0b"}
	b.PlatformVersion = "v1"

	return []annotation{js, av, mm, b}
}

func getExpectedV2Annotations() annotations {

	fb := getExpectedFakebookAnnotation()
	fb.FactsetID = "00AAA-E"
	fb.TmeIDs = []string{}
	fb.UUIDs = []string{"eac853f5-3859-4c08-8540-55e043719400"}
	fb.PlatformVersion = "v2"

	msj := getExpectedMallStreetJournalAnnotation()
	msj.FactsetID = "00BBBB-E"
	msj.TmeIDs = []string{}
	msj.UUIDs = []string{"5d1510f8-2779-4b74-adab-0a5eb138fca6"}
	msj.PlatformVersion = "v2"

	return []annotation{fb, msj}
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

func getExpectedBrandParentAnnotation() annotation {
	return annotation{
		Predicate: "http://www.ft.com/ontology/classification/isClassifiedBy",
		ID:        "http://api.ft.com/things/" + brandParentUUID,
		APIURL:    "http://api.ft.com/brands/" + brandParentUUID,
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/classification/Classification",
			"http://www.ft.com/ontology/product/Brand",
		},
		PrefLabel: "Financial Times",
	}
}

func getExpectedBrandChildAnnotation() annotation {
	return annotation{
		Predicate: "http://www.ft.com/ontology/classification/isClassifiedBy",
		ID:        "http://api.ft.com/things/" + brandChildUUID,
		APIURL:    "http://api.ft.com/brands/" + brandChildUUID,
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/classification/Classification",
			"http://www.ft.com/ontology/product/Brand",
		},
		PrefLabel: "Business School video",
	}
}

func getExpectedBrandGrandChildAnnotation() annotation {
	return annotation{
		Predicate: "http://www.ft.com/ontology/classification/isClassifiedBy",
		ID:        "http://api.ft.com/things/" + brandGrandChildUUID,
		APIURL:    "http://api.ft.com/brands/" + brandGrandChildUUID,
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/classification/Classification",
			"http://www.ft.com/ontology/product/Brand",
		},
		PrefLabel: "Child Business School video",
	}
}

func getExpectedBrandCircularAAnnotation() annotation {
	return annotation{
		Predicate: "http://www.ft.com/ontology/classification/isClassifiedBy",
		ID:        "http://api.ft.com/things/" + brandCircularAUUID,
		APIURL:    "http://api.ft.com/brands/" + brandCircularAUUID,
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/classification/Classification",
			"http://www.ft.com/ontology/product/Brand",
		},
		PrefLabel: "Circular Business School video - A",
	}
}

func getExpectedBrandCircularBAnnotation() annotation {
	return annotation{
		Predicate: "http://www.ft.com/ontology/classification/isClassifiedBy",
		ID:        "http://api.ft.com/things/" + brandCircularBUUID,
		APIURL:    "http://api.ft.com/brands/" + brandCircularBUUID,
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/classification/Classification",
			"http://www.ft.com/ontology/product/Brand",
		},
		PrefLabel: "Circular Business School video - B",
	}
}
