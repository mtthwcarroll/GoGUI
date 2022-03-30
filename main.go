package main

import (
	"crypto/rand"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel-examples/community/maze/stack"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"math/big"
)

var cells = 10     //how many cells in a row. actual number is cells^2
var cellWidth = 50 //how many pixels wide the cells are

// drawing function for the window
func drawGrid(grid [][]*cell, imd *imdraw.IMDraw) {
	cellWidth := float64(cellWidth)

	// draw each cell
	for x := 0; x < len(grid); x++ {
		for y := 0; y < len(grid[0]); y++ {
			xw := cellWidth * float64(x)
			yw := cellWidth * float64(y)
			poi := grid[x][y]

			// first fill the cells
			imd.Color = colornames.White
			if x == len(grid)-1 && y == len(grid[0])-1 {
				imd.Color = colornames.Gold
			}

			imd.Push(pixel.V(xw, yw), pixel.V(xw+cellWidth, yw+cellWidth))
			imd.Rectangle(0)

			// now we draw the walls
			imd.Color = colornames.Black
			if poi.walls[0] {
				imd.Push(pixel.V(xw, yw), pixel.V(xw+cellWidth, yw))
				imd.Line(4)
			}
			if poi.walls[1] {
				imd.Push(pixel.V(xw+cellWidth, yw), pixel.V(xw+cellWidth, yw+cellWidth))
				imd.Line(4)
			}
			if poi.walls[2] {
				imd.Push(pixel.V(xw, yw+cellWidth), pixel.V(xw+cellWidth, yw+cellWidth))
				imd.Line(4)
			}
			if poi.walls[3] {
				imd.Push(pixel.V(xw, yw), pixel.V(xw, yw+cellWidth))
				imd.Line(4)
			}
		}
	}
}

// taken from maze-generator pixel demo.
// is a structure for data storage
type cell struct {
	walls [4]bool // Wall order: top, right, bottom, left

	row     int
	col     int
	visited bool
}

// taken from maze-generator pixel demo. Is pretty much a
// constructor for a cell.
func newCell(row int, col int) *cell {
	newCell := new(cell)
	newCell.row = row
	newCell.col = col

	for i := range newCell.walls {
		newCell.walls[i] = true
	}
	return newCell
}

// generates a maze using the same algorithm as maze-generator.
func generateMaze(cols int, rows int) [][]*cell {
	trace := stack.NewStack(cols * rows) // backtracing stack
	var grid = make([][]*cell, cols)     // holds cell references
	for i := range grid {
		grid[i] = make([]*cell, rows)
	}
	for x := 0; x < rows; x++ { // init cells
		for y := 0; y < cols; y++ {
			grid[x][y] = newCell(x, y)
		}
	}

	// while loop that creates the maze. Algorithm goes as follows:
	// take a random neighbor of current cell that hasn't been visited and destroy
	// the walls between. Neighbor is now current cell and previous goes into the stack.
	// If there are no unvisited neighbors pop the stack and try again. Ends when stack is
	// 0 and there are no unvisited neighbors.
	working := true
	currentCell := grid[0][0] // starting cell
	for working {
		neighbor := getRandomNeighbor(currentCell, grid, &cols, &rows)

		if neighbor != nil {
			trace.Push(currentCell)
			removeWalls(currentCell, neighbor)
			currentCell.visited = true
			currentCell = neighbor
		} else if trace.Len() > 0 {
			if !currentCell.visited {
				currentCell.visited = true
			}
			currentCell = trace.Pop().(*cell)
		} else if neighbor == nil && trace.Len() == 0 {
			working = false
		}
	}

	return grid
}

