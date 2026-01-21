package dfx

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/df/dl"
)

const (
	LogTimeFormat = "[%8.3f]" // time formatting for log entries
)

var (
	LogTimeColor     = imgui.Vec4{X: 0.5, Y: 0.5, Z: 0.5, W: 1.0}
	LogDebugColor    = imgui.Vec4{X: 0.0, Y: 0.0, Z: 1.0, W: 1.0}
	LogWarningColor  = imgui.Vec4{X: 1.0, Y: 1.0, Z: 0.0, W: 1.0}
	LogErrorColor    = imgui.Vec4{X: 1.0, Y: 0.0, Z: 0.0, W: 1.0}
	LogFunctionColor = imgui.Vec4{X: 0.023, Y: 0.596, Z: 0.603, W: 1.0}
	LogFieldsColor   = imgui.Vec4{X: 0.203, Y: 0.886, Z: 0.886, W: 1.0}
)

// LogMessage represents a single log entry.
type LogMessage struct {
	Time    time.Time
	Level   slog.Level
	Func    string
	Fields  string
	Message string
}

// LogBuffer is a thread-safe circular buffer for log messages.
type LogBuffer struct {
	messages []LogMessage
	maxSize  int
	mu       sync.RWMutex
}

// NewLogBuffer creates a new log buffer with the specified maximum size.
func NewLogBuffer(maxSize int) *LogBuffer {
	return &LogBuffer{
		messages: make([]LogMessage, 0, maxSize),
		maxSize:  maxSize,
	}
}

// Add appends a log message to the buffer. if the buffer is full,
// the oldest message is removed.
func (lb *LogBuffer) Add(msg LogMessage) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.messages = append(lb.messages, msg)
	if len(lb.messages) > lb.maxSize {
		lb.messages = lb.messages[1:]
	}
}

// Messages returns a copy of all messages in the buffer.
func (lb *LogBuffer) Messages() []LogMessage {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	msgs := make([]LogMessage, len(lb.messages))
	copy(msgs, lb.messages)
	return msgs
}

// Range calls f for each log message in the buffer while holding the read lock.
// iteration stops early if f returns false.
// the message pointer is only valid during the callback.
func (lb *LogBuffer) Range(f func(index int, msg *LogMessage) bool) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	for i := range lb.messages {
		if !f(i, &lb.messages[i]) {
			break
		}
	}
}

// Clear removes all messages from the buffer.
func (lb *LogBuffer) Clear() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.messages = lb.messages[:0]
}

// AllText returns all log messages as a single formatted string.
func (lb *LogBuffer) AllText() string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	var out strings.Builder
	for _, msg := range lb.messages {
		fields := ""
		if msg.Fields != "" {
			fields = " " + msg.Fields
		}
		out.WriteString(strings.TrimSuffix(
			fmt.Sprintf("[%v] %8s %v%v %v",
				msg.Time.Format(time.RFC3339Nano),
				msg.Level,
				msg.Func,
				fields,
				msg.Message),
			"\n"))
		out.WriteString("\n")
	}
	return out.String()
}

// Count returns the number of messages in the buffer.
func (lb *LogBuffer) Count() int {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return len(lb.messages)
}

// LogViewer is a component that displays log messages from a LogBuffer.
type LogViewer struct {
	Container
	Buffer      *LogBuffer
	AutoScroll  bool
	LevelFilter slog.Level // minimum level to show
	ShowTime    bool
	ShowFunc    bool
	ShowFields  bool
}

// NewLogViewer creates a new log viewer component.
func NewLogViewer(buffer *LogBuffer) *LogViewer {
	return &LogViewer{
		Container:   Container{Visible: true},
		Buffer:      buffer,
		AutoScroll:  true,
		LevelFilter: slog.LevelInfo,
		ShowTime:    true,
		ShowFunc:    true,
		ShowFields:  true,
	}
}

// Draw renders the log viewer.
func (lv *LogViewer) Draw(state *State) {
	if !lv.Visible {
		return
	}

	// create scrollable child window for log messages
	imgui.PushStyleVarFloat(imgui.StyleVarScrollbarSize, 9)
	imgui.BeginChildStr("##logViewerContent")
	imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{X: 0, Y: 0})
	PushFont(MonospaceFont)

	// get count for clipper (single lock acquisition)
	count := lv.Buffer.Count()

	// use list clipper for efficient rendering
	clipper := imgui.NewListClipper()
	if count > 0 {
		clipper.Begin(int32(count))
		for clipper.Step() {
			start := int(clipper.DisplayStart())
			end := int(clipper.DisplayEnd())

			// iterate only over visible range using Range to avoid copying
			lv.Buffer.Range(func(index int, msg *LogMessage) bool {
				// only process messages in visible range
				if index < start {
					return true // continue to next message
				}
				if index >= end {
					return false // stop iteration (past visible range)
				}

				// skip messages below filter level
				if msg.Level < lv.LevelFilter {
					return true // continue to next message
				}

				lv.renderMessage(msg, state)
				return true // continue to next message
			})
		}
	}

	// auto-scroll to bottom
	if lv.AutoScroll && imgui.ScrollY() >= imgui.ScrollMaxY() {
		imgui.SetScrollHereYV(1.0)
	}

	imgui.PopFont()
	imgui.PopStyleVar()
	imgui.EndChild()
	imgui.PopStyleVar() // pop scrollbar size

	// call base container drawing
	if lv.OnDraw != nil {
		lv.OnDraw(state)
	}
	for _, child := range lv.Children {
		child.Draw(state)
	}
}

