package models

import (
	"crypto/md5"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/russross/blackfriday"
)

var (
	articleCollection = []byte("Document")

	// article field names
	fParentID = []byte("ParentId")
	fTitle    = []byte("Title")
	fContent  = []byte("Content")
	fDiagram  = []byte("Diagram")
)

// RootArticleID root article's id
const RootArticleID = "53e7d496-a969-4dce-9901-e21ec772b53b"

// errors definitions
var (
	ErrNoArticleCollection  = errors.New("no article collection")
	ErrArticleNotFound      = errors.New("article not found")
	ErrArticleAlreadyExists = errors.New("article already exists")
)

// Article model
type Article struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Diagram  string `json:"diagram"`
}

func (a *Article) newID() error {
	pool := "1234567890abcdefghijklmnopqrstuvwxyz"
	length := 40

	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return err
	}

	for k, v := range b {
		b[k] = pool[int(v)%len(pool)]
	}

	a.ID = string(b)

	return nil
}

// Create create an article
func (a *Article) Create() error {
	if err := a.newID(); err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		c, err := a.getCollection(tx)
		if err != nil {
			return err
		}

		// if article already exists
		if c.Bucket([]byte(a.ID)) != nil {
			return ErrArticleAlreadyExists
		}

		b, err := c.CreateBucket([]byte(a.ID))
		if err != nil {
			return err
		}

		if err = b.Put(fParentID, []byte(a.ParentID)); err != nil {
			return err
		}
		if err = b.Put(fTitle, []byte(a.Title)); err != nil {
			return err
		}
		if err = b.Put(fContent, []byte(a.Content)); err != nil {
			return err
		}
		if err = b.Put(fDiagram, []byte(a.Diagram)); err != nil {
			return err
		}

		return nil
	})
}

// Update update an article
func (a *Article) Update() error {
	return db.Update(func(tx *bolt.Tx) error {
		c, err := a.getCollection(tx)
		if err != nil {
			return err
		}

		b := c.Bucket([]byte(a.ID))
		if b == nil {
			return ErrArticleNotFound
		}

		if err = b.Put(fParentID, []byte(a.ParentID)); err != nil {
			return err
		}
		if err = b.Put(fTitle, []byte(a.Title)); err != nil {
			return err
		}
		if err = b.Put(fContent, []byte(a.Content)); err != nil {
			return err
		}
		if err = b.Put(fDiagram, []byte(a.Diagram)); err != nil {
			return err
		}

		return nil
	})
}

// IsAncestorOf check relationship
func (a *Article) IsAncestorOf(testID string) (yes bool, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		c, err := a.getCollection(tx)
		if err != nil {
			return err
		}

		for {
			t := c.Bucket([]byte(testID))
			if t == nil {
				break
			}

			testID = string(t.Get(fParentID))
			if testID == "" {
				// when t is root article
				break
			}
			if testID == a.ID {
				yes = true
				break
			}
		}

		return nil
	})

	return
}

// Delete article by id, without sub-articles
func (a *Article) Delete() error {
	return db.Update(func(tx *bolt.Tx) error {
		c, err := a.getCollection(tx)
		if err != nil {
			return err
		}

		if c.Bucket([]byte(a.ID)) == nil {
			return nil
		}

		return c.DeleteBucket([]byte(a.ID))
	})
}

// GetSubArticles get sub-articles
func (a *Article) GetSubArticles() (subs []*Article, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		c, err := a.getCollection(tx)
		if err != nil {
			return err
		}

		cursor := c.Cursor()
		for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
			b := c.Bucket(k)
			if string(b.Get(fParentID)) == a.ID {
				subs = append(subs, &Article{ID: string(k), Title: string(b.Get(fTitle))})
			}
		}

		return nil
	})

	return
}

