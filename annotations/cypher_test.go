package annotations

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	annrw "github.com/Financial-Times/annotations-rw-neo4j/annotations"
	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/concepts-rw-neo4j/concepts"
	"github.com/Financial-Times/content-rw-neo4j/content"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/Financial-Times/organisations-rw-neo4j/organisations"
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
	alphavilleSeriesUUID               = "747894f8-a231-4efb-805d-753635eca712"
	brandParentUUID                    = "dbb0bdae-1f0c-1a1a-b0cb-b2227cce2b54"
	brandChildUUID                     = "ff691bf8-8d92-1a1a-8326-c273400bff0b"
	brandGrandChildUUID                = "ff691bf8-8d92-2a2a-8326-c273400bff0b"
	brandCircularAUUID                 = "ff691bf8-8d92-3a3a-8326-c273400bff0b"
	brandCircularBUUID                 = "ff691bf8-8d92-4a4a-8326-c273400bff0b"
	contentWithBrandsDiffTypesUUID     = "3fc9fe3e-af8c-6a6a-961a-e5065392bb31"

	MSJConceptUUID         = "5d1510f8-2779-4b74-adab-0a5eb138fca6"
	FakebookConceptUUID    = "eac853f5-3859-4c08-8540-55e043719400"
	MetalMickeyConceptUUID = "0483bef8-5797-40b8-9b25-b12e492f63c6"
	JohnSmithConceptUUID   = "75e2f7e9-cb5e-40a5-a074-86d69fe09f69"
)

var allUUIDs = []string{contentUUID, contentWithNoAnnotationsUUID, contentWithParentAndChildBrandUUID,
	contentWithThreeLevelsOfBrandUUID, contentWithCircularBrandUUID, contentWithOnlyFTUUID, alphavilleSeriesUUID,
	brandParentUUID, brandChildUUID, brandGrandChildUUID, brandCircularAUUID, brandCircularBUUID, contentWithBrandsDiffTypesUUID,
	FakebookConceptUUID, MSJConceptUUID, MetalMickeyConceptUUID, JohnSmithConceptUUID}

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

