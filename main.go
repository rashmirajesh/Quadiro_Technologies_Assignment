package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Car struct {
	ID               uint   `gorm:"primaryKey" json:"id"`
	Name             string `json:"name"`
	ManufacturingYear int    `json:"manufacturing_year"`
	Price            float64 `json:"price"`
}

type Admin struct {
	Username string
	Password string
}

var admin = Admin{
	Username: "admin",
	Password: "password",
}

var DB *gorm.DB

func initDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	DB.AutoMigrate(&Car{})
}

func main() {
	initDB()
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.GET("/login", showLoginPage)
	r.POST("/login", handleLogin)

	adminGroup := r.Group("/admin")
	adminGroup.Use(authMiddleware)
	{
		adminGroup.GET("/dashboard", showDashboard)
		adminGroup.GET("/cars", getCars)
		adminGroup.POST("/cars", createCar)
		adminGroup.PUT("/cars/:id", updateCar)
		adminGroup.DELETE("/cars/:id", deleteCar)
	}

	r.Run(":8080")
}

func showLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Assignment for Quadiro Technologies",
	})
}

func handleLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == admin.Username && password == admin.Password {
		c.SetCookie("admin", "true", 3600, "/", "localhost", false, true)
		c.Redirect(http.StatusFound, "/admin/dashboard")
	} else {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"error": "Invalid credentials",
		})
	}
}

func authMiddleware(c *gin.Context) {
	cookie, err := c.Cookie("admin")
	if err != nil || cookie != "true" {
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
	}
	c.Next()
}

func showDashboard(c *gin.Context) {
	var cars []Car
	DB.Find(&cars)
	totalCars := len(cars)

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"cars":      cars,
		"totalCars": totalCars,
	})
}

func getCars(c *gin.Context) {
	var cars []Car
	DB.Find(&cars)
	c.JSON(http.StatusOK, cars)
}

func createCar(c *gin.Context) {
	var car Car
	if err := c.ShouldBindJSON(&car); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	DB.Create(&car)
	c.JSON(http.StatusCreated, car)
}

func updateCar(c *gin.Context) {
	id := c.Param("id")
	var car Car
	if err := DB.First(&car, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Car not found"})
		return
	}

	if err := c.ShouldBindJSON(&car); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	DB.Save(&car)
	c.JSON(http.StatusOK, car)
}

func deleteCar(c *gin.Context) {
	id := c.Param("id")
	if err := DB.Delete(&Car{}, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Car not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Car deleted"})
}
