package solver

import (
	"errors"
	"fmt"
)

type Cell struct {
	X, Y int
}

type Field struct {
	size          int
	neighborCache map[Cell]map[Cell]struct{}
}

func NewField(size int) (*Field, error) {
	if err := checkSize(size); err != nil {
		return nil, err
	}
	return &Field{
		size:          size,
		neighborCache: make(map[Cell]map[Cell]struct{}),
	}, nil
}

func checkSize(size int) error {
	if size < 2 {
		return errors.New("minimum field size is 2")
	}
	return nil
}

func (f *Field) Size() int {
	return f.size
}

func (f *Field) GetAllCells() []Cell {
	var cells []Cell
	for x := 0; x < f.size; x++ {
		for y := 0; y < f.size; y++ {
			cells = append(cells, Cell{x, y})
		}
	}
	return cells
}

func (f *Field) GetNeighbourCells(cell Cell) map[Cell]struct{} {
	if neighbors, found := f.neighborCache[cell]; found {
		return neighbors
	}
	directions := []Cell{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	neighbors := make(map[Cell]struct{})
	for _, d := range directions {
		neighbor := Cell{cell.X + d.X, cell.Y + d.Y}
		if neighbor.X >= 0 && neighbor.X < f.size && neighbor.Y >= 0 && neighbor.Y < f.size {
			neighbors[neighbor] = struct{}{}
		}
	}
	f.neighborCache[cell] = neighbors
	return neighbors
}

type FieldState struct {
	field *Field
	state map[Cell]int
}

func NewFieldState(field *Field) *FieldState {
	return &FieldState{
		field: field,
		state: make(map[Cell]int),
	}
}

func FromListToState(matrix [][]int) (map[string]interface{}, error) {
	size := len(matrix)
	field, err := NewField(size)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}, err
	}
	state := NewFieldState(field)

	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			state.SetState(Cell{x, y}, matrix[x][y])
		}
	}
	return map[string]interface{}{"state": state}, nil
}

func (fs *FieldState) ToList() [][]int {
	size := fs.field.Size()
	result := make([][]int, size)
	for x := 0; x < size; x++ {
		row := make([]int, size)
		for y := 0; y < size; y++ {
			row[y] = fs.state[Cell{x, y}]
		}
		result[x] = row
	}
	return result
}

func (fs *FieldState) SetState(coords Cell, value int) {
	fs.state[coords] = value
}

func (fs *FieldState) GetState(coords Cell) int {
	return fs.state[coords]
}

func (fs *FieldState) GetInvolved(cell Cell) []Cell {
	involved := []Cell{cell}
	value := fs.state[cell]
	notChecked := []Cell{cell}
	checked := map[Cell]struct{}{cell: {}}

	for len(notChecked) > 0 {
		cell = notChecked[len(notChecked)-1]
		notChecked = notChecked[:len(notChecked)-1]

		for neighbor := range fs.field.GetNeighbourCells(cell) {
			if _, ok := checked[neighbor]; !ok && fs.state[neighbor] == value {
				involved = append(involved, neighbor)
				notChecked = append(notChecked, neighbor)
				checked[neighbor] = struct{}{}
			}
		}
	}
	return involved
}

type CellsGroup struct {
	value                  int
	initialCells           []Cell
	possibleCells          []Cell
	possibleConnectionCells []Cell
}

func NewCellsGroup(value int, initialCells []Cell) *CellsGroup {
	return &CellsGroup{
		value:        value,
		initialCells: initialCells,
	}
}

func (cg *CellsGroup) GetValue() int {
	return cg.value
}

func (cg *CellsGroup) GetPossibleLength() int {
	return len(cg.initialCells) + len(cg.possibleCells) + len(cg.possibleConnectionCells)
}

func (cg *CellsGroup) AddPossibleCell(cell Cell) {
	cg.possibleCells = append(cg.possibleCells, cell)
}

func (cg *CellsGroup) AddConnection(cell Cell) {
	cg.possibleConnectionCells = append(cg.possibleConnectionCells, cell)
}

type PuzzleSolver struct {
	possibleValues map[Cell][]int
	involved       []Cell
	unfilledGroups map[Cell]*CellsGroup
	fieldState     *FieldState
	stateChanged   bool
}

