package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
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
	rawId, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Был введен неверный айди пользователя"})
		return
	}
	userID, correct := rawId.(int64)
	if !correct {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка трансфорамации id"})
		return
	}
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
		UserID:  uint(userID),
	}

	if err := db.Create(&note).Error; err != nil {
		log.Println("Ошибка создания заметки: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка создания заметки: ", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": note})
}

func deleteNote(c *gin.Context) {
	var note Note
	id := c.Param("id")
	if err := db.First(&note, id).Error; err != nil {
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
	var authReq AuthReq
	if err := c.ShouldBindJSON(&authReq); err != nil {
		log.Println("Вы ввели неккоректные данные: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Вы ввели некорректные данные"})
		return
	}

	var user User
	if err := db.Where("user_name = ?", authReq.UserName).First(&user).Error; err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка полечения пользователя"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(authReq.Password)); err == nil {
		token, err := generateJWT(user.Id, authReq.UserName)
		if err != nil {
			log.Println("Ошибка генерации JWT: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации JWT"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprint("Вы успешно вошли в аккаунт: ", token)})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Введен неверный пароль"})
	}
}

func signUp(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("Введены неверные данные: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Введены неверные данные"})
		return
	}
	var count int64
	if err := db.Model(&User{}).Where("user_name = ?", user.UserName).Count(&count).Error; err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользователя"})
		return
	}
	if count <= 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Println("Ошибка, генерации пароля: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации пароля"})
			return
		}
		user.Password = string(hashedPassword)
		if err := db.Create(&user).Error; err != nil {
			log.Println("Ошибка создания пользовтеля", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания пользовтеля"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Вы успешно создали новый аккаунт!"})
	} else {
		c.JSON(http.StatusConflict, gin.H{"error": "Имя занято введите другое"})
	}
}

func deleteUser(c *gin.Context) {
	struserID, exists := c.Get("userID")
	userID, ok := struserID.(int64)
	if !ok {
		log.Println("Неверный индентефикатор пользовтеля")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный индентефикатор пользовтеля"})
		return
	}
	if exists {
		if err := db.Delete(&User{}, userID).Error; err != nil {
			log.Println("Ошибка удаления учетной записи: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления учетной записи"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Учетная запись успешно удалена!"})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Несанкционированный доступ, доступ закрыт!"})
	}
}

func updateUser(c *gin.Context) {
	struserID, exists := c.Get("userID")
	userID, ok := struserID.(int64)
	if !ok {
		log.Println("Неверный индентефикатор пользовтеля")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный индентефикатор пользовтеля"})
		return
	}
	var request User
	if err := c.ShouldBind(&request); err != nil {
		log.Println("Ошибка введены неверные данные: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка введены неверные данные: "})
		return
	}
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользователя"})
		return
	}

	if exists {
		if request.UserName != "" {
			user.UserName = request.UserName
		}
		if request.Password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Println("Ошибка, генерации пароль: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка, генерации пароля"})
				return
			}
			user.Password = string(hashedPassword)
		}
		if err := db.Save(&user).Error; err != nil {
			log.Println("Ошибка обнолвения пользоватлеля: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обнолвения пользоватлеля"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Учетная запись успешно обновлена"})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неавторезеринванны доступ, доступ закрыт!"})
	}

}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Ошибка вы не ввели токен"})
			c.Abort()
			return
		}

		if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный формат токена авторизации"})
			c.Abort()
			return
		}

		tokenString = tokenString[7:]

		log.Println(tokenString)

		claims, err := ValidateJWT(tokenString)
		if err != nil {
			log.Printf("Ошибка валидации JWT: %v", err) // Логируем ошибку для отладки
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительный или просроченный токен"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userName", claims.UserName)

		c.Next()
	}
}
