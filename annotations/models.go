package annotations

type annotations []annotation

type annotation struct {
	Predicate string   `json:"predicate"`
	ID        string   `json:"id"`
	APIURL    string   `json:"apiUrl"`
	Types     []string `json:"types"`
	LeiCode   string   `json:"leiCode,omitempty"`
	PrefLabel string   `json:"prefLabel,omitempty"`
}

var predicates = map[string]string{
	"MENTIONS":         "http://www.ft.com/ontology/annotation/mentions",
	"IS_CLASSIFIED_BY": "http://www.ft.com/ontology/classification/isClassifiedBy",
	"ABOUT":            "http://www.ft.com/ontology/annotation/about",
	"IS_PRIMARILY_CLASSIFIED_BY": "http://www.ft.com/ontology/classification/isPrimarilyClassifiedBy",
}
