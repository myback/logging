package logging

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Logger struct {
	DebugSet       bool
	Stdout, Stderr io.Writer
}

func New(out, e io.Writer, debug bool) *Logger {
	return &Logger{
		DebugSet: debug,
		Stdout:   out,
		Stderr:   e,
	}
}

func (l *Logger) Errorf(format string, msg ...interface{}) {
	l.stderrWrite("error", fmt.Sprintf(format, msg...))
}

func (l *Logger) Error(msg interface{}) {
	l.stderrWrite("error", msgWrap(msg))
}

func (l *Logger) Infof(format string, msg ...interface{}) {
	l.stdoutWrite("info", fmt.Sprintf(format, msg...))
}

func (l *Logger) Info(msg string) {
	l.stdoutWrite("info", msgWrap(msg))
}

func (l *Logger) Debugf(format string, msg ...interface{}) {
	if l.DebugSet {
		l.stdoutWrite("debug", fmt.Sprintf(format, msg...))
	}
}

func (l *Logger) Debug(msg string) {
	if l.DebugSet {
		l.stdoutWrite("debug", msgWrap(msg))
	}
}

func (l *Logger) Fatalf(format string, msg ...interface{}) {
	l.stderrWrite("fatal", fmt.Sprintf(format, msg...))
	os.Exit(127)
}

func (l *Logger) Fatal(msg string) {
	l.stderrWrite("fatal", msgWrap(msg))
	os.Exit(127)
}

func (l *Logger) Access(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var u, h string
		login, _, ok := r.BasicAuth()
		if ok {
			u = login
		} else {
			u = "-"
		}

		if l.DebugSet {
			var hl []string
			for k, v := range r.Header {
				hl = append(hl, fmt.Sprintf(`http_%s="%s"`, k, v[0]))
			}

			h = " " + strings.Join(hl, " ")
		}

		l.stdoutWrite("access", fmt.Sprintf(`user="%s" address="%s" host="%s" uri="%s" method="%s" proto="%s" useragent="%s" referer="%s"`+h,
			u, r.RemoteAddr, r.Host, r.RequestURI, r.Method, r.Proto, r.UserAgent(), r.Referer()))

		next.ServeHTTP(w, r)
	})
}

func (l *Logger) stdoutWrite(level, msg string) {
	_, err := fmt.Fprintf(l.Stdout, `time="%s" level="%s" %s\n`, time.Now().UTC().Format("2006-01-02T15:04:05.000Z"), level, msg)
	if err != nil {
		fmt.Printf("Log write failed: %s", err)
	}

}

func (l *Logger) stderrWrite(level, msg string) {
	_, err := fmt.Fprintf(l.Stderr, `time="%s" level="%s" %s\n`, time.Now().UTC().Format("2006-01-02T15:04:05.000Z"), level, msg)
	if err != nil {
		fmt.Printf("Log write failed: %s", err)
	}

}

func msgWrap(msg interface{}) string {
	return fmt.Sprintf(`message="%s"`, msg)
}
