package broker

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cast"

	"genaiz.com/genaiz/lang/enumz"
	"genaiz.com/genaiz/task/shared"
)

const (
	genaizAuthUrlKey     = "GENAIZ_AUTH_URL"
	genaizAuthSessionKey = "GENAIZ_AUTH_SESSION"

	PropSpecTypeBoolean PropSpecType = "BOOL"
	PropSpecTypeDouble  PropSpecType = "DOUBLE"
	PropSpecTypeEnum    PropSpecType = "ENUM"
	PropSpecTypeInt     PropSpecType = "INT"
	PropSpecTypeString  PropSpecType = "STRING"

	WorkspaceVisibilityPrivate WorkspaceVisibility = "PRIVATE"
	WorkspaceVisibilityOrg     WorkspaceVisibility = "ORGANIZATION"
)

var (
	DataLinkFlags = &dataLinkFlags{
		Active:   1 << 0,
		Released: 1 << 1,
	}
	FunctionFlags = &functionFlags{
		Active:       1 << 0,
		Released:     1 << 1,
		Provisioning: 1 << 2,
	}
	ProxyFlags = &proxyFlags{
		Active:      1 << 0,
		ProtocolTcp: 1 << 1,
		ProtocolUdp: 1 << 2,
	}
	SolutionFlags = &solutionFlags{
		Active:   1 << 0,
		Released: 1 << 1,
	}
	WorkspaceFlags = &workspaceFlags{
		Active:    1 << 0,
		RcEnabled: 1 << 1,
	}
	PropSpecTypes = enumz.NewEnumType(PropSpecTypeBoolean, PropSpecTypeDouble,
		PropSpecTypeEnum, PropSpecTypeInt, PropSpecTypeString)
	WorkspaceVisibilities = enumz.NewEnumType(WorkspaceVisibilityPrivate, WorkspaceVisibilityOrg)

	ErrorDataPortNotFound     = errors.New("data port not found")
	ErrorPropIllegalBool      = errors.New("illegal default value for bool type")
	ErrorPropIllegalDouble    = errors.New("illegal default value for double type")
	ErrorPropIllegalInt       = errors.New("illegal default value for int type")
	ErrorPropIllegalEnum      = errors.New("illegal default value for enum type")
	ErrorWorkflowNotFound     = errors.New("workflow not found")
	ErrorWorkflowNodeNotFound = errors.New("workflow node not found")
)

type Broker struct {
	AuthFile string
	HostAddr string
	Username string
}

func (b Broker) GetClient() (Client, error) {
	var envHost = os.Getenv(genaizAuthUrlKey)
	var envSession = os.Getenv(genaizAuthSessionKey)

	if envHost != "" && envSession != "" {
		// override in effect
		return clientFactory.Seed(envHost, envSession)
	}

	if b.HostAddr == "" {
		return clientFactory.Active(b.AuthFile)
	}

	return clientFactory.Get(b.AuthFile, b.HostAddr)
}

