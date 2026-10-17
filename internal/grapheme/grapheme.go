package grapheme

import "github.com/rockorager/go-uucode"

// Next returns the end of the grapheme cluster starting at offset in str.
func Next(str string, offset int) int {
	it := uucode.NewGraphemeIterator(str[offset:])
	cluster, _ := it.Next()
	return offset + cluster.End
}

// Previous returns the start of the grapheme cluster ending at offset in str.
func Previous(str string, offset int) int {
	start := 0
	for end := Next(str, 0); end < offset; end = Next(str, end) {
		start = end
	}
	return start
}
