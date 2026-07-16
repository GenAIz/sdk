package broker

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/task/shared"
)

func TestDataLink_FindPropSpec(t *testing.T) {
	var testKey = "testKey"
	var testLink = &DataLink{}

	assert.Nil(t, testLink.FindPropSpec(testKey))
	testLink.PropSpecs = []PropSpec{{Key: testKey}}
	assert.Equal(t, &testLink.PropSpecs[0], testLink.FindPropSpec(testKey))
}

func TestDataLink_FindProxy(t *testing.T) {
	var expectedHost = "host"
	var expectedPort = 1337
	var testLink = &DataLink{}

	assert.Nil(t, testLink.FindProxy(expectedHost, expectedPort))
	testLink.OutboundProxies = []Proxy{{Host: expectedHost, Port: expectedPort}}
	assert.Equal(t, &testLink.OutboundProxies[0], testLink.FindProxy(expectedHost, expectedPort))
}

func TestDataLink_FindSecretSpec(t *testing.T) {
	var testKey = "testKey"
	var testLink = &DataLink{}

	assert.Nil(t, testLink.FindSecretSpec(testKey))
	testLink.SecretSpecs = []PropSpec{{Key: testKey}}
	assert.Equal(t, &testLink.SecretSpecs[0], testLink.FindSecretSpec(testKey))
}

func TestDataLink_GetBranch(t *testing.T) {
	var testDatalink = &DataLink{
		Oem:     "oem",
		Handle:  "handle",
		Version: "version",
	}
	var expectedBranch = fmt.Sprintf("%s/%s:%s", testDatalink.Oem, testDatalink.Handle, testDatalink.Version)

	assert.Equal(t, expectedBranch, testDatalink.GetBranch())
}

func TestDataLink_GetFqdn(t *testing.T) {
	var testDatalink = &DataLink{
		Oem:    "oem",
		Handle: "handle",
	}
	var expectedFqdn = fmt.Sprintf("%s/%s", testDatalink.Oem, testDatalink.Handle)

	assert.Equal(t, expectedFqdn, testDatalink.GetFqdn())
	testDatalink.Fqdn = new("expectedFqdn")
	assert.Equal(t, *testDatalink.Fqdn, testDatalink.GetFqdn())
}

func TestDataLink_GetVersion(t *testing.T) {
	var testDatalink = &DataLink{
		Version: "version",
	}

	assert.Equal(t, testDatalink.Version, testDatalink.GetVersion())
	testDatalink.Seq = new(1)
	assert.Equal(t, fmt.Sprintf("%s-rc-%d", testDatalink.Version, *testDatalink.Seq), testDatalink.GetVersion())
	testDatalink.Flags = new(DataLinkFlags.Active | DataLinkFlags.Released)
	assert.Equal(t, testDatalink.Version, testDatalink.GetVersion())
}

func TestDataLink_IsActive(t *testing.T) {
	var testLink = &DataLink{}

	assert.False(t, testLink.IsActive())
	testLink.Flags = new(DataLinkFlags.Active)
	assert.True(t, testLink.IsActive())
}

func TestDataLink_IsAfter(t *testing.T) {
	var branchOem = "oem"
	var branchHandle = "handle"
	var testDatalink1 = &DataLink{
		Oem:    branchOem,
		Handle: branchHandle,
		Seq:    new(37),
	}
	var testDatalink2 = &DataLink{
		Oem:    branchOem,
		Handle: branchHandle,
		Seq:    new(1337),
	}

	assert.False(t, testDatalink1.IsAfter(testDatalink2))
	assert.True(t, testDatalink2.IsAfter(testDatalink1))
}

func TestDataLink_IsAfter_DifferentBranch(t *testing.T) {
	var branchOem = "oem"
	var branchHandle = "handle"
	var testDatalink1 = &DataLink{
		Oem:    branchOem,
		Handle: branchHandle,
	}
	var testDatalink2 = &DataLink{
		Oem:    branchOem,
		Handle: "different",
	}

	assert.False(t, testDatalink1.IsAfter(testDatalink2))
}

func TestDataLink_IsAfter_NoSequence(t *testing.T) {
	var branchOem = "oem"
	var branchHandle = "handle"
	var testDatalink1 = &DataLink{
		Oem:    branchOem,
		Handle: branchHandle,
	}
	var testDatalink2 = &DataLink{
		Oem:    branchOem,
		Handle: branchHandle,
	}

	assert.False(t, testDatalink1.IsAfter(testDatalink2))
}

func TestDataLink_IsAfter_Partial(t *testing.T) {
	var branchOem = "oem"
	var branchHandle = "handle"
	var testDatalink1 = &DataLink{
		Oem:    branchOem,
		Handle: branchHandle,
		Seq:    new(37),
	}
	var testDatalink2 = &DataLink{
		Oem:    branchOem,
		Handle: branchHandle,
	}

	assert.True(t, testDatalink1.IsAfter(testDatalink2))
}

func TestDataLink_RemovePropSpec(t *testing.T) {
	var testProp = &PropSpec{Key: "testKey"}
	var testLink = &DataLink{PropSpecs: []PropSpec{*testProp}}

	assert.Nil(t, testLink.RemovePropSpec("wrongKey"))
	assert.Equal(t, testProp, testLink.RemovePropSpec(testProp.Key))
}

