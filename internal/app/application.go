package app

import (
	"fmt"
	"github.com/MikkelvtK/solipull/internal/database"
	"github.com/MikkelvtK/solipull/internal/database/sqlite"
	"github.com/MikkelvtK/solipull/internal/models"
	"github.com/MikkelvtK/solipull/internal/scraper"
	"github.com/MikkelvtK/solipull/internal/service"
	"github.com/gocolly/colly/v2/queue"
	"log"
	"log/slog"
	"os"
)

type Application struct {
	Serv *service.SolicitationService
	repo models.ComicBookRepository
}

func NewApplication() *Application {
	cfgDir, _ := os.UserConfigDir()

	db := database.MustOpen(cfgDir+"/solipull/solipull.db", "sqlite")
	fmt.Println(cfgDir)
	repo := sqlite.NewComicBookRepository(db)

	e := scraper.NewComicReleasesExtractor(slog.Default())
	q, err := queue.New(5, &queue.InMemoryQueueStorage{MaxSize: 10_000})
	if err != nil {
		log.Fatal(err)
	}

	navCollector, _ := scraper.NewCollector(service.Domain, 5)
	solCollector, _ := scraper.NewCollector(service.Domain, 5)

	cfg := scraper.SConfig{
		Nav:    navCollector,
		Sol:    solCollector,
		Q:      q,
		Ex:     e,
		Logger: slog.Default(),
	}

	s, _ := scraper.NewComicReleasesScraper(&cfg)

	serv := service.NewSolicitationService(s, repo)

	return &Application{
		Serv: serv,
		repo: repo,
	}
}
