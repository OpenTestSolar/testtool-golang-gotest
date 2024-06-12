package main

import (
	"gotest/cmd/build"
	"gotest/cmd/discover"
	"gotest/cmd/execute"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := cobra.Command{
		Use: "solar-gotest",
	}
	rootCmd.AddCommand(discover.NewCmdDiscover())
	rootCmd.AddCommand(execute.NewCmdExecute())
	rootCmd.AddCommand(build.NewCmdBuild())
	rootCmd.Execute()
}
