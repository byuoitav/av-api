package hateoas

type Link struct {
	Rel  string `json:"rel"`
	HREF string `json:"href"`
}

type Swagger struct {
	Info  Info            `json:"info,omitempty"`
	Paths map[string]Path `json:"paths"`
}

type Info struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
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

// Root is a generic struct utilized at the root of an API to provide initial HATEOAS links
type Root struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
	Links       []Link `json:"links,omitempty"`
}