func TestDataLink_RemoveProxy(t *testing.T) {
	var testProxy = &Proxy{Host: "testHost", Port: 1337}
	var testLink = &DataLink{OutboundProxies: []Proxy{*testProxy}}

	assert.Nil(t, testLink.RemoveProxy("wrongHost", 0))
	assert.Equal(t, testProxy, testLink.RemoveProxy(testProxy.Host, testProxy.Port))
}

func TestDataLink_RemoveSecretSpec(t *testing.T) {
	var testProp = &PropSpec{Key: "testKey"}
	var testLink = &DataLink{SecretSpecs: []PropSpec{*testProp}}

	assert.Nil(t, testLink.RemoveSecretSpec("wrongKey"))
	assert.Equal(t, testProp, testLink.RemoveSecretSpec(testProp.Key))
}

func TestDataLink_ReplacePropSpec(t *testing.T) {
	var expectedStaying = "staying"
	var replacement = &PropSpec{Key: "replaced", Name: "replacing"}
	var testLink = &DataLink{PropSpecs: []PropSpec{
		{
			Key: "staying",
		},
	}}

	testLink.ReplacePropSpec(replacement)
	assert.Equal(t, 1, len(testLink.PropSpecs))
	assert.Equal(t, expectedStaying, testLink.PropSpecs[0].Key)
	testLink.PropSpecs = append(testLink.PropSpecs, PropSpec{Key: "replaced", Name: "replaced"})
	testLink.ReplacePropSpec(replacement)
	assert.Equal(t, 2, len(testLink.PropSpecs))
	assert.Equal(t, replacement.Name, testLink.PropSpecs[0].Name)
	assert.Equal(t, expectedStaying, testLink.PropSpecs[1].Key)
}

func TestDataLink_ReplaceSecretSpec(t *testing.T) {
	var expectedStaying = "staying"
	var replacement = &PropSpec{Key: "replaced", Name: "replacing"}
	var testLink = &DataLink{SecretSpecs: []PropSpec{
		{
			Key: "staying",
		},
	}}

	testLink.ReplaceSecretSpec(replacement)
	assert.Equal(t, 1, len(testLink.SecretSpecs))
	assert.Equal(t, expectedStaying, testLink.SecretSpecs[0].Key)
	testLink.SecretSpecs = append(testLink.SecretSpecs, PropSpec{Key: "replaced", Name: "replaced"})
	testLink.ReplaceSecretSpec(replacement)
	assert.Equal(t, 2, len(testLink.SecretSpecs))
	assert.Equal(t, replacement.Name, testLink.SecretSpecs[0].Name)
	assert.Equal(t, expectedStaying, testLink.SecretSpecs[1].Key)
}

func TestDataLink_Sanitize(t *testing.T) {
	var testDataLink = &DataLink{
		OutboundProxies: []Proxy{
			{
				Host:  "host",
				Port:  0,
				Flags: ProxyFlags.Active,
			},
		},
		PropSpecs: []PropSpec{
			{
				Key:         "key",
				Type:        "int",
				Name:        "name",
				Description: "description",
				Value:       "value",
				Values:      []string{"VALUE1", "VALUE2"},
			},
		},
		SecretSpecs: []PropSpec{
			{
				Key:         "key",
				Type:        "string",
				Name:        "name",
				Description: "description",
				Value:       "value",
				Values:      []string{"VALUE1", "VALUE2"},
			},
		},
	}

	actual := testDataLink.Sanitize()
	assert.Equal(t, testDataLink.OutboundProxies, actual.OutboundProxies)

	assert.Equal(t, testDataLink.PropSpecs[0].Key, actual.PropSpecs[0].Key)
	assert.Equal(t, "INT", actual.PropSpecs[0].Type)
	assert.Equal(t, testDataLink.PropSpecs[0].Name, actual.PropSpecs[0].Name)
	assert.Equal(t, testDataLink.PropSpecs[0].Description, actual.PropSpecs[0].Description)
	assert.Equal(t, testDataLink.PropSpecs[0].Value, actual.PropSpecs[0].Value)
	assert.Equal(t, testDataLink.PropSpecs[0].Values, actual.PropSpecs[0].Values)

	assert.Equal(t, testDataLink.SecretSpecs[0].Key, actual.SecretSpecs[0].Key)
	assert.Equal(t, "STRING", actual.SecretSpecs[0].Type)
	assert.Equal(t, testDataLink.SecretSpecs[0].Name, actual.SecretSpecs[0].Name)
	assert.Equal(t, testDataLink.SecretSpecs[0].Description, actual.SecretSpecs[0].Description)
	assert.Equal(t, testDataLink.SecretSpecs[0].Value, actual.SecretSpecs[0].Value)
	assert.Equal(t, testDataLink.SecretSpecs[0].Values, actual.SecretSpecs[0].Values)
}

func TestFunction_FindDataPortByHandle(t *testing.T) {
	var testHandle = "handle"
	var testFunction = &Function{}
	var expectedDataPort = &DataPort{
		Handle:      testHandle,
		Description: "testDescription",
		Name:        "Test Description",
	}

	assert.Nil(t, testFunction.FindDataPortByHandle(testHandle))
	testFunction.OutputPorts = []DataPort{
		{
			Handle: "notTheHandle",
		},
		*expectedDataPort,
	}
	actual := testFunction.FindDataPortByHandle(testHandle)
	assert.Equal(t, expectedDataPort, actual)
}

