package controllers

import (
	"net/http"

	"github.com/alcoccoque/puzzle-solver-go/responses"
)

func (server *Server) Home(w http.ResponseWriter, r *http.Request) {
	responses.JSON(w, http.StatusOK, "Hello World!")

}
