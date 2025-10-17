package fullindex

import (
	"fmt"
	"log"
	"testing"
)

func TestSearch(t *testing.T) {
	data := map[string]interface{}{
		"title":   "Token ca Search",
		"content": "Bleve is a modern text search library for Go.",
	}
	err := AddToIndex("doc2", data)
	if err != nil {
		log.Fatalf("Failed to add document: %v", err)
	}

	// 搜索
	results, err := SearchIndex("Bleve", 10, true)
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}
	fmt.Println("Search Results:", results)
}
