package service

import (
	"context"
	"github.com/MikkelvtK/solipull/internal/models"
	"log/slog"
	"sync"
)

const (
	Domain = "comicreleases.com"
)

type ScrapingObserver interface {
	models.ErrorObserver
	OnUrlFound(int)
	OnComicBookScraped(int)
}

type SolicitationService struct {
	scraper models.DataProvider
	repo    models.ComicBookRepository
	logger  *slog.Logger
	stats   *models.RunStats
}

func NewSolicitationService(p models.DataProvider, r models.ComicBookRepository, l *slog.Logger, s *models.RunStats) *SolicitationService {
	return &SolicitationService{
		scraper: p,
		logger:  l,
		stats:   s,
		repo:    r,
	}
}

func (s *SolicitationService) Sync(ctx context.Context) error {
	results := make(chan models.ComicBook, 100)
	errCh := make(chan error, 1)
	wg := &sync.WaitGroup{}
	url := "https://" + Domain + "/sitemap.xml"

	defer close(errCh)

	go s.bulkSave(ctx, results, errCh, wg)

	err := s.scraper.GetData(ctx, url, results)
	close(results)
	wg.Wait()

	if err != nil {
		return err
	}

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func (s *SolicitationService) bulkSave(ctx context.Context, res <-chan models.ComicBook, errCh chan<- error, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	cbs := make([]models.ComicBook, 100)

	for cb := range res {
		cbs = append(cbs, cb)

		if len(cbs) >= 100 {
			if err := s.repo.BulkSave(ctx, cbs); err != nil {
				errCh <- err
			}

			cbs = cbs[:0]
		}
	}

	if len(cbs) > 0 {
		if err := s.repo.BulkSave(ctx, cbs); err != nil {
			errCh <- err
		}
	}
}
