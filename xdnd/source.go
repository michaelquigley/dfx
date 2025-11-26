package xdnd

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// Debug enables debug logging for XDND operations
var Debug = false

const (
	// XDND protocol version we support
	XDNDVersion = 5
)

// DragState represents the current state of a drag operation
type DragState int

const (
	DragStateIdle DragState = iota
	DragStateDragging
	DragStateDropped
)

// Source implements the XDND drag source protocol
type Source struct {
	conn   *xgb.Conn
	screen *xproto.ScreenInfo
	atoms  *Atoms

	// Our proxy window for receiving XDND responses
	// (We create this ourselves so we can receive events on our connection)
	proxyWindow xproto.Window

	// The actual application window (for reference only)
	appWindow xproto.Window

	// Current drag state
	mu        sync.Mutex
	state     DragState
	data      []byte
	mimeTypes []string
	tempFile  string // Path to temp file for file-based transfers
	fileURI   []byte // Cached file:// URI for text/uri-list

	// Current target window
	targetWindow   xproto.Window
	targetAccepted bool

	// Root window for coordinate translation
	rootWindow xproto.Window
}

// NewSource creates a new XDND drag source
// windowTitle is used to find the application's X11 window
func NewSource(windowTitle string) (*Source, error) {
	// Connect to X server
	conn, err := xgb.NewConnDisplay(X11DisplayName())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to X server: %w", err)
	}

	// Get the setup info
	setup := xproto.Setup(conn)
	if len(setup.Roots) == 0 {
		conn.Close()
		return nil, fmt.Errorf("no screens found")
	}
	screen := setup.DefaultScreen(conn)

	// Intern atoms
	atoms, err := InternAtoms(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to intern atoms: %w", err)
	}

	// Find the application window by title (for reference)
	appWindow, err := FindWindowByTitle(conn, windowTitle)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to find window '%s': %w", windowTitle, err)
	}

	// Create a small proxy window that we own on this connection
	// This window will be used as the source for XDND messages so we can receive responses
	proxyWindow, err := xproto.NewWindowId(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to allocate window ID: %w", err)
	}

	// Create a 1x1 input-only window (invisible)
	err = xproto.CreateWindowChecked(conn,
		0,                  // depth (0 = copy from parent for InputOnly)
		proxyWindow,        // window id
		screen.Root,        // parent
		0, 0,               // x, y
		1, 1,               // width, height
		0,                  // border width
		xproto.WindowClassInputOnly,
		0, // visual (0 = copy from parent)
		xproto.CwEventMask,
		[]uint32{xproto.EventMaskPropertyChange},
	).Check()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create proxy window: %w", err)
	}

	if Debug {
		log.Printf("[XDND] Created proxy window %d for receiving events", proxyWindow)
	}

	return &Source{
		conn:        conn,
		screen:      screen,
		atoms:       atoms,
		proxyWindow: proxyWindow,
		appWindow:   appWindow,
		rootWindow:  screen.Root,
		state:       DragStateIdle,
	}, nil
}

// Close releases resources
func (s *Source) Close() {
	s.mu.Lock()
	// Clean up temp file
	if s.tempFile != "" {
		os.Remove(s.tempFile)
		s.tempFile = ""
	}
	s.mu.Unlock()

	// Destroy our proxy window
	if s.proxyWindow != 0 && s.conn != nil {
		xproto.DestroyWindow(s.conn, s.proxyWindow)
	}

	if s.conn != nil {
		s.conn.Close()
	}
}

// IsDragging returns true if a drag operation is active
// This includes both the dragging phase and waiting for drop completion
func (s *Source) IsDragging() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state == DragStateDragging || s.state == DragStateDropped
}

