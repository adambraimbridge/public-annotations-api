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
	"github.com/Financial-Times/go-logger"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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

	contentWithBrandsDiffTypesUUID = "3fc9fe3e-af8c-6a6a-961a-e5065392bb31"
	financialInstrumentUUID        = "77f613ad-1470-422c-bf7c-1dd4c3fd1693"

	MSJConceptUUID         = "5d1510f8-2779-4b74-adab-0a5eb138fca6"
	FakebookConceptUUID    = "eac853f5-3859-4c08-8540-55e043719400"
	MetalMickeyConceptUUID = "0483bef8-5797-40b8-9b25-b12e492f63c6"
	JohnSmithConceptUUID   = "75e2f7e9-cb5e-40a5-a074-86d69fe09f69"
	brokenPacUUID          = "8d965e66-5454-4856-972d-f64cc1a18a5d"

	narrowerTopic = "7e22c8b8-b280-4e52-aa22-fa1c6dffd894"
	aboutTopic    = "ca982370-66cd-43bd-b2e3-7bfcb73efb1e"
	broaderTopicA = "fde5eee9-3260-4125-adb6-3d91a4888be5"
	broaderTopicB = "b6469cc2-f6ff-45aa-a9bb-3d1bb0f9a35d"
	cyclicTopicA  = "e404e3bd-beff-4324-83f4-beb044baf916"
	cyclicTopicB  = "77a410a3-6857-4654-80ef-6aae29be852a"

	v1PlatformVersion    = "v1"
	v2PlatformVersion    = "v2"
	v1Lifecycle          = "annotations-v1"
	nextVideoLifecycle   = "annotations-next-video"
	emptyPlatformVersion = ""

	brandType = "http://www.ft.com/ontology/product/Brand"
	topicType = "http://www.ft.com/ontology/Topic"
)

var (
	conceptLabels = map[string]string{
		brandGrandChildUUID: "Child Business School video",
		brandChildUUID:      "Business School video",
		brandParentUUID:     "Financial Times",
		brandCircularAUUID:  "Circular Business School video - A",
		brandCircularBUUID:  "Circular Business School video - B",
		aboutTopic:          "Ashes 2017",
		broaderTopicA:       "The Ashes",
		broaderTopicB:       "Cricket",
		narrowerTopic:       "England Ashes 2017 Victory",
		cyclicTopicA:        "Dodgy Cyclic Topic A",
		cyclicTopicB:        "Dodgy Cyclic Topic B",
	}

	conceptTypes = map[string][]string{
		brandType: {
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/classification/Classification",
			brandType,
		},
		topicType: {
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			topicType,
		},
	}

	conceptApiUrlTemplates = map[string]string{
		brandType: "http://api.ft.com/brands/%s",
		topicType: "http://api.ft.com/things/%s",
	}
)

type cypherDriverTestSuite struct {
	suite.Suite
	db neoutils.NeoConnection
}

var allUUIDs = []string{contentUUID, contentWithNoAnnotationsUUID, contentWithParentAndChildBrandUUID,
	contentWithThreeLevelsOfBrandUUID, contentWithCircularBrandUUID, contentWithOnlyFTUUID, alphavilleSeriesUUID,
	brandParentUUID, brandChildUUID, brandGrandChildUUID, brandCircularAUUID, brandCircularBUUID, contentWithBrandsDiffTypesUUID,
	FakebookConceptUUID, MSJConceptUUID, MetalMickeyConceptUUID, brokenPacUUID, financialInstrumentUUID, JohnSmithConceptUUID,
	aboutTopic, broaderTopicA, broaderTopicB, narrowerTopic, cyclicTopicA, cyclicTopicB}

func TestCypherDriverSuite(t *testing.T) {
	logger.InitLogger("public-annotations-api-test", log.DebugLevel.String())
	suite.Run(t, newCypherDriverTestSuite())
}

func newCypherDriverTestSuite() *cypherDriverTestSuite {
	return &cypherDriverTestSuite{}
}

