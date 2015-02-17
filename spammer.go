package urlspammer

import (
	"io"
	"net/http"
	"sync"
	"time"
)

// Spam runs a http.Get on each url from [urls] using [n] parallel threads.
// Runs [f] on each result, with the Get duration as argument.
// Be sure to close the urls channel at some point so that this method can return!
func SpamByThread(urls <-chan string, f func(io.Reader, time.Duration), nThreads int) {

	var waitGroup sync.WaitGroup

	for i := 0; i < nThreads; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for url := range urls {
				handleUrl(url, f)
			}
		}()
	}

	waitGroup.Wait()
}

// Spam runs a http.Get on each url from [urls] at the given rate per minute.
// Runs [f] on each result, with the Get duration as argument.
// Be sure to close the urls channel at some point so that this method can return!
func SpamByRate(urls <-chan string, f func(io.Reader, time.Duration), callPerMin int) {

	var callGroup sync.WaitGroup
	interval := time.Minute / time.Duration(callPerMin)

	for url := range urls {
		callGroup.Add(1)
		go func() {
			defer callGroup.Done()

			handleUrl(url, f)
		}()
		time.Sleep(interval)
	}

	callGroup.Wait()
}

func handleUrl(url string, f func(io.Reader, time.Duration)) {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		// HTTP get error.
		return
	}
	defer resp.Body.Close()

	f(resp.Body, time.Since(start))
}
