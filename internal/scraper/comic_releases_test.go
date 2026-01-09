package scraper

import (
	"errors"
	"github.com/MikkelvtK/solipull/internal/cache"
	"github.com/MikkelvtK/solipull/internal/models"
	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/xmlquery"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"reflect"
	"strings"
	"testing"
	"time"
)

var (
	OneUltramanCollection = `<div class="wp-block-columns is-layout-flex wp-container-core-columns-is-layout-28f84493 wp-block-columns-is-layout-flex">
<div class="wp-block-column is-layout-flow wp-block-column-is-layout-flow" style="flex-basis:66.66%">
<p>ULTRAMAN OMNIBUS HC 60TH ANNIVERSARY COVER</p>
<p>
Writer(s): MAT GROOM, KYLE HIGGINS &amp;MORE<br>
Pencils: FRANCESCO MANNA, DAVIDE TINTO &amp;MORE<br>
Cover Artist(s): TSUBURAYA PRODUCTIONS &amp;PEACH MOMOKO<br>
800 PGS./Rated T+ …$125.00<br>
ISBN: 9781302967482<br>Trim size: 7-1/4 x 10-7/8
</p>
<p>
For the first time ever, Marvel Comics collects its 21st-entury saga of Ultraman, the Japanese pop-cultural icon, together with never-before-collected Ultraman stories from the 1990s!<br>Marvel Comics presents a thrilling reimagination of the pop-culture phenomenon that is Ultraman! With writing and art from superstars around the world, this single volume is sure to be a one-of-a-kind classic! Discover how a young man named Shin Hayata merged with a mysterious warrior from beyond the stars and gained the incredible ability to become the towering Ultraman – for only three minutes! Working alongside the United Science Patrol, Ultraman protects the Earth from monstrous interdimensional invaders known as Kaiju – and is set on a multiversal collision course with the Avengers…and Galactus! Plus: never-before-collected classic comic-book adventures of Ultraman from the 1990s!
</p>
<p>Collecting THE RISE OF ULTRAMAN (2020) #1-5, THE TRIALS OF ULTRAMAN (2021) #1-5, ULTRAMAN: THE MYSTERY OF ULTRASEVEN (2022) #1-5, ULTRAMAN X THE AVENGERS (2024) #1-4, THE FALL OF ULTRAMAN (2026) #1, ULTRAMAN (1993) #1-3 and ULTRAMAN (1994) #-1 and #1-4.</p>
<p>
<strong>
<span style="text-decoration: underline;">ULTRAMAN OMNIBUS HC PEACH MOMOKO COVER [DM ONLY]</span>
</strong>
<br>
800 PGS./Rated T+ …$125.00<br>
ISBN: 9781302967499<br>Trim size: 7-1/4 x 10-7/8
</p>
</div>
</div>`

	detectiveComics = `<div class="wp-block-columns is-layout-flex wp-container-core-columns-is-layout-28f84493 wp-block-columns-is-layout-flex">
<div class="wp-block-column is-layout-flow wp-block-column-is-layout-flow" style="flex-basis:66.66%">
<p>DETECTIVE COMICS #1107</p>
<p>
Writer(s): TOM TAYLOR<br>
Artist(s): PETE WOODS<br>
Cover Artist(s): MIKEL JANIN<br>
Variant Covers:<br>
Variant covers by ESAD RIBIC and OZGUR YILDIRIM<br>
Corner Box variant by JORGE JIMENEZ<br>$4.99 US | 32 pages | Variant $5.99 US (card stock)
</p>
<p>On Sale: 3/25/26</p>
<p>The Dark Knight Detective is hot on the trail of an abducted teenager with a mysterious past, but he has found himself at a dead end. In a rare moment of desperation, Batman teams up with Black Canary and Gotham City’s newest resident, Green Arrow, to investigate a case with unexpected and terrifying implications for Bruce, Dinah, and Oliver’s shared history. Will this trio be enough to rescue this girl and unravel the mystery of her past? Find out in this thrilling new storyline! </p>
</div>
</div>`

	validXml = `
<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type="text/xsl" href="https://www.comicreleases.com/sitemap.xsl"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd">
<url>
<loc>https://www.comicreleases.com/2025/12/marvel-march-2026-solicitations/</loc>
<lastmod>2026-01-01T22:27:21+00:00</lastmod>
</url>
</urlset>
`
	invalidXml = `
<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type="text/xsl" href="https://www.comicreleases.com/sitemap.xsl"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd">
<url>
<loc>%%marvel-march-2026-solicitations/</loc>
<lastmod>2026-01-01T22:27:21+00:00</lastmod>
</url>
</urlset>
`
)

