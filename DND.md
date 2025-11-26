# DFX Drag-and-Drop Implementation

This document describes the XDND (X Drag and Drop) implementation for DFX, enabling drag-and-drop from DFX applications to other X11 applications like REAPER.

## Overview

DFX now supports dragging data (e.g., MIDI files) from a DFX application window to external applications on Linux/X11. The implementation uses the XDND protocol, which is the standard drag-and-drop protocol for X11.

## Architecture

### Key Files

- **`xdnd/source.go`** - Core XDND protocol implementation (drag source)
- **`xdnd/atoms.go`** - X11 atom management (protocol message types)
- **`xdnd/x11.go`** - X11 utilities (window finding by title)
- **`draghandle.go`** - DragHandle component for initiating drags from UI

### How It Works

1. **Proxy Window**: We create a small invisible X11 window on our own X connection. This is necessary because GLFW owns the main window, and we can't receive X events on windows we don't own. The proxy window receives XDND responses (XdndStatus, XdndFinished).

2. **Window Discovery**: Since cimgui-go/GLFW doesn't expose the native X11 window handle, we find it by searching the window tree for a window matching our title.

3. **XDND Protocol Flow**:
   ```
   Source                          Target
     |                               |
     |-- XdndEnter (types) -------->|
     |-- XdndPosition (x,y) ------->|
     |<---- XdndStatus (accept) ----|
     |   (repeat position/status)   |
     |-- XdndDrop ----------------->|
     |<---- SelectionRequest -------|  (target requests data)
     |-- SelectionNotify (data) --->|
     |<---- XdndFinished -----------|
   ```

4. **Mouse Tracking**: Once a drag starts, we use `QueryPointer` to track the mouse position directly from X11, allowing tracking even outside our window.

## Two Transfer Modes

### File-Based Mode (`UseFile = true`)
- Creates a temp file at `/tmp/dfx-drag-<filename>`
- Advertises `text/uri-list` MIME type with `file://` URI
- **Required for REAPER and most applications**
- Temp file cleaned up on next drag or app exit

### In-Memory Mode (`UseFile = false`)
- Transfers raw data directly via the specified MIME types
- No temp file created
- Only works with apps that specifically support the MIME type (rare)
- REAPER shows "could not import" error with this mode

## Usage

```go
// Create a drag handle
dragHandle := dfx.NewDragHandle(func() []byte {
    return midiData  // Return the data to drag
})

// Configure the drag handle
dragHandle.Label = "Drag MIDI"
dragHandle.MimeTypes = []string{"audio/midi"}  // MIME types to offer
dragHandle.UseFile = true                       // File-based (for compatibility)
dragHandle.Filename = "song.mid"               // Temp filename
dragHandle.Size = imgui.Vec2{X: 200, Y: 50}

// Draw it in your component
dragHandle.Draw(state)
```

## Debug Logging

Enable debug output:
```go
xdnd.Debug = true
```

This logs all XDND protocol messages, useful for troubleshooting.

## Known Limitations

1. **X11 Only**: No Wayland support (would require a different protocol)
2. **Drag Out Only**: No support for receiving drops (drag-in)
3. **File-Based Required**: Most apps (including REAPER) only accept `text/uri-list`, not raw MIME types
4. **Window Title Matching**: The X11 window is found by title, so the title must be unique

## Key Discoveries During Implementation

1. **Separate X Connection**: We couldn't use GLFW's X connection, so we open our own. This required creating a proxy window to receive events.

2. **No Pointer Grab**: Initially tried grabbing the pointer for tracking, but this conflicted with GLFW and locked up X. Switched to polling with `QueryPointer`.

3. **TARGETS Atom**: Drop targets first request `TARGETS` to see available formats before requesting actual data.

4. **URL Encoding**: File URIs must be properly URL-encoded (spaces as `%20`).

5. **Event Processing After Drop**: Must continue processing X events after `XdndDrop` to handle `SelectionRequest` and `XdndFinished`.

## Example Application

See `examples/dfx_example_dragmidi/main.go` for a complete example that loads a MIDI file and provides a drag handle to export it.

```bash
go build ./examples/dfx_example_dragmidi/
./dfx_example_dragmidi path/to/file.mid
```

## Future Work

- Wayland support (via `wl_data_device`)
- Windows and macOS support
- Drag-in (drop target) support
- Visual drag feedback (drag cursor/icon)
- Multiple file drag support
