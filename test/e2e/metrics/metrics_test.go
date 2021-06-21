package metrics

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kyverno/kyverno/test/e2e"
	. "github.com/onsi/gomega"
)

func Test_MetricsServerAvailability(t *testing.T) {
	RegisterTestingT(t)
	if os.Getenv("E2E") == "" {
		t.Skip("Skipping E2E Test")
	}
	requestObj := e2e.APIRequest{
		URL:  "http://localhost:8000/metrics",
		Type: "GET",
	}
	response, err := e2e.CallAPI(requestObj)
	Expect(err).NotTo(HaveOccurred())
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	newStr := buf.String()

	layout := "2006-01-02 15:04:05 -0700 MST"
	timeInTimeFormat, err := time.Parse(layout, "2021-06-20 18:04:50 +0000 UTC")
	if err != nil {
		fmt.Println("error occurred: ", err)
	}
	processMetrics(newStr, "multi-tenancy", timeInTimeFormat)
	Expect(response.StatusCode).To(Equal(200))
}

func processMetrics(newStr, e2ePolicyName string, e2eTime time.Time) {
	fmt.Println("e2eTime: ", e2eTime)
	var action, policyName string
	var timeInTimeFormat time.Time
	var err error
	splitByNewLine := strings.Split(newStr, "\n")
	for _, lineSplitedByNewLine := range splitByNewLine {
		if strings.HasPrefix(lineSplitedByNewLine, "kyverno_policy_changes_info{") {
			// fmt.Println(lineSplitedByNewLine)
			splitByComma := strings.Split(lineSplitedByNewLine, ",")
			for _, lineSplitedByComma := range splitByComma {
				// fmt.Println(lineSplitedByComma)
				if strings.HasPrefix(lineSplitedByComma, "policy_change_type=") {
					// action = lineSplitedByComma
					splitByQuote := strings.Split(lineSplitedByComma, "\"")
					action = splitByQuote[1]
				}
				if strings.HasPrefix(lineSplitedByComma, "policy_name=") {
					splitByQuote := strings.Split(lineSplitedByComma, "\"")
					policyName = splitByQuote[1]
				}
				if strings.HasPrefix(lineSplitedByComma, "timestamp=") {
					splitByQuote := strings.Split(lineSplitedByComma, "\"")
					layout := "2006-01-02 15:04:05 -0700 MST"
					timeInTimeFormat, err = time.Parse(layout, splitByQuote[1])
					if err != nil {
						fmt.Println("error occurred: ", err)
					}
				}
			}
			if policyName == e2ePolicyName {
				diff := e2eTime.Sub(timeInTimeFormat)
				// fmt.Println(diff)
				if diff < 0 {
					// fmt.Println("-------less------")
					if action == "created" {
						fmt.Println("************policy created**************")
						break
					}
				}
			}
		}
	}
	fmt.Println("action: ", action)
	fmt.Println("policyName: ", policyName)
	fmt.Println("timeInTimeFormat: ", timeInTimeFormat)

}
