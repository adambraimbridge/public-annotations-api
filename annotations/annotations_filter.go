package annotations

type annotationsFilter interface {
	filter(ann []annotation, chain *annotationsFilterChain) []annotation
}

type annotationsFilterChain struct {
	index   int
	filters []annotationsFilter
}

func newAnnotationsFilterChain(filters ...annotationsFilter) *annotationsFilterChain {
	size := len(filters)
	f := make([]annotationsFilter, size+1)
	for i, t := range filters {
		f[i] = t
	}
	f[size] = defaultDedupFilter
	return &annotationsFilterChain{0, f}
}

func (chain *annotationsFilterChain) doNext(ann []annotation) []annotation {
	if chain.index < len(chain.filters) {
		f := chain.filters[chain.index]
		chain.index++

		ann = f.filter(ann, chain)
	}

	return ann
}

type lifecycleFilter struct {
	lifecycle string
}

func newLifecycleFilter(lifecycle string) annotationsFilter {
	return &lifecycleFilter{lifecycle}
}

func (f *lifecycleFilter) filter(annotations []annotation, chain *annotationsFilterChain) []annotation {
	filtered := []annotation{}
	for _, annotation := range annotations {
		if annotation.Lifecycle == f.lifecycle {
			filtered = append(filtered, annotation)
		}
	}
	return chain.doNext(filtered)
}

type dedupFilter struct {
}

var defaultDedupFilter = &dedupFilter{}

func (f *dedupFilter) filter(in []annotation, chain *annotationsFilterChain) []annotation {
	out := []annotation{}

OUTER:
	for _, ann := range in {
		for _, copied := range out {
			if copied.Predicate == ann.Predicate && copied.ID == ann.ID {
				continue OUTER
			}
		}

		out = append(out, ann)
	}

	return chain.doNext(out)
}
