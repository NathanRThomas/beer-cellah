/** ****************************************************************************************************************** **
	Endpoints supported by the pi

** ****************************************************************************************************************** **/

package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	
	"fmt"
	"encoding/json"
	"net/http"
	"time"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- ROUTES ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *app) versionGet(w http.ResponseWriter, r *http.Request) {
	out, _ := json.Marshal(fmt.Sprintf("%s v%s", apiName, apiVersion))
	w.Write(out)
}


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- HANDLERS ----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

// Our custom router/handler for this
func (this *app) routes() http.Handler {
	// our base chi router
	// https://go-chi.io/#/pages/middleware
	r := chi.NewRouter()

	// rate limit by ip
	r.Use(httprate.LimitByIP(100, time.Minute)) // it is public afterall

	// finally our actual handlers
	r.HandleFunc("/status", this.getStatus) // for k8 health checks
	r.HandleFunc("/version", this.versionGet) // for k8 health checks

	return r
}
