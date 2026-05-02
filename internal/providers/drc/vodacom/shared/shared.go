package shared

import "encoding/json"

func StrPtr(s string) *string {
	return &s
}

func Deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func NormalizeBundleID(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return formatFloatID(x)
	case json.Number:
		return x.String()
	default:
		return ""
	}
}

func formatFloatID(v float64) string {
	if v == float64(int64(v)) {
		return itoa(int64(v))
	}
	return ""
}

func itoa(v int64) string {
	if v == 0 {
		return "0"
	}

	neg := v < 0
	if neg {
		v = -v
	}

	var buf [20]byte
	i := len(buf)

	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}

	if neg {
		i--
		buf[i] = '-'
	}

	return string(buf[i:])
}
