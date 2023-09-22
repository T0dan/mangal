package provider

import (
	"github.com/metafates/mangal/provider/generic"
	"github.com/metafates/mangal/provider/mangadex"
	"github.com/metafates/mangal/provider/manganato"
	"github.com/metafates/mangal/provider/manganelo"
	"github.com/metafates/mangal/provider/mangapill"
	"github.com/metafates/mangal/provider/mangaplus"
	"github.com/metafates/mangal/provider/onepiecetube"
	"github.com/metafates/mangal/source"
)

const CustomProviderExtension = ".lua"

var builtinProviders = []*Provider{
	{
		ID:   mangadex.ID,
		Name: mangadex.Name,
		CreateSource: func() (source.Source, error) {
			return mangadex.New(), nil
		},
	},
	{
		ID:   onepiecetube.ID,
		Name: onepiecetube.Name,
		CreateSource: func() (source.Source, error) {
			return onepiecetube.New(), nil
		},
	},
	{
		ID:   mangaplus.ID,
		Name: mangaplus.Name,
		CreateSource: func() (source.Source, error) {
			return mangaplus.New(), nil
		},
	},
}

func init() {
	for _, conf := range []*generic.Configuration{
		manganelo.Config,
		manganato.Config,
		mangapill.Config,
	} {
		conf := conf
		builtinProviders = append(builtinProviders, &Provider{
			ID:   conf.ID(),
			Name: conf.Name,
			CreateSource: func() (source.Source, error) {
				return generic.New(conf), nil
			},
		})
	}
}
