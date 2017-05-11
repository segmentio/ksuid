package ksuid

import "bytes"

// lexographic ordering (based on Unicode table) is 0-9A-Za-z
var base62Characters = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")

func base2base(in []byte, inBase int, outBase int) []byte {
	res := []byte{}
	bs := in[:]

	for len(bs) > 0 {
		length := len(bs)
		quotient := []byte{}
		remainder := 0

		for i := 0; i != length; i++ {
			acc := int(bs[i]) + remainder*inBase
			d := acc/outBase | 0
			remainder = acc % outBase

			if len(quotient) > 0 || d > 0 {
				quotient = append(quotient, byte(d))
			}
		}

		// this is really a prepend or remainder to result
		res = append([]byte{byte(remainder)}, res...)
		bs = quotient
	}

	return res
}

func encodeBase62(in []byte) string {
	encoded := base2base(in, 256, 62)
	for i, b := range encoded {
		encoded[i] = base62Characters[b]
	}
	return string(encoded)
}

func decodeBase62(in string) []byte {
	bs := []byte(in)
	decoded := make([]byte, len(bs))
	for i, b := range bs {
		// O(1)... technically. Has better real-world perf than a map
		decoded[i] = byte(bytes.IndexByte(base62Characters, b))
	}

	return base2base(decoded, 62, 256)
}
