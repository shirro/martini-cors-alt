// Package cors provides an alternative CORS implementation for Martini.
//
// It relies on the Route.MethodsFor to provide Access-Control-Allow-Methods.
//
// Due to the modularity of Martini, Routes is not available in Middleware
// handlers so both a Middleware and NotFound handler are required.
package cors

import (
	"github.com/codegangsta/martini"
	"net/http"
	"strings"
	"sync"
)

// Sites provides the origin check for Cors
type Sites interface {
	// SetContext return boolean indicating if origin is known.
	// It may set site in request context if required
	SetContext(string, martini.Context) bool
}

// Cors type allows configuration of CORS handling
type Cors struct {
	// Allowed origins is now provided through an interface
	Origins Sites
	// CORS headers. Please use write mutex if updating while live.
	HeadersMutex sync.RWMutex
	Headers      map[string]string
	// Default is to return 403 if Origin not a match. Set to true to disable.
	Tolerant bool
	DevKey   string
}

// StandardHeaders are not really a standard. Customised headers should be provided.
var StandardHeaders = map[string]string{
	"Access-Control-Max-Age":        "86000",
	"Access-Control-Allow-Headers":  "Content-Type, Origin, Authorization",
	"Access-Control-Expose-Headers": "Content-Length",
}

// Middleware checks the Origin header on requests and adds appropriate CORS headers to
// the response.
func (cors *Cors) MiddleWare(w http.ResponseWriter, r *http.Request, ctx martini.Context) {
	var origin string

	// Postman can't set Origin headers so I test with this
	if cors.DevKey != "" && r.Header.Get("X-Dev") == cors.DevKey {
		origin = r.Header.Get("X-Origin")
	} else {
		origin = r.Header.Get("Origin")
	}

	// Set Access-Control-Allow-Origin
	h := w.Header()
	originOk := cors.setOrigin(h, origin, ctx)

	// Conditionally set 403 if Origin was not a match
	if !originOk && !cors.Tolerant {
		w.WriteHeader(http.StatusForbidden)
		return
	}

}

// NotFound handles Cors preflight requests. Options routes will shadow it.
func (cors *Cors) NotFound(w http.ResponseWriter, r *http.Request, routes martini.Routes) {

	// Leave if not a preflight request
	if r.Method != "OPTIONS" || len(r.Header.Get("Origin")) == 0 {
		return
	}

	// MethodsFor could be expensive with lots of routes.
	// It might help to increase Access-Control-Max-Age
	methods := routes.MethodsFor(r.URL.Path)

	// If this Url has no methods leave it to the next handler
	if len(methods) == 0 {
		return
	}

	h := w.Header()
	// Set all the CORS headers other than Access-Control-Allow-{Origin,Methods}
	cors.setHeaders(h)
	h.Set("Access-Control-Allow-Methods", stringMethods(methods))
	w.WriteHeader(http.StatusOK)

}

func (cors *Cors) setOrigin(h http.Header, origin string, ctx martini.Context) bool {

	// Reader lock so we can change the map dynamically
	// Empty Origins map allows all domains
	if cors.Origins == nil {
		h.Set("Access-Control-Allow-Origin", "*")
		return true
	}

	// Allow request if Origin in map
	if ok := cors.Origins.SetContext(origin, ctx); ok {
		h.Set("Access-Control-Allow-Origin", origin)
		return true
	}

	// Default
	return false
}

func (cors *Cors) setHeaders(h http.Header) {
	// Reader lock so we can change the headers dynamically
	cors.HeadersMutex.RLock()
	defer cors.HeadersMutex.RUnlock()
	for header, value := range cors.Headers {
		h.Set(header, value)
	}
}

// RealNotFound provides an alternative to Martini's inbuilt handler as using a
// CORS NotFound means we have to handle it ourselves.
func RealNotFound(w http.ResponseWriter, r *http.Request, routes martini.Routes) {
	// We throw in 405 handling for free (or for the cost of a MethodsFor call)
	methods := routes.MethodsFor(r.URL.Path)

	// If no methods on this path it is a a 404 and return
	if len(methods) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Otherwise a 405 with Allow header
	w.Header().Set("Allow", stringMethods(methods))
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func stringMethods(methods []string) string {
	methods = append(methods, "OPTIONS")
	return strings.Join(methods, ",")
}
