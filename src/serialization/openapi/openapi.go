package openapi

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type Document struct {
	Components Components      `yaml:"components"`
	Paths      map[string]Path `yaml:"paths"`

	AbsolutePath string
}

func LoadDocument(path string) (*Document, error) {
	document, err := loadDocumentNoResolve(path)
	if err != nil {
		return nil, err
	}

	return document, document.ResolveRefs()
}

func loadDocumentNoResolve(path string) (*Document, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	document := Document{}
	document.AbsolutePath = path
	err = yaml.Unmarshal(content, &document)

	// Replace all "local" refs with absolute ones

	for _, schema := range document.Components.Schemas {
		if strings.HasPrefix(schema.Ref, "#") {
			schema.Ref = document.AbsolutePath + schema.Ref
		}
	}
	for _, response := range document.Components.Responses {
		if strings.HasPrefix(response.Ref, "#") {
			response.Ref = document.AbsolutePath + response.Ref
		}
	}
	for _, parameter := range document.Components.Parameters {
		if strings.HasPrefix(parameter.Ref, "#") {
			parameter.Ref = document.AbsolutePath + parameter.Ref
		}
	}

	for _, path := range document.Paths {
		for _, parameter := range path.Parameters {
			if strings.HasPrefix(parameter.Ref, "#") {
				parameter.Ref = document.AbsolutePath + parameter.Ref
			}
		}

		for _, op := range path.Operations {
			for _, response := range op.Responses {
				if strings.HasPrefix(response.Ref, "#") {
					response.Ref = document.AbsolutePath + response.Ref
				}

				for _, mediaType := range response.Content {
					if strings.HasPrefix(mediaType.Schema.Ref, "#") {
						mediaType.Schema.Ref = document.AbsolutePath + mediaType.Schema.Ref
					}
				}
			}
		}
	}

	return &document, err
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
