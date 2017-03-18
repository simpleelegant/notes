package main

import (
	"net/http"

	"github.com/simpleelegant/notes/api"
)

func registerRoutes(assetsHandler http.Handler) {
	http.Handle("/", http.RedirectHandler("/assets/", http.StatusMovedPermanently))
	http.Handle("/assets/", assetsHandler)

	a := new(api.Articles)
	http.HandleFunc("/articles/search", a.Search)
	http.HandleFunc("/articles/get", a.Get)
	http.HandleFunc("/articles/create", a.Create)
	http.HandleFunc("/articles/update", a.Update)
	http.HandleFunc("/articles/delete", a.Delete)

	http.HandleFunc("/diagram/render", api.RenderDiagram)
	http.HandleFunc("/md5", api.MD5)

	s := new(api.Settings)
	http.HandleFunc("/settings/get", s.Get)
	http.HandleFunc("/settings/restore", s.Restore)
}
