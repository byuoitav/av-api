package hateoas

type Link struct {
	Rel  string `json:"rel"`
	HREF string `json:"href"`
}

type Swagger struct {
	Paths map[string]Path `json:"paths"`
}

type Path struct {
	Get    *Method `yaml:"get,omitempty"`
	Post   *Method `yaml:"post,omitempty"`
	Put    *Method `yaml:"put,omitempty"`
	Delete *Method `yaml:"delete,omitempty"`
}

type Method struct {
	Summary    string      `yaml:"summary,omitempty"`
	Parameters []Parameter `yaml:"parameters,omitempty"`
}

type Parameter struct {
	Name     string `yaml:"name,omitempty"`
	In       string `yaml:"in,omitempty"`
	Required bool   `yaml:"required,omitempty"`
}
