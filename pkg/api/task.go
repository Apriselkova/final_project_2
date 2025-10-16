package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Apriselkova/final_project/pkg/db"
)

// Task представляет задачу в базе данных
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// AddTask добавляет задачу в таблицу scheduler и возвращает идентификатор добавленной записи
func AddTask(task *Task) (int64, error) {
	var id int64
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Database.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	id, err = res.LastInsertId()
	return id, err
}

// getTaskHandler обрабатывает GET запросы для получения задачи по ID
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Получаем параметр id из query string
	idStr := r.URL.Query().Get("id")

	// Проверяем, что id передан
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Не указан идентификатор"})
		return
	}

	// Конвертируем строку в число
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат идентификатора"})
		return
	}

	// Получаем задачу из базы данных
	task, err := GetTask(strconv.Itoa(id))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Задача не найдена"})
		return
	}

	// Возвращаем задачу в формате JSON
	json.NewEncoder(w).Encode(task)
}

// GetTask получает задачу из базы данных по ID
func GetTask(id string) (*Task, error) {
	var task Task

	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	row := db.Database.QueryRow(query, id)

	// Сканируем ID как число и конвертируем в строку
	var idInt int64
	err := row.Scan(&idInt, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return nil, err
	}

	task.ID = strconv.FormatInt(idInt, 10)

	return &task, nil
}

// updateTaskHandler обрабатывает PUT запросы для обновления задачи
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var task Task

	// Десериализуем JSON из тела запроса
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка десериализации JSON"})
		return
	}
	defer r.Body.Close()

	// Проверяем обязательные поля
	if task.ID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Не указан идентификатор задачи"})
		return
	}

	if task.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Не указан заголовок задачи"})
		return
	}

	// Валидация даты
	if !isValidDate(task.Date) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат даты"})
		return
	}

	// Валидация правила повтора (если указано)
	if task.Repeat != "" && !isValidRepeatPattern(task.Repeat) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат правила повтора"})
		return
	}

	// Обновляем задачу в базе данных
	err := UpdateTask(&task)
	if err != nil {
		if err.Error() == "задача не найдена" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Задача не найдена"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка при обновлении задачи"})
		}
		return
	}

	// Возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

// UpdateTask обновляет задачу в базе данных
func UpdateTask(task *Task) error {
	id, err := strconv.ParseInt(task.ID, 10, 64)
	if err != nil {
		return fmt.Errorf("неверный формат идентификатора")
	}

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	result, err := db.Database.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, id)
	if err != nil {
		return err
	}

	// Проверяем, была ли обновлена хотя бы одна запись
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("задача не найдена")
	}

	return nil
}

// isValidDate проверяет валидность формата даты YYYYMMDD
func isValidDate(date string) bool {
	if len(date) != 8 {
		return false
	}

	_, err := time.Parse("20060102", date)
	return err == nil
}

// isValidRepeatPattern проверяет валидность формата правила повтора
func isValidRepeatPattern(repeat string) bool {
	if repeat == "" {
		return true
	}

	// Простая проверка - должен начинаться с d, w, m, y и содержать число
	if len(repeat) < 3 {
		return false
	}

	firstChar := repeat[0]
	if firstChar != 'd' && firstChar != 'w' && firstChar != 'm' && firstChar != 'y' {
		return false
	}

	// Проверяем, что после пробела есть число
	if len(repeat) < 2 || repeat[1] != ' ' {
		return false
	}

	_, err := strconv.Atoi(repeat[2:])
	return err == nil
}

// deleteTaskHandler обрабатывает DELETE запросы для удаления задачи
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Получаем параметр id из query string
	idStr := r.URL.Query().Get("id")

	// Проверяем, что id передан
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Не указан идентификатор"})
		return
	}

	// Конвертируем строку в число для проверки валидности
	_, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат идентификатора"})
		return
	}

	// Удаляем задачу
	err = DeleteTask(idStr)
	if err != nil {
		if err.Error() == "задача не найдена" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Задача не найдена"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка при удалении задачи"})
		}
		return
	}

	// Возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

// DeleteTask удаляет задачу из базы данных по ID
func DeleteTask(id string) error {
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("неверный формат идентификатора")
	}

	query := `DELETE FROM scheduler WHERE id = ?`
	result, err := db.Database.Exec(query, idInt)
	if err != nil {
		return err
	}

	// Проверяем, была ли удалена хотя бы одна запись
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("задача не найдена")
	}

	return nil
}

// doneTaskHandler обрабатывает POST запросы для отметки задачи как выполненной
func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Получаем параметр id из query string
	idStr := r.URL.Query().Get("id")

	// Проверяем, что id передан
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Не указан идентификатор"})
		return
	}

	// Конвертируем строку в число для проверки валидности
	_, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат идентификатора"})
		return
	}

	// Получаем задачу из базы данных
	task, err := GetTask(idStr)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Задача не найдена"})
		return
	}

	// Если задача не повторяется - удаляем её
	if task.Repeat == "" {
		err = DeleteTask(idStr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка при удалении задачи"})
			return
		}
	} else {
		// Для повторяющейся задачи вычисляем следующую дату
		now := time.Now()
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка при вычислении следующей даты"})
			return
		}

		// Обновляем дату задачи
		err = UpdateTaskDate(idStr, nextDate)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка при обновлении даты задачи"})
			return
		}
	}

	// Возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

// UpdateTaskDate обновляет только дату задачи
func UpdateTaskDate(id string, date string) error {
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("неверный формат идентификатора")
	}

	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	result, err := db.Database.Exec(query, date, idInt)
	if err != nil {
		return err
	}

	// Проверяем, была ли обновлена хотя бы одна запись
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("задача не найдена")
	}

	return nil
}
