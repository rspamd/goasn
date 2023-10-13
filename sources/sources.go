package sources

const (
	AFRINIC_ASN = "https://ftp.afrinic.net/pub/stats/afrinic/delegated-afrinic-latest"
	APNIC_ASN   = "https://ftp.apnic.net/pub/stats/apnic/delegated-apnic-latest"
	ARIN_ASN    = "https://ftp.arin.net/pub/stats/arin/delegated-arin-extended-latest"
	BGP_LATEST  = "http://data.ris.ripe.net/rrc00/latest-bview.gz"
	LACNIC_ASN  = "https://ftp.lacnic.net/pub/stats/lacnic/delegated-lacnic-latest"
	RIPE_ASN    = "https://ftp.ripe.net/ripe/stats/delegated-ripencc-latest"
	IANA_ASN    = "https://www.iana.org/assignments/as-numbers/as-numbers.xml"
)

func GetAllSources() []string {
	allSources := append(GetRIRASN(), BGP_LATEST)
	allSources = append(allSources, IANA_ASN)
	return allSources
}

func GetASNSources() []string {
	return append(GetRIRASN(), IANA_ASN)
}

func GetRIRASN() []string {
	return []string{AFRINIC_ASN, APNIC_ASN, ARIN_ASN, LACNIC_ASN, RIPE_ASN}
}
