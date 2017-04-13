package annotations

import (

	"fmt"
	"strings"
	"reflect"
)

type filteredPredicateNames struct {
	MENTIONS string `default: "mentions"`
	MAJOR_MENTIONS string `default: "major_mentions"`
	ABOUT string `default: "about"`
	IS_CLASSIFIED_BY string `default: "is_classified_by"`
	IS_PRIMARILY_CLASSIFIED_BY string `default: "is_primarily_classified_by"`

}

func newFilteredPredicateNames() *filteredPredicateNames {
	v := filteredPredicateNames{}
	v.MENTIONS = "mentions"
	v.MAJOR_MENTIONS = "major_mentions"
	v.ABOUT = "about"
	v.IS_CLASSIFIED_BY = "is_classified_by"
	v.IS_PRIMARILY_CLASSIFIED_BY = "is_primarily_classified_by"
	return &v
}

func (f *filteredPredicateNames) contains(pred string) bool {
	enum := newFilteredPredicateNames()
	v := reflect.ValueOf(*enum)
	for i :=0; i < v.NumField(); i++ {
		//fmt.Println(v.Field(i))
		if(v.Field(i).String() == pred) {
			return true
		}
	}
	return false
}

type AnnotationsFilter struct {
	//definition of predicate groups to whom Rule of Importance should be applied
	ImportanceRuleConfig  [][]string
	//predicate names of annotations that should be considered for filtering
	enum                  *filteredPredicateNames
	//stores annotations to be filtered keyed by concept ID (uuid)
	unfilteredAnnotations map[string][]annotation
	//stored annotations not to be filtered keyed by concept ID (uuid)
	filteredAnnotations   map[string][]annotation
}

func NewAnnotationsFilter() *AnnotationsFilter {
	f := AnnotationsFilter{}
	f.enum = newFilteredPredicateNames()
	f.ImportanceRuleConfig = [][]string{
		{f.enum.MENTIONS,f.enum.MAJOR_MENTIONS, f.enum.ABOUT},
		{f.enum.IS_CLASSIFIED_BY, f.enum.IS_PRIMARILY_CLASSIFIED_BY},
	}
	f.filteredAnnotations = make(map[string][]annotation)
	f.unfilteredAnnotations =  make(map[string][]annotation)
	return &f
}
func (f *AnnotationsFilter) addFiltered(a annotation) {
	if f.filteredAnnotations[a.ID] == nil {
		//for each importance group we shell store 1 most important annotation
		f.filteredAnnotations[a.ID] = make([]annotation, len(f.ImportanceRuleConfig ))
		//fmt.Printf("\n initialised filtered annotations: %v \n", f.filteredAnnotations)
	}
	grpId, pos := f.getGroupIdAndImportanceValue(strings.ToLower(a.Predicate))
	if grpId == -1 || pos == -1  {
		fmt.Printf("could not find group for predicate %s \n", strings.ToLower(a.Predicate))
		return
	}
	arr := f.filteredAnnotations[a.ID]
	prevAnnotation := arr[grpId]
	//empty value indicates we have not seen annotations for this group before
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
	f.unfilteredAnnotations[a.ID] = append(f.unfilteredAnnotations[a.ID], a )
}

func (f *AnnotationsFilter) Add(a annotation) {
	//fmt.Printf("\n processing predicate: %s \n", strings.ToLower(a.Predicate))
	if f.enum.contains(strings.ToLower(a.Predicate)) {
		f.addFiltered(a)

	} else {
		f.addUnfiltered(a)
	}
}
func (f *AnnotationsFilter) getGroupIdAndImportanceValue(predicate string) (int, int) {
	for group, s := range f.ImportanceRuleConfig {
		for pos, val := range s {
			if val == predicate {
				return group, pos
			}
		}
	}
	return -1, -1
}

func (f *AnnotationsFilter) getImportanceValueForGroupId(predicate string, groupId int) int {
	for pos, val := range f.ImportanceRuleConfig[groupId] {
		if val == predicate {
			return pos
		}
	}
	return  -1
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

	//fmt.Printf("uuid:  %v \n",  out)
	return out
}

