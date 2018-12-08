package main

import (
	"bufio"
	"flag"
	// "errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"time"
)

type LogLine struct {
	remote_addr string
	remote_user string
	time_local  string
	// time_local      time.Time
	request         string
	status          string
	body_bytes_sent string
	http_referer    string
	http_user_agent string
}

func ConvertLogLine(logLine string) LogLine {
	re := regexp.MustCompile(`(?P<remote_addr>\d+\.\d+\.\d+\.\d+) - - \[(?P<time_local>[^\]]+)\] \"(?P<request>.*)\" (?P<status>[0-9]+) (?P<body_bytes_sent>[0-9]+) \"(?P<http_referer>.*)\" \"(?P<http_user_agent>.+)\"`)

	m := reSubMatchMap(re, logLine)

	parsedLine := LogLine{
		remote_addr:     m["remote_addr"],
		remote_user:     m["remote_user"],
		time_local:      m["time_local"],
		request:         m["request"],
		status:          m["status"],
		body_bytes_sent: m["body_bytes_sent"],
		http_referer:    m["http_referer"],
		http_user_agent: m["http_user_agent"],
	}
	return parsedLine
}

func reSubMatchMap(r *regexp.Regexp, str string) map[string]string {
	match := r.FindStringSubmatch(str)
	subMatchMap := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}
	return subMatchMap
}

func sortByPopularity(metric map[string]int) {
	n := map[int][]string{}
	var a []int
	for k, v := range metric {
		n[v] = append(n[v], k)
	}
	for k := range n {
		a = append(a, k)
	}
	sort.Sort(sort.IntSlice(a))
	for _, k := range a {
		for _, s := range n[k] {
			fmt.Printf("%d - %s\n", k, s)
		}
	}

}

func dateIsInInterval(line string, period string) bool {
	now := time.Now()
	var startDate time.Time
	switch period {
	case "week":
		duration, _ := time.ParseDuration("168h")
		startDate = now.Add(-duration)

	case "month":
		startDate = now.AddDate(0, -1, 0)
	default:
		return false
	}
	t, _ := time.Parse("02/Jan/2006:15:04:05 -0700", line)
	return startDate.Before(t)

}

func LogLineString(logLine LogLine) string {
	return fmt.Sprintf(`%s - - [%s] "%s" %s %s "%s" "%s"`, logLine.remote_addr, logLine.time_local, logLine.request, logLine.status, logLine.body_bytes_sent, logLine.http_referer, logLine.http_user_agent)
}

// func filter (str string, re *regexp) bool {

// }

func main() {

	period := flag.String("period", "week", "Period before current date to get logs. Ex: week|month")
	flag.Parse()
	if (*period != "week") && (*period != "month") {
		fmt.Printf("Wrong period value. Should be 'week' or 'moth', got '%s'\n", *period)
		os.Exit(1)

	}

	i := 0

	scanner := bufio.NewScanner(os.Stdin)
	userAgents := make(map[string]int)
	referers := make(map[string]int)
	requests := make(map[string]int)
	statuses := make(map[string]int)
	logLines := make([]LogLine, 10000)

	for scanner.Scan() {
		// `Text` returns the current token, here the next line,
		// from the input.

		// fmt.Println(scanner.Text())
		parsedLine := ConvertLogLine(scanner.Text())
		i += 1
		// Write out the uppercased line.
		logLines = append(logLines, parsedLine)
		// if i > 40 {
		// 	break
		// }

	}
	for line := range logLines {
		match, _ := regexp.MatchString(`GET (.+\.html|\/) HTTP\/1\.[1]`, logLines[line].request)
		if match != true {
			continue
		}
		match, _ = regexp.MatchString(`.*([Bb]ot|vkShare|Google-AMPHTML|feedly|[cC]rawler|[Pp]arser|curl|-).*`, logLines[line].http_user_agent)
		if match == true {
			continue
		}
		if !dateIsInInterval(logLines[line].time_local, *period) {
			continue
		}

		userAgents[logLines[line].http_user_agent] += 1
		referers[logLines[line].http_referer] += 1
		requests[logLines[line].request] += 1
		statuses[logLines[line].status] += 1
		fmt.Println(LogLineString(logLines[line]))
	}
	// sortByPopularity(userAgents)

	// Check for errors during `Scan`. End of file is
	// expected and not reported by `Scan` as an error.
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

}
