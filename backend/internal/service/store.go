package service

import (
	"github.com/example/bid-maker-backend/internal/model"
	"sync"
)

var documentStore struct {
	sync.RWMutex
	documents map[string]*model.Document
}

func init() {
	documentStore.documents = make(map[string]*model.Document)
}

func StoreDocument(doc *model.Document) {
	documentStore.Lock()
	defer documentStore.Unlock()
	documentStore.documents[doc.ID] = doc
}

func GetDocument(id string) (*model.Document, bool) {
	documentStore.RLock()
	defer documentStore.RUnlock()
	doc, ok := documentStore.documents[id]
	return doc, ok
}

func UpdateDocument(doc *model.Document) {
	documentStore.Lock()
	defer documentStore.Unlock()
	documentStore.documents[doc.ID] = doc
}