type DataLink struct {
	Id              *int64     `json:"id,omitempty"`
	Created         *int64     `json:"nco,omitempty"`
	Modified        *int64     `json:"nms,omitempty"`
	Flags           *int       `json:"flags,omitempty"`
	Seq             *int       `json:"seq,omitempty"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Oem             string     `json:"oem"`
	Handle          string     `json:"handle"`
	Fqdn            *string    `json:"fqdn,omitempty"`
	Version         string     `json:"version"`
	PropSpecs       []PropSpec `json:"propSpecs,omitempty"`
	SecretSpecs     []PropSpec `json:"secretSpecs,omitempty"`
	OutboundProxies []Proxy    `json:"outboundProxies,omitempty"`
}

func (dl *DataLink) FindPropSpec(key string) *PropSpec {
	if i := slices.IndexFunc(dl.PropSpecs, func(spec PropSpec) bool {
		return strings.EqualFold(spec.Key, key)
	}); i >= 0 {
		return &dl.PropSpecs[i]
	}

	return nil
}

func (dl *DataLink) FindProxy(host string, port int) *Proxy {
	if i := slices.IndexFunc(dl.OutboundProxies, func(proxy Proxy) bool {
		return strings.EqualFold(host, proxy.Host) && port == proxy.Port
	}); i >= 0 {
		return &dl.OutboundProxies[i]
	}

	return nil
}

func (dl *DataLink) FindSecretSpec(key string) *PropSpec {
	if i := slices.IndexFunc(dl.SecretSpecs, func(spec PropSpec) bool {
		return strings.EqualFold(spec.Key, key)
	}); i >= 0 {
		return &dl.SecretSpecs[i]
	}

	return nil
}

func (dl *DataLink) GetBranch() string {
	return fmt.Sprintf("%s:%s", dl.GetFqdn(), dl.Version)
}

func (dl *DataLink) GetFqdn() string {
	if dl.Fqdn != nil {
		return *dl.Fqdn
	}

	return fmt.Sprintf("%s/%s", dl.Oem, dl.Handle)
}

func (dl *DataLink) GetVersion() string {
	if dl.Seq != nil && !dl.IsReleased() {
		return fmt.Sprintf("%s-rc-%d", dl.Version, *dl.Seq)
	}

	return dl.Version
}

func (dl *DataLink) IsActive() bool {
	if dl.Flags == nil {
		return false
	}

	return (*dl.Flags & DataLinkFlags.Active) == DataLinkFlags.Active
}

func (dl *DataLink) IsAfter(dataLink *DataLink) bool {
	if dataLink != nil && (dl.GetBranch() == dataLink.GetBranch()) {
		// When a sequence value is blank, we assume it's a local build, and it is not the latest for the given branch
		if dl.Seq != nil {
			if dataLink.Seq == nil {
				return !dataLink.IsReleased()
			}

			return *dl.Seq > *dataLink.Seq
		}

		return dl.IsReleased()
	}

	return false
}

func (dl *DataLink) IsEqual(oem, handle, version string) bool {
	return dl.IsRevision(oem, handle) &&
		strings.EqualFold(dl.Version, version)
}

func (dl *DataLink) IsReleased() bool {
	if dl.Flags == nil {
		return false
	}

	return (*dl.Flags & DataLinkFlags.Released) == DataLinkFlags.Released
}

func (dl *DataLink) IsRevision(oem, handle string) bool {
	return strings.EqualFold(dl.Oem, oem) &&
		strings.EqualFold(dl.Handle, handle)
}

func (dl *DataLink) RemovePropSpec(key string) *PropSpec {
	if spec := dl.FindPropSpec(key); spec != nil {
		var result = *spec

		dl.PropSpecs = dl.removePropSpec(dl.PropSpecs, key)
		return &result
	}

	return nil
}

func (dl *DataLink) RemoveProxy(host string, port int) *Proxy {
	if proxy := dl.FindProxy(host, port); proxy != nil {
		var result = *proxy

		dl.OutboundProxies = dl.removeProxy(dl.OutboundProxies, host, port)
		return &result
	}

	return nil
}

func (dl *DataLink) RemoveSecretSpec(key string) *PropSpec {
	if spec := dl.FindSecretSpec(key); spec != nil {
		var result = *spec

		dl.SecretSpecs = dl.removePropSpec(dl.SecretSpecs, key)
		return &result
	}

	return nil
}

func (dl *DataLink) ReplacePropSpec(spec *PropSpec) {
	if newSpecs, err := dl.replacePropSpec(dl.PropSpecs, spec); err == nil {
		dl.PropSpecs = newSpecs
	}
}

func (dl *DataLink) ReplaceSecretSpec(spec *PropSpec) {
	if newSpecs, err := dl.replacePropSpec(dl.SecretSpecs, spec); err == nil {
		dl.SecretSpecs = newSpecs
	}
}

func (dl *DataLink) Sanitize() *DataLink {
	var result = &DataLink{
		Handle:      dl.Handle,
		Oem:         dl.Oem,
		Version:     dl.Version,
		Seq:         dl.Seq,
		Name:        dl.Name,
		Description: dl.Description,
	}

	for _, proxy := range dl.OutboundProxies {
		result.OutboundProxies = append(result.OutboundProxies, proxy)
	}

	for _, spec := range dl.PropSpecs {
		result.PropSpecs = append(result.PropSpecs, spec.Sanitize())
	}

	for _, spec := range dl.SecretSpecs {
		result.SecretSpecs = append(result.SecretSpecs, spec.Sanitize())
	}

	return result
}

func (dl *DataLink) removePropSpec(specs []PropSpec, key string) []PropSpec {
	return slices.DeleteFunc(specs, func(s PropSpec) bool {
		return strings.EqualFold(key, s.Key)
	})
}

func (dl *DataLink) removeProxy(proxies []Proxy, host string, port int) []Proxy {
	return slices.DeleteFunc(proxies, func(p Proxy) bool {
		return strings.EqualFold(host, p.Host) && port == p.Port
	})
}

func (dl *DataLink) replacePropSpec(specs []PropSpec, spec *PropSpec) ([]PropSpec, error) {
	if spec != nil {
		var replaced = []PropSpec{*spec}
		var purged = dl.removePropSpec(specs, spec.Key)

		if !slices.EqualFunc(specs, purged, func(spec2 PropSpec, spec PropSpec) bool {
			return strings.EqualFold(spec2.Key, spec.Key)
		}) {
			// Put the replaced property at the head of the list, so we can spot it easier
			return append(replaced, purged...), nil
		}
	}

	return nil, errors.New("not found")
}

type dataLinkFlags struct {
	Active   int
	Released int
}

type DataPort struct {
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Handle      string `json:"handle"`
	Name        string `json:"name"`
}

func ListDataPorts(ports any) []DataPort {
	var result []DataPort
	var list []interface{}
	var ok bool

	if list, ok = ports.([]interface{}); ok {
		var portMap map[string]interface{}

		for _, portInterface := range list {
			if portMap, ok = portInterface.(map[string]interface{}); ok {
				result = append(result, DataPort{
					Description: cast.ToString(portMap["description"]),
					Handle:      strings.ToLower(cast.ToString(portMap["handle"])),
					Name:        cast.ToString(portMap["name"]),
				})
			}
		}
	}

	return result
}

type DeviceAuth struct {
	DeviceCode              string `json:"device_code"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
	UserCode                string `json:"user_code"`
	VerificationUri         string `json:"verification_uri"`
	VerificationUriComplete string `json:"verification_uri_complete"`
}

