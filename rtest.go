package main

import (
	"bufio"
	"fmt"

	"io/ioutil"
	"net/http"
	"net/url"

	"strconv"
	"strings"

	"github.com/rawoke083/ColorPrint"
	"log"
	"os"

	"errors"
)

// TestCase holds the HTTP method, URL to call,
// return code, what text to check on return, and pass/fail
type TestCase struct {
	HTTPMethod       string
	URL              string
	HTTPReturnCode   string
	ResponseTXTCheck string
	Pass             bool
}

func (test *TestCase) runATest(mparams map[string]string) bool {

	for i, k := range mparams {
		test.URL = strings.Replace(test.URL, i, k, -1)
	}

	ColorPrint.ColWrite("\n\nTEST:"+ColorPrint.ToColor(test.HTTPMethod, ColorPrint.CL_LIGHT_CYAN)+" "+test.URL, ColorPrint.CL_WHITE)

	err := errors.New("")
	resp := new(http.Response)
	req := new(http.Request)

	// Parse URL and Query
	u, _ := url.Parse(test.URL)
	//q, _ := url.ParseQuery(u.RawQuery)
	c := &http.Client{}

	switch test.HTTPMethod {
	case "GET":
		req, err = http.NewRequest("GET", test.URL, nil)
		if err != nil {
			return false
		}
	case "POST":
		req, err = http.NewRequest("POST", test.URL, strings.NewReader(u.RawQuery))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		if err != nil {
			return false
		}
	case "PATCH":
		req, err = http.NewRequest("PATCH", test.URL, strings.NewReader(u.RawQuery))
		if err != nil {
			return false
		}
	case "DELETE":
		req, err = http.NewRequest("DELETE", test.URL, strings.NewReader(u.RawQuery))
		if err != nil {
			return false
		}
	}

	if resp, err = c.Do(req); err != nil {
		// handle error
		ColorPrint.ColWrite(fmt.Sprintf("\n\n=====>HTTP-ERROR:%s", test.URL), ColorPrint.CL_RED)
		return false
	}
	resp.Body.Close()

	test.Pass = true

	http_code := strconv.Itoa(resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	s := string(body)

	if strings.TrimSpace(test.HTTPReturnCode) != http_code {

		test.Pass = false
		ColorPrint.ColWrite("\n=>FAILED  - HttpCode Excepted |"+test.HTTPReturnCode+"| but got |"+http_code+"|", ColorPrint.CL_RED)
		return false
	}

	if len(test.ResponseTXTCheck) > 1 {

		if !strings.Contains(s, test.ResponseTXTCheck) {
			test.Pass = false

			ColorPrint.ColWrite("\n=>FAILED  - ResponseText ("+test.ResponseTXTCheck+") not found. "+s, ColorPrint.CL_RED)
			return false

		}

	}

	return true
}

func runTestSuite(testCases []TestCase, mparams map[string]string) int {

	var testRunCount int
	var testOKCount int

	for _, test := range testCases {

		testRunCount++
		if test.runATest(mparams) {
			testOKCount++
		}

	} //end for

	s_total_test := fmt.Sprintf("\n\nTotal Test %d", testRunCount)
	s_total_test_ok := fmt.Sprintf("\nTest(s) OK %d", testOKCount)
	s_total_test_failed := fmt.Sprintf("\nTest(s) Failed %d\n", int(testRunCount-testOKCount))

	ColorPrint.ColWrite(s_total_test, ColorPrint.CL_LIGHT_BLUE)
	ColorPrint.ColWrite(s_total_test_ok, ColorPrint.CL_LIGHT_GREEN)
	ColorPrint.ColWrite(s_total_test_failed, ColorPrint.CL_RED)

	return (testRunCount - testOKCount)
}

// LoadTest loads test data from test specification files
func LoadTest(filename string) []TestCase {

	//	valid_http_methods := []string{"GET", "POST", "PUT", "DELETE"}

	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	//max of 64 test for now...
	testCaseList := [64]TestCase{}
	testCount := 0

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {

		if strings.HasPrefix(scanner.Text(), "#") {
			continue
		}

		fields := strings.Split(scanner.Text(), "|")
		field_count := len(fields)

		if field_count < 2 {
			continue
		}

		testCaseList[testCount].HTTPMethod = fields[0]
		testCaseList[testCount].URL = fields[1]
		testCaseList[testCount].HTTPReturnCode = fields[2]
		if field_count > 3 {

			testCaseList[testCount].ResponseTXTCheck = fields[3]
		}
		testCount++

	} //for scanner

	return testCaseList[0:testCount]
}

func printUsage() {
	ColorPrint.ColWrite("Usage:\n", ColorPrint.CL_WHITE)
	ColorPrint.ColWrite("\nrtest filename=<url-list-file> [-DStringToReplace=NewString] [-DMoreStringToReplace=MoreNewString]\n\n", ColorPrint.CL_WHITE)
}

func main() {
	ColorPrint.ColWrite("\n###### REST Test ########\n", ColorPrint.CL_YELLOW)

	var mparams = make(map[string]string)
	var fileName string

	for i := range os.Args {
		if strings.Contains(os.Args[i], "-D") {
			fields := strings.Split(strings.Replace(os.Args[i], "-D", "", -1), "=")
			mparams[fields[0]] = fields[1]
		}

		if strings.Contains(os.Args[i], "filename=") {
			fields := strings.Split(os.Args[i], "=")
			fileName = fields[1]
		}
	}

	if len(fileName) < 1 {
		fmt.Println("\nError:No filename")
		printUsage()
		os.Exit(1)
	}

	os.Exit(runTestSuite(LoadTest(fileName), mparams))
}
