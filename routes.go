package main

import (
	js "encoding/json"
	"net/http"

	"github.com/simpleelegant/notes/api"
)

func registerRoutes(assetsHandler http.Handler) {
	http.Handle("/",
		http.RedirectHandler("/assets/", http.StatusMovedPermanently))
	http.Handle("/assets/", assetsHandler)

	http.HandleFunc("/articles/search", json(api.SearchArticles))
	http.HandleFunc("/articles/get", json(api.GetArticle))
	http.HandleFunc("/articles/create", post(json(api.CreateArticle)))
	http.HandleFunc("/articles/update", post(json(api.UpdateArticle)))
	http.HandleFunc("/articles/delete", post(json(api.DeleteArticle)))

	http.HandleFunc("/diagram/render", json(api.RenderDiagram))
	http.HandleFunc("/md5", json(api.MD5))

	http.HandleFunc("/restore", post(api.Restore))
	http.HandleFunc("/export", post(api.Export))
}

type handler func(*http.Request) (int, interface{})

func json(h handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, body := h(r)
		if err, ok := body.(error); ok {
			w.WriteHeader(status)
			w.Write([]byte(err.Error()))
			return
		}
		z, err := js.Marshal(body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		w.Write(z)
	}
}

func post(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		h(w, r)
	}
}
