# AstroCards

AstroCards is a command-line tool that generates flashcards (question/answer pairs) from PDF documents using a local Large Language Model (LLM) via [Ollama](https://ollama.com/). It is designed to help you quickly turn your study materials, lecture notes, or any PDF content into digital flashcards for efficient learning and review.

## Features
- Converts each page of a PDF into one or more flashcards using an LLM
- Outputs cards as structured JSON (with page number, question, and answer)
- Asynchronous processing with progress bar
- Customizable Ollama server URL and model

## Requirements
- [Go](https://go.dev/doc/install) 1.24.4 or newer
- [Ollama](https://ollama.com/) running locally or remotely
  - See [Ollama's installation guide](https://ollama.com/download) for your platform
  - Start Ollama with a model, e.g. `ollama run llama3`

## Installation

Install AstroCards using Go:

```sh
go install github.com/nilspolek/AstroCards/cmd/astro-cards@latest
```

This will install the `astro-cards` binary in your `$GOPATH/bin` (usually `$HOME/go/bin`). Make sure this directory is in your `PATH`.

## Usage

```sh
astro-cards --pdf <input.pdf> [--ollama <ollama_url>] [--model <model_name>] [--output <output.json>]
```

- `--pdf` (required): Path to the input PDF file
- `--ollama`: URL of the Ollama server (default: `http://localhost:11434`)
- `--model`: Name of the Ollama model to use (default: `llama3.2`)
- `--output`: Output file for the generated cards (default: stdout)

### Example

1. Start Ollama and pull a model (if not already done):
   ```sh
   ollama run llama3
   ```
2. Generate cards from a PDF:
   ```sh
   astro-cards --pdf my_notes.pdf --output cards.json
   ```

## Output Format

The output is a JSON array of cards. Each card has:
- `pageNumber`: The page number in the PDF
- `question`: The generated question
- `answer`: The generated answer

Example:
```json
[
  {
    "pageNumber": 1,
    "question": "What is the main topic of the introduction?",
    "answer": "The introduction covers ..."
  },
  {
    "pageNumber": 2,
    "question": "Explain the concept of ...",
    "answer": "..."
  }
]
```

## License

[MIT](LICENSE)
