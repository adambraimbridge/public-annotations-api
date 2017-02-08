package mapper

import (
	"errors"
	"sort"
)

var parentTypes = map[string]string{
	"Thing":                  "", // this is here to enable iterating over map keys to get all types
	"Concept":                "Thing",
	"Role":                   "Thing",
	"BoardRole":              "Role",
	"Classification":         "Concept",
	"Person":                 "Concept",
	"Organisation":           "Concept",
	"Membership":             "Concept",
	"Company":                "Organisation",
	"PublicCompany":          "Company",
	"PrivateCompany":         "Company",
	"Brand":                  "Classification",
	"Subject":                "Classification",
	"Section":                "Classification",
	"Genre":                  "Classification",
	"Location":               "Concept",
	"IndustryClassification": "Classification",
	"Topic":                  "Concept",
	"Content":                "Thing",
	"SpecialReport":          "Classification",
	"AlphavilleSeries":       "Classification",
	"FinancialInstrument":    "Concept",
}

// ParentType returns the immediate parent type for a given Type
func ParentType(t string) string {
	return parentTypes[t]
}

func isDescendent(descendent, ancestor string) bool {
	for t := descendent; t != ""; t = ParentType(t) {
		if t == ancestor {
			return true
		}
	}
	return false
}

// MostSpecific returns the most specific from a list of types in an hierarchy
// behaviour is undefined if any of the types are siblings.
func mostSpecific(types []string) (string, error) {
	if len(types) == 0 {
		return "", errors.New("no types supplied")
	}
	sorted, err := SortTypes(types)
	if err != nil {
		return "", err
	}
	return sorted[len(sorted)-1], nil
}

var ErrNotHierarchy = errors.New("provided types are not a consistent hierarchy")

// SortTypes sorts the given types from least specific to most specific
func SortTypes(types []string) ([]string, error) {
	ts := &typeSorter{types: make([]string, len(types))}
	copy(ts.types, types)
	sort.Sort(ts)
	if ts.invalid {
		return types, ErrNotHierarchy
	}
	return ts.types, nil
}

type typeSorter struct {
	types   []string
	invalid bool
}

func (ts *typeSorter) Len() int {
	return len(ts.types)
}

func (ts *typeSorter) Less(a, b int) bool {
	at := ts.types[a]
	bt := ts.types[b]
	if isDescendent(bt, at) {
		return true
	}
	if !isDescendent(at, bt) {
		ts.invalid = true
	}
	return false
}

func (ts *typeSorter) Swap(a, b int) {
	ts.types[a], ts.types[b] = ts.types[b], ts.types[a]
}
