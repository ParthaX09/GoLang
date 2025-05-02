package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var DB *sql.DB

func Connect() error {

    // var err error
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    // Retrieve environment variables
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    dbname := os.Getenv("DB_NAME")

    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, dbname)
    
    DB, err = sql.Open("mysql", dsn)
    if err != nil {
        return fmt.Errorf("error opening database: %v", err)
    }

    if err = DB.Ping(); err != nil {
        return fmt.Errorf("error connecting to database: %v", err)
    }

    fmt.Println("Connected to MySQL!")
    return nil
}
