package main

import (
	"fmt"

	"errors"

	"github.com/Financial-Times/neo-model-utils-go/mapper"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
)

// Driver interface
type driver interface {
	read(id string) (anns annotations, found bool, err error)
	checkConnectivity() error
}

// CypherDriver struct
type cypherDriver struct {
	db  *neoism.Database
	env string
}

func newCypherDriver(db *neoism.Database, env string) cypherDriver {
	return cypherDriver{db, env}
}

func (cd cypherDriver) checkConnectivity() error { //TODO - use the neo4j connectivity check library
	results := []struct {
		ID int
	}{}
	query := &neoism.CypherQuery{
		Statement: "MATCH (x) RETURN ID(x) LIMIT 1",
		Result:    &results,
	}
	err := cd.db.Cypher(query)
	log.Debugf("CheckConnectivity results:%+v  err: %+v", results, err)
	return err
}

type neoReadStruct struct {
}

func (cd cypherDriver) read(contentUUID string) (anns annotations, found bool, err error) {
	results := []annotation{}

	query := &neoism.CypherQuery{
		Statement: `
					MATCH (c:Thing{uuid:{contentUUID}})-[rel]->(cc:Thing)
					OPTIONAL MATCH (cc)<-[iden:IDENTIFIES]-(i:LegalEntityIdentifier)
					RETURN cc.uuid as id,
					type(rel) as predicate,
					labels(cc) as types,
					cc.prefLabel as prefLabel,
					i.value as leiCode
					`,
		Parameters: neoism.Props{"contentUUID": contentUUID},
		Result:     &results,
	}
	err = cd.db.Cypher(query)
	if err != nil {
		log.Errorf("Error looking up uuid %s with query %s from neoism: %+v", contentUUID, query.Statement, err)
		return annotations{}, false, fmt.Errorf("Error accessing Annotations datastore for uuid: %s", contentUUID)
	}
	log.Debugf("CypherResult Read Annotations for uuid: %s was: %+v", contentUUID, results)
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
	log.Infof("ann.types=%s", ann.Types)
	types := mapper.TypeURIs(ann.Types) //TODO - change the mapper so it returns a type of 'Thing' if nothing else
	if types == nil {
		return ann, errors.New("Concept not found")
	}
	log.Infof("types=%s", types)
	ann.Types = types
	predicate, err := getPredicateFromRelationship(ann.Predicate)
	if err != nil {
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
