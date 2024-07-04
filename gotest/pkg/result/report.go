package result

import (
	"log"

	"github.com/OpenTestSolar/testtool-sdk-golang/api"
	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/pkg/errors"
)

func ReportTestResults(testResults chan *sdkModel.TestResult, reporter api.Reporter) error {
	for result := range testResults {
		log.Printf("[PLUGIN]Reporting test result: %s, result: %d", result.Test.Name, result.ResultType)
		err := reporter.ReportCaseResult(result)
		if err != nil {
			return errors.Wrapf(err, "Report test result failed")
		}
	}
	log.Printf("[PLUGIN]Report test results finished")
	return nil
}
