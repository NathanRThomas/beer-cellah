/** ****************************************************************************************************************** **
	General handlers

** ****************************************************************************************************************** **/

package main

import (
	"beer-cellah/models"
	// "github.com/pkg/errors"

	"fmt"
	"html/template"
	"net/http"
	"log"
	"strings"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONST -------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- HANDLERS ----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *app) getStatus(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(opts.Templates + "templates/status.html")
	if err != nil {
		log.Fatal(err)
	}

	running, tempHistory, coolingHistory := models.ReturnStats()

	var data struct {
		Running bool 
		Temps, Target, CoolingHistory template.JS 
		CurrentTemp, MaxTemp, MinTemp string 
	}
	data.Running = running
	data.Target = template.JS(fmt.Sprintf("%.0f", opts.Target))

	var max, min float32
	
	if len(tempHistory) > 0 {
		max = tempHistory[0]
		min = tempHistory[0]
		data.CurrentTemp = fmt.Sprintf("%.1fF", tempHistory[len(tempHistory)-1])
	}
	
	var tmps []string 
	for _, t := range tempHistory {
		tmps = append (tmps, fmt.Sprintf("%.0f", t))

		if t > max {
			max = t 
		} else if t < min {
			min = t 
		}
	}

	data.Temps = template.JS(strings.Join(tmps, ","))

	var cools []string 
	for _, t := range coolingHistory {
		if t {
			cools = append (cools, "1")
		} else {
			cools = append (cools, "0")
		}
	}

	data.CoolingHistory = template.JS(strings.Join(cools, ","))

	if len(data.CurrentTemp) == 0 {
		data.CurrentTemp = "-"
		data.MaxTemp = "-"
		data.MinTemp = "-"
	} else {
		data.MaxTemp = fmt.Sprintf("%.1fF", max)
		data.MinTemp = fmt.Sprintf("%.1fF", min)
	}
	
	err = t.Execute(w, data)
	if err != nil {
        log.Print(err.Error())
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}
