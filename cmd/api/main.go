package main

import (

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/joho/godotenv"

	"log"
	"net"
	"os"
	"strings"


	"gitlab.com/ReneGumzej/kitchenops/pkg/database"
	"gitlab.com/ReneGumzej/kitchenops/pkg/models"
	"gitlab.com/ReneGumzej/kitchenops/pkg/api"
	"gitlab.com/ReneGumzej/kitchenops/pkg/llm"
	"gitlab.com/ReneGumzej/kitchenops/pkg/logging"

)

func main()  {

	// .env-Datei im Projektverzeichnis laden (falls vorhanden).
	// Bereits gesetzte Umgebungsvariablen haben Vorrang.
	_ = godotenv.Load()

	// Logging zuerst initialisieren (Level/Format aus env)
	logging.Init()
	logging.Info("kitchenops wird gestartet",
		"log_level", os.Getenv("LOG_LEVEL"),
		"log_format", os.Getenv("LOG_FORMAT"),
	)

	app := fiber.New(
		fiber.Config{
			ServerHeader: "KitchenOps",
			
		},
	)

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5500",
			"http://127.0.0.1:5500",
		},
		AllowOriginsFunc: func(origin string) bool {
			host := strings.TrimPrefix(origin, "http://")
			host = strings.TrimPrefix(host, "https://")
			if i := strings.LastIndex(host, ":"); i > 0 {
				host = host[:i]
			}
			if host == "localhost" {
				return true
			}
			ip := net.ParseIP(host)
			return ip != nil && (ip.IsLoopback() || ip.IsPrivate())
		},
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodOptions,
		},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))

	app.Use(requestid.New())
	// Strukturiertes Request-Logging über slog (ersetzt das alte
	// Format-Template des fiber-loggers)
	app.Use(api.RequestLogger())


	db, err := database.DatabaseEngine()
	
	if err != nil {
		logging.Error("datenbankverbindung fehlgeschlagen", "fehler", err.Error())
		log.Fatalf("Can`t create database engine: %v", err)
	}
	logging.Info("datenbank verbunden")

	if err := db.AutoMigrate(&models.Lebensmittel{}, &models.Sonstiges{}, &models.Rezept{}, &models.Einkaufsartikel{}); err != nil {
		logging.Error("migration fehlgeschlagen", "fehler", err.Error())
		log.Fatalf("Can`t migrate database: %v", err)
	}
	logging.Info("migration abgeschlossen")

	dbHandler := api.NewDatabaseHandler(db)
	rezeptHandler := api.NewRezeptHandler(db, llm.NewClient())

	// API

	app.Get("/",api.InitGreet)

	app.Get("/lebensmittel", dbHandler.GetAllLebensmittel)
	app.Get("/lebensmittel/:id", dbHandler.GetLebensmittelByID)
	app.Post("/lebensmittel", dbHandler.CreateNewLebensmittel)
	app.Put("/lebensmittel/:id", dbHandler.UpdateLebensmittelByID)
	app.Delete("/lebensmittel/:id", dbHandler.DeleteLebensmittelByID)
	app.Post("/lebensmittel/kategorisieren", rezeptHandler.KategorisierenLebensmittel)
	app.Post("/lebensmittel/erfassen", rezeptHandler.ErfasseLebensmittelPerText)

	// Einkaufsliste
	app.Get("/einkaufsliste", dbHandler.GetAllEinkaufsartikel)
	app.Post("/einkaufsliste", dbHandler.CreateEinkaufsartikel)
	app.Put("/einkaufsliste/:id", dbHandler.UpdateEinkaufsartikel)
	app.Delete("/einkaufsliste/:id", dbHandler.DeleteEinkaufsartikelByID)
	app.Delete("/einkaufsliste", dbHandler.ClearEinkaufsliste)

	// Rezepte
	app.Get("/rezepte", rezeptHandler.GetAllRezepte)
	app.Post("/rezepte", rezeptHandler.SaveRezept)
	app.Post("/rezepte/generieren", rezeptHandler.GenerateRezepte)

	logging.Info("server startet", "adresse", "http://0.0.0.0:3000")
	if err := app.Listen(":3000"); err != nil {
		logging.Error("server gestoppt", "fehler", err.Error())
		log.Fatal(err)
	}
}
