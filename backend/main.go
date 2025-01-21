package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Image struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Filename string `json:"filename"`
	Likes    int    `json:"likes"`
}

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var db *gorm.DB
var jwtSecret = []byte("your-secret-key")

func main() {
	// Initialize database
	fmt.Println("Starting server...")
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		log.Fatal("Database environment variables are not set correctly")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbName)
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		// Create database if not exists
		fmt.Println("Trying to create database : " + dbName + "...")
		dsn = dbUser + ":" + dbPassword + "@tcp(" + dbHost + ":3306)/"
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			fmt.Println(err.Error())
		}
		db.Exec("CREATE DATABASE " + dbName)
		db.Exec("USE " + dbName)
		fmt.Println("Database created : " + dbName)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&Image{}, &User{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Create Fiber app
	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New())

	// Routes
	app.Get("/api/images", protectedRoute(getImages))
	app.Post("/api/images", protectedRoute(uploadImage))
	app.Post("/api/images/:id/like", protectedRoute(toggleLike))
	app.Get("/api/images/expose/:filename", serveImage)

	// Authentication Routes
	app.Post("/api/register", register)
	app.Post("/api/login", login)

	// Start the server
	log.Fatal(app.Listen(":8080"))
}

// Middleware to protect routes
func protectedRoute(handler fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check JWT token
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(401).SendString("Unauthorized")
		}

		// Validate JWT
		_, err := validateToken(token)
		if err != nil {
			return c.Status(401).SendString("Invalid token")
		}

		// Call the next handler
		return handler(c)
	}
}

// Validate JWT token
func validateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure token's signing method matches
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	return token, err
}

// Register route
func register(c *fiber.Ctx) error {
	var user User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).SendString("Invalid input")
	}

	// Hash the password before storing it
	// Add password hashing here (e.g., bcrypt) for production

	if err := db.Create(&user).Error; err != nil {
		return c.Status(500).SendString("Failed to register user")
	}

	return c.Status(201).SendString("User registered successfully")
}

// Login route
func login(c *fiber.Ctx) error {
	var user User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).SendString("Invalid input")
	}

	var dbUser User
	if err := db.Where("username = ?", user.Username).First(&dbUser).Error; err != nil {
		return c.Status(400).SendString("Invalid username or password")
	}

	// Check password here (bcrypt or plain comparison for now)
	if dbUser.Password != user.Password {
		return c.Status(400).SendString("Invalid username or password")
	}

	// Generate JWT token
	token, err := generateToken(dbUser.ID)
	if err != nil {
		return c.Status(500).SendString("Error generating token")
	}

	return c.JSON(fiber.Map{"token": token})
}

// Generate JWT token
func generateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// Get list of images
func getImages(c *fiber.Ctx) error {
	var images []Image
	if err := db.Find(&images).Error; err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(images)
}

// Upload a new image
func uploadImage(c *fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(400).SendString("Invalid file")
	}

	// Save file to local storage
	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, os.ModePerm)
	}

	filePath := fmt.Sprintf("%s/%s", uploadDir, file.Filename)
	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	// Save file info to database
	image := Image{
		Filename: file.Filename,
		Likes:    0,
	}
	if err := db.Create(&image).Error; err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Status(201).SendString("Image uploaded successfully")
}

// Toggle like for an image
func toggleLike(c *fiber.Ctx) error {
	id := c.Params("id")

	var image Image
	if err := db.First(&image, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).SendString("Image not found")
		}
		return c.Status(500).SendString(err.Error())
	}

	image.Likes++
	if err := db.Save(&image).Error; err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.SendString("Like toggled successfully")
}

// Serve image from the "uploads" directory
func serveImage(c *fiber.Ctx) error {
	filename := c.Params("filename")

	// Check if file exists
	filePath := fmt.Sprintf("uploads/%s", filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return c.Status(404).SendString("Image not found")
	}

	return c.SendFile(filePath)
}
