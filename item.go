package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"sync"
)

type Item struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type itemReq struct {
	mu    sync.RWMutex
	items map[string]Item
}

func (s *itemReq) getItem(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	item, found := s.items[r.PathValue("id")]
	s.mu.RUnlock()
	if !found {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(item); err != nil {
		log.Printf("write response: %v", err)
	}
}
func (s *itemReq) postItem(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "cannot read request body", http.StatusBadRequest)
		}
		return
	}
	var newItem Item
	if err := json.Unmarshal(body, &newItem); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if newItem.ID == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	if _, exists := s.items[newItem.ID]; exists {
		s.mu.Unlock()
		http.Error(w, "item already exists", http.StatusConflict)
		return
	}
	s.items[newItem.ID] = newItem
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newItem); err != nil {
		log.Printf("write response: %v", err)
	}
}

func (s *itemReq) putItem(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "cannot read request body", http.StatusBadRequest)
		}
		return
	}

	updateId := r.PathValue("id")

	var updateItem struct {
		Name *string `json:"name"`
	}
	if err := json.Unmarshal(body, &updateItem); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if updateItem.Name == nil {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	item, exists := s.items[updateId]
	if !exists {
		s.mu.Unlock()
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}
	item.Name = *updateItem.Name
	s.items[updateId] = item
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(item); err != nil {
		log.Printf("write response: %v", err)
	}
}
