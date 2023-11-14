package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jmorganca/ollama/api"
	"go.teddydd.me/plundered/signals"
)

func noerr(err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}

var (
	flagModel = flag.String("model", os.Getenv("OLLAMA_DEFAULT_MODEL"), "model to use")
	flagDebug = flag.Bool("debug", false, "debug output")

	additionalPrompt string
	jsonFormat       bool
	filePath         string
	appendToInput    bool
	prefixInput      bool
)

func main() {
	flag.StringVar(&filePath, "f", "", "context file")
	flag.StringVar(&filePath, "file", "", "context file")

	flag.BoolVar(&appendToInput, "a", false, "append to input")
	flag.BoolVar(&appendToInput, "append", false, "append to input")
	flag.BoolVar(&prefixInput, "p", false, "prefix input")
	flag.BoolVar(&prefixInput, "prefix", false, "prefix  input")

	flag.BoolVar(&jsonFormat, "j", false, "JSON output")
	flag.BoolVar(&jsonFormat, "json", false, "JSON output")
	flag.BoolFunc("J", "JSON output with automatic „Respond using JSON” prompt", func(_ string) error {
		jsonFormat = true
		additionalPrompt = "Respond using JSON"
		return nil
	})
	flag.Parse()
	ctx := signals.SetupSignalHandler()

	prompt := strings.TrimSpace(strings.Join(flag.Args(), " "))

	c, err := api.ClientFromEnvironment()
	noerr(err)

	var input []byte

	if filePath != "" {
		input, err = os.ReadFile(filePath)
		noerr(err)
		prompt = fmt.Sprintf("%s\n\n%s", prompt, string(input))
	} else {
		input, err = io.ReadAll(os.Stdin)
		noerr(err)
		prompt = fmt.Sprintf("%s\n\n%s", prompt, string(input))
	}

	if additionalPrompt != "" {
		prompt = fmt.Sprintf("%s\n%s", prompt, additionalPrompt)
	}

	stream := false

	if *flagDebug {
		fmt.Fprintf(os.Stderr, "=== prompt:\n%s\n===\n", prompt)
	}

	if appendToInput {
		fmt.Println(string(input))
	}

	format := ""
	if jsonFormat {
		format = "json"
	}

	err = c.Generate(ctx, &api.GenerateRequest{
		Model:  *flagModel,
		Prompt: prompt,
		Stream: &stream,
		Format: format,
	}, func(r api.GenerateResponse) error {
		fmt.Println(r.Response)
		return nil
	})
	noerr(err)
	if prefixInput {
		fmt.Println(string(input))
	}
}
