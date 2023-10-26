package iana

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rspamd/goasn/sources"

	"github.com/asergeyev/nradix"
)

type IANAIP4Record struct {
	Prefix      string `xml:"prefix"`
	Designation string `xml:"designation"`
	Status      string `xml:"status"`
}

type IANAIP4Registry struct {
	XMLName xml.Name        `xml:"registry"`
	Records []IANAIP4Record `xml:"record"`
}

func GetReservedIP4(appCacheDir string) (*nradix.Tree, error) {
	tree := nradix.NewTree(0)

	ianaIP4File := sources.MustBasename(sources.IANA_IP4)
	xmlPath := filepath.Join(appCacheDir, ianaIP4File)
	f, err := os.Open(xmlPath)
	if err != nil {
		return tree, err
	}
	defer f.Close()

	dec := xml.NewDecoder(f)
	res := new(IANAIP4Registry)
	err = dec.Decode(res)
	if err != nil {
		return tree, err
	}

	for _, rec := range res.Records {
		if rec.Status == "RESERVED" {
			if !strings.HasSuffix(rec.Prefix, "/8") {
				return tree, fmt.Errorf("not prepared to deal with allocation: %s", rec.Prefix)
			}
			rePrefix := strings.TrimLeft(rec.Prefix, "0")
			if strings.HasPrefix(rePrefix, "/") {
				rePrefix = "0" + rePrefix
			}
			slashIdx := strings.Index(rePrefix, "/")
			rePrefix = rePrefix[:slashIdx] + ".0.0.0/8"
			tree.AddCIDR(rePrefix, 0)
		}
	}

	// these ranges are just footness in IANA XML
	tree.AddCIDR("100.64.0.0/10", 0)
	tree.AddCIDR("169.254.0.0/16", 0)
	tree.AddCIDR("172.16.0.0/12", 0)
	tree.AddCIDR("192.0.2.0/24", 0)
	tree.AddCIDR("192.88.99.0/24", 0)
	tree.AddCIDR("192.88.99.2/32", 0)
	tree.AddCIDR("192.168.0.0/16", 0)
	tree.AddCIDR("192.0.0.0/24", 0)
	tree.AddCIDR("198.18.0.0/15", 0)
	tree.AddCIDR("198.51.100.0/24", 0)
	tree.AddCIDR("203.0.113.0/24", 0)

	return tree, nil
}
