package internals

import (
	"fmt"
	"log"
	"net/http"
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

func (server *Server) Start(db *DBConn) {
	server.db = db
	http.HandleFunc("/upload", server.upload)
	http.Handle("/", http.FileServer(http.Dir("html")))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
