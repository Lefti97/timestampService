package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

const keyServerAddr = "serverAddr"

func getPtList(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()

	//hasPeriod := r.URL.Query().Has("period")
	period := r.URL.Query().Get("period")
	//hasTz := r.URL.Query().Has("tz")
	tz := r.URL.Query().Get("tz")
	//hasT1 := r.URL.Query().Has("t1")
	t1 := r.URL.Query().Get("t1")
	//hasT2 := r.URL.Query().Has("t2")
	t2 := r.URL.Query().Get("t2")

	fmt.Printf("got / request. period=%s, tz=%s, t1=%s, t2=%s\n",
		period, tz, t1, t2)
	io.WriteString(w, "period="+period+", tz="+tz+", t1="+t1+", t2="+t2+"\n")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ptlist", getPtList)

	ctx := context.Background()
	server := &http.Server{
		Addr:    ":3333",
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
