package datastreamer

import (
	"log"
	"time"
)

func GetIntCounter(l *log.Logger) chan int {
	counter := make(chan int)
	go func(c chan int) {
		var total int64 = 0
		average := 0
		part := 0
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case n := <-c:
				total += int64(n)
				part += n
			case t := <-ticker.C:
				if part == 0 {
					continue
				}
				average = part / (t.Nanosecond() / 1000000)
				part = 0
				l.Printf("%d Kb/s", average)
			}
		}
	}(counter)
	return counter
}
