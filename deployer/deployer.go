package deployer

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

const stackName = "swalker-test"

type client interface {
	CreateStack(context.Context, *cloudformation.CreateStackInput, ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error)
	UpdateStack(context.Context, *cloudformation.UpdateStackInput, ...func(*cloudformation.Options)) (*cloudformation.UpdateStackOutput, error)
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

func (d *Deployer) update(ctx context.Context) error {
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

func (d *Deployer) Undeploy(ctx context.Context) error {
	return nil
}
