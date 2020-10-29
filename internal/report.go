package internal

import (
	"errors"
	"path/filepath"

	"github.com/google/go-github/v32/github"
	"github.com/spf13/viper"
)

type Report struct {
	Totals struct {
		Errors     int `json:"errors"`
		FileErrors int `json:"file_errors"`
	} `json:"totals"`
	Files map[string]struct {
		Errors   int `json:"errors"`
		Messages []struct {
			Message   string `json:"message"`
			Line      int    `json:"line"`
			Ignorable bool   `json:"ignorable"`
		} `json:"messages"`
	} `json:"files"`
	Errors []string `json:"errors"`
}

func (r *Report) CreateCheckRunAnnotations() ([]*github.CheckRunAnnotation, error) {
	if !viper.IsSet("github.workspace") {
		return nil, errors.New("missing config key: github.workspace")
	}

	var as []*github.CheckRunAnnotation
	for k, v := range r.Files {
		for _, m := range v.Messages {
			p, err := filepath.Rel(viper.GetString("github.workspace"), k)
			if err != nil {
				return nil, err
			}

			a := &github.CheckRunAnnotation{
				Path:            github.String(p),
				StartLine:       github.Int(m.Line),
				EndLine:         github.Int(m.Line),
				AnnotationLevel: github.String("failure"),
				Message:         github.String(m.Message),
			}

			as = append(as, a)
		}
	}

	return as, nil
}
