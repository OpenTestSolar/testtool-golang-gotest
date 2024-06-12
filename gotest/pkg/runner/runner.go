package runner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	gotestResult "gotest/pkg/result"
	gotestTestcase "gotest/pkg/testcase"
	gotestUtil "gotest/pkg/util"

	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/pkg/errors"
	"github.com/sourcegraph/conc"
)

func RunTest(projPath, pkgBin string, testcases []*gotestTestcase.TestCase) error {
	var tcNames []string
	for _, testcase := range testcases {
		tcNames = append(tcNames, testcase.Name)
	}
	packPath, err := filepath.Rel(projPath, pkgBin)
	if err != nil {
		return err
	}
	packPath = strings.TrimSuffix(packPath, ".test")
	cmdline := fmt.Sprintf(`go tool test2json -t -p %s %s -test.v=test2json -test.run "%s$"`, packPath, pkgBin, strings.Join(tcNames, "|"))
	extra_args := os.Getenv("TESTSOLAR_TTP_EXTRAARGS")
	if extra_args != "" {
		cmdline += " " + extra_args
	}
	cmdline += " 2>&1"
	log.Printf("[PLUGIN]Run cmdline %s", cmdline)
	stdout, _, err := gotestUtil.RunCommand(cmdline, projPath, false, true)
	if err != nil {
		return errors.Wrap(err, "run cmd failed")
	}
	testResults := make(chan *sdkModel.TestResult)
	var wg conc.WaitGroup
	wg.Go(
		func() {
			gotestResult.ParseCaseLog(stdout, testResults)
		},
	)
	wg.Go(
		func() {
			gotestResult.ReportTestResults(testResults)
		},
	)
	wg.Wait()
	return nil
}