func TestFunction_GetDataSourceLinks(t *testing.T) {
	var expectedOem = "oem"
	var expectedHandle = "handle"
	var expectedVer = "version"
	var testFunction = &Function{
		DataSources: []string{
			fmt.Sprintf("%s/%s:%s", expectedOem, expectedHandle, expectedVer),
		},
	}

	actual := testFunction.GetDataSourceLinks()
	assert.Equal(t, expectedOem, actual[0].Oem)
	assert.Equal(t, expectedHandle, actual[0].Handle)
	assert.Equal(t, expectedVer, actual[0].Version)
}

func TestFunction_GetDataStoreLinks(t *testing.T) {
	var expectedOem = "oem"
	var expectedHandle = "handle"
	var expectedVer = "version"
	var testFunction = &Function{
		DataStores: []string{
			fmt.Sprintf("%s/%s:%s", expectedOem, expectedHandle, expectedVer),
		},
	}

	actual := testFunction.GetDataStoreLinks()
	assert.Equal(t, expectedOem, actual[0].Oem)
	assert.Equal(t, expectedHandle, actual[0].Handle)
	assert.Equal(t, expectedVer, actual[0].Version)
}

func TestFunction_GetFullVersion(t *testing.T) {
	var testFunction = &Function{
		Version: "version",
	}

	assert.Equal(t, testFunction.Version, testFunction.GetFullVersion())
	testFunction.Seq = new(37)
	assert.Equal(t, fmt.Sprintf("%s-rc-%d", testFunction.Version, *testFunction.Seq), testFunction.GetFullVersion())
}

func TestFunction_asIdentity(t *testing.T) {
	var actual *shared.Identity
	var function = Function{
		Id:      37,
		Digest:  "digest",
		Img:     "path",
		Version: "version",
	}

	actual = function.asIdentity()
	assert.Equal(t, function.Id, cast.ToInt(actual.Id))
	assert.Equal(t, function.Digest, actual.Hash)
	assert.Equal(t, function.Img, actual.Path)
	assert.Equal(t, function.Version, actual.Version)
}

func TestFunction_toModel(t *testing.T) {
	var testPropSpec = &PropSpec{
		Key:  "key",
		Type: "string",
	}
	var testFunction = &Function{
		Id:          37,
		Handle:      "handle",
		Oem:         "oem",
		Version:     "version",
		Name:        "name",
		Description: "description",
		PropSpecs:   []PropSpec{*testPropSpec},
	}

	actual := testFunction.toModel()
	assert.Equal(t, testFunction.Handle, actual.Handle)
	assert.Equal(t, testFunction.Oem, actual.Oem)
	assert.Equal(t, testFunction.Version, actual.Version)
	assert.Equal(t, testFunction.Name, actual.Name)
	assert.Equal(t, testFunction.Description, actual.Description)
	assert.Equal(t, testPropSpec.Key, actual.PropSpecs[0].Key)
	assert.Equal(t, "STRING", actual.PropSpecs[0].Type)
}

func TestMapFunction(t *testing.T) {
	var testOutboundProxy = &Proxy{
		Host: "expectedHost",
		Port: 37,
	}
	var testPropSpec = &PropSpec{
		Type: PropSpecTypeInt,
		Key:  "test_key",
		Name: "Test PropSpec",
	}
	var testOutPort = &DataPort{Handle: "test-out"}
	var testInPort = &DataPort{Handle: "test-in"}
	var testFunction = &Function{
		DataSources:     []string{"source1"},
		DataStores:      []string{"store1"},
		Handle:          "testHandle",
		Name:            "testName",
		Oem:             "testOem",
		Type:            "testType",
		OutboundProxies: []Proxy{*testOutboundProxy},
		OutputPorts:     []DataPort{*testOutPort},
		InputPorts:      []DataPort{*testInPort},
		PropSpecs:       []PropSpec{*testPropSpec},
		Description:     "testDescription",
		Version:         "testVersion",
		Arches:          []string{"arch"},
	}

	assert.Empty(t, MapFunction("something"))
	assert.Equal(t, testFunction, MapFunction(map[string]any{
		"datasources": testFunction.DataSources,
		"datastores":  testFunction.DataStores,
		"handle":      testFunction.Handle,
		"name":        testFunction.Name,
		"oem":         testFunction.Oem,
		"type":        testFunction.Type,
		"outboundproxies": []interface{}{
			map[string]any{
				"host": testOutboundProxy.Host,
				"port": testOutboundProxy.Port,
			},
		},
		"outputports": []interface{}{
			map[string]any{
				"handle": testOutPort.Handle,
			},
		},
		"inputports": []interface{}{
			map[string]any{
				"handle": testInPort.Handle,
			},
		},
		"propspecs": []interface{}{
			map[string]any{
				"key":  testPropSpec.Key,
				"type": testPropSpec.Type,
				"name": testPropSpec.Name,
			},
		},
		"description": testFunction.Description,
		"version":     testFunction.Version,
		"arches":      testFunction.Arches,
	}))
}

func TestPropSpec_Validate(t *testing.T) {
	var booleanSpec = &PropSpec{Type: PropSpecTypeBoolean}
	var intSpec = &PropSpec{Type: PropSpecTypeInt}
	var doubleSpec = &PropSpec{Type: PropSpecTypeDouble}
	var enumSpec = &PropSpec{Type: PropSpecTypeEnum, Values: []string{"value"}}

	assert.ErrorIs(t, booleanSpec.Validate("1"), ErrorPropIllegalBool)
	assert.ErrorIs(t, booleanSpec.Validate("0"), ErrorPropIllegalBool)
	assert.NoError(t, booleanSpec.Validate("true"))
	assert.NoError(t, booleanSpec.Validate("fALse"))
	assert.ErrorIs(t, intSpec.Validate("notAnInt"), ErrorPropIllegalInt)
	assert.ErrorIs(t, intSpec.Validate("28.98"), ErrorPropIllegalInt)
	assert.NoError(t, intSpec.Validate("37"))
	assert.ErrorIs(t, doubleSpec.Validate("notADouble"), ErrorPropIllegalDouble)
	assert.NoError(t, doubleSpec.Validate("42.2"))
	assert.ErrorIs(t, enumSpec.Validate("notListed"), ErrorPropIllegalEnum)
	assert.NoError(t, enumSpec.Validate("value"))
}