// StartDrag initiates a drag operation with the given MIME types and data (in-memory, no temp file)
// Note: This does NOT grab the pointer - mouse tracking is done via ImGui
func (s *Source) StartDrag(mimeTypes []string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != DragStateIdle {
		return fmt.Errorf("drag already in progress")
	}

	// Clean up any previous temp file
	if s.tempFile != "" {
		os.Remove(s.tempFile)
		s.tempFile = ""
	}
	s.fileURI = nil

	s.data = data
	s.mimeTypes = mimeTypes
	s.targetWindow = 0
	s.targetAccepted = false

	if Debug {
		log.Printf("[XDND] Starting in-memory drag with types: %v", mimeTypes)
	}

	// Take ownership of XdndSelection
	xproto.SetSelectionOwner(s.conn, s.proxyWindow, s.atoms.XdndSelection, xproto.TimeCurrentTime)

	s.state = DragStateDragging
	return nil
}

// StartDragWithFilename initiates a drag with a specified filename for the temp file
func (s *Source) StartDragWithFilename(mimeTypes []string, data []byte, filename string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != DragStateIdle {
		return fmt.Errorf("drag already in progress")
	}

	// Clean up any previous temp file
	if s.tempFile != "" {
		os.Remove(s.tempFile)
		s.tempFile = ""
	}

	s.data = data
	s.targetWindow = 0
	s.targetAccepted = false

	// Write data to temp file for file-based transfers
	tempDir := os.TempDir()
	tempPath := filepath.Join(tempDir, "dfx-drag-"+filename)
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	s.tempFile = tempPath

	// Create file URI (text/uri-list format requires \r\n termination)
	// URL-encode the path to handle spaces and special characters
	fileURL := &url.URL{Scheme: "file", Path: tempPath}
	s.fileURI = []byte(fileURL.String() + "\r\n")

	if Debug {
		log.Printf("[XDND] Created temp file: %s", tempPath)
		log.Printf("[XDND] File URI: %s", fileURL.String())
	}

	// Always offer text/uri-list (file path) in addition to any other types
	s.mimeTypes = []string{"text/uri-list"}
	for _, mt := range mimeTypes {
		if mt != "text/uri-list" {
			s.mimeTypes = append(s.mimeTypes, mt)
		}
	}

	// Take ownership of XdndSelection
	xproto.SetSelectionOwner(s.conn, s.proxyWindow, s.atoms.XdndSelection, xproto.TimeCurrentTime)

	s.state = DragStateDragging
	return nil
}

// CancelDrag cancels the current drag operation
func (s *Source) CancelDrag() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != DragStateDragging {
		return
	}

	// Send XdndLeave to current target if any
	if s.targetWindow != 0 {
		s.sendLeave()
	}

	s.state = DragStateIdle
	s.data = nil
	s.targetWindow = 0
}

// UpdateMousePosition updates the drag with current mouse position
// If screenX/screenY are -1, queries the X server for current pointer position
func (s *Source) UpdateMousePosition(screenX, screenY int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != DragStateDragging {
		return
	}

	// If position not provided, query from X server
	if screenX < 0 || screenY < 0 {
		cookie := xproto.QueryPointer(s.conn, s.rootWindow)
		reply, err := cookie.Reply()
		if err == nil {
			screenX = int(reply.RootX)
			screenY = int(reply.RootY)
		} else {
			return // Can't get position, skip this update
		}
	}

	s.handleMotion(int16(screenX), int16(screenY))
}

// QueryPointerPosition returns the current screen position of the pointer
func (s *Source) QueryPointerPosition() (x, y int, ok bool) {
	cookie := xproto.QueryPointer(s.conn, s.rootWindow)
	reply, err := cookie.Reply()
	if err != nil {
		return 0, 0, false
	}
	return int(reply.RootX), int(reply.RootY), true
}

// QueryPointerState returns pointer position and whether button1 (left) is pressed
func (s *Source) QueryPointerState() (x, y int, button1Pressed bool, ok bool) {
	cookie := xproto.QueryPointer(s.conn, s.rootWindow)
	reply, err := cookie.Reply()
	if err != nil {
		return 0, 0, false, false
	}
	// Button1 mask is 0x100 (256)
	button1Pressed = (reply.Mask & 0x100) != 0
	return int(reply.RootX), int(reply.RootY), button1Pressed, true
}

