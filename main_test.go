package main

import (
	// "fmt"
	"testing"
)

// type logLine struct {
// 	remote_addr string
// 	remote_user string
// 	time_local  string
// 	// time_local      time.Time
// 	request         string
// 	status          int
// 	body_bytes_sent int
// 	http_referer    string
// 	http_user_agent string
// }

func TestConvertLogLine(t *testing.T) {
	logLine := `52.87.65.11 - - [02/Nov/2018:06:55:13 +0000] "GET /feed.xml HTTP/1.1" 200 9810 "-" "curl"`
	parsedLine := ConvertLogLine(logLine)
	parsedStructure := LogLine{
		remote_addr:     "52.87.65.11",
		remote_user:     "",
		time_local:      "02/Nov/2018:06:55:13 +0000",
		request:         "GET /feed.xml HTTP/1.1",
		status:          "200",
		body_bytes_sent: "9810",
		http_referer:    "-",
		http_user_agent: "curl",
	}

	if parsedLine != parsedStructure {
		t.Errorf("expected:\n%v\n got: %v", parsedStructure, parsedLine)
	}
	logLine = `190.128.131.6 - - [03/Nov/2018:23:25:08 +0000] "" 400 0 "-" "-"`
	parsedLine = ConvertLogLine(logLine)
	parsedStructure = LogLine{
		remote_addr:     "190.128.131.6",
		remote_user:     "",
		time_local:      "03/Nov/2018:23:25:08 +0000",
		request:         "",
		status:          "400",
		body_bytes_sent: "0",
		http_referer:    "-",
		http_user_agent: "-",
	}

	if parsedLine != parsedStructure {
		t.Errorf("expected:\n%v\n got: %v", parsedStructure, parsedLine)
	}
	// fmt.Println(parsedLine)
}
func TestLogLineString(t *testing.T) {
	logLine := `52.87.65.11 - - [02/Nov/2018:06:55:13 +0000] "GET /feed.xml HTTP/1.1" 200 9810 "-" "curl"`
	parsedStructure := LogLine{
		remote_addr:     "52.87.65.11",
		remote_user:     "",
		time_local:      "02/Nov/2018:06:55:13 +0000",
		request:         "GET /feed.xml HTTP/1.1",
		status:          "200",
		body_bytes_sent: "9810",
		http_referer:    "-",
		http_user_agent: "curl",
	}
	outputLine := LogLineString(parsedStructure)
	if outputLine != logLine {
		t.Errorf("expected:\n%v\n got:\n%v", logLine, outputLine)
	}

}
