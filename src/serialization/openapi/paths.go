package openapi

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
