package main

import (
  "log"
  "net/http"

  "github.com/gorilla/mux"
)

func main() {

  r := mux.NewRouter()
  ep := NewEndPoints(r)

  ep.RegisterEndpoints()

  log.Fatal(http.ListenAndServe(":8000", r))
}