package annotations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLifecycleFilter(t *testing.T) {
	f := newLifecycleFilter("foo")
	chain := newAnnotationsFilterChain([]annotationsFilter{f})

	ann := []annotation{
		{
			Lifecycle:"foo",
		},
		{
			Lifecycle:"bar",
		},
	}

	actual := chain.doNext(ann)

	assert.Len(t, actual, 1)
	assert.Equal(t, ann[0], actual[0], "filtered annotations")
}

func TestDedupFilterPassthrough(t *testing.T) {
	f := newDedupFilter("baz", "bar")
	chain := newAnnotationsFilterChain([]annotationsFilter{f})

	ann := []annotation{
		{
			ID: "1",
			Predicate:"foo",
		},
	}

	actual := chain.doNext(ann)

	assert.Len(t, actual, 1)
	assert.Equal(t, ann[0], actual[0], "pass-through predicate")
}

func TestDedupFilterDropAnnotation(t *testing.T) {
	f := newDedupFilter("baz", "bar")
	chain := newAnnotationsFilterChain([]annotationsFilter{f})

	ann := []annotation{
		{
			ID: "2",
			Predicate:"bar",
		},
		{
			ID: "2",
			Predicate:"baz",
		},
	}

	actual := chain.doNext(ann)

	assert.Len(t, actual, 1)
	assert.Equal(t, ann[1], actual[0], "filtered annotations")
}

func TestDedupFilterRetainsOtherPredicatesForConcept(t *testing.T) {
	f := newDedupFilter("baz", "bar")
	chain := newAnnotationsFilterChain([]annotationsFilter{f})

	ann := []annotation{
		{
			ID: "2",
			Predicate:"foo",
		},
		{
			ID: "2",
			Predicate:"baz",
		},
	}

	actual := chain.doNext(ann)

	assert.Len(t, actual, 2)
	assert.Contains(t, actual, ann[0])
	assert.Contains(t, actual, ann[1])
}

func TestDedupFilterRetainsDifferentConcepts(t *testing.T) {
	f := newDedupFilter("baz", "bar")
	chain := newAnnotationsFilterChain([]annotationsFilter{f})

	ann := []annotation{
		{
			ID: "2",
			Predicate:"bar",
		},
		{
			ID: "3",
			Predicate:"baz",
		},
	}

	actual := chain.doNext(ann)

	assert.Len(t, actual, 2)
	assert.Contains(t, actual, ann[0])
	assert.Contains(t, actual, ann[1])
}

func TestDedupFilterDedups(t *testing.T) {
	f := newDedupFilter("baz", "bar")
	chain := newAnnotationsFilterChain([]annotationsFilter{f})

	ann := []annotation{
		{
			ID: "2",
			Predicate:"baz",
		},
		{
			ID: "2",
			Predicate:"baz",
		},
	}

	actual := chain.doNext(ann)

	assert.Len(t, actual, 1)
	assert.Equal(t, actual[0].ID, "2", "concept id")
	assert.Equal(t, actual[0].Predicate, "baz", "predicate")
}
