package cli

//lint:file-ignore faillint This file should be ignored by faillint (fmt in use).

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	kingpin "github.com/alecthomas/kingpin/v2"

	"github.com/TykTechnologies/tyk/cli/bundler"
	"github.com/TykTechnologies/tyk/cli/importer"
	"github.com/TykTechnologies/tyk/cli/linter"
	"github.com/TykTechnologies/tyk/cli/plugin"
	"github.com/TykTechnologies/tyk/cli/version"
	"github.com/TykTechnologies/tyk/internal/build"
	logger "github.com/TykTechnologies/tyk/log"
)

const (
	appName = "tyk"
	appDesc = "Tyk Gateway"
)

var (
	// Conf specifies the configuration file path.
	Conf *string
	// Port specifies the listen port.
	Port *string
	// MemProfile enables memory profiling.
	MemProfile *bool
	// CPUProfile enables CPU profiling.
	CPUProfile *bool
	// BlockProfile enables block profiling.
	BlockProfile *bool
	// MutexProfile enables block profiling.
	MutexProfile *bool
	// HTTPProfile exposes a HTTP endpoint for accessing profiling data.
	HTTPProfile *bool
	// DebugMode sets the log level to debug mode.
	DebugMode *bool
	// LogInstrumentation outputs instrumentation data to stdout.
	LogInstrumentation *bool

	// DefaultMode is set when default command is used.
	DefaultMode bool

	app *kingpin.Application

	log = logger.Get()
)

var initOnce sync.Once

// Init sets all flags and subcommands.
// It's only run once to avoid races over the globals.
// The arguments are ignored for subsequent runs.
func Init(confPaths []string) {
	initOnce.Do(func() {
		setup(confPaths)
	})
}

func setup(confPaths []string) {
	app = kingpin.New(appName, appDesc)
	app.HelpFlag.Short('h')
	app.Version(build.Version)

	// Start/default command:
	startCmd := app.Command("start", "Starts the Tyk Gateway")
	Conf = startCmd.Flag("conf", "load a named configuration file").PlaceHolder("FILE").String()
	Port = startCmd.Flag("port", "listen on PORT (overrides config file)").String()
	MemProfile = startCmd.Flag("memprofile", "generate a memory profile").Bool()
	CPUProfile = startCmd.Flag("cpuprofile", "generate a cpu profile").Bool()
	BlockProfile = startCmd.Flag("blockprofile", "generate a block profile").Bool()
	MutexProfile = startCmd.Flag("mutexprofile", "generate a mutex profile").Bool()
	HTTPProfile = startCmd.Flag("httpprofile", "expose runtime profiling data via HTTP").Bool()
	DebugMode = startCmd.Flag("debug", "enable debug mode").Bool()
	LogInstrumentation = startCmd.Flag("log-instrumentation", "output instrumentation output to stdout").Bool()

	startCmd.Action(func(ctx *kingpin.ParseContext) error {
		DefaultMode = true
		return nil
	})
	startCmd.Default()

	// Linter:
	lintCmd := app.Command("lint", "Runs a linter on Tyk configuration file")
	lintCmd.Action(func(c *kingpin.ParseContext) error {
		confSchema, err := ioutil.ReadFile("cli/linter/schema.json")
		if err != nil {
			return err
		}
		path, lines, err := linter.Run(string(confSchema), confPaths)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if len(lines) == 0 {
			fmt.Printf("found no issues in %s\n", path)
			os.Exit(0)
		}
		fmt.Printf("issues found in %s:\n", path)
		for _, line := range lines {
			fmt.Println(line)
		}
		os.Exit(1)
		return nil
	})

	// Add version command:
	version.AddTo(app)

	// Add import command:
	importer.AddTo(app)

	// Add bundler commands:
	bundler.AddTo(app)

	// Add plugin commands:
	plugin.AddTo(app)
}

// Parse parses the command-line arguments.
func Parse() {
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
