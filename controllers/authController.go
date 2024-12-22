package controllers

import (
	"context"
	"fmt"
	"os"

	"log"
	"net/http"
	"time"

	"golang-jwtauth/database"
	"golang-jwtauth/models"
	"golang-jwtauth/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()
var UserCollection *mongo.Collection

func CreateUniqueIndex() {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}}, // Ascending index on the 'email' field
		Options: options.Index().SetUnique(true),  // Unique index
	}
	_, err := UserCollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatalf("Error creating unique index: %v", err)
	}
}
func MakeCollection() {
	UserCollection = database.UserData(database.G_DBClient, "Users")
}
func isDuplicateKeyError(err error) bool {
	// Check if the error is a duplicate key error (MongoDB error code for unique constraint violation)
	if mongoErr, ok := err.(mongo.WriteException); ok {
		for _, we := range mongoErr.WriteErrors {
			if we.Code == 11000 { // Duplicate key error code
				return true
			}
		}
	}
	return false
}
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}
func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		// validation of incoming data
		//Encrypt the password
		// Create a context with a timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		// Bind the request body to the User struct
		var user models.User
		if err := c.BindJSON(&user); err != nil {
			log.Println("Error binding JSON:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}
		time.Sleep(100 * time.Second)
		// Log the received user data
		log.Printf("Received user: %+v", user)

		// Validate the struct using the validator package
		if err := validate.Struct(user); err != nil {
			log.Println("Validation error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validateerr := utils.ValidateSignUpData(&user)
		if validateerr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validateerr.Error()})
			return
		}
		user.Password = HashPassword(user.Password)
		_, inserterr := UserCollection.InsertOne(ctx, user)
		if inserterr != nil {
			// Handle duplicate key error (email already exists)
			if isDuplicateKeyError(inserterr) {
				c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting user: " + inserterr.Error()})
			return
		}

		// Respond with success
		c.JSON(http.StatusOK, gin.H{"message": "Sign-up successful"})
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var loginCredentials map[string]interface{}
		err := c.BindJSON(&loginCredentials)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}
		if !utils.IsValidEmail(loginCredentials["email"].(string)) {
			c.JSON(http.StatusBadRequest, "Invalid User")
			return
		}
		filter := bson.D{{"email", loginCredentials["email"].(string)}}
		var userData models.User
		err = UserCollection.FindOne(ctx, filter).Decode(&userData)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Invalid User!")
			return
		}
		fmt.Printf("db user data : %v", userData)
		err = bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(loginCredentials["password"].(string)))
		if err != nil {
			c.JSON(http.StatusBadRequest, "Invalid User!!")
			return
		}
		// Generate JWT token
		claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": userData.Email,
			"exp": time.Now().Add(time.Hour * 24).Unix(), // Expires in 24 hours
		})
		err = godotenv.Load(".env")
		if err != nil {
			c.JSON(http.StatusInternalServerError, "Something Wrong! Try Again")
			return
		}
		token, err := claims.SignedString([]byte(os.Getenv("SECRET_KEY")))
		if err != nil {
			c.JSON(http.StatusInternalServerError, "Something Wrong! Try Again")
			return
		}
		c.SetCookie("UserAuthorizationCredenatials", token, int(time.Now().Add(time.Hour*24).Unix()), "/", "", false, true)
		// Respond with success
		c.JSON(http.StatusOK, gin.H{"message": "Login Successful"})

	}
}
