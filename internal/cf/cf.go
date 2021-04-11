package cf

import (
	"fmt"

	"github.com/awslabs/goformation/v4"
	"github.com/awslabs/goformation/v4/cloudformation"
)

type Template struct {
	tpl *cloudformation.Template
}

func Parse(path string) (*Template, error) {
	raw, err := goformation.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening template: %w", err)
	}
	return &Template{tpl: raw}, nil
}
