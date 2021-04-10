package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/mindriot101/cfdeploy/deployer"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(undeployCmd)
}

var undeployCmd = &cobra.Command{
	Use:   "undeploy",
	Short: "Tear down a stack",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("undeploying")
		ctx := context.Background()
		d, err := deployer.New(ctx, templateName)
		if err != nil {
			log.Printf("error setting up deployer: %v", err)
			return
		}
		if err := d.Undeploy(ctx); err != nil {
			log.Printf("could not undeploy template: %v", err)
			return
		}
	},
}
