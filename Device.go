package main

import (
	"encoding/json"
	"net/http"
)

func (e *Endpoints) openLongPoll(w http.ResponseWriter, r *http.Request) {

	ch := make(chan int)
	lp := &LongPoll{
		ch: ch,
	}

	e.setLongPoll(lp)

	controlCode := <-ch
	json.NewEncoder(w).Encode(controlCode)
}
