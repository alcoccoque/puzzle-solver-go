package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type PuzzleGenerator struct {
	size int
}

func NewPuzzleGenerator(size int) *PuzzleGenerator {
	return &PuzzleGenerator{size: size}
}

func (pg *PuzzleGenerator) GeneratePuzzle(filledPercentage float64) ([][]int, error) {
	solvedPuzzle, err := pg.SolvePuzzle()
	if err != nil {
		return nil, err
	}

	totalCells := pg.size * pg.size
	filledCellsCount := int(float64(totalCells) * filledPercentage)
	var filledCells []Cell
	for x := 0; x < pg.size; x++ {
		for y := 0; y < pg.size; y++ {
			if solvedPuzzle[x][y] != 0 {
				filledCells = append(filledCells, Cell{x, y})
			}
		}
	}

	rand.Seed(time.Now().UnixNano())
	cellsToReset := randomSample(filledCells, len(filledCells)-filledCellsCount)
	for _, cell := range cellsToReset {
		solvedPuzzle[cell.X][cell.Y] = 0
	}

	return solvedPuzzle, nil
}

func (pg *PuzzleGenerator) SolvePuzzle() ([][]int, error) {
	field := make([][]int, pg.size)
	for i := range field {
		field[i] = make([]int, pg.size)
	}

	rand.Seed(time.Now().UnixNano())
	x, y := rand.Intn(pg.size), rand.Intn(pg.size)
	value := rand.Intn(pg.size) + 1
	field[x][y] = value

	result, err := FromListToState(field)
	if err != nil {
		return nil, err
	}

	fieldState := result["state"].(*FieldState)
	solver := NewPuzzleSolver(fieldState)
	solution, err := solver.Solve()
	if err != nil {
		return nil, err
	}

	return solution["solved_puzzle"].([][]int), nil
}

// randomSample returns a random sample of n elements from the slice.
func randomSample(cells []Cell, n int) []Cell {
	if n > len(cells) {
		n = len(cells)
	}
	rand.Seed(time.Now().UnixNano())
	perm := rand.Perm(len(cells))
	sample := make([]Cell, n)
	for i := 0; i < n; i++ {
		sample[i] = cells[perm[i]]
	}
	return sample
}

func main() {
	pg := NewPuzzleGenerator(5)
	puzzle, err := pg.GeneratePuzzle(0.5)
	if err != nil {
		fmt.Println("Error generating puzzle:", err)
		return
	}
	for _, row := range puzzle {
		fmt.Println(row)
	}
}
