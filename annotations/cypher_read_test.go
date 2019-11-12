package annotations

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
)

type MockNeoConnection struct {
	neoutils.CypherRunner
	neoutils.IndexEnsurer

	cypherBatch func(queries []*neoism.CypherQuery) error
}

func (mc MockNeoConnection) CypherBatch(queries []*neoism.CypherQuery) error {
	if mc.cypherBatch == nil {
		return errors.New("not implemented")
	}
	return mc.cypherBatch(queries)
}

func TestCypherDriverReadBrand(t *testing.T) {
	tests := []struct {
		name           string
		neoResult      []neoAnnotation
		expectedResult annotations
		expectedFound  bool
		expectedError  bool
	}{
		{
			name: "hasBrand predicate",
			neoResult: []neoAnnotation{
				{
					Predicate: "HAS_BRAND",
					ID:        "test",
					Types:     []string{"Brand"},
				},
			},
			expectedResult: annotations{
				{
					Predicate: "http://www.ft.com/ontology/classification/isClassifiedBy",
					ID:        "http://api.ft.com/things/test",
					APIURL:    "http://test.api.ft.com/brands/test",
					Types:     []string{"http://www.ft.com/ontology/product/Brand"},
				},
			},
			expectedFound: true,
			expectedError: false,
		},
		{
			name: "isClassifiedBy predicate",
			neoResult: []neoAnnotation{
				{
					Predicate: "IS_CLASSIFIED_BY",
					ID:        "test",
					Types:     []string{"Brand"},
				},
			},
			expectedResult: annotations{
				{
					Predicate: "http://www.ft.com/ontology/classification/isClassifiedBy",
					ID:        "http://api.ft.com/things/test",
					APIURL:    "http://test.api.ft.com/brands/test",
					Types:     []string{"http://www.ft.com/ontology/product/Brand"},
				},
			},
			expectedFound: true,
			expectedError: false,
		},
		{
			name: "hasBrand and isClassifiedBy predicate",
			neoResult: []neoAnnotation{
				{
					Predicate: "HAS_BRAND",
					ID:        "id1",
					Types:     []string{"Brand"},
				},
				{
					Predicate: "IS_CLASSIFIED_BY",
					ID:        "id2",
					Types:     []string{"Brand"},
				},
			},
			expectedResult: annotations{
				{
					Predicate: "http://www.ft.com/ontology/classification/isClassifiedBy",
					ID:        "http://api.ft.com/things/id1",
					APIURL:    "http://test.api.ft.com/brands/id1",
					Types:     []string{"http://www.ft.com/ontology/product/Brand"},
				},
				{
					Predicate: "http://www.ft.com/ontology/classification/isClassifiedBy",
					ID:        "http://api.ft.com/things/id2",
					APIURL:    "http://test.api.ft.com/brands/id2",
					Types:     []string{"http://www.ft.com/ontology/product/Brand"},
				},
			},
			expectedFound: true,
			expectedError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			mockConn := MockNeoConnection{
				cypherBatch: func(queries []*neoism.CypherQuery) error {
					if len(queries) < 1 {
						t.Fatal("Unexpected query param")
					}
					q := queries[0]
					// we use json marshall and unmarshall so we don't have to use reflection directly
					jsonAnn, err := json.Marshal(test.neoResult)
					assert.NoError(err, "Unexpected error marshalling Neo results")
					err = json.Unmarshal(jsonAnn, q.Result)
					assert.NoError(err, "Unexpected error unmarshalling Neo results")
					return nil
				},
			}

			testDriver := NewCypherDriver(mockConn, "test")
			result, found, err := testDriver.read("contentUUID")
			if !test.expectedError && err != nil {
				t.Fatalf("Unexpected error reading annotations: %v", err)
			}
			assert.Equal(found, test.expectedFound)
			if test.expectedError {
				assert.Error(err, "Expected error reading annotations")
			}
			assert.Equal(test.expectedResult, result)
		})
	}
}
