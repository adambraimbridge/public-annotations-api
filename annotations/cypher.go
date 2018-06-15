package annotations

import (
	"fmt"
	"time"

	"errors"

	"github.com/Financial-Times/neo-model-utils-go/mapper"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
	log "github.com/sirupsen/logrus"
)

// Driver interface
type driver interface {
	read(id string) (anns annotations, found bool, err error)
	checkConnectivity() error
}

// CypherDriver struct
type cypherDriver struct {
	conn neoutils.NeoConnection
	env  string
}

func NewCypherDriver(conn neoutils.NeoConnection, env string) cypherDriver {
	return cypherDriver{conn, env}
}

func (cd cypherDriver) checkConnectivity() error {
	return neoutils.Check(cd.conn)
}

type neoAnnotation struct {
	Predicate string
	ID        string
	APIURL    string
	Types     []string
	LeiCode   string
	FIGI      string
	PrefLabel string
	Lifecycle string

	// Canonical information
	PrefUUID           string
	CanonicalTypes     []string
	CanonicalLeiCode   string
	CanonicalPrefLabel string

	//the fields below are populated only for the /content/{uuid}/annotations/{plaformVersion} endpoint
	FactsetIDs      []string `json:"factsetID,omitempty"`
	TmeIDs          []string `json:"tmeIDs,omitempty"`
	UUIDs           []string `json:"uuids,omitempty"`
	PlatformVersion string   `json:"platformVersion,omitempty"`
}

func (cd cypherDriver) read(contentUUID string) (anns annotations, found bool, err error) {
	results := []neoAnnotation{}

	query := &neoism.CypherQuery{
		Statement: `
      MATCH (content:Thing{uuid:{contentUUID}})-[rel]-(concept:Concept)
      OPTIONAL MATCH (concept)-[:EQUIVALENT_TO]->(canonicalConcept:Concept)
      OPTIONAL MATCH (concept)<-[:IDENTIFIES]-(lei:LegalEntityIdentifier)
      OPTIONAL MATCH (concept)<-[:ISSUED_BY]-(:FinancialInstrument)<-[:IDENTIFIES]-(figi:FIGIIdentifier)
      RETURN coalesce(canonicalConcept.prefUUID, concept.uuid) as id, type(rel) as predicate, coalesce(labels(canonicalConcept), labels(concept)) as types,
				coalesce(canonicalConcept.prefLabel, concept.prefLabel) as prefLabel, lei.value as leiCode, figi.value as figi, rel.lifecycle as lifecycle

      UNION ALL
      MATCH (content:Thing{uuid:{contentUUID}})-[rel]-(brand:Brand)-[:EQUIVALENT_TO]->(canonicalBrand:Brand)
      OPTIONAL MATCH (canonicalBrand)-[:EQUIVALENT_TO]-(leafBrand:Brand)-[r:HAS_PARENT*0..]->(parentBrand:Brand)-[:EQUIVALENT_TO]->(canonicalParent:Brand)
      RETURN distinct coalesce(canonicalParent.prefUUID, parentBrand.uuid) as id, "IMPLICITLY_CLASSIFIED_BY" as predicate, coalesce(labels(canonicalParent), labels(parentBrand)) as types,
				coalesce(canonicalParent.prefLabel, parentBrand.prefLabel) as prefLabel, null as leiCode, null as figi, rel.lifecycle as lifecycle

      UNION ALL
      MATCH (content:Thing{uuid:{contentUUID}})-[rel:ABOUT]-(concept:Concept)-[:EQUIVALENT_TO]->(canonicalConcept:Concept)
      MATCH (canonicalConcept)<-[:EQUIVALENT_TO]-(leafConcept:Concept)-[:HAS_BROADER*1..]->(implicit:Concept)-[:EQUIVALENT_TO]->(canonicalImplicit)
      WHERE NOT (canonicalImplicit)<-[:EQUIVALENT_TO]-(:Concept)<-[:ABOUT]-(content) // filter out the original abouts
      RETURN distinct canonicalImplicit.prefUUID as id, "IMPLICITLY_ABOUT" as predicate, labels(canonicalImplicit) as types, canonicalImplicit.prefLabel as prefLabel, null as leiCode, null as figi, rel.lifecycle as lifecycle
      `,
		Parameters: neoism.Props{"contentUUID": contentUUID},
		Result:     &results,
	}

	start := time.Now()
	err = cd.conn.CypherBatch([]*neoism.CypherQuery{query})
	end := time.Now()

	log.WithField("contentUUID", contentUUID).WithField("duration", fmt.Sprintf("%vms", end.Sub(start).Nanoseconds()/1e6)).Info("Annotations query (including implicit relationships) completed.")

	if err != nil {
		log.Errorf("Error looking up uuid %s with query %s from neoism: %+v", contentUUID, query.Statement, err)
		return annotations{}, false, fmt.Errorf("Error accessing Annotations datastore for uuid: %s", contentUUID)
	}

	log.Debugf("Found %d Annotations for uuid: %s", len(results), contentUUID)
	if (len(results)) == 0 {
		return annotations{}, false, nil
	}

	mappedAnnotations := []annotation{}
	found = false

	for idx := range results {
		annotation, err := mapToResponseFormat(results[idx], cd.env)
		if err == nil {
			found = true
			mappedAnnotations = append(mappedAnnotations, annotation)
		}
	}

	lifecycleFilter := newLifecycleFilter()
	predicateFilter := NewAnnotationsPredicateFilter()

	chain := newAnnotationsFilterChain(lifecycleFilter, predicateFilter)
	return chain.doNext(mappedAnnotations), found, nil
}

func mapToResponseFormat(neoAnn neoAnnotation, env string) (annotation, error) {
	var ann annotation

	ann.PrefLabel = neoAnn.PrefLabel
	ann.LeiCode = neoAnn.LeiCode
	ann.FIGI = neoAnn.FIGI
	ann.APIURL = mapper.APIURL(neoAnn.ID, neoAnn.Types, env)
	ann.ID = mapper.IDURL(neoAnn.ID)
	types := mapper.TypeURIs(neoAnn.Types)
	if types == nil || len(types) == 0 {
		log.Debugf("Could not map type URIs for ID %s with types %s", ann.ID, ann.Types)
		return ann, errors.New("Concept not found")
	}
	ann.Types = types

	predicate, err := getPredicateFromRelationship(neoAnn.Predicate)
	if err != nil {
		log.Debugf("Could not find predicate for ID %s for relationship %s", ann.ID, ann.Predicate)
		return ann, err
	}
	ann.Predicate = predicate
	ann.Lifecycle = neoAnn.Lifecycle

	return ann, nil
}

func deduplicateList(inList []string) []string {
	outList := []string{}
	deduped := map[string]bool{}
	for _, v := range inList {
		deduped[v] = true
	}
	for k, o := range deduped {
		if o {
			outList = append(outList, k)
		}
	}
	return outList
}

func getPredicateFromRelationship(relationship string) (predicate string, err error) {
	predicate = predicates[relationship]
	if predicate == "" {
		return "", errors.New("Not a valid annotation type")
	}
	return predicate, nil
}
