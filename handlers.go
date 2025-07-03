package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func getNotes(c *gin.Context) {
	var notes []Note
	if err := db.Find(&notes).Error; err != nil {
		log.Println("Ошибка получения заметок: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка получения заметок: ", err)})
		return
	}
	c.JSON(http.StatusOK, notes)
}

func getNote(c *gin.Context) {
	id := c.Param("id")
	var note Note
	if err := db.First(&note, id).Error; err != nil {
		log.Println("Ошибка получения заметки: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка получения заметки: ", err)})
		return
	}
	c.JSON(http.StatusOK, note)
}

func addNote(c *gin.Context) {
	var note Note
	name := c.PostForm("name")
	content := c.PostForm("content")
	filesStr := c.PostFormArray("files")
	tags := c.PostFormArray("tags")
	var files []string
	for _, f := range filesStr {
		filePath, err := saveFile(c, "note", strings.ReplaceAll(f, " ", ""), "notes")
		if err != nil {
			log.Println("Ошибка получения пути файла: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка получения пути файла: ", err)})
			return
		}
		files = append(files, fmt.Sprintf("http://192.168.1.9:8080/%s", filePath))
	}
	note = Note{
		Name:    name,
		Content: content,
		Files:   files,
		Tags:    tags,
	}

	c.JSON(http.StatusCreated, gin.H{"message": note})
}

func deleteNote(c *gin.Context) {
	var note Note
	id := c.Param("id")
	if err := db.First(&note, id); err != nil {
		log.Println("Ошибка получения заметки: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка получения заметки: ", err)})
		return
	}
	if len(note.Files) > 0 {
		for _, f := range note.Files {
			if strings.TrimSpace(f) != "" {
				file, err := url.Parse(f)
				if err != nil {
					log.Println("Ошибка получения пути к файлу: ", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка получения пути к файлу: ", err)})
					return
				}
				os.Remove(file.Path)
			}
		}
	}
	if err := db.Delete(&note).Error; err != nil {
		log.Println("Ошибка удаления заметки: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка удаления заметки: ", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Заметка удалена"})
}

func updateNote(c *gin.Context) {
	id := c.Param("id")
	var note Note
	if err := db.First(&note, id).Error; err != nil {
		log.Println("Ошибка получения заметки: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка получения заметки: ", err)})
		return
	}
	if name := c.PostForm("name"); name != "" {
		note.Name = name
	}
	if content := c.PostForm("content"); content != "" {
		note.Content = content
	}
	if filesStr := c.PostFormArray("files"); len(filesStr) <= 0 {
		var files []string
		for _, f := range filesStr {
			filePath, err := saveFile(c, "note", strings.ReplaceAll(f, " ", ""), "notes")
			if err != nil {
				log.Println("Ошибка получения пути файла: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка получения пути файла: ", err)})
				return
			}
			files = append(files, fmt.Sprintf("http://192.168.1.9:8080/%s", filePath))
		}
		note.Files = files
	}
	if tags := c.PostFormArray("tags"); len(tags) <= 0 {
		note.Tags = tags
	}
	if err := db.Save(note).Error; err != nil {
		log.Println("Ошибка обнвления заметки: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка обнвления заметки: ", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": note})
}

func getUsers(c *gin.Context) {
	var users []User
	if err := db.Preload("Notes").Find(&users).Error; err != nil {
		log.Println("Ошибка получения пользователей: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка получения пользоватлей: ", err)})
		return
	}
	c.JSON(http.StatusOK, users)
}

func getUser(c *gin.Context) {
	id := c.Param("id")
	var user User
	if err := db.Preload("Notes").First(&user, id).Error; err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка получения пользователя: ", err)})
		return
	}
	c.JSON(http.StatusOK, user)
}

func addUser(c *gin.Context) {

}

func deleteUser(c *gin.Context) {

}

func updateUser(c *gin.Context) {

}
