package strcache

import (
	"math/rand"
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
	time.Sleep(time.Nanosecond * 5000) // for change time.Now().UnixNano()
}

func checkCacheSkl(info string, t *testing.T, c *Cache, items ...string) {
	x := c.skl.Head()
	tmp := make([]string, len(items))
	for i := 0; x != nil; i++ {
		tmp[i] = x.Key
		x = x.Next()
	}
	for i := 0; i < len(items); i++ {
		if tmp[i] != items[i] {
			c.skl.Print()
			t.Fatalf("%v: invalid skiplist items, excepted %v instead of %v \r\n", info, items, tmp)
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
	checkCacheSkl("Simple set", t, c, "2")

	err = c.Set("3", str(3))
	if err != nil {
		t.Fatal(err)
	}
	checkCache(t, c, 5, 2)
	checkCacheSkl("Simple set", t, c, "2", "3")

	// Repeat set.
	err = c.Set("2", str(4))
	if err != nil {
		t.Fatal(err)
	}
	checkCache(t, c, 7, 2)
	checkCacheSkl("Repeat set", t, c, "3", "2")

	// Remove key "3" by insert a new key.
	err = c.Set("4", str(4))
	if err != nil {
		t.Fatal(err)
	}
	checkCache(t, c, 8, 2)
	checkCacheSkl(`Remove key "3" by insert a new key`, t, c, "2", "4")

	// Simple delete.
	err = c.Del("2")
	if err != nil {
		t.Fatal(err)
	}
	checkCache(t, c, 4, 1)
	checkCacheSkl("Simple delete", t, c, "4")

	// Delete non-existent key.
	err = c.Del("non-existent key")
	if err != ErrKeyNotFound {
		t.Fatalf("invalid err: %v", err)
	}
	checkCache(t, c, 4, 1)
	checkCacheSkl("Delete non-existent key", t, c, "4")

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
	checkCacheSkl("Delete non-existent key", t, c, "1")

	// Test Get change the accessTime and score.
	c.Del("1")
	c.Set("1", str(1))
	time.Sleep(time.Nanosecond * 5000)
	c.Set("2", str(2))
	time.Sleep(time.Nanosecond * 5000)
	c.Set("3", str(3))
	checkCache(t, c, 6, 3)
	checkCacheSkl("Test Get change the accessTime and score", t, c, "1", "2", "3")

	c.Get("2")
	checkCache(t, c, 6, 3)
	checkCacheSkl("Test Get change the accessTime and score", t, c, "1", "3", "2")

	c.Get("3")
	checkCache(t, c, 6, 3)
	checkCacheSkl("Test Get change the accessTime and score", t, c, "1", "2", "3")

	c.Get("1")
	checkCache(t, c, 6, 3)
	checkCacheSkl("Test Get change the accessTime and score", t, c, "2", "3", "1")

	// Test values.
	for i := 1; i <= 3; i++ {
		if v, _ := c.Get(strconv.Itoa(i)); v != str(i) {
			t.Fatalf("error value: %v expected %v got %v", i, str(i), v)
		}
	}

	// Test ByteSize.
	c = New(10 * KB)
	if c.maxsize != 10*1024 {
		t.Fatal("invalid ByteSize")
	}
	c.Set("3kb", str(3*1024))
	checkCache(t, c, 3*1024, 1)
	checkCacheSkl("Test ByteSize", t, c, "3kb")

	// Test value.
	c = New(1000)
	for i := 0; i < 1000; i++ {
		c.Set(strconv.Itoa(i), strconv.Itoa(i))
	}
}

func BenchmarkCache_Set(b *testing.B) {
	c := New(1000)
	x := str(10)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Set(strconv.Itoa(rand.Intn(1000)), x)
		}
	})
}

func BenchmarkCache_Get(b *testing.B) {
	c := New(1000)
	for i := 0; i < 1000; i++ {
		c.Set(strconv.Itoa(i), str(10))
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Get(strconv.Itoa(rand.Intn(1000)))
		}
	})
}

func BenchmarkCache_BigMemory_Set(b *testing.B) {
	c := New(1000 * GB)
	x := str(10)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Set(strconv.Itoa(rand.Intn(100000)), x)
		}
	})
}

func BenchmarkCache_BigMemory_Get(b *testing.B) {
	c := New(1000 * GB)
	for i := 0; i < 100000; i++ {
		c.Set(strconv.Itoa(i), str(10))
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Get(strconv.Itoa(rand.Intn(100000)))
		}
	})
}
