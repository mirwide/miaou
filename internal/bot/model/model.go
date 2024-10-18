package model

import "errors"

var (
	Models = map[string]Model{
		"gemma2:9b":                  {},
		"gemma2:27b-instruct-q3_K_S": {Hidden: true},
		"llama3.1:8b":                {SupportTools: true},
		"mistral:7b":                 {SupportTools: true},
		"qwen2.5:14b":                {SupportTools: true},
		"owl/t-lite:q4_0-instruct":   {},
		"llava:13b":                  {SupportImages: true},
	}
)

type Model struct {
	Name          string
	SupportImages bool
	SupportTools  bool
	Hidden        bool
}

func NewModel(name string) (Model, error) {
	m, ok := Models[name]
	if ok {
		m.Name = name
		return m, nil
	}
	return Model{}, errors.New("model not found")
}
