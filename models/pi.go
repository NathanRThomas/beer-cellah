
package models 

import (
	"github.com/stianeikeland/go-rpio/v4"

	"fmt"
	"time"
	"sync"
	"os"
	"log"
	"strconv"
	"strings"
)

func runCooler (dur time.Duration, running *bool) {
	fmt.Println("starting cooloer")
	defer fmt.Println("done cooling")

	targetTm := time.Now().Add(dur)

	// pull the pin high
	pin := rpio.Pin(12)
	pin.Output()
	pin.High()
	
	for *running && targetTm.After(time.Now()) {

		time.Sleep(time.Second)
	}

	//pin.PullDown() // make it low again
	pin.Low()
	pin.PullOff()
}

func MonitorButton (wg *sync.WaitGroup, running *bool) {
	defer wg.Done()
	
	pin := rpio.Pin(24)
	pin.Input()
	pin.PullUp()
	pin.Detect(rpio.FallEdge) // look for it falling to ground

	triggered := false 
	for *running {
		if pin.EdgeDetected() {
			if triggered == false {
				runCooler(time.Minute, running) // open saysme
				triggered = true // i think we're getting re-entry stuff...
			}
			// else we ignore it, it's still triggering
		} else {
			triggered = false  // we're in the clear
		}
		time.Sleep(time.Second / 2)
	}
}

// monitors the temp to know when to run things
func MonitorTemp (wg *sync.WaitGroup, running *bool, c <-chan time.Time) {
	defer wg.Done()

	for {
		select {
		case <-c:
			// we got something to do
			tmp := CheckAirTemp()
			fmt.Println("Checking air temp: ", tmp)
			// check the temp, see if we need to do anything
			if tmp > 70 { 

				// we need to do something
				runCooler(time.Minute, running)
			}
		default:
			time.Sleep(time.Second)
		}

		if *running == false {
			return // we're done
		}
	}
}

// returns the air temp in degrees f
func CheckAirTemp () float64 {
	data, err := os.ReadFile("/sys/bus/w1/devices/28-3ce1d4434b6a/w1_slave")
	if err != nil {
		log.Printf("CheckAirTemp: %v\n", err)
		return 0
	}

	if len(data) < 30 {
		log.Printf("CheckAirTemp: didn't get a body: %s\n", string(data))
		return 0
	}

	deg := data[len(data)-7:]
	if deg[0] != '=' {
		log.Printf("CheckAirTemp: Not expected body: %s :: %s\n", string(deg), string(data))
		return 0
	}

	cString := strings.TrimSpace(string(deg[1:]))

	degC, err := strconv.Atoi(cString)
	if err != nil {
		log.Printf("CheckAirTemp: Not expected int: %v : %s : %s\n", err, cString, string(data))
		return 0
	}

	return ((float64(degC) / 1000) * (9/5)) + 32
}
