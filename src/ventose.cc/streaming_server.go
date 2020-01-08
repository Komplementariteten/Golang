package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"ventose.cc/datastreamer"
)

const defaultNetString = "localhost:62253"

func HandleSignals(ctrl chan bool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)
	s := <-c
	if s == os.Kill {
		ctrl <- false
	} else {
		ctrl <- true
	}
}

func GeneratreTestData() *datastreamer.Content {
	d := &datastreamer.Content{}
	var gates []*datastreamer.Gatesignal
	d.Fe = 1
	d.Ctp = 2345
	d.Channel = 1
	for i := 0; i < 50; i++ {
		g := &datastreamer.Gatesignal{}
		g.Amp = 1234
		g.GateATof = 110
		g.GateBTof = 120
		gates = append(gates, g)
	}
	d.Gate = gates
	return d
}

func initializeDefaultStreamer(counter chan int, ctrl <-chan bool) chan bool {
	defaultData := GeneratreTestData()
	quitter := datastreamer.NewProtoBStreamer(defaultNetString, defaultData, ctrl, counter)
	return quitter
}

func main() {
	def := flag.Bool("default", true, "Use default setup and type. Server listens on {"+defaultNetString+"}")
	ctl := make(chan bool, 2)

	go HandleSignals(ctl)

	logger := log.New(os.Stdout, "[ProtoBufferStreamer]", log.LstdFlags)
	counter := datastreamer.GetIntCounter(logger)

	if *def {
		q := initializeDefaultStreamer(counter, ctl)
		<-q
	} else {
		logger.Fatal("Not jet implemented")
	}
}
