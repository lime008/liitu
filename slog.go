package liitu

import (
	"context"
	"io"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
)

var (
	DefaultLevel      = slog.LevelInfo
	DefaultTimeFormat = "2006-01-02 15:04:05"
)

// Options for a slog.Handler that writes tinted logs. A zero Options consists
// entirely of default values.
//
// Options can be used as a drop-in replacement for [slog.HandlerOptions].
type SlogOptions struct {
	// Minimum level to log (Default: slog.LevelInfo)
	Level slog.Leveler
	// Disable color (Default: false)
	NoColor bool
	// Enable source code location (Default: false)
	AddSource bool
	// Enable caller function (Default: false)
	AddCaller bool
	// Time format (Default: time.StampMilli)
	TimeFormat string
	// ReplaceAttr is called to rewrite each non-group attribute before it is logged.
	// See https://pkg.go.dev/log/slog#HandlerOptions for details.
	ReplaceAttr func(groups []string, attr slog.Attr) slog.Attr
}

// NewHandler creates a [slog.Handler] that writes tinted logs to Writer w,
// using the default options. If opts is nil, the default options are used.
func NewSlogHandler(w io.Writer, opts *SlogOptions) slog.Handler {
	h := &handler{
		attrsPrefix: "",
		groups:      []string{},
		w:           w,
		p:           newPrinter(false),
		addSource:   false,
		addFunction: false,
		level:       DefaultLevel,
		replaceAttr: func([]string, slog.Attr) slog.Attr { panic("not implemented") },
		timeFormat:  DefaultTimeFormat,
		noColor:     false,
		jsonEncoder: &json.Encoder{},
	}
	if opts == nil {
		return h
	}

	if opts.NoColor {
		h.p.noColor = opts.NoColor
	}

	h.addSource = opts.AddSource
	if h.addSource {
		h.addFunction = opts.AddCaller
	}

	if opts.Level != nil {
		h.level = opts.Level
	}

	h.replaceAttr = opts.ReplaceAttr
	if opts.TimeFormat != "" {
		h.timeFormat = opts.TimeFormat
	}

	h.noColor = opts.NoColor
	h.p.noColor = opts.NoColor

	return h
}

// handler implements a [slog.Handler].
type handler struct {
	w io.Writer
	p *printer

	addSource   bool
	addFunction bool
	noColor     bool

	attrsPrefix string
	groups      []string

	level       slog.Leveler
	replaceAttr func([]string, slog.Attr) slog.Attr
	timeFormat  string
	jsonEncoder *json.Encoder
}

func (h *handler) clone() *handler {
	return &handler{
		attrsPrefix: h.attrsPrefix,
		groups:      h.groups,
		w:           h.w,
		p:           h.p,
		addSource:   h.addSource,
		addFunction: h.addFunction,
		level:       h.level,
		replaceAttr: h.replaceAttr,
		timeFormat:  h.timeFormat,
		noColor:     h.noColor,
		jsonEncoder: h.jsonEncoder,
	}
}

func (h *handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *handler) Handle(_ context.Context, record slog.Record) error {
	rep := h.replaceAttr

	h.p.Lock()
	defer h.p.Unlock()

	h.printTime(record.Time)

	// write level
	if rep == nil {
		h.p.PrintLevel(Level(record.Level))
	} else if a := rep(h.groups, slog.Int(slog.LevelKey, int(record.Level))); a.Key != "" {
		if a.Value.Kind() == slog.KindInt64 {
			h.p.PrintLevel(Level(a.Value.Int64()))
		}
	}

	// write handler groups
	if len(h.groups) > 0 {
		h.p.PrintGroup(h.groups)
	}

	// write message
	if rep == nil {
		_, _ = h.p.WriteString(record.Message + "\n")
	} else if a := rep(h.groups, slog.String(slog.MessageKey, record.Message)); a.Key != "" {
		_, _ = h.p.WriteString(a.Value.String() + "\n")
	}

	h.printSource(record)

	// write handler attributes
	if len(h.attrsPrefix) > 0 {
		h.p.PrintField(h.attrsPrefix, 1)
	}

	// write attributes
	record.Attrs(func(attr slog.Attr) bool {
		if rep != nil {
			attr = rep(h.groups, attr)
		}
		h.printAttr(attr, 1)
		return true
	})

	_, err := io.Copy(h.w, h.p)

	return err
}

func (h *handler) printTime(t time.Time) {
	if t.IsZero() {
		return
	}

	val := t.Round(0) // strip monotonic to match Attr behavior

	if h.replaceAttr == nil {
		h.p.PrintTime(t)
	} else if a := h.replaceAttr(h.groups, slog.Time(slog.TimeKey, val)); a.Key != "" {
		if a.Value.Kind() == slog.KindTime {
			h.p.PrintTime(a.Value.Time())
		}
	}
}

func (h *handler) printSource(r slog.Record) {
	if !h.addSource {
		return
	}

	fs := runtime.CallersFrames([]uintptr{r.PC})

	frame, _ := fs.Next()
	if frame.File == "" {
		return
	}

	src := &slog.Source{
		Function: frame.Function,
		File:     frame.File,
		Line:     frame.Line,
	}

	h.printSrcValue(src)
}

func (h *handler) printSrcValue(attr *slog.Source) {
	if h.replaceAttr == nil {
		h.p.Println(
			colorString,
			Indent+attr.File+":"+strconv.Itoa(
				attr.Line,
			),
		)

		if h.addFunction {
			h.p.Println(
				colorString,
				strings.Repeat(Indent, 2)+"func "+attr.Function,
			)
		}
	} else if a := h.replaceAttr(h.groups, slog.Any(slog.SourceKey, attr)); a.Key != "" {
		h.printValue(a.Value, 1)
	}
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	h2 := h.clone()

	// write attributes to buffer
	for _, attr := range attrs {
		if h.replaceAttr != nil {
			attr = h.replaceAttr(h.groups, attr)
		}

		h.printAttr(attr, 1)
	}

	h2.attrsPrefix = h.attrsPrefix

	return h2
}

func (h *handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	h2 := h.clone()
	h2.groups = append(h2.groups, name)

	return h2
}

func (h *handler) printGroup(attrs []slog.Attr, indentLevel int) {
	_ = h.p.WriteByte('\n')

	for _, a := range attrs {
		h.printAttr(a, indentLevel)
	}
}

func (h *handler) printAttr(attr slog.Attr, indentLevel int) {
	if attr.Equal(slog.Attr{}) {
		return
	}

	attr.Value = attr.Value.Resolve()

	h.p.PrintField(attr.Key, indentLevel)
	h.printValue(attr.Value, indentLevel)
}

func (h *handler) printValue(value slog.Value, indentLevel int) {
	switch value.Kind() {
	case slog.KindGroup:
		h.printGroup(value.Group(), indentLevel+1)
	case slog.KindInt64:
		h.p.Println(colorNumber, strconv.FormatInt(value.Int64(), 10))
	case slog.KindFloat64:
		h.p.Println(colorNumber, strconv.FormatFloat(value.Float64(), 'g', 8, 64))
	case slog.KindUint64:
		h.p.Println(colorNumber, strconv.FormatUint(value.Uint64(), 10))
	case slog.KindString:
		h.p.Println(colorString, strconv.Quote(value.String()))
	case slog.KindBool:
		h.p.Println(colorBool, strconv.FormatBool(value.Bool()))
	default:
		h.p.PrintJson(value.Any(), indentLevel)
	}
}
