package wf

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/dk"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

func TestPropExecutor_Add(t *testing.T) {
	var testFolder = t.TempDir()
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		WithOutput(testOutput).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var factory = newPropAddExecutorFactory(testLedger, testCli, &PropOptions{})
	var testExecutor = factory(&cobra.Command{})
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedNodeArg = "nodeHandle"
	var expectedWorkflowArg = "workflow"
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		var expectedFunctionHandle = "handle"
		var expectedFunctionOem = "oem"
		var expectedFunctionVersion = "version"

		filez.CloseSilently(fd)

		if err = os.MkdirAll(filepath.Join(testFolder, expectedFunctionHandle), 0750); err == nil {
			testViper.Set("solution", &broker.Solution{
				Handle: "solution",
				Workflows: []broker.Workflow{
					{
						Handle: expectedWorkflowArg,
						Nodes: []broker.WorkflowNode{
							{
								Handle: expectedNodeArg,
								Sf: &broker.WorkflowNodeFunction{
									Handle:  expectedFunctionHandle,
									Oem:     expectedFunctionOem,
									Version: expectedFunctionVersion,
								},
							},
						},
					},
				},
			})

			if err = testViper.WriteConfigAs(testConfigFile); err == nil {
				var testFunction = &broker.Function{
					Handle:  expectedFunctionHandle,
					Oem:     expectedFunctionOem,
					Version: expectedFunctionVersion,
				}

				testViper = viper.New()
				testViper.Set(schema.Genaiz.Function.Publish.Internal.Doc, &testFunction)

				if err = testViper.WriteConfigAs(filepath.Join(testFolder, expectedFunctionHandle, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
					var expectedKey = "key"
					var expectedValue = "value"

					assert.NoError(t, testExecutor.Add(expectedWorkflowArg, expectedNodeArg, expectedKey, expectedValue))
					actual := testOutput.String()
					assert.Contains(t, actual, expectedWorkflowArg)
					assert.Contains(t, actual, expectedNodeArg)
					assert.Contains(t, actual, expectedKey)
					assert.Contains(t, actual, expectedValue)
					return
				}
			}
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Add_ExternalFunction(t *testing.T) {
	var testFolder = t.TempDir()
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		WithOutput(testOutput).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},

		addProps: make(map[string]string),
	}
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedNodeArg = "node"
	var expectedWorkflowArg = "workflow"
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		filez.CloseSilently(fd)
		testViper.Set("solution", &broker.Solution{
			Handle: "solution",
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflowArg,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedNodeArg,
							Sf: &broker.WorkflowNodeFunction{
								Handle: "handle",
							},
						},
					},
				},
			},
		})

		if err = testViper.WriteConfigAs(testConfigFile); err == nil {
			assert.NoError(t, testExecutor.Add(expectedWorkflowArg, expectedNodeArg, "key", "value"))
			actual := testOutput.String()
			assert.Contains(t, actual, expectedWorkflowArg)
			assert.Contains(t, actual, expectedNodeArg)
			return
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Add_NoGenaizPathError(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithWorkDir(t.TempDir()).
		Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}

	assert.Error(t, testExecutor.Add("workflow", "node", "key", "value"))
}