func setupHtmlElement(html string, pub string, t *testing.T) *colly.HTMLElement {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Errorf("Error loading document: %s", err)
	}

	el := &colly.HTMLElement{
		DOM: doc.Find("div.wp-block-columns").First(),
		Request: &colly.Request{
			Ctx: colly.NewContext(),
		},
	}

	el.Request.Ctx.Put("publisher", pub)
	return el
}

func setupXmlElement(xml string, t *testing.T) *colly.XMLElement {
	t.Helper()

	c := colly.NewCollector()
	var capturedRequest *colly.Request
	c.OnRequest(func(r *colly.Request) {
		capturedRequest = r
		r.Abort()
	})
	_ = c.Visit("https://example.com")

	resp := &colly.Response{StatusCode: 200, Body: []byte(xml)}
	doc, _ := xmlquery.Parse(strings.NewReader(xml))
	xmlNode := xmlquery.FindOne(doc, "//loc")
	xmlElem := colly.NewXMLElementFromXMLNode(resp, xmlNode)
	xmlElem.Request = capturedRequest
	return xmlElem
}

func setupCrDetailParser(months []string, publishers []string, t *testing.T) *crDetailParser {
	t.Helper()

	cac := cache.NewCache()
	reg := newCrComicRegex(months, publishers)
	ex := newExtractor(reg)
	return newCrDetailParser(cac, ex)
}

func setupCrListParser(months []string, publishers []string, t *testing.T) *crListParser {
	t.Helper()

	reg := newCrComicRegex(months, publishers)
	qu, err := queue.New(5, &queue.InMemoryQueueStorage{MaxSize: 10_000})
	if err != nil {
		t.Errorf("Error loading queue: %s", err)
	}

	return newCrListParser(reg, qu)
}

func TestNewComicReleasesScraper(t *testing.T) {
	c := cache.NewCache()
	months := []string{"march"}
	publishers := []string{"dc"}

	scraper, err := NewComicReleasesScraper(c, months, publishers)

	if err != nil {
		t.Fatalf("Failed to create scraper: %v", err)
	}

	expectedURL := "https://comicreleases.com/sitemap.xml"
	if scraper.url != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, scraper.url)
	}

	if scraper.listCollector == nil || scraper.detailCollector == nil {
		t.Error("Collectors were not initialized")
	}

	if scraper.queue == nil {
		t.Error("Queue was not initialized")
	}
}

