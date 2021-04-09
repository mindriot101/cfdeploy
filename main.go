package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/mindriot101/cfdeploy/deployer"
)

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
	deployer := deployer.New(cf, tpl)

	switch os.Args[1] {
	case "deploy":
		err = deployer.Deploy(ctx)
	case "undeploy":
		err = deployer.Undeploy(ctx)
	default:
		usage()
	}

	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
