package main

import (
	"math/rand"
	"time"
	"fmt"
)

var (
	basepool int = 8
	pairs = []int{5,4,3,3,3,2,2,1}
	offpool int = 0
	devpool int = 0
	a_hit int = 0
	wins int = 0
	average_hits map[int]float64
	av_off float64 = 0
	av_def float64 = 0
	near = 0
	carry_over = 0
	carry_overHit = 0
)

func r() int {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(10)
	return r
}

func roll(dice int, diff int) (int, string) {
	var successes = 0
	var m = ""
	for i := 0; i < dice; i++ {
		if r() >= diff {
			successes++
			m = m + "1"
		}else {
			m = m + "0"
		}
	}
	return successes, m
}

func updatePools() {
	rand.Seed(time.Now().UnixNano())
	op := rand.Intn(len(pairs) - 1)

	offpool = basepool - pairs[op]
	devpool = pairs[op]
}

func damage(hits int) int {
	rand.Seed(time.Now().UnixNano())
	v := 10 - hits
	if v > 0 {
		return rand.Intn(v) + hits
	}else {
		return 15
	}
}

func maneuver(hits int, def int, th int) int {
	if hits + 1 > def {
		near++
		return 5
	}
	if carry_over > 0 {
		ret := hits + carry_over - def
		if ret > 0 && damage(ret) > 6 {
			carry_overHit++
			carry_over = 0
		}
		return 0
	} else {
		carry_over = hits
	}
	return 0
}

func main() {
	average_hits = make(map[int]float64)
	deffsum := 0
	offsum := 0
	hits := 0
	runs := 0
	for defvalue := 4; defvalue < 11; defvalue++ {
		wins = 0
		deffsum = 0
		offsum = 0
		for p:= 0; p < 200; p++ {
			updatePools()
			for i := 0; i < 1000; i++ {
				a, _ := roll(offpool, defvalue)
				if a > devpool {
					// Hit
					a = a - devpool
					a_hit = a_hit + a
				} else {
					a = a - devpool
					a = maneuver(a_hit + a, devpool, defvalue)
					if a > 0 {
						a_hit = a_hit + a
					}
				}
				runs++
				if a > 0 {
					damage := damage(a_hit)
					if damage > 6 {
						wins++

						deffsum = deffsum + devpool
						offsum = offsum + offpool
						hits = hits + runs
						a_hit = 0
						runs = 0
					}
				}
			}
		}
		average_hits[defvalue] = float64(hits) / float64(wins)
		av_off = (float64(offsum) / float64(wins))
		av_def = (float64(deffsum) / float64(wins))
		fmt.Printf("Def: %v => Ø %v hits to wound on Ø off: %v and Ø def: %v\n", defvalue, average_hits[defvalue], av_off, av_def)
		wonByNear := (float64(near) / float64(wins)) * 100
		fmt.Printf("%v% Won by near manipulation\n", wonByNear)
		wonByCarryOver := (float64(carry_overHit) / float64(wins)) * 100
		fmt.Printf("%v% Won by carry over\n", wonByCarryOver)
		near = 0
		carry_over = 0
	}
	av_off = (float64(offsum) / float64(wins))
	av_def = (float64(deffsum) / float64(wins))
	fmt.Printf("%v result in %v wins, Ø Off: %v - Ø Def: %v", runs, wins, av_off, av_def)
}
