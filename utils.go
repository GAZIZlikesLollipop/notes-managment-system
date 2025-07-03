package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
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

	absolutePath := filepath.Join(homeDir, "notes-media", directory)

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
