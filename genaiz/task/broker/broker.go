package broker

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cast"

	"genaiz.com/genaiz/lang/enumz"
	"genaiz.com/genaiz/task/shared"
)

const (
	PropSpecTypeBoolean PropSpecType = "BOOL"
	PropSpecTypeDouble  PropSpecType = "DOUBLE"
	PropSpecTypeEnum    PropSpecType = "ENUM"
	PropSpecTypeInt     PropSpecType = "INT"
	PropSpecTypeString  PropSpecType = "STRING"
)

var (
	FunctionFlags = &functionFlags{
		Active:       1 << 0,
		Released:     1 << 1,
		Provisioning: 1 << 2,
	}
	PropSpecTypes = enumz.NewEnumType(PropSpecTypeBoolean, PropSpecTypeDouble,
		PropSpecTypeEnum, PropSpecTypeInt, PropSpecTypeString)

	ErrorPropIllegalBool   = errors.New("illegal bool value")
	ErrorPropIllegalDouble = errors.New("illegal double value")
	ErrorPropIllegalInt    = errors.New("illegal int value")
	ErrorPropIllegalEnum   = errors.New("illegal enum value")
)

type PropSpecType = string

type Broker struct {
	AuthFile string
	HostAddr string
}

func (b Broker) GetClient() (Client, error) {
	if b.HostAddr == "" {
		return clientFactory.Active(b.AuthFile)
	}

	return clientFactory.Get(b.AuthFile, b.HostAddr)
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
	PropSpecs   []PropSpec
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

func (f Function) toModel() *functionModel {
	return &functionModel{
		Description: f.Description,
		Handle:      f.Handle,
		ImgDigest:   f.ImgDigest,
		Name:        f.Name,
		Oem:         f.Oem,
		PropSpecs:   f.PropSpecs,
		Type:        f.Type,
		Version:     f.Version,
	}
}

type functionFlags struct {
	Active       int
	Released     int
	Provisioning int
}

// functionModel provides a transport definition for provisioning smart functions
type functionModel struct {
	Description string     `json:"description"`
	Handle      string     `json:"handle"`
	ImgDigest   string     `json:"imgDigest"`
	Name        string     `json:"name"`
	Oem         string     `json:"oem"`
	PropSpecs   []PropSpec `json:"propSpecs"`
	Type        string     `json:"type"`
	Version     string     `json:"version"`
}

type PropSpec struct {
	Key         string       `yaml:"key" json:"key"`
	Name        string       `yaml:"name" json:"name"`
	Description string       `yaml:"description" json:"description"`
	Type        PropSpecType `yaml:"type" json:"type"`
	Value       string       `yaml:"value" json:"value"`
	Values      []string     `yaml:"values" json:"values"`
}

func (ps PropSpec) Validate(value any) error {
	var err error

	if strings.EqualFold(ps.Type, PropSpecTypeBoolean) {
		if !slices.Contains([]string{"true", "false"}, strings.ToLower(cast.ToString(value))) {
			err = ErrorPropIllegalBool
		}
	} else if strings.EqualFold(ps.Type, PropSpecTypeDouble) {
		if _, err = cast.ToFloat32E(value); err != nil {
			err = ErrorPropIllegalDouble
		}
	} else if strings.EqualFold(ps.Type, PropSpecTypeInt) {
		if _, err = strconv.Atoi(cast.ToString(value)); err != nil {
			err = ErrorPropIllegalInt
		}
	} else if strings.EqualFold(ps.Type, PropSpecTypeEnum) {
		if !slices.Contains(ps.Values, cast.ToString(value)) {
			err = ErrorPropIllegalEnum
		}
	}

	return err
}

func FindPropSpec(specs any, key string) *PropSpec {
	var result *PropSpec
	var list []interface{}
	var ok bool

	if list, ok = specs.([]interface{}); ok {
		var specMap map[string]interface{}

		for _, specInterface := range list {
			if specMap, ok = specInterface.(map[string]interface{}); ok {
				if key == specMap["key"] {
					result = &PropSpec{
						Key:         cast.ToString(specMap["key"]),
						Description: cast.ToString(specMap["description"]),
						Name:        cast.ToString(specMap["name"]),
						Type:        cast.ToString(specMap["type"]),
						Value:       cast.ToString(specMap["value"]),
						Values:      cast.ToStringSlice(specMap["values"]),
					}
					break
				}
			}
		}
	}

	return result
}

func ListPropSpecs(specs any) []PropSpec {
	var result []PropSpec
	var list []interface{}
	var ok bool

	if list, ok = specs.([]interface{}); ok {
		var specMap map[string]interface{}

		for _, specInterface := range list {
			if specMap, ok = specInterface.(map[string]interface{}); ok {
				result = append(result, PropSpec{
					Key:         cast.ToString(specMap["key"]),
					Description: cast.ToString(specMap["description"]),
					Name:        cast.ToString(specMap["name"]),
					Type:        cast.ToString(specMap["type"]),
					Value:       cast.ToString(specMap["value"]),
					Values:      cast.ToStringSlice(specMap["values"]),
				})
			}
		}
	}

	return result
}

type provisionData struct {
	Auth string
	Sf   Function
}

type publishingData struct {
	Sf Function
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

type SolutionRemote struct {
	Solution
	Id     int64
	Digest string
	Fqdn   string
}

func (r SolutionRemote) asIdentity() *shared.Identity {
	return &shared.Identity{
		Id:      strconv.FormatInt(r.Id, 10),
		Hash:    r.Digest,
		Path:    r.Fqdn,
		Version: r.Version,
	}
}

type Workflow struct {
	Name        string         `json:"name"`
	Description string         `json:"Description"`
	Handle      string         `json:"handle"`
	Links       []WorkflowLink `json:"links"`
	Nodes       []WorkflowNode `json:"nodes"`
}

func (wf Workflow) ContainsNode(handle string) bool {
	return slices.ContainsFunc(wf.Nodes, func(node WorkflowNode) bool {
		return strings.EqualFold(node.Handle, handle)
	})
}

func (wf Workflow) FindNodeHandleBySf(oem, handle, version string) (string, error) {
	var nodeIndex = slices.IndexFunc(wf.Nodes, func(node WorkflowNode) bool {
		var sf = node.Sf

		if sf != nil {
			return strings.EqualFold(sf.Oem, oem) &&
				strings.EqualFold(sf.Handle, handle) &&
				strings.EqualFold(sf.Version, version)
		}

		return false
	})

	if nodeIndex < 0 {
		return "", errors.New("node not found")
	}

	return wf.Nodes[nodeIndex].Handle, nil
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
	Name        string                `yaml:"name" json:"name"`
	Description string                `yaml:"description,omitempty" json:"description,omitempty"`
	Handle      string                `yaml:"handle" json:"handle"`
	Sf          *WorkflowNodeFunction `yaml:"sf,omitempty" json:"sf,omitempty"`
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
