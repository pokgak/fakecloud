package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"

	"fakecloud/database"
	"fakecloud/handlers/vm"
)

func main() {
	// Open database connection
	db, err := sql.Open("sqlite3", "./fakecloud.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create users table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS virtual_machines (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
		instance_type TEXT NOT NULL
    )`)
	if err != nil {
		log.Fatal(err)
	}

	// Set database connection for handlers
	database.SetDB(db)

	// Create router
	router := mux.NewRouter()

	// VirtualMachine
	router.HandleFunc("/vms", vm.CreateVirtualMachine).Methods("POST")
	router.HandleFunc("/vms", vm.GetVirtualMachines).Methods("GET")
	router.HandleFunc("/vms/{id}", vm.GetVirtualMachine).Methods("GET")
	router.HandleFunc("/vms/{id}", vm.UpdateVirtualMachine).Methods("PUT")
	router.HandleFunc("/vms/{id}", vm.DeleteVirtualMachine).Methods("DELETE")

	// Start server
	log.Fatal(http.ListenAndServe(":8000", router))
}
