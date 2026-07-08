package broker

import (
	"maps"
	"slices"
)

type Hierarchical interface {
	// GetBranch returns the Hierarchy's anchor string, all matching branches are part of the Hierarchy
	GetBranch() string

	// IsReleased denotes elements in a Hierarchy which released and are now the final leaf of the branch
	IsReleased() bool
}

type Hierarchy[H Hierarchical] struct {
	branchMap map[string]H
}

func (h *Hierarchy[H]) reduce(local []H) []H {
	var result []H

	for _, value := range slices.Collect(maps.Values(h.branchMap)) {
		if value.IsReleased() {
			result = append(result, value)
		} else if _, ok := h.branchMap[value.GetBranch()]; ok {
			result = append(result, value)
		}
	}

	for _, l := range local {
		if _, ok := h.branchMap[l.GetBranch()]; !ok {
			result = append(result, l)
		}
	}

	return result
}

func NewHierarchy[H Hierarchical](hierarchical []H, isAfter func(H, H) bool) *Hierarchy[H] {
	var result = &Hierarchy[H]{
		branchMap: make(map[string]H),
	}

	for _, h := range hierarchical {
		if h.IsReleased() {
			result.branchMap[h.GetBranch()] = h
		} else if latest, ok := result.branchMap[h.GetBranch()]; !ok || isAfter(h, latest) {
			result.branchMap[h.GetBranch()] = h
		}
	}

	return result
}
