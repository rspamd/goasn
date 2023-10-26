package iana

import (
	"encoding/xml"
	"os"
	"path/filepath"

	"github.com/rspamd/goasn/sources"

	"github.com/asergeyev/nradix"
)

type IANAIP6Record struct {
	Prefix      string `xml:"prefix"`
	Description string `xml:"description"`
}

type IANAIP6Registry struct {
	Records []IANAIP6Record `xml:"record"`
}

type IANAIP6Info struct {
	XMLName  xml.Name        `xml:"registry"`
	Registry IANAIP6Registry `xml:"registry"`
}

func GetReservedIP6(appCacheDir string) (*nradix.Tree, error) {
	tree := nradix.NewTree(0)

	ianaIP6File := sources.MustBasename(sources.IANA_IP6)
	xmlPath := filepath.Join(appCacheDir, ianaIP6File)
	f, err := os.Open(xmlPath)
	if err != nil {
		return tree, err
	}
	defer f.Close()

	dec := xml.NewDecoder(f)
	res := new(IANAIP6Info)
	err = dec.Decode(res)
	if err != nil {
		return tree, err
	}

	for _, rec := range res.Registry.Records {
		switch rec.Description {
		case "Reserved by IETF":
			fallthrough
		case "Link-Scoped Unicast":
			fallthrough
		case "Unique Local Unicast":
			tree.AddCIDR(rec.Prefix, 0)
		default:
		}
	}

	return tree, nil
}
