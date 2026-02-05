package liitu_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/goccy/go-json"

	"github.com/lime008/liitu"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type testStruct struct {
	Name   string
	Age    int
	Member bool
	Nested *testStruct
}

var myStruct = &testStruct{
	Name:   "John",
	Age:    32,
	Member: true,
	Nested: &testStruct{
		Name:   "Johnny",
		Age:    3,
		Member: false,
	},
}

var testSlice = []struct {
	Name string
	Age  int
}{
	{
		Name: "John",
		Age:  32,
	},
	{
		Name: "Jane",
		Age:  21,
	},
}

func ExampleZerologWriter() {
	log.Logger = log.Output(liitu.ZerologWriter{Out: os.Stderr})
	log.Logger.Level(zerolog.TraceLevel)

	log.Trace().Msg("hello")
	log.Debug().Msg("world")
	log.Info().Msg("foo")
	log.Warn().Msg("bar")
	log.Error().Msg("baz")
	log.WithLevel(zerolog.FatalLevel).Msg("quz")
	log.WithLevel(zerolog.PanicLevel).Msg("do not panic")
	log.Debug().Caller().
		Int("number", 14).
		Bytes("binary", []byte("Test")).
		Bool("boolean", false).
		Str("foo", "bar").
		RawJSON("value", []byte(`{"some": "json"}`)).Msg("printing log fields")
	log.Debug().Msg("hi")
	log.Warn().Msg("WARNING")
	log.Info().Interface("testStruct", myStruct).Msg("hello")

	// Output:
}

func ExampleNewSlogHandler() {
	fmt.Fprintf(os.Stderr, "\n\n\n")
	ctx := context.Background()

	programLevel := new(slog.LevelVar)

	logger := slog.New(
		liitu.NewSlogHandler(
			os.Stderr,
			&liitu.SlogOptions{Level: programLevel},
		),
	)

	programLevel.Set(liitu.LevelTrace)

	logger.Log(ctx, liitu.LevelTrace, "hello")
	logger.Debug("world")
	logger.Info("foo")
	logger.Warn("bar")
	logger.Error("baz")
	logger.Log(ctx, liitu.LevelFatal, "quz")
	logger.Log(ctx, liitu.LevelPanic, "do not panic")
	logger.Log(ctx, slog.LevelWarn+2, "custom levels")
	logger.Info("printing log message fields",
		slog.String("string", "bar"),
		slog.Int("integer", 14),
		slog.Float64("float", 18.35),
		slog.Bool("boolean", false),
		slog.Any("binary", []byte("Test")),
		slog.Any("raw json message value", json.RawMessage(`{"some": "json"}`)),
		slog.Any("structs", myStruct),
		slog.Any("slices", testSlice),
	)
	g := logger.WithGroup("group")
	g.Info("hello from a group")
	g.WithGroup("child").Warn("with nested log group")
	gc := g.WithGroup("2nd child")
	gc.Info("with two nested groups")

	f := logger.With("additional", "params")
	logger.Info("hello without additional params")
	f.Info("hello with additional params")
	f.Info("override attributes", slog.String("additional", "test"))

	f2 := f.With("nested", "attrs").With("Hello", "world")
	f2.Info("Nested attributes")

	logger.Error("field groups", slog.Group("group",
		slog.String("Foo", "bar"),
		slog.Group("child-group",
			slog.String("hello", "world"),
		),
	))

	loggerWithSource := slog.New(
		liitu.NewSlogHandler(
			os.Stderr,
			&liitu.SlogOptions{Level: programLevel, AddSource: true},
		),
	)
	loggerWithSource.Info(
		"you can automatically include source file in the logs by setting SlogOptions.AddSource=true",
		slog.String("foo", "bar"),
		slog.String("baz", "quz"),
	)

	loggerWithCaller := slog.New(
		liitu.NewSlogHandler(
			os.Stderr,
			&liitu.SlogOptions{Level: programLevel, AddSource: true, AddCaller: true},
		),
	)
	loggerWithCaller.Info(
		"you can also include the caller function with SlogOptions.AddCaller=true",
		slog.String("foo", "bar"),
		slog.String("baz", "quz"),
	)
	fmt.Fprintf(os.Stderr, "\n\n\n")

	// Output:
}
