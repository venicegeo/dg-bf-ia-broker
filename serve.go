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

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/venicegeo/bf-ia-broker/planet"
	"github.com/venicegeo/bf-ia-broker/util"
)

func serve() {

	portStr := ":8080"
	router := mux.NewRouter()

	context := &(util.BasicLogContext{})

	util.LogAudit(context, util.LogAuditInput{Actor: "serve()", Action: "startup", Actee: "self", Message: "Application Startup", Severity: util.INFO})

	router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		util.LogAudit(context, util.LogAuditInput{Actor: "anon user", Action: request.Method, Actee: request.URL.String(), Message: "Receiving / request", Severity: util.INFO})
		fmt.Fprintf(writer, "Hi")
		util.LogAudit(context, util.LogAuditInput{Actor: request.URL.String(), Action: request.Method + " response", Actee: "anon user", Message: "Sending / response", Severity: util.INFO})
	})
	router.HandleFunc("/planet/discover/{itemType}", planet.DiscoverHandler)
	router.HandleFunc("/planet/{itemType}/{id}", planet.MetadataHandler)
	router.HandleFunc("/planet/activate/{itemType}/{id}", planet.ActivateHandler)
	// 	case "/help":
	// 		fmt.Fprintf(writer, "We're sorry, help is not yet implemented.\n")
	// 	default:
	// 		fmt.Fprintf(writer, "Command undefined. \n")
	// 	}
	// })
	http.Handle("/", router)

	log.Fatal(http.ListenAndServe(portStr, nil))
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve Broker",
	Long: `
Serve the image archive broker`,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}
