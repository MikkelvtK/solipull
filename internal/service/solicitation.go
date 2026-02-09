package service

import (
	"context"
	"github.com/MikkelvtK/solipull/internal/models"
	"sync"
)

const (
	Domain = "comicreleases.com"
)

type DataProvider interface {
	GetData(ctx context.Context, url string, results chan<- models.ComicBook, observer ScrapingObserver) error
	SetInputs(months, publishers []string) error
}

type ScrapingObserver interface {
	models.ErrorObserver
	OnStart()
	OnUrlFound(n int)
	OnNavigationComplete()
	OnComicBookScraped(n int)
	OnScrapingComplete()
}

type SolicitationService struct {
	scraper DataProvider
	repo    models.ComicBookRepository
}

func NewSolicitationService(p DataProvider, r models.ComicBookRepository) *SolicitationService {
	return &SolicitationService{
		scraper: p,
		repo:    r,
	}
}

func (s *SolicitationService) Sync(ctx context.Context, observer ScrapingObserver, months, publishers []string) error {
	results := make(chan models.ComicBook, 100)
	errCh := make(chan error, 1)
	wg := &sync.WaitGroup{}
	url := "https://" + Domain + "/sitemap.xml"

	defer close(errCh)

	if err := s.scraper.SetInputs(months, publishers); err != nil {
		return err
	}

	wg.Add(1)
	go s.bulkSave(ctx, results, errCh, wg)

	err := s.scraper.GetData(ctx, url, results, observer)
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

func (s *SolicitationService) View(ctx context.Context, months, publishers []string) ([]models.ComicBook, error) {
	// TODO: implement to get comic books by months and publishers
	return s.repo.GetAll(ctx)
}

func (s *SolicitationService) bulkSave(ctx context.Context, res <-chan models.ComicBook, errCh chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	cbs := make([]models.ComicBook, 0, 100)

	for cb := range res {
		cbs = append(cbs, cb)

		if len(cbs) >= 100 {
			if err := s.repo.BulkSave(ctx, cbs); err != nil {
				errCh <- err
				return
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