// taken from maze-generator because I couldn't think of a better
// way to write the same thing.
func removeWalls(a *cell, b *cell) {
	x := a.row - b.row

	if x == 1 {
		a.walls[3] = false
		b.walls[1] = false
	} else if x == -1 {
		a.walls[1] = false
		b.walls[3] = false
	}

	y := a.col - b.col

	if y == 1 {
		a.walls[0] = false
		b.walls[2] = false
	} else if y == -1 {
		a.walls[2] = false
		b.walls[0] = false
	}
}

// function to change the player position color
func (cell *cell) highlightAsPlayer(imd *imdraw.IMDraw) {
	cellWidth := float64(cellWidth)
	x := float64(cell.row) * cellWidth
	y := float64(cell.col) * cellWidth

	imd.Color = colornames.Green
	imd.Push(pixel.V(x+(cellWidth*.2), y+(cellWidth*.2)), pixel.V(x+cellWidth-(cellWidth*.2), y+cellWidth-(cellWidth*.2)))
	imd.Rectangle(0)
}

// get a random unvisited neighbor from the passed cell poi
func getRandomNeighbor(poi *cell, grid [][]*cell, cols *int, rows *int) *cell {
	var neighbors []*cell
	x := poi.row
	y := poi.col

	if x-1 >= 0 {
		if !grid[x-1][y].visited {
			neighbors = append(neighbors, grid[x-1][y])
		}
	}
	if y-1 >= 0 {
		if !grid[x][y-1].visited {
			neighbors = append(neighbors, grid[x][y-1])
		}
	}
	if x+1 < *rows {
		if !grid[x+1][y].visited {
			neighbors = append(neighbors, grid[x+1][y])
		}
	}
	if y+1 < *cols {
		if !grid[x][y+1].visited {
			neighbors = append(neighbors, grid[x][y+1])
		}
	}

	if len(neighbors) > 0 {
		// took this from maze-generator because my random number generator was being predictable
		big, _ := rand.Int(rand.Reader, big.NewInt(int64(len(neighbors))))
		randIndex := big.Int64()
		return neighbors[randIndex]
	} else {
		return nil
	}
}

// run for the pixel window
func run() {
	//configs for window
	config := pixelgl.WindowConfig{
		Title:  "a-maze-ing game!",
		Bounds: pixel.R(0, 0, float64(cells*cellWidth), float64(cells*cellWidth)),
	}

	// make the maze
	grid := generateMaze(cells, cells)
	playerPOS := []int{0, 0} //player starts at bottom left corner (0,0)

	win, _ := pixelgl.NewWindow(config) // make the window

	imDraw := imdraw.New(nil) // make the drawer

	// while the window hasn't been closed this will run
	for !win.Closed() {
		// controls and validation for user input
		if win.JustReleased(pixelgl.KeyW) {
			if playerPOS[1]+1 < cells && !grid[playerPOS[0]][playerPOS[1]].walls[2] {
				playerPOS[1] += 1
			}
		}
		if win.JustReleased(pixelgl.KeyD) {
			if playerPOS[0]+1 < cells && !grid[playerPOS[0]][playerPOS[1]].walls[1] {
				playerPOS[0] += 1
			}
		}
		if win.JustReleased(pixelgl.KeyS) {
			if playerPOS[1] > 0 && !grid[playerPOS[0]][playerPOS[1]].walls[0] {
				playerPOS[1] -= 1
			}
		}
		if win.JustReleased(pixelgl.KeyA) {
			if playerPOS[0] > 0 && !grid[playerPOS[0]][playerPOS[1]].walls[3] {
				playerPOS[0] -= 1
			}
		}

		// re draw the grid
		imDraw.Clear()
		drawGrid(grid, imDraw)
		grid[playerPOS[0]][playerPOS[1]].highlightAsPlayer(imDraw)
		imDraw.Draw(win)
		win.Update()
		// win condition
		if playerPOS[0] == cells-1 && playerPOS[1] == cells-1 {
			grid = generateMaze(cells, cells)
			playerPOS = []int{0, 0}
		}
	}
}

// main function
func main() {
	pixelgl.Run(run)
}
