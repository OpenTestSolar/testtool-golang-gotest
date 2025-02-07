package cmdline

import "strings"

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
