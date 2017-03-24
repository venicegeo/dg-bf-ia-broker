// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

// various constants representing the levels of severity for a given audit message
const (
	FATAL    = 0
	CRITICAL = 2
	ERROR    = 3
	WARN     = 4
	NOTICE   = 5
	INFO     = 6
	DEBUG    = 7
)

// LogContext is an interface for all context objects, usable in logging
type LogContext interface {
	AppName() string    // The name of the calling application - "pzsvc-ossim", as an example
	SessionID() string  // Used in logs to indicate which session an event is associated with
	LogRootDir() string // The root directory that has all associated go packages that use pzsvc logging.  Helps keep file locs short.
}

// logFunc is the function used to add entries to the log
var (
	logFunc func(string)
)

func init() {
	logFunc = func(logString string) {
		fmt.Println(logString)
	}
}

// logMessage receives a string to put to the logs.  It formats it correctly
// and puts it in the right place.  This function exists partially in order
// to simplify the task of modifying log behavior in the future.  Note that
// logMessage will panic if no baseLogFunc has been set.  This is a feature,
// not a bug.  It helps you identify threads that have not been properly
// readied.  If logMessage panics in this way, the appropriate answer is
// to call ReadyLog before the first call to logMessage.
func logMessage(lc LogContext, prefix, message string) {
	_, file, line, _ := runtime.Caller(2)
	if lc.LogRootDir() != "" {
		splits := strings.SplitAfter(file, lc.LogRootDir())
		if len(splits) > 1 {
			file = lc.LogRootDir() + splits[len(splits)-1]
		}
	}
	outMsg := fmt.Sprintf("%s - [%s:%s %s %d] %s", prefix, lc.AppName(), lc.SessionID(), file, line, message)
	logFunc(outMsg)
}

// LogInfo posts a logMessage call for standard, non-error messages.  The
// point is mostly to maintain uniformity of appearance and behavior.
func LogInfo(lc LogContext, message string) {
	logMessage(lc, "INFO", message)
}

// LogAlert posts a logMessage call for messages that suggest that someone
// may be attempting to breach the security of the program, or point to the
// possibility of a significant security vulnerability.  The point of this
// function is mostly to maintain uniformity of appearance and behavior.
func LogAlert(lc LogContext, message string) {
	logMessage(lc, "ALERT", message)
}

// LogSimpleErr posts a logMessage call for simple error messages, and produces a pzsvc.Error
// from the result.  The point is mostly to maintain uniformity of appearance and behavior.
func LogSimpleErr(lc LogContext, message string, err error) LoggedError {
	// If by some chance we get back our own error message, catch it appropriately
	if ourError, ok := err.(*Error); ok {
		ourError.Log(lc, message)
		message += err.Error()
	} else {
		if err != nil {
			message += err.Error()
		}
		logMessage(lc, "ERROR", message)
	}
	return fmt.Errorf(message)
}

// LogAuditInput is the set of inputs for the LogAudit function
type LogAuditInput struct {
	Actor    string
	Action   string
	Actee    string
	Message  string
	Severity int
	Response *http.Response // only for LogAuditResponse
}

// LogAudit posts a logMessage call for messages that are generated to
// conform to Audit requirements.  This function is intended to maintain
// uniformity of appearance and behavior, and also to ease maintainability
// when routing requirements change.
func LogAudit(lc LogContext, input LogAuditInput) {
	time := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")

	hostName, _ := os.Hostname()
	outStr := fmt.Sprintf(`<%d>1 %s %s %s - ID%d [pzaudit@48851 actor="%s" action="%s" actee="%s"] %s`,
		8+input.Severity, time, hostName, lc.AppName(), os.Getpid(), input.Actor, input.Action, input.Actee, input.Message)
	logFunc(outStr)
}

// LogAuditResponse is LogAudit for those cases where it needs to include an HTTP response
// body, and that body is not being conveniently read and outputted by some other function.
// It reads the response, logs the result, and replaces the consumed response body with a
// fresh one made from the read buffer, so that it doesn't interfere with any other function
// that woudl wish to access the body.
func LogAuditResponse(lc LogContext, input LogAuditInput) {
	bbuff, _ := ioutil.ReadAll(input.Response.Body)
	input.Response.Body.Close()
	input.Response.Body = ioutil.NopCloser(bytes.NewBuffer(bbuff))
	input.Message = strings.Replace(string(bbuff), "\n", "", -1)
	LogAudit(lc, input)
}

// LoggedError is a duplicate of the "error" interface.  Its real point is to
// indicate, when it is returned from a function, that the error it represents
// has already been entered intot he log and does not need to be entered again.
// The string contained in the LoggedError should be a relatively simple
// description of the error, suitable for returning to the caller of a REST
// interface.
type LoggedError error

// A BasicLogContext is generally only used for testing or when no other
// context information is available.
type BasicLogContext struct {
	sessionID string
}

// AppName returns a hard-coded string
func (tc *BasicLogContext) AppName() string {
	return "bf-ia-broker"
}

// SessionID returns a session ID, creating one if needed
func (tc *BasicLogContext) SessionID() string {
	if tc.sessionID == "" {
		tc.sessionID, _ = PsuUUID()
	}
	return tc.sessionID
}

// LogRootDir returns an empty string
func (tc *BasicLogContext) LogRootDir() string {
	return ""
}
