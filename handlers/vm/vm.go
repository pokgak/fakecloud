package vm

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pokgak/fakecloud/sdk"

	"fakecloud/database"
)

func CreateVirtualMachine(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request for %s with headers %v", r.Method, r.URL.Path, r.Header)

	// Parse request body
	var vm sdk.VirtualMachine
	err := json.NewDecoder(r.Body).Decode(&vm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert vm into database
	result, err := database.GetDB().Exec("INSERT INTO virtual_machines (name, instance_type) VALUES (?, ?)", vm.Name, vm.InstanceType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get ID of inserted vm
	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vm.ID = int(id)

	// Return success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(vm)
}

func GetVirtualMachines(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request for %s with headers %v", r.Method, r.URL.Path, r.Header)

	// Get all virtual_machines from database
	rows, err := database.GetDB().Query("SELECT id, name, instance_type FROM virtual_machines")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Create slice of virtual_machines
	var virtual_machines []sdk.VirtualMachine

	// Iterate over rows and add virtual_machines to slice
	for rows.Next() {
		var vm sdk.VirtualMachine
		err := rows.Scan(&vm.ID, &vm.Name, &vm.InstanceType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		virtual_machines = append(virtual_machines, vm)
	}

	// Return virtual_machines as JSON response
	json.NewEncoder(w).Encode(virtual_machines)
}

func GetVirtualMachine(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request for %s with headers %v", r.Method, r.URL.Path, r.Header)

	// Get vm ID from URL parameter
	vars := mux.Vars(r)
	id := vars["id"]

	// Get vm from database
	var vm sdk.VirtualMachine
	err := database.GetDB().QueryRow("SELECT id, name, instance_type FROM virtual_machines WHERE id = ?", id).Scan(&vm.ID, &vm.Name, &vm.InstanceType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Return vm as JSON response
	json.NewEncoder(w).Encode(vm)
}

func UpdateVirtualMachine(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request for %s with headers %v", r.Method, r.URL.Path, r.Header)

	// Get vm ID from URL parameter
	vars := mux.Vars(r)
	id := vars["id"]

	// Parse request body
	var vm sdk.VirtualMachine
	err := json.NewDecoder(r.Body).Decode(&vm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update vm in database
	_, err = database.GetDB().Exec("UPDATE virtual_machines SET name = ?, instance_type = ? WHERE id = ?", vm.Name, vm.InstanceType, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vm)
}

func DeleteVirtualMachine(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request for %s with headers %v", r.Method, r.URL.Path, r.Header)

	// Get vm ID from URL parameter
	vars := mux.Vars(r)
	id := vars["id"]

	// Delete vm from database
	_, err := database.GetDB().Exec("DELETE FROM virtual_machines WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
}
