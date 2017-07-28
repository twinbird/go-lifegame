package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

const (
	// FieldColSize is screen column size
	FieldColSize = 40
	// FieldRowSize is screen row size
	FieldRowSize = 25
)

var patternMap map[string][][]bool

var gliderPattern = [][]bool{
	{true, true, true},
	{true, false, false},
	{false, true, false},
}

var blockPattern = [][]bool{
	{true, true},
	{true, true},
}

var honeycombPattern = [][]bool{
	{false, false, true, true, false, false},
	{true, false, false, false, false, true},
	{true, false, false, false, false, true},
	{false, false, true, true, false, false},
}

func init() {
	patternMap = make(map[string][][]bool, 0)

	patternMap["glider"] = gliderPattern
	patternMap["block"] = blockPattern
	patternMap["honeycomb"] = honeycombPattern
}

func printPatterns() {
	fmt.Println("Supported pattern is follow.")
	for k := range patternMap {
		fmt.Println(" * " + k)
	}
}

func main() {
	var pattern string

	flag.StringVar(&pattern, "pattern", "glider", "preset pattern")
	flag.Usage = func() {
		fmt.Println("Usage:")
		flag.PrintDefaults()
		printPatterns()
		os.Exit(0)
	}
	flag.Parse()

	if _, ok := patternMap[pattern]; ok == false {
		fmt.Printf("pattern '%s' isn't supported", pattern)
		os.Exit(1)
	}

	sc, err := newScreen(FieldColSize, FieldRowSize)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	sc.setPattern(patternMap[pattern], 10, 10)

	if err := gameStart(sc); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func gameStart(sc *screen) error {
	var err error

	t := time.Tick(time.Second)
	for {
		sc.rendering()
		sc, err = sc.nextGen()
		if err != nil {
			return err
		}
		<-t
	}
}

type screen struct {
	RowsCount int
	ColsCount int
	Field     [][]bool // true: live, false: dead
}

func newScreen(colSize, rowSize int) (*screen, error) {
	if colSize <= 0 {
		return nil, fmt.Errorf("colSize must be greater than 0")
	}
	if rowSize <= 0 {
		return nil, fmt.Errorf("rowSize must be greater than 0")
	}
	rows := make([][]bool, rowSize)

	for i := range rows {
		cols := make([]bool, colSize)
		rows[i] = cols
	}
	return &screen{
		RowsCount: rowSize,
		ColsCount: colSize,
		Field:     rows,
	}, nil
}

func (sc *screen) rendering() {
	for _, row := range sc.Field {
		for _, col := range row {
			if col == true {
				// living
				fmt.Print("■")
			} else {
				// dead
				fmt.Print("□")
			}
		}
		fmt.Println()
	}
}

func (sc *screen) checkPattern(pattern [][]bool, x, y int) error {
	// row size check
	rowSize := len(pattern)
	if rowSize <= 0 {
		return fmt.Errorf("pattern row size must be greater than 1")
	}
	// column size check
	colSize := len(pattern[0])
	for _, cols := range pattern {
		if colSize != len(cols) {
			return fmt.Errorf("pattern column size should be same each row")
		}
	}
	// field size and position check
	if len(sc.Field)+y < rowSize {
		return fmt.Errorf("pattern row size couldn't into field")
	}
	if len(sc.Field[0])+x < colSize {
		return fmt.Errorf("pattern column size couldn't into field")
	}

	return nil
}

func (sc *screen) setPattern(pattern [][]bool, x, y int) error {
	if err := sc.checkPattern(pattern, x, y); err != nil {
		return err
	}
	for i, row := range pattern {
		for k, col := range row {
			sc.Field[i+y][k+x] = col
		}
	}
	return nil
}

func (sc *screen) isSafeIdx(x, y int) bool {
	maxRow := len(sc.Field)
	maxCol := len(sc.Field[0])

	if x < 0 || y < 0 {
		return false
	}

	if maxRow <= y || maxCol <= x {
		return false
	}
	return true
}

func (sc *screen) countLivingNeighbor(x, y int) int {
	var counter int
	// left top
	if sc.isSafeIdx(x-1, y-1) && sc.Field[y-1][x-1] == true {
		counter++
	}
	// top
	if sc.isSafeIdx(x, y-1) && sc.Field[y-1][x] == true {
		counter++
	}
	// right top
	if sc.isSafeIdx(x+1, y-1) && sc.Field[y-1][x+1] == true {
		counter++
	}
	// left
	if sc.isSafeIdx(x-1, y) && sc.Field[y][x-1] == true {
		counter++
	}
	// right
	if sc.isSafeIdx(x+1, y) && sc.Field[y][x+1] == true {
		counter++
	}
	// left bottom
	if sc.isSafeIdx(x-1, y+1) && sc.Field[y+1][x-1] == true {
		counter++
	}
	// bottom
	if sc.isSafeIdx(x, y+1) && sc.Field[y+1][x] == true {
		counter++
	}
	// right bottom
	if sc.isSafeIdx(x+1, y+1) && sc.Field[y+1][x+1] == true {
		counter++
	}
	return counter
}

func (sc *screen) nextCellState(x, y int) bool {
	n := sc.countLivingNeighbor(x, y)
	// birth
	if n == 3 {
		return true
	}
	// living
	if n == 2 || n == 3 {
		return true
	}
	// depopulation
	if n <= 1 {
		return false
	}
	// overcrowding
	if 4 <= n {
		return false
	}
	panic("couldn't determine next status")
}

func (sc *screen) nextGen() (*screen, error) {
	next, err := newScreen(sc.ColsCount, sc.RowsCount)
	if err != nil {
		return nil, err
	}

	for y, rows := range sc.Field {
		for x := range rows {
			next.Field[y][x] = sc.nextCellState(x, y)
		}
	}
	return next, nil
}
