/*
Copyright © 2022 Martin Marsh martin@marshtrio.com

*/

package cmd

import (
	"fmt"
	"autohelm/helm"
	"autohelm/version"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "autohelm",
	Short: "Boat Helm driver controller",
	Long: `Boat Helm driver controller runs under TMUX so that as a background task it can
process inputs from a keypad to set desired course, control parameters etc.  

Listening on UDP ports the controller gets streams of data defining the compass heading and helm
position.

The controller writes to RPI hardware ports to control the power and direction of the
 helm motor using a motor contoller board and also to a buzzer via a buffer chip.

.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.navmux.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(runCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Autohelm",
	Long:  `Version Number of Autohelm - part of boat helm control system and stearing to compass heading`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Autohelm - Multiplexing data sources and data recording\nVersion: " + version.BuildVersion + "\nBuild: " + version.BuildTime)
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Autohelm starts processing",
	Long:  `Start Autohelm controlling helm motor and listening to keyboard,
	 UDP compass and UDP helm position data streams - runs until aborted`,
	Run: func(cmd *cobra.Command, args []string) {
		// to move up to start of oneline up use \033[F

		if len(args) > 0 {
			fmt.Printf("\nStarting Autohelm using %s\n\nruns until aborted\n", args[0])
		} else {
			fmt.Println("\nStarting Autohelm\nruns until aborted")
		}

		helm.Execute(loadConfig())

	},
}
