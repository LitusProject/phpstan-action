package cmd

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"phpstan-action/internal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

var (
	ghRepository string
	ghSHA        string
	ghToken      string
	ghWorkspace  string

	rootCmd = &cobra.Command{
		Use:          "phpstan-action",
		Short:        "PHPStan Action",
		SilenceUsage: true,
		RunE:         runRoot,
		PreRunE:      runRootPre,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().StringVar(
		&ghRepository,
		"github.repository",
		"",
		"owner and repository name",
	)

	rootCmd.Flags().StringVar(
		&ghSHA,
		"github.sha",
		"",
		"commit hash that triggered the workflow",
	)

	rootCmd.Flags().StringVar(
		&ghToken,
		"github.token",
		"",
		"installation access token for the job",
	)

	rootCmd.Flags().StringVar(
		&ghWorkspace,
		"github.workspace",
		"",
		"github workspace directory path",
	)

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		log.Fatal(err)
	}
}

func initConfig() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(
		strings.NewReplacer(".", "_", "-", "_"),
	)
}

func runRoot(cmd *cobra.Command, args []string) error {
	d, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	r := new(internal.Report)
	if err := json.Unmarshal(d, r); err != nil {
		return err
	}

	c, err := internal.NewClient()
	if err != nil {
		return err
	}

	cr, err := c.CreateCheckRun()
	if err != nil {
		return err
	}

	as, err := r.CreateCheckRunAnnotations()
	if err != nil {
		return err
	}

	if err := c.UpdateCheckRun(cr, as); err != nil {
		return err
	}

	if err := c.CompleteCheckRun(cr, as); err != nil {
		return err
	}

	if len(as) > 0 {
		p := message.NewPrinter(language.English)
		e := p.Sprintf("phpstan has identified %d issue(s)", len(as))

		return errors.New(e)
	}

	return nil
}

func runRootPre(cmd *cobra.Command, args []string) error {
	{
		err := message.Set(
			language.English,
			"PHPStan has identified %d issue(s).",
			catalog.Var("issues", plural.Selectf(1, "", plural.One, "issue", plural.Other, "issues")),
			catalog.String("PHPStan has identified %[1]d ${issues}."),
		)

		if err != nil {
			return err
		}
	}

	{
		err := message.Set(
			language.English,
			"phpstan has identified %d issue(s)",
			catalog.Var("issues", plural.Selectf(1, "", plural.One, "issue", plural.Other, "issues")),
			catalog.String("phpstan has identified %[1]d ${issues}"),
		)

		if err != nil {
			return err
		}
	}

	return nil
}
