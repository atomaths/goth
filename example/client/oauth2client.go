package main

// Use code.google.com/p/goauth2/oauth client to test
// Open url in browser:
// http://localhost:14000/app

import (
	"fmt"
	"net/http"
	"net/url"

	"code.google.com/p/goauth2/oauth"
	storage "github.com/atomaths/goth/storage/mongo"
	"github.com/atomaths/osin"
)

func main() {
	config := osin.NewServerConfig()
	// goauth2 checks errors using status codes
	config.ErrorStatusCode = 401
	// goauth2 passes client secret in params instead of Authorization headers
	config.AllowClientSecretInParams = true

	server := osin.NewServer(config, storage.NewTestStorage())
	output := osin.NewResponseOutputJSON()

	client := &oauth.Config{
		ClientId:     "1234",
		ClientSecret: "aabbccdd",
		//RedirectURL:  "http://localhost:14000/appauth/code",
		//AuthURL:      "http://localhost:14000/authorize",
		//TokenURL:     "http://localhost:14000/token",
		RedirectURL: "http://192.168.10.232:14001/appauth/code",
		AuthURL:     "http://192.168.10.232:14000/authorize",
		TokenURL:    "http://192.168.10.232:14000/token",
	}
	ctransport := &oauth.Transport{Config: client}

	// Authorization code endpoint
	http.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		resp := server.NewResponse()
		if ar := server.HandleAuthorizeRequest(resp, r); ar != nil {
			if !HandleLoginPage(ar, w, r) {
				return
			}
			ar.Authorized = true
			server.FinishAuthorizeRequest(resp, r, ar)
		}
		if resp.IsError && resp.InternalError != nil {
			fmt.Printf("ERROR: %s\n", resp.InternalError)
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
		if resp.IsError && resp.InternalError != nil {
			fmt.Printf("ERROR: %s\n", resp.InternalError)
		}
		output.Output(resp, w, r)
	})

	// Information endpoint
	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		resp := server.NewResponse()
		if ir := server.HandleInfoRequest(resp, r); ir != nil {
			server.FinishInfoRequest(resp, r, ir)
		}
		output.Output(resp, w, r)
	})

	// Application home endpoint
	http.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body>"))
		//w.Write([]byte(fmt.Sprintf("<a href=\"/authorize?response_type=code&client_id=1234&state=xyz&scope=everything&redirect_uri=%s\">Login</a><br/>", url.QueryEscape("http://localhost:14000/appauth/code"))))
		w.Write([]byte(fmt.Sprintf("<a href=\"%s\">Login</a><br/>", client.AuthCodeURL(""))))
		w.Write([]byte("</body></html>"))
	})

	// Application destination - CODE
	http.HandleFunc("/appauth/code", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		code := r.Form.Get("code")

		w.Write([]byte("<html><body>"))
		w.Write([]byte("APP AUTH - CODE<br/>"))

		if code != "" {

			var jr *oauth.Token
			var err error

			// if parse, download and parse json
			if r.Form.Get("doparse") == "1" {
				jr, err = ctransport.Exchange(code)
				if err != nil {
					jr = nil
					w.Write([]byte(fmt.Sprintf("ERROR: %s<br/>\n", err)))
				}
			}

			// show json access token
			if jr != nil {
				w.Write([]byte(fmt.Sprintf("ACCESS TOKEN: %s<br/>\n", jr.AccessToken)))
				if jr.RefreshToken != "" {
					w.Write([]byte(fmt.Sprintf("REFRESH TOKEN: %s<br/>\n", jr.RefreshToken)))
				}
			}

			w.Write([]byte(fmt.Sprintf("FULL RESULT: %+v<br/>\n", jr)))

			cururl := *r.URL
			curq := cururl.Query()
			curq.Add("doparse", "1")
			cururl.RawQuery = curq.Encode()
			w.Write([]byte(fmt.Sprintf("<a href=\"%s\">Download Token</a><br/>", cururl.String())))
		} else {
			w.Write([]byte("Nothing to do"))
		}

		w.Write([]byte("</body></html>"))
	})

	//http.ListenAndServe(":14000", nil)
	http.ListenAndServe("192.168.10.232:14001", nil)
}

// Login page
func HandleLoginPage(ar *osin.AuthorizeRequest, w http.ResponseWriter, r *http.Request) bool {
	r.ParseForm()
	if r.Method == "POST" && r.Form.Get("login") == "test" && r.Form.Get("password") == "test" {
		return true
	}

	w.Write([]byte("<html><body>"))

	w.Write([]byte(fmt.Sprintf("LOGIN %s (use test/test)<br/>", ar.Client.Id)))
	w.Write([]byte(fmt.Sprintf("<form action=\"/authorize?response_type=%s&client_id=%s&state=%s&redirect_uri=%s\" method=\"POST\">",
		ar.Type, ar.Client.Id, ar.State, url.QueryEscape(ar.RedirectUri))))

	w.Write([]byte("Login: <input type=\"text\" name=\"login\" /><br/>"))
	w.Write([]byte("Password: <input type=\"password\" name=\"password\" /><br/>"))
	w.Write([]byte("<input type=\"submit\"/>"))

	w.Write([]byte("</form>"))

	w.Write([]byte("</body></html>"))

	return false
}
