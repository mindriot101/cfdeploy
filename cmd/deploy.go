package cmd

import (
	"context"
	"log"

	"github.com/mindriot101/cfdeploy/deployer"
	"github.com/spf13/cobra"
)

var capabilities []string

func init() {
	deployCmd.Flags().StringSliceVarP(&capabilities, "cap", "c", nil, "Capabilities to add")
	rootCmd.AddCommand(deployCmd)
}

func run(cmd *cobra.Command, args []string) {
	log.Println("deploying")
	ctx := context.Background()
	// tpl, err := cf.Parse(templateName)
	// if err != nil {
	// 	log.Printf("error loading template: %v", err)
	// 	return
	// }
	// _ = tpl
	d, err := deployer.New(ctx, templateName)
	if err != nil {
		log.Printf("error setting up deployer: %v", err)
		return
	}
	if err := d.Deploy(ctx, capabilities); err != nil {
		log.Printf("could not deploy template: %v", err)
		return
	}
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy or update a template",
	Run:   run,
}
