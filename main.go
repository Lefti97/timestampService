package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

const keyServerAddr = "serverAddr"

func getPtList(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	tz := r.URL.Query().Get("tz")
	t1 := r.URL.Query().Get("t1")
	t2 := r.URL.Query().Get("t2")

	// call function to create json array response

	fmt.Printf("%s %s\n", r.Method, r.URL.JoinPath())
	io.WriteString(w, "period="+period+", tz="+tz+", t1="+t1+", t2="+t2+"\n")
}

func main() {
	portNum := ""

	if len(os.Args) >= 2 {
		portNum = os.Args[1]
	} else {
		fmt.Printf("Port number must be given as command line argument.\n")
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ptlist", getPtList)

	ctx := context.Background()
	server := &http.Server{
		Addr:    ":" + portNum,
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
			return ctx
		},
	}

	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error listening for server: %s\n", err)
	}
}
