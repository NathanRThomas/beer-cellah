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

	running, tempHistory := models.ReturnStats()

	var data struct {
		Running bool 
		Temps template.JS 
		CurrentTemp, MaxTemp, MinTemp string 
	}
	data.Running = running

	var max, min float32
	
	if len(tempHistory) > 0 {
		max = tempHistory[0]
		min = tempHistory[0]
		data.CurrentTemp = fmt.Sprintf("%.1fF", tempHistory[len(tempHistory)-1])
	}
	
	var tmps []string 
	for _, t := range tempHistory {
		tmps = append (tmps, fmt.Sprintf("%.1f", t))

		if t > max {
			max = t 
		} else if t < min {
			min = t 
		}
	}

	data.Temps = template.JS(strings.Join(tmps, ","))
	data.MaxTemp = fmt.Sprintf("%.1fF", max)
	data.MinTemp = fmt.Sprintf("%.1fF", min)
	
	err = t.Execute(w, data)
	if err != nil {
        log.Print(err.Error())
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}
