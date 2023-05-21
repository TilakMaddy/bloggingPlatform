package main

import "bloggingPlatform/internals"

func main() {

	var conn internals.DBConn
	var server internals.Server

	// open a tcp connection to the database
	conn.Setup()

	// start HTTP server locally and inject connection dependency
	server.Start(&conn)

}
