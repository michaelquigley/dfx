package dfx

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
)

// MultiGrid is a flexible component container that separates component management
// from layout strategy. Components are managed as a named collection, and different
// layout strategies can be applied to arrange them.
type MultiGrid struct {
	*Container
	components map[string]Component
	layout     Layout
}

// Layout defines how components are arranged and how user interaction is handled
type Layout interface {
	// Arrange renders the components according to the layout strategy
	Arrange(components map[string]Component, state *State)

	// HandleInput processes user input for layout-specific interactions (resizing, etc)
	HandleInput(state *State)
}

// NewMultiGrid creates a new MultiGrid with no components and no layout
func NewMultiGrid() *MultiGrid {
	return &MultiGrid{
		Container:  &Container{Visible: true},
		components: make(map[string]Component),
	}
}

// AddComponent adds a named component to the collection
func (mg *MultiGrid) AddComponent(id string, component Component) {
	mg.components[id] = component
}

// RemoveComponent removes a component from the collection
func (mg *MultiGrid) RemoveComponent(id string) {
	delete(mg.components, id)
}

// GetComponent retrieves a component by ID
func (mg *MultiGrid) GetComponent(id string) (Component, bool) {
	comp, exists := mg.components[id]
	return comp, exists
}

// SetLayout applies a layout strategy to the component collection
func (mg *MultiGrid) SetLayout(layout Layout) {
	mg.layout = layout
}

// ComponentIDs returns all component IDs in the collection
func (mg *MultiGrid) ComponentIDs() []string {
	ids := make([]string, 0, len(mg.components))
	for id := range mg.components {
		ids = append(ids, id)
	}
	return ids
}

// Draw renders the MultiGrid using the current layout strategy
func (mg *MultiGrid) Draw(state *State) {
	if !mg.Visible {
		return
	}

	// call OnDraw if defined
	if mg.OnDraw != nil {
		mg.OnDraw(state)
	}

	// handle input first (for resize operations, etc)
	if mg.layout != nil {
		mg.layout.HandleInput(state)
	}

	// arrange components using the current layout
	if mg.layout != nil {
		mg.layout.Arrange(mg.components, state)
	}

	// draw children (if any)
	for _, child := range mg.Children {
		child.Draw(state)
	}
}

// FlexLayout provides a resizable grid layout similar to the original MultiSurface
type FlexLayout struct {
	arrangement [][]string // component IDs arranged in rows/columns
	rowHeights  []int      // heights for each row (0 = auto-size)
	colWidths   [][]int    // widths for each column in each row (0 = auto-size)

	// resizing state
	dragging     bool
	dragType     DragType
	dragRowIndex int
	dragColIndex int
	dragRowPrev  int
	dragColPrev  int
	deltaRow     int
	deltaCol     int
}

type DragType int

const (
	DragNone DragType = iota
	DragRow
	DragColumn
)

const (
	multiGridMargin      = 2
	multiGridSpacing     = 4
	multiGridSplitWidth  = 10
	multiGridSplitHeight = 11
)

// NewFlexLayout creates a flexible layout with the given arrangement
func NewFlexLayout(arrangement [][]string) *FlexLayout {
	fl := &FlexLayout{
		arrangement: arrangement,
		rowHeights:  make([]int, len(arrangement)),
		colWidths:   make([][]int, len(arrangement)),
	}

	// initialize column width slices
	for i, row := range arrangement {
		fl.colWidths[i] = make([]int, len(row))
	}

	return fl
}

// HandleInput processes mouse input for resize operations
func (fl *FlexLayout) HandleInput(state *State) {
	// handle resize completion
	if fl.dragging {
		if fl.dragType == DragRow && fl.dragRowIndex >= 0 && fl.dragRowPrev >= 0 {
			fl.rowHeights[fl.dragRowIndex] -= fl.deltaRow
			fl.rowHeights[fl.dragRowPrev] += fl.deltaRow
		} else if fl.dragType == DragColumn && fl.dragRowIndex >= 0 && fl.dragColIndex >= 0 && fl.dragColPrev >= 0 {
			fl.colWidths[fl.dragRowIndex][fl.dragColIndex] -= fl.deltaCol
			fl.colWidths[fl.dragRowIndex][fl.dragColPrev] += fl.deltaCol
		}
		fl.dragging = false
		fl.dragType = DragNone
		fl.deltaRow = 0
		fl.deltaCol = 0
	}
}

