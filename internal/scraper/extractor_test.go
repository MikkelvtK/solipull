package scraper

import (
	"github.com/MikkelvtK/solipull/internal/models"
	"log/slog"
	"reflect"
	"regexp"
	"testing"
	"time"
)

type MockNode struct {
	text string
	name string
}

func (m MockNode) Each(f func(HTMLNode)) {
	f(m)
}

func (m MockNode) Text() string {
	return m.text
}

func (m MockNode) NodeName() string {
	return m.name
}

func TestNewComicReleasesExtractor(t *testing.T) {
	tests := []struct {
		name string
		want ComicBookExtractor
	}{
		{
			name: "nil == no errors",
			want: &comicReleasesExtractor{
				rePublisher:   regexp.MustCompile(`(?i)/(?P<Pub>\w+)-[a-zA-Z]+-\d{4}-solicitations`),
				rePages:       regexp.MustCompile(`(?P<Pages>\d+)\s*(?i)(?:pages?|pgs?.?)`),
				rePrice:       regexp.MustCompile(`\$(\d+\.\d{2})`),
				reReleaseDate: regexp.MustCompile(`(?i)(\d{1,2}/\d{1,2}/\d{2,4})`),
				creatorParser: newCreatorParser([]string{"writer", "artist", "cover artist"}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewComicReleasesExtractor([]string{"march"}, []string{"dc"}, slog.Default(), &models.RunStats{}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewComicReleasesExtractor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_comicReleasesExtractor_Creators(t *testing.T) {
	type fields struct {
		creatorParser *creatorParser
	}
	type args struct {
		n HTMLNode
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []models.Creator
	}{
		{
			name: "nil == no errors",
			fields: fields{
				creatorParser: nil,
			},
			args: args{
				n: nil,
			},
			want: []models.Creator{},
		},
		{
			name: "default test case",
			fields: fields{
				creatorParser: newCreatorParser([]string{"writer", "artist", "cover artist"}),
			},
			args: args{
				n: MockNode{text: "writer: Tom King, Tom Taylor", name: "p"},
			},
			want: []models.Creator{{Name: "Tom King", Role: "writer"}, {Name: "Tom Taylor", Role: "writer"}},
		},
		{
			name: "applies proper casing to the names",
			fields: fields{
				creatorParser: newCreatorParser([]string{"writer", "artist", "cover artist"}),
			},
			args: args{
				n: MockNode{text: "writer: TOM KING, TOM TAYLOR", name: "p"},
			},
			want: []models.Creator{{Name: "Tom King", Role: "writer"}, {Name: "Tom Taylor", Role: "writer"}},
		},
		{
			name: "skips unknown roles",
			fields: fields{
				creatorParser: newCreatorParser([]string{"writer", "artist", "cover artist"}),
			},
			args: args{
				n: MockNode{text: "variant cover artist: DAN MORA"},
			},
			want: []models.Creator{},
		},
		{
			name: "skips br nodes",
			fields: fields{
				creatorParser: newCreatorParser([]string{"writer", "artist", "cover artist"}),
			},
			args: args{
				n: MockNode{text: "writer: TOM KING, TOM TAYLOR", name: "br"},
			},
			want: []models.Creator{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &comicReleasesExtractor{
				creatorParser: tt.fields.creatorParser,
			}
			if got := c.Creators(tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Creators() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_comicReleasesExtractor_Issue(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nil == no errors",
			args: args{
				s: "",
			},
			want: "",
		},
		{
			name: "default test case",
			args: args{
				s: "Batman #1",
			},
			want: "1",
		},
		{
			name: "handles no issue number found",
			args: args{
				s: "Batman vol 1",
			},
			want: "",
		},
		{
			name: "handles special issues",
			args: args{
				s: "Batman #1 FACSIMILE EDITION",
			},
			want: "1 Facsimile Edition",
		},
		{
			name: "handles bad issue number",
			args: args{
				s: "Batman #1.5",
			},
			want: "1.5",
		},
		{
			name: "handles incoherent issue number",
			args: args{
				s: "Batman #1 FACSIMILE EDITION               dsgsdgsdgsdgdsgsdgg",
			},
			want: "1 Facsimile Edition               Dsgsdgsdgsdgdsgsdgg",
		},
		{
			name: "handles extra whitespace",
			args: args{
				s: "Batman #1 FACSIMILE EDITION                    ",
			},
			want: "1 Facsimile Edition",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &comicReleasesExtractor{}
			if got := c.Issue(tt.args.s); got != tt.want {
				t.Errorf("Issue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_comicReleasesExtractor_Pages(t *testing.T) {
	type fields struct {
		rePages *regexp.Regexp
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "nil == no errors",
			fields: fields{
				rePages: nil,
			},
			args: args{
				s: "",
			},
			want: "",
		},
		{
			name: "default test case",
			fields: fields{
				rePages: regexp.MustCompile(`(?P<Pages>\d+)\s*(?i)(?:pages?|pgs?.?)`),
			},
			args: args{
				s: "32 pages",
			},
			want: "32",
		},
		{
			name: "handles alternative page number",
			fields: fields{
				rePages: regexp.MustCompile(`(?P<Pages>\d+)\s*(?i)(?:pages?|pgs?.?)`),
			},
			args: args{
				s: "800 PGS.",
			},
			want: "800",
		},
		{
			name: "handles no pages found",
			fields: fields{
				rePages: regexp.MustCompile(`(?P<Pages>\d+)\s*(?i)(?:pages?|pgs?.?)`),
			},
			args: args{
				s: "$4.99",
			},
			want: "",
		},
		{
			name: "handles bad page number",
			fields: fields{
				rePages: regexp.MustCompile(`(?P<pages>\d+)\s*(?i)(?:pages?|pgs?.?)`),
			},
			args: args{
				s: "32 pages",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &comicReleasesExtractor{
				rePages: tt.fields.rePages,
			}
			if got := c.Pages(tt.args.s); got != tt.want {
				t.Errorf("Pages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_comicReleasesExtractor_Price(t *testing.T) {
	type fields struct {
		rePrice *regexp.Regexp
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "nil == no errors",
			fields: fields{
				rePrice: nil,
			},
			args: args{
				s: "",
			},
			want: "",
		},
		{
			name: "default test case",
			fields: fields{
				rePrice: regexp.MustCompile(`\$(\d+\.\d{2})`),
			},
			args: args{
				s: "$4.99",
			},
			want: "$4.99",
		},
		{
			name: "handles no price found",
			fields: fields{
				rePrice: regexp.MustCompile(`(?P<Price>\d+\.\d{2})`),
			},
			args: args{
				s: "32 pages",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &comicReleasesExtractor{
				rePrice: tt.fields.rePrice,
			}
			if got := c.Price(tt.args.s); got != tt.want {
				t.Errorf("Price() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_comicReleasesExtractor_Publisher(t *testing.T) {
	type fields struct {
		rePublisher *regexp.Regexp
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "nil == no errors",
			fields: fields{
				rePublisher: nil,
			},
			args: args{
				s: "",
			},
			want: "",
		},
		{
			name: "default test case",
			fields: fields{
				rePublisher: regexp.MustCompile(`(?i)/(?P<Pub>\w+)-[a-zA-Z]+-\d{4}-solicitations`),
			},
			args: args{
				s: "/dc-march-2026-solicitations",
			},
			want: "dc",
		},
		{
			name: "handles case insensitivity",
			fields: fields{
				rePublisher: regexp.MustCompile(`(?i)/(?P<Pub>\w+)-[a-zA-Z]+-\d{4}-solicitations`),
			},
			args: args{
				s: "/DC-MARCH-2026-SOLICITATIONS",
			},
			want: "dc",
		},
		{
			name: "handles any length",
			fields: fields{
				rePublisher: regexp.MustCompile(`(?i)/(?P<Pub>\w+)-[a-zA-Z]+-\d{4}-solicitations`),
			},
			args: args{
				s: "/IMAGE-MARCH-2026-SOLICITATIONS",
			},
			want: "image",
		},
		{
			name: "handles no pub found found",
			fields: fields{
				rePublisher: regexp.MustCompile(`(?i)/(?P<Pub>\w+)-[a-zA-Z]+-\d{4}-solicitations`),
			},
			args: args{
				s: "/MARCH-2026-SOLICITATIONS",
			},
			want: "",
		},
		{
			name: "handles no pub found found",
			fields: fields{
				rePublisher: regexp.MustCompile(`(?i)/(?P<pub>\w+)-[a-zA-Z]+-\d{4}-solicitations`),
			},
			args: args{
				s: "/DC-MARCH-2026-SOLICITATIONS",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &comicReleasesExtractor{
				rePublisher: tt.fields.rePublisher,
			}
			if got := c.Publisher(tt.args.s); got != tt.want {
				t.Errorf("Publisher() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_comicReleasesExtractor_ReleaseDate(t *testing.T) {
	type fields struct {
		reReleaseDate *regexp.Regexp
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   time.Time
	}{
		{
			name: "nil == no errors",
			fields: fields{
				reReleaseDate: nil,
			},
			args: args{
				s: "",
			},
			want: time.Time{},
		},
		{
			name: "default test case",
			fields: fields{
				reReleaseDate: regexp.MustCompile(`(?i)(\d{1,2}/\d{1,2}/\d{2,4})`),
			},
			args: args{
				s: "On Sale: 3/12/26",
			},
			want: time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "handles alternative date",
			fields: fields{
				reReleaseDate: regexp.MustCompile(`(?i)(\d{1,2}/\d{1,2}/\d{2,4})`),
			},
			args: args{
				s: "ON SALE: 3/12/26",
			},
			want: time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "handles wrong date",
			fields: fields{
				reReleaseDate: regexp.MustCompile(`(?i)(\d{1,2}/\d{1,2}/\d{2,4})`),
			},
			args: args{
				s: "12 Mar, 2026",
			},
			want: time.Time{},
		},
		{
			name: "handles no date found",
			fields: fields{
				reReleaseDate: regexp.MustCompile(`(?i)(\d{1,2}/\d{1,2}/\d{2,4})`),
			},
			args: args{
				s: "32 pages",
			},
			want: time.Time{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &comicReleasesExtractor{
				reReleaseDate: tt.fields.reReleaseDate,
			}
			if got := c.ReleaseDate(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReleaseDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_comicReleasesExtractor_Title(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nil == no errors",
			args: args{
				s: "",
			},
			want: "",
		},
		{
			name: "default test case",
			args: args{
				s: "Batman #1",
			},
			want: "Batman",
		},
		{
			name: "handles case insensitivity",
			args: args{
				s: "BATMAN #1",
			},
			want: "Batman",
		},
		{
			name: "handles any no split",
			args: args{
				s: "BATMAN",
			},
			want: "Batman",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &comicReleasesExtractor{}
			if got := c.Title(tt.args.s); got != tt.want {
				t.Errorf("Title() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newCreatorParser(t *testing.T) {
	type args struct {
		roles []string
	}
	tests := []struct {
		name string
		args args
		want *creatorParser
	}{
		{
			name: "nill == no errors",
			args: args{
				roles: nil,
			},
			want: &creatorParser{},
		},
		{
			name: "default test case",
			args: args{
				roles: []string{"writer"},
			},
			want: &creatorParser{roles: []string{"writer"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newCreatorParser(tt.args.roles); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newCreatorParser() = %v, want %v", got, tt.want)
			}
		})
	}
}
