package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

var debug = false

func main() {
	// Flags
	var login, passwordFile, db, url string
	flag.StringVar(&login, "u", "", "Odoo Login name")
	flag.StringVar(&passwordFile, "p", "", "File containing the Odoo password")
	flag.StringVar(&db, "db", "", "Odoo DB name")
	flag.StringVar(&url, "url", "", "Odoo URL")
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.Parse()

	// Read password
	pwBytes, err := ioutil.ReadFile(passwordFile)
	if err != nil {
		log.Fatalln(err)
	}
	password := strings.TrimSpace(string(pwBytes))
	log.Printf("Connecting to Odoo at '%s' (using DB '%s') as user '%s'", url, db, login)

	sess := doLogin(url, db, login, password)
	attendances := readAttendances(url, sess.ID, sess.UID)

	for _, a := range attendances {
		fmt.Printf("[%d] %s %8s %s\n", a.ID, a.Name, a.Action, *a.ActionDesc)
	}
}

// JsonRpcRequest represents a generic json-rpc request
type JsonRpcRequest struct {
	// ID should be a randomly generated value, either as a string or int. The
	// server will return this value in the response.
	ID string `json:"id,omitempty"`

	// Jsonrpc is always set to "2.0"
	Jsonrpc string `json:"jsonrpc,omitempty"`

	// Method to call, usually just "call"
	Method string `json:"method,omitempty"`

	// Params includes the actual request payload.
	Params map[string]interface{} `json:"params,omitempty"`
}

type AuthenticateResponse struct {
	// ID that was sent with the request
	ID string `json:"id,omitempty"`
	// Jsonrpc is always set to "2.0"
	Jsonrpc string `json:"jsonrpc,omitempty"`
	// Result payload
	Result Session `json:"result,omitempty"`
}

type Session struct {
	// ID is the session ID.
	// Is always set, no matter the authentication outcome.
	ID string `json:"session_id,omitempty"`

	// UID is the user's ID as an int, or the boolean `false` if authentication
	// failed.
	UID int `json:"uid,omitempty"`

	// Username is usually set to the LoginName that was sent in the request.
	// Is always set, no matter the authentication outcome.
	Username string `json:"username,omitempty"`
}

type ReadAttendancesResponse struct {
	// ID that was sent with the request
	ID string `json:"id,omitempty"`
	// Jsonrpc is always set to "2.0"
	Jsonrpc string `json:"jsonrpc,omitempty"`
	// Result payload
	Result ReadAttendanceResult `json:"result,omitempty"`
}

type ReadAttendanceResult struct {
	// Length is the total number of records in the DB
	Length int `json:"length,omitempty"`

	// Records includes a subset of the records based on the "offset" and
	// "limit" values in the request.
	Records []Attendance `json:"records,omitempty"`
}

type Attendance struct {
	// ID is an unique ID for each attendance entry
	ID int `json:"id,omitempty"`

	// Name is the entry timestamp in UTC
	// Format: '2006-01-02 15:04:05'
	Name string `json:"name,omitempty"`

	// Action is either "sign_in" or "sign_out"
	Action string `json:"action,omitempty"`
	// ActionDesc describes the "action reason" from Odoo.
	//
	// Example values:
	// * `[1, "Outside office hours"]`
	// * `[2, "Outside office hours"]`
	// * `[4, "Sick / Medical Consultation"]`
	ActionDesc *json.RawMessage `json:"action_desc,omitempty"`
}