// Arrange renders components in a flexible grid with resizable splitters
func (fl *FlexLayout) Arrange(components map[string]Component, state *State) {
	if len(fl.arrangement) == 0 {
		return
	}

	fl.sizeRows(state.Size)

	cursor := imgui.CursorStartPos()
	for i, row := range fl.arrangement {
		imgui.SetCursorPos(cursor)
		rowHeight := fl.rowHeights[i]
		rowSize := imgui.Vec2{X: state.Size.X - multiGridSpacing, Y: float32(rowHeight)}

		// draw row splitter (except for first row)
		if i > 0 {
			rowSize.Y -= multiGridSplitHeight
			imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{X: 0, Y: 0})
			imgui.InvisibleButton(fmt.Sprintf("row_%d_split", i), imgui.Vec2{X: state.Size.X, Y: multiGridSplitWidth})
			imgui.PopStyleVar()

			if imgui.IsItemHovered() {
				imgui.SetMouseCursor(imgui.MouseCursorResizeNS)
			}
			if imgui.IsItemActive() {
				fl.dragging = true
				fl.dragType = DragRow
				fl.deltaRow = int(imgui.CurrentIO().MouseDelta().Y)
				fl.dragRowIndex = i
				fl.dragRowPrev = i - 1
			}
			imgui.SetCursorPos(cursor.Add(imgui.Vec2{X: 0, Y: multiGridSplitWidth}))
		}

		// arrange columns in this row
		fl.sizeColumns(state.Size, i)
		colCursor := imgui.CursorPos()

		for j, componentID := range row {
			imgui.SetCursorPos(colCursor)
			colWidth := fl.colWidths[i][j]
			colSize := imgui.Vec2{X: float32(colWidth - multiGridSpacing), Y: rowSize.Y}

			// draw column splitter (except for first column)
			if j > 0 {
				colSize.X -= multiGridSplitHeight
				imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{X: 0, Y: 0})
				imgui.InvisibleButton(fmt.Sprintf("row_%d_col_%d_split", i, j), imgui.Vec2{X: multiGridSplitWidth, Y: rowSize.Y})
				imgui.PopStyleVar()

				if imgui.IsItemHovered() {
					imgui.SetMouseCursor(imgui.MouseCursorResizeEW)
				}
				if imgui.IsItemActive() {
					fl.dragging = true
					fl.dragType = DragColumn
					fl.deltaCol = int(imgui.CurrentIO().MouseDelta().X)
					fl.dragRowIndex = i
					fl.dragColIndex = j
					fl.dragColPrev = j - 1
				}
				imgui.SetCursorPos(colCursor.Add(imgui.Vec2{X: multiGridSplitWidth, Y: 0}))
			}

			// draw the component
			if component, exists := components[componentID]; exists {
				fl.drawComponent(component, colSize, componentID, state)
			}

			colCursor.X += float32(colWidth)
		}

		cursor.Y += float32(rowHeight)
	}
}

// drawComponent renders a component in a child window
func (fl *FlexLayout) drawComponent(component Component, size imgui.Vec2, id string, state *State) {
	if imgui.BeginChildStrV(fmt.Sprintf("mg_%s", id), size, 0, imgui.WindowFlagsNoScrollbar) {
		childState := &State{
			Size:     size,
			Position: imgui.Vec2{},
			IO:       imgui.CurrentIO(),
			App:      state.App,
			Parent:   state.Parent,
		}
		component.Draw(childState)
	}
	imgui.EndChild()
}

// sizeRows calculates row heights
func (fl *FlexLayout) sizeRows(size imgui.Vec2) {
	if len(fl.rowHeights) == 0 {
		return
	}

	maxY := int(size.Y - multiGridMargin)

	var needsHeight []int
	allocated := 0

	for i, height := range fl.rowHeights {
		if height > 0 {
			allocated += height
		} else {
			needsHeight = append(needsHeight, i)
		}
	}

	if len(needsHeight) > 0 {
		newHeight := maxY / len(fl.rowHeights)
		for _, i := range needsHeight {
			fl.rowHeights[i] = newHeight
			allocated += newHeight
		}
	}

	// distribute overage/underage
	if allocated != maxY {
		diff := maxY - allocated
		sharePerRow := diff / len(fl.rowHeights)
		for i := range fl.rowHeights {
			fl.rowHeights[i] += sharePerRow
		}
	}
}

