package wf

import "strings"

type node struct {
	part     string
	isWild   bool
	children []*node
	//叶子节点
	path     string
	handlers HandlersChain
}

func (n *node) insert(path string, parts []string, handlers HandlersChain, height int) {
	if len(parts) == height {
		n.path = path
		n.handlers = handlers
		return
	}

	part := parts[height]
	child := n.matchChild(part, false)
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	child.insert(path, parts, handlers, height+1)
}

func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || (len(n.part) != 0 && n.part[0] == '*') {
		if n.path == "" {
			return nil
		}
		return n
	}

	var ret *node = nil
	part := parts[height]
	children := n.matchChildren(part)
	for _, child := range children {
		res := child.search(parts, height+1)
		if res != nil {
			pathParts := parsePath(res.path)
			if strings.Contains(pathParts[height], "*") {
				if ret == nil {
					ret = res
				}
				continue
			} else if strings.Contains(pathParts[height], ":") {
				ret = res
				continue
			} else {
				ret = res
				break
			}
		}
	}
	return ret
}

func (n *node) matchChild(part string, wild bool) (n_ret *node) {
	for _, child := range n.children {
		if child.part == part {
			return child
		} else if wild && child.isWild {
			n_ret = child
		}
	}
	return n_ret
}

func (n *node) matchChildren(part string) (ns []*node) {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			ns = append(ns, child)
		}
	}
	return ns
}
