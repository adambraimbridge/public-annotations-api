package mapper

import "log"

var apiPaths = map[string]string{
	"Organisation": "organisations",
	"Person":       "people",
	"Brand":        "brands",
	"Thing":        "things",
	"Content":      "content",
}

var typeURIs = map[string]string{
	"Thing":                  "http://www.ft.com/ontology/core/Thing",
	"Concept":                "http://www.ft.com/ontology/concept/Concept",
	"Role":                   "http://www.ft.com/ontology/organisation/Role",
	"BoardRole":              "http://www.ft.com/ontology/organisation/BoardRole",
	"Classification":         "http://www.ft.com/ontology/classification/Classification",
	"IndustryClassification": "http://www.ft.com/ontology/industry/IndustryClassification",
	"Person":                 "http://www.ft.com/ontology/person/Person",
	"Organisation":           "http://www.ft.com/ontology/organisation/Organisation",
	"Membership":             "http://www.ft.com/ontology/organisation/Membership",
	"Company":                "http://www.ft.com/ontology/company/Company",
	"PublicCompany":          "http://www.ft.com/ontology/company/PublicCompany",
	"PrivateCompany":         "http://www.ft.com/ontology/company/PrivateCompany",
	"Brand":                  "http://www.ft.com/ontology/product/Brand",
	"Subject":                "http://www.ft.com/ontology/Subject",
	"Section":                "http://www.ft.com/ontology/Section",
	"Topic":                  "http://www.ft.com/ontology/Topic",
	"Location":               "http://www.ft.com/ontology/Location",
	"Genre":                  "http://www.ft.com/ontology/Genre",
	"SpecialReport":          "http://www.ft.com/ontology/SpecialReport",
	"AlphavilleSeries":       "http://www.ft.com/ontology/AlphavilleSeries",
	"FinancialInstrument":    "http://www.ft.com/ontology/FinancialInstrument",
}

// APIURL - Establishes the ApiURL given a whether the Label is a Person, Organisation or Company (Public or Private)
func APIURL(uuid string, labels []string, env string) string {
	base := "http://api.ft.com/"
	if env == "test" {
		base = "http://test.api.ft.com/"
	}

	path := ""
	mostSpecific, err := mostSpecific(labels)
	if err == nil {
		for t := mostSpecific; t != "" && path == ""; t = ParentType(t) {
			path = apiPaths[t]
		}
	}
	if path == "" {
		// TODO: I don't thing we should default to this, but I kept it
		// for compatability and because this function can't return an error
		path = "things"
	}
	return base + path + "/" + uuid
}

// IDURL - Adds the appropriate prefix e.g http://api.ft.com/things/
func IDURL(uuid string) string {
	return "http://api.ft.com/things/" + uuid
}

// TypeURIs - Builds up the type URI based on type e.g http://www.ft.com/ontology/Person
func TypeURIs(labels []string) []string {
	var results []string
	sorted, err := SortTypes(labels)
	if err != nil {
		log.Printf("ERROR - %v", err)
		return []string{}
	}
	for _, label := range sorted {
		uri := typeURIs[label]
		if uri != "" {
			results = append(results, uri)
		}
	}
	return results
}
