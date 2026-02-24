package wf

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

func TestLinksExecutor_Add(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedLink = "link"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddLinksOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newLinksAddExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(testCmd, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Add(expectedWorkflow, []string{expectedLink})
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow:[\s\t]*`+expectedWorkflow), actual)
	assert.Regexp(t, regexp.MustCompile(`add-0:[\s\t]*`+expectedLink), actual)
}

func TestLinksExecutor_Display(t *testing.T) {
	var expectedWorkflow = "workflow"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddLinksOptions()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		workflowArg:  expectedWorkflow,
		LinksOptions: testOptions,
	}

	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow:[\s\t]*`+expectedWorkflow), actual)
}

func TestLinksExecutor_Init(t *testing.T) {
	var testDir = t.TempDir()
	var expectedLeftPath = "leftPath"
	var expectedLeftPort = "left_port"
	var expectedRightPath = "rightPath"
	var expectedRightPort = "right_port"
	var expectedWorkflow = "workflow"
	var expectedSfOem = "sfOem"
	var expectedSfVersion = "sfVersion"
	var expectedLeftSfHandle = "leftHandle"
	var expectedRightSfHandle = "rightHandle"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions: NewAddLinksOptions(),

		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedLeftPath,
							Sf: &broker.WorkflowNodeFunction{
								Oem:     expectedSfOem,
								Handle:  expectedLeftSfHandle,
								Version: expectedSfVersion,
							},
						},
						{
							Handle: expectedRightPath,
							Sf: &broker.WorkflowNodeFunction{
								Oem:     expectedSfOem,
								Handle:  expectedRightSfHandle,
								Version: expectedSfVersion,
							},
						},
					},
				},
			},
		}),
	}
	var fd *os.File
	var err error

	if err = os.MkdirAll(filepath.Join(testDir, expectedLeftPath), 0750); err == nil {
		var sfViper = viper.New()

		sfViper.Set(schema.Genaiz.Function.Publish.Oem.Doc, expectedSfOem)
		sfViper.Set(schema.Genaiz.Function.Publish.Handle.Doc, expectedLeftSfHandle)
		sfViper.Set(schema.Genaiz.Function.Publish.Version.Doc, expectedSfVersion)
		sfViper.Set(schema.Genaiz.Function.Publish.OutputPorts.Doc, []map[string]interface{}{
			{
				"handle": expectedLeftPort,
			},
		})
		err = sfViper.WriteConfigAs(filepath.Join(testDir, expectedLeftPath, "Genaiz.yaml"))
	}

	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	if err = os.MkdirAll(filepath.Join(testDir, expectedRightPath), 0750); err == nil {
		var sfViper = viper.New()

		sfViper.Set(schema.Genaiz.Function.Publish.Oem.Doc, expectedSfOem)
		sfViper.Set(schema.Genaiz.Function.Publish.Handle.Doc, expectedRightSfHandle)
		sfViper.Set(schema.Genaiz.Function.Publish.Version.Doc, expectedSfVersion)
		sfViper.Set(schema.Genaiz.Function.Publish.InputPorts.Doc, []map[string]interface{}{
			{
				"handle": expectedRightPort,
			},
		})
		err = sfViper.WriteConfigAs(filepath.Join(testDir, expectedRightPath, "Genaiz.yaml"))
	}

	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	if fd, err = os.Create(filepath.Join(testDir, "Genaiz.yaml")); err == nil {
		var solutionBytes []byte
		var solution = &broker.Solution{Workflows: []broker.Workflow{
			{
				Handle: expectedWorkflow,
			},
		}}

		defer filez.CloseSilently(fd)

		if solutionBytes, err = yaml.Marshal(solution); err == nil {
			if _, err = fd.Write(solutionBytes); err == nil {
				var testLink = fmt.Sprintf("%s/%s:%s/%s", expectedLeftPath, expectedLeftPort, expectedRightPath, expectedRightPort)
				var expectedLink = fmt.Sprintf("%s[%s]:%s[%s]", expectedLeftPath, expectedLeftPort, expectedRightPath, expectedRightPort)
				var actualLinks []string

				t.Chdir(testDir)
				actualLinks, err = testExecutor.Init(expectedWorkflow, []string{testLink})
				assert.NoError(t, err)
				assert.Equal(t, 1, len(actualLinks))
				assert.Equal(t, expectedLink, actualLinks[0])
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestLinksExecutor_Init_BracketLink(t *testing.T) {
	var testDir = t.TempDir()
	var expectedLeftHandle = "leftHandle"
	var expectedLeftPort = "leftPort"
	var expectedRightHandle = "rightHandle"
	var expectedRightPort = "rightPort"
	var expectedWorkflow = "workflow"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions: NewAddLinksOptions(),

		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedRightHandle,
						},
						{
							Handle: expectedLeftHandle,
						},
					},
				},
			},
		}),
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(filepath.Join(testDir, "Genaiz.yaml")); err == nil {
		var solutionBytes []byte
		var solution = &broker.Solution{Workflows: []broker.Workflow{
			{
				Handle: expectedWorkflow,
			},
		}}

		defer filez.CloseSilently(fd)

		if solutionBytes, err = yaml.Marshal(solution); err == nil {
			if _, err = fd.Write(solutionBytes); err == nil {
				var testLink = fmt.Sprintf("%s[%s]:%s[%s]", expectedLeftHandle, expectedLeftPort, expectedRightHandle, expectedRightPort)
				var expectedLink = fmt.Sprintf("%s[%s]:%s[%s]", expectedLeftHandle, expectedLeftPort, expectedRightHandle, expectedRightPort)
				var actualLinks []string

				t.Chdir(testDir)
				actualLinks, err = testExecutor.Init(expectedWorkflow, []string{testLink})
				assert.NoError(t, err)
				assert.Equal(t, 1, len(actualLinks))
				assert.Equal(t, expectedLink, actualLinks[0])
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestLinksExecutor_Init_InvalidConfigType(t *testing.T) {
	var expectedWorkflow = "workflow"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions: NewAddLinksOptions(),
	}

	actualLinks, err := testExecutor.Init(expectedWorkflow, []string{})
	assert.Empty(t, actualLinks)
	assert.Error(t, err)
}

func TestLinksExecutor_Init_InvalidWorkflow(t *testing.T) {
	var testDir = t.TempDir()
	var expectedWorkflow = "invalid_workflow"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions: NewAddLinksOptions(),

		workflowWriterFactory: newWorkflowWriterStub,
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(filepath.Join(testDir, "Genaiz.yaml")); err == nil {
		var actualLinks []string

		defer filez.CloseSilently(fd)
		t.Chdir(testDir)
		actualLinks, err = testExecutor.Init(expectedWorkflow, []string{})
		assert.Empty(t, actualLinks)
		assert.Error(t, err)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestLinksExecutor_Init_FindLeftInternalConflict(t *testing.T) {
	// Finding a Genaiz.yaml with the sf.publish key set to something other than a function map
	var testDir = t.TempDir()
	var expectedPath = "testPath"
	var expectedNodeHandle = "nodeHandle"
	var expectedRightHandle = "rightHandle"
	var expectedRightPort = "rightPort"
	var expectedWorkflow = "workflow"
	var expectedSfOem = "sfOem"
	var expectedSfHandle = "sfHandle"
	var expectedSfVersion = "sfVersion"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions: NewAddLinksOptions(),

		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedNodeHandle,
							Sf: &broker.WorkflowNodeFunction{
								Oem:     expectedSfOem,
								Handle:  expectedSfHandle,
								Version: expectedSfVersion,
							},
						},
						{
							Handle: expectedRightHandle,
						},
					},
				},
			},
		}),
	}
	var fd *os.File
	var err error

	if err = os.MkdirAll(filepath.Join(testDir, expectedPath), 0750); err == nil {
		var sfViper = viper.New()

		sfViper.Set(schema.Genaiz.Function.Publish.Internal.Doc, "notAFunctionMap")
		err = sfViper.WriteConfigAs(filepath.Join(testDir, expectedPath, "Genaiz.yaml"))
	}

	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	if fd, err = os.Create(filepath.Join(testDir, "Genaiz.yaml")); err == nil {
		var solutionBytes []byte
		var solution = &broker.Solution{Workflows: []broker.Workflow{
			{
				Handle: expectedWorkflow,
			},
		}}

		defer filez.CloseSilently(fd)

		if solutionBytes, err = yaml.Marshal(solution); err == nil {
			if _, err = fd.Write(solutionBytes); err == nil {
				var testLink = fmt.Sprintf("%s:%s[%s]", expectedPath, expectedRightHandle, expectedRightPort)
				var actualLinks []string

				t.Chdir(testDir)
				actualLinks, err = testExecutor.Init(expectedWorkflow, []string{testLink})
				assert.Error(t, err)
				assert.Empty(t, actualLinks)
				return
			}
		}
	}

	assert.NoError(t, err)

}

