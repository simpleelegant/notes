package resources

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/russross/blackfriday"
)

var (
	articleCollectionName = []byte("Document")

	// article field names
	fParent  = []byte("ParentId")
	fTitle   = []byte("Title")
	fContent = []byte("Content")
	fDiagram = []byte("Diagram")
)

// RootArticleID root article's id
const RootArticleID = "53e7d496-a969-4dce-9901-e21ec772b53b"

// errors definitions
var (
	ErrNoArticleCollection = errors.New("article collection not exists")
	ErrArticleNotFound     = errors.New("article not found")
)

// Article resource
type Article struct {
	ID, Parent, Title, Content, Diagram string
}

// GetArticle get an article by its id
func GetArticle(id string) (*Article, error) {
	a := &Article{ID: id}
	err := db.View(func(tx *bolt.Tx) error {
		c, err := articleCollection(tx)
		if err != nil {
			return err
		}
		b := c.Bucket([]byte(id))
		if b == nil {
			return ErrArticleNotFound
		}

		a.Parent = string(b.Get(fParent))
		a.Title = string(b.Get(fTitle))
		a.Content = string(b.Get(fContent))
		a.Diagram = string(b.Get(fDiagram))
		return nil
	})
	return a, err
}

// SearchArticles search articles by title or content pattern
func SearchArticles(pattern string) (
	titleMatches, contentMatches []*ArticleTitle, e error) {
	p := []byte(strings.ToLower(pattern))
	e = db.View(func(tx *bolt.Tx) error {
		c, err := articleCollection(tx)
		if err != nil {
			return err
		}

		cursor := c.Cursor()
		for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
			b := c.Bucket(k)
			title := b.Get(fTitle)
			if bytes.Contains(bytes.ToLower(title), p) {
				titleMatches = append(titleMatches,
					&ArticleTitle{ID: string(k), Title: string(title)})
			} else if bytes.Contains(bytes.ToLower(b.Get(fContent)), p) {
				contentMatches = append(contentMatches,
					&ArticleTitle{ID: string(k), Title: string(title)})
			}
		}

		return nil
	})

	return
}

// Create create an article
func (a *Article) Create() error {
	var err error
	a.ID, err = newID()
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		c, err := articleCollection(tx)
		if err != nil {
			return err
		}

		// if article already exists
		if c.Bucket([]byte(a.ID)) != nil {
			return errors.New("article already exists")
		}

		b, err := c.CreateBucket([]byte(a.ID))
		if err != nil {
			return err
		}
		if err = b.Put(fParent, []byte(a.Parent)); err != nil {
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
func (a *Article) Update(parent, title, content, diagram bool) error {
	return db.Update(func(tx *bolt.Tx) error {
		c, err := articleCollection(tx)
		if err != nil {
			return err
		}
		b := c.Bucket([]byte(a.ID))
		if b == nil {
			return ErrArticleNotFound
		}

		if parent {
			if err = b.Put(fParent, []byte(a.Parent)); err != nil {
				return err
			}
		}
		if title {
			if err = b.Put(fTitle, []byte(a.Title)); err != nil {
				return err
			}
		}
		if content {
			if err = b.Put(fContent, []byte(a.Content)); err != nil {
				return err
			}
		}
		if diagram {
			if err = b.Put(fDiagram, []byte(a.Diagram)); err != nil {
				return err
			}
		}
		return nil
	})
}

// IsAncestorOf check relationship
func (a *Article) IsAncestorOf(testID string) (yes bool, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		c, err := articleCollection(tx)
		if err != nil {
			return err
		}

		for {
			t := c.Bucket([]byte(testID))
			if t == nil {
				break
			}

			testID = string(t.Get(fParent))
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
		c, err := articleCollection(tx)
		if err != nil {
			return err
		}
		if c.Bucket([]byte(a.ID)) == nil {
			return nil
		}
		return c.DeleteBucket([]byte(a.ID))
	})
}

type ArticleTitle struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// GetSubArticles get sub-articles
func (a *Article) GetSubArticles() (subs []*ArticleTitle, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		c, err := articleCollection(tx)
		if err != nil {
			return err
		}

		cursor := c.Cursor()
		for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
			b := c.Bucket(k)
			if string(b.Get(fParent)) == a.ID {
				subs = append(subs,
					&ArticleTitle{ID: string(k), Title: string(b.Get(fTitle))})
			}
		}
		return nil
	})
	return
}

// MD5 calculate md5 digests of content and diagram
func (a *Article) MD5() (contentMD5, diagramMD5 string) {
	return fmt.Sprintf("%x", md5.Sum([]byte(a.Content))),
		fmt.Sprintf("%x", md5.Sum([]byte(a.Diagram)))
}

// ContentHTML convert content which in markdown to HTML
func (a *Article) ContentHTML() []byte {
	return blackfriday.Run([]byte(a.Content))
}

func initArticleCollection() error {
	return db.Update(func(tx *bolt.Tx) error {
		c, err := articleCollection(tx)
		if err != nil {
			if err != ErrNoArticleCollection {
				return err
			}
			// create collection
			c, err = tx.CreateBucket(articleCollectionName)
			if err != nil {
				return err
			}
		}

		// create root article if not exists
		if c.Bucket([]byte(RootArticleID)) == nil {
			b, err := c.CreateBucket([]byte(RootArticleID))
			if err != nil {
				return err
			}
			err = b.Put(fTitle, []byte("First Article"))
			if err != nil {
				return err
			}
			err = b.Put(fContent, []byte("You can edit this article."))
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// CheckArticleCollection check if black have valid structure for Article
func CheckArticleCollection(black *bolt.DB) error {
	return black.View(func(tx *bolt.Tx) error {
		c, err := articleCollection(tx)
		if err != nil {
			return err
		}
		if c.Bucket([]byte(RootArticleID)) == nil {
			return errors.New("no root article in database")
		}

		// TODO more check

		return nil
	})
}

// RestoreArticlesFrom restore article collection from src
func RestoreArticlesFrom(src *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		c, err := articleCollection(tx)
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
			x, err := articleCollection(tx)
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
				if err = b.Put(fParent, y.Get(fParent)); err != nil {
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

func articleCollection(tx *bolt.Tx) (*bolt.Bucket, error) {
	c := tx.Bucket(articleCollectionName)
	if c == nil {
		return nil, ErrNoArticleCollection
	}
	return c, nil
}

func newID() (string, error) {
	const (
		pool   = "1234567890abcdefghijklmnopqrstuvwxyz"
		length = 40
	)

	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for k, v := range b {
		b[k] = pool[int(v)%len(pool)]
	}

	return string(b), nil
}
