package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/alcoccoque/puzzle-solver-go/api/models"
	"github.com/alcoccoque/puzzle-solver-go/api/responses"
	"github.com/alcoccoque/puzzle-solver-go/api/auth"
	"github.com/alcoccoque/puzzle-solver-go/api/utils/formaterror"
	"github.com/alcoccoque/puzzle-solver-go/api/solver"
)

func (server *Server) SolveMatrix(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	var solveMatrixSchema solver.SolveMatrix
	err = json.Unmarshal(body, &solveMatrixSchema)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	result, err := solver.FromListToState(solveMatrixSchema.Rows)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	state, ok := result["state"].(*models.FieldState)
	if !ok {
		responses.ERROR(w, http.StatusBadRequest, errors.New("Invalid state"))
		return
	}

	solver := solver.NewPuzzleSolver(state)
	solvedResult, err := solver.Solve()
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	solvedPuzzle, ok := solvedResult["solved_puzzle"].([][]int)
	if !ok {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("Failed to solve puzzle"))
		return
	}

	newMatrix := models.CreateMatrix{
		Coordinates: solvedPuzzle,
		UserID:      auth.ExtractTokenID(r),
	}

	matrix, err := models.CreateMatrix(server.DB, newMatrix)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	responses.JSON(w, http.StatusOK, matrix)
}

func (server *Server) GenerateMatrix(w http.ResponseWriter, r *http.Request) {
	size, err := strconv.Atoi(r.URL.Query().Get("size"))
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	filledPercentage, err := strconv.ParseFloat(r.URL.Query().Get("filled_percentage"), 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	generator := solver.NewPuzzleGenerator(size)
	board, err := solver.GeneratePuzzle(filledPercentage)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	newMatrix := models.CreateMatrix{
		Coordinates: board,
		UserID:      auth.ExtractTokenID(r),
	}

	matrix, err := models.CreateMatrix(server.DB, newMatrix)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	responses.JSON(w, http.StatusOK, matrix)
}
