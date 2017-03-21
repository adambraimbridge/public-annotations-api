package annotations

type annotations []annotation

type annotation struct {
	Predicate string   `json:"predicate"`
	ID        string   `json:"id"`
	APIURL    string   `json:"apiUrl"`
	Types     []string `json:"types"`
	LeiCode   string   `json:"leiCode,omitempty"`
	PrefLabel string   `json:"prefLabel,omitempty"`
	//the fields below are populated only for the /content/{uuid}/annotations/{plaformVersion} endpoint
	FactsetID       string   `json:"factsetID,omitempty"`
	TmeIDs          []string `json:"tmeIDs,omitempty"`
	UUIDs           []string `json:"uuids,omitempty"`
	PlatformVersion string   `json:"platformVersion,omitempty"`
}

var predicates = map[string]string{
	"MENTIONS":         "http://www.ft.com/ontology/annotation/mentions",
	"MAJOR_MENTIONS":   "http://www.ft.com/ontology/annotation/majorMentions",
	"IS_CLASSIFIED_BY": "http://www.ft.com/ontology/classification/isClassifiedBy",
	"ABOUT":            "http://www.ft.com/ontology/annotation/about",
	"IS_PRIMARILY_CLASSIFIED_BY": "http://www.ft.com/ontology/classification/isPrimarilyClassifiedBy",
	"HAS_AUTHOR":                 "http://www.ft.com/ontology/annotation/hasAuthor",
}
