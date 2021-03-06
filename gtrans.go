package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/translate"
	openbrowser "github.com/haya14busa/go-openbrowser"
	"golang.org/x/oauth2"
	"golang.org/x/text/language"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/option"
)

const usageMessage = "" +
	`Usage:	gtrans [flags] [input text]
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
`

var (
	targetLang    string
	doOpenBrowser bool
)

func init() {
	flag.StringVar(&targetLang, "to", "", "target language")
	flag.BoolVar(&doOpenBrowser, "open", false, "open Google Translate in browser instead of writing translated result to STDOUT")
}

func usage() {
	fmt.Fprintln(os.Stderr, usageMessage)
	fmt.Fprintln(os.Stderr, "Flags:")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if err := Main(os.Stdin, os.Stdout, targetLang, doOpenBrowser); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Main(r io.Reader, w io.Writer, targetLang string, doOpenBrowser bool) error {
	if targetLang == "" {
		var err error
		targetLang, err = detectTargetLang()
		if err != nil {
			return err
		}
	}

	text := strings.Join(flag.Args(), " ")
	if text == "" {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		text = string(b)
	}

	if doOpenBrowser {
		return openGoogleTranslate(w, targetLang, text)
	}
	return runTranslation(w, targetLang, text)
}

// https://translate.google.com/#auto/{lang}/{input}
func openGoogleTranslate(w io.Writer, targetLang, text string) error {
	u := fmt.Sprintf("https://translate.google.com/#auto/%s/%s", targetLang, url.QueryEscape(text))
	return openbrowser.Start(u)
}

func runTranslation(w io.Writer, targetLang, text string) error {
	ctx := context.Background()
	apiKey := os.Getenv("GOOGLE_TRANSLATE_API_KEY")
	if apiKey == "" {
		return errors.New("GOOGLE_TRANSLATE_API_KEY is not set")
	}

	client, err := translate.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return err
	}
	defer client.Close()

	if sec := os.Getenv("GOOGLE_TRANSLATE_SECOND_LANG"); sec != "" {
		detectionsList, err := client.DetectLanguage(ctx, []string{text})
		if err != nil {
			return err
		}
		for _, detections := range detectionsList {
			for _, detection := range detections {
				if detection.Language.String() == targetLang {
					targetLang = sec
				}
				break
			}
		}
	}
	targetLangTag, err := language.Parse(targetLang)
	if err != nil {
		return err
	}
	opt := &translate.Options{}
	translations, err := client.Translate(ctx, []string{text}, targetLangTag, opt)
	if err != nil {
		return err
	}
	for _, translation := range translations {
		fmt.Fprintln(w, translation.Text)
	}
	return nil
}

func oauthClient(ctx context.Context, apiKey string) *http.Client {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
		Transport: &transport.APIKey{Key: apiKey},
	})
	oauthConfig := &oauth2.Config{}
	token := &oauth2.Token{AccessToken: apiKey}
	httpClient := oauthConfig.Client(ctx, token)
	return httpClient
}

func detectTargetLang() (string, error) {
	if code := os.Getenv("GOOGLE_TRANSLATE_LANG"); code != "" {
		return code, nil
	}
	for _, env := range []string{"LANGUAGE", "LC_ALL", "LANG"} {
		code := langCodeFromLocale(os.Getenv(env))
		if code != "" {
			return code, nil
		}
	}
	return "", errors.New("cannot detect language. Please export $LANG or $GOOGLE_TRANSLATE_LANG (e.g. en, ja)")
}

// https://en.wikipedia.org/wiki/Locale_(computer_software)
func langCodeFromLocale(locale string) string {
	if strings.HasPrefix(locale, "zh_CN") || strings.HasPrefix(locale, "zh_SG") {
		return "zh-CN"
	}

	// Regions using Chinese Traditional: Taiwan, Hong Kong
	if strings.HasPrefix(locale, "zh_TW") || strings.HasPrefix(locale, "zh_HK") {
		return "zh-TW"
	}

	i := strings.Index(locale, "_")
	if i == -1 {
		return ""
	}

	return locale[:i]
}
