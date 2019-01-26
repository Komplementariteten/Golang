package TreeTest

import "fmt"

const (
	RED   = 0
	BLACK = 1
)

type node struct {
	left, right, parent *node
	color               int
	Value               TreeValue
	Key                 KeyType
}

type TreeValue interface{}

type KeyType interface {
	LessThan(interface{}) bool
}

type Tree struct {
	root *node
	size int
}

func NewTree() *Tree {
	return &Tree{}
}

//Empty check whether the rbtree is empty
func (t *Tree) Empty() bool {
	if t.root == nil {
		return true
	}
	return false
}

//Iterator create the rbtree's iterator that points to the minmum node
func (t *Tree) Iterator() *node {
	return minimum(t.root)
}

//Size return the size of the rbtree
func (t *Tree) Size() int {
	return t.size
}

//Clear destroy the rbtree
func (t *Tree) Clear() {
	t.root = nil
	t.size = 0
}

//Next return the node's successor as an iterator
func (n *node) Next() *node {
	return successor(n)
}

func (n *node) preorder() {
	fmt.Printf("(%v %v)", n.Key, n.Value)
	if n.parent == nil {
		fmt.Printf("nil")
	} else {
		fmt.Printf("whose parent is %v", n.parent.Key)
	}
	if n.color == RED {
		fmt.Println(" and color RED")
	} else {
		fmt.Println(" and color BLACK")
	}
	if n.left != nil {
		fmt.Printf("%v's left child is ", n.Key)
		n.left.preorder()
	}
	if n.right != nil {
		fmt.Printf("%v's right child is ", n.Key)
		n.right.preorder()
	}
}

//successor return the successor of the node
func successor(x *node) *node {
	if x.right != nil {
		return minimum(x.right)
	}
	y := x.parent
	for y != nil && x == y.right {
		x = y
		y = x.parent
	}
	return y
}

//getColor get color of the node
func getColor(n *node) int {
	if n == nil {
		return BLACK
	}
	return n.color
}

//minimum find the minimum node of subtree n.
func minimum(n *node) *node {
	for n.left != nil {
		n = n.left
	}
	return n
}

//maximum find the maximum node of subtree n.
func maximum(n *node) *node {
	for n.right != nil {
		n = n.right
	}
	return n
}
