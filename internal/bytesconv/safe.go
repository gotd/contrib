// +build appengine nounsafe

package bytesconv

func B2S(b []byte) string {
	return string(b)
}

func S2B(s string) []byte {
	return []byte(s)
}