func NewPuzzleSolver(fieldState *FieldState) *PuzzleSolver {
	return &PuzzleSolver{
		fieldState:     fieldState,
		stateChanged:   true,
		unfilledGroups: make(map[Cell]*CellsGroup),
		possibleValues: make(map[Cell][]int),
	}
}

func (ps *PuzzleSolver) Solve() (map[string]interface{}, error) {
	if err := ps.refreshState(); err != nil {
		return map[string]interface{}{"error": err.Error()}, err
	}
	ps.tryFillEmptyCells()
	if ps.checkForZeros() {
		return map[string]interface{}{"error": "Puzzle is unsolvable"}, errors.New("puzzle is unsolvable")
	}
	return map[string]interface{}{"solved_puzzle": ps.fieldState.ToList()}, nil
}

func (ps *PuzzleSolver) refreshState() error {
	ps.findUnfilledGroups()
	for _, cell := range ps.fieldState.field.GetAllCells() {
		if ps.fieldState.GetState(cell) != 0 {
			if _, ok := ps.unfilledGroups[cell]; ok {
				ps.findPossibleValues(cell)
			}
		}
	}

	var emptyCells []Cell
	for _, cell := range ps.fieldState.field.GetAllCells() {
		if ps.fieldState.GetState(cell) == 0 {
			emptyCells = append(emptyCells, cell)
		}
	}

	involved := make(map[Cell]struct{})
	for _, cell := range emptyCells {
		allNonOne := true
		for neighbor := range ps.fieldState.field.GetNeighbourCells(cell) {
			if ps.fieldState.GetState(neighbor) == 1 {
				allNonOne = false
				break
			}
		}
		if allNonOne {
			ps.addPossibleValue(cell, 1)
		}

		if _, ok := involved[cell]; !ok {
			emptyGroup := ps.fieldState.GetInvolved(cell)
			for _, e := range emptyGroup {
				involved[e] = struct{}{}
			}
			ps.findAdditionalValues(emptyGroup)
		}
	}

	ps.stateChanged = true
	return nil
}

func (ps *PuzzleSolver) findAdditionalValues(emptyGroup []Cell) {
	for value := 2; value <= min(len(emptyGroup)+1, 10); value++ {
		involvedForValue := make(map[Cell]struct{})
		for _, cell := range emptyGroup {
			involvedForValue[cell] = struct{}{}
		}
		notPossibleCells := make(map[Cell]struct{})

		for cell := range involvedForValue {
			for neighbor := range ps.fieldState.field.GetNeighbourCells(cell) {
				if ps.fieldState.GetState(neighbor) == value {
					notPossibleCells[cell] = struct{}{}
				}
			}
		}
		for cell := range notPossibleCells {
			delete(involvedForValue, cell)
		}

		if value <= len(involvedForValue) {
			for cell := range involvedForValue {
				ps.addPossibleValue(cell, value)
			}
		}
	}
}

func (ps *PuzzleSolver) findUnfilledGroups() {
	ps.unfilledGroups = make(map[Cell]*CellsGroup)
	ps.involved = nil
	ps.possibleValues = make(map[Cell][]int)

	for _, cell := range ps.fieldState.field.GetAllCells() {
		if ps.fieldState.GetState(cell) != 0 {
			if _, ok := ps.involved[cell]; !ok {
				initialCells := ps.fieldState.GetInvolved(cell)
				ps.involved = append(ps.involved, initialCells...)
				value := ps.fieldState.GetState(cell)

				if len(initialCells) < value {
					newGroup := NewCellsGroup(value, initialCells)
					for _, c := range initialCells {
						ps.unfilledGroups[c] = newGroup
					}
				}

				if len(initialCells) > value {
					panic("wrong group size")
				}
			}
		}
	}
}

