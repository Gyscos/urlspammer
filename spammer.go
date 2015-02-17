package urlspammer

import (
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type Query struct {
	Url  string
	Data string
}

func WrapUrls(urls <-chan string) <-chan Query {
	queries := make(chan Query, 20)
	go func() {
		for url := range urls {
			queries <- Query{Url: url}
		}
		close(queries)
	}()
	return queries
}

type HandlerFunc func(Query, []byte, time.Duration)

// Spam runs a http.Get on each url from [urls] using [n] parallel threads.
// Runs [f] on each result, with the Get duration as argument.
// Be sure to close the urls channel at some point so that this method can return!
func SpamByThread(nThreads int, queries <-chan Query, f HandlerFunc) {

	var waitGroup sync.WaitGroup

	for i := 0; i < nThreads; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for query := range queries {
				handleUrl(query, f)
			}
		}()
	}

	waitGroup.Wait()
}

// Spam runs a http.Get on each url from [urls] at the given rate per minute.
// Runs [f] on each result, with the Get duration as argument.
// Be sure to close the urls channel at some point so that this method can return!
func SpamByRate(callPerMin int, queries <-chan Query, f HandlerFunc) {

	var callGroup sync.WaitGroup
	interval := time.Minute / time.Duration(callPerMin)

	for query := range queries {
		callGroup.Add(1)
		go func() {
			defer callGroup.Done()
			handleUrl(query, f)
		}()
		time.Sleep(interval)
	}

	callGroup.Wait()
}

func handleUrl(query Query, f HandlerFunc) {
	start := time.Now()
	resp, err := http.Get(query.Url)
	if err != nil {
		// HTTP get error.
		log.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// HTTP read error?
		log.Println("Error:", err)
		return
	}

	f(query, body, time.Since(start))
}
