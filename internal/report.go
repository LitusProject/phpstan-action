package internal

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

type GitHubLogLevel string

const (
	GitHubLogLevelDebug   GitHubLogLevel = "debug"
	GitHubLogLevelWarning GitHubLogLevel = "warning"
	GitHubLogLevelError   GitHubLogLevel = "error"
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

func (r *Report) CreateMessages() ([]string, error) {
	if !viper.IsSet("github.workspace") {
		return nil, errors.New("missing config key: github.workspace")
	}

	var ms []string
	for k, v := range r.Files {
		for _, m := range v.Messages {
			p, err := filepath.Rel(viper.GetString("github.workspace"), k)
			if err != nil {
				return nil, err
			}

			ms = append(
				ms,
				fmt.Sprintf(
					"::%s file=%s,line=%d::%s",
					GitHubLogLevelError,
					p,
					m.Line,
					m.Message,
				),
			)
		}
	}

	return ms, nil
}
