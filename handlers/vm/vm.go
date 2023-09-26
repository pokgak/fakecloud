package vm

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"fakecloud/database"
)

type VirtualMachine struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	InstanceType string `json:"instance_type"`
}

func CreateVirtualMachine(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var vm VirtualMachine
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
	// Get all virtual_machines from database
	rows, err := database.GetDB().Query("SELECT id, name, instance_type FROM virtual_machines")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Create slice of virtual_machines
	var virtual_machines []VirtualMachine

	// Iterate over rows and add virtual_machines to slice
	for rows.Next() {
		var vm VirtualMachine
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
	// Get vm ID from URL parameter
	vars := mux.Vars(r)
	id := vars["id"]

	// Get vm from database
	var vm VirtualMachine
	err := database.GetDB().QueryRow("SELECT id, name, instance_type FROM virtual_machines WHERE id = ?", id).Scan(&vm.ID, &vm.Name, &vm.InstanceType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Return vm as JSON response
	json.NewEncoder(w).Encode(vm)
}

func UpdateVirtualMachine(w http.ResponseWriter, r *http.Request) {
	// Get vm ID from URL parameter
	vars := mux.Vars(r)
	id := vars["id"]

	// Parse request body
	var vm VirtualMachine
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
