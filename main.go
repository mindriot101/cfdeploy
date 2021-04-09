package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

const stackName = "swalker-test"

func isFile(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !s.IsDir()
}

func templateFilename() (string, error) {
	fnames := []string{
		"cloudformation.yml",
		"cloudformation.template",
	}
	for _, fname := range fnames {
		if isFile(fname) {
			return fname, nil
		}
	}

	return "", fmt.Errorf("cannot find template file")
}

func usage() {
	fmt.Printf("usage: %s <deploy|undeploy>\n", os.Args[0])
	os.Exit(1)
}

type deployer struct {
	cf  *cloudformation.Client
	tpl string
}

func (d *deployer) deploy(ctx context.Context) error {
	// Try to create the stack
	_, err := d.cf.CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(d.tpl),
	})
	if err != nil {
		// why can we not use Is here?
		var aex *types.AlreadyExistsException
		if errors.As(err, &aex) {
			return d.update(ctx)
		}
		return fmt.Errorf("error creating stack: %w", err)
	}
	return nil
}

func (d *deployer) update(ctx context.Context) error {
	log.Printf("updating stack")
	_, err := d.cf.UpdateStack(ctx, &cloudformation.UpdateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(d.tpl),
	})
	if err != nil {
		log.Fatalf("error updating stack: %v", err)
	}
	return nil
}

func (d *deployer) undeploy(ctx context.Context) error {
	return nil
}

func main() {
	if len(os.Args) != 2 {
		usage()
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	cf := cloudformation.NewFromConfig(cfg)
	if err != nil {
		log.Fatalf("error loading AWS config: %v", err)
	}
	tplF, err := templateFilename()
	if err != nil {
		log.Fatalf("cannot find template file: %v", err)
	}
	tplB, err := ioutil.ReadFile(tplF)
	if err != nil {
		log.Fatalf("cannot read template file: %v", err)
	}
	tpl := string(tplB)
	deployer := &deployer{
		cf:  cf,
		tpl: tpl,
	}

	switch os.Args[1] {
	case "deploy":
		err = deployer.deploy(ctx)
	case "undeploy":
		err = deployer.undeploy(ctx)
	default:
		usage()
	}

	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
