package broker

import (
	"slices"
	"strconv"
	"strings"

	"genaiz.com/genaiz/task/shared"
)

type Broker struct {
	AuthFile string
	HostAddr string
}

type Function struct {
	Id          int // Id is assigned by a publishing Broker and refers to the Smart Function release cycle
	Arches      []string
	Description string
	Fqdn        string
	Handle      string
	Img         string
	Digest      string
	Name        string
	Oem         string
	Type        string
	Version     string
}

func (f Function) asIdentity() *shared.Identity {
	return &shared.Identity{
		Id:      strconv.Itoa(f.Id),
		Hash:    f.Digest,
		Path:    f.Img,
		Version: f.Version,
	}
}

type Provision struct {
	Auth string
	Sf   Function
}

type Session struct {
	Id     int64
	Nco    int64
	Nms    int64
	Flags  int
	UserId int
	Expiry int64
}

type Solution struct {
	Description string
	Handle      string
	Name        string
	Oem         string
	Version     string
	Workflows   []Workflow `yaml:"workflows"`
}

func (s Solution) Merge(update Solution) *Solution {
	var result = &Solution{}

	if update.Description != "" && !strings.EqualFold(update.Description, s.Description) {
		result.Description = update.Description
	} else {
		result.Description = s.Description
	}

	if update.Name != "" && !strings.EqualFold(update.Name, s.Name) {
		result.Name = update.Name
	} else {
		result.Name = s.Name
	}

	if update.Version != "" && !strings.EqualFold(update.Version, s.Version) {
		result.Version = update.Version
	} else {
		result.Version = s.Version
	}

	for _, wf := range s.Workflows {
		result.Workflows = append(result.Workflows, wf)
	}

	for _, wf := range update.Workflows {
		if !slices.ContainsFunc(result.Workflows, WorkflowHandlePredicate(wf.Handle)) {
			result.Workflows = append(result.Workflows, wf)
		}
	}

	result.Handle = s.Handle
	result.Oem = s.Oem
	return result
}

type Workflow struct {
	Name        string `yaml:"name"`
	Description string
	Handle      string
	Links       []WorkflowLink
	Nodes       []WorkflowNode
}

type WorkflowLink struct {
	LhsNode     string
	LhsNodePort string
	RhsNode     string
	RhsNodePort string
}

func (wl WorkflowLink) Equals(wl2 WorkflowLink) bool {
	return strings.EqualFold(wl.LhsNode, wl2.LhsNode) &&
		strings.EqualFold(wl.LhsNodePort, wl2.LhsNodePort) &&
		strings.EqualFold(wl.RhsNode, wl2.RhsNode) &&
		strings.EqualFold(wl.RhsNodePort, wl2.RhsNodePort)
}

type WorkflowNode struct {
	Name        string
	Description string
	Handle      string
	Sf          *WorkflowNodeFunction
}

func (wn WorkflowNode) Equals(wn2 WorkflowNode) bool {
	return strings.EqualFold(wn.Handle, wn2.Handle)
}

type WorkflowNodeFunction struct {
	Oem     string
	Handle  string
	Version string
	Seq     int
}

func WorkflowHandlePredicate(handle string) func(Workflow) bool {
	return func(wf Workflow) bool {
		return strings.EqualFold(wf.Handle, handle)
	}
}

func WorkflowNamePredicate(name string) func(Workflow) bool {
	return func(wf Workflow) bool {
		return strings.EqualFold(wf.Name, name)
	}
}
