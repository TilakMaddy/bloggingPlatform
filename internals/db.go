package internals

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
)

type Author struct {
	ID   int64
	Name string
}

type Blog struct {
	ID       int64
	Title    string
	Content  string
	Images   []string
	AuthorID int64
}

type DBConn struct {
	DB      *sql.DB
	IsSetup bool
}

// Setup function should only run once will be used to populate DB
func (dbConn *DBConn) Setup() {

	cfg := mysql.Config{
		User:   os.Getenv("DB_USER"),
		Passwd: os.Getenv("DB_PASS"),
		Net:    "tcp",
		Addr:   os.Getenv("DB_HOST"),
		DBName: os.Getenv("DB_NAME"),
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())

	if err != nil {
		log.Fatalf("failed to open connection to database %v\n", err)
	}

	pingErr := db.Ping()

	if pingErr != nil {
		log.Fatalln("database failed to ping back")
	}

	dbConn.DB = db
	dbConn.IsSetup = true

}

func (dbConn *DBConn) publishBlog(w http.ResponseWriter, blog Blog) error {
	// todo: write an insert query
	return nil
}
