package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
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
	ghWorkspace string

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

	ms, err := r.CreateMessages()
	if err != nil {
		return err
	}

	for _, m := range ms {
		if _, err := fmt.Fprintln(os.Stdout, m); err != nil {
			return err
		}
	}

	if len(ms) > 0 {
		p := message.NewPrinter(language.English)
		e := p.Sprintf("phpstan has identified %d issue(s)", len(ms))

		return errors.New(e)
	}

	return nil
}

func runRootPre(cmd *cobra.Command, args []string) error {
	err := message.Set(
		language.English,
		"php_codesniffer has identified %d issue(s)",
		catalog.Var("issues", plural.Selectf(1, "", plural.One, "issue", plural.Other, "issues")),
		catalog.String("php_codesniffer has identified %[1]d ${issues}"),
	)

	if err != nil {
		return err
	}

	return nil
}