func Test_crDetailParser_Parse(t *testing.T) {
	type args struct {
		e *colly.HTMLElement
		c []models.ComicBook
	}
	tests := []struct {
		name    string
		fields  *crDetailParser
		args    args
		wantLen int
	}{
		{
			name:   "Successfully parsed marvel",
			fields: setupCrDetailParser([]string{"march"}, []string{"marvel"}, t),
			args: args{
				e: setupHtmlElement(OneUltramanCollection, "marvel", t),
				c: []models.ComicBook{
					{
						Title: "Ultraman Omnibus Hc 60Th Anniversary Cover",
						Pages: "800 PGS.",
						Price: "$125.00",
						Creators: func() map[string][]string {
							c := make(map[string][]string)
							c["cover artist"] = []string{"Tsuburaya Productions & Peach Momoko"}
							c["writer"] = []string{"Mat Groom", "Kyle Higgins", "More"}
							return c
						}(),
						Publisher:   "marvel",
						ReleaseDate: time.Time{},
					},
				},
			},
			wantLen: 1,
		},
		{
			name:   "Successfully parsed dc",
			fields: setupCrDetailParser([]string{"march"}, []string{"dc", "marvel"}, t),
			args: args{
				e: setupHtmlElement(detectiveComics, "dc", t),
				c: []models.ComicBook{
					{
						Title:     "Detective Comics",
						Issue:     "1107",
						Pages:     "32 pages",
						Price:     "$5.99",
						Publisher: "dc",
						Creators: func() map[string][]string {
							c := make(map[string][]string)
							c["artist"] = []string{"Pete Woods"}
							c["writer"] = []string{"Tom Taylor"}
							c["cover artist"] = []string{"Mikel Janin"}
							return c
						}(),
						ReleaseDate: time.Date(2026, time.March, 25, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			wantLen: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &crDetailParser{
				c: tt.fields.c,
				e: tt.fields.e,
			}
			p.Parse(tt.args.e)

			cbs, err := p.c.GetAll()
			if err != nil {
				t.Errorf("Failed to get all comic books: %v", err)
				return
			}
			if len(cbs) != tt.wantLen {
				t.Errorf("Expected 1 comic book, got %d", len(cbs))
			}
			for i, cb := range cbs {
				if cb.Title != tt.args.c[i].Title {
					t.Errorf("Expected title %s, got %s", tt.args.c[i].Title, cb.Title)
				}
				if cb.Pages != tt.args.c[i].Pages {
					t.Errorf("Expected pages %s, got %s", tt.args.c[i].Pages, cb.Pages)
				}
				if cb.Publisher != tt.args.c[i].Publisher {
					t.Errorf("Expected publisher %s, got %s", tt.args.c[i].Publisher, cb.Publisher)
				}
				if cb.ReleaseDate != tt.args.c[i].ReleaseDate {
					t.Errorf("Expected releaseDate %s, got %s", tt.args.c[i].ReleaseDate, cb.ReleaseDate)
				}
				if !reflect.DeepEqual(cb.Creators, tt.args.c[i].Creators) {
					t.Errorf("Expected creators %v, got %v", tt.args.c[i].Creators, cb.Creators)
				}
			}
		})
	}
}

func Test_crDetailParser_Selector(t *testing.T) {
	p := setupCrDetailParser([]string{"march"}, []string{"marvel"}, t)
	want := "div.wp-block-columns"
	t.Run("Selector returns proper string", func(t *testing.T) {
		if got := p.Selector(); got != want {
			t.Errorf("Selector() = %v, want %v", got, want)
		}
	})
}

func Test_crIssue(t *testing.T) {
	type args struct {
		s   string
		in1 *crComicRegex
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Finds issue for a single",
			args: args{
				s:   "DETECTIVE COMICS #1107",
				in1: &crComicRegex{},
			},
			want:    "1107",
			wantErr: false,
		},
		{
			name: "Finds no issue for collected comic",
			args: args{
				s:   "ULTRAMAN OMNIBUS HC 60TH ANNIVERSARY COVER",
				in1: &crComicRegex{},
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Finds issue for special editions",
			args: args{
				s:   "DETECTIVE COMICS #475 FACSIMILE EDITION",
				in1: &crComicRegex{},
			},
			want:    "475 Facsimile Edition",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := crIssue(tt.args.s, tt.args.in1)
			if (err != nil) != tt.wantErr {
				t.Errorf("crIssue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("crIssue() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_crListParser_Parse(t *testing.T) {
	type fields struct {
		reg *crComicRegex
		q   *queue.Queue
	}
	type args struct {
		e *colly.XMLElement
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantLen int
		wantErr bool
	}{
		{
			name: "Successfully find url",
			fields: fields{
				reg: newCrComicRegex([]string{"march"}, []string{"dc", "marvel"}),
				q: func() *queue.Queue {
					q, err := queue.New(5, &queue.InMemoryQueueStorage{MaxSize: 10})
					if err != nil {
						t.Errorf("Failed to create queue: %v", err)
					}
					return q
				}(),
			},
			args: args{
				e: setupXmlElement(validXml, t),
			},
			wantLen: 1,
		},
		{
			name: "Don't find url",
			fields: fields{
				reg: newCrComicRegex([]string{"march"}, []string{"dc"}),
				q: func() *queue.Queue {
					q, err := queue.New(5, &queue.InMemoryQueueStorage{MaxSize: 10})
					if err != nil {
						t.Errorf("Failed to create queue: %v", err)
					}
					return q
				}(),
			},
			args: args{
				e: setupXmlElement(validXml, t),
			},
			wantLen: 0,
		},
		{
			name: "invalid url",
			fields: fields{
				reg: newCrComicRegex([]string{"march"}, []string{"dc", "marvel"}),
				q: func() *queue.Queue {
					q, err := queue.New(5, &queue.InMemoryQueueStorage{MaxSize: 10})
					if err != nil {
						t.Errorf("Failed to create queue: %v", err)
					}
					return q
				}(),
			},
			args: args{
				e: setupXmlElement(invalidXml, t),
			},
			wantLen: 0,
		},
		{
			name: "queue is full",
			fields: fields{
				reg: newCrComicRegex([]string{"march"}, []string{"dc", "marvel"}),
				q: func() *queue.Queue {
					q, err := queue.New(5, &queue.InMemoryQueueStorage{MaxSize: 1})
					x := setupXmlElement(validXml, t)
					_ = q.AddRequest(x.Request)
					if err != nil {
						t.Errorf("Failed to create queue: %v", err)
					}
					return q
				}(),
			},
			args: args{
				e: setupXmlElement(validXml, t),
			},
			wantLen: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &crListParser{
				reg: tt.fields.reg,
				q:   tt.fields.q,
			}
			p.Parse(tt.args.e)

			got, err := p.q.Size()
			if err != nil {
				t.Errorf("crListParser.queue.Size() error = %v", err)
				return
			}

			if got != tt.wantLen {
				t.Errorf("crListParser.queue.Size() = %v, want %v", got, tt.wantLen)
			}
		})
	}
}

func Test_crListParser_Selector(t *testing.T) {
	p := setupCrListParser([]string{"march"}, []string{"marvel"}, t)
	want := "//loc"
	t.Run("Selector returns proper string", func(t *testing.T) {
		if got := p.Selector(); got != want {
			t.Errorf("Selector() = %v, want %v", got, want)
		}
	})
}

func Test_crPages(t *testing.T) {
	type args struct {
		s     string
		regex *crComicRegex
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "nil == no errors",
			args: args{
				s:     "",
				regex: nil,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "pages in format '32 pages' extracteds",
			args: args{
				s:     "$4.99 US | 32 pages | Variant $5.99 US (card stock)",
				regex: newCrComicRegex([]string{"march"}, []string{"dc", "marvel"}),
			},
			want:    "32 pages",
			wantErr: false,
		},
		{
			name: "pages in format '32 pages' extracted",
			args: args{
				s:     "800 PGS./Rated T+ …$125.00",
				regex: newCrComicRegex([]string{"march"}, []string{"dc", "marvel"}),
			},
			want:    "800 PGS.",
			wantErr: false,
		},
		{
			name: "no pages found",
			args: args{
				s:     "/Rated T+ …$125.00",
				regex: newCrComicRegex([]string{"march"}, []string{"dc", "marvel"}),
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := crPages(tt.args.s, tt.args.regex)
			if (err != nil) != tt.wantErr {
				t.Errorf("crPages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("crPages() got = %v, want %v", got, tt.want)
			}
		})
	}
}

//func Test_crParseTime(t *testing.T) {
//    type args struct {
//        s string
//        e *extractor
//    }
//    tests := []struct {
//        name string
//        args args
//        want time.Time
//    }{
//        // TODO: Add test cases.
//    }
//    for _, tt := range tests {
//        t.Run(tt.name, func(t *testing.T) {
//            if got := crParseTime(tt.args.s, tt.args.e); !reflect.DeepEqual(got, tt.want) {
//                t.Errorf("crParseTime() = %v, want %v", got, tt.want)
//            }
//        })
//    }
//}

func Test_crPrice(t *testing.T) {
	type args struct {
		s     string
		regex *crComicRegex
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "nil == no errors",
			args: args{
				s:     "",
				regex: nil,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "extracted price",
			args: args{
				s:     "$4.99 US | 32 pages | Variant $5.99 US (card stock)",
				regex: newCrComicRegex([]string{"march"}, []string{"dc", "marvel"}),
			},
			want:    "$4.99",
			wantErr: false,
		},
		{
			name: "extracted marvel price",
			args: args{
				s:     "800 PGS./Rated T+ …$125.00",
				regex: newCrComicRegex([]string{"march"}, []string{"dc", "marvel"}),
			},
			want:    "$125.00",
			wantErr: false,
		},
		{
			name: "no price found",
			args: args{
				s:     "/Rated T+ …",
				regex: newCrComicRegex([]string{"march"}, []string{"dc", "marvel"}),
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := crPrice(tt.args.s, tt.args.regex)
			if (err != nil) != tt.wantErr {
				t.Errorf("crPrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("crPrice() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_crReleaseDate(t *testing.T) {
	type args struct {
		s     string
		regex *crComicRegex
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "nil == no errors",
			args: args{
				s:     "",
				regex: nil,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "extracted releaseDate",
			args: args{
				s:     "On Sale: 4/8/26",
				regex: newCrComicRegex([]string{"march"}, []string{"dc", "marvel"}),
			},
			want:    "4/8/26",
			wantErr: false,
		},
		{
			name: "extracted releaseDate alternative",
			args: args{
				s:     "ON-SALE 03/25/26",
				regex: newCrComicRegex([]string{"march"}, []string{"dc", "marvel"}),
			},
			want:    "03/25/26",
			wantErr: false,
		},
		{
			name: "no release date found",
			args: args{
				s:     "",
				regex: newCrComicRegex([]string{"march"}, []string{"dc", "marvel"}),
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := crReleaseDate(tt.args.s, tt.args.regex)
			if (err != nil) != tt.wantErr {
				t.Errorf("crReleaseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("crReleaseDate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_crTitle(t *testing.T) {
	type args struct {
		s   string
		in1 *crComicRegex
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "nil == no errors",
			args: args{
				s:   "",
				in1: nil,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "extracted title with issue upper",
			args: args{
				s:   "BATMAN #1",
				in1: nil,
			},
			want:    "Batman",
			wantErr: false,
		},
		{
			name: "extracted title with issue lower",
			args: args{
				s:   "batman #1",
				in1: nil,
			},
			want:    "Batman",
			wantErr: false,
		},
		{
			name: "extracted title without issue",
			args: args{
				s:   "batman",
				in1: nil,
			},
			want:    "Batman",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := crTitle(tt.args.s, tt.args.in1)
			if (err != nil) != tt.wantErr {
				t.Errorf("crTitle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("crTitle() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractor_extract(t *testing.T) {
	type fields struct {
		regex *crComicRegex
	}
	type args struct {
		s           string
		extractFunc func(string, *crComicRegex) (string, error)
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
				regex: nil,
			},
			args: args{
				s: "",
				extractFunc: func(string, *crComicRegex) (string, error) {
					return "", nil
				},
			},
			want: "",
		},
		{
			name: "func returns error",
			fields: fields{
				regex: nil,
			},
			args: args{
				s: "",
				extractFunc: func(string, *crComicRegex) (string, error) {
					return "", errors.New("error")
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &extractor{
				regex: tt.fields.regex,
			}
			if got := e.extract(tt.args.s, tt.args.extractFunc); got != tt.want {
				t.Errorf("extract() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractor_extractCreators(t *testing.T) {
	type fields struct {
		regex *crComicRegex
	}
	type args struct {
		s *goquery.Selection
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string][]string
	}{
		{
			name: "nil == no errors",
			fields: fields{
				regex: nil,
			},
			args: args{
				s: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &extractor{
				regex: tt.fields.regex,
			}
			if got := e.extractCreators(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractCreators() = %v, want %v", got, tt.want)
			}
		})
	}
}

//func Test_generateUrlRegex(t *testing.T) {
//    type args struct {
//        months     []string
//        publishers []string
//    }
//    tests := []struct {
//        name string
//        args args
//        want string
//    }{
//        // TODO: Add test cases.
//    }
//    for _, tt := range tests {
//        t.Run(tt.name, func(t *testing.T) {
//            if got := generateUrlRegex(tt.args.months, tt.args.publishers); got != tt.want {
//                t.Errorf("generateUrlRegex() = %v, want %v", got, tt.want)
//            }
//        })
//    }
//}
//
//func Test_newCrComicRegex(t *testing.T) {
//    type args struct {
//        months     []string
//        publishers []string
//    }
//    tests := []struct {
//        name string
//        args args
//        want *crComicRegex
//    }{
//        // TODO: Add test cases.
//    }
//    for _, tt := range tests {
//        t.Run(tt.name, func(t *testing.T) {
//            if got := newCrComicRegex(tt.args.months, tt.args.publishers); !reflect.DeepEqual(got, tt.want) {
//                t.Errorf("newCrComicRegex() = %v, want %v", got, tt.want)
//            }
//        })
//    }
//}
//
//func Test_newCrDetailParser(t *testing.T) {
//    type args struct {
//        c *cache.Cache
//        e *extractor
//    }
//    tests := []struct {
//        name string
//        args args
//        want *crDetailParser
//    }{
//        // TODO: Add test cases.
//    }
//    for _, tt := range tests {
//        t.Run(tt.name, func(t *testing.T) {
//            if got := newCrDetailParser(tt.args.c, tt.args.e); !reflect.DeepEqual(got, tt.want) {
//                t.Errorf("newCrDetailParser() = %v, want %v", got, tt.want)
//            }
//        })
//    }
//}
//
//func Test_newCrListParser(t *testing.T) {
//    type args struct {
//        reg *crComicRegex
//        q   *queue.Queue
//    }
//    tests := []struct {
//        name string
//        args args
//        want *crListParser
//    }{
//        // TODO: Add test cases.
//    }
//    for _, tt := range tests {
//        t.Run(tt.name, func(t *testing.T) {
//            if got := newCrListParser(tt.args.reg, tt.args.q); !reflect.DeepEqual(got, tt.want) {
//                t.Errorf("newCrListParser() = %v, want %v", got, tt.want)
//            }
//        })
//    }
//}
//
//func Test_newExtractor(t *testing.T) {
//    type args struct {
//        regex *crComicRegex
//    }
//    tests := []struct {
//        name string
//        args args
//        want *extractor
//    }{
//        // TODO: Add test cases.
//    }
//    for _, tt := range tests {
//        t.Run(tt.name, func(t *testing.T) {
//            if got := newExtractor(tt.args.regex); !reflect.DeepEqual(got, tt.want) {
//                t.Errorf("newExtractor() = %v, want %v", got, tt.want)
//            }
//        })
//    }
//}
