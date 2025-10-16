package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Apriselkova/final_project/pkg/db"
)

// addTaskHandler обрабатывает POST-запросы для добавления задачи
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJson(w, map[string]string{"error": "Метод не поддерживается"}, http.StatusMethodNotAllowed)
		return
	}

	var task db.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJson(w, map[string]string{"error": "Ошибка десериализации JSON"}, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Проверка обязательного поля title
	if task.Title == "" {
		writeJson(w, map[string]string{"error": "Не указан заголовок задачи"}, http.StatusBadRequest)
		return
	}

	// Получаем текущее время
	now := time.Now()
	currentDate := now.Format("20060102")

	// Обработка специального значения "today"
	if task.Date == "today" {
		task.Date = currentDate
	}

	// Если дата не указана, устанавливаем текущую дату
	if task.Date == "" {
		task.Date = currentDate
	}

	// Валидация формата даты
	_, err := time.Parse("20060102", task.Date)
	if err != nil {
		writeJson(w, map[string]string{"error": "Дата представлена в неверном формате"}, http.StatusBadRequest)
		return
	}

	// Валидация корректности даты (проверяем, что дата существует)
	if !isValidDate(task.Date) {
		writeJson(w, map[string]string{"error": "Некорректная дата"}, http.StatusBadRequest)
		return
	}

	// Парсим дату задачи для дальнейших проверок
	taskTime, err := time.Parse("20060102", task.Date)
	if err != nil {
		writeJson(w, map[string]string{"error": "Дата представлена в неверном формате"}, http.StatusBadRequest)
		return
	}

	// Получаем начало текущего дня для сравнения
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Обработка повторяющихся задач
	if task.Repeat != "" {
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeJson(w, map[string]string{"error": "Неправильный формат правила повторения"}, http.StatusBadRequest)
			return
		}
		task.Date = nextDate
	} else {
		// Для неповторяющихся задач проверяем, что дата не в прошлом
		if taskTime.Before(todayStart) {
			task.Date = currentDate
		}
	}

	// Добавление задачи в базу данных
	id, err := db.AddTask(&task)
	if err != nil {
		writeJson(w, map[string]string{"error": "Ошибка при добавлении задачи"}, http.StatusInternalServerError)
		return
	}

	// Возврат идентификатора задачи
	writeJson(w, map[string]int64{"id": id}, http.StatusOK)
}

// writeJson отправляет данные в формате JSON
func writeJson(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
