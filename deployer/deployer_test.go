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
type call struct {
	name    string
	payload interface{}
}

type mockClient struct {
	calls    []call
	createfn createFnSig
	updatefn updateFnSig
}

func (m *mockClient) CreateStack(ctx context.Context, ip *cloudformation.CreateStackInput, opts ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
	m.calls = append(m.calls, call{
		name:    "CreateStack",
		payload: ip,
	})
	if m.createfn == nil {
		return nil, nil
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

func TestCreateStack(t *testing.T) {
	client := &mockClient{}
	d := New(client, tpl)

	d.Deploy(context.TODO())

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
}

func TestFallBackToUpdate(t *testing.T) {
	client := &mockClient{}
	client.createfn = func(ctx context.Context, ip *cloudformation.CreateStackInput, opts ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
		return nil, &types.AlreadyExistsException{}
	}

	d := New(client, tpl)

	d.Deploy(context.TODO())

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
