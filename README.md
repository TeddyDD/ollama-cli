# Ollama CLI

Simple CLI interface for [Ollama], designed to be integrated into Kakoune text
editor. Can be used as standalone tool as well.

## Installation

This program requires Go 1.21.
```sh
go install go.teddydd.me/ollama-cli@latest
```
You can also set the `OLLAMA_DEFAULT_MODEL` environment variable to specify
a default model. `OLLAMA_HOST` allows to specify Ollama endpoint.

## Usage

Program accepts prompt from arguments and standard input or file if `-f`
is specified. Prompt is constucted as follows:

1. CLI arguments
2. Empty line
3. Standard input or content of the files

The program accepts several options that allow you to customize its behavior:

* `-f`, `-file`: Path to the context file. If provided, the contents of this file will be used as input for the model.
* `-a`, `-append`: Print prompt befre output.
* `-p`, `-prefix`: Print output and then prompt.
* `-s`, `-system`: Override Ollama system prompt.
* `-j`, `-json`: Output JSON (you must mention JSON in prompt).
* `-o`, `-output-template`: Use template for output.
* `-i`, `-input-template`: Use template to build prompt.
* `-J`: Output JSON (appends „Respond using JSON” to prompt).
* `-model`: Model to use.

Here's an example usage:
```sh
ollama-cli -f hello.go add comments
echo '-- hello world' | ollama-cli -a \
    'write function in Lua based on the comment, dont wrap in markdown code block'
echo 'def lerp(a,b,x):' | ollama-cli -p 'write comment for this function'

# JSON
LC_TIME=en_US TZ=UTC date |
    ollama-cli -j date as json object with each component as separate field, skip timezone |
    jq

ollama-cli -n \
  -s 'You are AI model that talks like a pirate. Respond to questions.'\
  'why sky is blue?'

ollama-cli -o 'Mistral says: {{ .Output }}' -n hello Mistral
```

Append and prefix flags are designed for piping text from Kakoune with
<kbd>|</kbd> key.

## Templates

Flags `-o` and `-i` use Go [text/template](https://godocs.io/text/template)
syntax.

Prompt template values:

* Args: CLI args except flags, joined with spaces.
* Stdin
* File
* FromFlags: Additional prompt that would be inserted, based on flags (see `-J` flag).
* System: System prompt flag content.

Output template values:

* Prompt: struct with same fields as prompt template input.
* RenderededPrompt: final prompt used for this generation.
* Output: output produced by LLM.
* Append, Prefix: booleans, flags `-a` and `-p` respectively.

Additional functions:

* `s regex input`: run [structural regexp](https://github.com/zyedidia/sregx) on the input and return printed (`p` command) text.
* `codeBlock input`: extract markdown code block from input.

## Contributing

Send patches to https://lists.sr.ht/~teddy/public-inbox

[Ollama]: https://ollama.ai/
