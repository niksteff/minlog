package minlog

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"time"
)

type Level int8

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	DPanicLevel
	PanicLevel
	FatalLevel
)

// Option is our return type for the functional option pattern
type Option func(*config)

// WithTimeformat sets the desired time format
func WithTimeformat(format string) Option {
	return func(c *config) {
		c.dateformat = format
	}
}

// WithMinLevel sets the min logging level for the logger
func WithMinLevel(min Level) Option {
	return func(c *config) {
		c.minLevel = min
	}
}

// WithMaxLevel sets the max logging level for the logger
func WithMaxLevel(max Level) Option {
	return func(c *config) {
		c.maxLevel = max
	}
}

func WithTarget(target *os.File) Option {
	return func(c *config) {
		c.target = target
	}
}

func WithDateUpdateInterval(interval int) Option {
	return func(c *config) {
		c.dateUpdateInterval = interval
	}
}

type config struct {
	dateformat         string
	dateUpdateInterval int
	minLevel           Level
	maxLevel           Level
	target             *os.File
}

type MinLog struct {
	conf    config
	pool    *sync.Pool
	logTime string
	out     chan func() []byte
}

func New(opts ...Option) MinLog {
	conf := config{
		dateformat:         time.RFC3339, // use a default format with second precision
		minLevel:           DebugLevel,   // start from DebugLevel
		maxLevel:           FatalLevel,   // allow up to FatalLevel
		dateUpdateInterval: 1000,         // log date update interval in ms
	}

	// apply the given options
	for _, opt := range opts {
		opt(&conf)
	}

	bufPool := &sync.Pool{
		New: func() any {
			return bytes.NewBuffer(make([]byte, 0, 512))
		},
	}

	m := MinLog{
		pool:    bufPool,
		conf:    conf,
		out:     make(chan func() []byte),
		logTime: time.Now().Format(conf.dateformat),
	}

	// update the date string in the given interval to not render it in every
	// log on its own.
	timeTicker := time.NewTicker(time.Millisecond * time.Duration(conf.dateUpdateInterval))
	go func() {
		for range timeTicker.C {
			m.logTime = time.Now().Format(conf.dateformat)
		}
	}()

	// run a loop writing the incoming logs to the output buffer
	go m.writer()

	return m
}

func (l *MinLog) writer() {
	for log := range l.out {
		_, _ = l.conf.target.Write(log()) // TODO: should check the return value
	}
}

func composeMsg(msg string, args ...any) string {
	if len(args) == 0 {
		return msg
	}

	if msg != "" {
		return fmt.Sprintf(msg, args...)
	}

	return fmt.Sprint(args...)
}

// Log writes a message to the out stream
func (l *MinLog) Log(level Level, msg string, args ...any) {
	if level < l.conf.minLevel || level > l.conf.maxLevel {
		// do not write a log when the level is out of min max bounds
		return
	}

	// write the log
	l.out <- func() []byte {
		b := l.pool.Get().(*bytes.Buffer)
		defer l.pool.Put(b)

		b.Reset()

		b.WriteString(l.logTime)  // write the log time
		b.WriteString(" [INFO] ") // write the log level

		msg := composeMsg(msg, args...) // build the log message
		b.WriteString(msg)              // write the log message
		b.WriteByte('\n')               // add the new line char

		return b.Bytes()
	}
}

func (l *MinLog) Info(msg string) {
	l.Log(InfoLevel, msg)
}

func (l *MinLog) Infof(msg string, args any) {
	l.Log(InfoLevel, msg, args)
}
