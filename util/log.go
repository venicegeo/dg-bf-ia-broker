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
	"fmt"
	"runtime"
	"strings"
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

// logMessage receives a string to put to the logs.  It formats it correctly
// and puts it in the right place.  This function exists partially in order
// to simplify the task of modifying log behavior in the future.  Note that
// logMessage will panic if no baseLogFunc has been set.  This is a feature,
// not a bug.  It helps you identify threads that have not been properly
// readied.  If logMessage panics in this way, the appropriate answer is
// to call ReadyLog before the first call to logMessage.
func logMessage(lc LogContext, prefix, message string) {
	_, file, line, _ := runtime.Caller(2)
	if logFunc == nil {
		logFunc = func(logString string) {
			fmt.Println(logString)
		}
	}
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
func LogAlert(s LogContext, message string) {
	logMessage(s, "ALERT", message)
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
