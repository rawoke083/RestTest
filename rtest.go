package main

import (
	"bufio"
	"fmt"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"flag"
	"strconv"
	"strings"

	"bytes"
	"github.com/rawoke083/ColorPrint"
	"log"
	"os"

	"errors"
)

type arrayFlags []string

var Aliases arrayFlags
var Headers arrayFlags
var fileName string
var haveVariables = make(map[string]string)

// TestCase holds the HTTP method, URL to call,
// return code, what text to check on return, and pass/fail
type TestCase struct {
	HTTPMethod       string
	URL              string
	HTTPReturnCode   string
	ResponseTXTCheck string
	storeVar         string
	Headers          string
	Pass             bool
}

func getKeyType(keyname string, kn string, data interface{}) (bool, interface{}) {
	switch vv := data.(type) {

	case int, float64, string:
		if kn == keyname {
			return true, data
		}
	case json.Number:
		if kn == keyname {
			val, err := data.(json.Number).Int64()
			if err != nil {
				return true, nil
			}
			return true, val
		}
	case []interface{}:
		for _, u := range vv {
			return getKeyType(keyname, kn, u)
		}
	case map[string]interface{}:
		if kn == keyname {
			return true, data
		}

		for knn, ov := range vv {
			return getKeyType(keyname, knn, ov)
		}
	default:
		return false, 0
	}

	return false, 0
}

func getKey(keyname string, data map[string]interface{}) (bool, interface{}) {
	for k, v := range data {
		ok, x := getKeyType(keyname, k, v)
		if ok {
			return true, x
		}
	}

	return false, nil
}

func setKey(haveVariables map[string]string, keyname, value string) {
	if _, ok := haveVariables[keyname]; !ok {
		haveVariables[keyname] = value
	}
}

func setMultipleHeaders(req http.Request) {
	for _, val := range Headers {
		fields := strings.Split(val, ":")
		req.Header.Set(fields[0], fields[1])
	}
}

func setHeader(req http.Request, header string) {
	for i, k := range haveVariables {
		if strings.Contains(header, "%"+i+"%") {
			header = strings.Replace(header, "%"+i+"%", k, -1)
		}
	}
	fields := strings.Split(header, ":")

	req.Header.Set(fields[0], fields[1])
}

func (test *TestCase) runATest(mparams map[string]string) bool {

	for i, k := range mparams {
		test.URL = strings.Replace(test.URL, i, k, -1)
	}
	for i, k := range haveVariables {
		if strings.Contains(test.URL, "%"+i+"%") {
			test.URL = strings.Replace(test.URL, "%"+i+"%", k, -1)
		}
	}

	ColorPrint.ColWrite("\n\nTEST:"+ColorPrint.ToColor(test.HTTPMethod, ColorPrint.CL_LIGHT_CYAN)+" "+test.URL, ColorPrint.CL_WHITE)

	err := errors.New("")
	resp := new(http.Response)
	req := new(http.Request)

	// Parse URL and Query
	u, _ := url.Parse(test.URL)
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

	setMultipleHeaders(*req)
	if len(test.Headers) > 1 {
		h := strings.Split(test.Headers, ",")
		for _, v := range h {
			setHeader(*req, v)
		}
	}

	if resp, err = c.Do(req); err != nil {
		ColorPrint.ColWrite(fmt.Sprintf("\n\n=====>HTTP-ERROR:%s", test.URL), ColorPrint.CL_RED)
		return false
	}
	test.Pass = true

	http_code := strconv.Itoa(resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	s := string(body)
	defer resp.Body.Close()

	ctype := resp.Header.Get("Content-Type")
	switch ctype {
	case "application/javascript":

		data := json.NewDecoder(bytes.NewReader(body))
		data.UseNumber()

		var d map[string]interface{}
		err := data.Decode(&d)
		if err != nil {
			test.Pass = false
			ColorPrint.ColWrite("\n=>FAILED - Expected response application/javascript cannot be decoded", ColorPrint.CL_RED)
			break
		}
		if len(test.storeVar) > 1 {
			checkVar := strings.Split(test.storeVar, "=")

			ok, result := getKey(checkVar[0], d)
			if ok {
				var vvalue string
				switch vv := result.(type) {
				case float64:
					vvalue = fmt.Sprintf("%f", vv)
				case string:
					vvalue = vv
				case int64:
					vvalue = strconv.FormatInt(vv, 10)
				}
				if len(checkVar) > 1 {
					setKey(haveVariables, checkVar[1], vvalue)
				} else {
					setKey(haveVariables, checkVar[0], vvalue)
				}
			}
		}
	}

	if strings.TrimSpace(test.HTTPReturnCode) != http_code {

		test.Pass = false
		ColorPrint.ColWrite("\n=>FAILED - HttpCode Excepted |"+test.HTTPReturnCode+"| but got |"+http_code+"|", ColorPrint.CL_RED)
		return false
	}

	if len(test.ResponseTXTCheck) > 1 {

		if !strings.Contains(s, test.ResponseTXTCheck) {
			test.Pass = false

			ColorPrint.ColWrite("\n=>FAILED - ResponseText ("+test.ResponseTXTCheck+") not found. "+s, ColorPrint.CL_RED)
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
		if field_count > 4 {
			testCaseList[testCount].storeVar = fields[4]
		}
		if field_count > 5 {
			testCaseList[testCount].Headers = fields[5]
		}
		testCount++

	} //for scanner

	return testCaseList[0:testCount]
}

func printUsage() {
	ColorPrint.ColWrite("Usage:\n", ColorPrint.CL_WHITE)
	ColorPrint.ColWrite("\nrtest filename=<url-list-file> [-D search=replace..]\n\n", ColorPrint.CL_WHITE)
}

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
func init() {
	flag.Var(&Aliases, "D", "Replace text with some other text")
	flag.Var(&Headers, "H", "Request headers")
	flag.StringVar(&fileName, "filename", "", "File containing list of test rules")

	flag.Parse()
}

func main() {
	ColorPrint.ColWrite("\n###### REST Test ########\n", ColorPrint.CL_YELLOW)

	var mparams = make(map[string]string)

	for i := range Aliases {
		fields := strings.Split(Aliases[i], "=")
		mparams[fields[0]] = fields[1]
	}

	if len(fileName) < 1 {
		fmt.Println("\nError:No filename")
		printUsage()
		os.Exit(1)
	}

	os.Exit(runTestSuite(LoadTest(fileName), mparams))
}
