package main

import (
	"testing"
)

func assertMaps(t *testing.T, got, want map[string]string) {
	t.Helper()
	for k, _ := range want {
		if got[k] != want[k] {
			t.Errorf("\nexpected value: %s\n     got value: %s", want[k], got[k])
		}
	}

}

func TestConvertLogLine(t *testing.T) {
	logLine := ""
	t.Run("Parsing normal log entry", func(t *testing.T) {
		logLine = `52.87.65.11 - - [02/Nov/2018:06:55:13 +0000] "GET /feed.xml HTTP/1.1" 200 9810 "-" "curl"`
		expectedResult := map[string]string{
			"remote_addr":     "52.87.65.11",
			"remote_user":     "",
			"time_local":      "02/Nov/2018:06:55:13 +0000",
			"request":         "GET /feed.xml HTTP/1.1",
			"status":          "200",
			"body_bytes_sent": "9810",
			"http_referer":    "-",
			"http_user_agent": "curl",
		}
		parsedLine := ConvertLogLineToMap(logLine)
		assertMaps(t, parsedLine, expectedResult)
	})

	t.Run("Parsing normal log entry with empty user agent", func(t *testing.T) {
		logLine = `190.128.131.6 - - [03/Nov/2018:23:25:08 +0000] "" 400 0 "-" "-"`
		expectedResult := map[string]string{
			"remote_addr":     "190.128.131.6",
			"remote_user":     "",
			"time_local":      "03/Nov/2018:23:25:08 +0000",
			"request":         "",
			"status":          "400",
			"body_bytes_sent": "0",
			"http_referer":    "-",
			"http_user_agent": "-",
		}
		parsedLine := ConvertLogLineToMap(logLine)
		assertMaps(t, parsedLine, expectedResult)
	})

	t.Run("Parsing strange line that should fail", func(t *testing.T) {
		logLine = `190.XXXX.XXXXX.XXX - - [03/Nov/2018:23:25:08 +0000] "" 400 0 "-" "-"`
		parsedLine := ConvertLogLineToMap(logLine)
		assertMaps(t, parsedLine, nil)
	})
}

func TestAnonymizeIp(t *testing.T) {
	cases := map[string]string{
		"5.255.250.183":   "5.255.250.0",
		"141.8.144.9":     "141.8.144.0",
		"178.154.244.157": "178.154.244.0",
		"40.77.167.89":    "40.77.167.0",
		"66.249.69.70":    "66.249.69.0",
	}
	for testCase, expectedResult := range cases {
		actualResult := AnonymizeIp(testCase)
		if actualResult != expectedResult {
			t.Errorf("result for line: '%s'\n returned wrong value: %s, expected: %s", testCase, actualResult, expectedResult)
		}
	}
}

func TestLogLineOK(t *testing.T) {
	cases := map[string]bool{
		`159.203.112.40 - - [05/Jan/2019:22:35:12 +0000] "GET / HTTP/1.0" 200 13444 "-" "Mozilla/5.0 (compatible; NetcraftSurveyAgent/1.0; +info@netcraft.com)"`:                                                                                                        false,
		`213.138.93.47 - - [05/Jan/2019:23:36:41 +0000] "GET /2018/08/25/aws-certification-preparation.html HTTP/1.1" 200 7779 "https://www.google.com/" "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:64.0) Gecko/20100101 Firefox/64.0"`:                               true,
		`213.138.93.47 - - [05/Jan/2019:23:36:42 +0000] "GET /?facebook HTTP/1.1" 200 4402 "https://makvaz.com/2018/08/25/aws-certification-preparation.html" "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:64.0) Gecko/20100101 Firefox/64.0"`:                          true,
		`199.16.157.180 - - [02/Jan/2019:11:54:10 +0000] "GET /2018/09/25/effective-devops.html HTTP/1.1" 200 10399 "-" "Twitterbot/1.0"`:                                                                                                                               false,
		`199.16.157.183 - - [02/Jan/2019:11:54:10 +0000] "GET /assets/img/header-pic.jpeg HTTP/1.1" 200 696591 "-" "Twitterbot/1.0"`:                                                                                                                                    false,
		`54.36.148.130 - - [02/Jan/2019:12:06:17 +0000] "GET /blog/page3/ HTTP/1.1" 200 3586 "-" "Mozilla/5.0 (compatible; AhrefsBot/6.1; +http://ahrefs.com/robot/)"`:                                                                                                  false,
		`125.212.217.215 - - [02/Jan/2019:16:03:08 +0000] "GET /robots.txt HTTP/1.1" 200 40 "-" "-"`:                                                                                                                                                                    false,
		`40.77.167.146 - - [02/Jan/2019:16:03:59 +0000] "GET /about/ HTTP/1.1" 200 3193 "-" "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)"`:                                                                                                  false,
		`89.64.54.234 - - [25/Mar/2019:19:07:01 +0000] "GET /2019/03/07/ideal-cicd-on-practice1/ HTTP/1.1" 200 11379 "https://makvaz.com/" "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 Safari/537.36"`: true,
	}
	for testCase, expectedResult := range cases {
		parsedLine := ConvertLogLineToMap(testCase)
		actualResult := matchAllRequirements(parsedLine, "any")
		if actualResult != expectedResult {
			t.Errorf("result for line: '%s'\n returned wrong value: %t, expected: %t", testCase, actualResult, expectedResult)

		}
	}
}