func TestPropSpec_VarSpec(t *testing.T) {
	var expectedSpec = &PropSpec{
		Key:         "expectedKey",
		Value:       "expectedValue",
		Description: "expectedDesc",
		Type:        PropSpecTypeInt,
	}
	var actualSpec = expectedSpec.VarSpec()

	assert.Equal(t, expectedSpec.Key, actualSpec.GetKey())
	assert.Equal(t, expectedSpec.Value, actualSpec.GetDefaultValue())
	assert.Equal(t, expectedSpec.Description, actualSpec.GetDescription())
	assert.NoError(t, actualSpec.Validate(37))
	assert.Error(t, actualSpec.Validate("test"))
}

func TestFindPropSpec(t *testing.T) {
	var testOtherKey = "anotherKey"
	var expectedPropSpec = &PropSpec{Key: "expectedKey"}
	var testPropSpecs = []PropSpec{
		{
			Key: testOtherKey,
		},
		*expectedPropSpec,
	}

	assert.Equal(t, expectedPropSpec, FindPropSpec(testPropSpecs, expectedPropSpec.Key))
	assert.NotNil(t, expectedPropSpec, FindPropSpec(testPropSpecs, testOtherKey))
}

func TestFindPropSpec_Empty(t *testing.T) {
	assert.Nil(t, FindPropSpec([]PropSpec{}, "KEY"))
}

func TestListDataPorts(t *testing.T) {
	var expectedHandle = "expectedHandle"
	var expectedPortMap = map[string]any{
		"handle":      expectedHandle,
		"name":        "expectedName",
		"description": "expectedDescription",
	}
	var actualPorts []DataPort

	assert.Empty(t, ListDataPorts("notAList"))
	actualPorts = ListDataPorts([]interface{}{expectedPortMap, "notASpec"})
	assert.Equal(t, 1, len(actualPorts))
	assert.True(t, strings.EqualFold(cast.ToString(expectedPortMap["handle"]), actualPorts[0].Handle))
	assert.Equal(t, expectedPortMap["name"], actualPorts[0].Name)
	assert.Equal(t, expectedPortMap["description"], actualPorts[0].Description)
}

func TestListPropSpecs(t *testing.T) {
	var expectedKey = "expectedKey"
	var expectedSpecMap = map[string]any{
		"key":         expectedKey,
		"name":        "expectedName",
		"description": "expectedDescription",
		"value":       "expectedValue",
		"values":      []string{"expected1", "expected2"},
		"type":        PropSpecTypeInt,
	}
	var actualSpecs []PropSpec

	assert.Empty(t, ListPropSpecs("notAList"))
	actualSpecs = ListPropSpecs([]interface{}{expectedSpecMap, "notASpec"})
	assert.Equal(t, 1, len(actualSpecs))
	assert.Equal(t, expectedSpecMap["description"], actualSpecs[0].Description)
	assert.Equal(t, expectedSpecMap["key"], actualSpecs[0].Key)
	assert.Equal(t, expectedSpecMap["name"], actualSpecs[0].Name)
	assert.Equal(t, expectedSpecMap["value"], actualSpecs[0].Value)
	assert.Equal(t, expectedSpecMap["values"], actualSpecs[0].Values)
	assert.Equal(t, expectedSpecMap["type"], actualSpecs[0].Type)
}

func TestProxy_IsEqual(t *testing.T) {
	var expectedHost = "host"
	var expectedPort = 37
	var testProxy = &Proxy{}

	assert.False(t, testProxy.IsEqual(expectedHost, expectedPort))
	testProxy.Host = expectedHost
	assert.False(t, testProxy.IsEqual(expectedHost, expectedPort))
	testProxy.Port = expectedPort
	assert.True(t, testProxy.IsEqual(expectedHost, expectedPort))
	testProxy.Host = ""
	assert.False(t, testProxy.IsEqual(expectedHost, expectedPort))
}

func TestProxy_SetActive(t *testing.T) {
	var testProxy = &Proxy{}

	assert.False(t, testProxy.IsActive())
	testProxy.SetActive(true)
	assert.True(t, testProxy.IsActive())
	testProxy.SetActive(false)
	assert.False(t, testProxy.IsTcp())
}

func TestProxy_SetTcp(t *testing.T) {
	var testProxy = &Proxy{}

	assert.False(t, testProxy.IsTcp())
	testProxy.SetTcp(true)
	assert.True(t, testProxy.IsTcp())
	testProxy.SetTcp(false)
	assert.False(t, testProxy.IsTcp())
}

func TestProxy_SetUdp(t *testing.T) {
	var testProxy = &Proxy{}

	assert.False(t, testProxy.IsUdp())
	testProxy.SetUdp(true)
	assert.True(t, testProxy.IsUdp())
	testProxy.SetUdp(false)
	assert.False(t, testProxy.IsUdp())
}

