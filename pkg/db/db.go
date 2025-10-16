package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
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
func AddTask(task *Task, dateFormat string) (int64, error) {
	fmt.Println("Это task.Date в DB - ", task.Date)
	now := time.Now()
	todayTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	task.Date = todayTime.Format(dateFormat)

	if task.Date == "today" {
		todayTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		task.Date = todayTime.Format(dateFormat)
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
