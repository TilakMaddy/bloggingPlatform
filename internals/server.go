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

	// convert uploaded data to Blog struct
	blog, err := convertToBlog(w, r)
	if err != nil {
		return
	}

	// publish the blog by writing it to the database
	err = server.db.publishBlog(w, blog)
	if err != nil {
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