func TestListProxies(t *testing.T) {
	var expectedHost = "expectedHost"
	var expectedPort = 37
	var expectedProxyMap = map[string]any{
		"host": expectedHost,
		"port": expectedPort,
	}
	var actualProxies []Proxy

	assert.Empty(t, ListProxies("notAList"))
	actualProxies = ListProxies([]interface{}{expectedProxyMap, "notASpec"})
	assert.Equal(t, 1, len(actualProxies))
	assert.Equal(t, expectedProxyMap["host"], actualProxies[0].Host)
	assert.Equal(t, expectedProxyMap["port"], actualProxies[0].Port)
}

func TestSolution_FindWorkflowByHandle(t *testing.T) {
	var expectedHandle = "handle"
	var expectedName = "name"
	var testSolution = &Solution{
		Workflows: []Workflow{
			{
				Handle: expectedHandle,
				Name:   expectedName,
			},
		},
	}
	var actual, err = testSolution.FindWorkflowByHandle(expectedHandle)

	assert.NoError(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, expectedName, actual.Name)
	assert.Equal(t, expectedHandle, actual.Handle)
}

func TestSolution_FindWorkflowByHandle_NotFound(t *testing.T) {
	var testSolution = &Solution{}
	var actual, err = testSolution.FindWorkflowByHandle("notFound")

	assert.Nil(t, actual)
	assert.ErrorIs(t, err, ErrorWorkflowNotFound)
}

func TestSolution_GetBranch(t *testing.T) {
	var testSolution = &Solution{
		Oem:     "oem",
		Handle:  "handle",
		Version: "version",
	}
	var expectedBranch = fmt.Sprintf("%s/%s:%s", testSolution.Oem, testSolution.Handle, testSolution.Version)

	assert.Equal(t, expectedBranch, testSolution.GetBranch())
}

func TestSolution_GetFqdn(t *testing.T) {
	var testSolution = &Solution{
		Oem:    "oem",
		Handle: "handle",
	}
	var expectedFqdn = fmt.Sprintf("%s/%s", testSolution.Oem, testSolution.Handle)

	assert.Equal(t, expectedFqdn, testSolution.GetFqdn())
	testSolution.Fqdn = new("expectedFqdn")
	assert.Equal(t, *testSolution.Fqdn, testSolution.GetFqdn())
}

func TestSolution_GetVersion(t *testing.T) {
	var testSolution = &Solution{
		Version: "version",
	}

	assert.Equal(t, testSolution.Version, testSolution.GetVersion())
	testSolution.Seq = new(1)
	assert.Equal(t, fmt.Sprintf("%s-rc-%d", testSolution.Version, *testSolution.Seq), testSolution.GetVersion())
	testSolution.Flags = new(SolutionFlags.Active | SolutionFlags.Released)
	assert.Equal(t, testSolution.Version, testSolution.GetVersion())
}

func TestSolution_IsActive(t *testing.T) {
	var testSolution = &Solution{}

	assert.False(t, testSolution.IsActive())
	testSolution.Flags = new(SolutionFlags.Active)
	assert.True(t, testSolution.IsActive())
	testSolution.Flags = new(SolutionFlags.Released | *testSolution.Flags)
	assert.True(t, testSolution.IsActive())
}

func TestSolution_IsAfter(t *testing.T) {
	var branchOem = "oem"
	var branchHandle = "handle"
	var testSolution1 = &Solution{
		Oem:    branchOem,
		Handle: branchHandle,
		Seq:    new(37),
	}
	var testSolution2 = &Solution{
		Oem:    branchOem,
		Handle: branchHandle,
		Seq:    new(1337),
	}

	assert.False(t, testSolution1.IsAfter(testSolution2))
	assert.True(t, testSolution2.IsAfter(testSolution1))
}

func TestSolution_IsAfter_DifferentBranch(t *testing.T) {
	var branchOem = "oem"
	var branchHandle = "handle"
	var testSolution1 = &Solution{
		Oem:    branchOem,
		Handle: branchHandle,
	}
	var testSolution2 = &Solution{
		Oem:    branchOem,
		Handle: "different",
	}

	assert.False(t, testSolution1.IsAfter(testSolution2))
}

func TestSolution_IsAfter_NoSequence(t *testing.T) {
	var branchOem = "oem"
	var branchHandle = "handle"
	var testSolution1 = &Solution{
		Oem:    branchOem,
		Handle: branchHandle,
	}
	var testSolution2 = &Solution{
		Oem:    branchOem,
		Handle: branchHandle,
	}

	assert.False(t, testSolution1.IsAfter(testSolution2))
}

func TestSolution_IsAfter_Partial(t *testing.T) {
	var branchOem = "oem"
	var branchHandle = "handle"
	var testSolution1 = &Solution{
		Oem:    branchOem,
		Handle: branchHandle,
		Seq:    new(37),
	}
	var testSolution2 = &Solution{
		Oem:    branchOem,
		Handle: branchHandle,
	}

	assert.True(t, testSolution1.IsAfter(testSolution2))
}

func TestSolution_Merge(t *testing.T) {
	var expectedDescription = "description"
	var expectedHandle = "handle"
	var expectedName = "name"
	var expectedOem = "oem"
	var expectedVersion = "version"
	var testSolution = &Solution{
		Handle: expectedHandle,
		Oem:    expectedOem,
	}
	var testUpdate = &Solution{
		Description: expectedDescription,
		Name:        expectedName,
		Version:     expectedVersion,
	}
	var actual = testSolution.Merge(*testUpdate)

	assert.Equal(t, expectedDescription, actual.Description)
	assert.Equal(t, expectedHandle, actual.Handle)
	assert.Equal(t, expectedName, actual.Name)
	assert.Equal(t, expectedOem, actual.Oem)
	assert.Equal(t, expectedVersion, actual.Version)
	assert.Empty(t, actual.Workflows)
}

