package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

func getPageContent(url string) string {
	if url == "" {
		log.Fatalf("Unexpected argument '%s' value.\n", "url")
	}

	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()
	contentBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	return string(contentBytes)
}

func getWordCount(searchingWord string, text string) int {
	if searchingWord == "" {
		log.Fatalf("Unexpected argument '%s' value.\n", "searchingWord")
	}

	if text == "" {
		return 0
	}

	return strings.Count(text, searchingWord)
}

func getUrlsFromStdin() []string {
	result := []string{}

	stdinScanner := bufio.NewScanner(os.Stdin)
	for stdinScanner.Scan() {
		scanErr := stdinScanner.Err()
		if scanErr != nil {
			log.Fatal(scanErr)
		}

		scannedUrl := stdinScanner.Text()
		result = append(result, scannedUrl)
	}

	return result
}

func main() {
	const maxConcurrencyLevel int = 5
	const searchingWord string = "Go"

	// get input urls
	urls := getUrlsFromStdin()
	concurrencyLevel := maxConcurrencyLevel
	if len(urls) < maxConcurrencyLevel {
		concurrencyLevel = len(urls)
	}
	fmt.Printf("App started with concurrency level of %d.\n", concurrencyLevel)

	// do concurrency work
	works := make(chan string)
	waitGroup := new(sync.WaitGroup)
	var totalCount uint64 = 0
	for goroutineIndex := 0; goroutineIndex < concurrencyLevel; goroutineIndex++ {
		waitGroup.Add(1)
		go func() {
			for url := range works {
				pageContent := getPageContent(url)
				wordCountOnPage := getWordCount(searchingWord, pageContent)
				fmt.Printf("Count for %s: %d\n", url, wordCountOnPage)
				atomic.AddUint64(&totalCount, uint64(wordCountOnPage))
			}
			waitGroup.Done()
		}()
	}

	for _, url := range urls {
		works <- url
	}
	close(works)

	// wait all jobs done
	waitGroup.Wait()
	fmt.Printf("Total: %d\n", totalCount)
}