type DeviceClient struct {
	ClientId    string
	ClientScope string
	GrantType   string
}

func NewDeviceClient(clientId, clientScope, grantType string) *DeviceClient {
	return &DeviceClient{
		ClientId:    clientId,
		ClientScope: clientScope,
		GrantType:   grantType,
	}
}

type Error struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"error"`
}

type Function struct {
	Id              int        `json:"id,omitempty"`
	Flags           int        `json:"flags,omitempty"`
	Seq             *int       `json:"seq,omitempty"`
	Name            string     `json:"name,omitempty"`
	Description     string     `json:"description,omitempty"`
	Oem             string     `json:"oem"`
	Handle          string     `json:"handle"`
	Fqdn            string     `json:"fqdn,omitempty"`
	Img             string     `json:"img,omitempty"`
	Version         string     `json:"version"`
	Digest          string     `json:"digest,omitempty"`
	Arches          []string   `json:"-"`
	Type            string     `json:"type"`
	InputPorts      []DataPort `json:"inputPorts,omitempty"`
	OutputPorts     []DataPort `json:"outputPorts,omitempty"`
	PropSpecs       []PropSpec `json:"propSpecs,omitempty"`
	ResultValues    []string   `json:"resultValues,omitempty"`
	DataSources     []string   `json:"dataSources,omitempty"`
	DataStores      []string   `json:"dataStores,omitempty"`
	OutboundProxies []Proxy    `json:"outboundProxies,omitempty"`
	ImgDigest       string     `json:"imgDigest,omitempty"`
}

func (f Function) FindDataPortByHandle(handle string) *DataPort {
	var ports []DataPort

	ports = append(ports, f.InputPorts...)
	ports = append(ports, f.OutputPorts...)

	for _, port := range ports {
		if strings.EqualFold(port.Handle, handle) {
			return &port
		}
	}

	return nil
}

func (f Function) GetFullVersion() string {
	var result string

	if f.Seq == nil {
		result = f.Version
	} else {
		result = fmt.Sprintf("%s-rc-%d", f.Version, *f.Seq)
	}

	return result
}

func (f Function) GetDataSourceLinks() []DataLink {
	var result []DataLink

	for _, dl := range f.DataSources {
		var oem, handle, ver = ParseFqdnVersion(dl)

		result = append(result, DataLink{
			Oem:     oem,
			Handle:  handle,
			Version: ver,
		})
	}

	return result
}

func (f Function) GetDataStoreLinks() []DataLink {
	var result []DataLink

	for _, dl := range f.DataStores {
		var oem, handle, ver = ParseFqdnVersion(dl)

		result = append(result, DataLink{
			Oem:     oem,
			Handle:  handle,
			Version: ver,
		})
	}

	return result
}

func (f Function) asIdentity() *shared.Identity {
	return &shared.Identity{
		Id:      strconv.Itoa(f.Id),
		Flags:   f.Flags,
		Hash:    f.Digest,
		Path:    f.Img,
		Version: f.GetFullVersion(),
	}
}

func (f Function) toModel() *functionModel {
	var sanitizedProps []PropSpec

	for _, spec := range f.PropSpecs {
		sanitizedProps = append(sanitizedProps, spec.Sanitize())
	}

	return &functionModel{
		DataSources:     f.DataSources,
		DataStores:      f.DataStores,
		Description:     f.Description,
		Handle:          f.Handle,
		ImgDigest:       f.ImgDigest,
		InputPorts:      f.InputPorts,
		Name:            f.Name,
		Oem:             f.Oem,
		OutboundProxies: f.OutboundProxies,
		OutputPorts:     f.OutputPorts,
		PropSpecs:       sanitizedProps,
		ResultValues:    f.ResultValues,
		Type:            f.Type,
		Version:         f.Version,
	}
}

func MapFunction(fn any) *Function {
	if fnMap, ok := fn.(map[string]interface{}); ok {
		return &Function{
			Arches:          cast.ToStringSlice(fnMap["arches"]),
			DataSources:     cast.ToStringSlice(fnMap["datasources"]),
			DataStores:      cast.ToStringSlice(fnMap["datastores"]),
			Description:     cast.ToString(fnMap["description"]),
			Handle:          cast.ToString(fnMap["handle"]),
			InputPorts:      ListDataPorts(fnMap["inputports"]),
			Name:            cast.ToString(fnMap["name"]),
			Oem:             cast.ToString(fnMap["oem"]),
			OutboundProxies: ListProxies(fnMap["outboundproxies"]),
			OutputPorts:     ListDataPorts(fnMap["outputports"]),
			PropSpecs:       ListPropSpecs(fnMap["propspecs"]),
			ResultValues:    cast.ToStringSlice(fnMap["resultvalues"]),
			Type:            cast.ToString(fnMap["type"]),
			Version:         cast.ToString(fnMap["version"]),
		}
	}

	return nil
}

type functionFlags struct {
	Active       int
	Released     int
	Provisioning int
}

// functionModel provides a transport definition for provisioning smart functions
type functionModel struct {
	Description     string     `json:"description"`
	DataSources     []string   `json:"dataSources"`
	DataStores      []string   `json:"dataStores"`
	Handle          string     `json:"handle"`
	ImgDigest       string     `json:"imgDigest"`
	InputPorts      []DataPort `json:"inputPorts,omitempty"`
	Name            string     `json:"name"`
	Oem             string     `json:"oem"`
	OutboundProxies []Proxy    `json:"outboundProxies"`
	OutputPorts     []DataPort `json:"outputPorts,omitempty"`
	PropSpecs       []PropSpec `json:"propSpecs,omitempty"`
	ResultValues    []string   `json:"resultValues,omitempty"`
	Type            string     `json:"type"`
	Version         string     `json:"version"`
}

type PropSpecType = string

type PropSpec struct {
	Key         string       `yaml:"key" json:"key"`
	Name        string       `yaml:"name" json:"name"`
	Description string       `yaml:"description,omitempty" json:"description,omitempty"`
	Type        PropSpecType `yaml:"type" json:"type"`
	Value       string       `yaml:"value,omitempty" json:"value,omitempty"`
	Values      []string     `yaml:"values,omitempty" json:"values,omitempty"`
}

func (ps PropSpec) GetDefaultValue() string {
	return ps.Value
}

func (ps PropSpec) GetDescription() string {
	return ps.Description
}

func (ps PropSpec) GetKey() string {
	return ps.Key
}

func (ps PropSpec) Sanitize() PropSpec {
	return PropSpec{
		Key:         ps.Key,
		Name:        ps.Name,
		Description: ps.Description,
		Type:        strings.ToUpper(ps.Type),
		Value:       ps.Value,
		Values:      ps.Values,
	}
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

func (ps PropSpec) VarSpec() shared.VarSpec {
	return ps
}

func FindPropSpec(specs []PropSpec, key string) *PropSpec {
	if i := slices.IndexFunc(specs, func(spec PropSpec) bool {
		return strings.EqualFold(spec.Key, key)
	}); i >= 0 {
		return &specs[i]
	}

	return nil
}

// ListPropSpecs is deprecated, viper.UnmarshallKey was overlooked, see schema.Keys.Unmarshall
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

type Proxy struct {
	Host  string `json:"host" yaml:"host"`
	Port  int    `json:"port" yaml:"port"`
	Flags int    `json:"flags" yaml:"flags"`
}

func (p *Proxy) IsActive() bool {
	return (p.Flags & ProxyFlags.Active) == ProxyFlags.Active
}

func (p *Proxy) IsTcp() bool {
	return (p.Flags & ProxyFlags.ProtocolTcp) == ProxyFlags.ProtocolTcp
}

func (p *Proxy) IsUdp() bool {
	return (p.Flags & ProxyFlags.ProtocolUdp) == ProxyFlags.ProtocolUdp
}

func (p *Proxy) IsEqual(host string, port int) bool {
	return strings.EqualFold(p.Host, host) && p.Port == port
}

func (p *Proxy) SetActive(active bool) {
	if active {
		p.Flags |= ProxyFlags.Active
	} else {
		p.Flags &= ^ProxyFlags.Active
	}
}

func (p *Proxy) SetTcp(enabled bool) {
	if enabled {
		p.Flags |= ProxyFlags.ProtocolTcp
	} else {
		p.Flags &= ^ProxyFlags.ProtocolTcp
	}
}

func (p *Proxy) SetUdp(enabled bool) {
	if enabled {
		p.Flags |= ProxyFlags.ProtocolUdp
	} else {
		p.Flags &= ^ProxyFlags.ProtocolUdp
	}
}

func ListProxies(proxies any) []Proxy {
	var result []Proxy
	var list []interface{}
	var ok bool

	if list, ok = proxies.([]interface{}); ok {
		var proxyMap map[string]interface{}

		for _, proxyInterface := range list {
			if proxyMap, ok = proxyInterface.(map[string]interface{}); ok {
				result = append(result, Proxy{
					Host:  cast.ToString(proxyMap["host"]),
					Port:  cast.ToInt(proxyMap["port"]),
					Flags: cast.ToInt(proxyMap["flags"]),
				})
			}
		}
	}

	return result
}

type proxyFlags struct {
	Active      int
	ProtocolTcp int
	ProtocolUdp int
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
	// Keep the ordering for the marshaler
	Id          *int64     `json:"id,omitempty"`
	Created     *int64     `json:"nco,omitempty"`
	Modified    *int64     `json:"nms,omitempty"`
	Flags       *int       `json:"flags,omitempty"`
	Seq         *int       `json:"seq,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Oem         string     `json:"oem"`
	Handle      string     `json:"handle"`
	Fqdn        *string    `json:"fqdn,omitempty"`
	Version     string     `json:"version"`
	Digest      *string    `json:"digest,omitempty"`
	Workflows   []Workflow `json:"workflows,omitempty"`
}

