package database

import (
    "database/sql"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Connect() error {
    var err error
    dsn := "root:Cbnits@123@tcp(localhost:3306)/user"
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
