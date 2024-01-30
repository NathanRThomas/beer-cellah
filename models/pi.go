
package models 

import (
	"github.com/stianeikeland/go-rpio/v4"

	"fmt"
	"time"
	"sync"
)

func runCooler (dur time.Duration, running *bool) {
	fmt.Println("starting cooloer")
	defer fmt.Println("done cooling")

	targetTm := time.Now().Add(dur)

	// pull the pin high
	pin := rpio.Pin(20)
	pin.Output()
	pin.High()
	pin.PullUp()

	for *running && targetTm.After(time.Now()) {

		time.Sleep(time.Second)
	}

	pin.PullDown() // make it low again
	// pin.PullOff()
}

func openDoor () {
	fmt.Println ("opening door")

	pin := rpio.Pin(18)
	pin.Output()
	pin.High() // make it high
	time.Sleep(time.Second * 10)

	pin.Low()
	pin.PullOff()

	fmt.Println ("close door")
}

func MonitorButton (wg *sync.WaitGroup, running *bool) {
	defer wg.Done()
	
	pin := rpio.Pin(24)
	pin.Input()
	pin.PullUp()
	pin.Detect(rpio.FallEdge) // look for it falling to ground

	triggered := false 
	for *running {
		if triggered == false && pin.EdgeDetected() {
			runCooler(time.Minute, running) // open saysme
			triggered = true // i think we're getting re-entry stuff...
		} else {
			triggered = false 
		}
		time.Sleep(time.Second / 2)
	}
}

// returns the air temp in degrees
func CheckAirTemp () float64 {
	return 0
}
