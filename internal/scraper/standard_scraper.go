package scraper

import (
	"fmt"
	"github.com/MikkelvtK/pul/internal/models"
	"sync"
	"time"
)

type StandardScraper struct {
	baseEndpoint string
	callsMade    int
}

func NewLeagueOfComicGeeksScraper() *StandardScraper {
	return &StandardScraper{
		baseEndpoint: "https://leagueofcomicgeeks.com/solicitations",
		callsMade:    0,
	}
}

func (s *StandardScraper) Init() error {

	return nil
}

func (s *StandardScraper) Scrape(month string, publishers []string) ([]models.ComicBook, error) {

	return nil, nil
}

func (s *StandardScraper) worker(id int, job string, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("worker %d starting job %s\n", id, job)
	time.Sleep(time.Second)
	results <- id
}
