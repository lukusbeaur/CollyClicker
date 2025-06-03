// scrapeping/ping.go
package Util

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"
)

var client = http.Client{
	Timeout: 2 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 2 * time.Second,
		}).DialContext,
	},
}
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func Ping(url string) (int, error) {
	min := 5 * time.Second
	max := 15 * time.Second
	delay := min + time.Duration(rng.Int63n(int64(max-min)))
	fmt.Printf("Total Delay: %d\n", delay/1000000000)
	time.Sleep(delay)

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, nil
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}