func TestConvertMapToLogLine(t *testing.T) {
	// logLine := ""
	t.Run("Parsing normal map entry", func(t *testing.T) {
		expectedLogLine := `52.87.65.11 - - [02/Nov/2018:06:55:13 +0000] "GET /feed.xml HTTP/1.1" 200 9810 "-" "curl"`
		parsedMap := map[string]string{
			"remote_addr":     "52.87.65.11",
			"remote_user":     "",
			"time_local":      "02/Nov/2018:06:55:13 +0000",
			"request":         "GET /feed.xml HTTP/1.1",
			"status":          "200",
			"body_bytes_sent": "9810",
			"http_referer":    "-",
			"http_user_agent": "curl",
		}
		generatedLogLine := ConvertMapToLogLine(parsedMap)
		if expectedLogLine != generatedLogLine {
			t.Errorf("Log line did not match!.\nExpected: '%s'\n      Got:'%s'", expectedLogLine, generatedLogLine)
		}
	})
	t.Run("Parsing normal map entry with anonimized ip", func(t *testing.T) {
		expectedLogLine := `190.128.131.0 - - [03/Nov/2018:23:25:08 +0000] "" 400 0 "-" "-"`
		parsedMap := map[string]string{
			"remote_addr":     "190.128.131.6",
			"remote_user":     "",
			"time_local":      "03/Nov/2018:23:25:08 +0000",
			"request":         "",
			"status":          "400",
			"body_bytes_sent": "0",
			"http_referer":    "-",
			"http_user_agent": "-",
		}
		parsedMap["remote_addr"] = AnonymizeIp(parsedMap["remote_addr"])
		generatedLogLine := ConvertMapToLogLine(parsedMap)
		if expectedLogLine != generatedLogLine {
			t.Errorf("Log line did not match!.\nExpected: '%s'\n      Got:'%s'", expectedLogLine, generatedLogLine)
		}
	})
}

// func TestLogLineString(t *testing.T) {
// 	logLine := `52.87.65.11 - - [02/Nov/2018:06:55:13 +0000] "GET /feed.xml HTTP/1.1" 200 9810 "-" "curl"`
// 	parsedStructure := LogLine{
// 		remote_addr:     "52.87.65.11",
// 		remote_user:     "",
// 		time_local:      "02/Nov/2018:06:55:13 +0000",
// 		request:         "GET /feed.xml HTTP/1.1",
// 		status:          "200",
// 		body_bytes_sent: "9810",
// 		http_referer:    "-",
// 		http_user_agent: "curl",
// 	}
// 	outputLine := LogLineString(parsedStructure)
// 	if outputLine != logLine {
// 		t.Errorf("expected:\n%v\n got:\n%v", logLine, outputLine)
// 	}

// }
