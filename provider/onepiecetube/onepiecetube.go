package onepiecetube

import (
	"github.com/metafates/mangal/source"
)

const (
	Name = "OnePiece-Tube"
	ID   = Name + " built-in"
)

type Onepiecetube struct {
	cache struct {
		chapters *cacher[[]*source.Chapter]
	}
}

func (*Onepiecetube) Name() string {
	return Name
}

func (*Onepiecetube) ID() string {
	return ID
}

func New() *Onepiecetube {
	opt := &Onepiecetube{}

	opt.cache.chapters = newCacher[[]*source.Chapter](ID + "_chapters")

	return opt
}
