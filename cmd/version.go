package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SisyphusSQ/go-cli-starter/vars"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show starter version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("AppName:    %s\n", vars.AppName)
		fmt.Printf("AppVersion: %s\n", vars.AppVersion)
		fmt.Printf("GoVersion:  %s\n", vars.GoVersion)
		fmt.Printf("BuildTime:  %s\n", vars.BuildTime)
		fmt.Printf("GitCommit:  %s\n", vars.GitCommit)
		fmt.Printf("GitRemote:  %s\n", vars.GitRemote)
	},
}

func initVersion() {
	rootCmd.AddCommand(versionCmd)
}
