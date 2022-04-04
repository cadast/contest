package openapi

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
)

func (document Document) ResolveRefs() error {
	for _, schema := range document.Components.Schemas {
		err := schema.resolveRef(document.AbsolutePath)
		if err != nil {
			return err
		}
	}

	for _, path := range document.Paths {
		for _, parameter := range path.Parameters {
			err := parameter.resolveRef(document.AbsolutePath)
			if err != nil {
				return err
			}
		}

		for _, op := range path.Operations {
			for _, response := range op.Responses {
				err := response.resolveRef(document.AbsolutePath)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// resolveRelativePath takes an anchor and a path relative to that anchor and returns the resulting absolute path.
func resolveRelativePath(anchor string, path string) (string, error) {
	if strings.HasPrefix(path, "/") {
		return path, nil
	}

	anchor, err := filepath.Abs(anchor)
	if err != nil {
		return "", err
	}
	anchor = filepath.Dir(anchor)
	return filepath.Clean(filepath.Join(anchor, path)), nil
}

// resolveRef resolves the reference in a Schema and the references in the Schema's Properties and Items.
func (s *Schema) resolveRef(currentPath string) error {
	if s.Ref != "" {
		schema := &Schema{}
		var err error
		var fragment string

		currentPath, fragment, err = getAbsoluteFileFragment(currentPath, s.Ref)
		if err = resolveReference(currentPath, fragment, schema); err != nil {
			return err
		}
		if err != nil {
			return err
		}

		if schema != nil {
			*s = *schema
		}
	}

	// Recursively resolve refs in properties
	for _, property := range s.Properties {
		err := property.resolveRef(currentPath)
		if err != nil {
			return err
		}
	}

	// Recursively resolve refs in items
	if s.Items != nil {
		err := s.Items.resolveRef(currentPath)
		if err != nil {
			return err
		}
	}

	return nil
}

// resolveRef resolves the reference in a Parameter and the reference in the Parameter's Schema.
func (p *Parameter) resolveRef(currentPath string) error {
	if p.Ref != "" {
		parameter := &Parameter{}
		var err error
		var fragment string

		currentPath, fragment, err = getAbsoluteFileFragment(currentPath, p.Ref)
		if err = resolveReference(currentPath, fragment, parameter); err != nil {
			return err
		}
		if err != nil {
			return err
		}

		*p = *parameter
	}

	err := p.Schema.resolveRef(currentPath)
	return err
}

// resolveRef resolves the reference in a Response and the references in the Schema in the MediaType.
func (r *Response) resolveRef(currentPath string) error {
	if r.Ref != "" {
		response := &Response{}
		var err error
		var fragment string

		currentPath, fragment, err = getAbsoluteFileFragment(currentPath, r.Ref)
		if err = resolveReference(currentPath, fragment, response); err != nil {
			return err
		}
		if err != nil {
			return err
		}

		*r = *response
	}

	for _, mediaType := range r.Content {
		if err := mediaType.Schema.resolveRef(currentPath); err != nil {
			return err
		}
	}

	return nil
}

// getAbsoluteFileFragment takes a basePath and path and returns an absolute file path and a fragment.
//
// path is assumed to be relative to basePath.
// fragment is taken from path, if present, otherwise "/" is returned
func getAbsoluteFileFragment(basePath string, path string) (string, string, error) {
	var file, fragment string
	if strings.HasPrefix(path, "#") {
		// If the path is only a fragment, prepend basePath
		file = basePath
		fragment = strings.TrimPrefix(path, "#")
	} else {
		// Resolve the path relative to basePath
		parts := strings.Split(path, "#")
		if len(parts) == 1 {
			fragment = "/"
		} else if len(parts) == 2 {
			fragment = parts[1]
		} else if len(parts) > 2 {
			return "", "", fmt.Errorf("could not resolve $ref %s: more than one # character", path)
		}

		var err error
		file, err = resolveRelativePath(basePath, parts[0])
		if err != nil {
			return "", "", err
		}
	}
	return file, fragment, nil
}

// resolveReference finds the yaml object in a file at the location the fragment points to and marshals it to out.
func resolveReference(filename string, fragment string, out interface{}) error {
	var content []byte
	var err error
	if content, err = ioutil.ReadFile(filename); err != nil {
		return err
	}

	// Unmarshal file into a map
	m := make(map[string]interface{})
	if err = yaml.Unmarshal(content, m); err != nil {
		return err
	}

	// Find the object at the fragment location
	var resolved interface{}
	if resolved, err = resolveInternalReference(m, fragment); err != nil {
		return err
	}

	// Marshal the object we found
	marshalled, err := yaml.Marshal(resolved)
	if err != nil {
		return err
	}

	// Unmarshal found object again, this time into the correct interface{}
	return yaml.Unmarshal(marshalled, out)
}

// resolveInternalReference finds the object that is within the map m at the given path. The path uses / as separator.
func resolveInternalReference(m interface{}, path string) (interface{}, error) {
	// If the path is empty or just "/", we need to recurse further
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return m, nil
	}

	v := reflect.ValueOf(m)
	switch v.Kind() {
	case reflect.Map:
		parts := strings.SplitN(path, "/", 2)

		value := v.MapIndex(reflect.ValueOf(parts[0]))
		if value.IsValid() {
			if len(parts) < 2 {
				return value.Interface(), nil
			}
			return resolveInternalReference(value.Interface(), parts[1])
		}
	}
	return nil, fmt.Errorf("invalid path %s", path)
}
