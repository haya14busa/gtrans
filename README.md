## gtrans - Command-line translator using Google Translate

## Installation

```
$ go get github.com/haya14busa/gtrans
```

## Setup

### 1) Get Google Translation API key
- See: https://cloud.google.com/translate/v2/quickstart

### 2) Set Google Translation API key as an envitonment variable along with other options.

Setup example:

```
$ echo 'export GOOGLE_TRANSLATE_API_KEY=<Your API KEY>' >> ~/.gtrans.sh
$ echo 'export GOOGLE_TRANSLATE_LANG=ja' >> ~/.gtrans.sh
$ echo 'export GOOGLE_TRANSLATE_SECOND_LANG=en' >> ~/.gtrans.sh
```

#### Bash
```
$ echo '[ -f ~/.gtrans.sh ] && source ~/.gtrans.sh' >> ~/.bashrc
```
#### Zsh
```
$ echo '[ -f ~/.gtrans.sh ] && source ~/.gtrans.sh' >> ~/.zshrc
```

Be careful not to expose your API key! Please use it at your own risk.

## Usage

```
Usage:  gtrans [flags] [input text]
        gtrans translates input text specified by argument or STDIN using Google Translate.
        Source language will be automatically detected.

        export GOOGLE_TRANSLATE_API_KEY=<Your Google Translate API Key>

        [optional]
        export GOOGLE_TRANSLATE_LANG=<default target language (e.g. en, ja, ...)>
        export GOOGLE_TRANSLATE_SECOND_LANG=<second language (e.g. en, ja, ...)>

        If you set both GOOGLE_TRANSLATE_LANG and GOOGLE_TRANSLATE_SECOND_LANG,
        gtrans automatically switches target langage.

        Example:
                $ gtrans "Golang is awesome"
                Golangは素晴らしいです
                $ gtrans "Golangは素晴らしいです"
                Golang is great
                $ gtrans "Golangは素晴らしいです" | gtrans | gtrans | gtrans ...

Flags:
  -to string
        target language
```
