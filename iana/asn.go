package iana

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rspamd/goasn/ir"
	"github.com/rspamd/goasn/sources"
)

type IANAASNRecord struct {
	Number      string `xml:"number"`
	Description string `xml:"description"`
	lowN        uint32
	highN       uint32
	ir          ir.IRID
}

type IANAASNRegistry struct {
	Records []IANAASNRecord `xml:"record"`
}

type IANAASNInfo struct {
	Registries []IANAASNRegistry `xml:"registry"`
}

func ReadIANAASN(appCacheDir string) (func(uint32) ir.IRID, error) {
	res := &IANAASNInfo{}
	ianaInfo := func(asnNo uint32) ir.IRID {
		for _, reg := range res.Registries {
			for _, rec := range reg.Records {
				if asnNo >= rec.lowN && asnNo <= rec.highN {
					return rec.ir
				}
			}
		}
		return ir.UNKNOWN
	}

	ianaASNFile, err := sources.Basename(sources.IANA_ASN)
	if err != nil {
		return ianaInfo, fmt.Errorf("couldn't find basename for URL(%s): %v",
			sources.IANA_ASN, err)
	}

	xmlPath := filepath.Join(appCacheDir, ianaASNFile)
	f, err := os.Open(xmlPath)
	if err != nil {
		return ianaInfo, err
	}
	defer f.Close()

	dec := xml.NewDecoder(f)
	err = dec.Decode(res)
	if err != nil {
		return ianaInfo, err
	}

	for n, reg := range res.Registries {
		for i, rec := range reg.Records {

			switch rec.Description {
			case "Assigned by AFRINIC":
				rec.ir = ir.AFRINIC
			case "Assigned by ARIN":
				rec.ir = ir.ARIN
			case "Assigned by APNIC":
				rec.ir = ir.APNIC
			case "Assigned by LACNIC":
				rec.ir = ir.LACNIC
			case "Assigned by RIPE NCC":
				rec.ir = ir.RIPE
			case "Reserved":
				rec.ir = ir.RESERVED
			case "Unallocated":
				rec.ir = ir.UNALLOCATED
			default:
				rec.ir = ir.UNKNOWN
			}

			splitStr := strings.SplitN(rec.Number, "-", 2)
			if len(splitStr) == 1 {
				n, err := strconv.Atoi(rec.Number)
				if err != nil {
					return ianaInfo, err
				}
				rec.lowN = uint32(n)
				rec.highN = uint32(n)
			} else {
				lowN, err := strconv.Atoi(splitStr[0])
				if err != nil {
					return ianaInfo, err
				}
				highN, err := strconv.Atoi(splitStr[1])
				if err != nil {
					return ianaInfo, err
				}
				rec.lowN = uint32(lowN)
				rec.highN = uint32(highN)
			}
			reg.Records[i] = rec

		}
		res.Registries[n] = reg

	}
	return ianaInfo, nil
}
