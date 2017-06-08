package ksuid

import "strings"

// lexographic ordering (based on Unicode table) is 0-9A-Za-z
var base62Characters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func appendBase2Base(dst []byte, src []byte, inBase int, outBase int) []byte {
	bs := src[:]
	bq := [stringEncodedLength]byte{}

	for len(bs) > 0 {
		length := len(bs)
		quotient := bq[:0]
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
		dst = append(dst, byte(remainder))
		bs = quotient
	}

	reverse(dst)
	return dst
}

func base2base(src []byte, inBase int, outBase int) []byte {
	return appendBase2Base(nil, src, inBase, outBase)
}

func appendEncodeBase62(dst []byte, src []byte) []byte {
	dst = appendBase2Base(dst, src, 256, 62)
	for i, c := range dst {
		dst[i] = base62Characters[c]
	}
	return dst
}

func encodeBase62(in []byte) []byte {
	return appendEncodeBase62(nil, in)
}

func appendDecodeBase62(dst []byte, src []byte) []byte {
	// Kind of intrusive, we modify the input buffer... it's OK here, it saves
	// a memory allocation in Parse.
	for i, b := range src {
		// O(1)... technically. Has better real-world perf than a map
		src[i] = byte(strings.IndexByte(base62Characters, b))
	}
	return appendBase2Base(dst, src, 62, 256)
}

func decodeBase62(src []byte) []byte {
	return appendDecodeBase62(
		make([]byte, 0, len(src)*2),
		append(make([]byte, 0, len(src)), src...),
	)
}

func reverse(b []byte) {
	i := 0
	j := len(b) - 1

	for i < j {
		b[i], b[j] = b[j], b[i]
		i++
		j--
	}
}
