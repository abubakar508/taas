package util

import "regexp"

var vodacomDRCRegex = regexp.MustCompile(`^243[0-9]{9}$`)

func IsValidDRCVodacomMSISDN(msisdn string) bool {
	return vodacomDRCRegex.MatchString(msisdn)
}
