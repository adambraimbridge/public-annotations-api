package main

import (
	"net"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 32,
		Dial: (&net.Dialer{
			Timeout: 30 * time.Second,
		}).Dial,
	},
}
