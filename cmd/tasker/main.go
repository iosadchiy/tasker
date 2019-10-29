package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"tasker/tasker"
)

const maxWaitTime = 25 * time.Second

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")

	db, err := gorm.Open("postgres", dbURL)
	if err != nil {
		log.Panic(err)
	}

	svc := tasker.NewService(db)
	handlers := tasker.Handlers{
		MaxWaitTime: maxWaitTime,
		Service:     svc,
	}

	r := mux.NewRouter()
	r.HandleFunc("/task", handlers.HandleCreate).Methods(http.MethodGet)
	r.HandleFunc("/task/{id}", handlers.HandleRead).Methods(http.MethodGet)
	r.HandleFunc("/task/{id}/finished", handlers.HandlePoll).Methods(http.MethodGet)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Panic(err)
	}
}