// FinishDrag completes the drag operation (called on mouse release)
func (s *Source) FinishDrag() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != DragStateDragging {
		return
	}

	if s.targetWindow != 0 && s.targetAccepted {
		s.sendDrop()
		s.state = DragStateDropped
	} else {
		// No valid target, cancel
		if s.targetWindow != 0 {
			s.sendLeave()
		}
		s.state = DragStateIdle
		s.data = nil
	}
}

// ProcessEvents processes pending X11 events for the drag operation
// This should be called each frame during a drag
func (s *Source) ProcessEvents() {
	for {
		ev, err := s.conn.PollForEvent()
		if ev == nil && err == nil {
			break // No more events
		}
		if err != nil {
			continue
		}

		s.handleEvent(ev)
	}
}

func (s *Source) handleEvent(ev xgb.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch e := ev.(type) {
	case xproto.SelectionRequestEvent:
		s.handleSelectionRequest(e)

	case xproto.ClientMessageEvent:
		s.handleClientMessage(e)
	}
}

func (s *Source) handleMotion(rootX, rootY int16) {
	// Note: called with lock held

	// Find window under cursor
	target := s.findWindowAt(rootX, rootY)

	if target != s.targetWindow {
		// Target changed
		if s.targetWindow != 0 {
			s.sendLeave()
		}
		s.targetWindow = target
		s.targetAccepted = false
		if target != 0 && s.isXdndAware(target) {
			s.sendEnter(target)
		}
	}

	// Send position update
	if s.targetWindow != 0 && s.isXdndAware(s.targetWindow) {
		s.sendPosition(rootX, rootY)
	}
}

func (s *Source) handleSelectionRequest(e xproto.SelectionRequestEvent) {
	// Note: called with lock held

	targetName := s.atoms.GetAtomName(e.Target)
	if Debug {
		log.Printf("[XDND] SelectionRequest: selection=%d target=%d (%s) property=%d requestor=%d",
			e.Selection, e.Target, targetName, e.Property, e.Requestor)
	}

	// Only handle XdndSelection requests
	if e.Selection != s.atoms.XdndSelection {
		if Debug {
			log.Printf("[XDND] Ignoring non-XdndSelection request (selection=%d, want=%d)", e.Selection, s.atoms.XdndSelection)
		}
		return
	}

	// Handle TARGETS request - list available formats
	if e.Target == s.atoms.TARGETS {
		if Debug {
			log.Printf("[XDND] Target requests TARGETS (available formats)")
		}
		// Return list of available types
		var targets []uint32
		targets = append(targets, uint32(s.atoms.TARGETS)) // We support TARGETS
		for _, mime := range s.mimeTypes {
			if atom := s.atoms.GetMimeAtom(mime); atom != 0 {
				targets = append(targets, uint32(atom))
			}
		}

		// Write targets as ATOM array
		data := make([]byte, len(targets)*4)
		for i, t := range targets {
			data[i*4] = byte(t)
			data[i*4+1] = byte(t >> 8)
			data[i*4+2] = byte(t >> 16)
			data[i*4+3] = byte(t >> 24)
		}

		xproto.ChangeProperty(s.conn, xproto.PropModeReplace,
			e.Requestor, e.Property, xproto.AtomAtom, 32,
			uint32(len(targets)), data)

		s.sendSelectionNotify(e, e.Property)
		if Debug {
			log.Printf("[XDND] Sent TARGETS: %v", s.mimeTypes)
		}
		return
	}

	// Check if requested target matches our MIME type
	var propertyAtom xproto.Atom
	if e.Target == s.atoms.AudioMidi && s.containsMime("audio/midi") {
		propertyAtom = s.atoms.AudioMidi
		if Debug {
			log.Printf("[XDND] Target requests audio/midi, we have it")
		}
	} else if e.Target == s.atoms.TextUriList && s.containsMime("text/uri-list") {
		propertyAtom = s.atoms.TextUriList
		if Debug {
			log.Printf("[XDND] Target requests text/uri-list, we have it")
		}
	} else {
		// Send refusal (property = None)
		if Debug {
			log.Printf("[XDND] Target requests unknown type '%s' (target=%d), refusing", targetName, e.Target)
			log.Printf("[XDND] Our atoms: AudioMidi=%d, TextUriList=%d, TARGETS=%d", s.atoms.AudioMidi, s.atoms.TextUriList, s.atoms.TARGETS)
		}
		s.sendSelectionNotify(e, xproto.AtomNone)
		return
	}

	// Determine which data to send based on the requested type
	var dataToSend []byte
	if e.Target == s.atoms.TextUriList {
		dataToSend = s.fileURI // Send file:// URI for text/uri-list
	} else {
		dataToSend = s.data // Send raw data for other types (audio/midi)
	}

	if Debug {
		log.Printf("[XDND] Writing %d bytes to property %d on window %d (type: %s)", len(dataToSend), e.Property, e.Requestor, targetName)
		if e.Target == s.atoms.TextUriList {
			log.Printf("[XDND] Sending file URI: %s", string(s.fileURI))
		}
	}

	// Write data to the requested property
	xproto.ChangeProperty(s.conn, xproto.PropModeReplace,
		e.Requestor, e.Property, propertyAtom, 8,
		uint32(len(dataToSend)), dataToSend)

	// Notify requestor
	s.sendSelectionNotify(e, e.Property)

	if Debug {
		log.Printf("[XDND] Sent SelectionNotify")
	}
}

