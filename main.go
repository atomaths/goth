package main

import (
	"net/http"

	storage "github.com/atomaths/goth/storage/mongo"
	"github.com/atomaths/osin"
)

func main() {
	// TestStorage implements the "osin.Storage" interface
	serverConfig := osin.NewServerConfig()
	serverConfig.AllowGetAccessRequest = true
	server := osin.NewServer(serverConfig, storage.NewTestStorage())
	output := osin.NewResponseOutputJSON()

	// Authorization code endpoint
	http.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		resp := server.NewResponse()
		if ar := server.HandleAuthorizeRequest(resp, r); ar != nil {

			// HANDLE LOGIN PAGE HERE

			ar.Authorized = true
			server.FinishAuthorizeRequest(resp, r, ar)
		}
		output.Output(resp, w, r)
	})

	// Access token endpoint
	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		resp := server.NewResponse()
		if ar := server.HandleAccessRequest(resp, r); ar != nil {
			ar.Authorized = true
			server.FinishAccessRequest(resp, r, ar)
		}
		output.Output(resp, w, r)
	})

	http.ListenAndServe("192.168.10.232:14000", nil)
}
