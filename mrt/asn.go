package mrt

import (
	"fmt"
	"path/filepath"

	"github.com/rspamd/goasn/ir"
	"github.com/rspamd/goasn/log"
	"github.com/rspamd/goasn/sources"

	"github.com/asergeyev/nradix"
	"github.com/osrg/gobgp/v3/pkg/packet/bgp"
	"github.com/osrg/gobgp/v3/pkg/packet/mrt"
	"go.uber.org/zap"
)

type BGPASNInfo struct {
	Err             error
	ParseErrors     []string
	ParseErrorCount int
	V4              map[string]uint32
	V6              map[string]uint32
}

func NewBGPASNInfo() *BGPASNInfo {
	return &BGPASNInfo{
		V4: make(map[string]uint32),
		V6: make(map[string]uint32),
	}
}

func ASNFromBGP(appCacheDir string, ianaASN func(uint32) ir.IRID, rejectFile string, reserved4 *nradix.Tree, reserved6 *nradix.Tree) *BGPASNInfo {
	res := NewBGPASNInfo()

	bviewFile := sources.MustBasename(sources.BGP_LATEST)
	bviewPath := filepath.Join(appCacheDir, bviewFile)
	rdr, err := NewMRTReader(bviewPath, rejectFile)
	if err != nil {
		res.Err = fmt.Errorf("failed to create MRT reader: %v", err)
		return res
	}
	defer rdr.Close()

	mrtParseErrorMap := make(map[string]struct{})

	log.Logger.Debug("reading MRT file")
	for {
		more, message, err := rdr.Next()
		if err != nil {
			mrtParseErrorMap[err.Error()] = struct{}{}
			res.ParseErrorCount++
			if !more {
				break
			} else {
				continue
			}
		}
		if !more {
			break
		}
		prefixToAS := res.V4
		reservedIP := reserved4
		switch message.Header.Type {
		case mrt.TABLE_DUMPv2:
			switch message.Header.SubType {
			case uint16(mrt.PEER_INDEX_TABLE):
				// not directly interesting
			case uint16(mrt.RIB_IPV6_UNICAST):
				prefixToAS = res.V6
				reservedIP = reserved6
				fallthrough
			case uint16(mrt.RIB_IPV4_UNICAST):
				ribMessage := message.Body.(*mrt.Rib)
				prefix := ribMessage.Prefix.String()
				inf, err := reservedIP.FindCIDR(prefix)
				if err != nil {
					log.Logger.Error("radix lookup failed", zap.Error(err))
				} else if inf != nil {
					log.Logger.Debug("ignoring reserved range", zap.String("range", prefix))
					continue
				} else if prefix == "0.0.0.0/0" || prefix == "::/0" {
					log.Logger.Debug("ignoring null route", zap.String("range", prefix))
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
								// XXX: IANA XML's UNALLOCATED status appears to be nonsense?
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
		res.ParseErrors = append(res.ParseErrors, k)
	}
	log.Logger.Debug("read MRT file")
	return res
}
