package manticore

import "sort"

// make the array sortable
type uint64Slice []uint64

func (p uint64Slice) Len() int           { return len(p) }
func (p uint64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p uint64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type vlbcomp []byte

func (s *vlbcomp) compress(val uint64) int {
	if val == 0 {
		return 0
	}
	for true {
		char := byte(val & 0x7f)
		val >>= 7
		if val == 0 {
			*s = append(*s, char)
			return 1
		}
		*s = append(*s, char|0x80)
	}
	return 0
}

func buildUvarRequest(name string, values []uint64) func(*apibuf) {
	return func(buf *apibuf) {

		// make copy of original values
		cp := uint64Slice(make([]uint64, len(values)))
		copy(cp, values)

		// prepare delta-encoded vbl compressed blob

		// sort given blob
		sort.Sort(cp)

		var output vlbcomp

		// delta-encoding
		prev := uint64(0)
		nvalues := int(0) // will calculate N of non-zero deltas (i.e. uniq values)
		for i := 0; i < len(cp); i++ {
			// compress + uniq
			nvalues += output.compress(cp[i] - prev)
			prev = cp[i]
		}

		buf.putString(name)
		buf.putLen(nvalues)
		buf.putLen(len(output))
		buf.putBytes(output)
	}
}
