package scraper

import (
	"github.com/PuerkitoBio/goquery"
	"reflect"
	"testing"
)

func TestGoQueryNode_Each(t *testing.T) {
	type fields struct {
		sel *goquery.Selection
	}
	type args struct {
		fn func(HTMLNode)
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "nil == no errors",
			fields: fields{
				sel: nil,
			},
			args: args{
				fn: func(n HTMLNode) {},
			},
		},
		{
			name: "default test case",
			fields: fields{
				sel: &goquery.Selection{},
			},
			args: args{
				fn: func(n HTMLNode) {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := GoQueryNode{
				sel: tt.fields.sel,
			}
			g.Each(tt.args.fn)
		})
	}
}

func TestGoQueryNode_NodeName(t *testing.T) {
	type fields struct {
		sel *goquery.Selection
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "nil == no errors",
			fields: fields{
				sel: nil,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := GoQueryNode{
				sel: tt.fields.sel,
			}
			if got := g.NodeName(); got != tt.want {
				t.Errorf("NodeName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoQueryNode_Text(t *testing.T) {
	type fields struct {
		sel *goquery.Selection
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "nil == no errors",
			fields: fields{
				sel: nil,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := GoQueryNode{
				sel: tt.fields.sel,
			}
			if got := g.Text(); got != tt.want {
				t.Errorf("Text() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	type args struct {
		sel *goquery.Selection
	}
	tests := []struct {
		name string
		args args
		want GoQueryNode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Wrap(tt.args.sel); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Wrap() = %v, want %v", got, tt.want)
			}
		})
	}
}
