package annotations

import (
	"fmt"

	"errors"

	"github.com/Financial-Times/neo-model-utils-go/mapper"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
)

// Driver interface
type driver interface {
	read(id string) (anns annotations, found bool, err error)
	filteredRead(id string, platformVersion string) (anns annotations, found bool, err error)
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

func (cd cypherDriver) checkConnectivity() error { //TODO - use the neo4j connectivity check library
	return neoutils.Check(cd.conn)
}

func (cd cypherDriver) read(contentUUID string) (anns annotations, found bool, err error) {
	results := []annotation{}

	query := &neoism.CypherQuery{
		Statement: `
				MATCH (c:Thing{uuid:{contentUUID}})-[rel]->(cc:Concept)
				OPTIONAL MATCH (cc)<-[iden:IDENTIFIES]-(i:LegalEntityIdentifier)
				WITH c, collect({id: cc.uuid, predicate: type(rel), types: labels(cc), prefLabel:cc.prefLabel, leiCode:i.value}) as rows
				OPTIONAL MATCH (cc:Concept)<-[r:HAS_PARENT*0..]-(:Brand)<-[rel]-(c)
				WITH collect({id: cc.uuid, predicate:'IS_CLASSIFIED_BY', types: labels(cc), prefLabel:cc.prefLabel,leiCode:null}) + rows as allRows
				UNWIND allRows as row
				WITH DISTINCT(row) as drow
				RETURN drow.id as id, drow.predicate as predicate, drow.types as types, drow.prefLabel as prefLabel, drow.leiCode as leiCode
				`,
		Parameters: neoism.Props{"contentUUID": contentUUID},
		Result:     &results,
	}

	err = cd.conn.CypherBatch([]*neoism.CypherQuery{query})
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
		annotation, err := mapToResponseFormat(&results[idx], cd.env)
		if err == nil {
			mappedAnnotations = append(mappedAnnotations, *annotation)
			found = true
		}
	}

	return mappedAnnotations, found, nil
}

// Returns all the annotations with the specified platformVersion enriched with all the existing concept IDs for a given content
func (cd cypherDriver) filteredRead(contentUUID string, platformVersion string) (anns annotations, found bool, err error) {
	results := []annotation{}

	query := &neoism.CypherQuery{
		Statement: `
				MATCH (c:Thing{uuid:{contentUUID}})-[rel{platformVersion:{platformVersion}}]->(cc:Concept)
				MATCH (cc)<-[:IDENTIFIES]-(upp:UPPIdentifier)
				optional MATCH (cc)<-[:IDENTIFIES]-(lei:LegalEntityIdentifier)
				optional MATCH (cc)<-[:IDENTIFIES]-(tme:TMEIdentifier)
				optional MATCH (cc)<-[:IDENTIFIES]-(fs:FactsetIdentifier)
				WITH c, cc, rel, lei, fs, collect(distinct tme.value) as tme, collect(distinct upp.value) as upp
				WITH c, collect({id: cc.uuid, predicate: type(rel), types: labels(cc), prefLabel:cc.prefLabel, uuids:upp, tmeIDs:tme, leiCode:lei.value, factsetID:fs.value, platformVersion:rel.platformVersion}) as rows
				UNWIND rows as row
				WITH DISTINCT(row) as drow
				RETURN drow.id as id, drow.predicate as predicate, drow.types as types, drow.prefLabel as prefLabel, drow.leiCode as leiCode, drow.factsetID as factsetID, drow.uuids as uuids, drow.tmeIDs as tmeIDs, drow.platformVersion as platformVersion`,
		Parameters: neoism.Props{"contentUUID": contentUUID, "platformVersion": platformVersion},
		Result:     &results,
	}

	err = cd.conn.CypherBatch([]*neoism.CypherQuery{query})
	if err != nil {
		log.Errorf("Error looking up uuid %s with query %s from neoism: %+v", contentUUID, query.Statement, err)
		return annotations{}, false, fmt.Errorf("Error accessing Annotations datastore for uuid: %s", contentUUID)
	}
	log.Debugf("Found %d Annotations for uuid: %s with platformVersion: %s", len(results), contentUUID, platformVersion)
	if (len(results)) == 0 {
		return annotations{}, false, nil
	}

	mappedAnnotations := []annotation{}

	found = false

	for idx := range results {
		annotation, err := mapToResponseFormat(&results[idx], cd.env)
		if err == nil {
			mappedAnnotations = append(mappedAnnotations, *annotation)
			found = true
		}
	}

	return mappedAnnotations, found, nil
}

func mapToResponseFormat(ann *annotation, env string) (*annotation, error) {
	ann.APIURL = mapper.APIURL(ann.ID, ann.Types, env)
	ann.ID = mapper.IDURL(ann.ID)
	types := mapper.TypeURIs(ann.Types)
	if types == nil {
		log.Warnf("Could not map type URIs for ID %s with types %s", ann.ID, ann.Types)
		return ann, errors.New("Concept not found")
	}
	ann.Types = types
	predicate, err := getPredicateFromRelationship(ann.Predicate)
	if err != nil {
		log.Warnf("Could not find predicate for ID %s for relationship %s", ann.ID, ann.Predicate)
		return ann, err
	}
	ann.Predicate = predicate
	return ann, nil
}

func getPredicateFromRelationship(relationship string) (predicate string, err error) {
	predicate = predicates[relationship]
	if predicate == "" {
		return "", errors.New("Not a valid annotation type")
	}
	return predicate, nil
}
