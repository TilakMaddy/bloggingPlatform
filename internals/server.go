package internals

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Server struct {
	db *DBConn
}

func (server *Server) deleteBlogById(w http.ResponseWriter, r *http.Request) {
	blogId, err := strconv.ParseInt(r.FormValue("id"), 10, 64)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	deleteErr := server.db.deleteBlogByID(blogId)
	if deleteErr != nil {
		http.Error(w, deleteErr.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	// anything written to w, will henceforth be ignored by clients
	// because http.StatusNoContent says 'I have nothing to give you'
}

func (server *Server) getBlogsByAuthor(w http.ResponseWriter, r *http.Request) {
	authorID, err := strconv.ParseInt(r.FormValue("id"), 10, 64)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	blogs, searchErr := server.db.getBlogsByAuthor(authorID)
	if searchErr != nil {
		http.Error(w, searchErr.Error(), http.StatusInternalServerError)
		return
	}

	blogBytes, err := json.Marshal(blogs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// It is important to set the header before you WriteHeader
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if len(blogs) == 0 {
		_, _ = w.Write([]byte("0"))
	} else {
		_, _ = w.Write(blogBytes)
	}

}

func (server *Server) search(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.FormValue("q")

	blogs, searchErr := server.db.searchBlogs(searchTerm)
	if searchErr != nil {
		http.Error(w, searchErr.Error(), http.StatusInternalServerError)
		return
	}

	blogBytes, err := json.Marshal(blogs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// It is important to set the header before you WriteHeader
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(blogBytes)
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

	w.WriteHeader(http.StatusCreated)
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

	blogCountBytes, err := json.Marshal(blogCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(blogCountBytes)

}

func (server *Server) Start(db *DBConn) {
	server.db = db

	http.HandleFunc("/api/delete-blog", server.deleteBlogById) // ?id=3
	http.HandleFunc("/api/author", server.getBlogsByAuthor)    // ?id=3
	http.HandleFunc("/api/search", server.search)              // ?q=search_term
	http.HandleFunc("/api/blog-count", server.blogCount)       // ?authorID=2
	http.HandleFunc("/api/upload", server.upload)

	http.Handle("/images/",
		http.StripPrefix(
			"/images/",
			http.FileServer(http.Dir(os.Getenv("UPLOAD_DIR")))))

	http.Handle("/", http.FileServer(http.Dir("html")))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
