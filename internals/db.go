package internals

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

// Two queries - SELECT, DELETE both happen individually
func (dbConn *DBConn) deleteBlogByID(blogId int64) error {

	blog, fetchErr := dbConn.fetchBlogByID(blogId)
	if fetchErr == nil {
		deleteBlogAssets(&blog)
	}

	//goland:noinspection ALL
	_, err := dbConn.DB.Exec("DELETE FROM blogs WHERE id = ?", blogId)
	if err != nil {
		return err
	}
	return nil
}

// Single SELECT query to blogs table with WHERE clause
func (dbConn *DBConn) getBlogsByAuthor(authorID int64) ([]Blog, error) {
	//goland:noinspection ALL
	rows, err := dbConn.DB.Query(
		"SELECT * FROM blogs WHERE author = ?", authorID)
	if err != nil {
		return nil, err
	}
	return convertMySQLRowsToBlogs(rows)
}

// Single SELECT query to blogs table with LIKE clause
func (dbConn *DBConn) searchBlogs(substring string) ([]Blog, error) {
	//goland:noinspection ALL
	rows, err := dbConn.DB.Query(
		"SELECT * FROM blogs WHERE content LIKE CONCAT('%', ?, '%')", substring)
	if err != nil {
		return nil, err
	}
	return convertMySQLRowsToBlogs(rows)
}

// Single SELECT query to blogs table with COUNT aggregation
func (dbConn *DBConn) numberOfBlogs(authorID int64) (int64, error) {
	//goland:noinspection ALL
	rows := dbConn.DB.QueryRow("SELECT COUNT(*) as count FROM blogs WHERE author = ?", authorID)

	var count int64

	if err := rows.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			// ignore throwing an error case where authorID doesn't exist in authors because
			// it would become a loophole to exploit. Example:
			//
			// for id in  0...100:
			//		if FIND-NUMBER-OF-BLOGS(id) throws an error:
			// 			continue
			//		else:
			//			< id is valid >
			//
			// therefore, by returning 0 you put the user in an ambiguous state where
			// he/she wouldn't know if a particular authorID actually exists
			//
			return 0, nil
		}
		return -1, fmt.Errorf("failed to query database")
	}

	return count, nil
}

// Single SELECT query to fetch a blog by its ID
func (dbConn *DBConn) fetchBlogByID(blogId int64) (Blog, error) {

	//goland:noinspection ALL
	row := dbConn.DB.QueryRow("SELECT * FROM blogs WHERE id = ?", blogId)

	var (
		blog   Blog
		images string // will be eventually unmarshalled to images:[]string inside blob
	)

	if err := row.Scan(&blog.ID, &blog.AuthorID, &blog.Content, &images, &blog.Title); err != nil {
		return Blog{}, err
	}

	// populate blog by exploding the images string
	if err := json.Unmarshal([]byte(images), &blog.Images); err != nil {
		return Blog{}, err
	}

	return blog, nil
}

// Single INSERT query to blogs table
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
