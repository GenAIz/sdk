package broker

import (
	"slices"
	"strconv"
	"strings"

	"genaiz.com/genaiz/task/shared"
)

var (
	FunctionFlags = &functionFlags{
		Active:       1 << 0,
		Released:     1 << 1,
		Provisioning: 1 << 2,
	}
)

type Broker struct {
	AuthFile string
	HostAddr string
}

type Error struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"error"`
}

type Function struct {
	Id          int // Id is assigned by a publishing Broker and refers to the Smart Function release cycle
	Arches      []string
	Description string
	Flags       int
	Fqdn        string
	Handle      string
	Img         string
	Digest      string
	ImgDigest   string
	Name        string
	Oem         string
	Seq         int
	Type        string
	Version     string
}

func (f Function) asIdentity() *shared.Identity {
	return &shared.Identity{
		Id:      strconv.Itoa(f.Id),
		Flags:   f.Flags,
		Hash:    f.Digest,
		Path:    f.Img,
		Version: f.Version,
	}
}

type functionFlags struct {
	Active       int
	Released     int
	Provisioning int
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
	Description string     `json:"description"`
	Handle      string     `json:"handle"`
	Name        string     `json:"name"`
	Oem         string     `json:"oem"`
	Version     string     `json:"version"`
	Workflows   []Workflow `json:"workflows"`
}

type SolutionRemote struct {
	Solution
	Id     int64
	Digest string
	Fqdn   string
}

func (s SolutionRemote) asIdentity() *shared.Identity {
	return &shared.Identity{
		Id:      strconv.FormatInt(s.Id, 10),
		Hash:    s.Digest,
		Path:    s.Fqdn,
		Version: s.Version,
	}
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
	Name        string         `json:"name"`
	Description string         `json:"Description"`
	Handle      string         `json:"handle"`
	Links       []WorkflowLink `json:"links"`
	Nodes       []WorkflowNode `json:"nodes"`
}

type WorkflowLink struct {
	LhsNode     string `json:"lhsNode"`
	LhsNodePort string `json:"lhsNodePort"`
	RhsNode     string `json:"rhsNode"`
	RhsNodePort string `json:"rhsNodePort"`
}

func (wl WorkflowLink) Equals(wl2 WorkflowLink) bool {
	return strings.EqualFold(wl.LhsNode, wl2.LhsNode) &&
		strings.EqualFold(wl.LhsNodePort, wl2.LhsNodePort) &&
		strings.EqualFold(wl.RhsNode, wl2.RhsNode) &&
		strings.EqualFold(wl.RhsNodePort, wl2.RhsNodePort)
}

type WorkflowNode struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Handle      string                `json:"handle"`
	Sf          *WorkflowNodeFunction `json:"sf"`
}

func (wn WorkflowNode) Equals(wn2 WorkflowNode) bool {
	return strings.EqualFold(wn.Handle, wn2.Handle)
}

type WorkflowNodeFunction struct {
	Oem     string `yaml:"oem" json:"oem"`
	Handle  string `yaml:"handle" json:"handle"`
	Version string `yaml:"version" json:"version"`
	Seq     int    `yaml:"seq,omitempty" json:"seq,omitzero"`
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
