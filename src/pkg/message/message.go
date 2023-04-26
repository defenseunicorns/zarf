// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package message provides a rich set of functions for displaying messages to the user.
package message

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// LogLevel is the level of logging to display.
type LogLevel int

const (
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel LogLevel = iota
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)

// NoProgress tracks whether spinner/progress bars show updates.
var NoProgress bool

var logLevel = InfoLevel

// Write logs to stderr and a buffer for logFile generation.
var logFile *os.File

var useLogFile bool

// DebugWriter represents a writer interface that writes to message.Debug
type DebugWriter struct{}

func (d *DebugWriter) Write(raw []byte) (int, error) {
	Debug(raw)
	return len(raw), nil
}

func init() {
	pterm.ThemeDefault.SuccessMessageStyle = *pterm.NewStyle(pterm.FgLightGreen)
	// Customize default error.
	pterm.Success.Prefix = pterm.Prefix{
		Text:  " ✔",
		Style: pterm.NewStyle(pterm.FgLightGreen),
	}
	pterm.Error.Prefix = pterm.Prefix{
		Text:  "    ERROR:",
		Style: pterm.NewStyle(pterm.BgLightRed, pterm.FgBlack),
	}
	pterm.Info.Prefix = pterm.Prefix{
		Text: " •",
	}

	pterm.DefaultProgressbar.MaxWidth = 85
	pterm.SetDefaultOutput(os.Stderr)
}

// UseLogFile writes output to stderr and a logFile.
func UseLogFile() {
	// Prepend the log filename with a timestamp.
	ts := time.Now().Format("2006-01-02-15-04-05")

	// Try to create a temp log file.
	var err error
	if logFile, err = os.CreateTemp("", fmt.Sprintf("zarf-%s-*.log", ts)); err != nil {
		Error(err, "Error saving a log file")
	} else {
		useLogFile = true
		logStream := io.MultiWriter(os.Stderr, logFile)
		pterm.SetDefaultOutput(logStream)
		message := fmt.Sprintf("Saving log file to %s", logFile.Name())
		Note(message)
	}
}

// SetLogLevel sets the log level.
func SetLogLevel(lvl LogLevel) {
	logLevel = lvl
	if logLevel >= DebugLevel {
		pterm.EnableDebugMessages()
	}
}

// GetLogLevel returns the current log level.
func GetLogLevel() LogLevel {
	return logLevel
}

// Debug prints a debug message.
func Debug(payload ...any) {
	debugPrinter(2, payload...)
}

// Debugf prints a debug message.
func Debugf(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	Debug(message)
}

// Error prints an error message.
func Error(err any, message string) {
	Debug(err)
	Warnf(message)
}

// ErrorWebf prints an error message and returns a web response.
func ErrorWebf(err any, w http.ResponseWriter, format string, a ...any) {
	Debug(err)
	message := fmt.Sprintf(format, a...)
	Warn(message)
	http.Error(w, message, http.StatusInternalServerError)
}

// Errorf prints an error message.
func Errorf(err any, format string, a ...any) {
	Debug(err)
	Warnf(format, a...)
}

// Warn prints a warning message.
func Warn(message string) {
	Warnf("%s", message)
}

// Warnf prints a warning message.
func Warnf(format string, a ...any) {
	message := Paragraph(format, a...)
	pterm.Warning.Println(message)
}

// Fatal prints a fatal error message and exits with a 1.
func Fatal(err any, message string) {
	Debug(err)
	errorPrinter(2).Println(message)
	Debug(string(debug.Stack()))
	os.Exit(1)
}

// Fatalf prints a fatal error message and exits with a 1.
func Fatalf(err any, format string, a ...any) {
	message := Paragraph(format, a...)
	Fatal(err, message)
}

// Info prints an info message.
func Info(message string) {
	Infof("%s", message)
}

// Infof prints an info message.
func Infof(format string, a ...any) {
	if logLevel > 0 {
		message := Paragraph(format, a...)
		pterm.Info.Println(message)
	}
}

// Successf prints a success message.
func Successf(format string, a ...any) {
	message := Paragraph(format, a...)
	pterm.Success.Println(message)
}

// Question prints a formatted message used in conjunction with a user prompt.
func Question(text string) {
	Questionf("%s", text)
}

// Questionf prints a formatted message used in conjunction with a user prompt.
func Questionf(format string, a ...any) {
	pterm.Println()
	message := Paragraph(format, a...)
	pterm.FgLightGreen.Println(message)
}

// Note prints a formatted yellow message.
func Note(text string) {
	Notef("%s", text)
}

// Notef prints a formatted yellow message.
func Notef(format string, a ...any) {
	pterm.Println()
	message := Paragraph(format, a...)
	pterm.FgYellow.Println(message)
}

// HeaderInfof prints a large header with a formatted message.
func HeaderInfof(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	// Ensure the text is consistent for the header width
	padding := 85 - len(message)
	pterm.Println()
	pterm.DefaultHeader.
		WithBackgroundStyle(pterm.NewStyle(pterm.BgDarkGray)).
		WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
		WithMargin(2).
		Printfln(message + strings.Repeat(" ", padding))
}

// HorizontalRule prints a white horizontal rule to separate the terminal
func HorizontalRule() {
	pterm.Println()
	pterm.Println(strings.Repeat("━", 100))
}

// HorizontalNoteRule prints a yellow horizontal rule to separate the terminal
func HorizontalNoteRule() {
	pterm.Println()
	pterm.FgYellow.Println(strings.Repeat("━", 100))
}

// JSONValue prints any value as JSON.
func JSONValue(value any) string {
	bytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		Debug(err, "ERROR marshalling json")
	}
	return string(bytes)
}

// Paragraph formats text into a 100 column paragraph
func Paragraph(format string, a ...any) string {
	return Paragraphn(100, format, a...)
}

// Paragraphn formats text into an n column paragraph
func Paragraphn(n int, format string, a ...any) string {
	return pterm.DefaultParagraph.WithMaxWidth(n).Sprintf(format, a...)
}

// PrintDiff prints the differences between a and b with a as original and b as new
func PrintDiff(textA, textB string) {
	dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(textA, textB, true)

	diffs = dmp.DiffCleanupSemantic(diffs)

	pterm.Println(dmp.DiffPrettyText(diffs))
}

func debugPrinter(offset int, a ...any) {
	printer := pterm.Debug.WithShowLineNumber(logLevel > 2).WithLineNumberOffset(offset)
	now := time.Now().Format(time.RFC3339)
	// prepend to a
	a = append([]any{now, " - "}, a...)

	printer.Println(a...)

	// Always write to the log file
	if useLogFile {
		pterm.Debug.
			WithShowLineNumber(true).
			WithLineNumberOffset(offset).
			WithDebugger(false).
			WithWriter(logFile).
			Println(a...)
	}
}

func errorPrinter(offset int) *pterm.PrefixPrinter {
	return pterm.Error.WithShowLineNumber(logLevel > 2).WithLineNumberOffset(offset)
}
