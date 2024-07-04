package loader

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	gotestTestcase "gotest/pkg/testcase"

	"github.com/pkg/errors"
)

func ParseTestCaseInFile(projPath string, path string) ([]*gotestTestcase.TestCase, error) {
	if !strings.HasSuffix(path, "_test.go") {
		return nil, nil
	}
	projPath = strings.TrimSuffix(projPath, string(os.PathSeparator))
	selectorPath, err := filepath.Rel(projPath, path)
	if err != nil {
		return nil, errors.Wrapf(err, "get relative path of %s failed, project path: %s", path, projPath)
	}
	var testcaseList []*gotestTestcase.TestCase
	log.Printf("[PLUGIN]Parse testcase in file %s", path)
	code, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "read testcase file %s failed", path)
	}
	for _, line := range strings.Split(strings.TrimSuffix(string(code), "\n"), "\n") {
		re := regexp.MustCompile(`^func\s+(Test\w+)\s*\(t \*testing\.T\)`)
		match := re.FindStringSubmatch(line)
		if len(match) > 0 {
			log.Printf("find case %s", match[1])
			testcaseList = append(testcaseList, &gotestTestcase.TestCase{
				Path:       selectorPath,
				Name:       match[1],
				Attributes: map[string]string{},
			})
		}
	}
	log.Printf("[PLUGIN]%d testcases found in file %s", len(testcaseList), path)
	return testcaseList, nil
}

func LoadTestCase(projPath string, selectorPath string) ([]*gotestTestcase.TestCase, error) {
	var testcaseList []*gotestTestcase.TestCase
	selectorAbsPath := filepath.Join(projPath, selectorPath)
	fi, err := os.Stat(selectorAbsPath)
	if err != nil {
		return nil, errors.Wrapf(err, "stat selector abs path: %s failed", selectorAbsPath)
	}
	log.Printf("[PLUGIN]Try to load testcases from path %s", selectorAbsPath)
	if fi.IsDir() {
		err := filepath.Walk(selectorAbsPath, func(path string, fi os.FileInfo, e error) error {
			if e != nil {
				return errors.Wrapf(e, "failed to walk file path %s", path)
			}
			loadedTestCases, err := ParseTestCaseInFile(projPath, path)
			if err != nil {
				return errors.Wrapf(err, "parse testcase in path %s failed", path)
			}
			testcaseList = append(testcaseList, loadedTestCases...)
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		loadedTestCases, err := ParseTestCaseInFile(projPath, selectorAbsPath)
		if err != nil {
			return nil, errors.Wrapf(err, "parse testcase in file %s failed", selectorAbsPath)
		}
		testcaseList = append(testcaseList, loadedTestCases...)
	}
	return testcaseList, nil
}
