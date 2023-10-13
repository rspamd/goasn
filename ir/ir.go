package ir

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type IRID string

const (
	AFRINIC     = "afrinic"
	APNIC       = "apnic"
	ARIN        = "arin"
	LACNIC      = "lacnic"
	RESERVED    = "reserved"
	RIPE        = "ripencc"
	UNKNOWN     = "--"
	UNALLOCATED = "unallocated" // ???
)

type IRASNInfo struct {
	IR      IRID
	Country string // FIXME: type
}

func ReadIRData(appCacheDir string, sources []string) (map[uint32]IRASNInfo, error) {
	asnInfo := make(map[uint32]IRASNInfo, 0)

	for _, fPath := range sources {
		fullPath := filepath.Join(appCacheDir, fPath)
		f, err := os.Open(fullPath)
		if err != nil {
			return asnInfo, fmt.Errorf("couldn't open file(%s) for reading: %v", fullPath, err)
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			splitStr := strings.Split(scanner.Text(), "|")
			if len(splitStr) < 7 {
				continue
			}

			entIR := splitStr[0]
			entCountry := splitStr[1]
			entType := splitStr[2]
			entValue := splitStr[3]
			entPlusTxt := splitStr[4]

			entPlus, err := strconv.Atoi(entPlusTxt)
			if err != nil {
				return asnInfo, fmt.Errorf("couldn't convert string(%s) to number: %v",
					entPlusTxt, err)
			}

			if entType == "asn" {
				if entCountry == "" {
					entCountry = "--"
				}
				irInfo := &IRASNInfo{
					Country: entCountry,
					IR:      IRID(entIR),
				}
				as, err := strconv.Atoi(entValue)
				if err != nil {
					return asnInfo, fmt.Errorf("couldn't convert string(%s) to number: %v",
						entValue, err)
				}
				for i := 1; i <= entPlus; i++ {
					asnNumber := uint32(as)
					asnInfo[asnNumber] = *irInfo
					as = as + 1
				}
			}
		}
		if err := scanner.Err(); err != nil {
			return asnInfo, fmt.Errorf("scanner error: %v", err)
		}
	}

	return asnInfo, nil
}