// renderMessage renders a single log message with color formatting.
func (lv *LogViewer) renderMessage(msg *LogMessage, state *State) {
	// render time if enabled
	if lv.ShowTime {
		// calculate relative time
		relativeTime := msg.Time.Sub(state.App.startTime).Seconds()
		imgui.TextColored(LogTimeColor, fmt.Sprintf(LogTimeFormat, relativeTime))
		imgui.SameLine()
	}

	// render level with appropriate color
	switch msg.Level {
	case slog.LevelDebug:
		imgui.TextColored(LogDebugColor, "   DEBUG")
	case slog.LevelInfo:
		imgui.TextUnformatted("    INFO")
	case slog.LevelWarn:
		imgui.TextColored(LogWarningColor, " WARNING")
	case slog.LevelError:
		imgui.TextColored(LogErrorColor, "   ERROR")
	}

	// render function if enabled
	if lv.ShowFunc && msg.Func != "" {
		imgui.SameLine()
		imgui.TextColored(LogFunctionColor, " "+msg.Func+" ")
	}

	// render fields if enabled and present
	if lv.ShowFields && msg.Fields != "" {
		imgui.SameLine()
		imgui.TextColored(LogFieldsColor, msg.Fields+" ")
	}

	// render message
	imgui.SameLine()
	imgui.TextUnformatted(msg.Message)
}

// SlogHandlerOptions configures the slog handler integration.
type SlogHandlerOptions struct {
	TrimPrefix string
	MinLevel   slog.Level
	StartTime  time.Time
}

// SlogHandler is a slog.Handler implementation that writes to a LogBuffer.
// this provides integration with the df/dl logging framework.
type SlogHandler struct {
	buffer     *LogBuffer
	trimPrefix string
	minLevel   slog.Level
	startTime  time.Time
	attrs      []slog.Attr
}

// NewSlogHandler creates a new slog handler that writes to a log buffer.
func NewSlogHandler(buffer *LogBuffer, opts *SlogHandlerOptions) *SlogHandler {
	if opts == nil {
		opts = &SlogHandlerOptions{
			MinLevel:  slog.LevelInfo,
			StartTime: time.Now(),
		}
	}
	if opts.StartTime.IsZero() {
		opts.StartTime = time.Now()
	}
	return &SlogHandler{
		buffer:     buffer,
		trimPrefix: opts.TrimPrefix,
		minLevel:   opts.MinLevel,
		startTime:  opts.StartTime,
	}
}

// Enabled implements slog.Handler.
func (h *SlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

// Handle implements slog.Handler.
func (h *SlogHandler) Handle(_ context.Context, rec slog.Record) error {
	msg := LogMessage{
		Time:    rec.Time,
		Level:   rec.Level,
		Message: rec.Message,
	}

	// extract function name from caller
	fs := runtime.CallersFrames([]uintptr{rec.PC})
	f, _ := fs.Next()
	fStr := f.Function
	if h.trimPrefix != "" {
		fStr = strings.TrimPrefix(fStr, h.trimPrefix)
	}
	msg.Func = fStr

	// extract attributes
	rec.AddAttrs(h.attrs...)
	if rec.NumAttrs() > 0 {
		fieldsMap := make(map[string]interface{}, rec.NumAttrs())
		rec.Attrs(func(a slog.Attr) bool {
			// skip channel key (df/dl internal)
			if a.Key != dl.ChannelKey {
				fieldsMap[a.Key] = a.Value.Any()
			}
			return true
		})
		if len(fieldsMap) > 0 {
			fields, err := json.Marshal(fieldsMap)
			if err != nil {
				return err
			}
			msg.Fields = string(fields)
		}
	}

	h.buffer.Add(msg)
	h.attrs = nil

	return nil
}

// WithAttrs implements slog.Handler.
func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.attrs = attrs
	return h
}

// WithGroup implements slog.Handler.
func (h *SlogHandler) WithGroup(_ string) slog.Handler {
	return h
}
