package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getNotes(c *gin.Context) {
	var notes []Note
	if err := db.Find(&notes).Error; err != nil {
		log.Println("Ошибка получения заметок: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка получения заметок: ", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": notes})
}

func getNote(c *gin.Context) {
	var note Note
	if err := db.First(&note).Error; err != nil {
		log.Println("Ошибка получения заметки: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка получения заметки: ", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": note})
}
func addNote(c *gin.Context) {

}
func deleteNote(c *gin.Context) {
	id := c.Param("id")
	if err := db.Delete(&id); err != nil {
		log.Println("Ошибка удаления заметки: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("Ошибка удаления заметки: ", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Заметка удалена"})
}
func updateNote(c *gin.Context) {

}
