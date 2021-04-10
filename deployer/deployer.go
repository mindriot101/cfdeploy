package deployer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"
	"github.com/mindriot101/cfdeploy/internal/template"
)

const StackName = "swalker-test"

type client interface {
	CreateStack(context.Context, *cloudformation.CreateStackInput, ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error)
	UpdateStack(context.Context, *cloudformation.UpdateStackInput, ...func(*cloudformation.Options)) (*cloudformation.UpdateStackOutput, error)
	DeleteStack(context.Context, *cloudformation.DeleteStackInput, ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error)
	DescribeStacks(context.Context, *cloudformation.DescribeStacksInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
}

type Templater interface {
	String() (string, error)
}

type Deployer struct {
	cf  client
	tpl string
}

func (d *Deployer) Deploy(ctx context.Context, capabilities []string) error {
	// Try to create the stack
	var caps []types.Capability
	for _, c := range capabilities {
		switch c {
		case "named_iam":
			caps = append(caps, types.CapabilityCapabilityNamedIam)
		}
	}
	res, err := d.cf.CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:    aws.String(StackName),
		TemplateBody: aws.String(d.tpl),
		Capabilities: caps,
	})
	if err != nil {
		// why can we not use Is here?
		var aex *types.AlreadyExistsException
		if errors.As(err, &aex) {
			return d.update(ctx, caps)
		}
		return fmt.Errorf("error creating stack: %w", err)
	}
	d.waitForStack(ctx, *res.StackId)
	return nil
}

func (d *Deployer) update(ctx context.Context, caps []types.Capability) error {
	log.Printf("updating stack")
	res, err := d.cf.UpdateStack(ctx, &cloudformation.UpdateStackInput{
		StackName:    aws.String(StackName),
		TemplateBody: aws.String(d.tpl),
		Capabilities: caps,
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
			StackName: aws.String(StackName),
		})
		if err != nil {
			return fmt.Errorf("error fetching stack status: %w", err)
		}
		if len(res.Stacks) != 1 {
			return fmt.Errorf("unexpected number of stacks: %d", len(res.Stacks))
		}
		ss := res.Stacks[0].StackStatus
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
		StackName: aws.String(StackName),
	})
	return err
}

func newDeployer(ctx context.Context, c client, tpl Templater) (*Deployer, error) {
	t, err := tpl.String()
	if err != nil {
		return nil, fmt.Errorf("error reading template: %w", err)
	}
	d := &Deployer{
		cf:  c,
		tpl: t,
	}
	return d, nil
}

func New(ctx context.Context, templateFilename string) (*Deployer, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	cf := cloudformation.NewFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error loading AWS config: %v", err)
	}
	return newDeployer(ctx, cf, template.New(templateFilename))
}
