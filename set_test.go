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
		{
			scenario: "iterating over a nil compressed set returns no ids",
			function: testCompressedSetNil,
		},
		{
			scenario: "concatenating multiple compressed sets is supported",
			function: testCompressedSetConcat,
		},
		{
			scenario: "duplicate ids are appear only once in the compressed set",
			function: testCompressedSetDuplicates,
		},
		{
			scenario: "building a compressed set with a single id repeated multiple times produces the id only once",
			function: testCompressedSetSingle,
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

func testCompressedSetDuplicates(t *testing.T) {
	sequence := Sequence{Seed: New()}

	ksuids := [1000]KSUID{}
	for i := range ksuids[:10] {
		ksuids[i], _ = sequence.Next() // exercise dedupe on the id range code path
	}
	for i := range ksuids[10:] {
		ksuids[i+10] = New()
	}
	for i := 1; i < len(ksuids); i += 4 {
		ksuids[i] = ksuids[i-1] // generate many dupes
	}

	miss := make(map[KSUID]struct{})
	uniq := make(map[KSUID]struct{})

	for _, id := range ksuids {
		miss[id] = struct{}{}
	}

	set := Compress(ksuids[:]...)

	for it := set.Iter(); it.Next(); {
		if _, dupe := uniq[it.KSUID]; dupe {
			t.Errorf("duplicate id found in compressed set: %s", it.KSUID)
		}
		uniq[it.KSUID] = struct{}{}
		delete(miss, it.KSUID)
	}

	if len(miss) != 0 {
		t.Error("some ids were not found in the compressed set:")
		for id := range miss {
			t.Log(id)
		}
	}
}

func testCompressedSetSingle(t *testing.T) {
	id := New()

	set := Compress(
		id, id, id, id, id, id, id, id, id, id,
		id, id, id, id, id, id, id, id, id, id,
		id, id, id, id, id, id, id, id, id, id,
		id, id, id, id, id, id, id, id, id, id,
	)

	n := 0

	for it := set.Iter(); it.Next(); {
		if n != 0 {
			t.Errorf("too many ids found in the compressed set: %s", it.KSUID)
		} else if id != it.KSUID {
			t.Errorf("invalid id found in the compressed set: %s != %s", it.KSUID, id)
		}
		n++
	}

	if n == 0 {
		t.Error("no ids were produced by the compressed set")
	}
}

func testCompressedSetNil(t *testing.T) {
	set := CompressedSet(nil)

	for it := set.Iter(); it.Next(); {
		t.Error("too many ids returned by the iterator of a nil compressed set: %s", it.KSUID)
	}
}

func testCompressedSetConcat(t *testing.T) {
	ksuids := [100]KSUID{}

	for i := range ksuids {
		ksuids[i] = New()
	}

	set := CompressedSet(nil)
	set = AppendCompressed(set, ksuids[:42]...)
	set = AppendCompressed(set, ksuids[42:64]...)
	set = AppendCompressed(set, ksuids[64:]...)

	for i, it := 0, set.Iter(); it.Next(); i++ {
		if ksuids[i] != it.KSUID {
			t.Errorf("invalid ID at index %d: %s != %s", i, ksuids[i], it.KSUID)
		}
	}
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
		n := 0
		for i := 0; i != b.N; i++ {
			ksuids2 = ksuids1
			buf = AppendCompressed(buf[:0], ksuids2[:]...)
			n = len(buf)
		}
		b.SetBytes(int64(n + len(ksuids2)))
	})

	b.Run("read", func(b *testing.B) {
		n := 0
		for i := 0; i != b.N; i++ {
			n = 0
			for it := set.Iter(); true; {
				if !it.Next() {
					n++
					break
				}
			}
		}
		b.SetBytes(int64((n * byteLength) + len(set)))
	})
}
