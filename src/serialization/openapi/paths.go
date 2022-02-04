package openapi

type Path struct {
	Summary     string               `yaml:"summary"`
	Description string               `yaml:"description"`
	Parameters  interface{}          `yaml:"parameters"`
	Operations  map[string]Operation `yaml:",inline"`
}

type Operation struct {
	Summary     string   `yaml:"summary"`
	OperationId string   `yaml:"operationId"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
}