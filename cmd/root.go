package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cozy/cozy-stack/client"
	"github.com/cozy/cozy-stack/client/request"
	build "github.com/cozy/cozy-stack/pkg/config"
	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/pkg/tlsclient"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// DefaultStorageDir is the default directory name in which data
// is stored relatively to the cozy-dispers binary.
const DefaultStorageDir = "storage"

const defaultDevDomain = "cozy.tools:8080"

var cfgFile string

// ErrUsage is returned by the cmd.Usage() method
var ErrUsage = errors.New("Bad usage of command")

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "cozy-dispers <command>",
	Short: "cozy-dispers is the main command",
	Long: `Cozy-DISPERS is a privacy-by-designed platform that make Cozy Stacks
	participate to queries.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return config.Setup(cfgFile)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Display the usage/help by default
		return cmd.Usage()
	},
	// Do not display usage on error
	SilenceUsage: true,
	// We have our own way to display error messages
	SilenceErrors: true,
}

func newAdminClient() *client.Client {
	pass := []byte(os.Getenv("COZY_ADMIN_PASSWORD"))
	if !build.IsDevRelease() {
		if len(pass) == 0 {
			var err error
			fmt.Printf("Password:")
			pass, err = gopass.GetPasswdMasked()
			if err != nil {
				errFatalf("Could not get password from standard input: %s\n", err)
			}
		}
	}

	httpClient, adminURL, err := tlsclient.NewHTTPClient(tlsclient.HTTPEndpoint{
		Host:      config.GetConfig().AdminHost,
		Port:      config.GetConfig().AdminPort,
		Timeout:   10 * time.Minute,
		EnvPrefix: "COZY_ADMIN",
	})
	checkNoErr(err)

	return &client.Client{
		Scheme:     adminURL.Scheme,
		Addr:       adminURL.Host,
		Domain:     adminURL.Host,
		Client:     httpClient,
		Authorizer: &request.BasicAuthorizer{Password: string(pass)},
	}
}

func init() {
	usageFunc := RootCmd.UsageFunc()

	RootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		_ = usageFunc(cmd)
		return ErrUsage
	})

	flags := RootCmd.PersistentFlags()
	flags.StringVarP(&cfgFile, "config", "c", "", "configuration file (default \"$HOME/.cozy.yaml\")")

	flags.String("host", "localhost", "server host")
	checkNoErr(viper.BindPFlag("host", flags.Lookup("host")))

	flags.IntP("port", "p", 8080, "server port")
	checkNoErr(viper.BindPFlag("port", flags.Lookup("port")))

	flags.String("admin-host", "localhost", "administration server host")
	checkNoErr(viper.BindPFlag("admin.host", flags.Lookup("admin-host")))

	flags.Int("admin-port", 6060, "administration server port")
	checkNoErr(viper.BindPFlag("admin.port", flags.Lookup("admin-port")))
}

func checkNoErr(err error) {
	if err != nil {
		panic(err)
	}
}

func errPrintfln(format string, vals ...interface{}) {
	_, err := fmt.Fprintf(os.Stderr, format+"\n", vals...)
	if err != nil {
		panic(err)
	}
}

func errPrintf(format string, vals ...interface{}) {
	_, err := fmt.Fprintf(os.Stderr, format, vals...)
	if err != nil {
		panic(err)
	}
}

func errFatalf(format string, vals ...interface{}) {
	_, err := fmt.Fprintf(os.Stderr, format, vals...)
	if err != nil {
		panic(err)
	}
	os.Exit(1)
}
