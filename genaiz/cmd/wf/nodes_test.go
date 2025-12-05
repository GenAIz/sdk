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

	"genaiz.com/genaiz-lib/lang/errorz"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

func TestNodesExecutor_Add(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedNode = "node"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddNodesOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newNodesAddExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(testCmd, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	assert.NoError(t, testExecutor.Add(expectedWorkflow, expectedNode))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow:[\s\t]*`+expectedWorkflow), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.description:[\s\t]*`), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.handle:[\s\t]*`+expectedNode), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.name:[\s\t]*`+expectedNode), actual)
}

func TestNodesExecutor_AddOnlyOem(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedNode = "node"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddNodesOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newNodesAddExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(testCmd, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testViper.Set(testOptions.optionSfOem.Key, "oem")
	assert.ErrorIs(t, testExecutor.Add(expectedWorkflow, expectedNode), errorIncompleteSfSpec)
}

func TestNodesExecutor_AddOnlyHandle(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedNode = "node"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddNodesOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newNodesAddExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(testCmd, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testViper.Set(testOptions.optionSfHandle.Key, "handle")
	assert.ErrorIs(t, testExecutor.Add(expectedWorkflow, expectedNode), errorIncompleteSfSpec)
}

func TestNodesExecutor_AddOnlyVersion(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedNode = "node"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddNodesOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newNodesAddExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(testCmd, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testViper.Set(testOptions.optionSfVersion.Key, "0.0.1")
	assert.ErrorIs(t, testExecutor.Add(expectedWorkflow, expectedNode), errorIncompleteSfSpec)
}

func TestNodesExecutor_AddOnlyHandleVersion(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedNode = "node"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddNodesOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newNodesAddExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(testCmd, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testViper.Set(testOptions.optionSfHandle.Key, "handle")
	testViper.Set(testOptions.optionSfVersion.Key, "0.0.1")
	assert.ErrorIs(t, testExecutor.Add(expectedWorkflow, expectedNode), errorIncompleteSfSpec)
}

func TestNodesExecutor_AddOnlyOemHandle(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedNode = "node"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddNodesOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newNodesAddExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(testCmd, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testViper.Set(testOptions.optionSfOem.Key, "oem")
	testViper.Set(testOptions.optionSfHandle.Key, "handle")
	assert.ErrorIs(t, testExecutor.Add(expectedWorkflow, expectedNode), errorIncompleteSfSpec)
}

func TestNodesExecutor_AddOnlyOemVersion(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedNode = "node"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddNodesOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newNodesAddExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(testCmd, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testViper.Set(testOptions.optionSfOem.Key, "oem")
	testViper.Set(testOptions.optionSfVersion.Key, "0.0.1")
	assert.ErrorIs(t, testExecutor.Add(expectedWorkflow, expectedNode), errorIncompleteSfSpec)
}

func TestNodesExecutor_AddWithSf(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedNode = "node"
	var expectedSfHandle = "handle"
	var expectedSfOem = "genaiz.com"
	var expectedSfVersion = "0.0.1"
	var expectedSfRc = "4"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddNodesOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newNodesAddExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(testCmd, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testViper.Set(testOptions.optionSfSerialized.Key, fmt.Sprintf("%s/%s:%s-rc%s", expectedSfOem, expectedSfHandle, expectedSfVersion, expectedSfRc))
	assert.NoError(t, testExecutor.Add(expectedWorkflow, expectedNode))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow:[\s\t]*`+expectedWorkflow), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.description:[\s\t]*`), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.handle:[\s\t]*`+expectedNode), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.name:[\s\t]*`+expectedNode), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.sf.handle:[\s\t]*`+expectedSfHandle), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.sf.oem:[\s\t]*`+expectedSfOem), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.sf.version:[\s\t]*`+expectedSfVersion), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.sf.seq:[\s\t]*`+expectedSfRc), actual)
}

func TestNodesExecutor_Display(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedNode = "node"
	var expectedSfHandle = "_sf_handle"
	var expectedSfOem = "_sf_oem"
	var expectedSfVersion = "_sf_version"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddNodesOptions()
	var testCmd = &cobra.Command{}
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		NodesOptions: testOptions,
		addNode: &broker.WorkflowNode{
			Handle: expectedNode,
			Sf: &broker.WorkflowNodeFunction{
				Handle:  expectedSfHandle,
				Oem:     expectedSfOem,
				Version: expectedSfVersion,
				Seq:     0,
			},
		},
		workflowArg: expectedWorkflow,
	}

	testLedger.Register(testCmd, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow:[\s\t]*`+expectedWorkflow), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.description:[\s\t]*`), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.handle:[\s\t]*`+expectedNode), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.name:[\s\t]*`), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.sf.handle:[\s\t]*`+expectedSfHandle), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.sf.oem:[\s\t]*`+expectedSfOem), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.sf.version:[\s\t]*`+expectedSfVersion), actual)
}

func TestNodesExecutor_Find(t *testing.T) {
	var testDir = t.TempDir()
	var testPath = filepath.Join(testDir, "Genaiz.yaml")
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testPath); err == nil {
		var expectedHandle = "test-handle"
		var publishMap = map[string]any{
			"sf": map[string]any{
				"publish": map[string]any{
					"handle": expectedHandle,
				},
			},
		}
		var actualHandle string
		var testBytes []byte

		if testBytes, err = yaml.Marshal(publishMap); err == nil {
			if _, err = fd.Write(testBytes); err == nil {
				defer filez.CloseSilently(fd)
				actualHandle, err = testExecutor.Find(testDir)
				assert.NoError(t, err)
				assert.Equal(t, expectedHandle+"-node", actualHandle)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestNodesExecutor_Find_HandleNotFound(t *testing.T) {
	var testDir = t.TempDir()
	var testPath = filepath.Join(testDir, "Genaiz.yaml")
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testPath); err == nil {
		var publishMap = map[string]any{
			"sf": map[string]any{
				"publish": map[string]any{
					"oem": "test-oem",
				},
			},
		}
		var actualHandle string
		var testBytes []byte

		if testBytes, err = yaml.Marshal(publishMap); err == nil {
			if _, err = fd.Write(testBytes); err == nil {
				defer filez.CloseSilently(fd)
				actualHandle, err = testExecutor.Find(testDir)
				assert.NoError(t, err)
				assert.Equal(t, testDir, actualHandle)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestNodesExecutor_Find_ParseError(t *testing.T) {
	var testDir = t.TempDir()
	var testPath = filepath.Join(testDir, "Genaiz.yaml")
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testPath); err == nil {
		var actualHandle string

		if _, err = fd.Write([]byte(":notYaml")); err == nil {
			defer filez.CloseSilently(fd)
			actualHandle, err = testExecutor.Find(testDir)
			assert.Error(t, err)
			assert.False(t, errorz.IsPathError(err))
			assert.Empty(t, actualHandle)
			return
		}
	}

	assert.NoError(t, err)
}

func TestNodesExecutor_Find_PathError(t *testing.T) {
	var testPath = filepath.Join(t.TempDir(), "notExisting")
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var actualHandle, err = testExecutor.Find(testPath)

	assert.NoError(t, err)
	assert.Equal(t, testPath, actualHandle)
}

func TestNodesExecutor_Init(t *testing.T) {
	var testDir = t.TempDir()
	var testPath = filepath.Join(testDir, "Genaiz.yaml")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
		NodesOptions: NewAddNodesOptions(),
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testPath); err == nil {
		var publishMap = map[string]any{
			"sf": map[string]any{
				"publish": map[string]any{
					"oem":     "oem",
					"handle":  "test-handle",
					"version": "1.1.0",
				},
			},
		}
		var actualHandle string
		var testBytes []byte

		defer filez.CloseSilently(fd)

		if testBytes, err = yaml.Marshal(publishMap); err == nil {
			if _, err = fd.Write(testBytes); err == nil {
				actualHandle, err = testExecutor.Init(testDir)
				assert.NoError(t, err)
				assert.Equal(t, filepath.Base(filepath.Dir(testPath))+"-node", actualHandle)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestNodesExecutor_Init_ConflictHandle(t *testing.T) {
	var testDir = t.TempDir()
	var testPath = filepath.Join(testDir, "Genaiz.yaml")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
		NodesOptions: NewAddNodesOptions(),
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testPath); err == nil {
		var publishMap = map[string]any{
			"sf": map[string]any{
				"publish": map[string]any{
					"oem":     "test-oem",
					"handle":  "test-handle",
					"version": "1.1.0",
				},
			},
		}
		var actualHandle string
		var testBytes []byte

		defer filez.CloseSilently(fd)

		if testBytes, err = yaml.Marshal(publishMap); err == nil {
			if _, err = fd.Write(testBytes); err == nil {
				testViper.Set(schema.Genaiz.Workflow.Nodes.Add.Oem.Doc, "test-oem")
				testViper.Set(schema.Genaiz.Workflow.Nodes.Add.Handle.Doc, "different-handle")
				actualHandle, err = testExecutor.Init(testDir)
				assert.Error(t, err)
				assert.Empty(t, actualHandle)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestNodesExecutor_Init_ConflictOem(t *testing.T) {
	var testDir = t.TempDir()
	var testPath = filepath.Join(testDir, "Genaiz.yaml")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
		NodesOptions: NewAddNodesOptions(),
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testPath); err == nil {
		var publishMap = map[string]any{
			"sf": map[string]any{
				"publish": map[string]any{
					"oem":     "test-oem",
					"handle":  "test-handle",
					"version": "1.1.0",
				},
			},
		}
		var actualHandle string
		var testBytes []byte

		defer filez.CloseSilently(fd)

		if testBytes, err = yaml.Marshal(publishMap); err == nil {
			if _, err = fd.Write(testBytes); err == nil {
				testViper.Set(schema.Genaiz.Workflow.Nodes.Add.Oem.Doc, "different-oem")
				actualHandle, err = testExecutor.Init(testDir)
				assert.Error(t, err)
				assert.Empty(t, actualHandle)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestNodesExecutor_Init_ConflictVersion(t *testing.T) {
	var testDir = t.TempDir()
	var testPath = filepath.Join(testDir, "Genaiz.yaml")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
		NodesOptions: NewAddNodesOptions(),
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testPath); err == nil {
		var publishMap = map[string]any{
			"sf": map[string]any{
				"publish": map[string]any{
					"oem":     "oem",
					"handle":  "test-handle",
					"version": "1.1.0",
				},
			},
		}
		var actualHandle string
		var testBytes []byte

		defer filez.CloseSilently(fd)

		if testBytes, err = yaml.Marshal(publishMap); err == nil {
			if _, err = fd.Write(testBytes); err == nil {
				testViper.Set(schema.Genaiz.Workflow.Nodes.Add.Handle.Doc, "test-handle")
				testViper.Set(schema.Genaiz.Workflow.Nodes.Add.Version.Doc, "1.2.0")
				actualHandle, err = testExecutor.Init(testDir)
				assert.Error(t, err)
				assert.Empty(t, actualHandle)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestNodesExecutor_Init_InvalidHandle(t *testing.T) {
	var testDir = t.TempDir()
	var testPath = filepath.Join(testDir, "Genaiz.yaml")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
		NodesOptions: NewAddNodesOptions(),
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testPath); err == nil {
		var publishMap = map[string]any{
			"sf": map[string]any{
				"publish": map[string]any{
					"oem":     "oem",
					"version": "1.1.0",
				},
			},
		}
		var actualHandle string
		var testBytes []byte

		defer filez.CloseSilently(fd)

		if testBytes, err = yaml.Marshal(publishMap); err == nil {
			if _, err = fd.Write(testBytes); err == nil {
				testViper.Set(schema.Genaiz.Workflow.Nodes.Add.Handle.Doc, "test-handle")
				testViper.Set(schema.Genaiz.Workflow.Nodes.Add.Version.Doc, "1.2.0")
				actualHandle, err = testExecutor.Init(testDir)
				assert.Error(t, err)
				assert.Empty(t, actualHandle)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestNodesExecutor_Init_InvalidOem(t *testing.T) {
	var testDir = t.TempDir()
	var testPath = filepath.Join(testDir, "Genaiz.yaml")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
		NodesOptions: NewAddNodesOptions(),
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testPath); err == nil {
		var publishMap = map[string]any{
			"sf": map[string]any{
				"publish": map[string]any{
					"handle":  "test-handle",
					"version": "1.1.0",
				},
			},
		}
		var actualHandle string
		var testBytes []byte

		defer filez.CloseSilently(fd)

		if testBytes, err = yaml.Marshal(publishMap); err == nil {
			if _, err = fd.Write(testBytes); err == nil {
				testViper.Set(schema.Genaiz.Workflow.Nodes.Add.Handle.Doc, "test-handle")
				testViper.Set(schema.Genaiz.Workflow.Nodes.Add.Version.Doc, "1.2.0")
				actualHandle, err = testExecutor.Init(testDir)
				assert.Error(t, err)
				assert.Empty(t, actualHandle)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestNodesExecutor_Init_InvalidVersion(t *testing.T) {
	var testDir = t.TempDir()
	var testPath = filepath.Join(testDir, "Genaiz.yaml")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
		NodesOptions: NewAddNodesOptions(),
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testPath); err == nil {
		var publishMap = map[string]any{
			"sf": map[string]any{
				"publish": map[string]any{
					"oem":    "oem",
					"handle": "test-handle",
				},
			},
		}
		var actualHandle string
		var testBytes []byte

		defer filez.CloseSilently(fd)

		if testBytes, err = yaml.Marshal(publishMap); err == nil {
			if _, err = fd.Write(testBytes); err == nil {
				testViper.Set(schema.Genaiz.Workflow.Nodes.Add.Handle.Doc, "test-handle")
				testViper.Set(schema.Genaiz.Workflow.Nodes.Add.Version.Doc, "1.2.0")
				actualHandle, err = testExecutor.Init(testDir)
				assert.Error(t, err)
				assert.Empty(t, actualHandle)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestNodesExecutor_Init_ParseError(t *testing.T) {
	var testDir = t.TempDir()
	var testPath = filepath.Join(testDir, "Genaiz.yaml")
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testPath); err == nil {
		var actualHandle string

		if _, err = fd.Write([]byte(":notYaml")); err == nil {
			defer filez.CloseSilently(fd)
			actualHandle, err = testExecutor.Init(testDir)
			assert.Error(t, err)
			assert.False(t, errorz.IsPathError(err))
			assert.Empty(t, actualHandle)
			return
		}
	}

	assert.NoError(t, err)
}

func TestNodesExecutor_Init_PathError(t *testing.T) {
	var testPath = filepath.Join(t.TempDir(), "notExisting")
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var actualHandle, err = testExecutor.Init(testPath)

	assert.NoError(t, err)
	assert.Equal(t, testPath, actualHandle)
}

func TestNodesExecutor_Pretend(t *testing.T) {
	var calledWorkflow bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewAddNodesOptions()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		NodesOptions: testOptions,
		workflowArg:  "workflow",
		addNode: &broker.WorkflowNode{
			Handle: "handle",
		},
		workflowTaskFactory:   newWorkflowTaskPretendStub(&calledWorkflow),
		workflowWriterFactory: newWorkflowWriterStub,
	}

	testLedger.Register(&cobra.Command{}, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Pretend()
	assert.True(t, calledWorkflow)
}

func TestNodesExecutor_PretendInvalidConfigType(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewAddNodesOptions()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		NodesOptions:        testOptions,
		workflowTaskFactory: newWorkflowTaskPretendStub(&calledWorkflow),
	}

	defer patch.Unpatch()
	testViper.Set(testOptions.optionConfigType.Key, "invalid")
	testLedger.Register(&cobra.Command{}, testOptions.addDefiners()...)
	testExecutor.Pretend()
	assert.False(t, calledWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNodesExecutor_PretendInvalidParams(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewAddNodesOptions()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		NodesOptions:          testOptions,
		workflowTaskFactory:   newWorkflowTaskPretendStub(&calledWorkflow),
		workflowWriterFactory: newWorkflowWriterStub,
	}

	defer patch.Unpatch()
	testLedger.Register(&cobra.Command{}, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Pretend()
	assert.False(t, calledWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNodesExecutor_Proceed(t *testing.T) {
	var calledWorkflow bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewRemoveNodesOptions()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		NodesOptions:          testOptions,
		workflowArg:           "workflow",
		workflowTaskFactory:   newWorkflowTaskCompleteStub(&calledWorkflow),
		workflowWriterFactory: newWorkflowWriterStub,
	}

	testLedger.InitLogging()
	testLedger.Register(&cobra.Command{}, testOptions.removeDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Proceed()
	assert.True(t, calledWorkflow)
}

func TestNodesExecutor_ProceedDuplicateHandle(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewRemoveNodesOptions()
	var expectedNode = "duplicate-handle"
	var expectedWorkflow = "workflow-handle"
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		NodesOptions:        testOptions,
		workflowTaskFactory: newWorkflowTaskCompleteStub(&calledWorkflow),
		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedNode,
						},
					},
				},
			},
		}),
		addNode: &broker.WorkflowNode{
			Handle: expectedNode,
		},
		workflowArg: expectedWorkflow,
	}

	defer patch.Unpatch()
	testLedger.Register(&cobra.Command{}, testOptions.removeDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeYaml)
	testExecutor.Proceed()
	assert.False(t, calledWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNodesExecutor_ProceedInvalidConfigType(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewRemoveNodesOptions()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		NodesOptions:        testOptions,
		workflowTaskFactory: newWorkflowTaskCompleteStub(&calledWorkflow),
	}

	defer patch.Unpatch()
	testViper.Set(testOptions.optionConfigType.Key, "invalid")
	testLedger.Register(&cobra.Command{}, testOptions.removeDefiners()...)
	testExecutor.Proceed()
	assert.False(t, calledWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNodesExecutor_ProceedInvalidParams(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewRemoveNodesOptions()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		NodesOptions:          testOptions,
		workflowTaskFactory:   newWorkflowTaskCompleteStub(&calledWorkflow),
		workflowWriterFactory: newWorkflowWriterStub,
	}

	defer patch.Unpatch()
	testLedger.Register(&cobra.Command{}, testOptions.removeDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Proceed()
	assert.False(t, calledWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNodesExecutor_Remove(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedNode = "node"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewRemoveNodesOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newNodesRemoveExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(testCmd, testOptions.removeDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Remove(expectedWorkflow, expectedNode)
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow:[\s\t]*`+expectedWorkflow), actual)
	assert.Regexp(t, regexp.MustCompile(`node.remove\[0\].handle:[\s\t]*`+expectedNode), actual)
}

