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
		Indexes, Temps template.JS 
	}
	data.Running = running
	
	var idx, tmps []string 
	for i, t := range tempHistory {
		idx = append(idx, fmt.Sprintf("%d", i))
		tmps = append (tmps, fmt.Sprintf("%.1f", t))
	}

	data.Indexes = template.JS(strings.Join(idx, ","))
	data.Temps = template.JS(strings.Join(tmps, ","))
	
	err = t.Execute(w, data)
	if err != nil {
        log.Print(err.Error())
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}
