package zonefile

import (
	"fmt"
	"os"

	"github.com/rspamd/goasn/ir"
	"github.com/rspamd/goasn/log"
)

const (
	zoneHeader = `$SOA 43200 asn-ns.rspamd.com support.rspamd.com 0 600 300 86400 300
$NS  43200 asn-ns.rspamd.com asn-ns2.rspamd.com
`
)

func GenerateZone(asnToIRInfo map[uint32]ir.IRASNInfo, prefixToAS map[string]uint32, zonev4 string) error {
	log.Logger.Debug("generating zone")
	defer log.Logger.Debug("generated zone")

	if zonev4 == "" {
		log.Logger.Info("skipping v4 zone generation")
		return nil // FIXME: v6
	}

	f, err := os.Create(zonev4)
	if err != nil {
		return fmt.Errorf("failed to create zonefile: %v", err)
	}
	fmt.Fprint(f, zoneHeader)
	emptyInfo := ir.IRASNInfo{
		IR:      ir.UNKNOWN,
		Country: "--", // FIXME: types
	}
	for prefix, asnNo := range prefixToAS {
		irInfo, ok := asnToIRInfo[asnNo]
		if !ok {
			irInfo = emptyInfo
		}
		fmt.Fprintf(f, "%s %d|%s|%s|%s|\n", prefix, asnNo, prefix, irInfo.Country, irInfo.IR)
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("error closing zonefile: %v", err)
	}
	return nil
}
