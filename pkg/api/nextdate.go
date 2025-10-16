package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	nextDate, err := NextDate(now, startDate.Format(dateFormat), repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Возвращаем только дату как строку
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}

// NextDate вычисляет следующую дату для задачи в соответствии с указанным правилом
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	startDate, err := time.Parse("20060102", dstart)
	if err != nil {
		return "", errors.New("некорректная дата dstart")
	}

	// Проверяем, что правило повторения не пустое
	if repeat == "" {
		return "", errors.New("параметр repeat не может быть пустым")
	}

	parts := strings.Split(repeat, " ")
	if len(parts) == 0 {
		return "", errors.New("неверный формат repeat")
	}

	var nextDate time.Time
	nextDate = startDate

	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return "", errors.New("не указано количество дней")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days < 1 || days > 400 {
			return "", errors.New("недопустимое значение для дней")
		}

		// Используем цикл для добавления дней
		for {
			nextDate = nextDate.AddDate(0, 0, days)
			if afterNow(nextDate, now) {
				break
			}
		}
		return nextDate.Format("20060102"), nil

	case "y":
		for {
			nextDate = nextDate.AddDate(1, 0, 0)
			if afterNow(nextDate, now) {
				break
			}
		}
		return nextDate.Format("20060102"), nil

	default:
		return "", errors.New("неподдерживаемый формат")
	}
}

// afterNow проверяет, является ли date позже now
func afterNow(date, now time.Time) bool {
	return date.After(now)
}
