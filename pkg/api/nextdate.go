package api

import (
	"net/http"
	"time"

	"github.com/Apriselkova/final_project/pkg/db"
)

// Константа для формата даты
const dateFormat = "20060102"

// nextDayHandler обрабатывает запросы к API для вычисления следующей даты
func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	var now time.Time
	var err error

	// Если параметр now не указан, используем текущую дату
	if nowStr == "" {
		now = time.Now()
	} else {
		now, err = time.Parse(dateFormat, nowStr)
		if err != nil {
			http.Error(w, "некорректная дата now", http.StatusBadRequest)
			return
		}
	}

	// Парсим дату
	startDate, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		http.Error(w, "некорректная дата date", http.StatusBadRequest)
		return
	}

	// Вызываем функцию NextDate
	nextDate, err := db.NextDate(now, startDate.Format(dateFormat), repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Возвращаем только дату как строку
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}
