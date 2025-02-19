package bot

var (
	start command = command{"start", "Начать новый чат"}
	reset command = command{"reset", "Начать новый чат"}
	model command = command{"model", "Выбрать LLM модель"}
)

type command struct {
	command string
	desc    string
}