func (s *Source) sendSelectionNotify(req xproto.SelectionRequestEvent, property xproto.Atom) {
	ev := xproto.SelectionNotifyEvent{
		Time:      req.Time,
		Requestor: req.Requestor,
		Selection: req.Selection,
		Target:    req.Target,
		Property:  property,
	}

	xproto.SendEvent(s.conn, false, req.Requestor, 0, string(ev.Bytes()))
}

func (s *Source) handleClientMessage(e xproto.ClientMessageEvent) {
	// Note: called with lock held
	msgType := e.Type

	if Debug {
		typeName := s.atoms.GetAtomName(msgType)
		log.Printf("[XDND] ClientMessage received: type=%d (%s) window=%d", msgType, typeName, e.Window)
	}

	if msgType == s.atoms.XdndStatus {
		// Target is telling us if it accepts the drop
		// Data layout: [target_window, flags, x, y, width, height]
		// flags & 1 = will accept
		data := e.Data.Data32
		s.targetAccepted = (data[1] & 1) != 0
		if Debug {
			log.Printf("[XDND] XdndStatus: accepted=%v (flags=0x%x)", s.targetAccepted, data[1])
		}

	} else if msgType == s.atoms.XdndFinished {
		// Drop is complete
		if Debug {
			log.Printf("[XDND] XdndFinished received - drop complete")
		}
		s.state = DragStateIdle
		s.data = nil
		s.fileURI = nil
		s.targetWindow = 0
		// Note: We don't delete the temp file immediately - the target may still be reading it
		// It will be cleaned up on the next drag or when Close() is called
	}
}

