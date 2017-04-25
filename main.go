package main

import (
	"log"
	"os"
	"bufio"
	"sync"
	"fmt"
	"strings"
	"regexp"
	"net/url"
	"io/ioutil"
	"net/http"
)

func main() {
	var searchReg = regexp.MustCompile(`http(s)?://[a-z0-9-]+(.[a-z0-9-]+)*(:[0-9]+)?(/.*)?`)

	inChan := make(chan string, 100)
	countChan := make(chan int)

	var inWg sync.WaitGroup
	inWg.Add(1)
	go processInput(inChan, searchReg, &inWg)

	var processWg sync.WaitGroup
	for i := 0; i < 5; i++ {
		processWg.Add(1)
		go processRead(inChan, countChan, &processWg)
	}

	inWg.Add(1)
	go func (wg *sync.WaitGroup) {
		defer wg.Done()
		counter := 0
		for c := range countChan {
			counter += c
		}
		fmt.Println("Total: ", counter)
	}(&inWg)

	processWg.Wait()
	close(countChan)

	inWg.Wait()
}

func processInput(in chan string, re *regexp.Regexp, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		close(in)
	}()

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
			in <- string(str)
		}
	}
}

func processRead(in chan string, out chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for str := range in {
		count, err := sendRequest(str)
		if err != nil {
			log.Println(err)
			continue
		}
		out <- count
	}
}

func sendRequest(url string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	str := string(body)

	count := strings.Count(str, "Go")
	fmt.Println("Count for", url, count)

	return count, nil
}
