package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// define a level type to represent the severity level for log entries
type Level int8

// create const which represent severity level
const (
	LevelInfo  Level = iota // has val 0
	LevelError              // has val 1
	LevelFatal              // has val 2
	LevelOff                // has val 3
)

// get the string representation of the level
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}

}

// define custom Logger type to hold the output destination that the log
// will write to, minimum severity level that log entries will be written for,
// mutex for writing cordination
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

func NewLogger(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}

func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}

func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1) // for FATAL entry level
}

// print is method to write the log internally
func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {

	// if the severity level is of the log entry is below minimum level for log
	// then return with no action
	if level < l.minLevel {
		return 0, nil
	}

	// create a struct to hold log entry
	record := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC1123),
		Message:    message,
		Properties: properties,
	}
	// include a stack trace for entries at the Error or Fatal levels
	if level >= LevelError {
		record.Trace = string(debug.Stack())
	}
	// declare a line variable for log entry test
	var text []byte

	// Marshal the record struct for writing it to text variable
	// if there is error print that error to the logger
	text, err := json.Marshal(record)
	if err != nil {
		text = []byte(LevelError.String() + ": unable to write log message: " + err.Error())
	}

	//Lock the mutex so that no two writes to the output destination can happen
	// if there is concurrent write , there my be write overlap without mutex
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.out.Write(append(text, '\n'))

}

// implement Write() method on our Logger type so thatit satisfy  the io.Writer
// interface,
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.print(LevelError, string(message), nil)
}
