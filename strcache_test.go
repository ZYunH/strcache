package strcache

import (
	"strconv"
	"testing"
	"time"
)

// str returns a string which length equal to the input x.
func str(x int) string {
	s := make([]byte, 0)
	for i := 0; i < x; i++ {
		s = append(s, 65)
	}
	return string(s)
}

func checkCache(t *testing.T, c *Cache, nowsize int, items int) {
	if c.nowsize != ByteSize(nowsize) {
		t.Fatalf("invalid nowsize, excepted %v instead of %v", nowsize, c.nowsize)
	}
	if len(c.m) != items || int(c.skl.Len()) != items {
		t.Fatalf("invalid length of items, excepted <len(m):%v len(skl):%v>"+
			"instead of <len(m):%v len(skl):%v>", items, items, len(c.m), c.skl.Len())
	}
}

func checkCacheSkl(t *testing.T, c *Cache, items ...string) {
	x := c.skl.Head()
	tmp := make([]string, len(items))
	for i := 0; i < len(items); i++ {
		tmp = append(tmp, x.Key)
		if x.Key != items[0] {
			t.Fatalf("invalid skiplist items, excepted %v instead of %v \r\nfull:%v", items[:i+1], tmp, items)
		}
	}
}

func TestCache_All(t *testing.T) {
	c := New(10)
	err := c.Set("1", str(11))
	if err != ErrNoEnoughSpace {
		t.Fatal("no size limit")
	}

	// Simple set.
	err = c.Set("2", str(2))
	if err != nil {
		t.Fatal(err)
	}
	checkCache(t, c, 2, 1)
	checkCacheSkl(t, c, "2")

	err = c.Set("3", str(3))
	if err != nil {
		t.Fatal(err)
	}
	checkCache(t, c, 5, 2)
	checkCacheSkl(t, c, "2", "3")

	// Repeat set.
	err = c.Set("2", str(4))
	if err != nil {
		t.Fatal(err)
	}
	checkCache(t, c, 7, 2)
	checkCacheSkl(t, c, "3", "2")

	// Remove key "3" by insert a new key.
	err = c.Set("4", str(4))
	if err != nil {
		t.Fatal(err)
	}
	checkCache(t, c, 8, 2)
	checkCacheSkl(t, c, "2", "4")

	// Simple delete.
	err = c.Del("2")
	if err != nil {
		t.Fatal(err)
	}
	checkCache(t, c, 4, 1)
	checkCacheSkl(t, c, "4")

	// Delete non-existent key.
	err = c.Del("non-existent key")
	if err != ErrKeyNotFound {
		t.Fatalf("invalid err: %v", err)
	}
	checkCache(t, c, 4, 1)
	checkCacheSkl(t, c, "4")

	// Delete last key.
	err = c.Del("4")
	if err != nil {
		t.Fatal(err)
	}
	checkCache(t, c, 0, 0)

	// Simple Get.
	if err := c.Set("1", str(1)); err != nil {
		t.Fatal(err)
	}
	s, err := c.Get("1")
	if err != nil {
		t.Fatal(err)
	}
	if s != str(1) {
		t.Fatal("invalid value")
	}

	// Get non-existent key.
	_, err = c.Get("non-existent key")
	if err != ErrKeyNotFound {
		t.Fatalf("invalid err: %v", err)
	}
	checkCache(t, c, 1, 1)
	checkCacheSkl(t, c, "1")

	// Test Get change the accessTime and score.
	c.Del("1")
	c.Set("1", str(1))
	time.Sleep(time.Nanosecond * 100)
	c.Set("2", str(2))
	time.Sleep(time.Nanosecond * 100)
	c.Set("3", str(3))
	checkCache(t, c, 6, 3)
	checkCacheSkl(t, c, "1", "2", "3")

	c.Get("2")
	checkCache(t, c, 6, 3)
	checkCacheSkl(t, c, "1", "3", "2")

	c.Get("3")
	checkCache(t, c, 6, 3)
	checkCacheSkl(t, c, "1", "2", "3")

	c.Get("1")
	checkCache(t, c, 6, 3)
	checkCacheSkl(t, c, "2", "3", "1")

	// Test ByteSize.
	c = New(10 * KB)
	if c.maxsize != 10*1024 {
		t.Fatal("invalid ByteSize")
	}
	c.Set("3kb", str(3*1024))
	checkCache(t, c, 3*1024, 1)
	checkCacheSkl(t, c, "3kb")
}

func BenchmarkCache_Set(b *testing.B) {
	c := New(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(strconv.Itoa(i), "1111111111")
	}
}

func BenchmarkCache_Get(b *testing.B) {
	c := New(1000)
	for i := 0; i < 100; i++ {
		c.Set(strconv.Itoa(i), "1111111111")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(strconv.Itoa(i % 100))
	}
}

func BenchmarkCache_BigMemory_Set(b *testing.B) {
	c := New(1000 * MB)
	x := str(int(3 * KB))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(strconv.Itoa(i), x)
	}
}

func BenchmarkCache_BigMemory_Get(b *testing.B) {
	c := New(1000 * GB)
	for i := 0; i < 10000; i++ {
		c.Set(strconv.Itoa(i), str(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(strconv.Itoa(i % 10000))
	}
}
