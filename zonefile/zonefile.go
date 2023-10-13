package zonefile

import (
	"fmt"
	"os"

	"github.com/rspamd/goasn/ir"
	"github.com/rspamd/goasn/log"

	"go.uber.org/zap"
)

const (
	zoneHeader = `$SOA 43200 asn-ns.rspamd.com support.rspamd.com 0 600 300 86400 300
$NS  43200 asn-ns.rspamd.com asn-ns2.rspamd.com
`
)

func GenerateZones(asnToIRInfo map[uint32]ir.IRASNInfo, v4AS map[string]uint32, zonev4 string, v6AS map[string]uint32, zonev6 string) error {
	if zonev4 == "" {
		log.Logger.Info("skipping v4 zone generation")
	} else {
		err := GenerateZone(asnToIRInfo, v4AS, zonev4)
		if err != nil {
			return fmt.Errorf("v4 zone generation failed: %v", err)
		}
	}
	if zonev6 == "" {
		log.Logger.Info("skipping v6 zone generation")
	} else {
		err := GenerateZone(asnToIRInfo, v6AS, zonev6)
		if err != nil {
			return fmt.Errorf("v6 zone generation failed: %v", err)
		}
	}
	return nil
}

func GenerateZone(asnToIRInfo map[uint32]ir.IRASNInfo, prefixToAS map[string]uint32, zone string) error {
	log.Logger.Debug("generating zone", zap.String("name", zone))
	defer log.Logger.Debug("generated zone", zap.String("name", zone))

	f, err := os.Create(zone)
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
