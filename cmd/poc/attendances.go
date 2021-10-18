package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

func readAttendances(url, sid string, uid int) []Attendance {
	// Prepare "read Attendances"
	attendancesRequest := JsonRpcRequest{
		ID:      uuid.NewString(),
		Jsonrpc: "2.0",
		Method:  "call",
		Params: map[string]interface{}{
			"model": "hr.attendance",
			"domain": [][]interface{}{{
				"employee_id.user_id.id",
				"=",
				uid,
			}},
			"fields": []string{"employee_id", "name", "action", "action_desc"},
			"limit":  1000,
			"offset": 0,
		},
	}
	if debug {
		log.Printf("AttendancesRequest: %#v\n", attendancesRequest)
		json.NewEncoder(os.Stderr).Encode(&attendancesRequest)
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(&attendancesRequest); err != nil {
		log.Fatalln(err)
	}

	// Send readAttendances request
	req, err := http.NewRequest("POST", url+"/web/dataset/search_read", buf)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("cookie", "session_id="+sid)
	if debug {
		log.Printf("request: %#v\n", req)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	if res.StatusCode != http.StatusOK {
		log.Fatalf("Got unexpected HTTP status: %d %s", res.StatusCode, res.Status)
	}

	// decode readAttendances response
	var attendanceResponse ReadAttendancesResponse
	if err := json.NewDecoder(res.Body).Decode(&attendanceResponse); err != nil {
		log.Fatalln(err)
	}
	if debug {
		log.Printf("ReadAttendancesResponse: %#v\n", attendanceResponse)
	}

	return attendanceResponse.Result.Records
}
