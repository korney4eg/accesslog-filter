package main

import (
	"bufio"
	// "errors"
	// "flag"
	// "encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	flags "github.com/jessevdk/go-flags"
)

// type parsedLogLine map[string]string
// type logLineMapped map[int]parsedLogLine
// type logLineDoubleMapped map[int]logLineMapped

type opts struct {
	Period        string `short:"p" long:"period" required:"true" choice:"any" choice:"day" choice:"month"`
	OutputDest    string `short:"o" long:"output-file-path"env:"OUTPUT_PATH" description:"path to save file(s). If not set output to stdout"`
	DivideByMonth bool   `short:"m" long:"divide-by-month" description:"Divide by month. Create file for each month. Example filename: ./08.reqs"`
	DivideByYear  bool   `short:"y" long:"divide-by-year" description:"Divide by year. Create folder for each year"`
}

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
	request := regexp.MustCompile(`GET (.+\.html|\/\d*\/\d*\/\d*\/.*\/) HTTP\/1\.[10]`)
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

func getOutputFilePath(outputDest string, date string, divideByMonth bool, divideByYear bool) (string, error) {
	outputPath := outputDest
	t, err := time.Parse("02/Jan/2006:15:04:05 -0700", date)
	if err != nil {
		return "", err
	}
	if !divideByMonth && !divideByYear {
		outputPath += "outputs"
	}
	if divideByYear {
		outputPath += fmt.Sprintf("/%d", t.Year())
	}
	if divideByMonth {
		outputPath += fmt.Sprintf("/%d", t.Month())
	}
	outputPath += ".reqs"
	return outputPath, nil
}

func main() {

	o := opts{}
	if _, err := flags.Parse(&o); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Period=%v\n", o.Period)
	fmt.Fprintf(os.Stderr, "OutputDest=%v\n", o.OutputDest)
	fmt.Fprintf(os.Stderr, "DivideByMonth=%v\n", o.DivideByMonth)
	fmt.Fprintf(os.Stderr, "DivideByYear=%v\n", o.DivideByYear)

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		parsedLine := ConvertLogLineToMap(scanner.Text())

		if matchAllRequirements(parsedLine, o.Period) == false {
			continue
		}
		parsedLine["remote_addr"] = AnonymizeIp(parsedLine["remote_addr"])
		if o.OutputDest == "" {
			fmt.Println(ConvertMapToLogLine(parsedLine))
		} else {
			filePath, err := getOutputFilePath(o.OutputDest, parsedLine["time_local"], o.DivideByMonth, o.DivideByYear)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			folder := path.Dir(filePath)
			err = os.MkdirAll(folder, 0755)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}

			defer f.Close()

			if _, err = f.WriteString(fmt.Sprintf("%v\n", parsedLine)); err != nil {
				panic(err)
			}

		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

}
