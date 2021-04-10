package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var templateName string

var rootCmd = &cobra.Command{
	Use:   "cfdeploy",
	Short: "Deploy cloudformation in a friendly way",
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&templateName, "template", "t", "", "Cloudformation template")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
