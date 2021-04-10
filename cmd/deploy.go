package cmd

import (
	"context"
	"log"

	"github.com/mindriot101/cfdeploy/deployer"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy or update a template",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("deploying")
		ctx := context.Background()
		d, err := deployer.New(ctx, templateName)
		if err != nil {
			log.Printf("error setting up deployer: %v", err)
			return
		}
		if err := d.Deploy(ctx); err != nil {
			log.Printf("could not deploy template: %v", err)
			return
		}
	},
}
