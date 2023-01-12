package gee

import "strings"

type node struct {
	pattern  string  //待匹配路由  例如/p/:xxx
	part     string  //路由中的一部分
	children []*node //子节点 例如[xxx,xxx,xxx]
	isWild   bool    //是否精准匹配   part中有:和*时为true
}

//第一个匹配成功的节点，用于插入
func (n *node) matchChild(part string) *node {
	//如果是精准匹配，就直接返回
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

//所有匹配成功的节点  用于查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	//如果是精准匹配，就直接返回
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

//insert 插入节点
func (n *node) insert(pattern string, parts []string, height int) {
	//parts为空，说明已经到了最后一个节点
	if len(parts) == height {
		//如果pattern为空，说明已经存在了
		n.pattern = pattern
		return
	}
	part := parts[height]
	child := n.matchChild(part)
	//如果没有匹配到，就新建一个节点
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

//search 查找节点
func (n *node) search(parts []string, height int) *node {
	//parts为空，说明已经到了最后一个节点
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}
	part := parts[height]
	children := n.matchChildren(part)
	//如果没有匹配到，就返回nil
	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}
