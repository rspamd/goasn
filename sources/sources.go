package sources

const (
	AFRINIC_ASN = "https://ftp.afrinic.net/pub/stats/afrinic/delegated-afrinic-latest"
	APNIC_ASN   = "https://ftp.apnic.net/pub/stats/apnic/delegated-apnic-latest"
	ARIN_ASN    = "https://ftp.arin.net/pub/stats/arin/delegated-arin-extended-latest"
	BGP_LATEST  = "http://data.ris.ripe.net/rrc00/latest-bview.gz"
	LACNIC_ASN  = "https://ftp.lacnic.net/pub/stats/lacnic/delegated-lacnic-latest"
	RIPE_ASN    = "https://ftp.ripe.net/ripe/stats/delegated-ripencc-latest"
	IANA_ASN    = "https://www.iana.org/assignments/as-numbers/as-numbers.xml"
	IANA_IP4    = "https://www.iana.org/assignments/ipv4-address-space/ipv4-address-space.xml"
	IANA_IP6    = "https://www.iana.org/assignments/ipv6-address-space/ipv6-address-space.xml"

	// CAIDA RouteViews prefix2as — these "creation log" files enumerate dated
	// dumps; the newest entry is fetched at runtime. Used as additive enrichment
	// on top of the single RIPE RIS collector above.
	CAIDA_RV4_PFX2AS_LOG = "https://publicdata.caida.org/datasets/routing/routeviews-prefix2as/pfx2as-creation.log"
	CAIDA_RV6_PFX2AS_LOG = "https://publicdata.caida.org/datasets/routing/routeviews6-prefix2as/pfx2as-creation.log"
)

func GetAllSources() []string {
	allSources := append(GetRIRASN(), BGP_LATEST)
	allSources = append(allSources, GetIANASources()...)
	return allSources
}

// GetCAIDASources returns the pfx2as-creation.log URLs for CAIDA's v4 and v6
// RouteViews datasets. These are refreshed out-of-band by the caida package
// (rather than via download.RefreshSources) because the actual dated .pfx2as.gz
// files are discovered at runtime and share basenames across datasets.
func GetCAIDASources() []string {
	return []string{CAIDA_RV4_PFX2AS_LOG, CAIDA_RV6_PFX2AS_LOG}
}

// CAIDADatasetDir returns the subdirectory name (relative to the app cache dir)
// in which the caida package stores files for a given log URL. This keeps the
// two datasets' files from colliding — both logs are named pfx2as-creation.log.
func CAIDADatasetDir(logURL string) string {
	switch logURL {
	case CAIDA_RV4_PFX2AS_LOG:
		return "caida-rv4"
	case CAIDA_RV6_PFX2AS_LOG:
		return "caida-rv6"
	default:
		return "caida-unknown"
	}
}

func GetIANASources() []string {
	return []string{IANA_ASN, IANA_IP4, IANA_IP6}
}

func GetASNSources() []string {
	return append(GetRIRASN(), GetIANASources()...)
}

func GetRIRASN() []string {
	return []string{AFRINIC_ASN, APNIC_ASN, ARIN_ASN, LACNIC_ASN, RIPE_ASN}
}
