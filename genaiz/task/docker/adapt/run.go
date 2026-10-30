package adapt

import "github.com/moby/moby/api/types/container"

type RunArguments struct {
	Attached    bool
	Dispose     bool
	Interactive bool
}

// WaitCondition adapts the wait condition on a client.ContainAttach to automatically remove the container on exit
func (r RunArguments) WaitCondition() container.WaitCondition {
	if r.Dispose {
		return container.WaitConditionRemoved
	}

	return container.WaitConditionNextExit
}
