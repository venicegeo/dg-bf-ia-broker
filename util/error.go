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
	"net/http"
)

// Error is intended as a somewhat more full-featured way of handling the
// error niche
type Error struct {
	hasLogged  bool   // whether or not this Error has been logged
	LogMsg     string // message to enter into logs
	SimpleMsg  string // simplified message to return to user via rest endpoint
	Request    string // http request body associated with the error (if any)
	Response   string // http response body assocaited with the error (if any)
	URL        string // url associated with the error (if any)
	HTTPStatus int    // http status associated with the error (if any)
}

//
// // Error is a type designed for easy serialization to JSON
// type Error struct {
// 	Message string `json:"error"`
// }
//
// func (err Error) Error() string {
// 	return err.Message
// }
//

// GenExtendedMsg is used to generate extended log messages from Error objects
// for the cases where that's appropriate
func (err Error) GenExtendedMsg() string {
	lineBreak := "\n/**************************************/\n"
	outBody := "Http Error: " + err.LogMsg + lineBreak
	if err.URL != "" {
		outBody += "\nURL: " + err.URL + "\n"
	}
	if err.Request != "" {
		outBody += "\nRequest: " + err.Request + "\n"
	}
	if err.Response != "" {
		outBody += "\nResponse: " + err.Response + "\n"
	}
	if http.StatusText(err.HTTPStatus) != "" {
		outBody += "\nHTTP Status: " + http.StatusText(err.HTTPStatus) + "\n"
	}
	outBody += lineBreak
	return outBody
}

// Log is intended as the base way to generate logging information for an Error
// object.  It constructs an extended error if necessary, gathers the filename
// and line number data, and sends it to logMessage for formatting and output.
// It also ensures that any given error will only be logged once, and will be
// logged at the lowest level that calls for it.  In particular, the general
// expectation is that the message will be generated at a relatively low level,
// and then logged with additional context at some higher position.  Given our
// general level of complexity, that strikes a decent balance between providing
// enough detail to figure out the cause of an error and keepign thigns simple
// enough to readily understand.
func (err *Error) Log(s LogContext, msgAdd string) LoggedError {
	if !err.hasLogged {
		if msgAdd != "" {
			err.LogMsg = msgAdd + ": " + err.LogMsg
		}
		outMsg := err.LogMsg
		if err.Request != "" || err.Response != "" {
			outMsg = err.GenExtendedMsg()
		}
		logMessage(s, "ERROR", outMsg)
		err.hasLogged = true
	} else {
		logMessage(s, "ERROR", "Meta-error.  Tried to log same message for a second time.")
	}
	return fmt.Errorf(err.Error())
}

// Error here is intended to let pzsvc.Error objects serve the error interface, and,
// by extension, to let them be passed around as interfaces in palces that aren't
// importing pzsvc-lib and used in a reasonable manner
func (err Error) Error() string {
	if err.SimpleMsg != "" {
		return err.SimpleMsg
	}
	return err.LogMsg
}
