package mrt

import (
	"fmt"
	"path/filepath"

	"github.com/rspamd/goasn/ir"
	"github.com/rspamd/goasn/log"
	"github.com/rspamd/goasn/sources"

	"github.com/osrg/gobgp/v3/pkg/packet/bgp"
	"github.com/osrg/gobgp/v3/pkg/packet/mrt"
	"go.uber.org/zap"
)

func ASNFromBGP(appCacheDir string, ianaASN func(uint32) ir.IRID) (map[string]uint32, []string, int, error) {
	mrtParseErrors := make([]string, 0)
	mrtParseErrorCount := 0
	prefixToAS := make(map[string]uint32)

	bviewFile, err := sources.Basename(sources.BGP_LATEST)
	if err != nil {
		return prefixToAS, mrtParseErrors, mrtParseErrorCount,
			fmt.Errorf("couldn't get basename for URL: %v", err)
	}
	bviewPath := filepath.Join(appCacheDir, bviewFile)
	rdr, err := NewMRTReader(bviewPath)
	if err != nil {
		return prefixToAS, mrtParseErrors, mrtParseErrorCount,
			fmt.Errorf("failed to create MRT reader: %v", err)
	}
	defer rdr.Close()

	mrtParseErrorMap := make(map[string]struct{})

	log.Logger.Debug("reading MRT file")
	for {
		more, message, err := rdr.Next()
		if err != nil {
			mrtParseErrorMap[err.Error()] = struct{}{}
			mrtParseErrorCount++
			if !more {
				break
			} else {
				continue
			}
		}
		switch message.Header.Type {
		case mrt.TABLE_DUMPv2:
			switch message.Header.SubType {
			case uint16(mrt.PEER_INDEX_TABLE):
				// FIXME: anything useful to be done with this?
			case uint16(mrt.RIB_IPV4_UNICAST):
				ribMessage := message.Body.(*mrt.Rib)
				prefix := ribMessage.Prefix.String()
				if prefix == "0.0.0.0/0" {
					continue
				}
				for _, entry := range ribMessage.Entries {
					for _, pa := range entry.PathAttributes {
						switch pa.GetType() {
						case bgp.BGP_ATTR_TYPE_AS_PATH:
							asPath := pa.(*bgp.PathAttributeAsPath)
							asList := asPath.Value[0].GetAS()
							pathLen := len(asList)
							if pathLen == 0 {
								continue
							}
							lastIdx := pathLen - 1
							var originAS uint32
							var ianaAllocation ir.IRID
							for lastIdx >= 0 {
								originAS = asList[lastIdx]
								ianaAllocation := ianaASN(originAS)
								// IANA XML's UNALLOCATED status appears to be nonsense
								if ianaAllocation == ir.UNKNOWN || ianaAllocation == ir.RESERVED {
									lastIdx = lastIdx - 1
								} else {
									break
								}
							}
							if ianaAllocation == ir.UNKNOWN || ianaAllocation == ir.RESERVED {
								log.Logger.Warn("ignoring announcement from apparently bogus ASNs",
									zap.String("prefix", prefix),
									zap.Any("as_list", asList))
								continue
							}
							// XXX: we are using last seen conceivable origin AS
							prefixToAS[prefix] = originAS
						default:
						}
					}
				}
			case uint16(mrt.RIB_IPV6_UNICAST):
				// FIXME: IP6
			default:
				log.Logger.Warn("table dump subtype was not handled",
					zap.Uint16("subtype", uint16(message.Header.SubType)))
			}
		default:
			log.Logger.Warn("MRT type was unhandled",
				zap.Uint16("type", uint16(message.Header.Type)))
		}
		if !more {
			break
		}
	}
	for k := range mrtParseErrorMap {
		mrtParseErrors = append(mrtParseErrors, k)
	}
	log.Logger.Debug("read MRT file")
	return prefixToAS, mrtParseErrors, mrtParseErrorCount, nil
}
