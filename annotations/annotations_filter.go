package annotations

import (
	"reflect"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// Defines all names of predicates that have to be considered by the annotation filter.
// Predicates that are not defined in FilteredPredicateNames are not filtered.
type filteredPredicateNames struct {
	MENTIONS                   string
	MAJOR_MENTIONS             string
	ABOUT                      string
	IS_CLASSIFIED_BY           string
	IS_PRIMARILY_CLASSIFIED_BY string
}

func newFilteredPredicateNames() *filteredPredicateNames {
	v := filteredPredicateNames{}
	v.MENTIONS = "http://www.ft.com/ontology/annotation/mentions"
	v.MAJOR_MENTIONS = "http://www.ft.com/ontology/annotation/majormentions"
	v.ABOUT = "http://www.ft.com/ontology/annotation/about"
	v.IS_CLASSIFIED_BY = "http://www.ft.com/ontology/classification/isclassifiedby"
	v.IS_PRIMARILY_CLASSIFIED_BY = "http://www.ft.com/ontology/classification/isprimarilyclassifiedby"
	return &v
}

func (f *filteredPredicateNames) contains(pred string) bool {
	enum := newFilteredPredicateNames()
	v := reflect.ValueOf(*enum)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).String() == pred {
			return true
		}
	}
	return false
}

type AnnotationsFilter struct {
	// Definition of predicate groups to whom Rule of Importance should be applied.
	// Each group contains a list of predicate names in the order of increasing importance.
	ImportanceRuleConfig [][]string
	// Predicate names of annotations that should be considered for filtering
	enum *filteredPredicateNames
	// Stores annotations to be filtered keyed by concept ID (uuid).
	unfilteredAnnotations map[string][]annotation
	// Stores annotations not to be filtered keyed by concept ID (uuid).
	filteredAnnotations map[string][]annotation
}

func NewAnnotationsFilter() *AnnotationsFilter {
	f := AnnotationsFilter{}
	f.enum = newFilteredPredicateNames()
	// Configure groups of predicates that should be filtered according to their importance.
	f.ImportanceRuleConfig = [][]string{
		{f.enum.MENTIONS, f.enum.MAJOR_MENTIONS, f.enum.ABOUT},
		{f.enum.IS_CLASSIFIED_BY, f.enum.IS_PRIMARILY_CLASSIFIED_BY},
	}
	f.filteredAnnotations = make(map[string][]annotation)
	f.unfilteredAnnotations = make(map[string][]annotation)
	return &f
}

func (f *AnnotationsFilter) Add(a annotation) {
	if f.enum.contains(strings.ToLower(a.Predicate)) {
		f.addFiltered(a)
	} else {
		f.addUnfiltered(a)
	}
}

func (f *AnnotationsFilter) Filter() []annotation {
	out := []annotation{}

	for _, allFiltered := range f.filteredAnnotations {
		for _, a := range allFiltered {
			if a.ID != "" {
				out = append(out, a)
			}
		}
	}

	for _, allUnfiltered := range f.unfilteredAnnotations {
		out = append(out, allUnfiltered...)
	}
	return out
}

func (f *AnnotationsFilter) addFiltered(a annotation) {
	if f.filteredAnnotations[a.ID] == nil {
		// For each importance group we shell store 1 most important annotation
		f.filteredAnnotations[a.ID] = make([]annotation, len(f.ImportanceRuleConfig))
	}
	grpId, pos := f.getGroupIdAndImportanceValue(strings.ToLower(a.Predicate))
	if grpId == -1 || pos == -1 {
		log.Debugf("Could not find group for predicate %s \n", strings.ToLower(a.Predicate))
		return
	}
	arr := f.filteredAnnotations[a.ID]
	prevAnnotation := arr[grpId]
	// Empty value indicates we have not seen annotations for this group before.
	if prevAnnotation.ID == "" {
		f.filteredAnnotations[a.ID][grpId] = a
	} else {
		prevPos := f.getImportanceValueForGroupId(strings.ToLower(prevAnnotation.Predicate), grpId)
		if prevPos < pos {
			f.filteredAnnotations[a.ID][grpId] = a
		}
	}
}

func (f *AnnotationsFilter) addUnfiltered(a annotation) {
	if f.unfilteredAnnotations[a.ID] == nil {
		f.unfilteredAnnotations[a.ID] = []annotation{}
	}
	f.unfilteredAnnotations[a.ID] = append(f.unfilteredAnnotations[a.ID], a)
}

func (f *AnnotationsFilter) getGroupIdAndImportanceValue(predicate string) (int, int) {
	for group, s := range f.ImportanceRuleConfig {
		for pos, val := range s {
			if val == predicate {
				return group, pos
			}
		}
	}
	//should not occur in normal circumstances
	return -1, -1
}

func (f *AnnotationsFilter) getImportanceValueForGroupId(predicate string, groupId int) int {
	for pos, val := range f.ImportanceRuleConfig[groupId] {
		if val == predicate {
			return pos
		}
	}
	//should not occur in normal circumstances
	return -1
}
