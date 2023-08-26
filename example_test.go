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

func ExampleZerologWriter() {
	fmt.Fprint(os.Stderr, "\n\n------ZEROLOG------\n\n")
	log.Logger = log.Output(liitu.ZerologWriter{Out: os.Stderr})
	log.Logger.Level(zerolog.DebugLevel)

	type testStruct struct {
		Name string
		Age  int
		Foo  bool
	}

	testStructValue := &testStruct{
		Name: "John",
		Age:  32,
		Foo:  true,
	}

	log.Info().Str("foo", "bar").Msg("hello")
	log.Trace().Caller().
		Int("number", 14).
		Bytes("binary", []byte("Test")).
		Bool("boolean", false).
		Str("foo", "bar").
		RawJSON("value", []byte(`{"some": "json"}`)).Msg("hello")
	log.Debug().Msg("hi")
	log.Warn().Msg("WARNING")
	log.Info().Interface("testStruct", testStructValue).Msg("hello")

	testSlice := []struct {
		Name   string
		Age    int
		Foo    bool
		Nested *testStruct
	}{
		{
			Name:   "John",
			Age:    32,
			Foo:    true,
			Nested: testStructValue,
		},
		{
			Name: "Jane",
			Age:  21,
			Foo:  false,
		},
	}

	log.Error().Interface("testSlice", testSlice).Msg("This is a slice")

	// Output:
}

func ExampleNewSlogHandler() {
	fmt.Fprint(os.Stderr, "\n\n------SLOG------\n\n")
	programLevel := new(slog.LevelVar)
	logger := slog.New(
		liitu.NewSlogHandler(
			os.Stderr,
			&liitu.SlogOptions{Level: programLevel},
		),
	)
	programLevel.Set(liitu.LevelTrace)

	type testStruct struct {
		Name      string
		Age       int
		BoolValue bool
	}

	testStructValue := &testStruct{
		Name:      "John",
		Age:       32,
		BoolValue: true,
	}

	logger.Info("hello", slog.String("foo", "bar"))
	logger.Log(context.Background(), -2, "custom level")
	logger.Log(context.Background(), liitu.LevelTrace, "hello",
		slog.Int("number", 14),
		slog.Any("binary", []byte("Test")),
		slog.Bool("boolean", false),
		slog.String("foo", "bar"),
		slog.Any("value", json.RawMessage(`{"some": "json"}`)),
	)
	logger.Debug("hi")
	logger.Warn("WARNING")
	logger.Info("hello", slog.Any("testStruct", testStructValue))
	g := logger.WithGroup("log-group")
	g.Info("hello from a group", slog.Any("testStruct", testStructValue))
	g.Error("foo", slog.Group("My group",
		slog.String("Foo", "bar"),
		slog.String("hello", "world"),
		slog.String("baz", "quz"),
		slog.Group("nested group",
			slog.String("Foo", "bar"),
			slog.String("hello", "world"),
			slog.String("baz", "quz"),
			slog.Any("test", testStructValue),
		),
	))
	g.WithGroup("test").Warn("nested group")

	testSlice := []struct {
		Name   string
		Age    int
		Foo    bool
		Nested *testStruct
	}{
		{
			Name:   "John",
			Age:    32,
			Foo:    true,
			Nested: testStructValue,
		},
		{
			Name: "Jane",
			Age:  21,
			Foo:  false,
		},
	}

	logger.Error("This is a slice", slog.Any("testSlice", testSlice))

	// Output:
}
