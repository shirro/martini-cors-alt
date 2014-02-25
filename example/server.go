package main

import (
	"github.com/codegangsta/martini"
	//	"github.com/shirro/martini-cors-alt"
	"github.com/shirro/martini-cors-alt"
)

// Implement cors.Site for our specific requirements
type Site struct {
	Id   int
	Name string
}

type SiteMap map[string]*Site

func (sm SiteMap) SetContext(origin string, ctx martini.Context) bool {
	site, ok := sm[origin]
	//ctx.MapTo(site, (*Site)(nil))
	ctx.Map(site)
	return ok
}

func main() {
	m := martini.Classic()

	// If you want custom behaviour for different sites based on
	// Origins header. Just beware this is trivially faked so don't
	// use in place of authentication.
	myOrigins := &SiteMap{
		"http://127.0.0.1":   {1, "Localhost"},
		"https://shirro.com": {2, "My domain"},
	}

	myCors := &cors.Cors{
		Origins: myOrigins,
		Headers: cors.StandardHeaders,
		//		Tolerant: true,
	}

	m.Use(myCors.MiddleWare)
	m.NotFound(myCors.NotFound, cors.RealNotFound)

	m.Get("/hello/:name", func(params martini.Params, site *Site) string {
		// If we have Tolerent set we get here instead of a 403
		// and need to test value of site
		if site == nil {
			return "You don't belong here " + params["name"]
		}
		return "Hello " + params["name"] + " welcome to " + site.Name
	})

	m.Run()
}
