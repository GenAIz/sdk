package broker

import (
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

func TestDataLink_FindSecretSpec(t *testing.T) {
	var testKey = "testKey"
	var testLink = &DataLink{}

	assert.Nil(t, testLink.FindSecretSpec(testKey))
	testLink.SecretSpecs = []PropSpec{{Key: testKey}}
	assert.Equal(t, &testLink.SecretSpecs[0], testLink.FindSecretSpec(testKey))
}

func TestDataLink_IsActive(t *testing.T) {
	var testLink = &DataLink{}

	assert.False(t, testLink.IsActive())
	testLink.Flags = DataLinkFlags.Active
	assert.True(t, testLink.IsActive())
}

func TestDataLink_RemovePropSpec(t *testing.T) {
	var testProp = &PropSpec{Key: "testKey"}
	var testLink = &DataLink{PropSpecs: []PropSpec{*testProp}}

	assert.Nil(t, testLink.RemovePropSpec("wrongKey"))
	assert.Equal(t, testProp, testLink.RemovePropSpec(testProp.Key))
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

func TestDataLink_RemoveSecretSpec(t *testing.T) {
	var testProp = &PropSpec{Key: "testKey"}
	var testLink = &DataLink{SecretSpecs: []PropSpec{*testProp}}

	assert.Nil(t, testLink.RemoveSecretSpec("wrongKey"))
	assert.Equal(t, testProp, testLink.RemoveSecretSpec(testProp.Key))
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

func TestFindPropSpec(t *testing.T) {
	var expectedKey = "expectedKey"
	var expectedSpecMap = map[string]any{
		"key":         expectedKey,
		"name":        "expectedName",
		"description": "expectedDescription",
		"value":       "expectedValue",
		"values":      []string{"expected1", "expected2"},
		"type":        PropSpecTypeInt,
	}

	assert.Empty(t, FindPropSpec("notAList", expectedKey))
	if actual := FindPropSpec([]interface{}{expectedSpecMap, "notASpec"}, expectedKey); actual != nil {
		assert.Equal(t, expectedSpecMap["description"], actual.Description)
		assert.Equal(t, expectedSpecMap["key"], actual.Key)
		assert.Equal(t, expectedSpecMap["name"], actual.Name)
		assert.Equal(t, expectedSpecMap["value"], actual.Value)
		assert.Equal(t, expectedSpecMap["values"], actual.Values)
		assert.Equal(t, expectedSpecMap["type"], actual.Type)
	} else {
		assert.Fail(t, "propSpec not found")
	}
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

func TestSolution_asIdentity(t *testing.T) {
	var actual *shared.Identity
	var solution = SolutionRemote{
		Solution: Solution{
			Version: "version",
		},
		Id:     37,
		Digest: "digest",
		Fqdn:   "path",
	}

	actual = solution.asIdentity()
	assert.Equal(t, solution.Id, cast.ToInt64(actual.Id))
	assert.Equal(t, solution.Digest, actual.Hash)
	assert.Equal(t, solution.Fqdn, actual.Path)
	assert.Equal(t, solution.Version, actual.Version)
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

func TestWorkflowNamePredicate(t *testing.T) {
	var expectedName = "name"
	var testWorkflow = Workflow{Name: expectedName}

	assert.False(t, WorkflowNamePredicate("notName")(testWorkflow))
	assert.True(t, WorkflowNamePredicate(expectedName)(testWorkflow))
}

func TestWorkflowNode_Equals(t *testing.T) {
	var node1 = WorkflowNode{}
	var node2 = WorkflowNode{Handle: "handle"}

	assert.True(t, node1.Equals(node1))
	assert.False(t, node1.Equals(node2))
}
