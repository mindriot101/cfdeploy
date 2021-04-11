package deployer

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

const tpl = "template"

type createFnSig func(ctx context.Context, ip *cloudformation.CreateStackInput, opts ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error)
type updateFnSig func(ctx context.Context, ip *cloudformation.UpdateStackInput, opts ...func(*cloudformation.Options)) (*cloudformation.UpdateStackOutput, error)
type deleteFnSig func(ctx context.Context, ip *cloudformation.DeleteStackInput, opts ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error)
type describeFnSig func(ctx context.Context, ip *cloudformation.DescribeStacksInput, opts ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)

type call struct {
	name    string
	payload interface{}
}

type mockClient struct {
	calls      []call
	createfn   createFnSig
	updatefn   updateFnSig
	deletefn   deleteFnSig
	describefn describeFnSig
}

type mockTemplater struct {
	calls    []call
	stringfn func() (string, error)
}

func (t *mockTemplater) String() (string, error) {
	t.calls = append(t.calls, call{
		name: "String",
	})
	if t.stringfn == nil {
		return "template", nil
	}
	return t.stringfn()
}

func (m *mockClient) CreateStack(ctx context.Context, ip *cloudformation.CreateStackInput, opts ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
	m.calls = append(m.calls, call{
		name:    "CreateStack",
		payload: ip,
	})
	if m.createfn == nil {
		return &cloudformation.CreateStackOutput{
			StackId: aws.String("stack"),
		}, nil
	}
	return m.createfn(ctx, ip, opts...)
}
func (m *mockClient) UpdateStack(ctx context.Context, ip *cloudformation.UpdateStackInput, opts ...func(*cloudformation.Options)) (*cloudformation.UpdateStackOutput, error) {
	m.calls = append(m.calls, call{
		name:    "UpdateStack",
		payload: ip,
	})
	if m.updatefn == nil {
		return nil, nil
	}
	return m.updatefn(ctx, ip, opts...)
}
func (m *mockClient) DeleteStack(ctx context.Context, ip *cloudformation.DeleteStackInput, opts ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error) {
	m.calls = append(m.calls, call{
		name:    "DeleteStack",
		payload: ip,
	})
	if m.deletefn == nil {
		return nil, nil
	}
	return m.deletefn(ctx, ip, opts...)
}
func (m *mockClient) DescribeStacks(ctx context.Context, ip *cloudformation.DescribeStacksInput, opts ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	m.calls = append(m.calls, call{
		name:    "DescribeStacks",
		payload: ip,
	})
	if m.describefn == nil {
		return &cloudformation.DescribeStacksOutput{
			Stacks: []types.Stack{
				{
					StackStatus: types.StackStatusCreateComplete,
				},
			},
		}, nil
	}
	return m.describefn(ctx, ip, opts...)
}

func TestCreateStack(t *testing.T) {
	client := &mockClient{}
	tpl := &mockTemplater{}
	d, err := newDeployer(context.TODO(), client, tpl)
	if err != nil {
		t.Errorf("error creating deployer: %v", err)
	}

	d.Deploy(context.TODO(), []string{})

	c := client.calls[0]
	if c.name != "CreateStack" {
		t.Errorf("unexpected method name: %s", c.name)
	}

	cp := c.payload.(*cloudformation.CreateStackInput)
	if *cp.StackName != "swalker-test" {
		t.Errorf("unexpected stack name: %s != swalker-test", *cp.StackName)
	}
	if *cp.TemplateBody != "template" {
		t.Errorf("unexpected template body: %s", *cp.TemplateBody)
	}
	if len(cp.Capabilities) != 0 {
		t.Errorf("unexpected capabilities: %v", cp.Capabilities)
	}
}

func TestFallBackToUpdate(t *testing.T) {
	client := &mockClient{}
	client.createfn = func(ctx context.Context, ip *cloudformation.CreateStackInput, opts ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
		return nil, &types.AlreadyExistsException{}
	}

	tpl := &mockTemplater{}
	d, err := newDeployer(context.TODO(), client, tpl)
	if err != nil {
		t.Errorf("error creating deployer: %v", err)
	}

	d.Deploy(context.TODO(), []string{})

	c := client.calls[0]
	expected := call{
		name: "CreateStack",
		payload: &cloudformation.CreateStackInput{
			StackName:    aws.String("swalker-test"),
			TemplateBody: aws.String("template"),
		},
	}
	if !reflect.DeepEqual(c, expected) {
		t.Errorf("%+#v != %+#v", c, expected)
	}
	c = client.calls[1]
	expected = call{
		name: "UpdateStack",
		payload: &cloudformation.UpdateStackInput{
			StackName:    aws.String("swalker-test"),
			TemplateBody: aws.String("template"),
		},
	}
	if !reflect.DeepEqual(c, expected) {
		t.Errorf("%+#v != %+#v", c, expected)
	}
}
