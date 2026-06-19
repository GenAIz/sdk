package broker

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

type mockSolutionWriter struct {
	testOutput   string
	testSolution *Solution
	testWorkflow *Workflow
	writeError   error
}

func (msw *mockSolutionWriter) BuildSolution() (string, Solution) {
	if msw.testSolution != nil {
		return "solution", *msw.testSolution
	}

	return "", Solution{}
}

func (msw *mockSolutionWriter) GetWorkflowByHandle(testHandle string) (*Workflow, error) {
	if msw.testWorkflow != nil && msw.testWorkflow.Handle == testHandle {
		return msw.testWorkflow, nil
	}

	return nil, errors.New("not existing")
}

func (msw *mockSolutionWriter) WithSolution(testSolution *Solution) SolutionWriter {
	msw.testSolution = testSolution
	return msw
}

func (msw *mockSolutionWriter) WithWorkflow(testWorkflow *Workflow) SolutionWriter {
	msw.testWorkflow = testWorkflow
	return msw
}

func (msw *mockSolutionWriter) Write(testOutput string) error {
	msw.testOutput = testOutput

	if msw.writeError != nil {
		return msw.writeError
	}

	return nil
}

type mockWorkflowWriter struct {
	testOutput      string
	resultLinks     map[string][]WorkflowLink
	resultNodes     map[string][]WorkflowNode
	resultWorkflows []Workflow
	testWorkflows   []Workflow
	writeError      error
}

func (mww *mockWorkflowWriter) BuildWorkflows() (string, []Workflow) {
	return "workflow", mww.resultWorkflows
}

func (mww *mockWorkflowWriter) GetWorkflowByHandle(handle string) (*Workflow, error) {
	var i = slices.IndexFunc(mww.testWorkflows, func(workflow Workflow) bool {
		return workflow.Handle == handle
	})

	if i >= 0 {
		return &mww.testWorkflows[i], nil
	}

	return nil, errors.New("not found")
}

func (mww *mockWorkflowWriter) GetWorkflows() []Workflow {
	return mww.testWorkflows
}

func (mww *mockWorkflowWriter) WithWorkflow(testWorkflow *Workflow) WorkflowWriter {
	mww.resultWorkflows = append(mww.resultWorkflows, *testWorkflow)
	return mww
}

func (mww *mockWorkflowWriter) WithWorkflows(testWorkflows []Workflow) WorkflowWriter {
	mww.testWorkflows = testWorkflows
	mww.resultWorkflows = testWorkflows
	return mww
}

func (mww *mockWorkflowWriter) WithWorkflowLinks(handle string, links []WorkflowLink) WorkflowWriter {
	if mww.resultLinks == nil {
		mww.resultLinks = make(map[string][]WorkflowLink)
	}

	mww.resultLinks[handle] = links
	return mww
}

func (mww *mockWorkflowWriter) WithWorkflowNodes(handle string, nodes []WorkflowNode) WorkflowWriter {
	if mww.resultNodes == nil {
		mww.resultNodes = make(map[string][]WorkflowNode)
	}

	mww.resultNodes[handle] = nodes
	return mww
}

func (mww *mockWorkflowWriter) Write(testOutput string) error {
	mww.testOutput = testOutput

	if mww.writeError != nil {
		return mww.writeError
	}

	return nil
}

type stubSolutionClient struct {
	client
	brokerAddr      string
	listError       error
	listSolutions   []Solution
	publishError    error
	publishIdentity *shared.Identity
}

func (ssc stubSolutionClient) GetHostAddr() string {
	return ssc.brokerAddr
}

func (ssc stubSolutionClient) ListSolutions(string) ([]Solution, error) {
	return ssc.listSolutions, ssc.listError
}

func (ssc stubSolutionClient) PublishSolution(*Solution) (*shared.Identity, error) {
	if ssc.publishError != nil {
		return nil, ssc.publishError
	}

	if ssc.publishIdentity != nil {
		return ssc.publishIdentity, nil
	}

	return nil, nil
}

func TestSolutionParams_HasWorkflows(t *testing.T) {
	var testParams = &SolutionParams{}

	assert.False(t, testParams.HasWorkflows())
	testParams.Solution = &Solution{}
	assert.False(t, testParams.HasWorkflows())
	testParams.Solution = &Solution{
		Workflows: []Workflow{
			{
				Handle: "handle",
			},
		},
	}
	assert.True(t, testParams.HasWorkflows())
}

func TestSolutionPublishParams_HasProvision(t *testing.T) {
	var expectedHandle = "handle"
	var expectedOem = "oem"
	var testParams = &SolutionPublishParams{
		Provisions: []ProvisionParams{
			{
				Handle: expectedHandle,
				Oem:    expectedOem,
			},
		},
	}

	assert.True(t, testParams.HasProvision(expectedOem, expectedHandle))
	assert.False(t, testParams.HasProvision(expectedOem, "notAHandle"))
	assert.False(t, testParams.HasProvision("notAOem", "notAHandle"))
}

func TestWorkflowParams_GetHandle(t *testing.T) {
	var expectedHandle = "handle"
	var testWorkflow = &Workflow{
		Handle: expectedHandle,
	}

	assert.Empty(t, WorkflowParams{}.GetHandle())
	assert.Equal(t, expectedHandle, WorkflowParams{Workflow: testWorkflow}.GetHandle())
}

func TestWorkflowParams_GetName(t *testing.T) {
	var expectedName = "name"
	var testWorkflow = &Workflow{
		Name: expectedName,
	}

	assert.Empty(t, WorkflowParams{}.GetName())
	assert.Equal(t, expectedName, WorkflowParams{Workflow: testWorkflow}.GetName())
}

func TestWorkflowParams_workflowPredicate(t *testing.T) {
	var testParams = &WorkflowParams{
		Workflow: &Workflow{},
	}

	assert.False(t, testParams.workflowPredicate()(Workflow{}))
	testParams.Name = "name"
	assert.True(t, testParams.workflowPredicate()(Workflow{Name: testParams.Name}))
	testParams.Handle = "handle"
	assert.True(t, testParams.workflowPredicate()(Workflow{Handle: testParams.Handle}))
}

func TestNewSolutionListTask(t *testing.T) {
	var testTask = NewSolutionListTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnIncomplete)
	assert.NotEmpty(t, testTask.OnPrepare)
}

func TestNewSolutionPublishTask(t *testing.T) {
	var testTask = NewSolutionPublishTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPretend)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.Empty(t, testTask.OnIncomplete)
	assert.NotEmpty(t, testTask.OnPrepare)
}