func TestNewNodes(t *testing.T) {
	var testLedger = config.NewLedger()
	var testCmd = NewNodes(testLedger, &Cli{})

	assert.Equal(t, 2, len(testCmd.Commands()))
}

func TestSerializedOptions_getDefault_dashSeq(t *testing.T) {
	var expectedOem = "genaiz.com"
	var expectedHandle = "handle_test"
	var expectedVersion = "0.0.1"
	var expectedSeq = 37
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var serializedOptions = NewSerializedOptions()
	var actual *broker.WorkflowNodeFunction

	testViper.Set(serializedOptions.optionSerialized.Key, fmt.Sprintf("%s/%s:%s-rc-%d", expectedOem, expectedHandle, expectedVersion, expectedSeq))
	actual = serializedOptions.optionDeserialized.DefaultGetter(testLedger).(*broker.WorkflowNodeFunction)
	assert.Equal(t, expectedOem, actual.Oem)
	assert.Equal(t, expectedHandle, actual.Handle)
	assert.Equal(t, expectedVersion, actual.Version)
	assert.Equal(t, expectedSeq, actual.Seq)
}

func TestSerializedOptions_getDefault_dotSeq(t *testing.T) {
	var expectedOem = "genaiz.com"
	var expectedHandle = "handle_test"
	var expectedVersion = "0.0.1"
	var expectedSeq = 37
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var serializedOptions = NewSerializedOptions()
	var actual *broker.WorkflowNodeFunction

	testViper.Set(serializedOptions.optionSerialized.Key, fmt.Sprintf("%s/%s:%s-rc.%d", expectedOem, expectedHandle, expectedVersion, expectedSeq))
	actual = serializedOptions.optionDeserialized.DefaultGetter(testLedger).(*broker.WorkflowNodeFunction)
	assert.Equal(t, expectedOem, actual.Oem)
	assert.Equal(t, expectedHandle, actual.Handle)
	assert.Equal(t, expectedVersion, actual.Version)
	assert.Equal(t, expectedSeq, actual.Seq)
}

