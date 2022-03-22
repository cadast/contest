package openapi

import (
	"fmt"
	"strings"
)

type Path struct {
	Summary     string               `yaml:"summary"`
	Description string               `yaml:"description"`
	Parameters  []*Parameter         `yaml:"parameters"`
	Operations  map[string]Operation `yaml:",inline"`
}

type Operation struct {
	Summary     string               `yaml:"summary"`
	OperationId string               `yaml:"operationId"`
	Description string               `yaml:"description"`
	Tags        []string             `yaml:"tags"`
	Responses   map[string]*Response `yaml:"responses"`
}

type Response struct {
	Description string               `yaml:"description"`
	Content     map[string]MediaType `yaml:"content"`

	Ref string `yaml:"$ref"`
}

type MediaType struct {
	Schema *Schema `yaml:"schema"`
}

func (r *Response) resolveRef(document Document) error {
	if r.Ref != "" {
		if !strings.HasPrefix(r.Ref, "#/components/responses/") {
			return fmt.Errorf("unknown $ref format: %s", r.Ref)
		}
		ref := strings.TrimPrefix(r.Ref, "#/components/responses/")
		if _, ok := document.Components.Responses[ref]; !ok {
			return fmt.Errorf("could not resolve $ref %s", r.Ref)
		}
		resolved := document.Components.Responses[ref]
		*r = resolved
	}

	return nil
}