func (ps *PuzzleSolver) findPossibleValues(cell Cell) {
	var nextCells []Cell
	nextCells = append(nextCells, cell)
	value := ps.fieldState.GetState(cell)
	var previousCells []Cell
	group := ps.unfilledGroups[cell]

	for len(nextCells) > 0 {
		currentCell := nextCells[len(nextCells)-1]
		nextCells = nextCells[:len(nextCells)-1]
		previousCells = append(previousCells, currentCell)
		currentLength := len(previousCells)

		wayLength := currentLength + len(group.initialCells)

		if wayLength < value && currentCell != cell {
			ps.addPossibleValue(currentCell, value)
			group.AddPossibleCell(currentCell)
		} else if wayLength == value {
			ps.addPossibleValue(currentCell, value)
			group.AddPossibleCell(currentCell)
			continue
		}

		freeNeighbours := make(map[Cell]struct{})
		for neighbor := range ps.fieldState.field.GetNeighbourCells(currentCell) {
			if ps.fieldState.GetState(neighbor) == 0 {
				freeNeighbours[neighbor] = struct{}{}
			}
		}

		for neighbor := range freeNeighbours {
			if ps.connectionCellsFound(cell, neighbor, group) {
				continue
			}
			nextCells = append(nextCells, neighbor)
		}
	}
}

func (ps *PuzzleSolver) connectionCellsFound(neighbor Cell, cell Cell, group *CellsGroup) bool {
	value := group.GetValue()

	if ps.fieldState.GetState(neighbor) == value && !contains(group.initialCells, neighbor) && !contains(group.possibleCells, neighbor) {
		ps.fieldState.SetState(cell, value)
		intersectionLength := len(ps.fieldState.GetInvolved(cell))
		ps.fieldState.SetState(cell, 0)

		if intersectionLength <= value {
			group.AddConnection(cell)
			ps.addPossibleValue(cell, value)
		}
		return true
	}
	return false
}

func (ps *PuzzleSolver) addPossibleValue(cell Cell, value int) {
	for _, n := range ps.fieldState.field.GetNeighbourCells(cell) {
		if nValue := ps.fieldState.GetState(n); nValue != 0 && abs(nValue-value) == 1 {
			return
		}
	}
	ps.possibleValues[cell] = append(ps.possibleValues[cell], value)
}

func (ps *PuzzleSolver) tryFillEmptyCells() {
	var freeCells []Cell
	for _, cell := range ps.fieldState.field.GetAllCells() {
		if ps.fieldState.GetState(cell) == 0 {
			freeCells = append(freeCells, cell)
		}
	}
	possibleValues := ps.possibleValues
	backtrack(freeCells, possibleValues, ps.fieldState)
}

func backtrack(freeCells []Cell, possibleValues map[Cell][]int, fieldState *FieldState) bool {
	if len(freeCells) == 0 {
		return true
	}
	cell := freeCells[len(freeCells)-1]
	freeCells = freeCells[:len(freeCells)-1]
	for _, value := range possibleValues[cell] {
		fieldState.SetState(cell, value)
		if checkGroupSize(fieldState) && backtrack(freeCells, possibleValues, fieldState) {
			return true
		}
		fieldState.SetState(cell, 0)
	}
	return false
}

func checkGroupSize(fieldState *FieldState) bool {
	groups := make(map[*CellsGroup]struct{})
	for _, cell := range fieldState.field.GetAllCells() {
		value := fieldState.GetState(cell)
		if value == 0 {
			continue
		}
		group := fieldState.GetInvolved(cell)
		if len(group) != value {
			return false
		}
		groups[NewCellsGroup(value, group)] = struct{}{}
	}
	for group := range groups {
		if group.GetPossibleLength() < group.GetValue() && len(group.possibleConnectionCells) == 0 {
			return false
		}
	}
	return true
}

func (ps *PuzzleSolver) checkForZeros() bool {
	for _, row := range ps.fieldState.ToList() {
		for _, val := range row {
			if val == 0 {
				return true
			}
		}
	}
	return false
}

// Helper functions
func contains(cells []Cell, cell Cell) bool {
	for _, c := range cells {
		if c == cell {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func main() {
	// Test code for the puzzle solver would go here.
	matrix := [][]int{
		{0, 0, 0},
		{0, 0, 0},
		{0, 0, 0},
	}
	result, err := FromListToState(matrix)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fieldState := result["state"].(*FieldState)
	solver := NewPuzzleSolver(fieldState)
	solution, err := solver.Solve()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Solved Puzzle:", solution["solved_puzzle"])
}
