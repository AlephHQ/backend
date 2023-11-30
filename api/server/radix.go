package server

import (
	"net/http"
	"strings"
)

type edge struct {
	target *node
	label  string
}

type node struct {
	handler http.Handler
	edges   []*edge
}

func newnode(h http.Handler) *node {
	return &node{
		edges:   make([]*edge, 0),
		handler: h,
	}
}

type radix struct {
	root *node
}

func sharedPrefix(s1, s2 string) string {
	i := 0
	prefix := make([]byte, 0)
	for i < len(s1) && i < len(s2) && s1[i] == s2[i] {
		prefix = append(prefix, s1[i])
		i++
	}

	return string(prefix)
}

func (r *radix) insert(pattern string, h http.Handler) {
	if r.root == nil {
		r.root = newnode(nil)
	}

	current := r.root
	found := 0

	inserted := false
	for !inserted {
		var traversable *edge

		i := 0
		prefix := ""
		for traversable == nil && i < len(current.edges) {
			if prefix = sharedPrefix(pattern[found:], current.edges[i].label); prefix != "" {
				traversable = current.edges[i]
			}

			i++
		}

		found += len(prefix)
		if traversable == nil {
			if found == len(pattern) {
				// this handles the case where our pattern
				// already exists in the tree and ends at a node
				// with a nil handler.
				//
				// case example:
				// 	1. insert(classical, h1)
				// 	2. insert(clasmical, h2)
				//	3. insert(clas, h3)
				current.handler = h
			} else {
				current.edges = append(
					current.edges,
					&edge{
						label:  pattern[found:],
						target: newnode(h),
					},
				)
			}

			inserted = true
		} else {
			if len(prefix) == len(traversable.label) {
				current = traversable.target
			} else if len(prefix) < len(traversable.label) {
				newTarget := newnode(nil)
				newTarget.edges = append(
					newTarget.edges,
					&edge{
						target: traversable.target,
						label:  traversable.label[len(prefix):],
					},
				)

				// this avoids adding edges with the empty string
				// as a label when we've already found all the elements
				// in our pattern
				if found < len(pattern) {
					newTarget.edges = append(
						newTarget.edges,
						&edge{
							target: newnode(h),
							label:  pattern[found:],
						},
					)
				} else if found == len(pattern) {
					newTarget.handler = h
				}

				traversable.target = newTarget
				traversable.label = prefix
				inserted = true
			}
		}
	}
}

func (r *radix) find(pattern string) (http.Handler, string) {
	if r.root == nil {
		return nil, pattern
	}

	current := r.root
	found := 0

	for {
		var next *edge

		i := 0
		for next == nil && i < len(current.edges) {
			if strings.HasPrefix(pattern[found:], current.edges[i].label) {
				next = current.edges[i]
			}

			i++
		}

		if next == nil {
			return nil, pattern
		} else {
			found += len(next.label)
			current = next.target

			if found == len(pattern) {
				return next.target.handler, pattern
			}
		}
	}
}
