package liitu

import (
	"io"
	"strconv"

	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
)

type ZerologWriter struct {
	Out     io.Writer
	NoColor bool
	p       *printer
}

func (w ZerologWriter) Write(p []byte) (n int, err error) {
	if w.p == nil {
		w.p = newPrinter(w.NoColor)
	}
	w.p.Lock()
	defer w.p.Unlock()

	var event map[string]any
	err = json.UnmarshalWithOption(p, &event, json.DecodeFieldPriorityFirstWin())
	if err != nil {
		return
	}

	var level Level
	if l, ok := event[zerolog.LevelFieldName].(string); ok {
		level = setLevel(l)
		delete(event, zerolog.LevelFieldName)
	}
	w.p.PrintTime(event[zerolog.TimestampFieldName])
	delete(event, zerolog.TimestampFieldName)
	w.p.PrintLevel(level)
	if message, ok := event[zerolog.MessageFieldName].(string); ok {
		w.p.WriteString(message + "\n")
		delete(event, zerolog.MessageFieldName)
	}

	for field := range event {
		w.p.PrintField(field, 1)
		switch value := event[field].(type) {
		case string:
			w.p.Println(colorString, strconv.Quote(value))
		case json.Number:
			w.p.Println(colorNumber, value.String())
		default:
			w.p.PrintJson(value, 1)
		}
	}

	c, err := io.Copy(w.Out, w.p)
	return int(c), err
}

func setLevel(level string) Level {
	switch level {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarning
	case "trace":
		return LevelTrace
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	case "panic":
		return LevelPanic
	default:
		return LevelUnknown
	}
}
