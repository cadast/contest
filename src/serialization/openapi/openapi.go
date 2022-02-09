package openapi

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

type Document struct {
	Components Components      `yaml:"components"`
	Paths      map[string]Path `yaml:"paths"`
}

func LoadDocument(path string) (*Document, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	document := Document{}
	err = yaml.Unmarshal(content, &document)
	if err != nil {
		return nil, err
	}

	return &document, document.ResolveRefs()
}

func (document Document) ResolveRefs() error {
	for name, schema := range document.Components.Schemas {
		resolved, err := resolvePropertyRefs(&schema, document)
		if err != nil {
			return err
		}

		document.Components.Schemas[name] = *resolved
	}

	for _, path := range document.Paths {
		for _, op := range path.Operations {
			for _, response := range op.Responses {
				for _, content := range response.Content {
					resolved, err := resolvePropertyRefs(content.Schema, document)
					if err != nil {
						return err
					}
					content.Schema = resolved
				}
			}
		}
	}
	return nil
}

func resolvePropertyRefs(schema *Schema, doc Document) (*Schema, error) {
	schema, err := resolveRef(schema, doc)
	if err != nil {
		return nil, err
	}

	for name, property := range schema.Properties {
		resolved, err := resolvePropertyRefs(&property, doc)
		if err != nil {
			return nil, err
		}

		schema.Properties[name] = *resolved
	}
	if schema.Items != nil {
		resolved, err := resolvePropertyRefs(schema.Items, doc)
		if err != nil {
			return nil, err
		}

		schema.Items = resolved
	}
	return schema, nil
}

func resolveRef(schema *Schema, doc Document) (*Schema, error) {
	if schema.Ref != "" {
		if !strings.HasPrefix(schema.Ref, "#/components/schemas/") {
			return nil, fmt.Errorf("unknown $ref format: %s", schema.Ref)
		}
		ref := strings.TrimPrefix(schema.Ref, "#/components/schemas/")
		if _, ok := doc.Components.Schemas[ref]; !ok {
			return nil, fmt.Errorf("could not resolve $ref %s", schema.Ref)
		}
		resolved := doc.Components.Schemas[ref]
		return &resolved, nil
	}
	return schema, nil
}
