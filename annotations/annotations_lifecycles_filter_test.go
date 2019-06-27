package annotations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	v1Lifecycle        = "annotations-v1"
	nextVideoLifecycle = "annotations-next-video"
)

var pacAnnotationA = annotation{
	ID:        "6bbd0457-15ab-4ddc-ab82-0cd5b8d9ce18",
	Predicate: ABOUT,
	Lifecycle: pacLifecycle,
}

var pacAnnotationB = annotation{
	ID:        "0ab61bfc-a2b1-4b08-a864-4233fd72f250",
	Predicate: MENTIONS,
	Lifecycle: pacLifecycle,
}

var v1AnnotationA = annotation{
	ID:        "a0076026-f2e5-414f-b7a0-419bc16c4c51",
	Predicate: ABOUT,
	Lifecycle: v1Lifecycle,
}

var v1AnnotationB = annotation{
	ID:        "2ddd7896-b6c5-4726-846e-2e842a3f2aea",
	Predicate: MENTIONS,
	Lifecycle: v1Lifecycle,
}

var v2AnnotationA = annotation{
	ID:        "8886a23b-c3ee-49cc-813a-94292176ce8a",
	Predicate: ABOUT,
	Lifecycle: v2Lifecycle,
}

var v2AnnotationB = annotation{
	ID:        "6e416a42-6f49-420b-9209-faf123e6ff08",
	Predicate: MENTIONS,
	Lifecycle: v2Lifecycle,
}

var nextVideoAnnotationA = annotation{
	ID:        "f00adf2e-6a59-4e2e-8a18-4d63ae0a689f",
	Predicate: ABOUT,
	Lifecycle: nextVideoLifecycle,
}

var nextVideoAnnotationB = annotation{
	ID:        "0d0e6957-cdb4-40cf-a3a5-c61665680eb8",
	Predicate: MENTIONS,
	Lifecycle: nextVideoLifecycle,
}

func TestFilterOnPACAnnotationsOnly(t *testing.T) {
	annotations := []annotation{pacAnnotationA, pacAnnotationB}
	f := newLifecycleFilter()
	chain := newAnnotationsFilterChain(f)
	filtered := chain.doNext(annotations)

	assert.Len(t, filtered, 2)
	assert.Contains(t, filtered, pacAnnotationA)
	assert.Contains(t, filtered, pacAnnotationB)
}

func TestFilterOnV1AnnotationsOnly(t *testing.T) {
	annotations := []annotation{v1AnnotationA, v1AnnotationB}
	f := newLifecycleFilter()
	chain := newAnnotationsFilterChain(f)
	filtered := chain.doNext(annotations)

	assert.Len(t, filtered, 2)
	assert.Contains(t, filtered, v1AnnotationA)
	assert.Contains(t, filtered, v1AnnotationB)
}

func TestFilterOnV2AnnotationsOnly(t *testing.T) {
	annotations := []annotation{v2AnnotationA, v2AnnotationB}
	f := newLifecycleFilter()
	chain := newAnnotationsFilterChain(f)
	filtered := chain.doNext(annotations)

	assert.Len(t, filtered, 2)
	assert.Contains(t, filtered, v2AnnotationA)
	assert.Contains(t, filtered, v2AnnotationB)
}

func TestFilterOnVideoAnnotationsOnly(t *testing.T) {
	annotations := []annotation{nextVideoAnnotationA, nextVideoAnnotationB}
	f := newLifecycleFilter()
	chain := newAnnotationsFilterChain(f)
	filtered := chain.doNext(annotations)

	assert.Len(t, filtered, 2)
	assert.Contains(t, filtered, nextVideoAnnotationA)
	assert.Contains(t, filtered, nextVideoAnnotationB)
}

func TestFilterOnPACV2Annotations(t *testing.T) {
	annotations := []annotation{pacAnnotationA, pacAnnotationB, v2AnnotationA, v2AnnotationB}
	f := newLifecycleFilter()
	chain := newAnnotationsFilterChain(f)
	filtered := chain.doNext(annotations)

	assert.Len(t, filtered, 4)
	assert.Contains(t, filtered, pacAnnotationA)
	assert.Contains(t, filtered, pacAnnotationB)
	assert.Contains(t, filtered, v2AnnotationA)
	assert.Contains(t, filtered, v2AnnotationB)
}

func TestFilterOnV1V2Annotations(t *testing.T) {
	annotations := []annotation{v1AnnotationA, v1AnnotationB, v2AnnotationA, v2AnnotationB}
	f := newLifecycleFilter()
	chain := newAnnotationsFilterChain(f)
	filtered := chain.doNext(annotations)

	assert.Len(t, filtered, 4)
	assert.Contains(t, filtered, v1AnnotationA)
	assert.Contains(t, filtered, v1AnnotationB)
	assert.Contains(t, filtered, v2AnnotationA)
	assert.Contains(t, filtered, v2AnnotationB)
}

