package internals

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	db *DBConn
}

func (server *Server) upload(w http.ResponseWriter, r *http.Request) {

	// convert form data in the request to Blog struct
	blog, err := convertReqFormToBlog(r)
	if err.IsError() {
		http.Error(w, err.message, err.statusCode)
		return
	}

	// publish the blog by writing it to the database
	publishErr := server.db.publishBlog(blog)
	if publishErr != nil {
		http.Error(w, "failed to publish to db", http.StatusInternalServerError)
		return
	}

	_, _ = fmt.Fprintf(w, "Uploaded succesfully !")

}

func (server *Server) blogCount(w http.ResponseWriter, r *http.Request) {

	var (
		authorID  int64 // GET request key to be supplied
		blogCount int64
		parseErr  error
		dbError   error
	)

	// extract author id from request
	if authorID, parseErr = strconv.ParseInt(r.FormValue("authorID"), 10, 64); parseErr != nil {
		http.Error(w, "invalid author ID supplied", http.StatusBadRequest)
		return
	}

	// ask the database about the number of blogs
	if blogCount, dbError = server.db.numberOfBlogs(authorID); dbError != nil {
		http.Error(w, dbError.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = fmt.Fprintf(w, "%d", blogCount)

}

func (server *Server) Start(db *DBConn) {
	server.db = db
	http.HandleFunc("/upload", server.upload)
	http.HandleFunc("/blog-count", server.blogCount)
	http.Handle("/", http.FileServer(http.Dir("html")))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
