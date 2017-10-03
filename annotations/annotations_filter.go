package annotations

type annotationsFilter interface {
	filter(ann []annotation, chain *annotationsFilterChain) []annotation
}

type annotationsFilterChain struct {
	filters []annotationsFilter
	index int
}

func newAnnotationsFilterChain(filters []annotationsFilter) *annotationsFilterChain {
	return &annotationsFilterChain{filters, 0}
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
	predicateToRetain string
	inFavourOf        string
}

func newDedupFilter(retain string, inFavourOf string) annotationsFilter {
	return &dedupFilter{retain, inFavourOf}
}

func (f *dedupFilter) filter(in []annotation, chain *annotationsFilterChain) []annotation {
	concepts := map[string]struct{}{}

	out := []annotation{}

	for _, ann := range in {
		if ann.Predicate == f.predicateToRetain {
			concepts[ann.ID] = struct{}{}
		}
	}

	OUTER:
	for _, ann := range in {
		if ann.Predicate == f.inFavourOf {
			if _, duplicated := concepts[ann.ID]; duplicated {
				continue
			}
		}

		for _, copied := range out {
			if copied.Predicate == ann.Predicate && copied.ID == ann.ID {
				continue OUTER
			}
		}

		out = append(out, ann)
	}

	return chain.doNext(out)
}
