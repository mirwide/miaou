package model

const (
	Gemma2_2b                  = "gemma2:2b"
	Gemma2_9b                  = "gemma2:9b"
	Gemma2_27b                 = "gemma2:27b"
	Gemma2_27b_instruct_q3_K_S = "gemma2:27b-instruct-q3_K_S"
	Llama3_1_8b                = "llama3.1:8b"
	Llama3_1_70b               = "llama3.1:70b"
	Llama3_2_1b                = "llama3.2:1b"
	Llama3_2_3b                = "llama3.2:3b"
	Llava_7b                   = "llava:7b"
	Tlite_8b                   = "owl/t-lite:q4_0-instruct"
	Mistral_7b                 = "mistral:7b"
	Codellama_7b               = "codellama:7b"
	Solar_10_7b                = "solar:10.7b"
)

var (
	Models = map[string]Model{
		"gemma2":   Model{Tags: []string{"2b", "9b", "27b", "27b-instruct-q3_K_S"}},
		"llama3.1": Model{Tags: []string{"8b", "70b"}, SupportTools: true},
	}
)

type Model struct {
	Tags          []string
	SupportImages bool
	SupportTools  bool
}
