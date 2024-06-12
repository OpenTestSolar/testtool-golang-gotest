package execute

import (
	"errors"
	"fmt"
	gotestBuilder "gotest/pkg/builder"
	gotestRunner "gotest/pkg/runner"
	gotestTestcase "gotest/pkg/testcase"
	gotestUtil "gotest/pkg/util"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type ExecuteOptions struct {
	executePath string
}

// NewBuildOptions new build options with default value
func NewExecuteOptions() *ExecuteOptions {
	return &ExecuteOptions{}
}

// NewCmdBuild create a build command
func NewCmdExecute() *cobra.Command {
	o := NewExecuteOptions()
	cmd := cobra.Command{
		Use:   "execute",
		Short: "Execute testcases",
		Run: func(cmd *cobra.Command, args []string) {
			o.RunExecute(cmd)
		},
	}
	cmd.Flags().StringVarP(&o.executePath, "path", "p", "", "Path of testcase info")
	cmd.MarkFlagRequired("path")
	return &cmd
}

func groupTestCasesByPathAndName(projPath string, testcases []*gotestTestcase.TestCase) (map[string]map[string][]*gotestTestcase.TestCase, error) {
	packages := map[string]map[string][]*gotestTestcase.TestCase{}
	for _, testcase := range testcases {
		path, name, err := gotestUtil.GetPathAndFileName(projPath, testcase.Path)
		if err != nil {
			log.Printf("[PLUGIN]Get path and file name from %s failed, err: %s", testcase.Path, err.Error())
			return nil, err
		}
		path = strings.TrimSuffix(path, string(os.PathSeparator))
		_, ok := packages[path]
		if !ok {
			packages[path] = map[string][]*gotestTestcase.TestCase{}
		}
		_, ok = packages[path][name]
		if !ok {
			packages[path][name] = []*gotestTestcase.TestCase{}
		}
		packages[path][name] = append(packages[path][name], testcase)
	}
	return packages, nil
}

func executeTestcases(projPath string, packages map[string]map[string][]*gotestTestcase.TestCase) error {
	for path, filesCases := range packages {
		pkgBin := filepath.Join(projPath, path+".test")
		_, err := os.Stat(pkgBin)
		if err != nil {
			log.Printf("[PLUGIN]Can't find package bin file %s during running, try to build it...", pkgBin)
			_, err := gotestBuilder.BuildTestPackage(projPath, path, false)
			if err != nil {
				return fmt.Errorf("Build package %s during running failed, err: %s", path, err.Error())
			}
		}
		// test one suite each time
		for _, cases := range filesCases {
			tcNames := make([]string, len(cases))
			for i, tc := range cases {
				tcNames[i] = tc.Name
			}
			log.Printf("[PLUGIN]Run test cases: %v by bin file %s", tcNames, pkgBin)
			err = gotestRunner.RunTest(projPath, pkgBin, cases)
			if err != nil {
				log.Printf("[PLUGIN]Run test cases failed, err: %s", err.Error())
				continue
			}
		}
	}
	return nil
}

func parseTestcases(testSelectors []string) ([]*gotestTestcase.TestCase, error) {
	var testcases []*gotestTestcase.TestCase
	for _, selector := range testSelectors {
		testcase, err := gotestTestcase.ParseTestCaseBySelector(selector)
		if err != nil {
			log.Printf("[PLUGIN]parse testcase by selector [%s] failed, err: %s", selector, err.Error())
			continue
		}
		testcases = append(testcases, testcase)
	}
	if len(testcases) == 0 {
		return nil, errors.New("no available testcases")
	}
	return testcases, nil
}

func (o *ExecuteOptions) RunExecute(cmd *cobra.Command) error {
	// load case info from yaml file
	config, err := gotestTestcase.UnmarshalCaseInfo(o.executePath)
	if err != nil {
		return err
	}
	// parse testcases
	testcases, err := parseTestcases(config.TestSelectors)
	if err != nil {
		return err
	}
	// get workspace
	projPath := gotestUtil.GetWorkspace(config.ProjectPath)
	_, err = os.Stat(projPath)
	if err != nil {
		return fmt.Errorf("stat project path %s failed, err: %s", projPath, err.Error())
	}
	// get testcases grouped by path and name
	packages, err := groupTestCasesByPathAndName(projPath, testcases)
	if err != nil {
		return err
	}
	// run testcases
	return executeTestcases(projPath, packages)
}
