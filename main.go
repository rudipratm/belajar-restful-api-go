package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func initDB() *sql.DB {
	db, err := sql.Open("sqlite", "./data.db")
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	);`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

	return db
}

func main() {
	db := initDB()
	defer db.Close()

	r := gin.Default()

	r.GET("/items", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name FROM items")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var items []Item
		for rows.Next() {
			var item Item
			if err := rows.Scan(&item.ID, &item.Name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			items = append(items, item)
		}

		c.JSON(http.StatusOK, items)
	})

	r.POST("/items", func(c *gin.Context) {
		var newItem Item
		if err := c.ShouldBindJSON(&newItem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := db.Exec("INSERT INTO items (name) VALUES (?)", newItem.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		id, _ := result.LastInsertId()
		newItem.ID = int(id)
		c.JSON(http.StatusCreated, newItem)
	})

	r.PUT("/items/:id", func(c *gin.Context) {
		id := c.Param("id")
		itemId, _ := strconv.Atoi(id)
		var updateItem Item
		if err := c.ShouldBindJSON(&updateItem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec("UPDATE items SET name = ? WHERE id = ?", updateItem.Name, itemId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		updateItem.ID = itemId
		c.JSON(http.StatusOK, updateItem)
	})

	r.Run(":8080")
}

