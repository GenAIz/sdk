package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type BaseReader struct {
	configName string
	configPath string
	configType shared.ConfigType
}

func (br *BaseReader) GetSolutionFile() string {
	var configPath = br.GetSolutionPath()

	return filepath.Join(configPath, br.configName+"."+br.configType)
}

func (br *BaseReader) GetSolutionPath(path ...string) string {
	var paths []string

	paths = append(paths, br.configPath)
	paths = append(paths, path...)
	return filepath.Join(paths...)
}

func (br *BaseReader) Read(configType shared.ConfigType, path ...string) (*broker.Solution, error) {
	var configPath = br.GetSolutionPath(path...)
	var input = filepath.Join(configPath, br.configName+"."+configType)
	var result *broker.Solution
	var err error

	if result, err = br.ReadFile(input); err == nil {
		br.configType = configType
		br.configPath = configPath
		return result, nil
	}

	return nil, err
}

func (br *BaseReader) ReadFile(filePath string) (*broker.Solution, error) {
	var vp = viper.New()
	var err error

	vp.SetConfigFile(filePath)

	if err = vp.ReadInConfig(); err == nil {
		var solution *broker.Solution

		if err = vp.UnmarshalKey("solution", &solution); err == nil {
			return solution, nil
		}
	}

	return nil, err
}

type SolutionReader struct {
	BaseReader
	current *broker.Solution
	ledger  *Ledger
}

func (sr *SolutionReader) FindFunctionValues() map[string]*viper.Viper {
	var paths = make(map[string]*viper.Viper)

	if dirEntries, err := os.ReadDir(sr.configPath); err == nil {
		for _, entry := range dirEntries {
			if entry.IsDir() {
				var vp = viper.New()

				vp.SetConfigName(sr.configName)
				vp.SetConfigType(sr.configType)
				vp.AddConfigPath(sr.GetSolutionPath(entry.Name()))

				if err = vp.ReadInConfig(); err == nil {
					paths[vp.ConfigFileUsed()] = vp
				} else {
					sr.ledger.Logger.Errorf("could not parse %s: %s", vp.ConfigFileUsed(), err)
				}
			}
		}
	}

	return paths
}

func (sr *SolutionReader) GetSolution() *broker.Solution {
	return sr.current
}

func (sr *SolutionReader) GetVersion() string {
	if sr.current != nil {
		return sr.current.Version
	}

	return ""
}

func (sr *SolutionReader) Read(configType shared.ConfigType, path ...string) error {
	var solution *broker.Solution
	var err error

	if solution, err = sr.BaseReader.Read(configType, path...); err == nil {
		sr.configType = configType
		sr.current = solution
		return nil
	}

	return err
}

func (sr *SolutionReader) WithConfigPath(path string) *SolutionReader {
	sr.configPath = path
	return sr
}

func (sr *SolutionReader) WithConfigType(configType shared.ConfigType) *SolutionReader {
	sr.configType = configType
	return sr
}

func NewSolutionReader(ledger *Ledger) *SolutionReader {
	return &SolutionReader{
		BaseReader: BaseReader{
			configName: ledger.ConfigName,
		},
		ledger: ledger,
	}
}
