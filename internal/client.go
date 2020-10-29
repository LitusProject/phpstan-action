package internal

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	defaultRequestTimeout = 30 * time.Second
	maxAnnotationCount    = 50
)

type Client struct {
	Owner   string
	Repo    string
	HeadSHA string

	ghClient *github.Client
}

func (c *Client) CreateCheckRun() (*github.CheckRun, error) {
	opts := github.CreateCheckRunOptions{
		Name:    "PHPStan",
		HeadSHA: c.HeadSHA,
		Status:  github.String("in_progress"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancel()

	cr, _, err := c.ghClient.Checks.CreateCheckRun(ctx, c.Owner, c.Repo, opts)
	if err != nil {
		return nil, err
	}

	return cr, nil
}

func (c *Client) UpdateCheckRun(cr *github.CheckRun, as []*github.CheckRunAnnotation) error {
	p := message.NewPrinter(language.English)
	s := p.Sprintf("PHPStan has identified %d issue(s).", len(as))

	for i := 0; i < len(as); i += maxAnnotationCount {
		j := i + maxAnnotationCount
		if j > len(as) {
			j = len(as)
		}

		opts := github.UpdateCheckRunOptions{
			Name:    cr.GetName(),
			HeadSHA: github.String(cr.GetHeadSHA()),
			Output: &github.CheckRunOutput{
				Title:       github.String("Result"),
				Summary:     github.String(s),
				Annotations: as[i:j],
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
		defer cancel()

		_, _, err := c.ghClient.Checks.UpdateCheckRun(ctx, c.Owner, c.Repo, cr.GetID(), opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) CompleteCheckRun(cr *github.CheckRun, as []*github.CheckRunAnnotation) error {
	cc := "success"
	if len(as) > 0 {
		cc = "failure"
	}

	p := message.NewPrinter(language.English)
	s := p.Sprintf("PHPStan has identified %d issue(s).", len(as))

	opts := github.UpdateCheckRunOptions{
		Name:       cr.GetName(),
		HeadSHA:    github.String(cr.GetHeadSHA()),
		Conclusion: github.String(cc),
		Status:     github.String("completed"),
		Output: &github.CheckRunOutput{
			Title:   github.String("Result"),
			Summary: github.String(s),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancel()

	_, _, err := c.ghClient.Checks.UpdateCheckRun(ctx, c.Owner, c.Repo, cr.GetID(), opts)
	if err != nil {
		return err
	}

	return nil
}

func NewClient() (*Client, error) {
	if !viper.IsSet("github.token") {
		return nil, errors.New("missing config key: github.token")
	}
	if !viper.IsSet("github.repository") {
		return nil, errors.New("missing config key: github.repository")
	}
	if !viper.IsSet("github.sha") {
		return nil, errors.New("missing config key: github.sha")
	}

	gt := viper.GetString("github.token")
	gr := viper.GetString("github.repository")
	gs := viper.GetString("github.sha")

	or := strings.SplitN(gr, "/", 2)

	ghClient := github.NewClient(
		oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: gt,
				},
			),
		),
	)

	return &Client{
		Owner:   or[0],
		Repo:    or[1],
		HeadSHA: gs,

		ghClient: ghClient,
	}, nil
}