func TestSolution_MergeWorkflows(t *testing.T) {
	var expectedDescription = "description"
	var expectedHandle = "handle"
	var expectedName = "name"
	var expectedOem = "oem"
	var expectedVersion = "version"
	var mergedWorkflow = "merged"
	var sourceWorkflow = "source"
	var testSolution = &Solution{
		Handle:      expectedHandle,
		Oem:         expectedOem,
		Description: expectedDescription,
		Name:        expectedName,
		Version:     expectedVersion,
		Workflows: []Workflow{
			{
				Handle: sourceWorkflow,
			},
		},
	}
	var testUpdate = &Solution{
		Workflows: []Workflow{
			{
				Handle: mergedWorkflow,
			},
		},
	}
	var actual = testSolution.Merge(*testUpdate)

	assert.Equal(t, expectedDescription, actual.Description)
	assert.Equal(t, expectedHandle, actual.Handle)
	assert.Equal(t, expectedName, actual.Name)
	assert.Equal(t, expectedOem, actual.Oem)
	assert.Equal(t, expectedVersion, actual.Version)
	assert.Equal(t, 2, len(actual.Workflows))
	assert.Equal(t, sourceWorkflow, actual.Workflows[0].Handle)
	assert.Equal(t, mergedWorkflow, actual.Workflows[1].Handle)
}

func TestWorkflow_ContainsNode(t *testing.T) {
	var expectedHandle = "nodeHandle"
	var testWorkflow = &Workflow{
		Nodes: []WorkflowNode{
			{
				Handle: expectedHandle,
			},
		},
	}

	assert.False(t, testWorkflow.ContainsNode("notHandle"))
	assert.True(t, testWorkflow.ContainsNode(expectedHandle))
}

func TestWorkflow_FindNodeByHandle(t *testing.T) {
	var expectedHandle = "handle"
	var expectedName = "name"
	var testWorkflow = &Workflow{
		Nodes: []WorkflowNode{
			{
				Handle: expectedHandle,
				Name:   expectedName,
			},
		},
	}
	var actual, err = testWorkflow.FindNodeByHandle(expectedHandle)

	assert.NoError(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, expectedHandle, actual.Handle)
	assert.Equal(t, expectedName, actual.Name)
}

func TestWorkflow_FindNodeByHandle_NotFound(t *testing.T) {
	var testWorkflow = &Workflow{}
	var actual, err = testWorkflow.FindNodeByHandle("notFound")

	assert.Nil(t, actual)
	assert.ErrorIs(t, err, ErrorWorkflowNodeNotFound)
}

func TestWorkflow_FindNodeBySf(t *testing.T) {
	var expectedNodeHandle = "nodeHandle"
	var testFunction = &Function{
		Oem:     "oem",
		Handle:  "functionHandle",
		Version: "version",
	}
	var testWorkflow = &Workflow{
		Nodes: []WorkflowNode{
			{
				Handle: "noSF",
			},
			{
				Handle: expectedNodeHandle,
				Sf: &WorkflowNodeFunction{
					Oem:     testFunction.Oem,
					Handle:  testFunction.Handle,
					Version: testFunction.Version,
				},
			},
		},
	}
	var actual, err = testWorkflow.FindNodeBySf(testFunction)

	assert.NoError(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, expectedNodeHandle, actual.Handle)
	assert.Equal(t, testFunction.Oem, actual.Sf.Oem)
	assert.Equal(t, testFunction.Handle, actual.Sf.Handle)
	assert.Equal(t, testFunction.Version, actual.Sf.Version)
}

func TestWorkflow_FindNodeBySf_NotFound(t *testing.T) {
	var testFunction = &Function{
		Oem:     "oem",
		Handle:  "functionHandle",
		Version: "version",
	}
	var testWorkflow = &Workflow{}
	var actual, err = testWorkflow.FindNodeBySf(testFunction)

	assert.Error(t, err)
	assert.Nil(t, actual)
}

func TestWorkflow_FindNodeHandleBySf(t *testing.T) {
	var expectedHandle = "nodeHandle"
	var expectedSfOem = "sfOem"
	var expectedSfHandle = "sfHandle"
	var expectedSfVersion = "sfVersion"
	var testWorkflow = &Workflow{
		Nodes: []WorkflowNode{
			{
				Handle: "notTheRightHandle",
			},
			{
				Handle: expectedHandle,
				Sf: &WorkflowNodeFunction{
					Oem:     expectedSfOem,
					Handle:  expectedSfHandle,
					Version: expectedSfVersion,
				},
			},
		},
	}

	actual, err := testWorkflow.FindNodeHandleBySf("", "", "")
	assert.Empty(t, actual)
	assert.Error(t, err)
	actual, err = testWorkflow.FindNodeHandleBySf(expectedSfOem, "", "")
	assert.Empty(t, actual)
	assert.Error(t, err)
	actual, err = testWorkflow.FindNodeHandleBySf(expectedSfOem, expectedSfHandle, "")
	assert.Empty(t, actual)
	assert.Error(t, err)
	actual, err = testWorkflow.FindNodeHandleBySf(expectedSfOem, expectedSfHandle, expectedSfVersion)
	assert.Equal(t, actual, expectedHandle)
}

