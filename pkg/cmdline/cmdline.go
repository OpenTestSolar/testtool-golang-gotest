package cmdline

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	coverageFileName = "testsolar_gotest.gocov"
)

type RawCmdline struct {
	cmdline string
}

func NewRawCmdline(cmdline string) *RawCmdline {
	return &RawCmdline{
		cmdline: cmdline,
	}
}

func (o *RawCmdline) GetCmdline() string {
	return o.cmdline
}

func (o *RawCmdline) AppendVParam() {
	if !strings.Contains(o.cmdline, " -v ") {
		o.cmdline += " -v"
	}
}

func (o *RawCmdline) AppendJsonParam() {
	if !strings.Contains(o.cmdline, " -json ") {
		o.cmdline += " -json"
	}
}

func (o *RawCmdline) AppendRedirect() {
	if !strings.Contains(o.cmdline, "2>&1") {
		o.cmdline += " 2>&1 "
	}
}

func (o *RawCmdline) AppendCoverage() {
	caverageDir, err := os.UserCacheDir()
	if err != nil {
		log.Printf("[PLUGIN] Failed to get user cache dir: %s", err)
		return
	}
	coverageDir := filepath.Join(caverageDir, ".testsolar", "coverage")
	if err := os.MkdirAll(coverageDir, 0666); err != nil {
		log.Printf("[PLUGIN] Failed to create coverage dir: %s", err)
		return
	}
	o.cmdline += fmt.Sprintf(" -coverprofile=%s ", filepath.Join(coverageDir, coverageFileName))
}
