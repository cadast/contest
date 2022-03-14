package openapi

type Components struct {
	Schemas    map[string]*Schema   `yaml:"schemas"`
	Parameters map[string]Parameter `yaml:"parameters"`
}

type SchemaType string

const (
	SchemaTypeObject  SchemaType = "object"
	SchemaTypeArray   SchemaType = "array"
	SchemaTypeString  SchemaType = "string"
	SchemaTypeInteger SchemaType = "integer"
	SchemaTypeNumber  SchemaType = "number"
	SchemaTypeBoolean SchemaType = "boolean"
)

type Schema struct {
	Title       string             `yaml:"title"`
	Type        SchemaType         `yaml:"type"`
	Description string             `yaml:"description"`
	Properties  map[string]*Schema `yaml:"properties"`
	Required    []string           `yaml:"required"`
	Nullable    bool               `yaml:"nullable"`
	Items       *Schema            `yaml:"items"`

	AnyOf []*Schema `yaml:"anyOf"`
	OneOf []*Schema `yaml:"oneOf"`
	AllOf []*Schema `yaml:"allOf"`

	Ref string `yaml:"$ref"`
}

func (s Schema) Requires(key string) bool {
	for _, val := range s.Required {
		if val == key {
			return true
		}
	}
	return false
}

type ParameterIn string

const (
	ParameterInQuery  ParameterIn = "query"
	ParameterInHeader ParameterIn = "header"
	ParameterInPath   ParameterIn = "path"
	ParameterInCookie ParameterIn = "cookie"
)

type Parameter struct {
	Name            string      `yaml:"name"`
	In              ParameterIn `yaml:"in"`
	Description     string      `yaml:"description"`
	Required        bool        `yaml:"required"`
	Deprecated      bool        `yaml:"deprecated"`
	AllowEmptyValue bool        `yaml:"allowEmptyValue"`

	Ref string `yaml:"$ref"`

	Schema Schema `yaml:"schema"`
}
