package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Глобальная переменная для хранения открытой базы данных
var Database *sql.DB

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// Строковая константа для схемы базы данных
const schema = `
CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title VARCHAR(255) NOT NULL,
    comment TEXT,
    repeat VARCHAR(128)
);
CREATE INDEX idx_date ON scheduler(date);
`

// Init инициализирует базу данных и создает таблицу, если она не существует
func Init(dbFile string) error {
	_, err := os.Stat(dbFile)
	install := false
	if err != nil {
		install = true
	}

	// Открываем (или создаем) базу данных
	Database, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	// Если база данных новая, создаем таблицу
	if install {
		if _, err := Database.Exec(schema); err != nil {
			return err
		}
		log.Println("Таблица scheduler успешно создана.")
	} else {
		log.Println("База данных уже существует.")
	}

	return nil
}

// AddTask добавляет задачу в таблицу scheduler и возвращает идентификатор добавленной записи
func AddTask(task *Task) (int64, error) {
	fmt.Println("Это task.Date в DB - ", task.Date)
	now := time.Now()
	todayTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	task.Date = todayTime.Format("20060102")

	if task.Date == "today" {
		todayTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		task.Date = todayTime.Format("20060102")
	}
	var id int64
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := Database.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	id, err = res.LastInsertId()
	return id, err
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
