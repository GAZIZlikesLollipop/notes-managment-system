package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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

func signIn(c *gin.Context) {
	userName := c.Param("userName")
	password := c.Param("password")

	var user User
	if err := db.Where("userName = ?", userName).Preload("Notes").First(&user).Error; err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка получения пользователя: ", err)})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err == nil {
		c.JSON(http.StatusOK, user)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintln("Введен неверный пароль: ", err)})
	}
}

func signUp(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("Введены неверные данные: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintln("Введены неверные данные: ", err)})
		return
	}
	var users []User
	if err := db.Find(&users).Error; err != nil {
		log.Println("Ошибка получения пользователей: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка получения пользователей: ", err)})
		return
	}
	if slices.ContainsFunc(users, func(u User) bool {
		return u.UserName == user.UserName
	}) == false {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Println("Ошибка, введен неверный пароль: ", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintln("Ошибка, введен неверный пароль: ", err)})
			return
		}
		user.Password = string(hashedPassword)
		if err := db.Create(&user); err != nil {
			log.Println("Ошибка создания пользовтеля", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка создания пользовтеля", err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintln("Вы успешно создали новый аккаунт!")})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintln("Пользоватeл с таким именем уже имеется!")})
	}
}

func deleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := db.Delete(&User{}, id); err != nil {
		log.Println("Ошибка удаления вашего аккаунта", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка удаления вашего аккаунта", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Ваш аккаунт успешно удален!"})
}

func updateUser(c *gin.Context) {

}