func (s *cypherDriverTestSuite) SetupTest() {
	s.db = getDatabaseConnection(s.T())
	writeAllDataToDB(s.T(), s.db)
}

func (s *cypherDriverTestSuite) TearDownTest() {
	cleanDB(s.T(), s.db)
}

func getDatabaseConnection(t *testing.T) neoutils.NeoConnection {
	if testing.Short() {
		t.Skip("Skipping Neo4j integration tests.")
		return nil
	}

	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}

	conf := neoutils.DefaultConnectionConfig()
	conf.Transactional = false
	db, err := neoutils.Connect(url, conf)
	require.NoError(t, err, "Failed to connect to Neo4j")
	return db
}

func (s *cypherDriverTestSuite) TestRetrieveMultipleAnnotations() {
	expectedAnnotations := annotations{
		getExpectedMentionsFakebookAnnotation(v2Lifecycle),
		getExpectedMallStreetJournalAnnotation(v2Lifecycle),
		getExpectedMetalMickeyAnnotation(v1Lifecycle),
		getExpectedAlphavilleSeriesAnnotation(v1Lifecycle),
		expectedAnnotation(brandGrandChildUUID, brandType, predicates["IS_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandChildUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandParentUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
	}

	driver := NewCypherDriver(s.db, "prod")
	anns := getAndCheckAnnotations(driver, contentUUID, s.T())
	assert.Equal(s.T(), len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestRetrievePacAndV2AnnotationsAsPriority() {
	expectedAnnotations := annotations{
		getExpectedMetalMickeyAnnotation(pacLifecycle),
		getExpectedHasDisplayTagFakebookAnnotation(pacLifecycle),
		getExpectedAboutFakebookAnnotation(pacLifecycle),
		getExpectedJohnSmithAnnotation(pacLifecycle),
		getExpectedMallStreetJournalAnnotation(v2Lifecycle),
		expectedAnnotation(brandGrandChildUUID, brandType, predicates["IS_CLASSIFIED_BY"], pacLifecycle),
		expectedAnnotation(brandChildUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], pacLifecycle),
		expectedAnnotation(brandParentUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], pacLifecycle),
	}
	driver := NewCypherDriver(s.db, "prod")
	writePacAnnotations(s.T(), s.db)
	//assert data for filtering
	numOfV1Annotations, _ := count(v1Lifecycle, s.db)
	numOfV2Annotations, _ := count(v2Lifecycle, s.db)
	numOfPACAnnotations, _ := count(pacLifecycle, s.db)
	assert.True(s.T(), numOfV1Annotations > 0)
	assert.True(s.T(), numOfV2Annotations > 0)
	assert.True(s.T(), numOfPACAnnotations > 0)

	anns := getAndCheckAnnotations(driver, contentUUID, s.T())

	assert.Len(s.T(), anns, len(expectedAnnotations), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestRetrieveImplicitAbouts() {
	expectedAnnotations := annotations{
		expectedAnnotation(aboutTopic, topicType, predicates["ABOUT"], pacLifecycle),
		expectedAnnotation(broaderTopicA, topicType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		expectedAnnotation(broaderTopicB, topicType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		getExpectedMallStreetJournalAnnotation(v2Lifecycle),
		getExpectedMentionsFakebookAnnotation(v2Lifecycle),
	}

	driver := NewCypherDriver(s.db, "prod")
	writeAboutAnnotations(s.T(), s.db)

	anns := getAndCheckAnnotations(driver, contentUUID, s.T())

	assert.Len(s.T(), anns, len(expectedAnnotations), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestRetrieveCyclicImplicitAbouts() {
	expectedAnnotations := annotations{
		expectedAnnotation(narrowerTopic, topicType, predicates["ABOUT"], pacLifecycle),
		expectedAnnotation(aboutTopic, topicType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		expectedAnnotation(broaderTopicA, topicType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		expectedAnnotation(broaderTopicB, topicType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		expectedAnnotation(cyclicTopicA, topicType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		expectedAnnotation(cyclicTopicB, topicType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		getExpectedMentionsFakebookAnnotation(v2Lifecycle),
		getExpectedMallStreetJournalAnnotation(v2Lifecycle),
	}

	driver := NewCypherDriver(s.db, "prod")
	writeCyclicAboutAnnotations(s.T(), s.db)

	anns := getAndCheckAnnotations(driver, contentUUID, s.T())
	d, _ := json.MarshalIndent(anns, "", "   ")
	s.T().Log(string(d))

	assert.Len(s.T(), anns, len(expectedAnnotations), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestRetrieveMultipleAnnotationsIfPacAnnotationCannotBeMapped() {
	expectedAnnotations := annotations{
		getExpectedMentionsFakebookAnnotation(v2Lifecycle),
		getExpectedMallStreetJournalAnnotation(v2Lifecycle),
		getExpectedMetalMickeyAnnotation(v1Lifecycle),
		getExpectedAlphavilleSeriesAnnotation(v1Lifecycle),
		expectedAnnotation(brandGrandChildUUID, brandType, predicates["IS_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandChildUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandParentUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
	}

	driver := NewCypherDriver(s.db, "prod")
	writeBrokenPacAnnotations(s.T(), s.db)
	//assert data for filtering
	numOfV1Annotations, _ := count(v1Lifecycle, s.db)
	numOfv2Annotations, _ := count(v2Lifecycle, s.db)
	numOfPacAnnotations, _ := count(pacLifecycle, s.db)
	assert.True(s.T(), (numOfV1Annotations+numOfv2Annotations) > 0)
	assert.Equal(s.T(), numOfPacAnnotations, 1)

	anns := getAndCheckAnnotations(driver, contentUUID, s.T())
	assert.Equal(s.T(), len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestRetrieveContentWithParentBrand() {
	expectedAnnotations := annotations{
		expectedAnnotation(brandGrandChildUUID, brandType, predicates["IS_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandChildUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandParentUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
	}

	driver := NewCypherDriver(s.db, "prod")
	anns := getAndCheckAnnotations(driver, contentWithParentAndChildBrandUUID, s.T())
	assert.Equal(s.T(), len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestRetrieveContentWithGrandParentBrand() {
	expectedAnnotations := annotations{
		expectedAnnotation(brandGrandChildUUID, brandType, predicates["IS_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandChildUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandParentUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
	}

	driver := NewCypherDriver(s.db, "prod")
	anns := getAndCheckAnnotations(driver, contentWithThreeLevelsOfBrandUUID, s.T())
	assert.Equal(s.T(), len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestRetrieveContentWithCircularBrand() {
	expectedAnnotations := annotations{
		expectedAnnotation(brandCircularAUUID, brandType, predicates["IS_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandCircularBUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
	}

	driver := NewCypherDriver(s.db, "prod")
	anns := getAndCheckAnnotations(driver, contentWithCircularBrandUUID, s.T())
	assert.Equal(s.T(), len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestRetrieveContentWithJustParentBrand() {
	expectedAnnotations := annotations{
		expectedAnnotation(brandParentUUID, brandType, predicates["IS_CLASSIFIED_BY"], v1Lifecycle),
	}

	driver := NewCypherDriver(s.db, "prod")
	anns := getAndCheckAnnotations(driver, contentWithOnlyFTUUID, s.T())
	assert.Equal(s.T(), len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

//Tests filtering Annotations where content is related to Brand A as isClassifiedBy and to Brand B as isPrimarilyClassifiedBy
// and Brands A and B have a circular relation HasParent
func (s *cypherDriverTestSuite) TestRetrieveContentBrandsOfDifferentTypes() {
	expectedAnnotations := annotations{
		expectedAnnotation(brandCircularAUUID, brandType, predicates["IS_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandCircularBUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
	}

	driver := NewCypherDriver(s.db, "prod")
	anns := getAndCheckAnnotations(driver, contentWithCircularBrandUUID, s.T())
	assert.Equal(s.T(), len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func TestRetrieveNoAnnotationsWhenThereAreNonePresentExceptBrands(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnection(t)

	writeContent(t, db)
	writeBrands(t, db)

	defer cleanDB(t, db)

	driver := NewCypherDriver(db, "prod")
	anns, found, err := driver.read(contentWithNoAnnotationsUUID)
	assert.NoError(err, "Unexpected error for content %s", contentWithNoAnnotationsUUID)
	assert.False(found, "Found annotations for content %s", contentWithNoAnnotationsUUID)
	assert.Equal(0, len(anns), "Didn't get the same number of annotations") // Two brands, child and parent
}

func TestRetrieveAnnotationWithCorrectValues(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnection(t)
	writeContent(t, db)
	writeOrganisations(t, db)
	writeFinancialInstruments(t, db)
	writeV2Annotations(t, db)
	defer cleanDB(t, db)

	expectedAnnotations := annotations{
		getExpectedMentionsFakebookAnnotation(v2Lifecycle),
		getExpectedMallStreetJournalAnnotation(v2Lifecycle),
	}

	driver := NewCypherDriver(db, "prod")
	anns := getAndCheckAnnotations(driver, contentUUID, t)

	assert.Equal(len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(t, anns, expectedAnnotations)

	for _, ann := range anns {
		for _, expected := range expectedAnnotations {
			if expected.ID == ann.ID {
				assert.Equal(expected.FIGI, ann.FIGI, "Didn't get the expected FIGI value")
				assert.Equal(expected.LeiCode, ann.LeiCode, "Didn't get the expected Leicode value")
				break
			}
		}
	}
}

func TestRetrieveNoAnnotationsWhenThereAreNoConceptsPresent(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnection(t)

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

// Utility functions
func writeAllDataToDB(t testing.TB, db neoutils.NeoConnection) {
	writeBrands(t, db)
	writeContent(t, db)
	writeOrganisations(t, db)
	writePeople(t, db)
	writeFinancialInstruments(t, db)
	writeSubjects(t, db)
	writeAlphavilleSeries(t, db)
	writeV1Annotations(t, db)
	writeV2Annotations(t, db)
	writeTopics(t, db)
}

func writeBrands(t testing.TB, db neoutils.NeoConnection) concepts.ConceptService {
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
	writeJSONToBaseService(contentRW, "./fixtures/Content-3fc9fe3e-af8c-4f7f-961a-e5065392bb31.json", t)
	writeJSONToBaseService(contentRW, "./fixtures/Content-3fc9fe3e-af8c-1a1a-961a-e5065392bb31.json", t)
	writeJSONToBaseService(contentRW, "./fixtures/Content-3fc9fe3e-af8c-2a2a-961a-e5065392bb31.json", t)
	writeJSONToBaseService(contentRW, "./fixtures/Content-3fc9fe3e-af8c-3a3a-961a-e5065392bb31.json", t)
	writeJSONToBaseService(contentRW, "./fixtures/Content-3fc9fe3e-af8c-4a4a-961a-e5065392bb31.json", t)
	writeJSONToBaseService(contentRW, "./fixtures/Content-3fc9fe3e-af8c-5a5a-961a-e5065392bb31.json", t)
	writeJSONToBaseService(contentRW, "./fixtures/Content-3fc9fe3e-af8c-6a6a-961a-e5065392bb31.json", t)
	return contentRW
}

func writeTopics(t testing.TB, db neoutils.NeoConnection) concepts.ConceptService {
	topicsRW := concepts.NewConceptService(db)
	assert.NoError(t, topicsRW.Initialise())
	writeJSONToService(topicsRW, "./fixtures/Topics-7e22c8b8-b280-4e52-aa22-fa1c6dffd894.json", t)
	writeJSONToService(topicsRW, "./fixtures/Topics-b6469cc2-f6ff-45aa-a9bb-3d1bb0f9a35d.json", t)
	writeJSONToService(topicsRW, "./fixtures/Topics-ca982370-66cd-43bd-b2e3-7bfcb73efb1e.json", t)
	writeJSONToService(topicsRW, "./fixtures/Topics-fde5eee9-3260-4125-adb6-3d91a4888be5.json", t)
	writeJSONToService(topicsRW, "./fixtures/Topics-77a410a3-6857-4654-80ef-6aae29be852a.json", t)
	writeJSONToService(topicsRW, "./fixtures/Topics-e404e3bd-beff-4324-83f4-beb044baf916.json", t)
	return topicsRW
}

func writeOrganisations(t testing.TB, db neoutils.NeoConnection) concepts.ConceptService {
	organisationRW := concepts.NewConceptService(db)
	assert.NoError(t, organisationRW.Initialise())
	writeJSONToService(organisationRW, "./fixtures/Organisation-MSJ-5d1510f8-2779-4b74-adab-0a5eb138fca6.json", t)
	writeJSONToService(organisationRW, "./fixtures/Organisation-Fakebook-eac853f5-3859-4c08-8540-55e043719400.json", t)
	return organisationRW
}

func writePeople(t testing.TB, db neoutils.NeoConnection) concepts.ConceptService {
	peopleRW := concepts.NewConceptService(db)
	assert.NoError(t, peopleRW.Initialise())
	writeJSONToService(peopleRW, "./fixtures/People-75e2f7e9-cb5e-40a5-a074-86d69fe09f69.json", t)
	return peopleRW
}

func writeFinancialInstruments(t testing.TB, db neoutils.NeoConnection) concepts.ConceptService {
	fiRW := concepts.NewConceptService(db)
	assert.NoError(t, fiRW.Initialise())
	writeJSONToService(fiRW, "./fixtures/FinancialInstrument-77f613ad-1470-422c-bf7c-1dd4c3fd1693.json", t)
	return fiRW
}

func writeSubjects(t testing.TB, db neoutils.NeoConnection) concepts.ConceptService {
	subjectsRW := concepts.NewConceptService(db)
	assert.NoError(t, subjectsRW.Initialise())
	writeJSONToService(subjectsRW, "./fixtures/Subject-MetalMickey-0483bef8-5797-40b8-9b25-b12e492f63c6.json", t)
	return subjectsRW
}

func writeAlphavilleSeries(t testing.TB, db neoutils.NeoConnection) concepts.ConceptService {
	alphavilleSeriesRW := concepts.NewConceptService(db)
	assert.NoError(t, alphavilleSeriesRW.Initialise())
	writeJSONToService(alphavilleSeriesRW, "./fixtures/TestAlphavilleSeries.json", t)
	return alphavilleSeriesRW
}

func writeV1Annotations(t testing.TB, db neoutils.NeoConnection) annrw.Service {
	service := annrw.NewCypherAnnotationsService(db)
	assert.NoError(t, service.Initialise())

	writeJSONToAnnotationsService(t, service, v1PlatformVersion, v1Lifecycle, contentUUID, "./fixtures/Annotations-3fc9fe3e-af8c-4f7f-961a-e5065392bb31-v1.json")
	writeJSONToAnnotationsService(t, service, v1PlatformVersion, v1Lifecycle, contentWithParentAndChildBrandUUID, "./fixtures/Annotations-3fc9fe3e-af8c-2a2a-961a-e5065392bb31-v1.json")
	writeJSONToAnnotationsService(t, service, v1PlatformVersion, v1Lifecycle, contentWithThreeLevelsOfBrandUUID, "./fixtures/Annotations-3fc9fe3e-af8c-3a3a-961a-e5065392bb31-v1.json")
	writeJSONToAnnotationsService(t, service, v1PlatformVersion, v1Lifecycle, contentWithCircularBrandUUID, "./fixtures/Annotations-3fc9fe3e-af8c-4a4a-961a-e5065392bb31-v1.json")
	writeJSONToAnnotationsService(t, service, v1PlatformVersion, v1Lifecycle, contentWithOnlyFTUUID, "./fixtures/Annotations-3fc9fe3e-af8c-5a5a-961a-e5065392bb31-v1.json")
	writeJSONToAnnotationsService(t, service, v1PlatformVersion, v1Lifecycle, contentWithBrandsDiffTypesUUID, "./fixtures/Annotations-3fc9fe3e-af8c-6a6a-961a-e5065392bb31-v1.json")
	return service
}

func writeV2Annotations(t testing.TB, db neoutils.NeoConnection) annrw.Service {
	service := annrw.NewCypherAnnotationsService(db)
	assert.NoError(t, service.Initialise())
	writeJSONToAnnotationsService(t, service, v2PlatformVersion, v2Lifecycle, contentUUID, "./fixtures/Annotations-3fc9fe3e-af8c-4f7f-961a-e5065392bb31-v2.json")

	return service
}

func writePacAnnotations(t testing.TB, db neoutils.NeoConnection) annrw.Service {
	service := annrw.NewCypherAnnotationsService(db)
	assert.NoError(t, service.Initialise())
	writeJSONToAnnotationsService(t, service, "pac", "annotations-pac", contentUUID, "./fixtures/Annotations-3fc9fe3e-af8c-4f7f-961a-e5065392bb31-pac.json")
	return service
}

func writeAboutAnnotations(t testing.TB, db neoutils.NeoConnection) annrw.Service {
	service := annrw.NewCypherAnnotationsService(db)
	assert.NoError(t, service.Initialise())
	writeJSONToAnnotationsService(t, service, "pac", "annotations-pac", contentUUID, "./fixtures/Annotations-ca982370-66cd-43bd-b2e3-7bfcb73efb1e-implicit-abouts.json")
	return service
}

func writeCyclicAboutAnnotations(t testing.TB, db neoutils.NeoConnection) annrw.Service {
	service := annrw.NewCypherAnnotationsService(db)
	assert.NoError(t, service.Initialise())
	writeJSONToAnnotationsService(t, service, "pac", "annotations-pac", contentUUID, "./fixtures/Annotations-7e22c8b8-b280-4e52-aa22-fa1c6dffd894-cyclic-implicit-abouts.json")
	return service
}

func writeBrokenPacAnnotations(t testing.TB, db neoutils.NeoConnection) annrw.Service {
	service := annrw.NewCypherAnnotationsService(db)
	assert.NoError(t, service.Initialise())
	writeJSONToAnnotationsService(t, service, emptyPlatformVersion, pacLifecycle, contentUUID, "./fixtures/Annotations-3fc9fe3e-af8c-4f7f-961a-e5065392bb31-broken-pac.json")
	return service
}

func writeJSONToBaseService(service baseftrwapp.Service, pathToJSONFile string, t testing.TB) {
	absPath, _ := filepath.Abs(pathToJSONFile)
	f, err := os.Open(absPath)
	assert.NoError(t, err)
	dec := json.NewDecoder(f)
	inst, _, err := service.DecodeJSON(dec)
	assert.NoError(t, err)
	err = service.Write(inst, "TEST_TRANS_ID")
	assert.NoError(t, err)
}

func writeJSONToService(service concepts.ConceptService, pathToJSONFile string, t testing.TB) {
	absPath, _ := filepath.Abs(pathToJSONFile)
	f, err := os.Open(absPath)
	assert.NoError(t, err)
	dec := json.NewDecoder(f)
	inst, _, err := service.DecodeJSON(dec)
	assert.NoError(t, err)
	_, err = service.Write(inst, "TEST_TRANS_ID")
	assert.NoError(t, err)
}

func writeJSONToAnnotationsService(t testing.TB, service annrw.Service, platformVersion string, lifecycle string, contentUUID string, pathToJSONFile string) {
	absPath, _ := filepath.Abs(pathToJSONFile)
	f, err := os.Open(absPath)
	assert.NoError(t, err)
	dec := json.NewDecoder(f)
	inst, err := service.DecodeJSON(dec)
	assert.NoError(t, err, "Error parsing file %s", pathToJSONFile)
	err = service.Write(contentUUID, lifecycle, platformVersion, "TID_TEST", inst)
	assert.NoError(t, err)
}

func assertListContainsAll(t *testing.T, list interface{}, items ...interface{}) {
	if reflect.TypeOf(items[0]).Kind().String() == "slice" {
		expected := reflect.ValueOf(items[0])
		expectedLength := expected.Len()
		for i := 0; i < expectedLength; i++ {
			assert.Contains(t, list, expected.Index(i).Interface())
		}
	} else {
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

func getExpectedMentionsFakebookAnnotation(lifecycle string) annotation {
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
		FIGI:      "BB8000C3P0-R2D2",
		PrefLabel: "Fakebook, Inc.",
		Lifecycle: lifecycle,
	}
}

func getExpectedAboutFakebookAnnotation(lifecycle string) annotation {
	return annotation{
		Predicate: "http://www.ft.com/ontology/annotation/about",
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
		FIGI:      "BB8000C3P0-R2D2",
		PrefLabel: "Fakebook, Inc.",
		Lifecycle: lifecycle,
	}
}

func getExpectedMallStreetJournalAnnotation(lifecycle string) annotation {
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
		Lifecycle: lifecycle,
	}
}

func getExpectedMetalMickeyAnnotation(lifecycle string) annotation {
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
		Lifecycle: lifecycle,
	}
}

func getExpectedHasDisplayTagFakebookAnnotation(lifecycle string) annotation {
	return annotation{
		Predicate: "http://www.ft.com/ontology/hasDisplayTag",
		ID:        "http://api.ft.com/things/eac853f5-3859-4c08-8540-55e043719400",
		APIURL:    "http://api.ft.com/organisations/eac853f5-3859-4c08-8540-55e043719400",
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/organisation/Organisation",
			"http://www.ft.com/ontology/company/Company",
			"http://www.ft.com/ontology/company/PublicCompany",
		},
		PrefLabel: "Fakebook, Inc.",
		Lifecycle: lifecycle,
		LeiCode:   "BQ4BKCS1HXDV9TTTTTTTT",
		FIGI:      "BB8000C3P0-R2D2",
	}
}

func getExpectedJohnSmithAnnotation(lifecycle string) annotation {
	return annotation{
		Predicate: "http://www.ft.com/ontology/hasContributor",
		ID:        "http://api.ft.com/things/75e2f7e9-cb5e-40a5-a074-86d69fe09f69",
		APIURL:    "http://api.ft.com/people/75e2f7e9-cb5e-40a5-a074-86d69fe09f69",
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/person/Person",
		},
		PrefLabel: "John Smith",
		Lifecycle: lifecycle,
	}
}

func getExpectedAlphavilleSeriesAnnotation(lifecycle string) annotation {
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
		Lifecycle: lifecycle,
	}
}

func expectedAnnotation(conceptUuid string, conceptType string, predicate string, lifecycle string) annotation {
	return annotation{
		Predicate: predicate,
		ID:        fmt.Sprintf("http://api.ft.com/things/%s", conceptUuid),
		APIURL:    fmt.Sprintf(conceptApiUrlTemplates[conceptType], conceptUuid),
		Types:     conceptTypes[conceptType],
		PrefLabel: conceptLabels[conceptUuid],
		Lifecycle: lifecycle,
	}
}

func count(annotationLifecycle string, db neoutils.NeoConnection) (int, error) {
	var results []struct {
		Count int `json:"c"`
	}
	query := &neoism.CypherQuery{
		Statement: `MATCH (c:Content)-[r]->( t:Thing)
					WHERE r.lifecycle = {lifecycle}
                	RETURN count(r) as c`,
		Parameters: neoism.Props{"lifecycle": annotationLifecycle},
		Result:     &results,
	}
	err := db.CypherBatch([]*neoism.CypherQuery{query})

	if err != nil {
		return 0, err
	}
	return results[0].Count, nil
}
