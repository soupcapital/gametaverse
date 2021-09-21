package api

import (
	"fmt"
	"net/http"
)

func init() {

}

//URLHandler is the interface of URL Handler
type URLHandler interface {
	// Get is GET request
	Get(w http.ResponseWriter, r *http.Request)
	// Post is POST request
	Post(w http.ResponseWriter, r *http.Request)
}

//URLHdl is Handler's base class
type URLHdl struct {
	server *Server
}

func (hdl *URLHdl) svrError(w http.ResponseWriter) {
	fmt.Fprintf(w, "500 Server Error")
}
