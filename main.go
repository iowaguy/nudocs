package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	myRouter := mux.NewRouter()
	myRouter.HandleFunc("/operations/generate", operationReceiver).Methods("POST")
	myRouter.HandleFunc("/operations/receive", operationSender).Methods("GET")
	http.Handle("/", myRouter)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func operationReceiver(w http.ResponseWriter, r *http.Request) {
	var o Operation

	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		log.Fatal(err)
	}

	w.Write([]byte("{\"status\":\"operation received\"}"))
}

func operationSender(w http.ResponseWriter, r *http.Request) {
	var o Operation
	o.OpType = "insert"
	o.Character = "x"
	o.Position = 5
	o.Priority = 1

	json.NewEncoder(w).Encode(o)
}

type Operation struct {
	OpType    string `json:"opType"`
	Character string `json:"character"`
	Position  int    `json:"position,string"`
	Priority  int    `json:"priority,string"`
}
