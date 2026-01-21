package scraper

import "github.com/PuerkitoBio/goquery"

type GoQueryNode struct {
	sel *goquery.Selection
}

func Wrap(sel *goquery.Selection) GoQueryNode {
	return GoQueryNode{sel: sel}
}

func (g GoQueryNode) Each(fn func(HTMLNode)) {
	if g.sel == nil {
		return
	}

	g.sel.Contents().Each(func(_ int, s *goquery.Selection) {
		fn(GoQueryNode{sel: s})
	})
}

func (g GoQueryNode) Text() string {
	if g.sel == nil {
		return ""
	}

	return g.sel.Text()
}

func (g GoQueryNode) NodeName() string {
	if g.sel == nil {
		return ""
	}

	return goquery.NodeName(g.sel)
}
