package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	stathat "github.com/stathat/go"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func KelvinToFahr(kelvin float64) float64 {
	return (kelvin-273.15)*1.8 + 32
}

type TickList struct {
	Message string
	Code    string `json:"cod"`
	List    []*Tick
}

type Tick struct {
	Kelvin float64 `json:"temp"`
	Fahr   float64
	Dt     int64
	Date   time.Time
}

func NewTickList(stationID string) (list *TickList, err error) {
	url := "http://api.openweathermap.org/data/2.1/history/station/" +
		stationID + "?type=tick"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buf, &list)
	if err != nil {
		return nil, err
	}
	list.init()
	return list, nil
}

func (tick *Tick) init() {
	tick.Fahr = KelvinToFahr(tick.Kelvin)
	tick.Date = time.Unix(tick.Dt, 0)
}

func (list *TickList) init() {
	for _, tick := range list.List {
		tick.init()
	}
}

func (list *TickList) Len() int {
	return len(list.List)
}

func (list *TickList) Tick(i int) *Tick {
	return list.List[i]
}

func (tick *Tick) Print() {
	fmt.Printf("% 3.1f %v\n", tick.Fahr, tick.Date)
}

func (list *TickList) Print() {
	if len(list.List) < 0 {
		fmt.Println("no data from station!")
		return
	}
	fmt.Printf("%5v %v\n", "Temp", "Time")
	for _, tick := range list.List {
		tick.Print()
	}
}

type UsageError struct {
	error
}

func errMain() (err error) {
	help := false
	flag.BoolVar(&help, "help", false, "print this help message")
	verbose := false
	flag.BoolVar(&verbose, "verbose", false, "print actions to stdout")
	email := ""
	flag.StringVar(&email, "email", "",
		"email address registered with StatHat")
	id := ""
	flag.StringVar(&id, "station", "", "station id. ex: 1348")
	name := ""
	flag.StringVar(&name, "stat", "",
		`stat name. defaults to "weather-" + station. ex: weather-temp-KBGR`)
	interval := time.Minute * 10
	flag.DurationVar(&interval, "interval", time.Minute*10,
		"time since last run")
	flag.Parse()
	if help {
		flag.Usage()
		return nil
	}
	if email == "" {
		return UsageError{errors.New("email required")}
	}
	if id == "" {
		return UsageError{errors.New("station id required")}
	}
	if name == "" {
		name = "weather-" + id
	}

	prev := time.Now().Add(-interval)

	list, err := NewTickList(id)
	if err != nil {
		return err
	}

	if list.Len() < 1 {
		return fmt.Errorf("no weather data returned: %v: %v", list.Code,
			list.Message)
	}

	recent := list.Tick(list.Len() - 1)
	if recent.Date.Before(prev) {
		if verbose {
			fmt.Println("already reported latest data")
		}
		return nil
	}

	if verbose {
		fmt.Println(name, email, recent.Fahr)
	}
	// the return value is a farce
	_ = stathat.PostEZValue(name, email, recent.Fahr)
	// if you make anything print in here, convert the whole thing to use
	// log instead of fmt.
	finished := stathat.WaitUntilFinished(time.Second * 5)
	if !finished {
		return errors.New("stathat timed out")
	}

	return nil
}

func main() {
	err := errMain()
	if err == nil {
		os.Exit(0)
	}
	if _, ok := err.(UsageError); ok {
		flag.Usage()
	}
	fmt.Fprintf(os.Stderr, "%v: Error: %v\n", filepath.Base(os.Args[0]), err)
	os.Exit(1)
}
