package dfx

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx/xdnd"
)

const (
	// DragThreshold is the minimum distance in pixels to initiate a drag
	DragThreshold = 5.0
)

// DragHandle is a component that allows initiating drag-and-drop operations
type DragHandle struct {
	// OnDragStart is called when a drag is initiated, should return the data to drag
	OnDragStart func() []byte

	// MimeTypes specifies the MIME types to offer (defaults to ["audio/midi"])
	MimeTypes []string

	// Filename is the name to use for the temp file (defaults to "data.mid")
	// Only used when UseFile is true
	Filename string

	// UseFile enables file-based transfer (writes to temp file, advertises text/uri-list)
	// When false, uses in-memory transfer with raw MIME types
	UseFile bool

	// Label is the text shown on the drag handle (defaults to "Drag")
	Label string

	// Size specifies the size of the drag handle (defaults to auto-size)
	Size imgui.Vec2

	// xdnd source reference (set by App)
	xdndSource *xdnd.Source

	// internal state
	actions    *ActionRegistry
	mouseDown  bool
	mouseStart imgui.Vec2
	dragging   bool
}

// NewDragHandle creates a new DragHandle component
func NewDragHandle(onDragStart func() []byte) *DragHandle {
	return &DragHandle{
		OnDragStart: onDragStart,
		MimeTypes:   []string{"audio/midi"},
		Label:       "Drag",
	}
}

// Draw implements Component
func (d *DragHandle) Draw(state *State) {
	// Get xdnd source from app if not set
	if d.xdndSource == nil && state.App != nil {
		d.xdndSource = state.App.XDNDSource()
	}

	// Use a button-like appearance for the drag handle
	label := d.Label
	if label == "" {
		label = "Drag"
	}

	// Draw the button
	var clicked bool
	if d.Size.X > 0 && d.Size.Y > 0 {
		clicked = imgui.ButtonV(label+"##draghandle", d.Size)
	} else {
		clicked = imgui.Button(label + "##draghandle")
	}
	_ = clicked // We use the button for visual, but handle drag ourselves

	// Check if button is hovered for drag detection
	if imgui.IsItemHovered() {
		// Change cursor to indicate draggable
		imgui.SetMouseCursor(imgui.MouseCursorHand)

		// Check for mouse down
		if imgui.IsMouseClickedBool(imgui.MouseButtonLeft) {
			d.mouseDown = true
			d.mouseStart = imgui.MousePos()
			d.dragging = false
		}
	}

	// Track mouse while button is held (before drag starts)
	if d.mouseDown && !d.dragging {
		if imgui.IsMouseReleased(imgui.MouseButtonLeft) {
			// Mouse released before drag threshold - just cancel
			d.mouseDown = false
		} else {
			// Check if we've moved past the threshold
			currentPos := imgui.MousePos()
			dx := currentPos.X - d.mouseStart.X
			dy := currentPos.Y - d.mouseStart.Y
			distance := dx*dx + dy*dy

			if distance > DragThreshold*DragThreshold {
				// Start the drag
				d.startDrag()
			}
		}
	}

	// Once dragging, use X11 for all tracking (works outside window)
	if d.dragging && d.xdndSource != nil {
		// Query pointer state directly from X11
		x, y, buttonPressed, ok := d.xdndSource.QueryPointerState()
		if ok {
			d.xdndSource.UpdateMousePosition(x, y)

			// Check if mouse button was released
			if !buttonPressed {
				d.xdndSource.FinishDrag()
				// Don't set d.dragging = false here - keep processing events
				// until XdndFinished is received
				d.mouseDown = false
			}
		}

		// Process X11 events (for SelectionRequest, XdndStatus, XdndFinished)
		d.xdndSource.ProcessEvents()

		// Check if drag is complete (e.g., XdndFinished received)
		if !d.xdndSource.IsDragging() {
			d.dragging = false
			d.mouseDown = false
		}
	}
}

// Actions implements Component
func (d *DragHandle) Actions() *ActionRegistry {
	if d.actions == nil {
		d.actions = NewActionRegistry()
	}
	return d.actions
}

// SetXDNDSource sets the XDND source (called by App during setup)
func (d *DragHandle) SetXDNDSource(source *xdnd.Source) {
	d.xdndSource = source
}

func (d *DragHandle) startDrag() {
	if d.xdndSource == nil || d.OnDragStart == nil {
		return
	}

	// Get the data to drag
	data := d.OnDragStart()
	if data == nil {
		return
	}

	// Get MIME types
	mimeTypes := d.MimeTypes
	if len(mimeTypes) == 0 {
		mimeTypes = []string{"audio/midi"}
	}

	var err error
	if d.UseFile {
		// File-based transfer (writes temp file, advertises text/uri-list)
		filename := d.Filename
		if filename == "" {
			filename = "data.mid"
		}
		err = d.xdndSource.StartDragWithFilename(mimeTypes, data, filename)
	} else {
		// In-memory transfer (raw MIME types only)
		err = d.xdndSource.StartDrag(mimeTypes, data)
	}

	if err != nil {
		// Log error but don't crash
		return
	}

	d.dragging = true
}
