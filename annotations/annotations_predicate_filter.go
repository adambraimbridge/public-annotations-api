package annotations

import (
	"strings"
)

// Defines all names of predicates that have to be considered by the annotation filter.
// Predicates that are not defined in FilteredPredicateNames are not filtered.

const (
	Mentions                = "http://www.ft.com/ontology/annotation/mentions"
	MajorMentions           = "http://www.ft.com/ontology/annotation/majormentions"
	About                   = "http://www.ft.com/ontology/annotation/about"
	IsClassifiedBy          = "http://www.ft.com/ontology/classification/isclassifiedby"
	IsPrimarilyClassifiedBy = "http://www.ft.com/ontology/classification/isprimarilyclassifiedby"
)

type AnnotationsPredicateFilter struct {
	// Definition of predicate groups to whom Rule of Importance should be applied.
	// Each group contains a list of predicate names in the order of increasing importance.
	ImportanceRuleConfig [][]string
	// Predicate names of annotations that should be considered for filtering
	enum []string
	// Stores annotations to be filtered keyed by concept ID (uuid).
	unfilteredAnnotations map[string][]annotation
	// Stores annotations not to be filtered keyed by concept ID (uuid).
	filteredAnnotations map[string][]annotation
}

func NewAnnotationsPredicateFilter() *AnnotationsPredicateFilter {
	return &AnnotationsPredicateFilter{
		enum: []string{
			Mentions,
			MajorMentions,
			About,
			IsClassifiedBy,
			IsPrimarilyClassifiedBy,
		},
		// Configure groups of predicates that should be filtered according to their importance.
		ImportanceRuleConfig: [][]string{
			{
				Mentions,
				MajorMentions,
				About,
			},
			{
				IsClassifiedBy,
				IsPrimarilyClassifiedBy,
			},
		},
		filteredAnnotations:   make(map[string][]annotation),
		unfilteredAnnotations: make(map[string][]annotation),
	}
}

func (f *AnnotationsPredicateFilter) FilterAnnotations(annotations []annotation) {
	for _, ann := range annotations {
		f.Add(ann)
	}
}

func (f *AnnotationsPredicateFilter) Add(a annotation) {

	pred := strings.ToLower(a.Predicate)
	for _, p := range f.enum {
		if p == pred {
			f.addFiltered(a)
			return
		}
	}

	f.addUnfiltered(a)
}

func (f *AnnotationsPredicateFilter) ProduceResponseList() []annotation {
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

func (f *AnnotationsPredicateFilter) addFiltered(a annotation) {
	if f.filteredAnnotations[a.ID] == nil {
		// For each importance group we shell store 1 most important annotation
		f.filteredAnnotations[a.ID] = make([]annotation, len(f.ImportanceRuleConfig))
	}
	grpID, pos := f.getGroupIDAndImportanceValue(strings.ToLower(a.Predicate))
	if grpID == -1 || pos == -1 {
		return
	}
	arr := f.filteredAnnotations[a.ID]
	prevAnnotation := arr[grpID]
	// Empty value indicates we have not seen annotations for this group before.
	if prevAnnotation.ID == "" {
		f.filteredAnnotations[a.ID][grpID] = a
	} else {
		prevPos := f.getImportanceValueForGroupID(strings.ToLower(prevAnnotation.Predicate), grpID)
		if prevPos < pos {
			f.filteredAnnotations[a.ID][grpID] = a
		}
	}
}

func (f *AnnotationsPredicateFilter) addUnfiltered(a annotation) {
	if f.unfilteredAnnotations[a.ID] == nil {
		f.unfilteredAnnotations[a.ID] = []annotation{}
	}
	f.unfilteredAnnotations[a.ID] = append(f.unfilteredAnnotations[a.ID], a)
}

func (f *AnnotationsPredicateFilter) getGroupIDAndImportanceValue(predicate string) (int, int) {
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

func (f *AnnotationsPredicateFilter) getImportanceValueForGroupID(predicate string, groupId int) int {
	for pos, val := range f.ImportanceRuleConfig[groupId] {
		if val == predicate {
			return pos
		}
	}
	//should not occur in normal circumstances
	return -1
}

func (f *AnnotationsPredicateFilter) filter(in []annotation, chain *annotationsFilterChain) []annotation {
	f.FilterAnnotations(in)
	return chain.doNext(f.ProduceResponseList())
}
