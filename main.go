package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

type Note struct {
	Id        int64  `json:"id" gorm:"primaryKey"`
	Name      string `json:"name"`
	UserID    uint
	Content   string    `json:"content"`
	Files     []string  `json:"files" gorm:"type:text;serializer:json"`
	Tags      []string  `json:"tags" gorm:"type:text;serializer:json"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type User struct {
	Id        int64     `json:"id" gorm:"primaryKey"`
	UserName  string    `json:"userName"`
	Password  string    `json:"password"`
	Notes     []Note    `json:"notes" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foregignKey:UserID"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
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

	r.GET("/signIn/:userName/:password", signIn)
	r.POST("/signUp", signUp)
	r.DELETE("/user/:userName/:password", deleteUser)
	r.PATCH("/user/:userName/:password", updateUser)

	r.Run("0.0.0.0:8080")
}
