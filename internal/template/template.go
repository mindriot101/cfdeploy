package template

import (
	"fmt"
	"io/ioutil"
)

type Template struct {
	path string
}

func New(path string) *Template {
	return &Template{
		path: path,
	}
}

func (t *Template) String() (string, error) {
	s, err := ioutil.ReadFile(t.path)
	if err != nil {
		return "", fmt.Errorf("error reading template: %w", err)
	}
	return string(s), nil
}
