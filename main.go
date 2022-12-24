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

type JsonError struct {
	Status      string
	Description string
}

const keyServerAddr = "serverAddr"

func main() {
	portNum := ""

	//Check for console argument
	if len(os.Args) >= 2 {
		portNum = os.Args[1]
	} else {
		fmt.Printf("Port number must be given as command line argument.\n")
		os.Exit(1)
	}

	//Create server
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

// The HTTP service
func getPtList(w http.ResponseWriter, r *http.Request) {
	//Get url parameters
	period := r.URL.Query().Get("period")
	tz := r.URL.Query().Get("tz")
	t1 := r.URL.Query().Get("t1")
	t2 := r.URL.Query().Get("t2")

	requestStr := r.Method + " " + r.URL.JoinPath().String() + "\n"
	fmt.Printf(requestStr)

	jsonRes, err := getResult(period, tz, t1, t2)
	// Print response to both client and server
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, requestStr)
		io.WriteString(w, getJsonError(err))
		fmt.Println(getJsonError(err))
	} else {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, requestStr)
		io.WriteString(w, string(jsonRes))
		fmt.Println(string(jsonRes))
	}
}

// Get timestamps JSON results
func getResult(period string, tz string, t1 string, t2 string) ([]byte, error) {
	//Check if time parameters are valid and convert to Time type
	t1Time, err1 := handleTimeFormat(t1)
	if err1 != nil {
		return []byte{}, err1
	}
	t2Time, err2 := handleTimeFormat(t2)
	if err2 != nil {
		return []byte{}, err2
	}

	//Check if period parameter is valid, get step and create temporary Time
	tTimeTmp, periodType, step, err3 := getPeriod(t1Time, period)
	if err3 != nil {
		return []byte{}, err3
	}

	//Get array of timestamps between t1 and t2
	timeStamps := getTimestamps(tTimeTmp, t2Time, step, periodType)
	timeStampsJson, err4 := json.MarshalIndent(timeStamps, "", " ")

	return timeStampsJson, err4
}

// Check if URL time parameters are valid and convert to Time type
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

	// Convert input date/time to separate integers
	year, err1 := strconv.Atoi(t[:4])
	month, err2 := strconv.Atoi(t[4:6])
	day, err3 := strconv.Atoi(t[6:8])
	hour, err4 := strconv.Atoi(t[9:11])
	min, err5 := strconv.Atoi(t[11:13])
	sec, err6 := strconv.Atoi(t[13:15])

	// If there were characters other than numbers in the
	// conversions above returns error
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil {
		return time.Time{}, errors.New("Date or time has invalid number")
	}

	// Convert to Time type
	timeRes = timeRes.AddDate(year-1, month-1, day-1)
	timeRes = timeRes.Add(time.Hour*time.Duration(hour) +
		time.Minute*time.Duration(min) +
		time.Second*time.Duration(sec))

	// If converted time is different than timestamp given return error
	// e.g. 20210231T204603Z returns error since february doesn't have 31 days
	if int(timeRes.Year()) != year || int(timeRes.Month()) != month || timeRes.Day() != day ||
		int(timeRes.Hour()) != hour || int(timeRes.Minute()) != min || timeRes.Second() != sec {
		return time.Time{}, errors.New("Date or time given is not possible")
	}

	return timeRes, nil
}

// Check if period parameter is valid, get step and create temporary Time
func getPeriod(t1Time time.Time, per string) (time.Time, byte, int, error) {
	tTimeTmp := time.Time{}
	periodType := per[len(per)-1]
	step := 0
	err := errors.New("")

	switch periodType {
	case 'h': // Hour
		tTimeTmp = tTimeTmp.AddDate(t1Time.Year()-1, int(t1Time.Month())-1, t1Time.Day()-1)
		tTimeTmp = tTimeTmp.Add(time.Hour * time.Duration(t1Time.Hour()))
		step, err = strconv.Atoi(per[:len(per)-1])
	case 'd': // Day
		tTimeTmp = tTimeTmp.AddDate(t1Time.Year()-1, int(t1Time.Month())-1, t1Time.Day()-1)
		step, err = strconv.Atoi(per[:len(per)-1])
	case 'o': // Month
		if per[len(per)-2] == 'm' {
			tTimeTmp = tTimeTmp.AddDate(t1Time.Year()-1, int(t1Time.Month())-1, 0)
			step, err = strconv.Atoi(per[:len(per)-2])
		} else {
			err = errors.New("Period is not valid")
		}
	case 'y': // Year
		tTimeTmp = tTimeTmp.AddDate(t1Time.Year()-1, 0, 0)
		step, err = strconv.Atoi(per[:len(per)-1])
	default:
		err = errors.New("Period is not valid")
	}

	if err != nil {
		err = errors.New("Period is not valid")
	}

	return tTimeTmp, periodType, step, err
}

// Get array of timestamps between t1 and t2
func getTimestamps(t1 time.Time, t2 time.Time, step int, periodType byte) []string {
	timeStamps := []string{}

	// Loop until t1 passes t2.
	// Each loop adds step amount of period type to Time.
	for t1.Before(t2) {
		switch periodType {
		case 'h':
			t1 = t1.Add(time.Hour * time.Duration(step))
		case 'd':
			t1 = t1.AddDate(0, 0, step)
		case 'o':
			t1 = t1.AddDate(0, step, 0)
		case 'y':
			t1 = t1.AddDate(step, 0, 0)
		}

		if t1.Before(t2) {
			timeStamps = append(timeStamps, toTimestamp(t1))
		}
	}

	return timeStamps
}

// Converts Time to timestamp
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

// Convert error to JSON
func getJsonError(err error) string {
	var jError JsonError
	jError.Status = "Error"
	jError.Description = err.Error()

	errStr, _ := json.MarshalIndent(jError, "", " ")

	return string(errStr)
}
