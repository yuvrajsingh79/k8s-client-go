package main

import (
	"client-go/k8s-client-go/controller"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/getPods", controller.GetPodList).Methods("GET")
	log.Fatal(http.ListenAndServe(":2000", r))
}
