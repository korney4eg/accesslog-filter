package main

import (
	"bufio"
	// "errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

func ConvertLogLineToMap(logLine string) map[string]string {
	regexPattern := `(?P<remote_addr>\d+\.\d+\.\d+\.\d+) - -` +
		` \[(?P<time_local>[^\]]+)\] \"(?P<request>.*)\" (?P<status>[0-9]+)` +
		` (?P<body_bytes_sent>[0-9]+) \"(?P<http_referer>.*)\" \"` +
		`(?P<http_user_agent>.+)\"`
	re := regexp.MustCompile(regexPattern)
	parsedMap := make(map[string]string)
	match := re.FindStringSubmatch(logLine)
	if match == nil {
		return nil
	}
	for i, name := range re.SubexpNames() {
		if i != 0 {
			parsedMap[name] = match[i]
		}
	}
	return parsedMap
}

func matchAllRequirements(parsedLine map[string]string, period string) bool {
	request := regexp.MustCompile(`GET (.+\.html|\/\d*\/\d*\/\d*\/.*\/|\/\?.*)(\?.*)? HTTP\/1\.[10]`)
	http_user_agent := regexp.MustCompile(`.*([Bb]ot|vkShare|Google-AMPHTML|feedly|[cC]rawler|[Pp]arser|curl|-|Disqus).*`)
	switch {
	case parsedLine["status"] != "200":
		return false
	case !request.MatchString(parsedLine["request"]):
		return false
	case http_user_agent.MatchString(parsedLine["http_user_agent"]):
		return false
	case !dateIsInInterval(parsedLine["time_local"], period):
		return false
	default:
		return true
	}
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

func AnonymizeIp(ip string) string {
	return ip[:strings.LastIndex(ip, ".")+1] + "0"
}

func ConvertMapToLogLine(parsedLine map[string]string) string {
	remote_addr := parsedLine["remote_addr"]
	time_local := parsedLine["time_local"]
	request := parsedLine["request"]
	status := parsedLine["status"]
	body_bytes_sent := parsedLine["body_bytes_sent"]
	http_referer := parsedLine["http_referer"]
	http_user_agent := parsedLine["http_user_agent"]
	return fmt.Sprintf(`%s - - [%s] "%s" %s %s "%s" "%s"`, remote_addr,
		time_local, request, status, body_bytes_sent, http_referer,
		http_user_agent)
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
	case "any":
		return true
	default:
		return false
	}
	t, _ := time.Parse("02/Jan/2006:15:04:05 -0700", line)
	return startDate.Before(t)
}

func main() {

	period := flag.String("period", "week", "Period before current date to get logs. Ex: week|month")
	flag.Parse()
	if (*period != "week") && (*period != "month") {
		fmt.Printf("Wrong period value. Should be 'week' or 'moth', got '%s'\n", *period)
		os.Exit(1)

	}

	// i := 0

	scanner := bufio.NewScanner(os.Stdin)
	// userAgents := make(map[string]int)
	// referers := make(map[string]int)
	// requests := make(map[string]int)
	// statuses := make(map[string]int)
	// logLines := make([]LogLine, 10000)

	for scanner.Scan() {
		// `Text` returns the current token, here the next line,
		// from the input.

		// fmt.Println(scanner.Text())
		// fmt.Println(scanner.Text())
		parsedLine := ConvertLogLineToMap(scanner.Text())
		// i += 1
		// Write out the uppercased line.
		// logLines = append(logLines, parsedLine)
		// if i > 40 {
		// 	break
		// }

		// }
		// for line := range logLines {

		if matchAllRequirements(parsedLine, *period) == false {
			continue
		}
		parsedLine["remote_addr"] = AnonymizeIp(parsedLine["remote_addr"])
		fmt.Println(ConvertMapToLogLine(parsedLine))
		// userAgents[parsedLine["http_user_agent"]] += 1
		// referers[parsedLine["http_referer"]] += 1
		// 	requests[logLines[line].request] += 1
		// 	statuses[logLines[line].status] += 1
		// 	fmt.Println(LogLineString(logLines[line]))
	}
	// sortByPopularity(referers)

	// // Check for errors during `Scan`. End of file is
	// // expected and not reported by `Scan` as an error.
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	// logLine := `52.87.65.11 - - [02/Nov/2018:06:55:13 +0000] "GET /feed.xml HTTP/1.1" 200 9810 "-" "curl"`
	// parsedLine := ConvertLogLineToMap(logLine)
	// fmt.Println(parsedLine)

}
