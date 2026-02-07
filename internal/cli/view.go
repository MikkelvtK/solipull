package cli

import (
	"context"
	"fmt"
	"github.com/MikkelvtK/solipull/internal/models"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v3"
	"slices"
	"strings"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func (c *CLI) view() *cli.Command {
	return &cli.Command{
		Name:  "view",
		Usage: "View and export comic book solicitations.",
		Description: "Displays solicitation data in a formatted and interactive table by default. Supports JSON and " +
			"CSV exports via flags for use in scripts and external tools.",
		Action: func(ctx context.Context, cmd *cli.Command) error {

			cbs, err := c.solService.View(ctx, []string{}, []string{})
			if err != nil {
				return err
			}

			m := newModel(cbs)
			if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
				return err
			}
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "publisher",
				Aliases: []string{"p"},
				Usage:   "Publishers to view",
			},
			&cli.StringSliceFlag{
				Name:    "month",
				Aliases: []string{"m"},
				Usage:   "Months to view",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Output as JSON",
			},
			&cli.BoolFlag{
				Name:  "csv",
				Usage: "Output as CSV",
			},
			&cli.BoolFlag{
				Name:  "csv--no-header",
				Usage: "Removes header names from csv output",
			},
		},
	}
}

type comicItem struct {
	cb *models.ComicBook
}

func (i comicItem) Title() string {
	return fmt.Sprintf("%s #%s", i.cb.Title, i.cb.Issue)
}

func (i comicItem) Description() string {
	pub := strings.ToUpper(i.cb.Publisher)
	dt := i.cb.ReleaseDate.Format("Jan 02")

	return fmt.Sprintf("[%s] | %s | %s", pub, dt, i.cb.Price)
}

func (i comicItem) FilterValue() string {
	c := slices.Collect(func(yield func(string) bool) {
		for _, cr := range i.cb.Creators {
			if !yield(cr.Name) {
				return
			}
		}
	})

	s := strings.ToLower(fmt.Sprintf("%s %s %s %s %s",
		i.cb.Title,
		i.cb.Issue,
		i.cb.ReleaseDate.Month(),
		i.cb.Publisher,
		strings.Join(c, " ")))

	return s
}

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func newModel(cbs []models.ComicBook) *model {
	items := slices.Collect(func(yield func(list.Item) bool) {
		for _, cb := range cbs {
			item := comicItem{cb: &cb}

			if !yield(item) {
				return
			}
		}
	})

	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "Comic Book Solicitations"

	return &m
}
