package validation

import (
	"github.com/nyaruka/phonenumbers"
)

// Phone validates and normalizes phone numbers to E.164.
//
// DefaultRegion is used when the input has no country code
// (e.g. "0712..." → KE). If empty, "KE" is assumed.
type Phone struct {
	DefaultRegion string
}

// allowedRegions: Africa (54) + diaspora
// sanctioned regions IR, KP, CU, SY.
var allowedRegions = map[string]struct{}{
	// Africa (54)
	"DZ": {}, "AO": {}, "BJ": {}, "BW": {}, "BF": {}, "BI": {}, "CM": {},
	"CV": {}, "CF": {}, "TD": {}, "KM": {}, "CG": {}, "CD": {}, "DJ": {},
	"EG": {}, "GQ": {}, "ER": {}, "SZ": {}, "ET": {}, "GA": {}, "GM": {},
	"GH": {}, "GN": {}, "GW": {}, "CI": {}, "KE": {}, "LS": {}, "LR": {},
	"LY": {}, "MG": {}, "MW": {}, "ML": {}, "MR": {}, "MU": {}, "MA": {},
	"MZ": {}, "NA": {}, "NE": {}, "NG": {}, "RW": {}, "ST": {}, "SN": {},
	"SC": {}, "SL": {}, "SO": {}, "ZA": {}, "SS": {}, "SD": {}, "TZ": {},
	"TG": {}, "TN": {}, "UG": {}, "ZM": {}, "ZW": {},
	
	// Diaspora (non-EU)
	"US": {}, "CA": {},
	"AE": {}, "SA": {}, "QA": {}, "KW": {}, "BH": {}, "OM": {},
	"IL": {}, "JO": {}, "LB": {},
	"IN": {}, "PK": {}, "BD": {}, "PH": {},
	"AU": {}, "NZ": {},
	"JM": {}, "TT": {},

	// Europe (Now allowed via Google Meet)
	"GB": {}, "DE": {}, "FR": {}, "NL": {}, "IE": {}, "IT": {}, "ES": {}, "PT": {},
	"BE": {}, "CH": {}, "AT": {}, "SE": {}, "NO": {}, "DK": {}, "FI": {},
	"PL": {}, "CZ": {}, "GR": {}, "HU": {}, "RO": {}, "BG": {}, "HR": {},
	"SK": {}, "SI": {}, "EE": {}, "LV": {}, "LT": {}, "LU": {}, "MT": {}, "CY": {},
	// (You can add the rest of the EU here)
}

// NormalizeAndValidate parses, validates, and normalizes a phone number
// to E.164 (+<country><national>). Returns:
//
//	("", nil)                         when raw is empty (phone is optional)
//	(e164, nil)                       on success
//	("", ErrInvalidPhone)             on malformed/invalid number
//	("", ErrPhoneRegionNotSupported)  when the resolved region is not allowed
func (p Phone) NormalizeAndValidate(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}

	region := p.DefaultRegion
	if region == "" {
		region = "KE"
	}

	num, err := phonenumbers.Parse(raw, region)
	if err != nil {
		return "", ErrInvalidPhone
	}
	if !phonenumbers.IsValidNumber(num) {
		return "", ErrInvalidPhone
	}

	resolved := phonenumbers.GetRegionCodeForNumber(num)
	if _, ok := allowedRegions[resolved]; !ok {
		return "", ErrPhoneRegionNotSupported
	}

	return phonenumbers.Format(num, phonenumbers.E164), nil
}

