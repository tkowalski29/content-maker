# Content Maker - Content Creation Toolkit

A comprehensive toolkit for automating the content creation process: from SEO audits, through image generation, to metadata extraction. The project contains 4 CLI applications written in Go and integration with Claude Code.

## 📋 Table of Contents

- [CLI Applications](#-cli-applications)
- [Requirements](#-requirements)
- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Usage](#-usage)
- [Tests](#-tests)
- [Documentation](#-documentation)
- [Project Structure](#-project-structure)

## 🛠️ CLI Applications

The project contains 4 independent CLI applications:

### 1. `brief` - SEO Audits and Competitor Analysis

Automatic SEO audit generation, TOP 10 competitor analysis, keyword suggestions, and content gaps.

**Main features**:
- Fetch SERP from SearXNG (TOP 20 results)
- Keyword suggestions (Google, Bing, DuckDuckGo)
- HTML content extraction (3 methods: Jina Reader, Netlify, direct)
- Competitor analysis and outline generation
- Identify content gaps and entities
- Markdown report ready for article writing

**Usage example**:
```bash
# Build
go build -o brief ./cmd/brief/

# SERP
./brief serp -query="best beaches in spain" -lang=en

# Suggestions
./brief suggest -query="beaches" -provider=google

# Fetch HTML
./brief fetch -url="https://example.com" -method=jina

# Analyze audits
./brief analyze -audits=.spec/t1/brief/all_audits.json
```

📖 **Documentation**: [cmd/brief/README.md](cmd/brief/README.md)

### 2. `colab` - Google Colab Automation

Control Google Colab notebooks through Playwright. Automatic launching on T4 GPU, keep-alive, webhook notifications.

**Main features**:
- Automatic login and Chrome session management
- Launch notebooks on T4 GPU
- Listen for gradio.live URLs
- Send webhooks with Gradio links
- Keep-alive mechanism (scroll + events)
- Debug snapshots (screenshots, HTML, logs)

**Usage example**:
```bash
# Build
go build -o colab ./cmd/colab/

# Run
./colab -env ./env.json -user test1
./colab -env /path/to/config/env.json -user test1
```

📖 **Documentation**: [cmd/colab/README.md](cmd/colab/README.md)

### 3. `extract` - Metadata Extraction from Markdown

Extract image placeholders and metadata from Markdown files (front-matter, FAQ).

**Main features**:
- Extract `{{IMAGE_X}}` placeholders with metadata
- Parse YAML front-matter
- Extract FAQ sections to schema.org
- Automatic slug generation
- Export to JSON

**Usage example**:
```bash
# Build
go build -o extract ./cmd/extract/

# Extract images
./extract images -input content/article.md -output output/images.json

# Extract front-matter
./extract frontmatter -input content/article.md -output output/cms.json
```

📖 **Documentation**: [cmd/extract/README.md](cmd/extract/README.md)

### 4. `gradio` - Image Generation

Automatic image generation through Gradio API backend.

**Main features**:
- Batch generation from JSON file
- Support for different aspect ratios (16:9, 4:3, 1:1, etc.)
- Event stream monitoring (SSE)
- Save PNG images + JSON metadata
- Retry mechanism

**Usage example**:
```bash
# Build
go build -o gradio ./cmd/gradio/

# Generate images
./gradio -input requests.json -gradio_url https://your-instance.gradio.live
```

📖 **Documentation**: [cmd/gradio/README.md](cmd/gradio/README.md)

## 📦 Requirements

### Required:
- **Go** 1.21+ (for compilation)
- **jq** (for JSON processing in hooks)
- **bash** (for verification hooks)

### Optional (depending on application):
- **Claude Code** (for running `/brief` command)
- **Playwright + Chrome/Chromium** (for `colab` application)
- **Gradio Backend** (for `gradio` application)

### Installing dependencies:

```bash
# macOS
brew install jq

# Linux
sudo apt-get install jq

# Playwright (for colab)
go run github.com/playwright-community/playwright-go/cmd/playwright install chromium
```

## 🚀 Installation

### Option 1: Build all applications

```bash
# Clone repository
git clone git@github.com:tkowalski29/content-maker.git
cd content_maker

# Build all applications
go build -o brief ./cmd/brief/
go build -o colab ./cmd/colab/
go build -o extract ./cmd/extract/
go build -o gradio ./cmd/gradio/

# Optionally move to /usr/local/bin
sudo mv brief colab extract gradio /usr/local/bin/
```

### Option 2: Build selected application

```bash
# Only brief
go build -o brief ./cmd/brief/

# Only colab
go build -o colab ./cmd/colab/

# Only extract
go build -o extract ./cmd/extract/

# Only gradio
go build -o gradio ./cmd/gradio/
```

### Option 3: Using Makefile

```bash
# Build brief
make build

# Clean
make clean

# Tests
make test
```

## ⚡ Quick Start

### 1. Brief - SEO Audit

```bash
# In Claude Code CLI
/brief t1

# Or directly with application
./brief serp -query="best beaches in spain"
```

### 2. Colab - Launch notebook

```bash
# Prepare env.json
cp cmd/colab/env.json.example env.json
# Edit env.json with your data

# Run
./colab -env ./env.json -user test1
```

### 3. Extract - Extract from Markdown

```bash
# Extract image placeholders
./extract images -input article.md -output images.json

# Extract metadata
./extract frontmatter -input article.md -output cms.json
```

### 4. Gradio - Generate images

```bash
# Prepare requests file
cat > requests.json << EOF
{
  "folder_name": "output/images",
  "items": [
    {
      "id": "HERO",
      "alt": "Hero image",
      "prompt": "A futuristic city at sunrise",
      "style": "photorealistic",
      "aspect_ratio": "16:9",
      "position_in_article": 120,
      "platform": "web"
    }
  ]
}
EOF

# Generate
./gradio -input requests.json -gradio_url https://your-instance.gradio.live
```

## 📖 Usage

### `/brief` Command (Claude Code)

The `/brief` command in Claude Code performs a full SEO audit in 7 phases:

1. **Initialization** - Read task.md, create directories
2. **Data gathering** - SERP + suggestions (parallel)
3. **Generate outline** - Intent analysis, article structure
4. **Fetch TOP 10** - Download competitor HTML
5. **Per-URL audits** - Detailed analysis of each URL
6. **Aggregation** - Content gaps, entities, unique angles
7. **Markdown report** - Final brief

**Example**:

```bash
# 1. Create task directory
mkdir -p .spec/t2
echo "Write an article about best restaurants in Krakow" > .spec/t2/task.md

# 2. Run in Claude Code
/brief t2

# 3. Result
cat .spec/t2/output/out_brief_t2.md
```

### Workflow: Brief → Colab → Gradio → Extract

Typical workflow for full automation:

```bash
# 1. Generate SEO brief
/brief my-article

# 2. Launch Colab (Flux1 model)
./colab -env ./env.json -user test1
# You'll receive webhook with URL to Gradio

# 3. Generate images through Gradio
./gradio -input images_requests.json -gradio_url https://xyz.gradio.live

# 4. Write Markdown article (manually or via AI)
# ...

# 5. Extract metadata
./extract images -input article.md -output images.json
./extract frontmatter -input article.md -output cms.json
```

## 🧪 Tests

### Test coverage: 87.3%

```bash
# All tests
make test

# Tests with verbose output
make test-verbose

# Coverage report
make test-coverage

# HTML report (opens in browser)
make test-coverage-html
```

### Testing components

```bash
# Test SERP
make run-serp QUERY='best beaches in spain'

# Test suggestions
make run-suggest QUERY='beaches' PROVIDER=google

# Test fetch HTML
make run-fetch URL='https://example.com'

# Test analyze
make run-analyzer AUDITS_FILE=.spec/t1/brief/all_audits.json
```

## 📚 Documentation

### Application-specific documentation:
- **[cmd/brief/README.md](cmd/brief/README.md)** - Brief CLI (SERP, suggestions, fetch, analyze)
- **[cmd/colab/README.md](cmd/colab/README.md)** - Colab CLI (Playwright automation)
- **[cmd/extract/README.md](cmd/extract/README.md)** - Extract CLI (images, frontmatter)
- **[cmd/gradio/README.md](cmd/gradio/README.md)** - Gradio CLI (image generation)

### Project documentation:
- **[MAKEFILE_README.md](MAKEFILE_README.md)** - Full Makefile documentation
- **[TEST_DOCUMENTATION.md](TEST_DOCUMENTATION.md)** - Test documentation
- **[.spec/README.md](.spec/README.md)** - Brief Tool documentation
- **[.claude/commands/brief.md](.claude/commands/brief.md)** - /brief command description

### Links:
- **[GitHub Repository](https://github.com/tkowalski29/content-maker)** - Source code
- **[Releases](https://github.com/tkowalski29/content-maker/releases)** - Precompiled binaries

## 📂 Project Structure

```
.
├── .claude/
│   ├── commands/
│   │   └── brief.md              # /brief command definition
│   ├── hooks/                    # Verification hooks
│   │   ├── validate_*.sh
│   │   └── phase*_complete.sh
│   └── prompts/                  # Prompts in Polish
│       ├── outline_system.md
│       ├── outline_user.md
│       ├── audit_system.md
│       └── audit_user.md
├── .spec/                        # Task directory
│   ├── t1/
│   │   ├── task.md              # Article topic
│   │   ├── task.json            # Metadata
│   │   ├── brief/               # Generated data
│   │   └── output/              # Final reports
│   └── README.md
├── cmd/                          # CLI applications
│   ├── brief/
│   │   ├── main.go              # Brief CLI
│   │   └── README.md
│   ├── colab/
│   │   ├── main.go              # Colab CLI
│   │   ├── README.md
│   │   └── env.json.example
│   ├── extract/
│   │   ├── main.go              # Extract CLI
│   │   └── README.md
│   └── gradio/
│       ├── main.go              # Gradio CLI
│       ├── README.md
│       └── example.json
├── internal/
│   ├── brief/                    # Brief logic
│   │   ├── serp/
│   │   ├── suggest/
│   │   ├── fetch/
│   │   └── analyze/
│   ├── colab/                    # Colab logic
│   │   └── runner.go
│   ├── extractor/                # Extract logic
│   │   ├── images.go
│   │   ├── frontmatter.go
│   │   └── types.go
│   ├── gradio/                   # Gradio logic
│   │   └── runner.go
│   └── cli/                      # CLI handlers
│       ├── brief/
│       ├── colab/
│       └── ...
├── Makefile                      # Automation
├── go.mod
├── go.sum
├── README.md                     # This file
├── MAKEFILE_README.md
└── TEST_DOCUMENTATION.md
```

## 🔧 Makefile - Key Commands

```bash
make help                # Help
make build              # Build brief
make test               # Tests
make test-coverage      # Tests + coverage
make clean              # Clean
make check              # All checks (fmt, vet, lint, test)
make run-serp QUERY='...'       # Test SERP
make run-suggest QUERY='...'    # Test suggestions
make run-fetch URL='...'        # Test fetch
```

Full documentation: [MAKEFILE_README.md](MAKEFILE_README.md)

## 🎯 Example Workflows

### Workflow 1: Full content creation cycle

```bash
# 1. SEO brief
mkdir -p .spec/article
echo "Best beaches in Spain for families" > .spec/article/task.md
/brief article

# 2. Launch Colab (Flux1 model)
./colab -env ./env.json -user test1
# You'll receive URL: https://xyz.gradio.live

# 3. Prepare image requests
cat > images.json << EOF
{
  "folder_name": ".spec/article/images",
  "items": [
    {
      "id": "IMAGE_1",
      "alt": "Costa Brava beach with families",
      "prompt": "Family beach Costa Brava, children playing",
      "style": "photorealistic",
      "aspect_ratio": "16:9"
    }
  ]
}
EOF

# 4. Generate images
./gradio -input images.json -gradio_url https://xyz.gradio.live

# 5. Write article (manually or AI)
# vim .spec/article/article.md

# 6. Extract metadata
./extract images -input .spec/article/article.md -output .spec/article/images_meta.json
./extract frontmatter -input .spec/article/article.md -output .spec/article/cms.json
```

### Workflow 2: Brief and analysis only

```bash
# Test SERP
./brief serp -query="spain beaches" -lang=en > serp.json

# Test suggestions
./brief suggest -query="beaches" -provider=google > suggestions.json

# Fetch competitor
./brief fetch -url="https://example.com" -method=jina > content.json

# Full brief through Claude
/brief my-topic
```

## 🔍 Verification Hooks

Each stage of the brief process is verified by bash hooks:

- ✅ `validate_suggestions.sh` - Check keyword suggestions
- ✅ `validate_serp.sh` - Verify SERP results (min 5 URLs)
- ✅ `validate_urls.sh` - Check URL list
- ✅ `validate_html.sh` - Verify HTML (title, text, word_count)
- ✅ `validate_keywords.sh` - Check keyword aggregation
- ✅ `validate_outline.sh` - Verify outline (min 3 sections, min 5 entities)
- ✅ `validate_audit.sh` - Verify audit (required fields)
- ✅ `phase{N}_complete.sh` - Verify phase completeness

## 🌐 External Services

- **SearXNG**: Search engine (self-hosted or public instance)
- **Jina Reader**: https://r.jina.ai (free markdown reader)
- **Netlify Extractor**:  (custom)
- **Google Suggest**: http://suggestqueries.google.com
- **Bing Suggestions**: https://www.bing.com/AS/Suggestions
- **DuckDuckGo**: https://duckduckgo.com/ac/
- **Google Colab**: https://colab.research.google.com
- **Gradio**: https://gradio.app (custom instances)

## 📝 File Formats

### task.md (brief input)
```markdown
Write an article about best beaches in Spain for families
```

### env.json (colab config)
```json
{
  "webhook_url": "https://example.com/hooks/gradio",
  "users": [
    {
      "id": "test1",
      "email": "user@example.com",
      "password": "secret",
      "colab_url": "https://colab.research.google.com/drive/..."
    }
  ]
}
```

### requests.json (gradio input)
```json
{
  "folder_name": "output/images",
  "items": [
    {
      "id": "HERO",
      "alt": "Hero image",
      "prompt": "A futuristic city at sunrise",
      "style": "photorealistic",
      "aspect_ratio": "16:9"
    }
  ]
}
```

## 🤝 Contributing

### Before commit

```bash
make check  # Format, vet, lint, tests
```

### Adding new tests

1. Create `*_test.go` in appropriate package
2. Use table-driven approach
3. Run `make test-coverage`
4. Goal: coverage > 80%

### Adding new hooks

1. Create file `.claude/hooks/validate_*.sh`
2. Return exit code 0 (success) or 1 (error)
3. Use emoji in messages: ✅ ❌ ⚠️
4. Add to documentation in `brief.md`

## 📊 Statistics

- **CLI Applications**: 4 (brief, colab, extract, gradio)
- **Tests**: 42+
- **Code coverage**: 87.3%
- **Languages**: Go, Bash, Markdown, Python (notebooks)
- **Go code lines**: ~3500 (without tests)
- **Verification hooks**: 10
- **Fetch methods**: 3 (Jina, Netlify, direct)
- **Suggestion providers**: 3 (Google, Bing, DuckDuckGo)
- **Platforms**: Linux (amd64, arm64), macOS (Intel, Apple Silicon), Windows

## 🐛 Troubleshooting

### Brief

**Problem: Tests timeout**
```bash
make test-unit  # Skip network tests
```

**Problem: SearXNG returns error 400**
```bash
# Check lang and country parameters
./brief serp -query='test' -lang=en -country=US
```

**Problem: Fetch doesn't work**
```bash
# Try different method
./brief fetch -url='https://example.com' -method=netlify
```

### Colab

**Problem: Playwright not installed**
```bash
go run github.com/playwright-community/playwright-go/cmd/playwright install chromium
```

**Problem: Chrome doesn't close**
```bash
# Manually kill processes
pkill -f chrome-user-data
```

**Problem: No Google session**
```bash
# On first run, log in manually in the opened browser
# Session will be saved in chrome-user-data/<user_id>/
```

### Gradio

**Problem: Connection timeout**
```bash
# Check if Gradio instance is running
curl https://your-instance.gradio.live/gradio_api/queue/status
```

**Problem: Invalid aspect ratio**
```bash
# Use correct format: "16:9", "4:3", "1:1" or "1200x627"
```

### Extract

**Problem: Missing jq**
```bash
# macOS
brew install jq

# Linux
sudo apt-get install jq
```

## 📄 License

Internal tool for Content Maker project.

## 👥 Authors

- Tomasz Kowalski
- Claude (Anthropic) - AI assistant

---

**Last updated**: October 29, 2025
**Version**: 2.0
**Status**: Production ✅
