package annotations

const pacLifecycle = "annotations-pac"
const v2Lifecycle = "annotations-v2"

type lifecycleFilter struct{}

func newLifecycleFilter() annotationsFilter {
	return &lifecycleFilter{}
}

func (f *lifecycleFilter) filter(annotations []annotation, chain *annotationsFilterChain) []annotation {
	if containsPACLifecycle(annotations) {
		filtered := filterPACAndV2Lifecycles(annotations)
		return chain.doNext(filtered)
	}
	return chain.doNext(annotations)
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
