package annotations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLifecycleFilter(t *testing.T) {
	f := newLifecycleFilter("foo")
	chain := newAnnotationsFilterChain(f)

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
	f := defaultDedupFilter
	chain := newAnnotationsFilterChain(f)

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

func TestDedupFilterDedups(t *testing.T) {
	f := defaultDedupFilter
	chain := newAnnotationsFilterChain(f)

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
