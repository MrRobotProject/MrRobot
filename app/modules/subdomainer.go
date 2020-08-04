package modules

// === IMPORTS ===

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/cosasdepuma/elliot/app/utils"
)

// === PUBLIC METHODS ===

// Subdomains is a concurrent method to obtain the sub-domains associated to a domain using different services.
func Subdomains(domain string, output *chan []string) {
	availableMethods := [](func(string) ([]string, error)){
		subdomainsInHackerTarget, subdomainsInThreatCrowd,
	}
	// Concurrency
	wg := sync.WaitGroup{}
	wg.Add(len(availableMethods))
	channel := make(chan []string, len(availableMethods))
	defer close(channel)

	// Initialize the concurrency
	for _, method := range availableMethods {
		go concurrentSubdomainer(method, domain, &wg, &channel)
	}

	// Retrieve the results
	subdomains, i := make([]string, 0), 0
	for i < len(availableMethods) {
		i++
		subdomains = append(subdomains, <-channel...)
	}

	// Filter the duplicates
	if len(subdomains) == 0 {
		*output <- nil
		return
	}
	subdomains = utils.FilterDuplicates(subdomains)
	*output <- utils.FilterDuplicates(subdomains)
}

// === PRIVATE METHODS ===

// ==== Subconcurrency Method ====

func concurrentSubdomainer(method func(string) ([]string, error), domain string, wg *sync.WaitGroup, channel *chan []string) {
	defer wg.Done()
	result, err := method(domain)
	if err != nil {
		*channel <- nil
	} else {
		*channel <- result
	}
}

// ==== Subdomain Methods ====

func subdomainsInThreatCrowd(domain string) ([]string, error) {
	// Compose the URL
	url := fmt.Sprintf("https://www.threatcrowd.org/searchApi/v2/domain/report/?domain=%s", domain)
	// Request the data
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return nil, errors.New("ThreatCrowd is not available")
	}
	// Grab the content
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("ThreatCrowd does not respond correctly")
	}
	// Parse the JSON
	subdomains := struct {
		Results []string `json:"subdomains"`
	}{}
	err = json.Unmarshal([]byte(body), &subdomains)
	if err != nil {
		return nil, errors.New("Bad JSON format using ThreatCrowd")
	}
	// Return the JSON
	return subdomains.Results, nil
}

func subdomainsInHackerTarget(domain string) ([]string, error) {
	// Compose the URL
	url := fmt.Sprintf("https://api.hackertarget.com/hostsearch/?q=%s", domain)
	// Request the data
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return nil, errors.New("HackerTarget is not available")
	}
	// Grab the content
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("HackerTarget does not respond correctly")
	}
	// Parse the Response
	subdomains := make([]string, 0)
	sc := bufio.NewScanner(bytes.NewReader(body))
	for sc.Scan() {
		splitter := strings.SplitN(sc.Text(), ",", 2)
		subdomains = append(subdomains, splitter[0])
	}
	return subdomains, nil
}
