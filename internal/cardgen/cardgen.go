package cardgen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/nilspolek/AstroCards/internal/pdfreader"
	ollamaapi "github.com/ollama/ollama/api"
)

// Card represents a generated card from a PDF page.
type Card struct {
	PageNumber int
	Question   string
	Answer     string
}

// CardGenerator holds configuration for generating cards using Ollama.
type CardGenerator struct {
	OllamaURL string
	Model     string
	Client    *ollamaapi.Client
}

// NewCardGenerator creates a new CardGenerator with the given Ollama URL and model name.
func NewCardGenerator(ollamaURL, model string) *CardGenerator {
	u, err := url.Parse(ollamaURL)
	if err != nil {
		u, _ = url.Parse("http://localhost:11434") // fallback
	}
	client := ollamaapi.NewClient(u, http.DefaultClient)
	return &CardGenerator{
		OllamaURL: ollamaURL,
		Model:     model,
		Client:    client,
	}
}

const CardPrompt = `Erstelle mit den folgenden content Karteikarten
Content:
%s`

// CardsJSONSchema is the JSON schema for the expected response format from Ollama.
var CardsJSONSchema = []byte(`{
  "type": "object",
  "properties": {
    "cards": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "question": {"type": "string"},
          "answer": {"type": "string"}
        },
        "required": ["question", "answer"]
      }
    }
  },
  "required": ["cards"]
}`)

// DefaultPrompt is the template used for generating cards from PDF pages.
const DefaultPrompt = CardPrompt

// CardFromOllama generates Cards from a pdfreader.Page using the Ollama API and the CardGenerator's config.
func (cg *CardGenerator) CardFromOllama(page pdfreader.Page) ([]Card, error) {
	prompt := fmt.Sprintf(DefaultPrompt, page.Content)
	var response string
	err := cg.Client.Generate(context.Background(), &ollamaapi.GenerateRequest{
		Model:  cg.Model,
		Prompt: prompt,
		Format: CardsJSONSchema,
	}, func(resp ollamaapi.GenerateResponse) error {
		response += resp.Response
		return nil
	})
	if err != nil {
		return nil, err
	}
	cards := parseCardsJSON(response, page.Number, page.Content)
	return cards, nil
}

// parseCardsJSON parses a JSON response with a 'cards' array of question/answer objects.
func parseCardsJSON(response string, pageNumber int, fallbackContent string) []Card {
	type cardObj struct {
		Question string `json:"question"`
		Answer   string `json:"answer"`
	}
	type cardsResp struct {
		Cards []cardObj `json:"cards"`
	}
	var result cardsResp
	err := json.Unmarshal([]byte(response), &result)
	if err != nil || len(result.Cards) == 0 {
		// fallback: single card with page content as question
		return []Card{{
			PageNumber: pageNumber,
			Question:   fallbackContent,
			Answer:     "",
		}}
	}
	cards := make([]Card, 0, len(result.Cards))
	for _, c := range result.Cards {
		if c.Answer == "" {
			continue
		}
		cards = append(cards, Card{
			PageNumber: pageNumber,
			Question:   strings.TrimSpace(c.Question),
			Answer:     strings.TrimSpace(c.Answer),
		})
	}
	return cards
}

// GenerateCards creates cards from a slice of pdfreader.Page using Ollama for content generation.
func (cg *CardGenerator) GenerateCards(pages []pdfreader.Page) ([]Card, error) {
	var (
		cards   []Card
		lastErr error
	)
	for _, page := range pages {
		pageCards, err := cg.CardFromOllama(page)
		if err != nil {
			lastErr = err
			continue
		}
		cards = append(cards, pageCards...)
	}
	return cards, lastErr
}

// GenerateCardsAsync creates cards from a slice of pdfreader.Page using Ollama for content generation asynchronously.
// It writes the current percentage (int) to the progress channel as it processes pages.
// Returns a channel for the resulting cards slice and error.
func (cg *CardGenerator) GenerateCardsAsync(pages []pdfreader.Page, progress chan<- int) (<-chan []Card, <-chan error) {
	resultChan := make(chan []Card, 1)
	errChan := make(chan error, 1)
	go func() {
		var cards []Card
		total := len(pages)
		for i, page := range pages {
			pageCards, err := cg.CardFromOllama(page)
			if err != nil {
				// fallback: use original content if Ollama fails
				pageCards = []Card{{
					PageNumber: page.Number,
					Question:   page.Content,
					Answer:     "",
				}}
			}
			cards = append(cards, pageCards...)
			if progress != nil && total > 0 {
				percent := int(float64(i+1) / float64(total) * 100)
				progress <- percent
			}
		}
		resultChan <- cards
		close(resultChan)
		close(progress)
		errChan <- nil
		close(errChan)
	}()
	return resultChan, errChan
}
