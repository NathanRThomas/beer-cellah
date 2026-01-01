
package models 

import (
	"github.com/stianeikeland/go-rpio/v4"

	"fmt"
	"time"
	"sync"
	"os"
	"log"
	"strconv"
	"net/http"
	"strings"
)

const coolerRelayPin	= 12

var tempHistory []float32
var coolingHistory []bool
var coolerRunning bool 

func addTemp (t float64) {
	if len(tempHistory) > 60 * 24 * 2 { // keeping about 48 hours of history
		tempHistory = tempHistory[1:]
		coolingHistory = coolingHistory[1:]
	}

	tempHistory = append (tempHistory, float32(t)) // i'm keeping these as 32 bit simply for the ram usage
	coolingHistory = append (coolingHistory, coolerRunning) // just record if we happened to be running
}

func ReturnStats () (bool, []float32, []bool) {
	return coolerRunning, tempHistory, coolingHistory
}

func waitForIt (dur time.Duration, running *bool) {
	targetTm := time.Now().Add(dur)

	for *running && targetTm.After(time.Now()) {

		time.Sleep(time.Second)
	}
}

func runCooler (dur time.Duration, running *bool) {
	fmt.Println("starting cooloer")
	defer fmt.Println("done cooling")

	// pull the pin high
	pin := rpio.Pin(coolerRelayPin)
	pin.Output()
	pin.High()
	
	waitForIt (dur, running)

	// stop it
	StopCooler()
}

// runs the pump to get cool water into the radiator
// we don't want to run this too long, we'll max the radiator quickly
// going to run this for 30 seconds at a time
// with 10 minutes between runs
func runPump (pumpUrl string) {
	
	for coolerRunning { // loop forever		
		
		fmt.Println("running pump")
		// API call to start the pump
		resp, err := http.Get(fmt.Sprintf("%s/rpc/Switch.Set?id=0&on=true", pumpUrl))
		if err != nil {
			fmt.Printf("runPump: Error starting pump: %v\n", err)
			waitForIt (time.Second * 30, &coolerRunning)
			continue
		}
		resp.Body.Close() // close the response body

		// wait for 30 seconds
		waitForIt (time.Second * 30, &coolerRunning) // wait for 30 seconds

		fmt.Println("stopping pump")
		// now turn the pump off
		resp, err = http.Get(fmt.Sprintf("%s/rpc/Switch.Set?id=0&on=false", pumpUrl))
		if err != nil {
			fmt.Printf("runPump: Error stopping pump: %v\n", err)
			waitForIt (time.Second * 30, &coolerRunning)
			continue
		}
		resp.Body.Close() // close the response body

		// now wait 10 minutes
		waitForIt (time.Minute * 10, &coolerRunning)
	} // end of for loop
}

func StopCooler () {
	// pull the pin high
	pin := rpio.Pin(coolerRelayPin)
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

func IsNight () bool {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		fmt.Println("error loading location: ", err)
		return false
	}
	now := time.Now().In(loc)
	return now.Hour() >= 21 || now.Hour() < 7 // 9pm to 7am
}

// monitors the temp to know when to run things
func MonitorTemp (wg *sync.WaitGroup, running *bool, c <-chan time.Time, target float64, device string, pumpUrl string) {
	defer wg.Done()

	for {
		select {
		case <-c:
			// we got something to do
			tmp := CheckAirTemp(device)
			// fmt.Println("Checking air temp: ", tmp)
			// check the temp, see if we need to do anything
			if tmp > target { 
				fmt.Printf("Air temp %.1fF over target %.1fF\n", tmp - target, target)

				// don't run the cooler between 10pm and 7am ET
				if IsNight() {
					fmt.Println("not running cooler at night")
					waitForIt (time.Minute * 10, running)
					break
				}

				// pull the pin high
				pin := rpio.Pin(coolerRelayPin)
				pin.Output()
				pin.High()
				coolerRunning = true

				// first launch the pump, this runs its own cycles
				go runPump (pumpUrl)

				// this is for the fan control, which we want running anytime it's too warm
				for {
					// now we loop for 1 minute at a time, checking for the temp to be lower
					waitForIt(time.Minute, running)

					tmp = CheckAirTemp(device)

					// make it a little colder than the target to reduce switching all the time
					if tmp < target - 0.6 || *running == false || IsNight() {
						break // bail early
					}
				}

				// we're good now
				pin.Low()
				pin.PullOff()
				coolerRunning = false // this disables the pump thread as well

				fmt.Printf("Done Cooling. Air temp %.1fF\n", tmp)
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
func CheckAirTemp (device string) float64 {
	data, err := os.ReadFile(fmt.Sprintf("/sys/bus/w1/devices/%s/w1_slave", device))
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

	degF := ((float64(degC) / 1000) * 9) / 5 + 32

	addTemp(degF)
	return degF
}
