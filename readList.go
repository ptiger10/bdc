package bdc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

// convert JSON values into a URL string
func encodeReadListData(c *Client, start int, max int, filters, sorts []map[string]interface{}) io.Reader {
	// common pagination operators
	values := map[string]interface{}{"start": start * max, "max": max}
	// query specific filters, if any
	values["filters"] = filters
	values["sort"] = sorts
	// encode payload as URL
	data := url.Values{}
	jsonValues, _ := json.Marshal(values)
	data.Set("data", string(jsonValues))
	data.Set("sessionId", c.sessionID)
	data.Set("devKey", c.devKey)

	body := strings.NewReader(data.Encode())
	return body
}

// Get up to pageMax records from an endpoint starting at record number "start" with optional filters
func (c *Client) getPage(start int, max int, endpoint string, filters, sorts []map[string]interface{}) resultError {
	body := encodeReadListData(c, start, max, filters, sorts)
	resp, err := makeRequest(endpoint, body)
	if err != nil {
		return resultError{err: err}
	}
	return resultError{result: resp}
}

// goroutine to check whether an item exists at a specific location,
// for use with countPages
func (c *Client) countRoutine(pages <-chan int, result chan<- int, endpoint string, filters, sorts []map[string]interface{}) {
	for {
		p := <-pages
		position := p * pageMax
		resp := c.getPage(position, 1, endpoint, filters, sorts)
		var goodResponse baseResponse
		json.Unmarshal(resp.result, &goodResponse)
		if len(goodResponse.Data) == 0 {
			result <- p
			return
		}
	}
}

// count the max number of pages to fetch by inspecting whether each subsequent page returns a value
func (c *Client) countPages(endpoint string, filters, sorts []map[string]interface{}) int {
	countRoutines := workersMax
	// var countRoutines int
	// if workersMax > 1 {
	// 	countRoutines = workersMax - 1 // sometimes a trailing goroutine exceeds the allowable concurrent threads
	// } else {
	// 	countRoutines = 1
	// }
	pages := make(chan int)
	result := make(chan int)
	for i := 0; i < countRoutines; i++ {
		go c.countRoutine(pages, result, endpoint, filters, sorts)
	}

	go func() {
		for i := 1; ; i++ {
			pages <- i
		}
	}()

	for {
		select {
		case r := <-result:
			return r
		}
	}

}

// goroutine for fetching a page of results at a specified List endpoint;
// for use with getAll()
func (c *Client) fetchRoutine(pages <-chan int, results chan<- resultError, endpoint string, filters, sorts []map[string]interface{}) {
	for {
		p, ok := <-pages
		if !ok {
			wg.Done()
			return
		}
		ch := make(chan resultError)
		go func() {
			ch <- c.getPage(p, pageMax, endpoint, filters, sorts)
		}()
		select {
		case <-time.After(8 * time.Second):
			results <- resultError{err: fmt.Errorf("Timed out request; try again later")}
		case r := <-ch:
			results <- r
		}
	}
}

// getAll is called by a specific resource, eg invoiceResource,
// deploys asynchronous countRoutines to count the number of pages available,
// then deploys asynchronous fetchRoutines to fetch all the items from all the available pages
func (c *Client) getAll(suffix string, parameters []*Parameters) (result []resultError) {
	endpoint := "List/" + suffix
	filters, sorts := encodeParameters(parameters)
	numPages := c.countPages(endpoint, filters, sorts)
	time.Sleep(500 * time.Millisecond) // sometimes a trailing goroutine exceeds the allowable concurrent threads

	pages := make(chan int)
	results := make(chan resultError, numPages)
	wg.Add(workersMax)
	for i := 0; i < workersMax; i++ {
		go c.fetchRoutine(pages, results, endpoint, filters, sorts)
	}
	for i := 0; i < numPages; i++ {
		pages <- i
	}
	close(pages)
	wg.Wait()
	close(results)

	for resp := range results {
		result = append(result, resp)
	}
	if len(result) > 0 {
		sort.Slice(result, func(i, j int) bool {
			return result[i].page < result[j].page
		})
	}
	return result

}
