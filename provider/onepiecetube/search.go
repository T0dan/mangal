package onepiecetube

import (
	"github.com/metafates/mangal/source"
)

func (o *Onepiecetube) Search(query string) ([]*source.Manga, error) {
	var mangas []*source.Manga

	manga := source.Manga{
		Name:   "One Piece",
		URL:    "https://onepiece-tube.com/kapitel-mangaliste",
		Index:  0,
		ID:     "https://onepiece-tube.com/kapitel-mangaliste",
		Source: o,
	}
	manga.Metadata.LanguageISO = "de"

	mangas = append(mangas, &manga)

	return mangas, nil
}