// findWindowAt finds the topmost XDND-aware window at the given root coordinates
func (s *Source) findWindowAt(rootX, rootY int16) xproto.Window {
	// Translate coordinates to find the window
	cookie := xproto.TranslateCoordinates(s.conn, s.rootWindow, s.rootWindow, rootX, rootY)
	reply, err := cookie.Reply()
	if err != nil {
		return 0
	}

	// Start from the child window at this position
	child := reply.Child
	if child == 0 {
		return 0
	}

	// Walk down the window tree to find the deepest XDND-aware window
	var lastAware xproto.Window
	for child != 0 {
		if s.isXdndAware(child) {
			lastAware = child
		}

		// Get child at this position within current window
		cookie := xproto.TranslateCoordinates(s.conn, s.rootWindow, child, rootX, rootY)
		reply, err := cookie.Reply()
		if err != nil {
			break
		}
		child = reply.Child
	}

	return lastAware
}

// isXdndAware checks if a window has the XdndAware property
func (s *Source) isXdndAware(win xproto.Window) bool {
	cookie := xproto.GetProperty(s.conn, false, win, s.atoms.XdndAware, xproto.AtomAtom, 0, 1)
	reply, err := cookie.Reply()
	if err != nil || reply.ValueLen == 0 {
		return false
	}
	return true
}

func (s *Source) sendEnter(target xproto.Window) {
	// XdndEnter message format:
	// data32[0] = source window
	// data32[1] = version (high 24 bits) | flags (low 8 bits, bit 0 = more than 3 types)
	// data32[2-4] = first 3 MIME type atoms (or 0)

	var data [5]uint32
	data[0] = uint32(s.proxyWindow)
	data[1] = uint32(XDNDVersion << 24)

	// Set MIME type atoms (up to 3 in the message)
	for i, mime := range s.mimeTypes {
		if i >= 3 {
			data[1] |= 1 // Set "more types" flag
			break
		}
		data[2+i] = uint32(s.atoms.GetMimeAtom(mime))
	}

	if Debug {
		log.Printf("[XDND] Sending XdndEnter to window %d with types: %v (atoms: %d, %d, %d)",
			target, s.mimeTypes, data[2], data[3], data[4])
	}

	s.sendClientMessage(target, s.atoms.XdndEnter, data)
}

func (s *Source) sendPosition(rootX, rootY int16) {
	// XdndPosition message format:
	// data32[0] = source window
	// data32[1] = reserved (0)
	// data32[2] = coordinates (x << 16 | y)
	// data32[3] = timestamp
	// data32[4] = action atom

	var data [5]uint32
	data[0] = uint32(s.proxyWindow)
	data[1] = 0
	data[2] = uint32(rootX)<<16 | uint32(uint16(rootY))
	data[3] = uint32(xproto.TimeCurrentTime)
	data[4] = uint32(s.atoms.XdndActionCopy)

	if Debug {
		log.Printf("[XDND] Sending XdndPosition to %d at (%d, %d)", s.targetWindow, rootX, rootY)
	}

	s.sendClientMessage(s.targetWindow, s.atoms.XdndPosition, data)
}

func (s *Source) sendLeave() {
	var data [5]uint32
	data[0] = uint32(s.proxyWindow)
	s.sendClientMessage(s.targetWindow, s.atoms.XdndLeave, data)
}

func (s *Source) sendDrop() {
	var data [5]uint32
	data[0] = uint32(s.proxyWindow)
	data[2] = uint32(xproto.TimeCurrentTime)

	if Debug {
		log.Printf("[XDND] Sending XdndDrop to window %d", s.targetWindow)
	}

	s.sendClientMessage(s.targetWindow, s.atoms.XdndDrop, data)
}

func (s *Source) sendClientMessage(target xproto.Window, msgType xproto.Atom, data [5]uint32) {
	ev := xproto.ClientMessageEvent{
		Format: 32,
		Window: target,
		Type:   msgType,
		Data:   xproto.ClientMessageDataUnionData32New(data[:]),
	}

	xproto.SendEvent(s.conn, false, target, 0, string(ev.Bytes()))

	// Flush to ensure the message is sent immediately
	s.conn.Sync()
}

func (s *Source) containsMime(mime string) bool {
	for _, m := range s.mimeTypes {
		if m == mime {
			return true
		}
	}
	return false
}
