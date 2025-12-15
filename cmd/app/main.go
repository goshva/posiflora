package main

import (
	"context"
	_ "embed"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"log"
	"net/http"
	"os"
	"posiflora-mvp/internal/database"
	"posiflora-mvp/internal/handlers"
	"posiflora-mvp/internal/repository"
	"posiflora-mvp/internal/service"
	"posiflora-mvp/internal/telegram"
	"strconv"
)

// ======================
// Embedded static files
// ======================

//go:embed index.html
var indexHTML []byte

//go:embed script.js
var scriptJS []byte

//go:embed swagger/swagger.json
var swaggerJSON []byte

//go:embed style.css
var styleCSS []byte

func mustLoadTemplates() *template.Template {
	tpl, err := template.New("index").Parse(string(indexHTML))
	if err != nil {
		log.Fatal("failed to parse embedded html template:", err)
	}
	return tpl
}

func main() {
	ctx := context.Background()

	// ======================
	// ENV
	// ======================
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ======================
	// Database
	// ======================
	pool, err := database.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer pool.Close()

	// ======================
	// Telegram client
	// ======================
	var telegramClient telegram.Client
	if os.Getenv("TELEGRAM_MOCK") == "true" {
		log.Println("Telegram client: MOCK")
		telegramClient = &telegram.MockClient{}
	} else {
		log.Println("Telegram client: REAL")
		telegramClient = telegram.NewHTTPClient()
	}

	// ======================
	// Repositories
	// ======================
	orderRepo := repository.NewOrderRepository(pool)
	telegramRepo := repository.NewTelegramRepository(pool)

	// ======================
	// Services
	// ======================
	orderService := service.NewOrderService(orderRepo, telegramRepo, telegramClient)

	// ======================
	// Handlers
	// ======================
	orderHandler := handlers.NewOrderHandler(orderService)
	telegramHandler := handlers.NewTelegramHandler(telegramRepo)
	statusHandler := handlers.NewStatusHandler(telegramRepo)

	pageTemplate := mustLoadTemplates()

	// ======================
	// Router
	// ======================
	router := httprouter.New()

	router.POST("/shops/:shopId/orders", orderHandler.CreateOrder)
	router.POST("/shops/:shopId/telegram/connect", telegramHandler.ConnectTelegram)
	router.GET("/shops/:shopId/telegram/status", statusHandler.GetStatus)

	// HTML-страница
	router.GET("/shops/:shopId/growth/telegram", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		shopIDStr := ps.ByName("shopId")
		shopID, err := strconv.ParseInt(shopIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid shopId", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := pageTemplate.Execute(w, struct{ ShopID int64 }{ShopID: shopID}); err != nil {
			http.Error(w, "failed to render page", http.StatusInternalServerError)
			log.Println("template error:", err)
			return
		}
	})

	// ======================
	// Статика: CSS, JS
	// ======================
	router.GET("/static/script.js", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(scriptJS)
	})
	router.GET("/static/style.css", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "text/css")
		w.Write(styleCSS)
	})
	router.GET("/health", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	router.GET("/swagger/swagger.json", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(swaggerJSON)
	})

	router.GET("/swagger", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<head>
		<title>Swagger UI</title>
		<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css">
		</head>
		<body>
		<div id="swagger-ui"></div>
		<script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
		<script>
		SwaggerUIBundle({
		url: '/swagger/swagger.json',
		dom_id: '#swagger-ui'
		});
		</script>
		</body>
		</html>
		`))
	})

	// ======================
	// Start server
	// ======================
	log.Println("Server started on :" + port)
	log.Println("Open: http://localhost:" + port + "/shops/1/growth/telegram")
	log.Println("Swagger: http://localhost:" + port + "/swagger")

	log.Fatal(http.ListenAndServe(":"+port, router))
}
