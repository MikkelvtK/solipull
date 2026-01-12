package scraper

import "github.com/PuerkitoBio/goquery"

type GoQueryNode struct {
	sel *goquery.Selection
}

func Wrap(sel *goquery.Selection) GoQueryNode {
	return GoQueryNode{sel: sel}
}

func (g GoQueryNode) Each(fn func(HTMLNode)) {
	g.sel.Contents().Each(func(_ int, s *goquery.Selection) {
		fn(GoQueryNode{sel: s})
	})
}

func (g GoQueryNode) Text() string {
	return g.sel.Text()
}

func (g GoQueryNode) NodeName() string {
	return goquery.NodeName(g.sel)
}
