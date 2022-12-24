package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const keyServerAddr = "serverAddr"

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

func getPtList(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	tz := r.URL.Query().Get("tz")
	t1 := r.URL.Query().Get("t1")
	t2 := r.URL.Query().Get("t2")

	// call function to create json array response
	// use time module for date and time calculation
	getResult(period, tz, t1, t2)

	fmt.Printf("%s %s\n", r.Method, r.URL.JoinPath())
	io.WriteString(w, "period="+period+", tz="+tz+", t1="+t1+", t2="+t2+"\n")
}

func getResult(period string, tz string, t1 string, t2 string) {
	t1Time, err1 := handleTimeFormat(t1)
	if err1 != nil {
		return
	}
	t2Time, err2 := handleTimeFormat(t2)
	if err2 != nil {
		return
	}

	fmt.Printf("t1 = %s\n", t1Time)
	fmt.Printf("t2 = %s\n", t2Time)

	tTimeTmp, periodType, step, err := getPeriod(t1Time, period)
	if err != nil {
		fmt.Printf(err.Error())
	}

	fmt.Printf("tTmp = %s\n", tTimeTmp)
	fmt.Printf("period type = %c\n", periodType)
	fmt.Printf("step = %d\n", step)
	fmt.Printf("-----------------------------\n")

	timeStamps := []string{}

	for tTimeTmp.Before(t2Time) {
		switch periodType {
		case 'h':
			tTimeTmp = tTimeTmp.Add(time.Hour * time.Duration(step))
		case 'd':
			tTimeTmp = tTimeTmp.AddDate(0, 0, step)
		case 'o':
			tTimeTmp = tTimeTmp.AddDate(0, step, 0)
		case 'y':
			tTimeTmp = tTimeTmp.AddDate(step, 0, 0)
		}

		if tTimeTmp.Before(t2Time) {
			timeStamps = append(timeStamps, toTimestamp(tTimeTmp))
		}
	}

	timeStampsJson, _ := json.Marshal(timeStamps)
	fmt.Println(string(timeStampsJson))

}

func handleTimeFormat(t string) (time.Time, error) {
	timeRes := time.Time{} // Initialize empty Time

	//If timestamp is different size of standard format return error
	if len(t) != 16 {
		return time.Time{}, errors.New("Timestamp size must be 16")
	}
	//If T and Z are in correct position return error
	if t[8] != 'T' || t[15] != 'Z' {
		return time.Time{}, errors.New("T or Z is in wrong position")
	}
	//If timestamp contains - symbol return error
	if strings.Contains(t, "-") {
		return time.Time{}, errors.New("Timestamp can't contain '-'")
	}

	year, err1 := strconv.Atoi(t[:4])
	month, err2 := strconv.Atoi(t[4:6])
	day, err3 := strconv.Atoi(t[6:8])
	hour, err4 := strconv.Atoi(t[9:11])
	min, err5 := strconv.Atoi(t[11:13])
	sec, err6 := strconv.Atoi(t[13:15])

	// If there were conversion errors return error
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil {
		return time.Time{}, errors.New("Conversion error")
	}

	timeRes = timeRes.AddDate(year-1, month-1, day-1)
	timeRes = timeRes.Add(time.Hour*time.Duration(hour) +
		time.Minute*time.Duration(min) +
		time.Second*time.Duration(sec))

	// If converted time is different than timestamp given return error
	if int(timeRes.Year()) != year || int(timeRes.Month()) != month || timeRes.Day() != day ||
		int(timeRes.Hour()) != hour || int(timeRes.Minute()) != min || timeRes.Second() != sec {
		return time.Time{}, errors.New("Date or time is incorrect")
	}

	return timeRes, nil
}

func getPeriod(t1Time time.Time, per string) (time.Time, byte, int, error) {
	tTimeTmp := time.Time{}
	periodType := per[len(per)-1]
	step := 0
	err := errors.New("")

	switch periodType {
	case 'h':
		tTimeTmp = tTimeTmp.AddDate(t1Time.Year()-1, int(t1Time.Month())-1, t1Time.Day()-1)
		tTimeTmp = tTimeTmp.Add(time.Hour * time.Duration(t1Time.Hour()))
		step, err = strconv.Atoi(per[:len(per)-1])
	case 'd':
		tTimeTmp = tTimeTmp.AddDate(t1Time.Year()-1, int(t1Time.Month())-1, t1Time.Day()-1)
		step, err = strconv.Atoi(per[:len(per)-1])
	case 'o':
		if per[len(per)-2] == 'm' {
			tTimeTmp = tTimeTmp.AddDate(t1Time.Year()-1, int(t1Time.Month())-1, 0)
			step, err = strconv.Atoi(per[:len(per)-2])
		} else {
			fmt.Printf("Period %s not valid\n", per)
		}
	case 'y':
		tTimeTmp = tTimeTmp.AddDate(t1Time.Year()-1, 0, 0)
		step, err = strconv.Atoi(per[:len(per)-1])
	default:
		fmt.Printf("Period %s not valid\n", per)
	}

	return tTimeTmp, periodType, step, err
}

func toTimestamp(t time.Time) string {
	year := strconv.Itoa(t.Year())
	for len(year) < 4 {
		year = "0" + year
	}
	month := strconv.Itoa(int(t.Month()))
	if len(month) < 2 {
		month = "0" + month
	}
	day := strconv.Itoa(t.Day())
	if len(day) < 2 {
		day = "0" + day
	}
	hour := strconv.Itoa(t.Hour())
	if len(hour) < 2 {
		hour = "0" + hour
	}
	min := strconv.Itoa(t.Minute())
	if len(min) < 2 {
		min = "0" + min
	}
	sec := strconv.Itoa(t.Second())
	if len(sec) < 2 {
		sec = "0" + sec
	}

	return year + month + day + "T" + hour + min + sec + "Z"
}
