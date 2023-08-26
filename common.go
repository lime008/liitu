package liitu

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/goccy/go-json"
)

var (
	IndentSize = 2
	IndentChar = " "
	Indent     = strings.Repeat(IndentChar, IndentSize)
)

var (
	colorDefault = colorFormat(color.Reset)

	colorTime   = colorFormat(color.FgHiBlack)
	colorField  = colorFormat(color.FgCyan)
	colorString = colorFormat(color.FgGreen)
	colorNumber = colorFormat(color.FgMagenta)
	colorBool   = colorFormat(color.FgYellow)
	colorBinary = colorFormat(color.FgBlue)
	colorNull   = colorFormat(color.FgHiYellow)

	colorDebug = colorFormat(color.FgMagenta, color.Bold)
	colorInfo  = colorFormat(color.FgGreen, color.Bold)
	colorWarn  = colorFormat(color.FgYellow, color.Bold)
	colorTrace = colorFormat(color.FgBlue, color.Bold)
	colorError = colorFormat(color.FgRed, color.Bold)
)

var jsonScheme = &json.ColorScheme{
	Int:       jsonFormat(colorNumber),
	Uint:      jsonFormat(colorNumber),
	Float:     jsonFormat(colorNumber),
	Bool:      jsonFormat(colorBool),
	String:    jsonFormat(colorString),
	Binary:    jsonFormat(colorBinary),
	ObjectKey: jsonFormat(colorField),
	Null:      jsonFormat(colorNull),
}

func jsonFormat(f string) json.ColorFormat {
	return json.ColorFormat{
		Header: f,
		Footer: colorDefault,
	}
}

func colorFormat(params ...color.Attribute) string {
	format := make([]string, len(params))
	for i, v := range params {
		format[i] = strconv.Itoa(int(v))
	}

	return fmt.Sprintf("%s[%sm", "\x1b", strings.Join(format, ";"))
}

var timeFormat = "2006-01-02 15:04:05" // time.Kitchen

// Level is a extended log level definition compatible with slog.Level.
// It adds TRACE, FATAL and PANIC levels to the level definitions.
type Level int

const (
	LevelTrace   = -8
	LevelDebug   = -4
	LevelInfo    = 0
	LevelWarning = 4
	LevelError   = 8
	LevelFatal   = 12
	LevelPanic   = 15
	LevelUnknown = 16
)

func levelDelta(level string, val Level) string {
	if val == 0 {
		return level
	}

	return level + "+" + strconv.FormatInt(int64(val), 10)
}

func levelText(level Level) string {
	switch {
	case level < LevelDebug:
		return levelDelta("TRC", level-LevelTrace)
	case level < LevelInfo:
		return levelDelta("DBG", level-LevelDebug)
	case level < LevelWarning:
		return levelDelta("INF", level-LevelInfo)
	case level < LevelError:
		return levelDelta("WRN", level-LevelWarning)
	case level < LevelFatal:
		return levelDelta("ERR", level-LevelError)
	case level < LevelPanic:
		return levelDelta("FTL", level-LevelFatal)
	case level == LevelPanic:
		return levelDelta("PNC", 0)
	default:
		return levelDelta("???", level)
	}
}

func levelColor(level Level) string {
	switch {
	case level < LevelDebug:
		return colorTrace
	case level < LevelInfo:
		return colorDebug
	case level < LevelWarning:
		return colorInfo
	case level < LevelError:
		return colorWarn
	case level < LevelUnknown:
		return colorError
	default:
		return colorDefault
	}
}

func formatTime(t any) string {
	switch t := t.(type) {
	case string:
		u, _ := time.Parse(time.RFC3339, t)
		return u.Format(timeFormat)
	case json.Number:
		u, _ := t.Int64()
		return time.Unix(u, 0).Format(timeFormat)
	}
	return time.Now().Format(timeFormat)
}

func jsonOptions(noColor bool) func(*json.EncodeOption) {
	return func(opts *json.EncodeOption) {
		json.DisableHTMLEscape()(opts)
		json.DisableNormalizeUTF8()(opts)

		if noColor {
			return
		}

		json.Colorize(jsonScheme)(opts)
	}
}

type printer struct {
	noColor bool
	jsonEnc *json.Encoder
	bytes.Buffer
	sync.Mutex
}

func newPrinter(noColor bool) *printer {
	p := &printer{
		noColor: noColor,
	}

	p.jsonEnc = json.NewEncoder(p)

	return p
}

func (p *printer) Print(color string, message string) {
	if p.noColor {
		_, _ = p.WriteString(message)
		return
	}

	_, _ = p.WriteString(color + message + colorDefault)
}

func (p *printer) Println(color string, message string) {
	p.Print(color, message+"\n")
}

func (p *printer) PrintField(s string, indentLevel int) {
	p.Print(colorField, strings.Repeat(Indent, indentLevel)+s)
	_, _ = p.WriteString(": ")
}

func (p *printer) PrintLevel(level Level) {
	p.Print(levelColor(level), " ["+levelText(level)+"] ")
}

func (p *printer) PrintTime(time any) {
	p.Print(colorTime, formatTime(time))
}

func (p *printer) PrintGroup(groups []string) {
	p.Print(colorField, "("+strings.Join(groups, "/")+") ")
}

func (p *printer) PrintJson(value any, indentLevel int) {
	p.jsonEnc.SetIndent(
		strings.Repeat(Indent, indentLevel),
		Indent,
	)

	if v, ok := value.(json.RawMessage); ok {
		var event map[string]any
		_ = json.UnmarshalWithOption(v, &event, json.DecodeFieldPriorityFirstWin())
		_ = p.jsonEnc.EncodeWithOption(event, jsonOptions(p.noColor))
	}

	_ = p.jsonEnc.EncodeWithOption(value, jsonOptions(p.noColor))
}
