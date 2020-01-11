package skiplist

import (
	"strconv"
	"testing"
)

func printSkipList(l *SkipList) {
	for i := l.level - 1; i >= 0; i-- {
		print(i, " ")
		x := l.header.levels[i]

		for x != nil {
			print("[val:", x.Val, " score:", x.Score, "] -> ")
			x = x.levels[i]
		}
		print("nil")
		print("\r\n")
	}
	print("\r\n")
}

func TestNew(t *testing.T) {
	l := New(10, 0.3, 2)
	l.Insert(12, "12", "12")
	l.Insert(13, "13", "13")
	n0 := l.Insert(14, "14", "14")
	l.Insert(15, "15", "15")
	l.Insert(11, "11", "11")
	l.Insert(16, "16", "16")
	n1 := l.Insert(17, "17", "17")
	printSkipList(l)
	l.Delete(17, "17", n1.LastAccess)
	printSkipList(l)
	l.Update(14, "14", n0.LastAccess, 16)
	printSkipList(l)
}

func BenchmarkSkipList_Insert(b *testing.B) {
	l := NewDefault()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Insert(int64(i), "", strconv.Itoa(i))
	}
}
