package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

type Note struct {
	Id      int64    `json:"id" gorm:"primaryKey"`
	Name    string   `json:"name"`
	Content string   `json:"content"`
	Files   []string `json:"files"`
	Tags    []string `json:"tags"`
}

type User struct {
	Id            int64  `json:"id" gorm:"primaryKey"`
	Login         string `json:"login"`
	Password_Hash string `json:"password_hash"`
	Password_Salt string `json:"password_salt"`
	Notes         []Note `json:"notes"`
}

func main() {
	r := gin.Default()
	var err error
	dsn := "host=localhost user=postgres dbname=notesdb port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln("Ошибка открытия базы данных: ", err)
	}
	if err := db.AutoMigrate(&User{}, &Note{}); err != nil {
		log.Fatalln("Ошибка миграции элемнтов базы данных: ", err)
	}

	r.GET("/notes", getNotes)
	r.GET("/notes/:id", getNote)
	r.POST("/notes", addNote)
	r.DELETE("/notes/:id", deleteNote)
	r.PATCH("/notes/:id", updateNote)

	r.GET("/users", getUsers)
	r.GET("/users/:id", getUser)
	r.POST("/users", addUser)
	r.DELETE("/users/:id", deleteUser)
	r.PATCH("/users/:id", updateUser)

	r.Run("0.0.0.0:8080")
}
