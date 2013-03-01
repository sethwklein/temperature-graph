package main

import (
	"encoding/json"
	"errors"
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
	List []*Tick
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
		return
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(buf, &list)
	if err != nil {
		return
	}
	list.init()
	return
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

func errMain() (err error) {
	verbose := true
	email := ""
	id := "1348"
	name := "KBGR"
	interval := 60
	interval = 60 * 3

	prev := time.Now().Add(time.Minute * -time.Duration(interval))

	list, err := NewTickList(id)
	if err != nil {
		return err
	}

	if list.Len() < 1 {
		// BUG(sk): likely there's useful data in the json on error
		return errors.New("no weather data returned")
	}

	recent := list.Tick(list.Len() - 1)
	if recent.Date.Before(prev) {
		if verbose {
			fmt.Println("already reported latest data")
		}
		return nil
	}

	name = "weather-temp" + name
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
	fmt.Fprintf(os.Stderr, "%v: Error: %v\n", filepath.Base(os.Args[0]), err)
	os.Exit(1)
}
