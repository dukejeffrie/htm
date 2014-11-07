// This code implements a configurable logger to inspect the state of the system.
//
// Usage:
// 1. Run the program with --htmlog=<filename>. Use special key '-' for Stdout.
// 2. Call log.HtmLogger.SetEnabled(true) to start logging.
//
// The logger can be enabled and disabled as often as needed by the program, but
// the file will not be reset.

package log

import "flag"
import "fmt"
import "os"
import orig "log"

const htmLoggerPrefix = "htm) "

var HtmLogger logger
var htmLoggerFilename = flag.String("htmlog", "",
	"Filename to write the trace log to.")

func init() {
	HtmLogger = logger{nil, false}
}

type logger struct {
	delegate *orig.Logger
	enabled  bool
}

func (l *logger) Print(args ...interface{}) {
	if l.Enabled() {
		l.delegate.Print(args)
	}
}

func (l *logger) Printf(format string, args ...interface{}) {
	if l.Enabled() {
		l.delegate.Printf(format, args...)
	}
}

func (l logger) Enabled() bool {
	return l.enabled && l.delegate != nil
}

func (l *logger) SetEnabled(value bool) {
	l.enabled = value
	if value && l.delegate == nil && len(*htmLoggerFilename) > 0 {
		l.delegate = createDelegate()
	}
}

func (l logger) String() string {
	return fmt.Sprint("HTM.TraceLog(\"%s\",enabled=%t,initialized=%t)",
		*htmLoggerFilename, l.enabled, l.delegate != nil)
}

func createDelegate() *orig.Logger {
	if *htmLoggerFilename == "-" {
		return orig.New(os.Stdout, htmLoggerPrefix, 0)
	}
	f, err := os.Create(*htmLoggerFilename)
	if err != nil {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not open file '%s' for writing. HTM trace log will be disabled.", *htmLoggerFilename)
			return nil
		}
	}
	return orig.New(f, htmLoggerPrefix, 0)
}
