package api

import (
	"encoding/json"
	"net/http"

	"github.com/Apriselkova/final_project/pkg/db"
)

// TasksResp структура для ответа с задачами
type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// tasksHandler обрабатывает GET-запросы для /api/tasks
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tasks, err := db.Tasks(20)
	if err != nil {
		http.Error(w, `{"error":"Ошибка при получении задач"}`, http.StatusInternalServerError)
		return
	}

	// Если tasks равен nil, создаем пустой слайс
	if tasks == nil {
		tasks = []*db.Task{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TasksResp{Tasks: tasks})
}
