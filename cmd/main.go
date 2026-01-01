/** ****************************************************************************************************************** **
	Beer Cellah

** ****************************************************************************************************************** **/

package main

import (
	"beer-cellah/models"

	"github.com/jessevdk/go-flags"
	"github.com/stianeikeland/go-rpio/v4"
	
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
	"syscall"
	"context"
	"sync"
)

  //-------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------//

// give us a name
const apiName = "Beer Cellah"
const apiVersion = "0.3.0"

  //-------------------------------------------------------------------------------------------------------------------//
 //----- CONFIG ------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------//

// final local options object for this executable
var opts struct {
	Help bool `short:"h" long:"help" description:"Shows help message"`
	Port string `short:"p" long:"port" description:"Specifies the target port to run on"`
	Device string `long:"device" description:"Specifies the device name to target for temp" default:"28-3ce1d443d7f3"`
	Target float64 `long:"target" description:"target temperature to stay below in F" default:"55"`
	Templates string `long:"templates" description:"Specifies the folder where the templates are stored"`
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information -v max of -vv"`
	PumpUrl string `long:"pump" description:"Specifies the URL to call to start the pump" default:"http://192.168.68.64"`
}

  //-------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS ---------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------//

func showHelp () {
	fmt.Printf("*****************************\n%s : Version %s\n\n", apiName, apiVersion)

	fmt.Printf("\n*****************************\n")
}

// handles parsing command arguments as well as setting up our opts object
func parseCommandLineArgs () ([]string) {
	// parse things
	args, err := flags.Parse(&opts)
	if err != nil { log.Fatal(err) }

	if opts.Help {
		showHelp()
		os.Exit(0)
	}

	// see what they're trying to do here
	if len(opts.Port) == 0 {
		opts.Port = "8080" // default port
	}

	// see what they're trying to do here
	if len(opts.Templates) == 0 {
		opts.Templates = "/var/www/go/beer-cellah/" // default port
	}

	// check any args
	for _, arg := range args {
		switch strings.ToLower(arg) {
		case "help":
			showHelp()
			os.Exit(0)

		case "version":
			fmt.Printf("%s v%s\n", apiName, apiVersion)
			os.Exit(0)
		}
	}

	return args // return any arguments we don't know what to do with... yet

}

  //-------------------------------------------------------------------------------------------------------------------//
 //----- APP ---------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------//

type app struct {
	running bool 
}

  //-------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS ---------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------//

  //-------------------------------------------------------------------------------------------------------------------//
 //----- MAIN --------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------//

func main() {
	log.SetFlags(log.Lshortfile)
	// first step, parse the command line params
	parseCommandLineArgs()

	log.Printf("Starting %s v%s\nTargetting: %.1fF\n", apiName, apiVersion, opts.Target)

	// main app for everything
	app := &app{
		running: true,
	}

	if len(opts.Device) > 0 {
		err := rpio.Open()
		if err != nil {
			log.Fatal(err)
		}
	}

	// create our server server
	srv := &http.Server {
		Addr: fmt.Sprintf(":%s", opts.Port),
		Handler: app.routes(), 
		ReadTimeout: time.Second * 30,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-c // this sits until something comes into the channel, eg the notify interupts from above
		app.running = false

		srv.Shutdown(context.Background()) // shutdown the server
	}()

	// make sure the cooler isn't running
	if len(opts.Device) > 0 {
		models.StopCooler()
	}

	// make sure the pump isn't running
	models.StopPump(opts.PumpUrl)

	var wg sync.WaitGroup

	/*
	wg.Add(1)
	go models.MonitorButton (&wg, &app.running)
	*/

	// create a ticker for monitoring air temp
	if len(opts.Device) > 0 {
		wg.Add(1)
		go models.MonitorTemp (&wg, &app.running, time.Tick(time.Minute), opts.Target, opts.Device, opts.PumpUrl)
	}

	log.Printf("%s v%s started on port %s\n", apiName, apiVersion, opts.Port) // going to always record this starting message
	if err := srv.ListenAndServe(); err != http.ErrServerClosed { // Error starting or closing listener:
		log.Printf("ListenAndServe: %v", err) // we want to know if this failed for another reason
	}

	log.Println("exiting...")
	wg.Wait()

	if len(opts.Device) > 0 {
		rpio.Close()
	}
	
	os.Exit(0) //final exit
}