func (s Solution) FindWorkflowByHandle(handle string) (*Workflow, error) {
	var nodeIndex = slices.IndexFunc(s.Workflows, func(wf Workflow) bool {
		return strings.EqualFold(wf.Handle, handle)
	})

	if nodeIndex < 0 {
		return nil, ErrorWorkflowNotFound
	}

	return &s.Workflows[nodeIndex], nil
}

func (s Solution) GetBranch() string {
	// The difference with GetVersion is omission of the sequence
	return fmt.Sprintf("%s:%s", s.GetFqdn(), s.Version)
}

func (s Solution) GetFqdn() string {
	if s.Fqdn != nil {
		return *s.Fqdn
	}

	return fmt.Sprintf("%s/%s", s.Oem, s.Handle)
}

func (s Solution) GetVersion() string {
	if s.Seq != nil && !s.IsReleased() {
		return fmt.Sprintf("%s-rc-%d", s.Version, *s.Seq)
	}

	return s.Version
}

func (s Solution) IsAfter(solution *Solution) bool {
	if solution != nil && (s.GetBranch() == solution.GetBranch()) {
		// When a sequence value is blank, we assume it's a local build, and it is not the latest for the given branch
		if s.Seq != nil {
			if solution.Seq == nil {
				return !solution.IsReleased()
			}

			return *s.Seq > *solution.Seq
		}

		return s.IsReleased()
	}

	return false
}

