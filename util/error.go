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
	request    string // http request body associated with the error (if any)
	response   string // http response body assocaited with the error (if any)
	url        string // url associated with the error (if any)
	httpStatus int    // http status associated with the error (if any)
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
	if err.url != "" {
		outBody += "\nURL: " + err.url + "\n"
	}
	if err.request != "" {
		outBody += "\nRequest: " + err.request + "\n"
	}
	if err.response != "" {
		outBody += "\nResponse: " + err.response + "\n"
	}
	if http.StatusText(err.httpStatus) != "" {
		outBody += "\nHTTP Status: " + http.StatusText(err.httpStatus) + "\n"
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
		if err.request != "" || err.response != "" {
			outMsg = err.GenExtendedMsg()
		}
		logMessage(s, "ERROR", outMsg)
		err.hasLogged = true
	} else {
		logMessage(s, "ERROR", "Meta-error.  Tried to log same message for a second time.")
	}
	return fmt.Errorf(err.Error())
}

// LogSimpleErr posts a logMessage call for simple error messages, and produces a pzsvc.Error
// from the result.  The point is mostly to maintain uniformity of appearance and behavior.
func LogSimpleErr(s LogContext, message string, err error) LoggedError {
	if err != nil {
		message += err.Error()
	}
	logMessage(s, "ERROR", message)
	return fmt.Errorf(message)
}

// Error here is intended to let pzsvc.Error objects serve the error interface, and,
// by extension, to let them be passed around as interfaces in palces that aren't
// importing pzsvc-lib and used in a reasonable manner
func (err Error) Error() string {
	fmt.Print("E1")
	if err.SimpleMsg != "" {
		fmt.Print("E2")
		return err.SimpleMsg
	}
	fmt.Print("E3")
	return err.LogMsg
}
