package cmd

import "github.com/spf13/cobra"

func init() {
	RootCmd.AddCommand(startCmd)
}

var RootCmd = &cobra.Command{
	Use:     "drogo",
	Short:   "drogo: Eventackle Admin Panel server",
	Version: "1.0.0",
}
