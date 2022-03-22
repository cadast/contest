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
	for _, schema := range document.Components.Schemas {
		err := schema.resolveRef(document)
		if err != nil {
			return err
		}
	}

	for _, path := range document.Paths {
		for _, parameter := range path.Parameters {
			err := parameter.resolveRef(document)
			if err != nil {
				return err
			}
		}

		for _, op := range path.Operations {
			for _, response := range op.Responses {
				err := response.resolveRef(document)
				if err != nil {
					return err
				}

				for _, content := range response.Content {
					err = content.Schema.resolveRef(document)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// FindOperationById gets the operation with the given id from a document.
// It returns the URL, method, Operation and if an operation was found.
func (document Document) FindOperationById(id string) (string, string, *Operation, bool) {
	for url, path := range document.Paths {
		for method, operation := range path.Operations {
			if operation.OperationId == id {
				return url, method, &operation, true
			}
		}
	}
	return "", "", nil, false
}

func (s *Schema) resolveRef(document Document) error {
	// Resolve ref in self
	if s.Ref != "" {
		if !strings.HasPrefix(s.Ref, "#/components/schemas/") {
			return fmt.Errorf("unknown $ref format: %s", s.Ref)
		}
		ref := strings.TrimPrefix(s.Ref, "#/components/schemas/")
		if _, ok := document.Components.Schemas[ref]; !ok {
			return fmt.Errorf("could not resolve $ref %s", s.Ref)
		}
		resolved := document.Components.Schemas[ref]
		*s = *resolved
	}

	// Recursively resolve refs in properties
	for _, property := range s.Properties {
		err := property.resolveRef(document)
		if err != nil {
			return err
		}
	}

	// Recursively resolve refs in items
	if s.Items != nil {
		err := s.Items.resolveRef(document)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Parameter) resolveRef(document Document) error {
	if p.Ref != "" {
		if !strings.HasPrefix(p.Ref, "#/components/parameters/") {
			return fmt.Errorf("unknown $ref format: %s", p.Ref)
		}
		ref := strings.TrimPrefix(p.Ref, "#/components/parameters/")
		if _, ok := document.Components.Parameters[ref]; !ok {
			return fmt.Errorf("could not resolve $ref %s", p.Ref)
		}
		resolved := document.Components.Parameters[ref]
		*p = resolved
	}

	err := p.Schema.resolveRef(document)
	if err != nil {
		return err
	}

	return nil
}
