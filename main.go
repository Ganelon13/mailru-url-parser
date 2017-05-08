package main

import (
	"log"
	"os"
	"bufio"
	"fmt"
	"strings"
	"regexp"
	"net/url"
	"io/ioutil"
	"net/http"
	"sync"
)

// echo -e 'https://golang.org\nhttps://golang.org\nhttp://mail.ru\nhttp://4gophers.ru\nhttps://golang.org\nhttp://golangshow.com\nhttps://golang.org\nhttps://golang.org\nhttps://golang.org' | go run main.go

func main() {
	var searchReg = regexp.MustCompile(`http(s)?://[a-z0-9-]+(.[a-z0-9-]+)*(:[0-9]+)?(/.*)?`)

	buf := make(chan struct{}, 5)
	urlChan := make(chan string)
	countChan := make(chan int)

	go processInput(urlChan, searchReg)
	go processRead(urlChan, countChan, buf)
	processStats(countChan)
}

func processInput(out chan <- string, re *regexp.Regexp) {
	defer close(out)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		record := scanner.Text()
		if record == `:q` {
			return
		}
		lines := strings.Split(record, `\n`)
		for _, line := range lines {
			str := re.FindString(line)
			if _, err := url.ParseRequestURI(str); err != nil {
				log.Println(err)
				continue
			}
			out <- str
		}
	}
}

func processRead(in <-chan string, out chan<- int, buf chan struct{}) {
	var wg sync.WaitGroup

	for str := range in {
		wg.Add(1)
		buf <- struct{}{}

		go func(str string) {
			defer wg.Done()
			count, err := sendRequest(str)
			if err != nil {
				log.Println(err)
				return
			}
			out <- count
			<-buf
		}(str)
	}

	wg.Wait()
	close(out)
}

func processStats(in <-chan int) {
	counter := 0
	for c := range in {
		counter += c
	}
	fmt.Println("Total:", counter)
}

func sendRequest(url string) (int, error) {
	resp, err := http.Get(url)
	// https://habrahabr.ru/company/mailru/blog/314804/#36
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return 0, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	str := string(body)

	count := strings.Count(str, "Go")
	fmt.Println("Count for", url, count)

	return count, nil
}
