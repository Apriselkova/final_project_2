package api

import "net/http"

// Init инициализирует обработчики API
func Init() {
	http.HandleFunc("/api/nextdate", nextDayHandler)
	http.HandleFunc("/api/task", taskHandler)          // обрабатывает get для одного
	http.HandleFunc("/api/task/done", doneTaskHandler) // обрабатывает get list
	http.HandleFunc("/api/tasks", tasksHandler)
}

// taskHandler обрабатывает различные HTTP-методы для /api/task
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)

	default:
		http.Error(w, `{"error":"Метод не поддерживается"}`, http.StatusMethodNotAllowed)
	}
}
