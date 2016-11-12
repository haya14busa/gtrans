package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/translate/v2"
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

var targetLang string

func init() {
	flag.StringVar(&targetLang, "to", "", "target language")
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
	if err := Main(os.Stdin, os.Stdout, targetLang); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type Gtrans struct {
	srv *translate.Service
}

func (gtrans *Gtrans) Translate(text, target string) (string, error) {
	call := gtrans.srv.Translations.List([]string{text}, target)
	call = call.Format("text")
	resp, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("fail to call translate API: %v", err)
	}
	return resp.Translations[0].TranslatedText, nil
}

func (gtrans *Gtrans) Detect(text string) (string, error) {
	call := gtrans.srv.Detections.List([]string{text})
	resp, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("fail to call detection API: %v", err)
	}
	return resp.Detections[0][0].Language, nil
}

func Main(r io.Reader, w io.Writer, targetLang string) error {
	ctx := context.Background()
	apiKey := os.Getenv("GOOGLE_TRANSLATE_API_KEY")
	if apiKey == "" {
		return errors.New("GOOGLE_TRANSLATE_API_KEY is not set")
	}
	service, err := translate.New(oauthClient(ctx, apiKey))
	if err != nil {
		return err
	}
	gtrans := &Gtrans{srv: service}

	text := strings.Join(flag.Args(), " ")
	if text == "" {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		text = string(b)
	}

	if targetLang == "" {
		targetLang, err = detectTargetLang()
		if err != nil {
			return err
		}
	}

	if sec := os.Getenv("GOOGLE_TRANSLATE_SECOND_LANG"); sec != "" {
		detectedSourceLang, err := gtrans.Detect(text)
		if err != nil {
			return err
		}
		if detectedSourceLang == targetLang {
			targetLang = sec
		}
	}

	translatedText, err := gtrans.Translate(text, targetLang)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, translatedText)
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
