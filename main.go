package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/google/uuid"
	"github.com/mindriot101/cfdeploy/cmd"
	"github.com/mindriot101/cfdeploy/deployer"
)

func isFile(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !s.IsDir()
}

func usage() {
	fmt.Printf("usage: %s <deploy|undeploy>\n", os.Args[0])
	os.Exit(1)
}

func wait(ctx context.Context, cf *cloudformation.Client, res interface{}) error {
	if cs, ok := res.(*cloudformation.CreateChangeSetOutput); ok {
		terminalStates := []types.ChangeSetStatus{
			types.ChangeSetStatusCreateComplete,
			types.ChangeSetStatusDeleteComplete,
			types.ChangeSetStatusDeleteFailed,
			types.ChangeSetStatusFailed,
		}
		errorStates := []types.ChangeSetStatus{
			types.ChangeSetStatusDeleteFailed,
			types.ChangeSetStatusFailed,
		}
		for {
			res, err := cf.DescribeChangeSet(ctx, &cloudformation.DescribeChangeSetInput{
				ChangeSetName: cs.Id,
			})
			if err != nil {
				return fmt.Errorf("error waiting for change set: %w", err)
			}
			for _, termState := range terminalStates {
				if res.Status == termState {
					for _, eState := range errorStates {
						if res.Status == eState {
							return fmt.Errorf("change set creation failed, status: %s", eState)
						}
					}
					return nil
				}
			}
			time.Sleep(5 * time.Second)
		}
	} else {
		return fmt.Errorf("do not now how to wait for %v", res)
	}
}

func testChangeSet(cf *cloudformation.Client, tpl string) {
	ctx := context.Background()
	log.Println("testing change set")
	csName := fmt.Sprintf("cs-%s", uuid.New())
	cs, err := cf.CreateChangeSet(ctx, &cloudformation.CreateChangeSetInput{
		ChangeSetName: aws.String(csName),
		StackName:     aws.String(deployer.StackName),
		ChangeSetType: types.ChangeSetTypeUpdate,
		Description:   aws.String("test chnage set"),
		TemplateBody:  aws.String(tpl),
	})
	if err != nil {
		log.Printf("error creating change set: %v", err)
		return
	}
	defer func() {
		log.Printf("deleting change set")
		// delete the change set
		_, err := cf.DeleteChangeSet(context.TODO(), &cloudformation.DeleteChangeSetInput{
			ChangeSetName: cs.Id,
		})
		if err != nil {
			log.Printf("error deleting change set: %v", err)
			return
		}
	}()
	if err := wait(ctx, cf, cs); err != nil {
		log.Printf("could not wait for change set: %v", err)
		return
	}

	res, err := cf.DescribeChangeSet(context.TODO(), &cloudformation.DescribeChangeSetInput{
		ChangeSetName: cs.Id,
	})
	if err != nil {
		log.Printf("describing change set: %v", err)
		return
	}
	presentChangeSet(res, os.Stdout)
}

func presentChangeSet(res *cloudformation.DescribeChangeSetOutput, out io.Writer) error {
	fmt.Fprintf(out, "CHANGES:\n")
	for _, ch := range res.Changes {
		ch := ch.ResourceChange
		fmt.Fprintf(out, "Change type: %s\n", ch.Action)
		fmt.Fprintf(out, "\tType: %s\n", *ch.ResourceType)
		fmt.Fprintf(out, "\tName: %s\n", *ch.LogicalResourceId)
		if ch.Action == types.ChangeActionModify {
			for _, detail := range ch.Details {
				_ = detail
			}
		}
	}
	return nil
}

func main() {
	cmd.Execute()
}

/*
func main2() {
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
*/
