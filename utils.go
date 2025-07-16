package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func saveFile(
	c *gin.Context,
	field string,
	name string,
	directory string,
) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println("Ошибка получшения пользовтельской диреткории: ", err)
		return "", err
	}

	absolutePath := filepath.Join(homeDir, "notes-files", directory)

	file, err := c.FormFile(field)
	if err != nil {
		log.Println("Ошиибка получения файла: ", err)
		return "", err
	}

	if err := os.MkdirAll(absolutePath, 0755); err != nil {
		log.Println("Ошибка создания директории: ", err)
		return "", err
	}

	fileName := fmt.Sprintf("%s-%s%s", name, uuid.New(), strings.ToLower(filepath.Ext(file.Filename)))
	saveFilePath := filepath.Join(absolutePath, fileName)

	if err := c.SaveUploadedFile(file, saveFilePath); err != nil {
		log.Printf("Ошибка сохранения файла: %v", err)
		return "", err
	}

	filePath := filepath.Join("media", directory, fileName)

	return filePath, nil
}

func generateJwtSecretKey() error {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Println("Ошибка генерации случайных байт: ", err)
		return err
	}
	if err := os.Setenv("JWT_SECRET_KEY", base64.URLEncoding.EncodeToString(b)); err != nil {
		log.Println("Ошибка сохранения в пренмную окружения: ", err)
		return err
	}
	return nil
}

func generateJWT(userID int64, userName string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:   userID,
		UserName: userName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "your-gin-crud-app",
			Subject:   strconv.FormatInt(userID, 10), // Преобразуем int в string для Subject
			Audience:  []string{"users"},
		},
	}
	secretKey, _ := base64.URLEncoding.DecodeString(os.Getenv("JWT_SECRET_KEY"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		log.Println("Ошибка преобрзаовния токена: ", err)
		return "", err
	}
	return tokenString, nil
}

func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		secretKey, _ := base64.URLEncoding.DecodeString(os.Getenv("JWT_SECRET_KEY"))
		return secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге токена: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("токен недействителен")
	}

	return claims, nil
}