// Read read an article by id
func (a *Article) Read() error {
	return db.View(func(tx *bolt.Tx) error {
		c, err := a.getCollection(tx)
		if err != nil {
			return err
		}

		b := c.Bucket([]byte(a.ID))
		if b == nil {
			return ErrArticleNotFound
		}

		a.ParentID = string(b.Get(fParentID))
		a.Title = string(b.Get(fTitle))
		a.Content = string(b.Get(fContent))
		a.Diagram = string(b.Get(fDiagram))

		return nil
	})
}

// SearchByTitle search articles by title pattern
func (a *Article) SearchByTitle(t string) (s []*Article, e error) {
	t = strings.ToLower(t)

	e = db.View(func(tx *bolt.Tx) error {
		c, err := a.getCollection(tx)
		if err != nil {
			return err
		}

		cursor := c.Cursor()
		for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
			title := string(c.Bucket(k).Get(fTitle))
			if strings.Contains(strings.ToLower(title), t) {
				s = append(s, &Article{ID: string(k), Title: title})
			}
		}

		return nil
	})

	return
}

// CalculateMD5 calculate md5 digests of content and diagram
func (a *Article) CalculateMD5() (contentMD5, diagramMD5 string) {
	return fmt.Sprintf("%x", md5.Sum([]byte(a.Content))),
		fmt.Sprintf("%x", md5.Sum([]byte(a.Diagram)))
}

// ConvertContentToHTML convert content which in markdown to HTML
func (a *Article) ConvertContentToHTML() []byte {
	return blackfriday.MarkdownCommon([]byte(a.Content))
}

func (a *Article) getCollection(tx *bolt.Tx) (*bolt.Bucket, error) {
	c := tx.Bucket(articleCollection)
	if c == nil {
		return nil, ErrNoArticleCollection
	}

	return c, nil
}

func (a *Article) initCollection() error {
	return db.Update(func(tx *bolt.Tx) error {
		c, err := a.getCollection(tx)
		switch err {
		case nil:
			// ignore
		case ErrNoArticleCollection:
			// create collection
			c, err = tx.CreateBucket(articleCollection)
			if err != nil {
				return err
			}
		default:
			return err
		}

		// create root article if not exists
		if c.Bucket([]byte(RootArticleID)) == nil {
			b, err := c.CreateBucket([]byte(RootArticleID))
			if err != nil {
				return err
			}

			if err := b.Put(fTitle, []byte("First Article")); err != nil {
				return err
			}
			if err := b.Put(fContent, []byte("You can edit this article.")); err != nil {
				return err
			}
		}

		return nil
	})
}

// CheckCollection check if black have valid structure for Article
func (a *Article) CheckCollection(black *bolt.DB) error {
	return black.View(func(tx *bolt.Tx) error {
		c, err := a.getCollection(tx)
		if err != nil {
			return err
		}

		if c.Bucket([]byte(RootArticleID)) == nil {
			return errors.New("no root article in database")
		}

		// XXX more checking

		return nil
	})
}

// Restore restore collection from src
func (a *Article) Restore(src *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		c, err := a.getCollection(tx)
		if err != nil {
			return err
		}

		// clean db
		cursor := c.Cursor()
		for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
			if err := c.DeleteBucket(k); err != nil {
				return err
			}
		}

		// copy articles from src
		return src.View(func(tx *bolt.Tx) error {
			x, err := a.getCollection(tx)
			if err != nil {
				return err
			}

			cursor := x.Cursor()
			for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
				y := x.Bucket(k)
				b, err := c.CreateBucketIfNotExists(k)
				if err != nil {
					return err
				}
				if err = b.Put(fParentID, y.Get(fParentID)); err != nil {
					return err
				}
				if err = b.Put(fTitle, y.Get(fTitle)); err != nil {
					return err
				}
				if err = b.Put(fContent, y.Get(fContent)); err != nil {
					return err
				}
				if err = b.Put(fDiagram, y.Get(fDiagram)); err != nil {
					return err
				}
			}

			return nil
		})
	})
}
