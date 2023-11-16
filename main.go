package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/jmorganca/ollama/api"
	"github.com/zyedidia/sregx/syntax"
	"go.teddydd.me/plundered/signals"
)

func noerr(err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}

var (
	debug = flag.Bool("debug", false, "debug output")

	model                string
	formatJson           bool
	filePath             string
	customPromptTemplate string
	customOutputTemplate string
	noStdin              bool

	additionalPrompt string
	appendToInput    bool
	prefixInput      bool
)

type Prompt struct {
	Args      string
	Stdin     string
	File      string
	FromFlags string
	System    string
}

var funcMap = template.FuncMap{
	"s": func(regex, input string) string {
		b := &bytes.Buffer{}
		cmd, err := syntax.Compile(regex, b, nil)
		noerr(err)
		cmd.Evaluate([]byte(input))

		return b.String()
	},
	"codeBlock": func(input string) string {
		scanner := bufio.NewScanner(strings.NewReader(input))
		inCodeBlock := false
		var codeBlockContent strings.Builder

		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "```") {
				if inCodeBlock {
					break // End of the code block
				}
				inCodeBlock = true
				continue
			}

			// Append the line to the code block content if inside a code block
			if inCodeBlock {
				codeBlockContent.WriteString(line)
				codeBlockContent.WriteString("\n") // Add newline between lines
			}
		}

		return strings.TrimSpace(codeBlockContent.String())
	},
}

const defaultPromptTpl = `{{- .Args }}
{{ or .File .Stdin "" }}
{{- with .FromFlags }}
{{ . }}
{{- end -}}`

var defaultPrompt = template.Must(template.New("prompt").Parse(defaultPromptTpl))

func (p Prompt) Render(t *template.Template) string {
	b := &bytes.Buffer{}
	err := t.Execute(b, p)
	noerr(err)
	return b.String()
}

type Output struct {
	Prompt           Prompt
	RenderededPrompt string
	Output           string
	Append, Prefix   bool
}

const defaultOutputTpl = `{{- if .Append -}}
{{ with .Prompt -}}{{ or .File .Stdin "" }}{{ end -}}
{{ end -}} 
{{ .Output }}
{{ if .Prefix -}}
{{ with .Prompt -}}{{ or .File .Stdin "" -}}{{ end -}}
{{ end -}} 
`

var defaultOutput = template.Must(template.New("output").Parse(defaultOutputTpl))

func main() {
	prompt := Prompt{}

	flag.StringVar(&filePath, "f", "", "include file in prompt")
	flag.StringVar(&filePath, "file", "", "include file in prompt")

	flag.StringVar(&customPromptTemplate, "i", "", "custom prompt template")
	flag.StringVar(&customPromptTemplate, "input-template", "", "custom prompt template")

	flag.StringVar(&customOutputTemplate, "o", "", "custom output template")
	flag.StringVar(&customOutputTemplate, "output-template", "", "custom output template")

	flag.BoolVar(&noStdin, "n", false, "no standard input")
	flag.BoolVar(&noStdin, "nostdin", false, "no standard input")

	flag.BoolVar(&appendToInput, "a", false, "write input and then output using default template")
	flag.BoolVar(&appendToInput, "append", false, "write input and then output using default template")
	flag.BoolVar(&prefixInput, "p", false, "write output and then input using default template")
	flag.BoolVar(&prefixInput, "prefix", false, "write output and then input using default template")

	flag.StringVar(&model, "m", os.Getenv("OLLAMA_DEFAULT_MODEL"), "model to use")
	flag.StringVar(&model, "model", os.Getenv("OLLAMA_DEFAULT_MODEL"), "model to use")

	flag.BoolVar(&formatJson, "j", false, "JSON output")
	flag.BoolVar(&formatJson, "json", false, "JSON output")
	flag.BoolFunc("J", "JSON output with automatic „Respond using JSON” prompt", func(_ string) error {
		formatJson = true
		prompt.FromFlags = "Respond using JSON"
		return nil
	})
	flag.Parse()
	ctx := signals.SetupSignalHandler()

	prompt.Args = strings.TrimSpace(strings.Join(flag.Args(), " "))

	c, err := api.ClientFromEnvironment()
	noerr(err)

	var input []byte

	if filePath != "" {
		input, err = os.ReadFile(filePath)
		noerr(err)
		prompt.File = string(input)
	}
	if !noStdin {
		input, err = io.ReadAll(os.Stdin)
		noerr(err)
		prompt.Stdin = string(input)
	}

	stream := false

	format := ""
	if formatJson {
		format = "json"
	}

	promptTpl := defaultPrompt
	renderedPrompt := prompt.Render(promptTpl)
	if *debug {
		fmt.Fprintln(os.Stderr, "=== PROMPT ===")
		fmt.Fprintln(os.Stderr, renderedPrompt)
		fmt.Fprintln(os.Stderr, "=== END ===")
	}

	outputTemplate := defaultOutput
	if customOutputTemplate != "" {
		outputTemplate, err = template.New("custom").Funcs(funcMap).Parse(customOutputTemplate)
		noerr(err)
	}

	err = c.Generate(ctx, &api.GenerateRequest{
		Model:  model,
		Prompt: renderedPrompt,
		Stream: &stream,
		Format: format,
	}, func(r api.GenerateResponse) error {
		o := Output{
			Prompt:           prompt,
			RenderededPrompt: renderedPrompt,
			Output:           r.Response,

			Append: appendToInput,
			Prefix: prefixInput,
		}

		b := &bytes.Buffer{}
		err = outputTemplate.Execute(b, o)
		noerr(err)

		fmt.Println(b.String())
		return nil
	})
	noerr(err)
}
