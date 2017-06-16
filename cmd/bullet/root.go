package main

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "bullet",
	Short: "Bullet is a fast deploy tool",
	Long:  `Bullet is a fast and flexible deploy tool built by Furqan Software and friends. Complete documentation is available at https://bullettool.com/.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}
