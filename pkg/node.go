package snap

import "fmt"

type Cx struct {
	Chs  string
	Body string
}
type Mapper func(Cx) fmt.Stringer

type MapperNode struct {
	Chars                 string
	Fn                    Mapper
	Children              []*MapperNode
	maxLenOfCharsThisPath int
}

func (n *MapperNode) Find(chs string) Mapper {
	if len(chs) == 0 {
		return n.Fn
	}

	ch := chs[0:1]
	chs = chs[1:]

	for _, child := range n.Children {
		if child.Chars == ch {
			return child.Find(chs)
		}
	}

	return nil
}

func (n *MapperNode) Insert(chs string, fn Mapper) {
	n._insert(chs, fn, len(chs))
}

func (n *MapperNode) _insert(chs string, fn Mapper, originalLenOfChars int) {
	if n.maxLenOfCharsThisPath < originalLenOfChars {
		n.maxLenOfCharsThisPath = originalLenOfChars
	}
	if len(chs) == 0 {
		n.Fn = fn
		return
	}

	ch := chs[0:1]
	chs = chs[1:]

	for _, child := range n.Children {
		if child.Chars == ch {
			child.maxLenOfCharsThisPath = originalLenOfChars
			child._insert(chs, fn, originalLenOfChars)
			return
		}
	}

	child := &MapperNode{
		Chars:                 ch,
		maxLenOfCharsThisPath: originalLenOfChars,
	}
	child._insert(chs, fn, originalLenOfChars)
	n.Children = append(n.Children, child)

}

func (n *MapperNode) isWordOfLenPossible(len int) bool {
	if n.maxLenOfCharsThisPath == 0 {
		return false
	}

	return len <= n.maxLenOfCharsThisPath
}
