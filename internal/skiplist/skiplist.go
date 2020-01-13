package skiplist

import (
	"math/rand"
)

const (
	defaultMaxLevel = 32
	defaultP        = 0.25
	defaultRandSeed = 0
)

var uts = newUniqueTS()

type SkipList struct {
	header *Node
	tail   *Node
	length int64
	level  int

	maxLevel int
	p        float64
	rnd      *rand.Rand
}

func New(maxlevel int, p float64, randseed int64) *SkipList {
	if maxlevel <= 1 || p <= 0 {
		panic("maxLevel must greater than 1, p must greater than 0")
	}

	s := &SkipList{
		header:   nil,
		tail:     nil,
		length:   0,
		level:    1,
		maxLevel: maxlevel,
		p:        p,
		rnd:      rand.New(rand.NewSource(randseed)),
	}

	s.header = newNode(s.maxLevel, 0, "", "", 0)

	for j := 0; j < s.maxLevel; j++ {
		s.header.next[j] = nil
	}

	s.header.pre = nil
	s.tail = nil
	return s
}

func NewDefault() *SkipList {
	return New(defaultMaxLevel, defaultP, defaultRandSeed)
}

func (s *SkipList) Len() int64 {
	return s.length
}

func (s *SkipList) Head() *Node {
	if s.length == 0 {
		return nil
	}
	return s.header.next[0]
}

func (s *SkipList) Tail() *Node { return s.tail }

func (s *SkipList) randomLevel() int {
	level := 1

	for s.rnd.Float64() < s.p {
		level += 1
	}

	if level > s.maxLevel {
		return s.maxLevel
	}
	return level
}

func (s *SkipList) Insert(score int64, key, val string) *Node {
	update := make([]*Node, s.maxLevel)
	lastAccess := uts.Now()

	// Search the insert location, also calculates `update` and `rank`.
	// The search process is begin from the highest level's header.
	for i, n := s.level-1, s.header; i >= 0; i-- {
		for n.next[i] != nil &&
			(n.next[i].Score < score ||
				(n.next[i].Score == score && n.next[i].LastAccess < lastAccess)) {
			n = n.next[i]
		}
		update[i] = n
	}

	// Make a random level for the insert node.
	level := s.randomLevel()
	// If the insert process will create new levels, we need to
	// update the `rank` and `update`.
	if level > s.level {
		for i := s.level; i < level; i++ {
			// s.header is the only node in every levels,
			// since it doesn't has tail, so its pan is
			// the length of skiplist.
			update[i] = s.header
		}
		s.level = level
	}

	// Insert the new node into levels. Keep in mind here, we just
	// insert it to `node.levels`(it only includes next pointer).
	// But the level[0] is actually a doubled link list.
	n := newNode(level, score, key, val, lastAccess)
	for i := 0; i < level; i++ {
		n.next[i] = update[i].next[i]
		update[i].next[i] = n
	}

	// Update new node's pre.
	if update[0] != s.header {
		n.pre = update[0]
	}

	// Update new node's next's pre, Because the levels[0] is
	// doubled link list. But if new node's next is NIL, we
	// need to change s.tail to the new node.
	if n.next[0] != nil {
		n.next[0].pre = n
	} else {
		s.tail = n
	}

	s.length++

	return n
}

func (s *SkipList) Delete(score int64, key string, lastAccess int64) bool {
	update := make([]*Node, s.maxLevel)
	n := s.header
	for i := s.level - 1; i >= 0; i-- {
		for n.next[i] != nil &&
			(n.next[i].Score < score ||
				(n.next[i].Score == score && n.next[i].LastAccess < lastAccess)) {
			n = n.next[i]
		}
		update[i] = n
	}

	n = n.next[0]
	if n != nil && n.Score == score && n.Key == key {
		s.delete(n, update)
		return true
	}
	return false
}

func (s *SkipList) delete(n *Node, update []*Node) {
	// Delete node and update span for all levels.
	for i := 0; i < s.level; i++ {
		if update[i].next[i] == n {
			update[i].next[i] = n.next[i]
		}
	}

	// Update n.pre if possible.
	if n.next[0] != nil {
		n.next[0].pre = n.pre
	} else {
		s.tail = n.pre
	}

	// Update skiplist.level if some levels only includes header.
	for s.level > 1 && s.header.next[s.level-1] == nil {
		s.level--
	}
	s.length--
}

// We don't need to update it's value here.
// The lastAccess must be modified to Now().
// So we update `newscore` and `newLastAccess` in this function.
func (s *SkipList) Update(score int64, key string, lastAccess int64, newscore int64) *Node {
	update := make([]*Node, s.maxLevel)
	n := s.header
	for i := s.level - 1; i >= 0; i-- {
		for n.next[i] != nil &&
			(n.next[i].Score < score ||
				(n.next[i].Score == score && n.next[i].LastAccess < lastAccess)) {
			n = n.next[i]
		}
		update[i] = n
	}

	n = n.next[0]

	if (n.pre == nil || n.pre.Score < newscore) &&
		(n.next[0] == nil || n.next[0].Score > newscore) {
		n.Score = newscore
		n.LastAccess = uts.Now()
		return n
	}
	s.delete(n, update)
	return s.Insert(newscore, n.Key, n.Val)
}

// For debug only.
func (s *SkipList) Print() {
	for i := s.level - 1; i >= 0; i-- {
		print(i, " ")
		x := s.header.next[i]

		for x != nil {
			print("[val:", x.Val, " key:", x.Key, " score:", x.Score, " lastAccess:", x.LastAccess, "] -> ")
			x = x.next[i]
		}
		print("nil")
		print("\r\n")
	}
	print("\r\n")
}

type Node struct {
	Key        string
	Val        string
	Score      int64
	LastAccess int64
	pre        *Node
	next      []*Node
}

func newNode(level int, score int64, key, val string, lastAccess int64) *Node {
	return &Node{
		Key:        key,
		Val:        val,
		Score:      score,
		LastAccess: lastAccess,
		pre:        nil,
		next:      make([]*Node, level),
	}
}

func (n *Node) Next() *Node { return n.next[0] }
