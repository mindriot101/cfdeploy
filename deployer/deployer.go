package deployer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"
)

const stackName = "swalker-test"

type client interface {
	CreateStack(context.Context, *cloudformation.CreateStackInput, ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error)
	UpdateStack(context.Context, *cloudformation.UpdateStackInput, ...func(*cloudformation.Options)) (*cloudformation.UpdateStackOutput, error)
	DeleteStack(context.Context, *cloudformation.DeleteStackInput, ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error)
	DescribeStacks(context.Context, *cloudformation.DescribeStacksInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
}

type Deployer struct {
	cf  client
	tpl string
}

func New(c client, tpl string) *Deployer {
	return &Deployer{
		cf:  c,
		tpl: tpl,
	}
}

func (d *Deployer) Deploy(ctx context.Context) error {
	// Try to create the stack
	res, err := d.cf.CreateStack(ctx, &cloudformation.CreateStackInput{
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
	d.waitForStack(ctx, *res.StackId)
	return nil
}

func (d *Deployer) update(ctx context.Context) error {
	log.Printf("updating stack")
	res, err := d.cf.UpdateStack(ctx, &cloudformation.UpdateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(d.tpl),
	})
	if err != nil {
		var ge *smithy.GenericAPIError
		if errors.As(err, &ge) {
			if ge.Message == "No updates are to be performed." {
				return nil
			}
		}
		return fmt.Errorf("error updating stack: %w", err)
	}
	d.waitForStack(ctx, *res.StackId)
	return nil
}

func (d *Deployer) waitForStack(ctx context.Context, stackId string) error {
	terminalStates := []types.StackStatus{
		types.StackStatusCreateFailed,
		types.StackStatusCreateComplete,
		types.StackStatusRollbackFailed,
		types.StackStatusRollbackComplete,
		types.StackStatusDeleteFailed,
		types.StackStatusDeleteComplete,
		types.StackStatusUpdateComplete,
		types.StackStatusUpdateRollbackFailed,
		types.StackStatusUpdateRollbackComplete,
		types.StackStatusImportComplete,
		types.StackStatusImportRollbackFailed,
		types.StackStatusImportRollbackComplete,
	}
	failureStates := []types.StackStatus{
		types.StackStatusCreateFailed,
		types.StackStatusRollbackFailed,
		types.StackStatusDeleteFailed,
		types.StackStatusUpdateRollbackFailed,
		types.StackStatusImportRollbackFailed,
	}

	for {
		res, err := d.cf.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		})
		if err != nil {
			return fmt.Errorf("error fetching stack status: %w", err)
		}
		if len(res.Stacks) != 1 {
			return fmt.Errorf("unexpected number of stacks: %d", len(res.Stacks))
		}
		ss := res.Stacks[0].StackStatus
		log.Printf("stack status: %s", ss)
		for _, st := range terminalStates {
			if st == ss {
				for _, st2 := range failureStates {
					if st2 == ss {
						return fmt.Errorf("stack failure: %v", ss)
					}
				}
				return nil
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (d *Deployer) Undeploy(ctx context.Context) error {
	log.Printf("deleting stack")
	_, err := d.cf.DeleteStack(ctx, &cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	})
	return err
}
