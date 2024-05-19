package main

import (
	"forum/internal/handlers"
	"forum/internal/handlers/middleware"
	"forum/internal/managers"
	"forum/internal/storage/mongo"
	"forum/internal/storage/mysql"
	"forum/internal/storage/redis"
	"log/slog"
	"net/http"
	"os"

	handls "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Покачто будет так, надо сделать конфиг
const (
	mysqlDBUser     = "root"
	mysqlDBPassword = "57ry4ardpa77"
	mysqlAddr       = "mysql:3306"
	mysqlDBName     = "forum"

	mongoDBName = "forum"
	mongoAddr   = "mongodb://mongodb"

	redisAddr     = "redis:6379"
	redisPassword = ""
	redisDBName   = 0

	userTable      = "users"
	postCollection = "post"
)

func main() {
	logger := slog.Default()

	dbMySQL, err := mysql.GetConnection(mysqlDBUser, mysqlDBPassword, mysqlAddr, mysqlDBName)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	dbMongo, err := mongo.GetConnection(mongoDBName, mongoAddr)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	redisClient := redis.GetConnect(redisDBName, redisPassword, redisAddr)
	if redisClient == nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	sessionStorage := redis.NewRedisStorage(redisClient)
	userStorage := mysql.NewUserStorage(dbMySQL, userTable)
	postStorage := mongo.NewPostStorage(dbMongo, postCollection)

	authManager := managers.NewSeesionManager(userStorage, sessionStorage)
	postManager := managers.NewPostManager(postStorage)

	userHandler := handlers.UserHandler{
		Logger:      logger,
		AuthManager: authManager,
	}

	postHandler := handlers.PostHandler{
		Logger:      logger,
		PostManager: postManager,
	}

	authMiddleware := middleware.NewAuthMiddleware(authManager, logger).GetHandler

	router := mux.NewRouter()

	router.HandleFunc("/api/register", userHandler.Register).Methods(http.MethodPost)
	router.HandleFunc("/api/login", userHandler.Login).Methods(http.MethodPost)

	// пути, требующие аутентификацию
	router.HandleFunc("/api/posts", authMiddleware(postHandler.Create)).Methods(http.MethodPost)
	router.HandleFunc("/api/post/{postID}", authMiddleware(postHandler.Delete)).Methods(http.MethodDelete)
	router.HandleFunc("/api/post/{postID}/{action}", authMiddleware(postHandler.UpdateVotes)).Methods(http.MethodGet)
	router.HandleFunc("/api/post/{postID}", authMiddleware(postHandler.AddComment)).Methods(http.MethodPost)
	router.HandleFunc("/api/post/{postID}/{commentID}", authMiddleware(postHandler.DeleteComment)).Methods(http.MethodDelete)

	router.HandleFunc("/api/posts/", postHandler.GetAll).Methods(http.MethodGet)
	router.HandleFunc("/api/posts/{category}", postHandler.GetAllByCategory).Methods(http.MethodGet)
	router.HandleFunc("/api/post/{postID}", postHandler.GetByID).Methods(http.MethodGet)
	router.HandleFunc("/api/user/{username}", postHandler.GetAllByUser).Methods(http.MethodGet)

	panicMiddleware := middleware.Panic(logger, router)

	origin := os.Getenv("ORIGIN_ALLOWED")

	headersOk := handls.AllowedHeaders([]string{"Authorization","Content-Type", "application/json"})
	originsOk := handls.AllowedOrigins([]string{origin})
	methodsOk := handls.AllowedMethods([]string{"GET", "POST", "DELETE"})

	addr := ":8080"
	logger.Info(
		"starting server",
		"address", addr,
	)
	err = http.ListenAndServe(addr, handls.CORS(headersOk, originsOk, methodsOk)(panicMiddleware))
	if err != nil {
		logger.Error(err.Error())
	}
}
