package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	portNum := ""
	period := ""
	tz := ""
	t1 := ""
	t2 := ""

	//Check for console argument
	if len(os.Args) >= 6 {
		portNum = os.Args[1]
		period = os.Args[2]
		tz = os.Args[3]
		t1 = os.Args[4]
		t2 = os.Args[5]
	} else {
		fmt.Printf("Needed parameters: <port> <period> <tz> <t1> <t2>\n")
		os.Exit(1)
	}

	//Set URL
	requestURL := fmt.Sprintf("http://localhost:%s/ptlist?period=%s&tz=%s&t1=%s&t2=%s", portNum, period, tz, t1, t2)
	res, err := http.Get(requestURL) //Call URL
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("URL: %s\n", requestURL)
	b, err := io.ReadAll(res.Body)
	fmt.Printf(string(b))
}
