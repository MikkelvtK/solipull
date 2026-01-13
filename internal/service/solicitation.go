package service

import (
	"context"
	"fmt"
	"github.com/MikkelvtK/solipull/internal/models"
	"log/slog"
	"sync"
)

const (
	Domain = "comicreleases.com"
)

type SolicitationService struct {
	scraper models.DataProvider
	logger  *slog.Logger
	stats   *models.RunStats
}

func NewSolicitationService(p models.DataProvider, l *slog.Logger, s *models.RunStats) *SolicitationService {
	return &SolicitationService{
		scraper: p,
		logger:  l,
		stats:   s,
	}
}

func (s *SolicitationService) Sync(ctx context.Context) error {
	results := make(chan models.ComicBook, 100)
	wg := &sync.WaitGroup{}
	url := "https://" + Domain + "/sitemap.xml"

	go func() {
		wg.Add(1)
		defer wg.Done()

		for cb := range results {
			fmt.Println(cb)
		}
	}()

	if err := s.scraper.GetData(ctx, url, results); err != nil {
		return err
	}

	close(results)
	wg.Wait()
	return nil
}
