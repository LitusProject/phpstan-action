package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

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

func (r *Report) UnmarshalJSON(data []byte) error {
	type Alias Report

	if err := json.Unmarshal(data, (*Alias)(r)); err != nil {
		a := &struct {
			Files []string `json:"files"`
			*Alias
		}{
			Alias: (*Alias)(r),
		}

		if err := json.Unmarshal(data, &a); err != nil {
			return err
		}

		r.Files = nil
	}

	return nil
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
					strings.ReplaceAll(m.Message, "\n", "%0A"),
				),
			)
		}
	}

	return ms, nil
}
