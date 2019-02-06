package main

import (
  "net/http"
  "github.com/gorilla/mux"
)

type LongPoll struct {
  ch (chan int)
}

type Endpoints struct {
  Rtr *mux.Router
  lp  *LongPoll
}

func NewEndPoints(r *mux.Router) *Endpoints {
  ep := &Endpoints{
    Rtr: r,
    lp: nil,
  }

  return ep
}

func (e *Endpoints) RegisterEndpoints() {
  e.Rtr.HandleFunc("/dashboard/test", e.testFn).Methods("GET")
  e.Rtr.HandleFunc("/dashboard/control", e.sendControl).Methods("GET")


  e.Rtr.HandleFunc("/device/olp", e.openLongPoll).Methods("GET")
}

func (e *Endpoints) setLongPoll(lp *LongPoll) {
  e.lp = lp
}