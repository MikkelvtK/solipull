# 🚀 Solipull

**Solipull** is a high-performance CLI tool designed for comic book enthusiasts to synchronize, browse, and manage comic book solicitations directly from the terminal.

> [!WARNING]  
> **Status: Work in Progress.** This tool is currently in active development. Basic features like background logging and collection management are currently being implemented.

---

## ✨ Current Features

- **Automated Sync**: Scrapes the [Comic Releases](https://www.comicreleases.com) sitemap and solicitation pages with regex-based precision using [Colly](https://github.com/gocolly/colly).
- **Interactive TUI**: A searchable, fuzzy-filtered list powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea).
- **Smart Persistence**: Robust [SQLite](https://sqlite.org) backend using **Upsert** logic to handle creating no duplicate entries.
- **Modern Architecture**: Built on **Clean Architecture** principles with a central dependency injection container.

## 🗺️ Roadmap

- [ ] **Structured Logging**: Implementation of `slog` JSON logging for background diagnostics.
- [ ] **Pull List Management**: Ability to "subscribe" to titles and track personal collections.
- [ ] **Export Formats**: Support for CSV and JSON data exports.
- [ ] **Homebrew Support**: Automated distribution via GoReleaser.

---

## 🛠 Installation (Development)

Requires Go 1.23+

```bash
# Clone the repository
git clone https://github.com/MikkelvtK/solipull.git
cd solipull

# Run directly
go run cmd/solipull/main.go --help
```

## 🏗 Architecture

Solipull follows below architecture design to ensure the different logic and infratstructure are clearly defined and separated:

- **`cmd/`**: The "Glue" layer. Contains the application entry point and manages the dependency graph via a central **Container**.
- **`internal/cli/`**: The "Transport" layer. Handles user interaction, command routing ([urfave/cli/v3](https://cli.urfave.org)), and the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI.
- **`internal/service/`**: The "Orchestration" layer. Defines domain use cases and manages the producer-consumer flow between the scraper and the repository.
- **`internal/scraper/`**: The "Ingestion" layer. Contains [Colly](https://github.com/gocolly/colly) collectors and regex-based extraction logic.
- **`internal/database/`**: The "Storage" layer. Implements the Repository pattern with DTO mapping to keep business models uncoupled from SQLite structures.

## 🛡 License

Distributed under the **Apache 2.0 License**.

Copyright (c) 2026 MikkelvtK
