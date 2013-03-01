package main

import (
	"encoding/json"
	"fmt"
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
	id := "1348"

	list, err := NewTickList(id)
	if err != nil {
		return err
	}
	list.Print()

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
