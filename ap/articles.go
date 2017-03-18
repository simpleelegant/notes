package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/simpleelegant/notes/models"
)

// Articles resource
type Articles struct{}

// Create create an article
func (*Articles) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	a := &models.Article{
		ParentID: r.FormValue("parent_id"),
		Title:    r.FormValue("title"),
		Content:  r.FormValue("content"),
	}
	if err := a.Create(); err != nil {
		replyBadRequest(w, err)
		return
	}

	a.CalculateMD5()

	reply(w, http.StatusOK, a)
}

type GetResult struct {
	*models.Article
	ContentMD5    string `json:"content_md5"`
	DiagramMD5    string `json:"diagram_md5"`
	ContentInHTML string `json:"content_in_html"`

	SuperArticleTitle string `json:"super_article_title"`

	SubArticles     []*models.Article `json:"sub_articles"`
	SiblingArticles []*models.Article `json:"sibling_articles"`
}

// Get get an article
func (*Articles) Get(w http.ResponseWriter, r *http.Request) {
	a := &models.Article{ID: r.FormValue("id")}
	if a.ID == "" {
		a.ID = models.RootArticleID
	}
	err := a.Read()
	if err != nil {
		replyBadRequest(w, err)
		return
	}

	re := &GetResult{Article: a}
	re.ContentMD5, re.DiagramMD5 = a.CalculateMD5()

	// render content in HTML if asked for
	if r.FormValue("html") != "" {
		re.ContentInHTML = string(a.ConvertContentToHTML())
	}

	// get sub-articles if asked for
	if r.FormValue("sub") != "" {
		re.SubArticles, err = a.GetSubArticles()
		if err != nil {
			replyBadRequest(w, err)
			return
		}
	}

	if a.ParentID != "" {
		// get super article if asked for
		if r.FormValue("sup") != "" {
			s := &models.Article{ID: a.ParentID}
			if err := s.Read(); err != nil {
				replyBadRequest(w, err)
				return
			}
			re.SuperArticleTitle = s.Title

			// get subling articles if asked for
			if r.FormValue("subling") != "" {
				re.SiblingArticles, err = s.GetSubArticles()
				if err != nil {
					replyBadRequest(w, err)
					return
				}
			}
		}
	}

	reply(w, http.StatusOK, re)
}

// Search search articles
func (*Articles) Search(w http.ResponseWriter, r *http.Request) {
	s, err := (*models.Article)(nil).SearchByTitle(r.FormValue("title"))
	if err != nil {
		replyBadRequest(w, err)
		return
	}

	reply(w, http.StatusOK, s)
}

// Delete delete an articles
func (*Articles) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	a := &models.Article{ID: r.FormValue("id")}
	if err := a.Read(); err != nil {
		replyBadRequest(w, err)
		return
	}

	// deny to delete root article
	if a.ID == models.RootArticleID {
		replyBadRequest(w, errors.New("unable to delete root article"))
		return
	}

	// deny if has sub-articles
	subs, err := a.GetSubArticles()
	if err != nil {
		replyBadRequest(w, err)
		return
	}
	if len(subs) != 0 {
		replyBadRequest(w, errors.New("unable to delete this article, because it has sub-articles"))
		return
	}

	if err := a.Delete(); err != nil {
		replyBadRequest(w, err)
		return
	}

	reply(w, http.StatusNoContent, "")
}

// Update update an article
func (*Articles) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	a := &models.Article{ID: r.FormValue("id")}
	if err := a.Read(); err != nil {
		replyBadRequest(w, err)
		return
	}

	contentMD5, diagramMD5 := a.CalculateMD5()

	if v, ok := formValue(r, "title"); ok {
		a.Title = v
	}
	if v, ok := formValue(r, "content"); ok {
		if r.FormValue("before_content_md5") != contentMD5 {
			replyBadRequest(w, errors.New("Remote content was changed by another operation."))
			return
		}
		a.Content = v
	}
	if v, ok := formValue(r, "diagram"); ok {
		if r.FormValue("before_diagram_md5") != diagramMD5 {
			replyBadRequest(w, errors.New("Remote diagram was changed by another operation."))
			return
		}
		a.Diagram = v
	}
	if v, ok := formValue(r, "parent_id"); ok {
		v = strings.TrimSpace(v)
		if v == a.ID {
			replyBadRequest(w, errors.New("parent article do not equal to current article"))
			return
		}

		// check wanted parent document exists
		if err := (&models.Article{ID: v}).Read(); err != nil {
			replyBadRequest(w, err)
			return
		}

		yes, err := a.IsAncestorOf(v)
		if err != nil {
			replyBadRequest(w, err)
			return
		}
		if yes {
			replyBadRequest(w, errors.New("specified article has been sub-article of current article"))
			return
		}
	}

	if err := a.Update(); err != nil {
		replyBadRequest(w, err)
		return
	}

	a.CalculateMD5()
	a.ConvertContentToHTML()

	reply(w, http.StatusOK, a)
}
