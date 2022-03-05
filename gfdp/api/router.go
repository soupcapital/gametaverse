package api

import (
	"net/http"
)

//Router is a URL router
type Router struct {
	rawHdls map[string]URLHandler
}

//NewRouter create a router
func NewRouter() (router *Router) {
	router = new(Router)
	router.rawHdls = make(map[string]URLHandler)
	return router
}

func (r *Router) RegistRaw(URL string, hdl URLHandler) (err error) {
	r.rawHdls[URL] = hdl
	return nil
}

func (r *Router) DealRaw(URL string, w http.ResponseWriter, req *http.Request) (err error) {
	if hdl, ok := r.rawHdls[URL]; ok {
		switch req.Method {
		case http.MethodPost:
			hdl.Post(w, req)
		case http.MethodGet:
			hdl.Get(w, req)
		case http.MethodDelete:
			hdl.Delete(w, req)
		default:
		}
	}
	return nil
}
