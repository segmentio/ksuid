package ksuid

import (
	"testing"
	"time"
)

func TestCompressedSet(t *testing.T) {
	tests := []struct {
		scenario string
		function func(*testing.T)
	}{
		{
			scenario: "sparse",
			function: testCompressedSetSparse,
		},
		{
			scenario: "packed",
			function: testCompressedSetPacked,
		},
		{
			scenario: "mixed",
			function: testCompressedSetMixed,
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, test.function)
	}
}

func testCompressedSetSparse(t *testing.T) {
	now := time.Now()

	times := [100]time.Time{}
	for i := range times {
		times[i] = now.Add(time.Duration(i) * 2 * time.Second)
	}

	ksuids := [1000]KSUID{}
	for i := range ksuids {
		ksuids[i], _ = NewRandomWithTime(times[i%len(times)])
	}

	set := Compress(ksuids[:]...)

	for i, it := 0, set.Iter(); it.Next(); {
		if i >= len(ksuids) {
			t.Error("too many KSUIDs were produced by the set iterator")
			break
		}
		if ksuids[i] != it.KSUID {
			t.Errorf("bad KSUID at index %d: expected %s but found %s", i, ksuids[i], it.KSUID)
		}
		i++
	}

	reportCompressionRatio(t, ksuids[:], set)
}

func testCompressedSetPacked(t *testing.T) {
	sequences := [10]Sequence{}
	for i := range sequences {
		sequences[i] = Sequence{Seed: New()}
	}

	ksuids := [1000]KSUID{}
	for i := range ksuids {
		ksuids[i], _ = sequences[i%len(sequences)].Next()
	}

	set := Compress(ksuids[:]...)

	for i, it := 0, set.Iter(); it.Next(); {
		if i >= len(ksuids) {
			t.Error("too many KSUIDs were produced by the set iterator")
			break
		}
		if ksuids[i] != it.KSUID {
			t.Errorf("bad KSUID at index %d: expected %s but found %s", i, ksuids[i], it.KSUID)
		}
		i++
	}

	reportCompressionRatio(t, ksuids[:], set)
}

func testCompressedSetMixed(t *testing.T) {
	now := time.Now()

	times := [20]time.Time{}
	for i := range times {
		times[i] = now.Add(time.Duration(i) * 2 * time.Second)
	}

	sequences := [200]Sequence{}
	for i := range sequences {
		seed, _ := NewRandomWithTime(times[i%len(times)])
		sequences[i] = Sequence{Seed: seed}
	}

	ksuids := [1000]KSUID{}
	for i := range ksuids {
		ksuids[i], _ = sequences[i%len(sequences)].Next()
	}

	set := Compress(ksuids[:]...)

	for i, it := 0, set.Iter(); it.Next(); {
		if i >= len(ksuids) {
			t.Error("too many KSUIDs were produced by the set iterator")
			break
		}
		if ksuids[i] != it.KSUID {
			t.Errorf("bad KSUID at index %d: expected %s but found %s", i, ksuids[i], it.KSUID)
		}
		i++
	}

	reportCompressionRatio(t, ksuids[:], set)
}

func reportCompressionRatio(t *testing.T, ksuids []KSUID, set CompressedSet) {
	len1 := byteLength * len(ksuids)
	len2 := len(set)
	t.Logf("original %d B, compressed %d B (%.4g%%)", len1, len2, 100*(1-(float64(len2)/float64(len1))))
}

func BenchmarkCompressedSet(b *testing.B) {
	ksuids1 := [1000]KSUID{}
	ksuids2 := [1000]KSUID{}

	for i := range ksuids1 {
		ksuids1[i] = New()
	}

	ksuids2 = ksuids1
	buf := make([]byte, 0, 1024)
	set := Compress(ksuids2[:]...)

	b.Run("write", func(b *testing.B) {
		for i := 0; i != b.N; i++ {
			ksuids2 = ksuids1
			buf = AppendCompressed(buf[:0], ksuids2[:]...)
		}
	})

	b.Run("read", func(b *testing.B) {
		for i := 0; i != b.N; i++ {
			for it := set.Iter(); true; {
				if !it.Next() {
					break
				}
			}
		}
	})
}