func TestPropExecutor_Add_NoFunctionError(t *testing.T) {
	var testFolder = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedNodeArg = "node"
	var expectedWorkflowArg = "workflow"
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		filez.CloseSilently(fd)
		testViper.Set("solution", &broker.Solution{
			Handle: "solution",
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflowArg,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedNodeArg,
						},
					},
				},
			},
		})

		if err = testViper.WriteConfigAs(testConfigFile); err == nil {
			assert.ErrorIs(t, testExecutor.Add(expectedWorkflowArg, expectedNodeArg, "key", "value"), errorNoFunction)
			return
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Add_NoMemberError(t *testing.T) {
	var testFolder = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedNodeArg = "node"
	var expectedWorkflowArg = "workflow"
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		filez.CloseSilently(fd)
		testViper.Set("solution", &broker.Solution{
			Handle: "solution",
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflowArg,
					Nodes: []broker.WorkflowNode{
						{
							Handle: "notTheRightHandle",
						},
					},
				},
			},
		})

		if err = testViper.WriteConfigAs(testConfigFile); err == nil {
			assert.Error(t, testExecutor.Add(expectedWorkflowArg, expectedNodeArg, "key", "value"))
			return
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Add_NoSolutionError(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithWorkDir(t.TempDir()).
		Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(filepath.Join(testLedger.WorkDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
		filez.CloseSilently(fd)

		assert.ErrorIs(t, testExecutor.Add("workflow", "node", "key", "value"), errorNoSolution)
		return
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Edit(t *testing.T) {
	var testFolder = t.TempDir()
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		WithOutput(testOutput).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var factory = newPropEditExecutorFactory(testLedger, testCli, &PropOptions{})
	var testExecutor = factory(&cobra.Command{})
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedNodeArg = "node"
	var expectedWorkflowArg = "workflow"
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		var expectedKey = "key"
		var expectedValue = "edited"

		filez.CloseSilently(fd)
		testViper.Set("solution", &broker.Solution{
			Handle: "solution",
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflowArg,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedNodeArg,
							Props: map[string]string{
								expectedKey: "value",
							},
							Sf: &broker.WorkflowNodeFunction{
								Handle: "functionHandle",
							},
						},
					},
				},
			},
		})

		if err = testViper.WriteConfigAs(testConfigFile); err == nil {
			assert.NoError(t, testExecutor.Edit(expectedWorkflowArg, expectedNodeArg, "key", expectedValue))
			actual := testOutput.String()
			assert.Contains(t, actual, expectedWorkflowArg)
			assert.Contains(t, actual, expectedNodeArg)
			assert.Regexp(t, regexp.MustCompile(`prop.rm.key:[\s\t]*`+expectedKey), actual)
			assert.Contains(t, actual, expectedValue)
			assert.Regexp(t, regexp.MustCompile(`prop.add.key:[\s\t]*`+expectedKey), actual)
			return
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Edit_OnFolder(t *testing.T) {
	var testFolder = t.TempDir()
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		WithOutput(testOutput).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var factory = newPropEditExecutorFactory(testLedger, testCli, &PropOptions{})
	var testExecutor = factory(&cobra.Command{})
	var expectedNodeArg = "node"
	var expectedNodeHandle = "nodeHandle"
	var expectedWorkflowArg = "workflow"
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var testFunctionPath = filepath.Join(testFolder, expectedNodeArg)
	var testFunctionFile = filepath.Join(testFunctionPath, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		var expectedFunctionHandle = "functionHandle"
		var expectedFunctionOem = "functionOem"
		var expectedFunctionVersion = "1.0.0"
		var expectedKey = "key"
		var expectedValue = "edited"

		filez.CloseSilently(fd)
		testViper.Set("solution", &broker.Solution{
			Handle: "solution",
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflowArg,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedNodeHandle,
							Props: map[string]string{
								expectedKey: "value",
							},
							Sf: &broker.WorkflowNodeFunction{
								Handle:  expectedFunctionHandle,
								Oem:     expectedFunctionOem,
								Version: expectedFunctionVersion,
							},
						},
					},
				},
			},
		})

		if err = testViper.WriteConfigAs(testConfigFile); err == nil {
			testViper = viper.New()
			testViper.Set(schema.Genaiz.Function.Publish.Internal.Doc, &broker.Function{
				Handle:  expectedFunctionHandle,
				Oem:     expectedFunctionOem,
				Version: expectedFunctionVersion,
			})

			if err = os.MkdirAll(testFunctionPath, 0750); err == nil {
				if err = testViper.WriteConfigAs(testFunctionFile); err == nil {
					t.Chdir(testFolder)
					assert.NoError(t, testExecutor.Edit(expectedWorkflowArg, expectedNodeArg, "key", expectedValue))
					actual := testOutput.String()
					assert.Contains(t, actual, expectedWorkflowArg)
					assert.Contains(t, actual, expectedNodeArg)
					assert.Regexp(t, regexp.MustCompile(`prop.rm.key:[\s\t]*`+expectedKey), actual)
					assert.Contains(t, actual, expectedValue)
					assert.Regexp(t, regexp.MustCompile(`prop.add.key:[\s\t]*`+expectedKey), actual)
					return
				}
			}
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Edit_NoNodeError(t *testing.T) {
	var testFolder = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedNodeArg = "node"
	var expectedWorkflowArg = "workflow"
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		filez.CloseSilently(fd)
		testViper.Set("solution", &broker.Solution{
			Handle: "solution",
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflowArg,
				},
			},
		})

		if err = testViper.WriteConfigAs(testConfigFile); err == nil {
			assert.Error(t, testExecutor.Edit(expectedWorkflowArg, expectedNodeArg, "key", "value"))
			return
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Edit_NoPropError(t *testing.T) {
	var testFolder = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedNodeArg = "node"
	var expectedWorkflowArg = "workflow"
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		filez.CloseSilently(fd)
		testViper.Set("solution", &broker.Solution{
			Handle: "solution",
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflowArg,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedNodeArg,
							Sf: &broker.WorkflowNodeFunction{
								Handle: "functionHandle",
							},
						},
					},
				},
			},
		})

		if err = testViper.WriteConfigAs(testConfigFile); err == nil {
			assert.Error(t, testExecutor.Edit(expectedWorkflowArg, expectedNodeArg, "key", "value"))
			return
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Init(t *testing.T) {
	var testFolder = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedNodeHandle = "nodeHandle"
	var expectedHandle = "handle"
	var expectedOem = "oem"
	var expectedVersion = "1.0.0"
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		var functionPath = filepath.Join(testFolder, expectedHandle)

		filez.CloseSilently(fd)
		testViper.Set("solution", &broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: "workflow",
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedNodeHandle,
							Sf: &broker.WorkflowNodeFunction{
								Handle:  expectedHandle,
								Oem:     expectedOem,
								Version: expectedVersion,
							},
						},
					},
				},
			},
		})

		if err = testViper.WriteConfigAs(testConfigFile); err == nil {
			if err = os.MkdirAll(functionPath, 0750); err == nil {
				if fd, err = os.Create(filepath.Join(functionPath, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
					testViper = viper.New()
					testViper.Set(schema.Genaiz.Function.Publish.Internal.Doc, &broker.Function{
						Handle:  expectedHandle,
						Oem:     expectedOem,
						Version: expectedVersion,
					})

					if err = testViper.WriteConfigAs(fd.Name()); err == nil {
						var actual string

						filez.CloseSilently(fd)
						t.Chdir(testFolder)
						actual, err = testExecutor.Init(expectedHandle)
						assert.NoError(t, err)
						assert.Equal(t, expectedNodeHandle, actual)
						return
					}
				}
			}
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Init_InvalidFunction(t *testing.T) {
	var testFolder = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedHandle = "handle"
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		var functionPath = filepath.Join(testFolder, expectedHandle)

		filez.CloseSilently(fd)

		if err = os.MkdirAll(functionPath, 0750); err == nil {
			if fd, err = os.Create(filepath.Join(functionPath, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
				var actual string

				filez.CloseSilently(fd)
				t.Chdir(testFolder)
				actual, err = testExecutor.Init(expectedHandle)
				assert.ErrorIs(t, err, errorNoFunction)
				assert.Empty(t, actual)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Init_InvalidFunctionHandle(t *testing.T) {
	var testFolder = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedHandle = "handle"
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		var functionPath = filepath.Join(testFolder, expectedHandle)

		filez.CloseSilently(fd)

		if err = os.MkdirAll(functionPath, 0750); err == nil {
			if fd, err = os.Create(filepath.Join(functionPath, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
				testViper.Set(schema.Genaiz.Function.Publish.Internal.Doc, &broker.Function{
					Oem: "oem",
				})

				if err = testViper.WriteConfigAs(fd.Name()); err == nil {
					var actual string

					filez.CloseSilently(fd)
					t.Chdir(testFolder)
					actual, err = testExecutor.Init(expectedHandle)
					assert.Error(t, err)
					assert.Contains(t, err.Error(), expectedHandle)
					assert.Empty(t, actual)
					return
				}
			}
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Init_InvalidFunctionOem(t *testing.T) {
	var testFolder = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedHandle = "handle"
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		var functionPath = filepath.Join(testFolder, expectedHandle)

		filez.CloseSilently(fd)

		if err = os.MkdirAll(functionPath, 0750); err == nil {
			if fd, err = os.Create(filepath.Join(functionPath, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
				testViper.Set(schema.Genaiz.Function.Publish.Internal.Doc, &broker.Function{
					Handle: expectedHandle,
				})

				if err = testViper.WriteConfigAs(fd.Name()); err == nil {
					var actual string

					filez.CloseSilently(fd)
					t.Chdir(testFolder)
					actual, err = testExecutor.Init(expectedHandle)
					assert.Error(t, err)
					assert.Contains(t, err.Error(), expectedHandle)
					assert.Empty(t, actual)
					return
				}
			}
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Init_InvalidFunctionVersion(t *testing.T) {
	var testFolder = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedHandle = "handle"
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		var functionPath = filepath.Join(testFolder, expectedHandle)

		filez.CloseSilently(fd)

		if err = os.MkdirAll(functionPath, 0750); err == nil {
			if fd, err = os.Create(filepath.Join(functionPath, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
				testViper.Set(schema.Genaiz.Function.Publish.Internal.Doc, &broker.Function{
					Handle: expectedHandle,
					Oem:    "oem",
				})

				if err = testViper.WriteConfigAs(fd.Name()); err == nil {
					var actual string

					filez.CloseSilently(fd)
					t.Chdir(testFolder)
					actual, err = testExecutor.Init(expectedHandle)
					assert.Error(t, err)
					assert.Contains(t, err.Error(), expectedHandle)
					assert.Empty(t, actual)
					return
				}
			}
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Init_InvalidSolution(t *testing.T) {
	var testFolder = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedHandle = "handle"
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		var functionPath = filepath.Join(testFolder, expectedHandle)

		filez.CloseSilently(fd)

		if err = os.MkdirAll(functionPath, 0750); err == nil {
			if fd, err = os.Create(filepath.Join(functionPath, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
				testViper.Set(schema.Genaiz.Function.Publish.Internal.Doc, &broker.Function{
					Handle:  expectedHandle,
					Oem:     "oem",
					Version: "1.0.0",
				})

				if err = testViper.WriteConfigAs(fd.Name()); err == nil {
					var actual string

					filez.CloseSilently(fd)
					t.Chdir(testFolder)
					actual, err = testExecutor.Init(expectedHandle)
					assert.ErrorIs(t, err, errorNoSolution)
					assert.Empty(t, actual)
					return
				}
			}
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Init_NodeNotFound(t *testing.T) {
	var testFolder = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedHandle = "handle"
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		var functionPath = filepath.Join(testFolder, expectedHandle)

		filez.CloseSilently(fd)
		testViper.Set("solution", &broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: "workflow",
					Nodes: []broker.WorkflowNode{
						{
							Handle: "node",
						},
					},
				},
			},
		})

		if err = testViper.WriteConfigAs(testConfigFile); err == nil {
			if err = os.MkdirAll(functionPath, 0750); err == nil {
				if fd, err = os.Create(filepath.Join(functionPath, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
					testViper = viper.New()
					testViper.Set(schema.Genaiz.Function.Publish.Internal.Doc, &broker.Function{
						Handle:  expectedHandle,
						Oem:     "oem",
						Version: "1.0.0",
					})

					if err = testViper.WriteConfigAs(fd.Name()); err == nil {
						var actual string

						filez.CloseSilently(fd)
						t.Chdir(testFolder)
						actual, err = testExecutor.Init(expectedHandle)
						assert.ErrorIs(t, err, errorNoWorkflowNode)
						assert.Empty(t, actual)
						return
					}
				}
			}
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Init_NoNodeWorkflows(t *testing.T) {
	var testFolder = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedHandle = "handle"
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		var functionPath = filepath.Join(testFolder, expectedHandle)

		filez.CloseSilently(fd)
		testViper.Set("solution", &broker.Solution{})

		if err = testViper.WriteConfigAs(testConfigFile); err == nil {
			if err = os.MkdirAll(functionPath, 0750); err == nil {
				if fd, err = os.Create(filepath.Join(functionPath, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
					testViper = viper.New()
					testViper.Set(schema.Genaiz.Function.Publish.Internal.Doc, &broker.Function{
						Handle:  expectedHandle,
						Oem:     "oem",
						Version: "1.0.0",
					})

					if err = testViper.WriteConfigAs(fd.Name()); err == nil {
						var actual string

						filez.CloseSilently(fd)
						t.Chdir(testFolder)
						actual, err = testExecutor.Init(expectedHandle)
						assert.ErrorIs(t, err, errorNoWorkflowNode)
						assert.Empty(t, actual)
						return
					}
				}
			}
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Init_PathError(t *testing.T) {
	var testFolder = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedHandle = "handle"
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		var actual string

		filez.CloseSilently(fd)
		actual, err = testExecutor.Init(expectedHandle)
		assert.NoError(t, err)
		assert.Equal(t, expectedHandle, actual)
		return
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Pretend(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testWorkDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testWorkDir).
		Build()
	var testConfigFile = filepath.Join(testWorkDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedWorkflow = "workflow"
	var testParams broker.WorkflowParams
	var testPropParams broker.WorkflowPropParams
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
		PropOptions:  NewAddPropOptions(),

		function:                &broker.Function{},
		workflow:                &broker.Workflow{Handle: expectedWorkflow},
		workflowPropTaskFactory: newWorkflowPropTaskPretendCapture(&testPropParams),
		workflowTaskFactory:     newWorkflowTaskPretendCapture(&testParams),
		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
				},
			},
		}),
	}
	var err error

	defer patch.Unpatch()
	testViper.Set("solution", &broker.Solution{})

	if err = testViper.WriteConfigAs(testConfigFile); err == nil {
		testExecutor.Pretend()
		assert.False(t, patch.Called)
		return
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Pretend_ConfigInvalid(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testLedger = config.NewBuilder().Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}

	defer patch.Unpatch()
	testExecutor.Pretend()
	assert.True(t, patch.Called)
	assert.Equal(t, patch.CalledWith, 1)
}

func TestPropExecutor_Pretend_NoValidation(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testWorkDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testWorkDir).
		Build()
	var testConfigFile = filepath.Join(testWorkDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedWorkflow = "workflow"
	var testParams broker.WorkflowParams
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
		PropOptions:  &PropOptions{},

		workflow:            &broker.Workflow{Handle: expectedWorkflow},
		workflowTaskFactory: newWorkflowTaskPretendCapture(&testParams),
		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
				},
			},
		}),
	}
	var err error

	defer patch.Unpatch()
	testViper.Set("solution", &broker.Solution{})

	if err = testViper.WriteConfigAs(testConfigFile); err == nil {
		testExecutor.Pretend()
		assert.False(t, patch.Called)
		return
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Pretend_ValidateNoSyncError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testWorkDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testWorkDir).
		Build()
	var testConfigFile = filepath.Join(testWorkDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedWorkflow = "workflow"
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
		PropOptions:  NewEditPropOptions(),

		workflow: &broker.Workflow{Handle: expectedWorkflow},
		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
				},
			},
		}),
	}
	var err error

	defer patch.Unpatch()
	testViper.Set("solution", &broker.Solution{})

	if err = testViper.WriteConfigAs(testConfigFile); err == nil {
		testExecutor.Pretend()
		assert.True(t, patch.Called)
		assert.Equal(t, patch.CalledWith, 1)
		return
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Pretend_WorkflowNotFound(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testWorkDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testWorkDir).
		Build()
	var testConfigFile = filepath.Join(testWorkDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},

		workflow:              &broker.Workflow{Handle: "handle"},
		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{}),
	}
	var err error

	defer patch.Unpatch()
	testViper.Set("solution", &broker.Solution{})

	if err = testViper.WriteConfigAs(testConfigFile); err == nil {
		testExecutor.Pretend()
		assert.True(t, patch.Called)
		assert.Equal(t, patch.CalledWith, 1)
		return
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Proceed(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testWorkDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testWorkDir).
		Build()
	var testConfigFile = filepath.Join(testWorkDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedWorkflow = "workflow"
	var expectedWorkflowNode = "node"
	var capturedPropParams broker.WorkflowPropParams
	var capturedWorkflowParams broker.WorkflowParams
	var capturedDataLinkParams broker.DataLinkParams
	var testDataLink = &broker.DataLink{
		Oem:     "oem",
		Handle:  "handle",
		Version: "0.1.1",
	}
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
		SyncBridge: dk.NewSyncBridgeBuilder().
			WithDataLinksWriterFactory(newDataLinksWriterTestFactory([]broker.DataLink{*testDataLink})).
			WithExportLinkTaskFactory(newExportLinkCompleteCapture(&capturedDataLinkParams)).
			Build(),
		PropOptions: NewEditPropOptions(),

		rmProps:  []string{"notExisting"},
		addProps: map[string]string{"PROP": "value"},
		function: &broker.Function{
			DataSources: []string{"oem/handle:version"},
			PropSpecs: []broker.PropSpec{
				{
					Key: "Test",
				},
			},
		},
		workflow:                &broker.Workflow{Handle: expectedWorkflow},
		workflowNode:            &broker.WorkflowNode{Handle: expectedWorkflowNode},
		workflowPropTaskFactory: newWorkflowPropTaskCompleteCapture(&capturedPropParams),
		workflowTaskFactory:     newWorkflowTaskCompleteCapture(&capturedWorkflowParams),
		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedWorkflowNode,
						},
					},
				},
			},
		}),
	}
	var err error

	defer patch.Unpatch()
	testViper.Set("solution", &broker.Solution{})

	if err = testViper.WriteConfigAs(testConfigFile); err == nil {
		testLedger.Logger = logrus.New()
		testExecutor.Proceed()
		assert.False(t, patch.Called)
		assert.NotNil(t, capturedPropParams)
		assert.NotNil(t, capturedWorkflowParams)
		assert.NotNil(t, capturedDataLinkParams)
		return
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Proceed_ConfigInvalid(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testLedger = config.NewBuilder().Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}

	defer patch.Unpatch()
	testExecutor.Proceed()
	assert.True(t, patch.Called)
	assert.Equal(t, patch.CalledWith, 1)
}

func TestPropExecutor_Proceed_DuplicateKeyError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedKey = "key"
	var expectedWorkflow = "workflow"
	var expectedWorkflowNode = "node"
	var testWorkDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testWorkDir).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testSolution = &broker.Solution{
		Handle: "solution",
		Workflows: []broker.Workflow{
			{
				Handle: expectedWorkflow,
				Nodes: []broker.WorkflowNode{
					{
						Handle: expectedWorkflowNode,
						Props: map[string]string{
							expectedKey: "value1",
						},
						Sf: &broker.WorkflowNodeFunction{
							Handle: "handle",
						},
					},
				},
			},
		},
	}
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PropOptions: NewAddPropOptions(),

		addProps: map[string]string{
			expectedKey: "value",
		},

		function:              &broker.Function{DataSources: []string{"notValid"}},
		workflow:              &broker.Workflow{Handle: expectedWorkflow},
		workflowNode:          &broker.WorkflowNode{Handle: expectedWorkflowNode},
		workflowWriterFactory: newWorkflowWriterFactory(testSolution),
	}
	var testConfigFile = filepath.Join(testWorkDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var fd *os.File
	var err error

	defer patch.Unpatch()

	if fd, err = os.Create(testConfigFile); err == nil {
		filez.CloseSilently(fd)
		testViper.Set("solution", testSolution)

		if err = testViper.WriteConfigAs(testConfigFile); err == nil {
			testExecutor.Proceed()
			assert.True(t, patch.Called)
			assert.Equal(t, patch.CalledWith, 1)
			return
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Proceed_ValidateExternal(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testWorkDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testWorkDir).
		Build()
	var testConfigFile = filepath.Join(testWorkDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedWorkflow = "workflow"
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
		PropOptions:  NewEditPropOptions(),

		external: &broker.Function{},
		workflow: &broker.Workflow{Handle: expectedWorkflow},
		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
				},
			},
		}),
	}
	var err error

	defer patch.Unpatch()
	testViper.Set("solution", &broker.Solution{})

	if err = testViper.WriteConfigAs(testConfigFile); err == nil {
		testExecutor.Proceed()
		assert.True(t, patch.Called)
		assert.Equal(t, patch.CalledWith, 1)
		return
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Proceed_ValidateInvalidDataLink(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testWorkDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testWorkDir).
		Build()
	var testConfigFile = filepath.Join(testWorkDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedWorkflow = "workflow"
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
		PropOptions:  NewEditPropOptions(),

		function: &broker.Function{DataSources: []string{"notValid"}},
		workflow: &broker.Workflow{Handle: expectedWorkflow},
		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
				},
			},
		}),
	}
	var err error

	defer patch.Unpatch()
	testViper.Set("solution", &broker.Solution{})

	if err = testViper.WriteConfigAs(testConfigFile); err == nil {
		testExecutor.Proceed()
		assert.True(t, patch.Called)
		assert.Equal(t, patch.CalledWith, 1)
		return
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Proceed_WorkflowNodeNotFound(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testWorkDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testWorkDir).
		Build()
	var testConfigFile = filepath.Join(testWorkDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedWorkflow = "workflow"
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},

		addProps:     map[string]string{"PROP": "value"},
		workflow:     &broker.Workflow{Handle: expectedWorkflow},
		workflowNode: &broker.WorkflowNode{Handle: "node"},
		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflow,
				},
			},
		}),
	}
	var err error

	defer patch.Unpatch()
	testViper.Set("solution", &broker.Solution{})

	if err = testViper.WriteConfigAs(testConfigFile); err == nil {
		testExecutor.Proceed()
		assert.True(t, patch.Called)
		assert.Equal(t, patch.CalledWith, 1)
		return
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Proceed_WorkflowNotFound(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testWorkDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testWorkDir).
		Build()
	var testConfigFile = filepath.Join(testWorkDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},

		workflow:              &broker.Workflow{Handle: "handle"},
		workflowWriterFactory: newWorkflowWriterFactory(&broker.Solution{}),
	}
	var err error

	defer patch.Unpatch()
	testViper.Set("solution", &broker.Solution{})

	if err = testViper.WriteConfigAs(testConfigFile); err == nil {
		testExecutor.Proceed()
		assert.True(t, patch.Called)
		assert.Equal(t, patch.CalledWith, 1)
		return
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Remove(t *testing.T) {
	var testFolder = t.TempDir()
	var testViper = viper.New()
	var testOutput = new(bytes.Buffer)
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		WithOutput(testOutput).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var factory = newPropRemoveExecutorFactory(testLedger, testCli, &PropOptions{})
	var testExecutor = factory(&cobra.Command{})
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		var expectedFunctionPath = "path"
		var expectedNodeHandle = "node"
		var expectedWorkflowHandle = "workflow"
		var testFunction = &broker.Function{
			Handle:  "function",
			Oem:     "oem",
			Version: "1.0.0",
		}

		filez.CloseSilently(fd)
		testViper.Set("solution", &broker.Solution{
			Handle: "solution",
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflowHandle,
					Nodes: []broker.WorkflowNode{
						{
							Handle: expectedNodeHandle,
							Sf: &broker.WorkflowNodeFunction{
								Handle:  testFunction.Handle,
								Oem:     testFunction.Oem,
								Version: testFunction.Version,
							},
						},
					},
				},
			},
		})

		if err = testViper.WriteConfigAs(testConfigFile); err == nil {
			var functionPath = filepath.Join(testFolder, expectedFunctionPath)

			if err = os.MkdirAll(functionPath, 0750); err == nil {
				testViper = viper.New()
				testViper.Set(schema.Genaiz.Function.Publish.Internal.Doc, testFunction)

				if err = testViper.WriteConfigAs(filepath.Join(functionPath, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
					var expectedKey = "key"

					t.Chdir(testFolder)
					assert.NoError(t, testExecutor.Remove(expectedWorkflowHandle, expectedNodeHandle, expectedKey))
					actual := testOutput.String()
					assert.Contains(t, actual, expectedWorkflowHandle)
					assert.Contains(t, actual, expectedNodeHandle)
					assert.Contains(t, actual, expectedKey)
					return
				}
			}
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Remove_FunctionNotFound(t *testing.T) {
	var testFolder = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		var expectedFunctionPath = "path"
		var expectedWorkflowHandle = "workflow"

		filez.CloseSilently(fd)
		testViper.Set("solution", &broker.Solution{
			Handle: "solution",
			Workflows: []broker.Workflow{
				{
					Handle: expectedWorkflowHandle,
				},
			},
		})

		if err = testViper.WriteConfigAs(testConfigFile); err == nil {
			var functionPath = filepath.Join(testFolder, expectedFunctionPath)

			if err = os.MkdirAll(functionPath, 0750); err == nil {
				testViper = viper.New()
				testViper.Set(schema.Genaiz.Function.Publish.Internal.Doc, &broker.Function{
					Handle:  "function",
					Oem:     "oem",
					Version: "1.0.0",
				})

				if err = testViper.WriteConfigAs(filepath.Join(functionPath, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
					t.Chdir(testFolder)
					assert.Error(t, testExecutor.Remove(expectedWorkflowHandle, expectedFunctionPath, "key"))
					return
				}
			}
		}
	}

	assert.NoError(t, err)
}

func TestPropExecutor_Remove_WorkflowNotFound(t *testing.T) {
	var testFolder = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithWorkDir(testFolder).
		Build()
	var testExecutor = &PropExecutor{
		BaseExecutor: BaseExecutor{Ledger: testLedger},
	}
	var testConfigFile = filepath.Join(testFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var fd *os.File
	var err error

	if fd, err = os.Create(testConfigFile); err == nil {
		filez.CloseSilently(fd)
		testViper.Set("solution", &broker.Solution{
			Handle: "solution",
			Workflows: []broker.Workflow{
				{
					Handle: "NotWorkflow",
				},
			},
		})

		if err = testViper.WriteConfigAs(testConfigFile); err == nil {
			assert.Error(t, testExecutor.Remove("workflow", "node", "key"))
			return
		}
	}

	assert.NoError(t, err)
}

func newDataLinksWriterTestFactory(current []broker.DataLink) dk.DataLinksWriterFactory {
	return func(ledger *config.Ledger, s string) *dk.DataLinksWriter {
		var reader = &config.DataLinksReader{}

		return &dk.DataLinksWriter{
			DataLinksWriter: &config.DataLinksWriter{
				DataLinksReader: *reader.WithCurrent(current),
			},
		}
	}
}

func newExportLinkCompleteCapture(capture *broker.DataLinkParams) dk.ExportLinkTaskFactory {
	return func(writer broker.DataLinkWriter) *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataLinkParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newWorkflowPropTaskCompleteCapture(capture *broker.WorkflowPropParams) WorkflowPropTaskFactory {
	return func() *task.Task[broker.WorkflowPropParams] {
		return &task.Task[broker.WorkflowPropParams]{
			OnPrepare: func(params *broker.WorkflowPropParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkflowPropParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newWorkflowPropTaskPretendCapture(capture *broker.WorkflowPropParams) WorkflowPropTaskFactory {
	return func() *task.Task[broker.WorkflowPropParams] {
		return &task.Task[broker.WorkflowPropParams]{
			OnPrepare: func(params *broker.WorkflowPropParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.WorkflowPropParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}
