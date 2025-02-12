package runner

import (
	"io"
	"log"

	"github.com/OpenTestSolar/testtool-golang-gotest/pkg/cmdline"
	gotestResult "github.com/OpenTestSolar/testtool-golang-gotest/pkg/result"
	gotestUtil "github.com/OpenTestSolar/testtool-golang-gotest/pkg/util"
	"github.com/OpenTestSolar/testtool-sdk-golang/api"
	sdkClient "github.com/OpenTestSolar/testtool-sdk-golang/client"
	"github.com/pkg/errors"
)

type Runner interface {
	Run() (io.ReadCloser, io.ReadCloser, error)
}

type RawCmdlineRunner struct {
	cmdline  *cmdline.RawCmdline
	projPath string
}

func NewRawCmdlineRunner(cmdline *cmdline.RawCmdline, projPath string, reporter api.Reporter) Runner {
	return &RawCmdlineRunner{
		cmdline:  cmdline,
		projPath: projPath,
	}
}

func (o *RawCmdlineRunner) appendExtraParams() {
	log.Printf("[PLUGIN]Raw cmdline: %s", o.cmdline.GetCmdline())
	o.cmdline.AppendJsonParam()
	o.cmdline.AppendVParam()
	o.cmdline.AppendRedirect()
	o.cmdline.AppendCoverage()
	log.Printf("[PLUGIN]Raw cmdline after append extra params: %s", o.cmdline.GetCmdline())
}

func (o *RawCmdlineRunner) Run() (io.ReadCloser, io.ReadCloser, error) {
	o.appendExtraParams()
	stdout, stderr, err := gotestUtil.RunCommand(o.cmdline.GetCmdline(), o.projPath, false, true)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "run raw cmd %s failed", o.cmdline.GetCmdline())
	}
	return stdout, stderr, nil
}

type GoTestExecutor interface {
	Execute() error
}

type RawCmdlineExecutor struct {
	runner   Runner
	reporter api.Reporter
}

func NewRawCmdlineExecutor(rawCmdline, projPath, reportPath string) (GoTestExecutor, error) {
	reporter, err := sdkClient.NewReporterClient(reportPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create reporter when initializing raw cmdline executor")
	}
	return &RawCmdlineExecutor{
		runner:   NewRawCmdlineRunner(cmdline.NewRawCmdline(rawCmdline), projPath, reporter),
		reporter: reporter,
	}, nil
}

func (o *RawCmdlineExecutor) Execute() error {
	stdout, _, err := o.runner.Run()
	if err != nil {
		return errors.Wrap(err, "execute raw cmd by runner failed")
	}
	return gotestResult.ParseAndReportResult(stdout, "", o.reporter)
}
