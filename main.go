package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Apriselkova/final_project/pkg/api"
	"github.com/Apriselkova/final_project/pkg/db"
)

func main() {
	// Инициализация базы данных
	err := db.Init("scheduler.db")
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
		db.Database.Close()
	}

	// Определяем переменную окружения или используем порт по умолчанию
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540" // Порт по умолчанию
	}

	// Указываем директорию для файлов
	webDir := "./web"

	// Обрабатываем запросы на корень и возвращаем файлы из директории
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	// Инициализация API обработчиков
	api.Init()

	// Запускаем сервер
	log.Printf("Сервер запущен на порту %s\n", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
