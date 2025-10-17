package fullindex

import (
	"chaos/api/config"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/blevesearch/bleve"
)

// const indexPath = "/data/chaos/fullindex"

// const indexPath = "/Users/jclee/Desktop/solprobe/index"

var (
	index bleve.Index
	mutex sync.Mutex
)

type SearchResult struct {
	ID    string
	Score float64
}

func init() {
	indexPath := config.GetConfig().IndexerRoot.Path
	err := initIndex(indexPath)
	if err != nil {
		fmt.Printf("Failed to initialize index: %v\n", err)
		os.Exit(1)
	}
}

func initIndex(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if mkdirErr := os.MkdirAll(dir, 0755); mkdirErr != nil {
			return fmt.Errorf("failed to create index directory: %v", mkdirErr)
		}
	}
	var err error
	index, err = bleve.Open(path)
	if err != nil {
		if err == bleve.ErrorIndexPathDoesNotExist {
			indexMapping := bleve.NewIndexMapping()
			index, err = bleve.New(path, indexMapping)
			if err != nil {
				return fmt.Errorf("failed to create index: %v", err)
			}
			return nil
		}
		return fmt.Errorf("failed to open index: %v", err)
	}
	return nil
}

func AddToIndex(id string, data map[string]interface{}) error {
	mutex.Lock()
	defer mutex.Unlock()

	if index == nil {
		return fmt.Errorf("index is not initialized")
	}

	err := index.Index(id, data)
	if err != nil {
		return fmt.Errorf("failed to add document to index: %v", err)
	}
	return nil
}

func UpdateIndex(id string, data map[string]interface{}) error {
	mutex.Lock()
	defer mutex.Unlock()

	if index == nil {
		return fmt.Errorf("index is not initialized")
	}

	err := index.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %v", err)
	}

	err = index.Index(id, data)
	if err != nil {
		return fmt.Errorf("failed to update document: %v", err)
	}
	return nil
}

func DeleteIndex(id string) error {
	mutex.Lock()
	defer mutex.Unlock()

	if index == nil {
		return fmt.Errorf("index is not initialized")
	}

	err := index.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %v", err)
	}

	return nil
}

func ExistsInIndex(id string) (bool, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if index == nil {
		return false, fmt.Errorf("index is not initialized")
	}
	doc, err := index.Document(id)
	if err != nil {
		return false, fmt.Errorf("failed to check existence of document: %v", err)
	}

	return doc != nil, nil
}

func SearchIndex(queryString string, size int, sort bool) ([]SearchResult, error) {
	if index == nil {
		return nil, fmt.Errorf("index is not initialized")
	}

	query := bleve.NewQueryStringQuery(queryString)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = size

	if sort {
		searchRequest.SortBy([]string{"-_score"})
	}

	searchResult, err := index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %v", err)
	}

	results := []SearchResult{}
	for _, hit := range searchResult.Hits {
		results = append(results, SearchResult{
			ID:    hit.ID,
			Score: hit.Score,
		})
	}

	return results, nil
}