func (s Solution) IsActive() bool {
	return s.Flags != nil && (*s.Flags&SolutionFlags.Active) == SolutionFlags.Active
}

func (s Solution) IsReleased() bool {
	return s.Flags != nil && (*s.Flags&SolutionFlags.Released) == SolutionFlags.Released
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

type solutionFlags struct {
	Active   int
	Released int
}

type Workflow struct {
	// Keep the ordering for the marshaler
	Id          *int64         `yaml:"-" json:"id,omitempty"`
	Created     *int64         `yaml:"-" json:"nco,omitempty"`
	Modified    *int64         `yaml:"-" json:"nms,omitempty"`
	Flags       *int           `yaml:"-" json:"flags,omitempty"`
	Handle      string         `json:"handle"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Links       []WorkflowLink `json:"links,omitempty"`
	Nodes       []WorkflowNode `json:"nodes,omitempty"`
}

func (wf Workflow) ContainsNode(handle string) bool {
	return slices.ContainsFunc(wf.Nodes, func(node WorkflowNode) bool {
		return strings.EqualFold(node.Handle, handle)
	})
}

func (wf Workflow) FindNodeByHandle(handle string) (*WorkflowNode, error) {
	var nodeIndex = slices.IndexFunc(wf.Nodes, func(node WorkflowNode) bool {
		return strings.EqualFold(node.Handle, handle)
	})

	if nodeIndex < 0 {
		return nil, ErrorWorkflowNodeNotFound
	}

	return &wf.Nodes[nodeIndex], nil
}

func (wf Workflow) FindNodeBySf(fn *Function) (*WorkflowNode, error) {
	if fn != nil {
		var nodeIndex = slices.IndexFunc(wf.Nodes, func(node WorkflowNode) bool {
			if node.Sf != nil {
				return strings.EqualFold(node.Sf.Oem, fn.Oem) &&
					strings.EqualFold(node.Sf.Handle, fn.Handle) &&
					strings.EqualFold(node.Sf.Version, fn.Version)
			}

			return false
		})

		if nodeIndex >= 0 {
			return &wf.Nodes[nodeIndex], nil
		}
	}

	return nil, ErrorWorkflowNodeNotFound
}

func (wf Workflow) FindNodeHandleBySf(oem, handle, version string) (string, error) {
	var nodeIndex = slices.IndexFunc(wf.Nodes, func(node WorkflowNode) bool {
		if node.Sf != nil {
			return strings.EqualFold(node.Sf.Oem, oem) &&
				strings.EqualFold(node.Sf.Handle, handle) &&
				strings.EqualFold(node.Sf.Version, version)
		}

		return false
	})

	if nodeIndex < 0 {
		return "", ErrorWorkflowNodeNotFound
	}

	return wf.Nodes[nodeIndex].Handle, nil
}

func (wf Workflow) HasNodeProps() bool {
	if len(wf.Nodes) > 0 {
		if i := slices.IndexFunc(wf.Nodes, func(node WorkflowNode) bool {
			return len(node.Props) > 0
		}); i >= 0 {
			return true
		}
	}

	return false
}

type WorkflowLink struct {
	LhsNode     string `json:"lhsNode"`
	LhsNodeId   *int64 `yaml:"-" json:"lhsNodeId,omitempty"`
	LhsNodePort string `json:"lhsNodePort,omitempty"`
	RhsNode     string `json:"rhsNode"`
	RhsNodeId   *int64 `yaml:"-" json:"rhsNodeId,omitempty"`
	RhsNodePort string `json:"rhsNodePort"`
	WorkflowId  *int64 `yaml:"-" json:"workflowId,omitempty"`
}

func (wl WorkflowLink) Equals(wl2 WorkflowLink) bool {
	return strings.EqualFold(wl.LhsNode, wl2.LhsNode) &&
		strings.EqualFold(wl.LhsNodePort, wl2.LhsNodePort) &&
		strings.EqualFold(wl.RhsNode, wl2.RhsNode) &&
		strings.EqualFold(wl.RhsNodePort, wl2.RhsNodePort)
}

func (wl WorkflowLink) String() string {
	var leftSide, rightSide string

	if wl.LhsNode != "" {
		leftSide = wl.LhsNode
	}

	if wl.LhsNodePort != "" {
		leftSide = fmt.Sprintf("%s[%s]", leftSide, wl.LhsNodePort)
	}

	if wl.RhsNode != "" {
		rightSide = wl.RhsNode
	}

	if wl.RhsNodePort != "" {
		rightSide = fmt.Sprintf("%s[%s]", rightSide, wl.RhsNodePort)
	}

	return fmt.Sprintf("%s:%s", leftSide, rightSide)
}

type WorkflowNode struct {
	// Keep the ordering for the marshaler
	Id              *int64                `yaml:"-" json:"id"`
	Handle          string                `yaml:"handle" json:"handle"`
	Name            string                `yaml:"name" json:"name"`
	Description     string                `yaml:"description,omitempty" json:"description,omitempty"`
	Props           map[string]string     `yaml:"props,omitempty" json:"props,omitempty"`
	Sf              *WorkflowNodeFunction `yaml:"sf,omitempty" json:"sf,omitempty"`
	WorkflowId      *int64                `yaml:"-" json:"workflowId,omitempty"`
	SmartFunctionId *int64                `yaml:"-" json:"smartFunctionId,omitempty"`
}

func (wn *WorkflowNode) AssignProp(key, value string) {
	if wn.Props == nil {
		wn.Props = make(map[string]string)
	}

	wn.Props[strings.ToUpper(key)] = value
}

func (wn *WorkflowNode) Equals(wn2 WorkflowNode) bool {
	return strings.EqualFold(wn.Handle, wn2.Handle)
}

func (wn *WorkflowNode) HasProp(key string) bool {
	if _, ok := wn.Props[strings.ToLower(key)]; ok {
		return true
	}

	if _, ok := wn.Props[strings.ToUpper(key)]; ok {
		return true
	}

	return false
}

func (wn *WorkflowNode) NormalizeProps() {
	if wn.Props != nil {
		for k, v := range wn.Props {
			delete(wn.Props, strings.ToLower(k))
			wn.Props[strings.ToUpper(k)] = v
		}
	}
}

func (wn *WorkflowNode) RemoveProp(key string) {
	if wn.Props != nil {
		delete(wn.Props, strings.ToUpper(key))
		delete(wn.Props, strings.ToLower(key))
	}
}

func (wn *WorkflowNode) ValidateProps(specs []shared.VarSpec) error {
	if wn.Props != nil {
		for key, value := range wn.Props {
			if i := slices.IndexFunc(specs, func(p shared.VarSpec) bool {
				return strings.EqualFold(p.GetKey(), key)
			}); i >= 0 {
				if err := specs[i].Validate(value); err == nil {
					return nil
				}

				return fmt.Errorf("value [%s] is not valid for key [%s]", value, key)
			}

			return fmt.Errorf("the key [%s] is invalid for node [%s]", key, wn.Handle)
		}
	}

	return nil
}

type WorkflowNodeFunction struct {
	Oem         string     `yaml:"oem" json:"oem"`
	Handle      string     `yaml:"handle" json:"handle"`
	Version     string     `yaml:"version" json:"version"`
	Seq         int        `yaml:"seq,omitempty" json:"seq,omitzero"`
	DataSources []DataLink `yaml:"-" json:"dataSources,omitempty"`
	DataStores  []DataLink `yaml:"-" json:"dataStores,omitempty"`
	InputPorts  []DataPort `yaml:"-" json:"InputPorts,omitempty"`
	OutputPorts []DataPort `yaml:"-" json:"OutputPorts,omitempty"`
}

func (wnf WorkflowNodeFunction) IsEqual(fn *Function) bool {
	if fn == nil {
		return false
	}

	return strings.EqualFold(wnf.Oem, fn.Oem) &&
		strings.EqualFold(wnf.Handle, fn.Handle) &&
		strings.EqualFold(wnf.Version, fn.Version)
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

type WorkspaceVisibility = string

type Workspace struct {
	Id          int64  `yaml:"id,omitempty" json:"id,omitempty"`
	Created     int64  `yaml:"nco,omitempty" json:"nco,omitempty"`
	Modified    int64  `yaml:"nms,omitempty" json:"nms,omitempty"`
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	OwnerAppId  int    `yaml:"ownerAppId,omitempty" json:"ownerAppId,omitempty"`
	OwnerUserId int    `yaml:"OwnerUserId,omitempty" json:"OwnerUserId,omitempty"`
	Visibility  string `yaml:"visibility" json:"visibility"`
	RcEnabled   bool   `yaml:"-" json:"-"`
	Flags       *int   `yaml:"Flags,omitempty"`
}

func (w Workspace) IsActive() bool {
	return w.Flags != nil && (*w.Flags&WorkspaceFlags.Active) == WorkspaceFlags.Active
}

func (w Workspace) IsRcEnabled() bool {
	if w.Flags != nil {
		return (*w.Flags & WorkspaceFlags.RcEnabled) == WorkspaceFlags.RcEnabled
	}

	return w.RcEnabled
}

func (w Workspace) MarshalJSON() ([]byte, error) {
	var bytes []byte
	var flags int

	if w.Flags == nil {
		if w.RcEnabled {
			flags = WorkspaceFlags.Active | WorkspaceFlags.RcEnabled
		} else {
			flags = WorkspaceFlags.Active
		}
	} else {
		flags = *w.Flags
	}

	bytes, _ = json.Marshal(struct {
		Id          string `json:"id"`
		Created     int64  `json:"created"`
		Modified    int64  `json:"modified"`
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Visibility  string `json:"visibility"`
		Flags       int    `json:"flags"`
	}{
		Id:          cast.ToString(w.Id),
		Created:     w.Created,
		Modified:    w.Modified,
		Name:        w.Name,
		Description: w.Description,
		Visibility:  strings.ToUpper(w.Visibility),
		Flags:       flags,
	})
	return bytes, nil
}

type workspaceFlags struct {
	Active    int
	RcEnabled int
}

type WorkspaceFlow struct {
	Id          int64
	Created     int64  `json:"nco,omitempty"`
	Modified    int64  `json:"nms,omitempty"`
	WorkspaceId int64  `json:"workspaceId"`
	SolutionId  int64  `json:"solutionId"`
	WorkflowId  int64  `json:"workflowId"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Flags       *int   `json:"Flags,omitempty"`
}

func ParseFqdnVersion(value string) (string, string, string) {
	var oem, handle, ver string
	var oemHandleParts = strings.Split(value, "/")

	if len(oemHandleParts) > 1 {
		var handleVersionParts = strings.Split(oemHandleParts[1], ":")

		if len(handleVersionParts) > 1 {
			ver = handleVersionParts[1]
		}

		handle = handleVersionParts[0]
		oem = oemHandleParts[0]
	} else {
		oem = value
	}

	return oem, handle, ver
}
