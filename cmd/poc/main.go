package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/vshn/odootools/pkg/odoo"
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
	c := odoo.NewClient(url, db)
	sess, err := c.Login(login, password)
	if err != nil {
		log.Fatalf("Authentication failed: %v\n", err)
	}

	attendances := readAttendances(url, sess.ID, sess.UID)

	apd := make(map[string]time.Duration)
	start := 0
	if attendances[0].Action == "sign_in" {
		// last entry was "sign in", so we can't count that one yet
		start = 1
	}
	for i := start; i < len(attendances)-1; i += 2 {
		signIn := attendances[i+1].Name.ToTime()
		signOut := attendances[i].Name.ToTime()
		date := signOut.Format(AttendanceDateFormat)
		apd[date] += signOut.Sub(signIn)
	}

	rows := make([]string, 0, len(apd))
	for date, hours := range apd {
		rows = append(rows, fmt.Sprintf("%s %.2f", date, hours.Hours()))
	}
	sort.Strings(rows)
	for _, row := range rows {
		fmt.Println(row)
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
	Name *AttendanceTime `json:"name,omitempty"`

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

const (
	AttendanceDateFormat     = "2006-01-02"
	AttendanceTimeFormat     = "15:04:05"
	AttendanceDateTimeFormat = AttendanceDateFormat + " " + AttendanceTimeFormat
)

type AttendanceTime time.Time

func (at *AttendanceTime) String() string {
	t := time.Time(*at)
	return t.Format(AttendanceDateTimeFormat)
}
func (at AttendanceTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, at.String())), nil
}
func (at *AttendanceTime) UnmarshalJSON(b []byte) error {
	ts := bytes.Trim(b, `"`)
	t, err := time.Parse(AttendanceDateTimeFormat, string(ts))
	if err != nil {
		return err
	}

	*at = AttendanceTime(t)
	return nil
}
func (at *AttendanceTime) ToTime() time.Time {
	return time.Time(*at)
}
