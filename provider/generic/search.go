package generic

import (
	"github.com/metafates/mangal/source"
	"github.com/metafates/mangal/util"
)

// Search for mangas by given title
func (s *Scraper) Search(query string) ([]*source.Manga, error) {
	address := s.config.GenerateSearchURL(query)
	address_path := util.UrlGetPath(address)
	if urls, ok := s.mangas[address_path]; ok {
		return urls, nil
	}

	err := s.mangasCollector.Visit(address)

	if err != nil {
		return nil, err
	}

	s.mangasCollector.Wait()
	return s.mangas[address_path], nil
}
