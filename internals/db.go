package internals

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"log"
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

func (dbConn *DBConn) publishBlog(blog Blog) error {
	//goland:noinspection ALL
	var _, err = dbConn.DB.Exec(
		"INSERT INTO blogs (author, content, image, title) VALUES (?, ?, ?, ?)",
		blog.AuthorID,
		blog.Content,
		stringifyToMySQLJSONArray(blog.Images),
		blog.Title,
	)

	if err != nil {
		return err
	}

	return nil
}