func TestSerializedOptions_getDefault_noOem(t *testing.T) {
	var expectedHandle = "handle_test"
	var expectedVersion = "0.0.1"
	var expectedSeq = 37
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var serializedOptions = NewSerializedOptions()
	var actual *broker.WorkflowNodeFunction

	testViper.Set(serializedOptions.optionSerialized.Key, fmt.Sprintf("%s:%s-rc%d", expectedHandle, expectedVersion, expectedSeq))
	actual = serializedOptions.optionDeserialized.DefaultGetter(testLedger).(*broker.WorkflowNodeFunction)
	assert.Equal(t, expectedHandle, actual.Handle)
	assert.Equal(t, expectedVersion, actual.Version)
	assert.Equal(t, expectedSeq, actual.Seq)
}

func TestSerializedOptions_getDefault_noSeq(t *testing.T) {
	var expectedOem = "genaiz.com"
	var expectedHandle = "handle_test"
	var expectedVersion = "0.0.1"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var serializedOptions = NewSerializedOptions()
	var actual *broker.WorkflowNodeFunction

	testViper.Set(serializedOptions.optionSerialized.Key, fmt.Sprintf("%s/%s:%s", expectedOem, expectedHandle, expectedVersion))
	actual = serializedOptions.optionDeserialized.DefaultGetter(testLedger).(*broker.WorkflowNodeFunction)
	assert.Equal(t, expectedOem, actual.Oem)
	assert.Equal(t, expectedHandle, actual.Handle)
	assert.Equal(t, expectedVersion, actual.Version)
}

