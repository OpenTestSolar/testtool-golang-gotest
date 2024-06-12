package result

import (
	"fmt"
	"log"

	sdkClient "github.com/OpenTestSolar/testtool-sdk-golang/client"
	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
)

func ReportTestResults(testResults chan *sdkModel.TestResult) error {
	reporter, err := sdkClient.NewReporterClient()
	if err != nil {
		fmt.Printf("[PLUGIN]Failed to create reporter: %v\n", err)
		return err
	}
	defer reporter.Close()
	for result := range testResults {
		log.Printf("[PLUGIN]Reporting test result: %s, result: %s", result.Test.Name, result.ResultType)
		err = reporter.ReportCaseResult(result)
		if err != nil {
			log.Printf("[PLUGIN]Failed to report load result: %v\n", err)
			return err
		}
	}
	return nil
}
