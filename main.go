package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

type Note struct {
	Id        int64     `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	Content   string    `json:"content"`
	Files     []string  `json:"files" gorm:"type:text;serializer:json"`
	Tags      []string  `json:"tags" gorm:"type:text;serializer:json"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	Id        int64     `json:"id" gorm:"primaryKey"`
	UserName  string    `json:"user_name"`
	Password  string    `json:"password"`
	Notes     []Note    `json:"notes" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foregignKey:UserID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	UserName string `json:"user_name"`
	jwt.RegisteredClaims
}

type AuthReq struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

func main() {
	if strings.TrimSpace(os.Getenv("JWT_SECRET_KEY")) == "" {
		if err := generateJwtSecretKey(); err != nil {
			log.Fatalln("Ошибка геенрации секретного ключа: ", err)
		}
	}
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

	protected := r.Group("/api")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/notes", getNotes)
		protected.GET("/notes/:id", getNote)
		protected.POST("/notes", addNote)
		protected.DELETE("/notes/:id", deleteNote)
		protected.PATCH("/notes/:id", updateNote)

		protected.DELETE("/user", deleteUser)
		protected.PATCH("/user", updateUser)
	}
	r.POST("/signIn", signIn)
	r.POST("/signUp", signUp)

	r.Run("0.0.0.0:8080")
}