func TestWorkflow_HasNodeProps(t *testing.T) {
	var testWorkflow = &Workflow{
		Nodes: []WorkflowNode{
			{
				Handle: "test",
				Props: map[string]string{
					"PROP": "value",
				},
			},
		},
	}

	assert.True(t, testWorkflow.HasNodeProps())
}

func TestWorkflow_HasNodeProps_NoProps(t *testing.T) {
	var testWorkflow = &Workflow{
		Nodes: []WorkflowNode{
			{
				Handle: "test",
			},
		},
	}

	assert.False(t, testWorkflow.HasNodeProps())
}

func TestWorkflow_HasNodeProps_NoNodes(t *testing.T) {
	var testWorkflow = &Workflow{}

	assert.False(t, testWorkflow.HasNodeProps())
}

func TestWorkflowHandlePredicate(t *testing.T) {
	var expectedHandle = "handle"
	var testWorkflow = Workflow{Handle: expectedHandle}

	assert.False(t, WorkflowHandlePredicate("notHandle")(testWorkflow))
	assert.True(t, WorkflowHandlePredicate(expectedHandle)(testWorkflow))
}

func TestWorkflowLink_Equals(t *testing.T) {
	var link1 = WorkflowLink{}
	var link2 = WorkflowLink{RhsNodePort: "port"}
	var link3 = WorkflowLink{RhsNode: "node"}
	var link4 = WorkflowLink{LhsNodePort: "otherPort"}

	assert.True(t, link1.Equals(link1))
	assert.False(t, link1.Equals(link2))
	assert.False(t, link1.Equals(link3))
	assert.False(t, link1.Equals(link4))
}

func TestWorkflowLink_String(t *testing.T) {
	var testLink = &WorkflowLink{}

	assert.Equal(t, ":", testLink.String())
	testLink.LhsNode = "expectedLeft"
	assert.Equal(t, "expectedLeft:", testLink.String())
	testLink.RhsNode = "expectedRight"
	assert.Equal(t, "expectedLeft:expectedRight", testLink.String())
	testLink.LhsNodePort = "leftPort"
	assert.Equal(t, "expectedLeft[leftPort]:expectedRight", testLink.String())
	testLink.RhsNodePort = "rightPort"
	assert.Equal(t, "expectedLeft[leftPort]:expectedRight[rightPort]", testLink.String())
}

func TestWorkflowNamePredicate(t *testing.T) {
	var expectedName = "name"
	var testWorkflow = Workflow{Name: expectedName}

	assert.False(t, WorkflowNamePredicate("notName")(testWorkflow))
	assert.True(t, WorkflowNamePredicate(expectedName)(testWorkflow))
}

func TestWorkflowNode_AssignProp(t *testing.T) {
	var expectedKey = "KEY"
	var expectedValue = "value"
	var workflowNode = &WorkflowNode{}

	workflowNode.AssignProp("key", "Keys are case insensitive")
	workflowNode.AssignProp(expectedKey, expectedValue)
	assert.NotNil(t, workflowNode.Props)
	assert.Equal(t, 1, len(workflowNode.Props))
	assert.Equal(t, expectedValue, workflowNode.Props[expectedKey])
}

func TestWorkflowNode_Equals(t *testing.T) {
	var node1 = WorkflowNode{}
	var node2 = WorkflowNode{Handle: "handle"}

	assert.True(t, node1.Equals(node1))
	assert.False(t, node1.Equals(node2))
}

func TestWorkflowNode_HasProp(t *testing.T) {
	var expectedKey = "key"
	var expectedKey2 = "key2"
	var workflowNode = &WorkflowNode{
		Props: make(map[string]string),
	}

	workflowNode.Props[expectedKey] = "value"
	workflowNode.Props[strings.ToUpper(expectedKey2)] = "value"
	assert.NotNil(t, workflowNode.Props)
	assert.True(t, workflowNode.HasProp(expectedKey))
	assert.True(t, workflowNode.HasProp(strings.ToUpper(expectedKey2)))
	assert.False(t, workflowNode.HasProp("notFound"))
}

func TestWorkflowNode_NormalizeProps(t *testing.T) {
	var lowerCaseKey = "lower_key"
	var expectedValue = "value"
	var workflowNode = &WorkflowNode{}

	workflowNode.NormalizeProps()
	assert.Empty(t, workflowNode.Props)
	workflowNode.Props = map[string]string{lowerCaseKey: expectedValue}
	workflowNode.NormalizeProps()
	assert.Equal(t, expectedValue, workflowNode.Props[strings.ToUpper(lowerCaseKey)])
}

func TestWorkflowNode_RemoveProp(t *testing.T) {
	var expectedKey = "KEY"
	var expectedValue = "value"
	var workflowNode = &WorkflowNode{}

	workflowNode.RemoveProp(expectedKey)
	workflowNode.AssignProp(expectedKey, expectedValue)
	workflowNode.RemoveProp(expectedKey)
	assert.NotNil(t, workflowNode.Props)
	assert.Empty(t, workflowNode.Props)
}

func TestWorkflowNode_ValidateProps(t *testing.T) {
	var workflowNode = &WorkflowNode{}
	var expectedKey = "KEY"
	var expectedValue = "VALUE"
	var varSpec = PropSpec{
		Key:  expectedKey,
		Type: "STRING",
	}

	assert.NoError(t, workflowNode.ValidateProps([]shared.VarSpec{}))
	workflowNode.AssignProp(expectedKey, expectedValue)
	assert.NoError(t, workflowNode.ValidateProps([]shared.VarSpec{varSpec}))
}