func TestSerializedOptions_getNode_illegalCast(t *testing.T) {
	var testOptions = NewSerializedOptions()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()

	testViper.Set(testOptions.optionDeserialized.Key, "notANode")
	actual := testOptions.getNode(testLedger)
	assert.NotNil(t, actual)
	assert.Empty(t, actual.Handle)
	assert.Empty(t, actual.Oem)
	assert.Empty(t, actual.Seq)
	assert.Empty(t, actual.Version)
}

func TestSerializedOptions_getDefault_noVersion(t *testing.T) {
	var expectedOem = "genaiz.com"
	var expectedHandle = "handle_test"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var serializedOptions = NewSerializedOptions()
	var actual *broker.WorkflowNodeFunction

	testViper.Set(serializedOptions.optionSerialized.Key, fmt.Sprintf("%s/%s", expectedOem, expectedHandle))
	actual = serializedOptions.optionDeserialized.DefaultGetter(testLedger).(*broker.WorkflowNodeFunction)
	assert.Equal(t, expectedOem, actual.Oem)
	assert.Equal(t, expectedHandle, actual.Handle)
}

func Test_validateArgNodes(t *testing.T) {
	assert.NoError(t, validateArgNodes("handle", "more-handle"))
	assert.Error(t, validateArgNodes("valid.handle", "_invalid_handle"))
}

func newWorkflowWriterFactory(stubSolution *broker.Solution) workflowWriterFactory {
	return func(*config.Ledger, string) *workflowWriter {
		var stub = &workflowWriter{
			WorkflowWriter: config.NewWorkflowWriter(),
		}

		stub.WithCurrent(stubSolution)
		return stub
	}
}
