package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/vertica/vertica-sql-go"
)

var db *sql.DB

func main() {
	// Open a connection to the Vertica database
	connInfo := "vertica://dbadmin:db@min@192.168.50.179:5433/smartech"
	var err error
	db, err = sql.Open("vertica", connInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Create a new instance of a mux router
	r := mux.NewRouter()
	r.HandleFunc("/clients", handleClients).Methods("POST")

	// Start the HTTP server
	log.Fatal(http.ListenAndServe(":8080", r))
}

type Client struct {
	CID          int    `json:"cid"`
	PayloadParam string `json:"payload_param"`
	EventID      int    `json:"event_id"`
}

type RequestPayload struct {
	Action string `json:"action"`
	Client Client `json:"client"`
}

func handleClients(w http.ResponseWriter, r *http.Request) {
	var req RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch req.Action {
	case "create":
		createClient(w, req.Client)
	case "update":
		updateClient(w, req.Client)
	case "delete":
		deleteClient(w, req.Client.CID)
	case "list":
		getClients(w)
	case "get":
		getClient(w, req.Client.CID)
	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
	}
}

func createClient(w http.ResponseWriter, client Client) {
	// Check if client exists
	var exists int
	err := db.QueryRow("SELECT COUNT(*) FROM migration.ClientPayloadInfo WHERE cid=?", client.CID).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists > 0 {
		http.Error(w, "Client already exists", http.StatusConflict)
		return
	}

	// Insert the new client into the database
	_, err = db.Exec("INSERT INTO migration.ClientPayloadInfo (cid, payload_param, event_id) VALUES (?, ?, ?)", client.CID, client.PayloadParam, client.EventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success message
	fmt.Fprintf(w, "Client created successfully")
}

func updateClient(w http.ResponseWriter, client Client) {
	// Check if client exists
	var exists int
	err := db.QueryRow("SELECT COUNT(*) FROM migration.ClientPayloadInfo WHERE cid=?", client.CID).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists == 0 {
		http.Error(w, "Client does not exist", http.StatusNotFound)
		return
	}

	// Update the client in the database
	_, err = db.Exec("UPDATE migration.ClientPayloadInfo SET payload_param=?, event_id=? WHERE cid=?", client.PayloadParam, client.EventID, client.CID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success message
	fmt.Fprintf(w, "Client updated successfully")
}

func deleteClient(w http.ResponseWriter, cid int) {
	// Check if client exists
	var exists int
	err := db.QueryRow("SELECT COUNT(*) FROM migration.ClientPayloadInfo WHERE cid=?", cid).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists == 0 {
		http.Error(w, "Client does not exist", http.StatusNotFound)
		return
	}

	// Delete the client from the database
	_, err = db.Exec("DELETE FROM migration.ClientPayloadInfo WHERE cid=?", cid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success message
	fmt.Fprintf(w, "Client deleted successfully")
}

func getClients(w http.ResponseWriter) {
	// Your database query to retrieve all clients
	rows, err := db.Query("SELECT * FROM migration.ClientPayloadInfo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Iterate through rows and create a slice of Client objects
	clients := make([]Client, 0)
	for rows.Next() {
		var client Client
		if err := rows.Scan(&client.CID, &client.PayloadParam, &client.EventID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		clients = append(clients, client)
	}

	// Convert slice of clients to JSON and write to response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clients)
}

func getClient(w http.ResponseWriter, cid int) {
	// Your database query to retrieve client by CID
	row := db.QueryRow("SELECT * FROM migration.ClientPayloadInfo WHERE cid=?", cid)

	// Scan row into a Client object
	var client Client
	if err := row.Scan(&client.CID, &client.PayloadParam, &client.EventID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Convert client to JSON and write to response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(client)
}
