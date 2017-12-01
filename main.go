package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

// bestTime will return new Time with year, month and day from "now" time and
// hour and minute from "later" time.
// If "later" time is behind "now" hour-wise it will shift the "now" time 1 day
// forward.
// For example, if now is 15:00 and later is 14:00 then later should happen the
// next day.
func bestTime(now, later time.Time) time.Time {
	nowh, nowm, _ := now.Clock()
	laterh, laterm, _ := later.Clock()

	var nextDay bool

	if laterh < nowh {
		nextDay = true
	} else if laterh == nowh && laterm <= nowm {
		nextDay = true
	}

	if nextDay {
		now = now.AddDate(0, 0, 1)
	}

	return time.Date(now.Year(), now.Month(), now.Day(), later.Hour(), later.Minute(), 0, 0, now.Location())
}

func playMP3(r io.ReadCloser) error {
	s, format, err := mp3.Decode(r)
	if err != nil {
		return err
	}

	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
		return err
	}

	done := make(chan struct{})

	speaker.Play(beep.Seq(s, beep.Callback(func() {
		close(done)
	})))

	<-done

	return nil
}

func main() {
	var userTime string
	var filename string

	flag.StringVar(&userTime, "when", "", "when to ring the alarm -- required")
	flag.StringVar(&filename, "file", "", "mp3 file to play -- required")
	flag.Parse()

	if userTime == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if filename == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	now := time.Now().Local()

	alarmTime, err := time.Parse("15:04", userTime)
	if err != nil {
		log.Fatalf("can't parse time: %s.\nMake sure time is in hh:mm format.", err)
	}

	later := bestTime(now, alarmTime)
	diff := later.Sub(now)
	fmt.Printf("now: %s,\nlater: %s,\ndiff: %s\n\n", now, later, diff)

	fmt.Printf("sleeping for %s\n", diff)
	time.Sleep(diff)
	fmt.Println("woke up, playing")

	// resp, err := http.Get("https://www.npr.org/streams/mp3/nprlive24.m3u")
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	if err := playMP3(f); err != nil {
		log.Fatal(err)
	}
}