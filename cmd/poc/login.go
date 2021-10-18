package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func doLogin(url, db, login, password string) Session {
	// Prepare login request
	loginRequest := JsonRpcRequest{
		ID:      uuid.NewString(),
		Jsonrpc: "2.0",
		Method:  "call",
		Params: map[string]interface{}{
			"db":    db,
			"login": login,
		},
	}
	if debug {
		log.Printf("LoginRequest: %#v\n", loginRequest)
	}
	loginRequest.Params["password"] = password
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(&loginRequest); err != nil {
		log.Fatalln(err)
	}

	// Send login request
	res, err := http.Post(url+"/web/session/authenticate", "application/json", buf)
	if err != nil {
		log.Fatalln(err)
	}
	if res.StatusCode != http.StatusOK {
		log.Fatalf("Got unexpected HTTP status: %d %s", res.StatusCode, res.Status)
	}

	// Decode login response
	var loginResponse AuthenticateResponse
	if err := json.NewDecoder(res.Body).Decode(&loginResponse); err != nil {
		log.Fatalln(err)
	}
	if debug {
		log.Printf("LoginResponse: %#v\n", loginResponse)
	}
	return loginResponse.Result
}
