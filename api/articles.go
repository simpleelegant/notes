package api

import (
	"errors"
	"net/http"

	"github.com/simpleelegant/notes/diagram"
	"github.com/simpleelegant/notes/resources"
)

// CreateArticle create an article
func CreateArticle(r *http.Request) (int, interface{}) {
	a := &resources.Article{
		Parent:  formValue(r, "parent"),
		Title:   formValue(r, "title"),
		Content: formValue(r, "content"),
	}
	if err := a.Create(); err != nil {
		return http.StatusBadRequest, err
	}
	return http.StatusOK, map[string]string{"id": a.ID}
}

// GetArticle get an article
func GetArticle(r *http.Request) (int, interface{}) {
	id := formValue(r, "id")
	if id == "" {
		id = resources.RootArticleID
	}
	a, err := resources.GetArticle(id)
	if err != nil {
		return http.StatusBadRequest, err
	}
	contentMD5, diagramMD5 := a.MD5()
	var diagramSVG string
	if a.Diagram != "" {
		out, err := diagram.Parse([]byte(a.Diagram))
		if err != nil {
			diagramSVG = err.Error()
		} else {
			diagramSVG = string(out)
		}
	}
	b := map[string]string{
		"id":         a.ID,
		"title":      a.Title,
		"content":    a.Content,
		"html":       string(a.ContentHTML()),
		"diagram":    a.Diagram,
		"diagramSVG": diagramSVG,
		"contentMD5": contentMD5,
		"diagramMD5": diagramMD5,
	}

	// get sub-articles of a
	subArticles, err := a.GetSubArticles()
	if err != nil {
		return http.StatusBadRequest, err
	}

	// get parent and subling articles
	var (
		parent  *resources.ArticleTitle
		subling []*resources.ArticleTitle
	)
	if a.Parent != "" {
		p, err := resources.GetArticle(a.Parent)
		if err != nil {
			return http.StatusBadRequest, err
		}
		parent = &resources.ArticleTitle{ID: p.ID, Title: p.Title}
		subling, err = p.GetSubArticles()
		if err != nil {
			return http.StatusBadRequest, err
		}
	}

	return http.StatusOK, map[string]interface{}{
		"parent":            parent,
		"childrenOfParent":  subling,
		"current":           b,
		"childrenOfCurrent": subArticles,
	}
}

// SearchArticles search articles
func SearchArticles(r *http.Request) (int, interface{}) {
	titleMatches, contentMatches, err := resources.SearchArticles(formValue(r, "pattern"))
	if err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, map[string]interface{}{
		"titleMatches":   titleMatches,
		"contentMatches": contentMatches,
	}
}

// DeleteArticle delete an articles
func DeleteArticle(r *http.Request) (int, interface{}) {
	a, err := resources.GetArticle(formValue(r, "id"))
	if err != nil {
		return http.StatusBadRequest, err
	}

	// deny to delete root article
	if a.ID == resources.RootArticleID {
		return http.StatusBadRequest, errors.New("unable to delete root article")
	}

	// deny if has sub-articles
	subs, err := a.GetSubArticles()
	if err != nil {
		return http.StatusBadRequest, err
	}
	if len(subs) != 0 {
		return http.StatusBadRequest, errors.New("must delete sub-articles")
	}

	if err := a.Delete(); err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, ""
}

// UpdateArticle update an article
func UpdateArticle(r *http.Request) (int, interface{}) {
	a, err := resources.GetArticle(formValue(r, "id"))
	if err != nil {
		return http.StatusBadRequest, err
	}
	contentMD5, diagramMD5 := a.MD5()

	var uParent, uTitle, uContent, uDiagram bool
	{
		if formValue(r, "uParent") == "true" {
			uParent = true
		}
		if formValue(r, "uTitle") == "true" {
			uTitle = true
		}
		if formValue(r, "uContent") == "true" {
			uContent = true
		}
		if formValue(r, "uDiagram") == "true" {
			uDiagram = true
		}
	}

	if uParent {
		parent := formValue(r, "parent")
		if err := checkChangeParent(a, parent); err != nil {
			return http.StatusBadRequest, err
		}
		a.Parent = parent
	}
	if uTitle {
		a.Title = formValue(r, "title")
	}
	if uContent {
		if formValue(r, "originalContentMD5") != contentMD5 {
			return http.StatusBadRequest,
				errors.New("content was changed by another operation")
		}
		a.Content = formValue(r, "content")
	}
	if uDiagram {
		if formValue(r, "originalDiagramMD5") != diagramMD5 {
			return http.StatusBadRequest,
				errors.New("diagram was changed by another operation.")
		}
		a.Diagram = formValue(r, "diagram")
	}

	if err := a.Update(uParent, uTitle, uContent, uDiagram); err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, "updated"
}

func checkChangeParent(a *resources.Article, parent string) error {
	if a.ID == parent {
		return errors.New("parent article cannot equal to current article")
	}
	if _, err := resources.GetArticle(parent); err != nil {
		return err
	}
	yes, err := a.IsAncestorOf(parent)
	if err != nil {
		return err
	}
	if yes {
		return errors.New("specified article has been sub-article of current article")
	}
	return nil
}
