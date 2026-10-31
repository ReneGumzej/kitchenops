package database

import (
	"errors"
	"os"
	"path/filepath"
	//"gorm.io/driver/postgres"
  	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	//"fmt"
	"gorm.io/driver/sqlite" 
	//"github.com/knadh/koanf/parsers/dotenv"
	
)

// var relPath string = "pkg/config/.env"
// var absPath, err = filepath.Abs(relPath)

// if err != nil {
// 	log.Fatalf("Path '%s' doesn`t exist: %v",relPath, err)
// }

// env_file := file.Provider(absPath)
// if err := env.Load(env_file,dotenv.Parser()); err != nil {
// 	log.Fatalf("can`t open .env-file: %v", err)
// }


// type PostgresConfig struct{
// 	Host string
// 	Port uint16
// 	Database string
// 	Username string
// 	Password string
// 	Schema string 
	
// }

// func (cfg PostgresConfig) ConnectionString() string {
	
// 	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Europe/Berlin search_path=%s", 
// 	cfg.Host, cfg.Username, cfg.Password, cfg.Database, cfg.Port, cfg.Schema)
// }



func DatabaseEngine() (*gorm.DB, error){
	// Einheitlicher Speicherort für Docker und lokale Entwicklung:
	// DB_PATH konfigurierbar, Standard data/kitchenops.db (wird bei Bedarf
	// angelegt - Docker mountet ./data nach /data und setzt DB_PATH absolut).
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/kitchenops.db"
	}

	// Verzeichnis sicherstellen (frischer Clone ohne data/-Ordner)
	if dir := filepath.Dir(dbPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, errors.New("Kann Datenbank-Verzeichnis nicht anlegen: " + err.Error())
		}
	}

	// GORM-Logging stillschalten: erwartete 404-Suchen ("record not found")
	// erzeugen sonst Rauschen. Anwendungsfehler werden strukturiert über
	// pkg/logging protokolliert.
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	
	if err != nil {
		return nil, errors.New("Cant establish database connection!")
	}

	return db, nil
}
