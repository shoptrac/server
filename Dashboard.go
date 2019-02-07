package main

import (
  "encoding/json"
  "net/http"
)

func (e *Endpoints) sendControl(w http.ResponseWriter, r *http.Request) {
  if e.lp != nil {
    e.lp.ch <- 1
    json.NewEncoder(w).Encode("Successfully closed")
  } else {
    json.NewEncoder(w).Encode("No LP open")
  }
}