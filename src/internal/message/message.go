package message

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

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

// NoProgress tracks whether spinner/progress bars show updates
var NoProgress bool

var logLevel = InfoLevel

// Write logs to stderr and a buffer for logfile generation
var logFile *os.File

var useLogFile bool

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

// UseLogfile writes output to stderr and a logfile
func UseLogFile() {
	// Prepend the log filename with a timestampe
	ts := time.Now().Format("2006-01-02-15-04-05")

	// Try to create a temp log file
	if logFile, err := os.CreateTemp("", fmt.Sprintf("zarf-%s-*.log", ts)); err != nil {
		Error(err, "Error saving a log file")
	} else {
		useLogFile = true
		logStream := io.MultiWriter(os.Stderr, logFile)
		pterm.SetDefaultOutput(logStream)
		message := fmt.Sprintf("Saving log file to %s", logFile.Name())
		Note(message)
	}
}

func SetLogLevel(lvl LogLevel) {
	logLevel = lvl
	if logLevel >= DebugLevel {
		pterm.EnableDebugMessages()
	}
}

func GetLogLevel() LogLevel {
	return logLevel
}

func Debug(payload ...any) {
	debugPrinter(2, payload...)
}

func Debugf(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	debugPrinter(3, message)
}

func Error(err any, message string) {
	debugPrinter(2, err)
	Warnf(message)
}

func ErrorWebf(err any, w http.ResponseWriter, format string, a ...any) {
	debugPrinter(2, err)
	message := fmt.Sprintf(format, a...)
	Warn(message)
	http.Error(w, message, http.StatusInternalServerError)
}

func Errorf(err any, format string, a ...any) {
	debugPrinter(2, err)
	Warnf(format, a...)
}

func Warn(message string) {
	Warnf(message)
}

func Warnf(format string, a ...any) {
	message := paragraph(format, a...)
	pterm.Warning.Println(message)
}

func Fatal(err any, message string) {
	debugPrinter(2, err)
	errorPrinter(2).Println(message)
	os.Exit(1)
}

func Fatalf(err any, format string, a ...any) {
	debugPrinter(2, err)
	message := paragraph(format, a...)
	errorPrinter(2).Println(message)
	os.Exit(1)
}

func Info(message string) {
	Infof(message)
}

func Infof(format string, a ...any) {
	if logLevel > 0 {
		message := paragraph(format, a...)
		pterm.Info.Println(message)
	}
}

func SuccessF(format string, a ...any) {
	message := paragraph(format, a...)
	pterm.Success.Println(message)
}

func Question(text string) {
	pterm.Println()
	message := paragraph(text)
	pterm.FgMagenta.Println(message)
}

func Notef(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	Note(message)
}

func Note(text string) {
	pterm.Println()
	message := paragraph(text)
	pterm.FgYellow.Println(message)
}

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

func JsonValue(value any) string {
	bytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		Debug(err, "ERROR marshalling json")
	}
	return string(bytes)
}

func paragraph(format string, a ...any) string {
	return pterm.DefaultParagraph.WithMaxWidth(100).Sprintf(format, a...)
}

func debugPrinter(offset int, a ...any) {
	printer := pterm.Debug.WithShowLineNumber(logLevel > 2).WithLineNumberOffset(offset)
	now := time.Now().Format(time.RFC3339)
	// prepend to a
	a = append([]any{now, " - "}, a...)

	printer.Println(a...)

	// Always write to the log file
	if useLogFile && !pterm.PrintDebugMessages {
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