func TestWorkflowNode_ValidateProps_Error(t *testing.T) {
	var workflowNode = &WorkflowNode{}
	var expectedKey = "KEY"
	var expectedValue = "VALUE"
	var varSpec = PropSpec{
		Key:  expectedKey,
		Type: "INT",
	}

	workflowNode.AssignProp(expectedKey, expectedValue)
	assert.Error(t, workflowNode.ValidateProps([]shared.VarSpec{}))
	assert.Error(t, workflowNode.ValidateProps([]shared.VarSpec{varSpec}))
}

func TestWorkflowNodeFunction_IsEqual(t *testing.T) {
	var expectedOem = "oem"
	var expectedHandle = "handle"
	var expectedVersion = "ver"
	var testNodeFn = &WorkflowNodeFunction{}
	var testFn = &Function{}

	assert.False(t, testNodeFn.IsEqual(nil))
	assert.True(t, testNodeFn.IsEqual(testFn))
	testNodeFn.Oem = expectedOem
	assert.False(t, testNodeFn.IsEqual(testFn))
	testFn.Oem = expectedOem
	assert.True(t, testNodeFn.IsEqual(testFn))
	testNodeFn.Handle = expectedHandle
	assert.False(t, testNodeFn.IsEqual(testFn))
	testFn.Handle = expectedHandle
	assert.True(t, testNodeFn.IsEqual(testFn))
	testNodeFn.Version = expectedVersion
	assert.False(t, testNodeFn.IsEqual(testFn))
	testFn.Version = expectedVersion
	assert.True(t, testNodeFn.IsEqual(testFn))
}

func TestWorkspace_IsActive(t *testing.T) {
	var testWorkspace = &Workspace{}

	assert.False(t, testWorkspace.IsActive())
	testWorkspace.Flags = new(WorkspaceFlags.RcEnabled)
	assert.False(t, testWorkspace.IsActive())
	testWorkspace.Flags = new(WorkspaceFlags.Active | WorkspaceFlags.RcEnabled)
	assert.True(t, testWorkspace.IsActive())
}

func TestWorkspace_IsRcEnabled(t *testing.T) {
	var testWorkspace = &Workspace{}

	assert.False(t, testWorkspace.IsRcEnabled())
	testWorkspace.RcEnabled = true
	assert.True(t, testWorkspace.IsRcEnabled())
	testWorkspace.Flags = new(WorkspaceFlags.Active)
	assert.False(t, testWorkspace.IsRcEnabled())
	testWorkspace.Flags = new(WorkspaceFlags.Active | WorkspaceFlags.RcEnabled)
	assert.True(t, testWorkspace.IsRcEnabled())
}

func TestWorkspace_MarshalJSON(t *testing.T) {
	var testWorkspace = &Workspace{
		Name:        "expectedName",
		Description: "expectedDesc",
		Visibility:  "expectedlyVisible",
		Flags:       new(37),
	}

	bytes, err := testWorkspace.MarshalJSON()
	assert.NoError(t, err)
	actual := string(bytes)
	assert.Contains(t, actual, testWorkspace.Name)
	assert.Contains(t, actual, testWorkspace.Description)
	assert.Contains(t, actual, strings.ToUpper(testWorkspace.Visibility))
	assert.Contains(t, actual, cast.ToString(testWorkspace.Flags))
}

func TestWorkspace_MarshalJSON_DefaultActive(t *testing.T) {
	var testWorkspace = &Workspace{
		Name:        "expectedName",
		Description: "expectedDesc",
		Visibility:  "expectedlyVisible",
	}

	bytes, err := testWorkspace.MarshalJSON()
	assert.NoError(t, err)
	actual := string(bytes)
	assert.Contains(t, actual, testWorkspace.Name)
	assert.Contains(t, actual, testWorkspace.Description)
	assert.Contains(t, actual, strings.ToUpper(testWorkspace.Visibility))
	assert.Contains(t, actual, cast.ToString(WorkspaceFlags.Active))
}

func TestWorkspace_MarshalJSON_RcEnabled(t *testing.T) {
	var testWorkspace = &Workspace{
		Name:        "expectedName",
		Description: "expectedDesc",
		Visibility:  "expectedlyVisible",
		RcEnabled:   true,
	}

	bytes, err := testWorkspace.MarshalJSON()
	assert.NoError(t, err)
	actual := string(bytes)
	assert.Contains(t, actual, testWorkspace.Name)
	assert.Contains(t, actual, testWorkspace.Description)
	assert.Contains(t, actual, strings.ToUpper(testWorkspace.Visibility))
	assert.Contains(t, actual, cast.ToString(WorkspaceFlags.RcEnabled|WorkspaceFlags.Active))
}

func TestParseFqdnVersion(t *testing.T) {
	var expectedOem = "oem"
	var expectedHandle = "handle"
	var expectedVersion = "version"
	var actualOem, actualHandle, actualVersion = ParseFqdnVersion(
		fmt.Sprintf("%s/%s:%s", expectedOem, expectedHandle, expectedVersion))

	assert.Equal(t, expectedOem, actualOem)
	assert.Equal(t, expectedHandle, actualHandle)
	assert.Equal(t, expectedVersion, actualVersion)
}

func TestParseFqdnVersion_SingleString(t *testing.T) {
	var expectedSingle = "single"
	var oem, handle, ver = ParseFqdnVersion(expectedSingle)

	assert.Equal(t, expectedSingle, oem)
	assert.Empty(t, handle)
	assert.Empty(t, ver)
}