// sizeColumns calculates column widths for a specific row
func (fl *FlexLayout) sizeColumns(size imgui.Vec2, rowIndex int) {
	if rowIndex >= len(fl.colWidths) || len(fl.colWidths[rowIndex]) == 0 {
		return
	}

	maxX := int(size.X - multiGridMargin)
	colWidths := fl.colWidths[rowIndex]

	var needsWidth []int
	allocated := 0

	for j, width := range colWidths {
		if width > 0 {
			allocated += width
		} else {
			needsWidth = append(needsWidth, j)
		}
	}

	if len(needsWidth) > 0 {
		newWidth := maxX / len(colWidths)
		for _, j := range needsWidth {
			colWidths[j] = newWidth
			allocated += newWidth
		}
	}

	// distribute overage/underage
	if allocated != maxX {
		diff := maxX - allocated
		sharePerCol := diff / len(colWidths)
		for j := range colWidths {
			colWidths[j] += sharePerCol
		}
	}
}

// GridLayout provides fixed-position grid layout with no interactive resizing
type GridLayout struct {
	cells      map[string]GridCell // component ID -> grid position
	gridWidth  int                 // number of columns
	gridHeight int                 // number of rows
	cellSize   imgui.Vec2          // size of each grid cell (0 = auto-size)
}

// GridCell defines a component's position in the grid
type GridCell struct {
	Row, Col int        // grid position (0-based)
	Span     imgui.Vec2 // rowspan, colspan (1,1 = single cell)
}

// NewGridLayout creates a fixed grid layout
func NewGridLayout(gridWidth, gridHeight int) *GridLayout {
	return &GridLayout{
		cells:      make(map[string]GridCell),
		gridWidth:  gridWidth,
		gridHeight: gridHeight,
	}
}

// SetCell positions a component in the grid
func (gl *GridLayout) SetCell(componentID string, row, col int, rowSpan, colSpan int) {
	gl.cells[componentID] = GridCell{
		Row:  row,
		Col:  col,
		Span: imgui.Vec2{X: float32(rowSpan), Y: float32(colSpan)},
	}
}

// HandleInput processes input (no interactive resizing for grid layout)
func (gl *GridLayout) HandleInput(state *State) {
	// grid layout is fixed - no interactive resize
	// could add drag-and-drop reordering here in the future
}

// Arrange renders components at fixed grid positions
func (gl *GridLayout) Arrange(components map[string]Component, state *State) {
	if gl.gridWidth <= 0 || gl.gridHeight <= 0 {
		return
	}

	// calculate cell dimensions
	cellWidth := state.Size.X / float32(gl.gridWidth)
	cellHeight := state.Size.Y / float32(gl.gridHeight)

	// override with fixed cell size if specified
	if gl.cellSize.X > 0 {
		cellWidth = gl.cellSize.X
	}
	if gl.cellSize.Y > 0 {
		cellHeight = gl.cellSize.Y
	}

	// render each component at its grid position
	for componentID, cell := range gl.cells {
		component, exists := components[componentID]
		if !exists {
			continue
		}

		// calculate component position and size
		posX := float32(cell.Col) * cellWidth
		posY := float32(cell.Row) * cellHeight
		sizeX := cell.Span.Y * cellWidth  // span.Y = colSpan
		sizeY := cell.Span.X * cellHeight // span.X = rowSpan

		// ensure component doesn't go outside bounds
		if posX+sizeX > state.Size.X {
			sizeX = state.Size.X - posX
		}
		if posY+sizeY > state.Size.Y {
			sizeY = state.Size.Y - posY
		}

		// draw component at calculated position
		imgui.SetCursorPos(imgui.Vec2{X: posX, Y: posY})
		componentSize := imgui.Vec2{X: sizeX, Y: sizeY}

		if imgui.BeginChildStrV(fmt.Sprintf("grid_%s", componentID), componentSize, 0, imgui.WindowFlagsNoScrollbar) {
			childState := &State{
				Size:     componentSize,
				Position: imgui.Vec2{X: posX, Y: posY},
				IO:       imgui.CurrentIO(),
				App:      state.App,
				Parent:   state.Parent,
			}
			component.Draw(childState)
		}
		imgui.EndChild()
	}
}