func TestNewSolutionReduceTask(t *testing.T) {
	var testTask = NewSolutionReduceTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnPrepare)
}

func TestNewSolutionUpdateTask(t *testing.T) {
	var testTask = NewSolutionUpdateTask(nil)

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPretend)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnIncomplete)
	assert.NotEmpty(t, testTask.OnPrepare)
}

func TestNewWorkflowDeleteTask(t *testing.T) {
	var testTask = NewWorkflowDeleteTask(nil)

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPretend)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.Empty(t, testTask.OnIncomplete)
	assert.NotEmpty(t, testTask.OnPrepare)
}

func TestNewWorkflowPropTask(t *testing.T) {
	var testTask = NewWorkflowPropTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPretend)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnIncomplete)
}

func TestNewWorkflowUpdateTask(t *testing.T) {
	var testTask = NewWorkflowUpdateTask(nil)

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPretend)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnIncomplete)
	assert.NotEmpty(t, testTask.OnPrepare)
}

func Test_handleSolutionCreateConfig(t *testing.T) {
	var expectedOutput = "output"
	var mockWriter = &mockSolutionWriter{}
	var testParams = &SolutionParams{
		ConfigParams: shared.ConfigParams{
			ConfigType:   lang.Ref(shared.ConfigTypeJson),
			ConfigFolder: "folder",
			ConfigName:   "name",
		},
		Solution: &Solution{
			Handle: "handle",
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: expectedOutput,
	}

	assert.NoError(t, handleSolutionCreateConfig(mockWriter, testParams, testState))
	assert.Same(t, testParams.Solution, mockWriter.testSolution)
	actual := strings.Join(testState.Reports, "\n")
	assert.Contains(t, actual, testParams.Handle)
	assert.Contains(t, actual, testParams.ConfigFolder)
	assert.NotContains(t, actual, testParams.ConfigName)
	assert.NotContains(t, actual, shared.ConfigTypeJson)
}

func Test_handleSolutionCreateConfig_Error(t *testing.T) {
	var expectedError = errors.New("expected")
	var expectedOutput = "output"
	var mockWriter = &mockSolutionWriter{writeError: expectedError}
	var testParams = &SolutionParams{
		Solution: &Solution{},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: expectedOutput,
	}

	assert.ErrorIs(t, handleSolutionCreateConfig(mockWriter, testParams, testState), expectedError)
	assert.Same(t, testParams.Solution, mockWriter.testSolution)
	assert.Equal(t, expectedOutput, mockWriter.testOutput)
}

func Test_handleSolutionCreateConfig_NoFileOutput(t *testing.T) {
	assert.ErrorIs(t, handleSolutionCreateConfig(nil, &SolutionParams{}, &task.State{}), errorSolutionFileInvalid)
}

func Test_handleSolutionCreateContext(t *testing.T) {
	var expectedOutput = "test." + shared.ConfigTypeYaml
	var testParams = &SolutionParams{
		ConfigParams: shared.ConfigParams{
			ConfigName: "test",
			ConfigType: lang.Ref(shared.ConfigTypeYaml),
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.NoError(t, handleSolutionCreateContext(testParams, testState))
	assert.Equal(t, expectedOutput, testState.Output)
}

func Test_handleSolutionCreateContext_InvalidFileOutput(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.Error(t, handleSolutionCreateContext(&SolutionParams{}, testState))
}

func Test_handleSolutionCreateContext_OutputAlreadySelected(t *testing.T) {
	var testState = &task.State{
		Output: "output",
	}

	assert.NoError(t, handleSolutionCreateContext(&SolutionParams{}, testState))
}

func Test_handleSolutionListContext(t *testing.T) {
	var expectedBrokerAddr = "brokerAddr"
	var restoredFactory = clientFactory.Active
	var testLogger, testHook = test.NewNullLogger()
	var testParams = &SolutionListParams{
		Oem: "oem",
	}
	var testState = &task.State{Logger: testLogger}

	defer func() {
		clientFactory.Active = restoredFactory
	}()
	clientFactory.Active = func(authFile string) (Client, error) {
		return &stubSolutionClient{
			brokerAddr: expectedBrokerAddr,
		}, nil
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.NoError(t, handleSolutionListContext(testParams, testState))
	assert.Equal(t, 0, len(testHook.Entries))
	assert.Equal(t, expectedBrokerAddr, testState.Output)
}

func Test_handleSolutionListContext_AccountOnly(t *testing.T) {
	var expectedBrokerAddr = "brokerAddr"
	var restoredFactory = clientFactory.Active
	var testLogger, testHook = test.NewNullLogger()
	var testParams = &SolutionListParams{
		Oem:         "oem",
		AccountOnly: true,
	}
	var testState = &task.State{Logger: testLogger}

	defer func() {
		clientFactory.Active = restoredFactory
	}()
	clientFactory.Active = func(authFile string) (Client, error) {
		return &stubSolutionClient{
			brokerAddr: expectedBrokerAddr,
		}, nil
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.NoError(t, handleSolutionListContext(testParams, testState))
	assert.Equal(t, 1, len(testHook.Entries))
	assert.Equal(t, expectedBrokerAddr, testState.Output)
}

func Test_handleSolutionListContext_CheckedOutput(t *testing.T) {
	var testState = &task.State{Output: "brokerAddr"}

	assert.Nil(t, handleSolutionListContext(&SolutionListParams{}, testState))
}

func Test_handleSolutionListContext_LocalAccountOnlyConflict(t *testing.T) {
	var expectedBrokerAddr = "brokerAddr"
	var restoredFactory = clientFactory.Active
	var testLogger, testHook = test.NewNullLogger()
	var testParams = &SolutionListParams{
		Oem:         "oem",
		AccountOnly: true,
		Local: []Solution{
			{
				Id: lang.Ref(int64(37)),
			},
		},
	}
	var testState = &task.State{Logger: testLogger}

	defer func() {
		clientFactory.Active = restoredFactory
	}()
	clientFactory.Active = func(authFile string) (Client, error) {
		return &stubSolutionClient{
			brokerAddr: expectedBrokerAddr,
		}, nil
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.NoError(t, handleSolutionListContext(testParams, testState))
	assert.Equal(t, 2, len(testHook.Entries))
	assert.Equal(t, expectedBrokerAddr, testState.Output)
}

func Test_handleSolutionListContext_NoOem(t *testing.T) {
	assert.ErrorIs(t, handleSolutionListContext(&SolutionListParams{}, &task.State{}), errorSolutionOemRequired)
}

func Test_handleSolutionListContext_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var restoredFactory = clientFactory.Active
	var testParams = &SolutionListParams{Oem: "oem"}
	var testState = &task.State{}

	defer func() {
		clientFactory.Active = restoredFactory
	}()
	clientFactory.Active = func(authFile string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleSolutionListContext(testParams, testState), expectedError)
}

func Test_handleSolutionListComplete(t *testing.T) {
	var restoredFactory = clientFactory.Active
	var expectedSolutions = []Solution{
		{
			Id: lang.Ref(int64(37)),
		},
	}
	var testParams = &SolutionListParams{Oem: "oem"}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	defer func() {
		clientFactory.Active = restoredFactory
	}()
	clientFactory.Active = func(authFile string) (Client, error) {
		return &stubSolutionClient{
			listSolutions: expectedSolutions,
		}, nil
	}

	assert.NoError(t, handleSolutionListComplete(testParams, testState))

	if actual, ok := testState.Internal.([]Solution); ok {
		assert.Equal(t, expectedSolutions, actual)
	} else {
		assert.Fail(t, "expected a list of solutions")
	}

}

func Test_handleSolutionListComplete_Failure(t *testing.T) {
	var expectedError = errors.New("expected")
	var restoredFactory = clientFactory.Active
	var testParams = &SolutionListParams{Oem: "oem"}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	defer func() {
		clientFactory.Active = restoredFactory
	}()
	clientFactory.Active = func(authFile string) (Client, error) {
		return &stubSolutionClient{
			listError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleSolutionListComplete(testParams, testState), expectedError)
}

func Test_handleSolutionListComplete_InternalSkip(t *testing.T) {
	var testState = &task.State{Internal: "something"}

	assert.Nil(t, handleSolutionListComplete(&SolutionListParams{}, testState))
}

func Test_handleSolutionListComplete_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var restoredFactory = clientFactory.Active
	var testParams = &SolutionListParams{Oem: "oem"}
	var testState = &task.State{}

	defer func() {
		clientFactory.Active = restoredFactory
	}()
	clientFactory.Active = func(authFile string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleSolutionListComplete(testParams, testState), expectedError)
}

func Test_handleSolutionListIncomplete(t *testing.T) {
	var testParams = &SolutionListParams{}
	var testState = &task.State{
		Error:  errors.New("expected"),
		Logger: logrus.New(),
	}

	assert.NoError(t, handleSolutionListIncomplete(testParams, testState))
}

func Test_handleSolutionListIncomplete_AccountOnlyError(t *testing.T) {
	var testParams = &SolutionListParams{AccountOnly: true}
	var testState = &task.State{
		Error:  errors.New("expected"),
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleSolutionListIncomplete(testParams, testState), testState.Error)
}

func Test_handleSolutionListIncomplete_NoError(t *testing.T) {
	assert.NoError(t, handleSolutionListIncomplete(&SolutionListParams{}, &task.State{}))
}

func Test_handleSolutionPublishComplete(t *testing.T) {
	var testParams = &SolutionPublishParams{
		Solution: &Solution{
			Handle: "handle",
			Oem:    "oem",
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var restoredFactory = clientFactory.Active
	var expectedPath = "solutionPath"
	var expectedAddr = "mockAddr"

	defer func() {
		clientFactory.Active = restoredFactory
	}()
	clientFactory.Active = func(authFile string) (Client, error) {
		return &stubSolutionClient{
			brokerAddr: expectedAddr,
			publishIdentity: &shared.Identity{
				Path:    expectedPath,
				Version: "version",
			},
		}, nil
	}

	assert.NoError(t, handleSolutionPublishComplete(testParams, testState))
	actual := strings.Join(testState.Reports, "\n")
	assert.Contains(t, actual, expectedPath)
	assert.Contains(t, actual, expectedAddr)
}

func Test_handleSolutionPublishComplete_Failure(t *testing.T) {
	var expectedError = errors.New("expected")
	var restoredFactory = clientFactory.Active
	var testParams = &SolutionPublishParams{
		Solution: &Solution{},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	defer func() {
		clientFactory.Active = restoredFactory
	}()
	clientFactory.Active = func(authFile string) (Client, error) {
		return &stubSolutionClient{
			publishError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleSolutionPublishComplete(testParams, testState), expectedError)
}

func Test_handleSolutionPublishComplete_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var restoredFactory = clientFactory.Active
	var testParams = &SolutionPublishParams{
		Solution: &Solution{},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	defer func() {
		clientFactory.Active = restoredFactory
	}()
	clientFactory.Active = func(authFile string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleSolutionPublishComplete(testParams, testState), expectedError)
}

func Test_handleSolutionPublishContext(t *testing.T) {
	var expectedOem = "oem"
	var expectedHandle = "handle"
	var testParams = &SolutionPublishParams{
		Provisions: []ProvisionParams{
			{
				Oem:    expectedOem,
				Handle: expectedHandle,
			},
		},
		Solution: &Solution{
			Workflows: []Workflow{
				{
					Handle: "Handle",
					Nodes: []WorkflowNode{
						{
							Handle: "Node1",
							Sf: &WorkflowNodeFunction{
								Handle: "sfHandle1",
								Oem:    "sfOem1",
							},
						},
						{
							Handle: "Node2",
						},
						{
							Handle: "Node3",
							Sf: &WorkflowNodeFunction{
								Handle: expectedHandle,
								Oem:    expectedOem,
							},
						},
					},
				},
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.NoError(t, handleSolutionPublishContext(testParams, testState))
}

func Test_handleSolutionPublishContext_InvalidLinks(t *testing.T) {
	var testParams = &SolutionPublishParams{
		Solution: &Solution{
			Workflows: []Workflow{
				{
					Handle: "NoNodesBad",
					Links: []WorkflowLink{
						{
							LhsNode:     "node1",
							RhsNode:     "node2",
							RhsNodePort: "port2",
						},
						{
							LhsNode:     "node2",
							LhsNodePort: "port1",
							RhsNode:     "node1",
						},
					},
					Nodes: []WorkflowNode{
						{
							Handle: "node1",
						},
						{
							Handle: "node2",
						},
					},
				},
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.Error(t, handleSolutionPublishContext(testParams, testState))
}

func Test_handleSolutionPublishContext_NoWorkflows(t *testing.T) {
	var testParams = &SolutionPublishParams{
		Solution: &Solution{
			Workflows: []Workflow{},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.NoError(t, handleSolutionPublishContext(testParams, testState))
}

func Test_handleSolutionPublishContext_NoSolution(t *testing.T) {
	assert.ErrorIs(t, handleSolutionPublishContext(&SolutionPublishParams{}, &task.State{}), errorSolutionInvalid)
}

func Test_handleSolutionPublishContext_NoWorkflowNodes(t *testing.T) {
	var testParams = &SolutionPublishParams{
		Solution: &Solution{
			Workflows: []Workflow{
				{
					Handle: "NoNodesBad",
				},
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.Error(t, handleSolutionPublishContext(testParams, testState))
}

func Test_handleSolutionPublishPretend(t *testing.T) {
	var expectedSolution = &Solution{
		Oem:    "testOem",
		Handle: "testHandle",
	}
	var testParams = &SolutionPublishParams{
		Solution: expectedSolution,
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var restoredFactory = clientFactory.Active
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	defer func() {
		clientFactory.Active = restoredFactory
	}()
	clientFactory.Active = func(authFile string) (Client, error) {
		return &stubSolutionClient{}, nil
	}

	assert.NoError(t, handleSolutionPublishPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, expectedSolution.Oem)
	assert.Contains(t, output, expectedSolution.Handle)
}

func Test_handleSolutionPublishPretend_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var restoredFactory = clientFactory.Active

	defer func() {
		clientFactory.Active = restoredFactory
	}()
	clientFactory.Active = func(authFile string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleSolutionPublishPretend(&SolutionPublishParams{}, &task.State{}), expectedError)
}

func Test_handleSolutionPublishPretend_StateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error: expectedError,
	}

	assert.ErrorIs(t, handleSolutionPublishPretend(&SolutionPublishParams{}, testState), expectedError)
}

func Test_handleSolutionReduceComplete(t *testing.T) {
	var expectedReleased = &Solution{
		Id:      lang.Ref(int64(37)),
		Fqdn:    lang.Ref("releasedFqdn"),
		Version: "1.3.37",
		Flags:   lang.Ref(SolutionFlags.Active | SolutionFlags.Released),
	}
	var expectedCandidate = &Solution{
		Id:      lang.Ref(int64(42)),
		Fqdn:    lang.Ref("candidateFqdn"),
		Version: "1.0.0",
		Seq:     lang.Ref(1),
		Flags:   lang.Ref(SolutionFlags.Active),
	}
	var testSolutions = []Solution{
		*expectedReleased,
		{
			Id:      lang.Ref(int64(372)),
			Fqdn:    lang.Ref("releasedFqdn"),
			Version: "1.3.37",
			// Makes no sense, but if it happens it should be reduced
			Seq:   lang.Ref(16),
			Flags: lang.Ref(SolutionFlags.Active),
		},
		*expectedCandidate,
		{
			Id:      lang.Ref(int64(42)),
			Fqdn:    lang.Ref("candidateFqdn"),
			Version: "1.0.0",
			Seq:     lang.Ref(0),
			Flags:   lang.Ref(SolutionFlags.Active),
		},
	}
	var testState = &task.State{
		Internal: testSolutions,
		Logger:   logrus.New(),
	}
	var testParams = &SolutionListParams{
		Local: []Solution{
			{
				Oem:     "oem",
				Handle:  "handle",
				Version: "1.0.0",
			},
		},
	}

	assert.NoError(t, handleSolutionReduceComplete(testParams, testState))

	if actual, ok := testState.Internal.([]Solution); ok {
		assert.Equal(t, 3, len(actual))
		assert.Equal(t, *expectedReleased, actual[0])
		assert.Equal(t, *expectedCandidate, actual[1])
		assert.Equal(t, testParams.Local[0], actual[2])
	} else {
		assert.Fail(t, "expected a list of solutions")
	}
}

func Test_handleSolutionReduceComplete_NoInternal(t *testing.T) {
	var testState = &task.State{}
	var testParams = &SolutionListParams{
		Local: []Solution{
			{
				Oem:    "oem",
				Handle: "handle",
			},
		},
	}

	assert.NoError(t, handleSolutionReduceComplete(testParams, testState))

	if actual, ok := testState.Internal.([]Solution); ok {
		assert.Equal(t, testParams.Local, actual)
	} else {
		assert.Fail(t, "expected a list of solutions")
	}
}

func Test_handleSolutionReduceContext(t *testing.T) {
	var testLogger, testHook = test.NewNullLogger()
	var testParams = &SolutionListParams{
		Local: []Solution{
			{
				Id: lang.Ref(int64(37)),
			},
		},
	}
	var testState = &task.State{
		Logger: testLogger,
	}

	assert.NoError(t, handleSolutionReduceContext(testParams, testState))
	assert.Equal(t, 0, len(testHook.Entries))
}

func Test_handleSolutionReduceContext_NoLocal(t *testing.T) {
	assert.NoError(t, handleSolutionReduceContext(&SolutionListParams{}, &task.State{}))
}

func Test_handleSolutionUpdateConfig(t *testing.T) {
	var expectedError = errors.New("expected")
	var expectedOutput = "output"
	var mockWriter = &mockSolutionWriter{
		testWorkflow: &Workflow{
			Handle: "default",
		},
		writeError: expectedError,
	}
	var testParams = &SolutionParams{
		Solution: &Solution{
			Workflows: []Workflow{
				{
					Handle: "default",
				},
				{
					Handle: "notDefault",
				},
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: expectedOutput,
	}

	assert.ErrorIs(t, handleSolutionUpdateConfig(mockWriter, testParams, testState), expectedError)
	assert.Same(t, testParams.Solution, mockWriter.testSolution)
	assert.Equal(t, expectedOutput, mockWriter.testOutput)
}

func Test_handleSolutionUpdateConfig_NoFileOutput(t *testing.T) {
	assert.ErrorIs(t, handleSolutionUpdateConfig(nil, &SolutionParams{}, &task.State{}), errorSolutionFileInvalid)
}

func Test_handleSolutionUpdateConfig_NewWorkflow(t *testing.T) {
	var expectedOutput = "output"
	var mockWriter = &mockSolutionWriter{}
	var testParams = &SolutionParams{
		Solution: &Solution{},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: expectedOutput,
	}

	assert.NoError(t, handleSolutionUpdateConfig(mockWriter, testParams, testState))
	assert.Same(t, testParams.Solution, mockWriter.testSolution)
	assert.Equal(t, expectedOutput, mockWriter.testOutput)
}

func Test_handleSolutionUpdatePretend(t *testing.T) {
	var testSolution = &Solution{
		Description: "testDescription",
		Handle:      "testHandle",
		Name:        "testName",
		Oem:         "testOem",
		Version:     "testVersion",
	}
	var testParams = &SolutionParams{
		Solution: testSolution,
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output.yaml",
	}
	var testWriter = &mockSolutionWriter{}
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	assert.NoError(t, handleSolutionUpdatePretend(testWriter, testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, testSolution.Description)
	assert.Contains(t, output, testSolution.Handle)
	assert.Contains(t, output, testSolution.Name)
	assert.Contains(t, output, testSolution.Oem)
	assert.Contains(t, output, testSolution.Version)
}

func Test_handleSolutionUpdatePretend_NoFileOutput(t *testing.T) {
	assert.ErrorIs(t, handleSolutionUpdatePretend(nil, &SolutionParams{}, &task.State{}), errorSolutionFileInvalid)
}

func Test_handleWorkflowCreateConfig(t *testing.T) {
	var expectedError = errors.New("expected")
	var expectedOutput = "output"
	var mockWriter = &mockWorkflowWriter{writeError: expectedError}
	var testParams = &WorkflowParams{
		Workflow: &Workflow{},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: expectedOutput,
	}

	assert.ErrorIs(t, handleWorkflowCreateConfig(mockWriter, testParams, testState), expectedError)
	assert.Equal(t, *testParams.Workflow, mockWriter.resultWorkflows[0])
	assert.Equal(t, expectedOutput, mockWriter.testOutput)
}

func Test_handleWorkflowCreateConfig_NoFileOutput(t *testing.T) {
	assert.ErrorIs(t, handleWorkflowCreateConfig(nil, &WorkflowParams{}, &task.State{}), errorWorkflowFileInvalid)
}

func Test_handleWorkflowCreateContext(t *testing.T) {
	var expectedOutput = "test." + shared.ConfigTypeYaml
	var testParams = &WorkflowParams{
		ConfigParams: shared.ConfigParams{
			ConfigName: "test",
			ConfigType: lang.Ref(shared.ConfigTypeYaml),
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.NoError(t, handleWorkflowCreateContext(testParams, testState))
	assert.Equal(t, expectedOutput, testState.Output)
}

func Test_handleWorkflowCreateContext_InvalidFileOutput(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.Error(t, handleWorkflowCreateContext(&WorkflowParams{}, testState))
}

func Test_handleWorkflowCreateContext_OutputAlreadySelected(t *testing.T) {
	var testState = &task.State{
		Output: "output",
	}

	assert.NoError(t, handleWorkflowCreateContext(&WorkflowParams{}, testState))
}

func Test_handleWorkflowDeleteConfig(t *testing.T) {
	var testWorkflow = []Workflow{
		{
			Handle: "testHandle",
		},
		{
			Handle: "notHandle",
		},
	}
	var mockWriter = &mockWorkflowWriter{testWorkflows: testWorkflow}
	var testParams = &WorkflowParams{
		Workflow: &Workflow{
			Handle: "testHandle",
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output.yaml",
	}

	assert.NoError(t, handleWorkflowDeleteConfig(mockWriter, testParams, testState))
	assert.Equal(t, 1, len(mockWriter.resultWorkflows))
	assert.Equal(t, testState.Output, mockWriter.testOutput)
}

func Test_handleWorkflowDeleteConfig_NoFileOutput(t *testing.T) {
	assert.ErrorIs(t, handleWorkflowDeleteConfig(nil, &WorkflowParams{}, &task.State{}), errorWorkflowFileInvalid)
}

func Test_handleWorkflowDeleteConfig_PathError(t *testing.T) {
	var testDir = t.TempDir()
	var mockWriter = &mockWorkflowWriter{
		testWorkflows: []Workflow{{
			Handle: "testHandle",
		}},
		writeError: errors.New("expected"),
	}
	var testParams = &WorkflowParams{
		Workflow: &Workflow{
			Handle: "testHandle",
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: filepath.Join(testDir, "not", "exist.yaml"),
	}

	assert.ErrorIs(t, handleWorkflowDeleteConfig(mockWriter, testParams, testState), mockWriter.writeError)
}

func Test_handleWorkflowDeleteContext(t *testing.T) {
	var testDir = t.TempDir()
	var testParams = &WorkflowParams{
		ConfigParams: shared.ConfigParams{
			ConfigFolder: testDir,
			ConfigName:   "test",
			ConfigType:   lang.Ref(shared.ConfigTypeYaml),
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	if fd, err := os.Create(filepath.Join(testDir, "test.yaml")); err == nil {
		defer filez.CloseSilently(fd)

		assert.NoError(t, handleWorkflowDeleteContext(testParams, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleWorkflowDeleteContext_NoConfigType(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.Error(t, handleWorkflowDeleteContext(&WorkflowParams{}, testState))
}

func Test_handleWorkflowDeleteContext_NoExistingFile(t *testing.T) {
	var testParams = &WorkflowParams{
		ConfigParams: shared.ConfigParams{
			ConfigType: lang.Ref(shared.ConfigTypeYaml),
			ConfigName: "test",
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleWorkflowDeleteContext(testParams, testState), errorWorkflowFileNotFound)
}

func Test_handleWorkflowDeletePretend(t *testing.T) {
	var expectedHandle = "wfHandle"
	var expectedName = "wfName"
	var testParams = &WorkflowParams{Workflow: &Workflow{}}
	var testState = &task.State{Output: "output.yaml"}
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	testParams.Workflow.Name = expectedName
	assert.NoError(t, handleWorkflowDeletePretend(testParams, testState))
	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Contains(t, output, expectedName)
	r, w, _ = os.Pipe()
	os.Stdout = w
	testParams.Workflow.Handle = expectedHandle
	assert.NoError(t, handleWorkflowDeletePretend(testParams, testState))
	_ = w.Close()
	b, _ = io.ReadAll(r)
	output = string(b)
	assert.Contains(t, output, expectedHandle)
}

func Test_handleWorkflowDeletePretend_NoFileOutput(t *testing.T) {
	assert.ErrorIs(t, handleWorkflowDeletePretend(&WorkflowParams{}, &task.State{}), errorWorkflowFileInvalid)
}

func Test_handleWorkflowDeletePretend_NoWorkflow(t *testing.T) {
	var testState = &task.State{
		Output: "output.yaml",
	}

	assert.ErrorIs(t, handleWorkflowDeletePretend(&WorkflowParams{}, testState), errorWorkflowNotFound)
}

func Test_handleWorkflowPropContext(t *testing.T) {
	var testParams = &WorkflowPropParams{
		Workflow: &Workflow{
			Nodes: []WorkflowNode{
				{
					Handle: "test",
					Props: map[string]string{
						"PROP": "value",
					},
				},
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Internal: shared.VarSpecTracking{
			VarSpecs: []shared.VarSpec{
				PropSpec{
					Key: "testKey",
				},
			},
		},
	}

	assert.NoError(t, handleWorkflowPropContext(testParams, testState))
}

func Test_handleWorkflowPropContext_NoVarSpecs(t *testing.T) {
	var testParams = &WorkflowPropParams{
		Workflow: &Workflow{
			Nodes: []WorkflowNode{
				{
					Handle: "test",
					Props: map[string]string{
						"PROP": "value",
					},
				},
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleWorkflowPropContext(testParams, testState), ErrorWorkflowPropIncomplete)
}

func Test_handleWorkflowPropContext_NoProps_NoVarSpecs(t *testing.T) {
	var testParams = &WorkflowPropParams{
		Workflow: &Workflow{},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.NoError(t, handleWorkflowPropContext(testParams, testState))
}

func Test_handleWorkflowPropContext_NoWorkflow(t *testing.T) {
	assert.ErrorIs(t, handleWorkflowPropContext(&WorkflowPropParams{}, &task.State{}), ErrorWorkflowNotFound)
}

func Test_handleWorkflowPropComplete(t *testing.T) {
	var testParams = &WorkflowPropParams{
		Workflow: &Workflow{
			Nodes: []WorkflowNode{
				{
					// We validate that no props is always valid
					Handle: "test",
				},
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.NoError(t, handleWorkflowPropComplete(testParams, testState))
}

func Test_handleWorkflowPropComplete_InvalidProps(t *testing.T) {
	var expectedValue = "notInteger"
	var expectedProp = "PROP"
	var testParams = &WorkflowPropParams{
		Workflow: &Workflow{
			Nodes: []WorkflowNode{
				{
					Handle: "test",
					Props: map[string]string{
						expectedProp: expectedValue,
					},
				},
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Internal: shared.VarSpecTracking{
			VarSpecs: []shared.VarSpec{
				PropSpec{
					Key:  expectedProp,
					Type: "int",
				},
			},
		},
	}

	actual := handleWorkflowPropComplete(testParams, testState)

	if actual != nil {
		assert.Contains(t, actual.Error(), expectedValue)
		assert.Contains(t, actual.Error(), expectedProp)
	} else {
		assert.Fail(t, "expected error")
	}
}

func Test_handleWorkflowPropComplete_NoNodes(t *testing.T) {
	var testParams = &WorkflowPropParams{
		Workflow: &Workflow{},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.NoError(t, handleWorkflowPropComplete(testParams, testState))
}

func Test_handleWorkflowPropComplete_NoWorkflow(t *testing.T) {
	assert.ErrorIs(t, handleWorkflowPropComplete(&WorkflowPropParams{}, &task.State{}), ErrorWorkflowNotFound)
}

func Test_handleWorkflowPropIncomplete(t *testing.T) {
	var expectedProp = "PROP"
	var expectedHandle = "handle"
	var testParams = &WorkflowPropParams{
		Workflow: &Workflow{
			Nodes: []WorkflowNode{
				{
					Handle: expectedHandle,
					// Incomplete is about listing the node with props which can not be validated
					Props: map[string]string{
						expectedProp: "value",
					},
				},
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Internal: shared.VarSpecTracking{
			VarSpecs: []shared.VarSpec{
				PropSpec{
					Key:  expectedProp,
					Type: "int",
				},
			},
		},
	}

	actual := handleWorkflowPropIncomplete(testParams, testState)

	if actual != nil {
		assert.Contains(t, actual.Error(), expectedHandle)
		assert.Contains(t, actual.Error(), expectedProp)
	} else {
		assert.Fail(t, "expected error")
	}
}

func Test_handleWorkflowPropIncomplete_NoNodes(t *testing.T) {
	var testParams = &WorkflowPropParams{
		Workflow: &Workflow{},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.NoError(t, handleWorkflowPropIncomplete(testParams, testState))
}

func Test_handleWorkflowPropIncomplete_NoProps(t *testing.T) {
	var testParams = &WorkflowPropParams{
		Workflow: &Workflow{
			Nodes: []WorkflowNode{
				{
					Handle: "test",
				},
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.NoError(t, handleWorkflowPropIncomplete(testParams, testState))
}

func Test_handleWorkflowPropIncomplete_NoWorkflow(t *testing.T) {
	assert.ErrorIs(t, handleWorkflowPropIncomplete(&WorkflowPropParams{}, &task.State{}), ErrorWorkflowNotFound)
}

func Test_handleWorkflowPropPretend(t *testing.T) {
	var testParams = &WorkflowPropParams{
		Workflow: &Workflow{},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.NoError(t, handleWorkflowPropPretend(testParams, testState))
}

func Test_handleWorkflowPropPretend_NoNodes(t *testing.T) {
	var expectedHandle = "handle"
	var expectedProp = "PROP"
	var testParams = &WorkflowPropParams{
		Workflow: &Workflow{
			Nodes: []WorkflowNode{
				{
					Handle: expectedHandle,
					Props: map[string]string{
						expectedProp: "value",
					},
				},
			},
		},
	}
	var testLogger, loggerHook = test.NewNullLogger()
	var testState = &task.State{
		Logger: testLogger,
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.NoError(t, handleWorkflowPropPretend(testParams, testState))

	if actual := loggerHook.LastEntry(); actual != nil {
		assert.Contains(t, actual.Message, expectedProp)
		assert.Contains(t, actual.Message, expectedHandle)
	} else {
		assert.Fail(t, "no log entries")
	}
}

func Test_handleWorkflowPropPretend_NoProps(t *testing.T) {
	var testParams = &WorkflowPropParams{
		Workflow: &Workflow{
			Nodes: []WorkflowNode{
				{
					Handle: "test",
				},
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.NoError(t, handleWorkflowPropPretend(testParams, testState))
}

func Test_handleWorkflowPropPretend_NoWorkflow(t *testing.T) {
	assert.ErrorIs(t, handleWorkflowPropPretend(&WorkflowPropParams{}, &task.State{}), ErrorWorkflowNotFound)
}

func Test_handleWorkflowUpdateConfig(t *testing.T) {
	var expectedDescription = "description"
	var expectedName = "name"
	var initWorkflow = Workflow{
		Handle:      "testHandle",
		Description: "initDesc",
		Name:        "initName",
	}
	var testWorkflow = &Workflow{Handle: "testHandle"}
	var testParams = &WorkflowParams{Workflow: testWorkflow, WorkflowUpdate: true}
	var testState = &task.State{
		Output: "output.yaml",
		Logger: logrus.New(),
	}
	var mockWriter = &mockWorkflowWriter{
		testWorkflows: []Workflow{initWorkflow},
	}

	assert.NoError(t, handleWorkflowUpdateConfig(mockWriter, testParams, testState))
	assert.Equal(t, 1, len(mockWriter.resultWorkflows))
	assert.NotEqual(t, expectedDescription, mockWriter.resultWorkflows[0].Description)
	assert.NotEqual(t, expectedName, mockWriter.resultWorkflows[0].Name)
	testWorkflow.Description = expectedDescription
	assert.NoError(t, handleWorkflowUpdateConfig(mockWriter, testParams, testState))
	assert.Equal(t, 1, len(mockWriter.resultWorkflows))
	assert.Equal(t, expectedDescription, mockWriter.resultWorkflows[0].Description)
	assert.NotEqual(t, expectedName, mockWriter.resultWorkflows[0].Name)
	testWorkflow.Name = expectedName
	assert.NoError(t, handleWorkflowUpdateConfig(mockWriter, testParams, testState))
	assert.Equal(t, 1, len(mockWriter.resultWorkflows))
	assert.Equal(t, expectedDescription, mockWriter.resultWorkflows[0].Description)
	assert.Equal(t, expectedName, mockWriter.resultWorkflows[0].Name)
	mockWriter.testWorkflows = []Workflow{initWorkflow}
	testWorkflow.Description = ""
	assert.NoError(t, handleWorkflowUpdateConfig(mockWriter, testParams, testState))
	assert.Equal(t, 1, len(mockWriter.resultWorkflows))
	assert.NotEqual(t, expectedDescription, mockWriter.resultWorkflows[0].Description)
	assert.Equal(t, expectedName, mockWriter.resultWorkflows[0].Name)
	mockWriter.writeError = errors.New("expected")
	assert.ErrorIs(t, handleWorkflowUpdateConfig(mockWriter, testParams, testState), mockWriter.writeError)
}

func Test_handleWorkflowUpdateConfig_NewWorkflow(t *testing.T) {
	var testWorkflow = &Workflow{Handle: "testHandle"}
	var testParams = &WorkflowParams{Workflow: testWorkflow}
	var testState = &task.State{
		Output: "output.yaml",
		Logger: logrus.New(),
	}
	var mockWriter = &mockWorkflowWriter{
		testWorkflows: []Workflow{
			{
				Handle: "AnotherWorkflow",
			},
		},
	}

	assert.NoError(t, handleWorkflowUpdateConfig(mockWriter, testParams, testState))
	assert.Equal(t, 2, len(mockWriter.resultWorkflows))
	assert.Equal(t, testState.Output, mockWriter.testOutput)
}

func Test_handleWorkflowUpdateConfig_NoFileOutput(t *testing.T) {
	assert.ErrorIs(t, handleWorkflowUpdateConfig(nil, &WorkflowParams{}, &task.State{}), errorWorkflowFileInvalid)
}

func Test_handleWorkflowUpdateConfig_NoWorkflow(t *testing.T) {
	var testState = &task.State{
		Output: "output.yaml",
	}

	assert.Error(t, handleWorkflowUpdateConfig(nil, &WorkflowParams{}, testState))
}

func Test_handleWorkflowUpdateConfig_NoWorkflowUpdate(t *testing.T) {
	var testWorkflow = &Workflow{Handle: "testHandle"}
	var testParams = &WorkflowParams{Workflow: testWorkflow}
	var testState = &task.State{
		Output: "output.yaml",
		Logger: logrus.New(),
	}
	var mockWriter = &mockWorkflowWriter{
		testWorkflows: []Workflow{
			{
				Handle: "testHandle",
			},
		},
	}

	assert.ErrorIs(t, handleWorkflowUpdateConfig(mockWriter, testParams, testState), errorWorkflowConflict)
}

func Test_handleWorkflowUpdatePretend(t *testing.T) {
	var expectedDescription = "testDescription"
	var expectedName = "testName"
	var expectedHandle = "expectedHandle"
	var testWorkflow = &Workflow{
		Handle:      expectedHandle,
		Description: "notExpectedDescription",
		Name:        "expectedName",
		Links: []WorkflowLink{
			{
				LhsNode:     "expectLhsNode",
				LhsNodePort: "expectLhsNodePort",
				RhsNode:     "expectRhsNode",
				RhsNodePort: "expectRhsNodePort",
			},
		},
		Nodes: []WorkflowNode{
			{
				Description: "nodeDescription",
				Handle:      "nodeHandle",
				Name:        "nodeName",
				Sf: &WorkflowNodeFunction{
					Handle:  "sfHandle",
					Oem:     "sfOem",
					Version: "sfVersion",
					Seq:     37,
				},
			},
		},
	}
	var testParams = &WorkflowParams{WorkflowUpdate: true, Workflow: &Workflow{
		Handle:      expectedHandle,
		Description: expectedDescription,
		Name:        expectedName,
	}}
	var testState = &task.State{
		Error:  errors.New("exist"),
		Logger: logrus.New(),
		Output: "output.yaml",
	}
	var mockWriter = &mockWorkflowWriter{
		testWorkflows: []Workflow{*testWorkflow},
	}
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	assert.NoError(t, handleWorkflowUpdatePretend(mockWriter, testParams, testState))
	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.NotContains(t, output, testWorkflow.Name)
	assert.NotContains(t, output, testWorkflow.Description)
	assert.Contains(t, output, expectedDescription)
	assert.Contains(t, output, expectedName)
	assert.Contains(t, output, testWorkflow.Links[0].LhsNode)
	assert.Contains(t, output, testWorkflow.Links[0].LhsNodePort)
	assert.Contains(t, output, testWorkflow.Links[0].RhsNode)
	assert.Contains(t, output, testWorkflow.Links[0].RhsNodePort)
	assert.Contains(t, output, testWorkflow.Nodes[0].Description)
	assert.Contains(t, output, testWorkflow.Nodes[0].Handle)
	assert.Contains(t, output, testWorkflow.Nodes[0].Name)
	assert.Contains(t, output, testWorkflow.Nodes[0].Sf.Handle)
	assert.Contains(t, output, testWorkflow.Nodes[0].Sf.Oem)
	assert.Contains(t, output, testWorkflow.Nodes[0].Sf.Version)
	assert.Contains(t, output, cast.ToString(testWorkflow.Nodes[0].Sf.Seq))
}

func Test_handleWorkflowUpdatePretend_NewWorkflow(t *testing.T) {
	var expectedDescription = "testDescription"
	var expectedName = "testName"
	var expectedHandle = "expectedHandle"
	var testWorkflow = &Workflow{
		Handle:      expectedHandle,
		Description: "notExpectedDescription",
		Name:        "expectedName",
	}
	var testParams = &WorkflowParams{WorkflowUpdate: true, Workflow: &Workflow{
		Handle:      expectedHandle,
		Description: expectedDescription,
		Name:        expectedName,
		Links: []WorkflowLink{
			{
				LhsNode:     "expectLhsNode",
				LhsNodePort: "expectLhsNodePort",
				RhsNode:     "expectRhsNode",
				RhsNodePort: "expectRhsNodePort",
			},
		},
		Nodes: []WorkflowNode{
			{
				Description: "nodeDescription",
				Handle:      "nodeHandle",
				Name:        "nodeName",
				Sf: &WorkflowNodeFunction{
					Handle:  "sfHandle",
					Oem:     "sfOem",
					Version: "sfVersion",
					Seq:     37,
				},
			},
		},
	}}
	var testState = &task.State{
		Error:  errors.New("exist"),
		Logger: logrus.New(),
		Output: "output.yaml",
	}
	var mockWriter = &mockWorkflowWriter{
		testWorkflows: []Workflow{},
	}
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	assert.NoError(t, handleWorkflowUpdatePretend(mockWriter, testParams, testState))
	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.NotContains(t, output, testWorkflow.Name)
	assert.NotContains(t, output, testWorkflow.Description)
	assert.Contains(t, output, expectedDescription)
	assert.Contains(t, output, expectedName)
	assert.Contains(t, output, testParams.Links[0].LhsNode)
	assert.Contains(t, output, testParams.Links[0].LhsNodePort)
	assert.Contains(t, output, testParams.Links[0].RhsNode)
	assert.Contains(t, output, testParams.Links[0].RhsNodePort)
	assert.Contains(t, output, testParams.Nodes[0].Description)
	assert.Contains(t, output, testParams.Nodes[0].Handle)
	assert.Contains(t, output, testParams.Nodes[0].Name)
	assert.Contains(t, output, testParams.Nodes[0].Sf.Handle)
	assert.Contains(t, output, testParams.Nodes[0].Sf.Oem)
	assert.Contains(t, output, testParams.Nodes[0].Sf.Version)
	assert.Contains(t, output, cast.ToString(testParams.Nodes[0].Sf.Seq))
}

func Test_handleWorkflowUpdatePretend_NoFileOutput(t *testing.T) {
	assert.ErrorIs(t, handleWorkflowUpdatePretend(nil, &WorkflowParams{}, &task.State{}), errorWorkflowFileInvalid)
}

func Test_handleWorkflowUpdatePretend_NoUpdate(t *testing.T) {
	var testWorkflow = &Workflow{
		Handle:      "expectedHandle",
		Description: "expectedDescription",
		Name:        "expectedName",
	}
	var testParams = &WorkflowParams{Workflow: testWorkflow, WorkflowUpdate: false}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output.yaml",
	}
	var mockWriter = &mockWorkflowWriter{
		testWorkflows: []Workflow{*testWorkflow},
	}
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	assert.NoError(t, handleWorkflowUpdatePretend(mockWriter, testParams, testState))
	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Contains(t, output, testWorkflow.Name)
	assert.Contains(t, output, testWorkflow.Description)
}

func Test_handleWorkflowUpdatePretend_UpdateLinks(t *testing.T) {
	var expectedDescription = "testDescription"
	var expectedHandle = "expectedHandle"
	var testWorkflow = &Workflow{
		Handle:      expectedHandle,
		Description: "notExpectedDescription",
		Name:        "expectedName",
		Links: []WorkflowLink{
			{
				LhsNode:     "expectLhsNode",
				LhsNodePort: "expectLhsNodePort",
				RhsNode:     "expectRhsNode",
				RhsNodePort: "expectRhsNodePort",
			},
		},
	}
	var testParams = &WorkflowParams{WorkflowUpdate: true, Workflow: &Workflow{
		Handle:      expectedHandle,
		Description: expectedDescription,
	}}
	var testState = &task.State{
		Error:  errors.New("exist"),
		Logger: logrus.New(),
		Output: "output.yaml",
	}
	var mockWriter = &mockWorkflowWriter{
		testWorkflows: []Workflow{*testWorkflow},
	}
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	assert.NoError(t, handleWorkflowUpdatePretend(mockWriter, testParams, testState))
	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.NotContains(t, output, testWorkflow.Name)
	assert.NotContains(t, output, testWorkflow.Description)
	assert.Contains(t, output, expectedDescription)
	assert.Contains(t, output, testWorkflow.Links[0].LhsNode)
	assert.Contains(t, output, testWorkflow.Links[0].LhsNodePort)
	assert.Contains(t, output, testWorkflow.Links[0].RhsNode)
	assert.Contains(t, output, testWorkflow.Links[0].RhsNodePort)
}

func Test_handleWorkflowUpdatePretend_UpdateNodes(t *testing.T) {
	var expectedName = "testName"
	var expectedHandle = "expectedHandle"
	var testWorkflow = &Workflow{
		Handle:      expectedHandle,
		Description: "notExpectedDescription",
		Name:        "expectedName",
		Nodes: []WorkflowNode{
			{
				Description: "nodeDescription",
				Handle:      "nodeHandle",
				Name:        "nodeName",
			},
		},
	}
	var testParams = &WorkflowParams{WorkflowUpdate: true, Workflow: &Workflow{
		Handle: expectedHandle,
		Name:   expectedName,
	}}
	var testState = &task.State{
		Error:  errors.New("exist"),
		Logger: logrus.New(),
		Output: "output.yaml",
	}
	var mockWriter = &mockWorkflowWriter{
		testWorkflows: []Workflow{*testWorkflow},
	}
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	assert.NoError(t, handleWorkflowUpdatePretend(mockWriter, testParams, testState))
	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.NotContains(t, output, testWorkflow.Name)
	assert.NotContains(t, output, testWorkflow.Description)
	assert.Contains(t, output, expectedName)
	assert.Contains(t, output, testWorkflow.Nodes[0].Description)
	assert.Contains(t, output, testWorkflow.Nodes[0].Handle)
	assert.Contains(t, output, testWorkflow.Nodes[0].Name)
}
