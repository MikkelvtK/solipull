package app

import (
	"github.com/MikkelvtK/solipull/internal/models"
	"github.com/MikkelvtK/solipull/internal/scraper"
	"github.com/MikkelvtK/solipull/internal/service"
	"github.com/gocolly/colly/v2/queue"
	"log"
	"log/slog"
)

type Application struct {
	Serv *service.SolicitationService
}

func NewApplication(months, publishers []string) *Application {
	e := scraper.NewComicReleasesExtractor(months, publishers, slog.Default(), models.NewRunStats())
	q, err := queue.New(5, &queue.InMemoryQueueStorage{MaxSize: 10_000})
	if err != nil {
		log.Fatal(err)
	}

	col, err := scraper.NewCollector(service.Domain, 5)
	if err != nil {
		log.Fatal(err)
	}

	navCollector := col.Clone()
	solCollector := col.Clone()

	cfg := scraper.SConfig{
		Nav:    navCollector,
		Sol:    solCollector,
		Q:      q,
		Ex:     e,
		Logger: slog.Default(),
		Stats:  models.NewRunStats(),
	}

	s := scraper.NewComicReleasesScraper(&cfg)

	serv := service.NewSolicitationService(s, slog.Default(), models.NewRunStats())

	return &Application{
		Serv: serv,
	}
}
