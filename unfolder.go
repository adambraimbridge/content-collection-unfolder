package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

const unfolderPath = "/unfold/{collectionType}/{uuid}"

type unfolder struct {
}

func newUnfolder() *unfolder {
	return &unfolder{}
}

func (u *unfolder) handle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	answer := fmt.Sprintf("Got called with type: %s and uuid %s", vars["collectionType"], vars["uuid"])
	w.Write([]byte(answer))
	w.WriteHeader(http.StatusOK)
}