func TestLinksExecutor_Init_FindLeftNoValidation(t *testing.T) {
	var testDir = t.TempDir()
	var expectedPath = "testPath"
	var expectedRightHandle = "rightHandle"
	var expectedRightPort = "rightPort"
	var expectedWorkflow = "workflow"
	var expectedSfOem = "sfOem"
	var expectedSfHandle = "sfHandle"
	var expectedSfVersion = "sfVersion"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions: NewAddLinksOptions(),

		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedRightHandle,
						},
					},
				},
			},
		}),
	}
	var fd *os.File
	var err error

	if err = os.MkdirAll(filepath.Join(testDir, expectedPath), 0750); err == nil {
		var sfViper = viper.New()

		sfViper.Set(schema.Genaiz.Function.Publish.Oem.Doc, expectedSfOem)
		sfViper.Set(schema.Genaiz.Function.Publish.Handle.Doc, expectedSfHandle)
		sfViper.Set(schema.Genaiz.Function.Publish.Version.Doc, expectedSfVersion)
		sfViper.Set(schema.Genaiz.Function.Publish.OutputPorts.Doc, []map[string]interface{}{
			{
				"handle": expectedRightPort,
			},
		})
		err = sfViper.WriteConfigAs(filepath.Join(testDir, expectedPath, "Genaiz.yaml"))
	}

	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	if fd, err = os.Create(filepath.Join(testDir, "Genaiz.yaml")); err == nil {
		var solutionBytes []byte
		var solution = &broker.Solution{Workflows: []broker.Workflow{
			{
				Handle: expectedWorkflow,
			},
		}}

		defer filez.CloseSilently(fd)

		if solutionBytes, err = yaml.Marshal(solution); err == nil {
			if _, err = fd.Write(solutionBytes); err == nil {
				var testLink = fmt.Sprintf("%s:%s[%s]", expectedPath, expectedRightHandle, expectedRightPort)
				var expectedLink = fmt.Sprintf("%s:%s[%s]", expectedPath, expectedRightHandle, expectedRightPort)
				var actualLinks []string

				testViper.Set(schema.Genaiz.Workflow.Links.Add.NoValidation.Doc, true)
				t.Chdir(testDir)
				actualLinks, err = testExecutor.Init(expectedWorkflow, []string{testLink})
				assert.NoError(t, err)
				assert.Equal(t, 1, len(actualLinks))
				assert.Equal(t, expectedLink, actualLinks[0])
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestLinksExecutor_Init_FindLeftValue(t *testing.T) {
	var testDir = t.TempDir()
	var expectedPath = "testPath"
	var expectedNodeHandle = "nodeHandle"
	var expectedRightHandle = "rightHandle"
	var expectedRightPort = "rightPort"
	var expectedWorkflow = "workflow"
	var expectedSfOem = "sfOem"
	var expectedSfHandle = "sfHandle"
	var expectedSfVersion = "sfVersion"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions: NewAddLinksOptions(),

		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedNodeHandle,
							Sf: &broker.WorkflowNodeFunction{
								Oem:     expectedSfOem,
								Handle:  expectedSfHandle,
								Version: expectedSfVersion,
							},
						},
						{
							Handle: expectedRightHandle,
						},
					},
				},
			},
		}),
	}
	var fd *os.File
	var err error

	if err = os.MkdirAll(filepath.Join(testDir, expectedPath), 0750); err == nil {
		var sfViper = viper.New()

		sfViper.Set(schema.Genaiz.Function.Publish.Oem.Doc, expectedSfOem)
		sfViper.Set(schema.Genaiz.Function.Publish.Handle.Doc, expectedSfHandle)
		sfViper.Set(schema.Genaiz.Function.Publish.Version.Doc, expectedSfVersion)
		sfViper.Set(schema.Genaiz.Function.Publish.OutputPorts.Doc, []map[string]interface{}{
			{
				"handle": expectedRightPort,
			},
		})
		err = sfViper.WriteConfigAs(filepath.Join(testDir, expectedPath, "Genaiz.yaml"))
	}

	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	if fd, err = os.Create(filepath.Join(testDir, "Genaiz.yaml")); err == nil {
		var solutionBytes []byte
		var solution = &broker.Solution{Workflows: []broker.Workflow{
			{
				Handle: expectedWorkflow,
			},
		}}

		defer filez.CloseSilently(fd)

		if solutionBytes, err = yaml.Marshal(solution); err == nil {
			if _, err = fd.Write(solutionBytes); err == nil {
				var testLink = fmt.Sprintf("%s:%s[%s]", expectedPath, expectedRightHandle, expectedRightPort)
				var expectedLink = fmt.Sprintf("%s:%s[%s]", expectedNodeHandle, expectedRightHandle, expectedRightPort)
				var actualLinks []string

				t.Chdir(testDir)
				actualLinks, err = testExecutor.Init(expectedWorkflow, []string{testLink})
				assert.NoError(t, err)
				assert.Equal(t, 1, len(actualLinks))
				assert.Equal(t, expectedLink, actualLinks[0])
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestLinksExecutor_Init_FindLeftYamlError(t *testing.T) {
	var testDir = t.TempDir()
	var expectedPath = "testPath"
	var expectedNodeHandle = "nodeHandle"
	var expectedRightHandle = "rightHandle"
	var expectedRightPort = "rightPort"
	var expectedWorkflow = "workflow"
	var expectedSfOem = "sfOem"
	var expectedSfHandle = "sfHandle"
	var expectedSfVersion = "sfVersion"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions: NewAddLinksOptions(),

		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedNodeHandle,
							Sf: &broker.WorkflowNodeFunction{
								Oem:     expectedSfOem,
								Handle:  expectedSfHandle,
								Version: expectedSfVersion,
							},
						},
						{
							Handle: expectedRightHandle,
						},
					},
				},
			},
		}),
	}
	var fd *os.File
	var err error

	if err = os.MkdirAll(filepath.Join(testDir, expectedPath), 0750); err == nil {
		var cfd *os.File

		if cfd, err = os.Create(filepath.Join(testDir, expectedPath, "Genaiz.yaml")); err == nil {
			defer filez.CloseSilently(cfd)
			_, err = cfd.Write([]byte(":notYaml"))
		}
	}

	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	if fd, err = os.Create(filepath.Join(testDir, "Genaiz.yaml")); err == nil {
		var solutionBytes []byte
		var solution = &broker.Solution{Workflows: []broker.Workflow{
			{
				Handle: expectedWorkflow,
			},
		}}

		defer filez.CloseSilently(fd)

		if solutionBytes, err = yaml.Marshal(solution); err == nil {
			if _, err = fd.Write(solutionBytes); err == nil {
				var testLink = fmt.Sprintf("%s:%s[%s]", expectedPath, expectedRightHandle, expectedRightPort)
				var actualLinks []string

				t.Chdir(testDir)
				actualLinks, err = testExecutor.Init(expectedWorkflow, []string{testLink})
				assert.Error(t, err)
				assert.Empty(t, actualLinks)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestLinksExecutor_Init_FindRightByHandle(t *testing.T) {
	var testDir = t.TempDir()
	var expectedPath = "testPath"
	var expectedLeftHandle = "leftHandle"
	var expectedRightHandle = "rightHandle"
	var expectedRightPort = "rightPort"
	var expectedWorkflow = "workflow"
	var expectedSfOem = "sfOem"
	var expectedSfVersion = "sfVersion"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions: NewAddLinksOptions(),

		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedRightHandle,
							Sf: &broker.WorkflowNodeFunction{
								Oem:     expectedSfOem,
								Handle:  expectedRightHandle,
								Version: expectedSfVersion,
							},
						},
						{
							Handle: expectedLeftHandle,
						},
					},
				},
			},
		}),
	}
	var fd *os.File
	var err error

	if err = os.MkdirAll(filepath.Join(testDir, expectedPath), 0750); err == nil {
		var sfViper = viper.New()

		sfViper.Set(schema.Genaiz.Function.Publish.Oem.Doc, expectedSfOem)
		sfViper.Set(schema.Genaiz.Function.Publish.Handle.Doc, expectedRightHandle)
		sfViper.Set(schema.Genaiz.Function.Publish.Version.Doc, expectedSfVersion)
		sfViper.Set(schema.Genaiz.Function.Publish.InputPorts.Doc, []broker.DataPort{
			{
				Handle: expectedRightPort,
			},
		})
		err = sfViper.WriteConfigAs(filepath.Join(testDir, expectedPath, "Genaiz.yaml"))
	}

	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	if fd, err = os.Create(filepath.Join(testDir, "Genaiz.yaml")); err == nil {
		var solutionBytes []byte
		var solution = &broker.Solution{Workflows: []broker.Workflow{
			{
				Handle: expectedWorkflow,
			},
		}}

		defer filez.CloseSilently(fd)

		if solutionBytes, err = yaml.Marshal(solution); err == nil {
			if _, err = fd.Write(solutionBytes); err == nil {
				var expectedLink = fmt.Sprintf("%s:%s[%s]", expectedLeftHandle, expectedRightHandle, expectedRightPort)
				var actualLinks []string

				t.Chdir(testDir)
				actualLinks, err = testExecutor.Init(expectedWorkflow, []string{expectedLink})
				assert.NoError(t, err)
				assert.Equal(t, 1, len(actualLinks))
				assert.Equal(t, expectedLink, actualLinks[0])
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestLinksExecutor_Init_FindRightPortError(t *testing.T) {
	var testDir = t.TempDir()
	var expectedPath = "testPath"
	var expectedNodeHandle = "nodeHandle"
	var expectedLeftHandle = "leftHandle"
	var expectedLeftPort = "leftPort"
	var expectedWorkflow = "workflow"
	var expectedSfOem = "sfOem"
	var expectedSfHandle = "sfHandle"
	var expectedSfVersion = "sfVersion"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions: NewAddLinksOptions(),

		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedNodeHandle,
							Sf: &broker.WorkflowNodeFunction{
								Oem:     expectedSfOem,
								Handle:  expectedSfHandle,
								Version: expectedSfVersion,
							},
						},
						{
							Handle: expectedLeftHandle,
						},
					},
				},
			},
		}),
	}
	var fd *os.File
	var err error

	if err = os.MkdirAll(filepath.Join(testDir, expectedPath), 0750); err == nil {
		var sfViper = viper.New()

		sfViper.Set(schema.Genaiz.Function.Publish.Oem.Doc, expectedSfOem)
		sfViper.Set(schema.Genaiz.Function.Publish.Handle.Doc, expectedSfHandle)
		sfViper.Set(schema.Genaiz.Function.Publish.Version.Doc, expectedSfVersion)
		err = sfViper.WriteConfigAs(filepath.Join(testDir, expectedPath, "Genaiz.yaml"))
	}

	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	if fd, err = os.Create(filepath.Join(testDir, "Genaiz.yaml")); err == nil {
		var solutionBytes []byte
		var solution = &broker.Solution{Workflows: []broker.Workflow{
			{
				Handle: expectedWorkflow,
			},
		}}

		defer filez.CloseSilently(fd)

		if solutionBytes, err = yaml.Marshal(solution); err == nil {
			if _, err = fd.Write(solutionBytes); err == nil {
				var testLink = fmt.Sprintf("%s[%s]:%s/%s", expectedLeftHandle, expectedLeftPort, expectedPath, "noAPort")
				var actualLinks []string

				t.Chdir(testDir)
				actualLinks, err = testExecutor.Init(expectedWorkflow, []string{testLink})
				assert.ErrorIs(t, err, broker.ErrorDataPortNotFound)
				assert.Empty(t, actualLinks)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestLinksExecutor_Init_FindRightValue(t *testing.T) {
	var testDir = t.TempDir()
	var expectedPath = "testPath"
	var expectedNodeHandle = "nodeHandle"
	var expectedRightHandle = "rightHandle"
	var expectedRightPort = "rightPort"
	var expectedWorkflow = "workflow"
	var expectedSfOem = "sfOem"
	var expectedSfHandle = "sfHandle"
	var expectedSfVersion = "sfVersion"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions: NewAddLinksOptions(),

		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedNodeHandle,
							Sf: &broker.WorkflowNodeFunction{
								Oem:     expectedSfOem,
								Handle:  expectedSfHandle,
								Version: expectedSfVersion,
							},
						},
						{
							Handle: expectedRightHandle,
						},
					},
				},
			},
		}),
	}
	var fd *os.File
	var err error

	if err = os.MkdirAll(filepath.Join(testDir, expectedPath), 0750); err == nil {
		var sfViper = viper.New()

		sfViper.Set(schema.Genaiz.Function.Publish.Oem.Doc, expectedSfOem)
		sfViper.Set(schema.Genaiz.Function.Publish.Handle.Doc, expectedSfHandle)
		sfViper.Set(schema.Genaiz.Function.Publish.Version.Doc, expectedSfVersion)
		err = sfViper.WriteConfigAs(filepath.Join(testDir, expectedPath, "Genaiz.yaml"))
	}

	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	if fd, err = os.Create(filepath.Join(testDir, "Genaiz.yaml")); err == nil {
		var solutionBytes []byte
		var solution = &broker.Solution{Workflows: []broker.Workflow{
			{
				Handle: expectedWorkflow,
			},
		}}

		defer filez.CloseSilently(fd)

		if solutionBytes, err = yaml.Marshal(solution); err == nil {
			if _, err = fd.Write(solutionBytes); err == nil {
				var testLink = fmt.Sprintf("%s:%s[%s]", expectedPath, expectedRightHandle, expectedRightPort)
				var expectedLink = fmt.Sprintf("%s:%s[%s]", expectedNodeHandle, expectedRightHandle, expectedRightPort)
				var actualLinks []string

				t.Chdir(testDir)
				actualLinks, err = testExecutor.Init(expectedWorkflow, []string{testLink})
				assert.NoError(t, err)
				assert.Equal(t, 1, len(actualLinks))
				assert.Equal(t, expectedLink, actualLinks[0])
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestLinksExecutor_Init_InvalidRightValue(t *testing.T) {
	var testDir = t.TempDir()
	var expectedPath = "testPath"
	var expectedNodeHandle = "nodeHandle"
	var expectedLeftHandle = "leftHandle"
	var expectedLeftPort = "leftPort"
	var expectedWorkflow = "workflow"
	var expectedSfOem = "sfOem"
	var expectedSfHandle = "sfHandle"
	var expectedSfVersion = "sfVersion"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions: NewAddLinksOptions(),

		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedNodeHandle,
							Sf: &broker.WorkflowNodeFunction{
								Oem:     expectedSfOem,
								Handle:  expectedSfHandle,
								Version: expectedSfVersion,
							},
						},
						{
							Handle: expectedLeftHandle,
						},
					},
				},
			},
		}),
	}
	var fd *os.File
	var err error

	if err = os.MkdirAll(filepath.Join(testDir, expectedPath), 0750); err == nil {
		var sfViper = viper.New()

		sfViper.Set(schema.Genaiz.Function.Publish.Oem.Doc, expectedSfOem)
		sfViper.Set(schema.Genaiz.Function.Publish.Handle.Doc, expectedSfHandle)
		sfViper.Set(schema.Genaiz.Function.Publish.Version.Doc, expectedSfVersion)
		err = sfViper.WriteConfigAs(filepath.Join(testDir, expectedPath, "Genaiz.yaml"))
	}

	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	if fd, err = os.Create(filepath.Join(testDir, "Genaiz.yaml")); err == nil {
		var solutionBytes []byte
		var solution = &broker.Solution{Workflows: []broker.Workflow{
			{
				Handle: expectedWorkflow,
			},
		}}

		defer filez.CloseSilently(fd)

		if solutionBytes, err = yaml.Marshal(solution); err == nil {
			if _, err = fd.Write(solutionBytes); err == nil {
				var testLink = fmt.Sprintf("%s[%s]:%s", expectedLeftHandle, expectedLeftPort, expectedPath)
				var actualLinks []string

				t.Chdir(testDir)
				actualLinks, err = testExecutor.Init(expectedWorkflow, []string{testLink})
				assert.Error(t, err)
				assert.Empty(t, actualLinks)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestLinksExecutor_Init_ParseLeftConflict(t *testing.T) {
	var testDir = t.TempDir()
	var expectedWorkflow = "workflow"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions: NewAddLinksOptions(),

		workflowWriterFactory: newWorkflowWriterStub,
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(filepath.Join(testDir, "Genaiz.yaml")); err == nil {
		var solutionBytes []byte
		var solution = &broker.Solution{Workflows: []broker.Workflow{
			{
				Handle: expectedWorkflow,
			},
		}}

		defer filez.CloseSilently(fd)

		if solutionBytes, err = yaml.Marshal(solution); err == nil {
			if _, err = fd.Write(solutionBytes); err == nil {
				var actualLinks []string

				t.Chdir(testDir)
				actualLinks, err = testExecutor.Init(expectedWorkflow, []string{"testPath/testPort[anotherPort]:testHandle"})
				assert.Empty(t, actualLinks)
				assert.Error(t, err)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestLinksExecutor_Init_ParseRightConflict(t *testing.T) {
	var testDir = t.TempDir()
	var expectedWorkflow = "workflow"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions: NewAddLinksOptions(),

		workflowWriterFactory: newWorkflowWriterStub,
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(filepath.Join(testDir, "Genaiz.yaml")); err == nil {
		var solutionBytes []byte
		var solution = &broker.Solution{Workflows: []broker.Workflow{
			{
				Handle: expectedWorkflow,
			},
		}}

		defer filez.CloseSilently(fd)

		if solutionBytes, err = yaml.Marshal(solution); err == nil {
			if _, err = fd.Write(solutionBytes); err == nil {
				var actualLinks []string

				t.Chdir(testDir)
				actualLinks, err = testExecutor.Init(expectedWorkflow, []string{"testHandle:testPath/testPort[anotherPort]"})
				assert.Empty(t, actualLinks)
				assert.Error(t, err)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestLinksExecutor_Pretend(t *testing.T) {
	var calledWorkflow bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewAddLinksOptions()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions:          testOptions,
		workflowArg:           "workflow",
		workflowTaskFactory:   newWorkflowTaskPretendStub(&calledWorkflow),
		workflowWriterFactory: newWorkflowWriterStub,
	}

	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Pretend()
	assert.True(t, calledWorkflow)
}

func TestLinksExecutor_PretendInvalidConfigType(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewAddLinksOptions()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions:        testOptions,
		workflowTaskFactory: newWorkflowTaskPretendStub(&calledWorkflow),
	}

	defer patch.Unpatch()
	testViper.Set(testOptions.optionConfigType.Key, "invalid")
	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testExecutor.Pretend()
	assert.False(t, calledWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestLinksExecutor_PretendInvalidParams(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewAddLinksOptions()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions:          testOptions,
		workflowTaskFactory:   newWorkflowTaskPretendStub(&calledWorkflow),
		workflowWriterFactory: newWorkflowWriterStub,
	}

	defer patch.Unpatch()
	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Pretend()
	assert.False(t, calledWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestLinksExecutor_Proceed(t *testing.T) {
	var calledWorkflow bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewAddLinksOptions()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions:          testOptions,
		workflowArg:           "workflow",
		workflowTaskFactory:   newWorkflowTaskCompleteStub(&calledWorkflow),
		workflowWriterFactory: newWorkflowWriterStub,
	}

	testLedger.InitLogging()
	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Proceed()
	assert.True(t, calledWorkflow)
}

func TestLinksExecutor_ProceedInvalidConfigType(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewAddLinksOptions()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions:        testOptions,
		workflowTaskFactory: newWorkflowTaskCompleteStub(&calledWorkflow),
	}

	defer patch.Unpatch()
	testViper.Set(testOptions.optionConfigType.Key, "invalid")
	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testExecutor.Proceed()
	assert.False(t, calledWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestLinksExecutor_ProceedInvalidParams(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewAddLinksOptions()
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions:          testOptions,
		workflowTaskFactory:   newWorkflowTaskCompleteStub(&calledWorkflow),
		workflowWriterFactory: newWorkflowWriterStub,
	}

	defer patch.Unpatch()
	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Proceed()
	assert.False(t, calledWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestLinksExecutor_Remove(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedLink = "link"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddLinksOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newLinksRemoveExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Remove(expectedWorkflow, []string{expectedLink})
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow:[\s\t]*`+expectedWorkflow), actual)
	assert.Regexp(t, regexp.MustCompile(`rm-0:[\s\t]*`+expectedLink), actual)
}

func TestNewLinks(t *testing.T) {
	var testLedger = config.NewLedger()
	var testCmd = NewLinks(testLedger, &Cli{})

	assert.Equal(t, 2, len(testCmd.Commands()))
}

func Test_parseArgsLinks(t *testing.T) {
	var expectedInvalid = "invalid"
	var expectedLeft = "left"
	var expectedPort = "port"
	var expectedRight = "right"
	var testLinks = []string{
		expectedInvalid + "[nil]",
		fmt.Sprintf("%s[%s]:%s", expectedLeft, expectedPort, expectedRight),
	}
	var actualLinks = parseArgsLinks(testLinks...)

	assert.Equal(t, 1, len(actualLinks))
	assert.Equal(t, expectedLeft, actualLinks[0].LhsNode)
	assert.Equal(t, expectedPort, actualLinks[0].LhsNodePort)
	assert.Equal(t, expectedRight, actualLinks[0].RhsNode)
}

func Test_parseNodeRefs(t *testing.T) {
	var expectedNode = "normal"
	var expectedPort = "port"
	var expected = []string{expectedNode, expectedPort}
	var actualNode, actualPort string
	var err error

	// Not a valid case, but parsing does not validate beyond conflicting field specs
	actualNode, actualPort, err = parseNodeRefs("[port]", "")
	assert.NoError(t, err)
	assert.Equal(t, []string{"[port]", ""}, []string{actualNode, actualPort})
	actualNode, actualPort, err = parseNodeRefs("normal", "port")
	assert.NoError(t, err)
	assert.Equal(t, expected, []string{actualNode, actualPort})
	actualNode, actualPort, err = parseNodeRefs("normal/", "port")
	assert.NoError(t, err)
	assert.Equal(t, expected, []string{actualNode, actualPort})
	actualNode, actualPort, err = parseNodeRefs("normal/port", "")
	assert.NoError(t, err)
	assert.Equal(t, expected, []string{actualNode, actualPort})
	actualNode, actualPort, err = parseNodeRefs("normal/port/", "")
	assert.NoError(t, err)
	assert.Equal(t, expected, []string{actualNode, actualPort})
	actualNode, actualPort, err = parseNodeRefs("normal/child/port", "")
	assert.NoError(t, err)
	assert.Equal(t, expected, []string{actualNode, actualPort})
	actualNode, actualPort, err = parseNodeRefs("normal/child/port/", "")
	assert.NoError(t, err)
	assert.Equal(t, expected, []string{actualNode, actualPort})
	actualNode, actualPort, err = parseNodeRefs("/normal//port/", "")
	assert.NoError(t, err)
	assert.Equal(t, expected, []string{actualNode, actualPort})
	actualNode, actualPort, err = parseNodeRefs("normal/port", "port2")
	assert.Error(t, err)
	assert.Empty(t, actualNode)
	assert.Empty(t, actualPort)
}

func Test_validateArgsLinks(t *testing.T) {
	assert.Error(t, validateArgLinks([]string{"valid:valid", "notValid[]"}))
	assert.NoError(t, validateArgLinks([]string{"valid1:valid2[port]"}))
}

func newWorkflowWriterStub(*config.Ledger, string) *workflowWriter {
	var stub = &workflowWriter{
		WorkflowWriter: config.NewWorkflowWriter(),
	}

	stub.WithCurrent(&broker.Solution{
		Workflows: []broker.Workflow{
			{
				Handle: "workflow",
			},
		},
	})
	return stub
}
