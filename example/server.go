package main

import (
	"github.com/codegangsta/martini"
	//	"github.com/shirro/martini-cors-alt"
	"shirro.com/martini-cors-alt"
)

func main() {
	m := martini.Classic()

	/*
		empty := struct{}{}
		corsOrigins := map[string]struct{}{
			"https://github.com": empty, // Go really needs sets
			"https://shirro.com": empty,
			"https://127.0.0.1":  empty,
		}
		corsHeaders := map[string]string{
			"Access-Control-Max-Age":        "604800",
			"Access-Control-Allow-Headers":  "Content-Type, Origin, Authorization",
			"Access-Control-Expose-Headers": "Content-Length, X-My-Header",
		}

		myCors := &cors.Cors{Origins: corsOrigins, Headers: corsHeaders, Tolerant: true}
	*/
	myCors := &cors.Cors{Headers: cors.StandardHeaders}

	m.Use(myCors.MiddleWare)
	m.NotFound(myCors.NotFound, cors.RealNotFound)

	m.Get("/hello/:name", func(params martini.Params) string {
		return "Hello " + params["name"]
	})

	m.Run()
}
