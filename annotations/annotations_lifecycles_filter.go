package annotations

const (
	pacLifecycle = "annotations-pac"
	v2Lifecycle  = "annotations-v2"
)

var lifecycleMap = map[string]string{
	"next-video": "annotations-next-video",
	"v1":         "annotations-v1",
	"pac":        "annotations-pac",
	"v2":         "annotations-v2",
}

type lifecycleFilter struct {
	lifecycles []string
}

func newLifecycleFilter(opts ...func(*lifecycleFilter)) annotationsFilter {
	lf := lifecycleFilter{}
	for _, opt := range opts {
		opt(&lf)
	}

	return &lf
}

func withLifecycles(lifecycles []string) func(*lifecycleFilter) {
	return func(f *lifecycleFilter) {
		f.lifecycles = lifecycles
	}
}

func (f *lifecycleFilter) filter(annotations []annotation, chain *annotationsFilterChain) []annotation {
	if containsPACLifecycle(annotations) {
		filtered := filterPACAndV2Lifecycles(annotations)
		return chain.doNext(f.applyAdditionalFiltering(filtered))
	}

	return chain.doNext(f.applyAdditionalFiltering(annotations))
}

func (f *lifecycleFilter) applyAdditionalFiltering(annotations []annotation) []annotation {
	if len(f.lifecycles) == 0 {
		return annotations
	}

	var filtered []annotation
	for _, annotation := range annotations {
		for _, lc := range f.lifecycles {
			if annotation.Lifecycle == lifecycleMap[lc] {
				filtered = append(filtered, annotation)
			}
		}
	}
	return filtered
}

func containsPACLifecycle(annotations []annotation) bool {
	for _, annotation := range annotations {
		if annotation.Lifecycle == pacLifecycle {
			return true
		}
	}
	return false
}

func filterPACAndV2Lifecycles(annotations []annotation) []annotation {
	var filtered []annotation
	for _, annotation := range annotations {
		if annotation.Lifecycle == pacLifecycle || annotation.Lifecycle == v2Lifecycle {
			filtered = append(filtered, annotation)
		}
	}
	return filtered
}