func TestFilterOnV1PACAnnotations(t *testing.T) {
	annotations := []annotation{pacAnnotationA, pacAnnotationB, v1AnnotationA, v1AnnotationB}
	f := newLifecycleFilter()
	chain := newAnnotationsFilterChain(f)
	filtered := chain.doNext(annotations)

	assert.Len(t, filtered, 2)
	assert.Contains(t, filtered, pacAnnotationA)
	assert.Contains(t, filtered, pacAnnotationB)
}

func TestFilterOnVideoPACAnnotations(t *testing.T) {
	annotations := []annotation{pacAnnotationA, pacAnnotationB, nextVideoAnnotationA, nextVideoAnnotationB}
	f := newLifecycleFilter()
	chain := newAnnotationsFilterChain(f)
	filtered := chain.doNext(annotations)

	assert.Len(t, filtered, 2)
	assert.Contains(t, filtered, pacAnnotationA)
	assert.Contains(t, filtered, pacAnnotationB)
}

func TestFilterOnV1V2PACAnnotations(t *testing.T) {
	annotations := []annotation{
		pacAnnotationA,
		pacAnnotationB,
		v1AnnotationA,
		v1AnnotationB,
		v2AnnotationA,
		v2AnnotationB,
	}
	f := newLifecycleFilter()
	chain := newAnnotationsFilterChain(f)
	filtered := chain.doNext(annotations)

	assert.Len(t, filtered, 4)
	assert.Contains(t, filtered, pacAnnotationA)
	assert.Contains(t, filtered, pacAnnotationB)
	assert.Contains(t, filtered, v2AnnotationA)
	assert.Contains(t, filtered, v2AnnotationA)
}

func TestAdditionalFilteringOnV1V2PACAnnotations(t *testing.T) {
	tests := map[string]struct {
		lifecycles []string
		expected   []annotation
	}{
		"additional pac filtering should return only pac annotations": {
			lifecycles: []string{"pac"},
			expected:   []annotation{pacAnnotationA, pacAnnotationB},
		},
		"additional v2 filtering should return only v2 annotations": {
			lifecycles: []string{"v2"},
			expected:   []annotation{v2AnnotationA, v2AnnotationB},
		},
		"additional v1 filtering should return nil": {
			lifecycles: []string{"v1"},
			expected:   nil,
		},
		"additional next-video filtering should return nil": {
			lifecycles: []string{"next-video"},
			expected:   nil,
		},
		"additional v1&next-video filtering should return nil": {
			lifecycles: []string{"v1", "next-video"},
			expected:   nil,
		},
		"additional pac&v2 filtering should return pac&v2 annotations": {
			lifecycles: []string{"pac", "v2"},
			expected:   []annotation{pacAnnotationA, pacAnnotationB, v2AnnotationA, v2AnnotationB},
		},
		"additional pac&v1&v2&next-video filtering should return pac&v2 annotations": {
			lifecycles: []string{"pac", "v1", "v2", "next-video"},
			expected:   []annotation{pacAnnotationA, pacAnnotationB, v2AnnotationA, v2AnnotationB},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			annotations := []annotation{pacAnnotationA, pacAnnotationB, v1AnnotationA, v1AnnotationB, v2AnnotationA, v2AnnotationB}
			f := newLifecycleFilter(withLifecycles(tc.lifecycles))
			chain := newAnnotationsFilterChain(f)
			filtered := chain.doNext(annotations)

			assert.Len(t, filtered, len(tc.expected))
			assert.Equal(t, filtered, tc.expected)
		})
	}
}

func TestAdditionalFilteringNoPACAnnotations(t *testing.T) {
	tests := map[string]struct {
		lifecycles []string
		expected   []annotation
	}{
		"additional v1 filtering should return only v1 annotations": {
			lifecycles: []string{"v1"},
			expected:   []annotation{v1AnnotationA, v1AnnotationB},
		},
		"additional v2 filtering should return only v2 annotations": {
			lifecycles: []string{"v2"},
			expected:   []annotation{v2AnnotationA, v2AnnotationB},
		},
		"additional next-video filtering should return only next-video annotations": {
			lifecycles: []string{"next-video"},
			expected:   []annotation{nextVideoAnnotationA, nextVideoAnnotationB},
		},
		"additional v1&v2 filtering should return v1&v2 annotations": {
			lifecycles: []string{"v1", "v2"},
			expected:   []annotation{v1AnnotationA, v1AnnotationB, v2AnnotationA, v2AnnotationB},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			annotations := []annotation{v1AnnotationA, v1AnnotationB, v2AnnotationA, v2AnnotationB, nextVideoAnnotationA, nextVideoAnnotationB}
			f := newLifecycleFilter(withLifecycles(tc.lifecycles))
			chain := newAnnotationsFilterChain(f)
			filtered := chain.doNext(annotations)

			assert.Len(t, filtered, len(tc.expected))
			assert.Equal(t, filtered, tc.expected)
		})
	}
}