func TestCypherQueries(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnection(t, assert)
	writeAllDataToDB(t, db)
	defer cleanDB(t, db)

	t.Run("RetrieveMultipleAnnotations", func(t *testing.T) {
		expectedAnnotations := annotations{
			getExpectedFakebookAnnotation(),
			getExpectedMallStreetJournalAnnotation(),
			getExpectedMetalMickeyAnnotation(),
			getExpectedAlphavilleSeriesAnnotation(),
			getExpectedBrandChildAnnotation(),
			getExpectedBrandGrandChildAnnotation(),
			getExpectedBrandParentAnnotation(),
		}

		driver := NewCypherDriver(db, "prod")
		anns := getAndCheckAnnotations(driver, contentUUID, t)
		assert.Equal(len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
		assertListContainsAll(t, anns, expectedAnnotations)

	})

	t.Run("RetrieveContentWithParentBrand", func(t *testing.T) {
		expectedAnnotations := annotations{getExpectedBrandChildAnnotation(),
			getExpectedBrandParentAnnotation(),
			getExpectedBrandGrandChildAnnotation()}

		driver := NewCypherDriver(db, "prod")
		anns := getAndCheckAnnotations(driver, contentWithParentAndChildBrandUUID, t)
		log.Info(anns)
		assert.Equal(len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
		assertListContainsAll(t, anns, expectedAnnotations)
	})

	t.Run("RetrieveContentWithGrandParentBrand", func(t *testing.T) {
		expectedAnnotations := annotations{getExpectedBrandChildAnnotation(),
			getExpectedBrandParentAnnotation(),
			getExpectedBrandGrandChildAnnotation()}

		driver := NewCypherDriver(db, "prod")
		anns := getAndCheckAnnotations(driver, contentWithThreeLevelsOfBrandUUID, t)
		assert.Equal(len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
		assertListContainsAll(t, anns, expectedAnnotations)
	})

	t.Run("RetrieveContentWithCircularBrand", func(t *testing.T) {
		expectedAnnotations := annotations{getExpectedBrandCircularAAnnotation(),
			getExpectedBrandCircularBAnnotation()}

		driver := NewCypherDriver(db, "prod")
		anns := getAndCheckAnnotations(driver, contentWithCircularBrandUUID, t)
		assert.Equal(len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
		assertListContainsAll(t, anns, expectedAnnotations)
	})

	t.Run("RetrieveContentWithJustParentBrand", func(t *testing.T) {
		expectedAnnotations := annotations{getExpectedBrandParentAnnotation()}

		driver := NewCypherDriver(db, "prod")
		anns := getAndCheckAnnotations(driver, contentWithOnlyFTUUID, t)
		assert.Equal(len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
		assertListContainsAll(t, anns, expectedAnnotations)
	})

	//Tests filtering Annotations where content is related to Brand A as isClassifiedBy and to Brand B as isPrimarilyClassifiedBy
	// and Brands A and B have a circular relation HasParent
	t.Run("RetrieveContentBrandsOfDifferentTypes", func(t *testing.T) {
		expectedAnnotations := annotations{getExpectedBrandCircularAAnnotation(),
			getExpectedBrandCircularBAnnotation()}

		driver := NewCypherDriver(db, "prod")
		anns := getAndCheckAnnotations(driver, contentWithCircularBrandUUID, t)
		assert.Equal(len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
		assertListContainsAll(t, anns, expectedAnnotations)
	})

	t.Run("RetrieveMultipleV1Annotations", func(t *testing.T) {
		expectedAnnotations := getExpectedV1Annotations()
		driver := NewCypherDriver(db, "prod")
		anns := getAndCheckFilteredAnnotations(driver, contentUUID, "v1", t)
		assert.Equal(len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
		assertListContainsAll(t, anns, expectedAnnotations)

		for _, ann := range anns {
			log.Info(ann)
			assert.Equal("v1", ann.PlatformVersion)
		}
	})

	t.Run("RetrieveMultipleV2Annotations", func(t *testing.T) {
		expectedAnnotations := getExpectedV2Annotations()
		driver := NewCypherDriver(db, "prod")
		anns := getAndCheckFilteredAnnotations(driver, contentUUID, "v2", t)
		assert.Equal(len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
		assertListContainsAll(t, anns, expectedAnnotations)

		for _, ann := range anns {
			log.Info(ann)
			assert.Equal("v2", ann.PlatformVersion)
		}
	})

}

func TestRetrieveNoAnnotationsWhenThereAreNonePresentExceptBrands(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnection(t, assert)

	writeContent(t, db)
	writeBrands(t, db)

	defer cleanDB(t, db)

	driver := NewCypherDriver(db, "prod")
	anns, found, err := driver.read(contentWithNoAnnotationsUUID)
	assert.NoError(err, "Unexpected error for content %s", contentWithNoAnnotationsUUID)
	assert.False(found, "Found annotations for content %s", contentWithNoAnnotationsUUID)
	assert.Equal(0, len(anns), "Didn't get the same number of annotations") // Two brands, child and parent
}

func TestRetrieveNoAnnotationsWhenThereAreNoConceptsPresent(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnection(t, assert)

	writeContent(t, db)
	writeV1Annotations(t, db)
	writeV2Annotations(t, db)

	defer cleanDB(t, db)

	driver := NewCypherDriver(db, "prod")
	anns, found, err := driver.read(contentUUID)
	assert.NoError(err, "Unexpected error for content %s", contentUUID)
	assert.False(found, "Found annotations for content %s", contentUUID)
	assert.Equal(0, len(anns), "Didn't get the same number of annotations, anns=%s", anns)
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

// Utility functions
func writeAllDataToDB(t testing.TB, db neoutils.NeoConnection) {
	writeBrands(t, db)
	writeContent(t, db)
	writeOrganisations(t, db)
	writeSubjects(t, db)
	writeAlphavilleSeries(t, db)
	writeV1Annotations(t, db)
	writeV2Annotations(t, db)
}

func writeBrands(t testing.TB, db neoutils.NeoConnection) baseftrwapp.Service {
	brandRW := concepts.NewConceptService(db)
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
	writeJSONToService(contentRW, "./fixtures/Content-3fc9fe3e-af8c-6a6a-961a-e5065392bb31.json", t)
	return contentRW
}

func writeOrganisations(t testing.TB, db neoutils.NeoConnection) baseftrwapp.Service {
	organisationRW := organisations.NewCypherOrganisationService(db)
	assert.NoError(t, organisationRW.Initialise())
	writeJSONToService(organisationRW, "./fixtures/Organisation-MSJ-5d1510f8-2779-4b74-adab-0a5eb138fca6.json", t)
	writeJSONToService(organisationRW, "./fixtures/Organisation-Fakebook-eac853f5-3859-4c08-8540-55e043719400.json", t)
	return organisationRW
}

func writeSubjects(t testing.TB, db neoutils.NeoConnection) baseftrwapp.Service {
	subjectsRW := concepts.NewConceptService(db)
	assert.NoError(t, subjectsRW.Initialise())
	writeJSONToService(subjectsRW, "./fixtures/Subject-MetalMickey-0483bef8-5797-40b8-9b25-b12e492f63c6.json", t)
	return subjectsRW
}

func writeAlphavilleSeries(t testing.TB, db neoutils.NeoConnection) baseftrwapp.Service {
	alphavilleSeriesRW := concepts.NewConceptService(db)
	assert.NoError(t, alphavilleSeriesRW.Initialise())
	writeJSONToService(alphavilleSeriesRW, "./fixtures/TestAlphavilleSeries.json", t)
	return alphavilleSeriesRW
}

func writeV1Annotations(t testing.TB, db neoutils.NeoConnection) annrw.Service {
	service := annrw.NewCypherAnnotationsService(db, "v1", "annotations-v1")
	assert.NoError(t, service.Initialise())
	writeJSONToAnnotationsService(service, contentUUID, "./fixtures/Annotations-3fc9fe3e-af8c-4f7f-961a-e5065392bb31-v1.json", t)
	writeJSONToAnnotationsService(service, contentWithParentAndChildBrandUUID, "./fixtures/Annotations-3fc9fe3e-af8c-2a2a-961a-e5065392bb31-v1.json", t)
	writeJSONToAnnotationsService(service, contentWithThreeLevelsOfBrandUUID, "./fixtures/Annotations-3fc9fe3e-af8c-3a3a-961a-e5065392bb31-v1.json", t)
	writeJSONToAnnotationsService(service, contentWithCircularBrandUUID, "./fixtures/Annotations-3fc9fe3e-af8c-4a4a-961a-e5065392bb31-v1.json", t)
	writeJSONToAnnotationsService(service, contentWithOnlyFTUUID, "./fixtures/Annotations-3fc9fe3e-af8c-5a5a-961a-e5065392bb31-v1.json", t)
	writeJSONToAnnotationsService(service, contentWithBrandsDiffTypesUUID, "./fixtures/Annotations-3fc9fe3e-af8c-6a6a-961a-e5065392bb31-v1.json", t)
	return service
}

func writeV2Annotations(t testing.TB, db neoutils.NeoConnection) annrw.Service {
	service := annrw.NewCypherAnnotationsService(db, "v2", "annotations-v2")
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
	errrr := service.Write(inst, "TEST_TRANS_ID")
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

func cleanDB(t testing.TB, db neoutils.NeoConnection) {
	qs := make([]*neoism.CypherQuery, len(allUUIDs))
	for i, uuid := range allUUIDs {
		qs[i] = &neoism.CypherQuery{
			Statement: fmt.Sprintf(`
			MATCH (a:Thing {uuid: "%s"})
			OPTIONAL MATCH (a)<-[iden:IDENTIFIES]-(i:Identifier)
			OPTIONAL MATCH (a)-[:EQUIVALENT_TO]-(t:Thing)
			DELETE iden, i, t
			DETACH DELETE a`, uuid)}
	}
	err := db.CypherBatch(qs)
	assert.NoError(t, err)
}

func getExpectedV1Annotations() annotations {

	av := getExpectedAlphavilleSeriesAnnotation()
	av.TmeIDs = []string{"FOOBAR"}
	av.UUIDs = []string{"747894f8-a231-4efb-805d-753635eca712"}
	av.PlatformVersion = "v1"

	mm := getExpectedMetalMickeyAnnotation()
	mm.TmeIDs = []string{"TWV0YWwgTWlja2V5-U3ViamVjdHM="}
	mm.UUIDs = []string{"0483bef8-5797-40b8-9b25-b12e492f63c6"}
	mm.PlatformVersion = "v1"

	b := getExpectedBrandGrandChildAnnotation()
	b.TmeIDs = []string{"MTVkNjNmNzctOTA3Mi00GrandChildUtMmI4MGIyODRiNmI0-QnJhbmRz"}
	b.UUIDs = []string{"ff691bf8-8d92-2a2a-8326-c273400bff0b"}
	b.PlatformVersion = "v1"

	return []annotation{av, mm, b}
}

func getExpectedV2Annotations() annotations {

	fb := getExpectedFakebookAnnotation()
	fb.FactsetIDs = []string{"00AAA-E"}
	fb.UUIDs = []string{"eac853f5-3859-4c08-8540-55e043719400"}
	fb.PlatformVersion = "v2"

	msj := getExpectedMallStreetJournalAnnotation()
	msj.FactsetIDs = []string{"00BBBB-E"}
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
