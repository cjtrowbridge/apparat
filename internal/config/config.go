// Defines configuration values and their defaults.
// Uses package-local contracts and Go standard-library types.
package config

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"runtime"
)

type Mode string

const (
	ModeGUI      Mode = "gui"
	ModeHeadless Mode = "headless"
	ModeAuto     Mode = "auto"
)

type Config struct {
	Mode         Mode
	BinaryName   string
	RootDir      string
	LastRunPath  string
	DatabasePath string
	LogsDir      string
	IdentityDir  string
	CacheDir     string
	ArtifactsDir string
	BackupsDir   string
	RecoveryDir  string
	Doctor       bool
	SmokeTest    bool
	Listen       string
	Demo         bool
	Inspect      bool
	Backup       string
	Restore      string
	LegacyRoots  []string
}

type Options struct {
	Args        []string
	Env         map[string]string
	DefaultMode Mode
	BinaryName  string
}

func Load(options Options) (Config, error) {
	env := options.Env
	if env == nil {
		env = environ()
	}
	mode := options.DefaultMode
	if mode == "" {
		mode = ModeAuto
	}
	binaryName := options.BinaryName
	if binaryName == "" {
		binaryName = "apparat"
	}
	base := defaultRootDir(env)
	root := env["APPARAT_RUNTIME_DIR"]
	implicit := root == ""
	if root == "" {
		root = filepath.Join(base, "node")
	}
	flags := flag.NewFlagSet("apparat", flag.ContinueOnError)
	flags.Var((*stringMode)(&mode), "mode", "runtime mode: gui, headless, or auto")
	flags.StringVar(&root, "runtime-dir", root, "runtime data directory")
	listen := "127.0.0.1:0"
	demo := false
	inspect := false
	backup := ""
	restore := ""
	flags.StringVar(&restore, "restore-backup", "", "restore verified SQLite backup to a new runtime directory and exit")
	flags.BoolVar(&inspect, "services", false, "print local services as JSON and exit")
	flags.StringVar(&backup, "backup", "", "write a verified SQLite backup to a new path and exit")
	flags.StringVar(&listen, "listen", listen, "read-only REST loopback IP:port (0 chooses a free port)")
	flags.BoolVar(&demo, "demo-services", false, "create the three mock services if absent")
	doctor := false
	smoke := false
	flags.BoolVar(&doctor, "doctor", false, "run runtime doctor")
	flags.BoolVar(&smoke, "smoke-test", false, "run non-window smoke test")
	{
		err := flags.Parse(options.Args)
		if err != nil {
			return Config{}, err
		}
	}
	{
		err := validateMode(mode)
		if err != nil {
			return Config{}, err
		}
	}
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "runtime-dir" {
			implicit = false
		}
	})
	root = filepath.Clean(root)
	cfg := Config{
		Mode:         mode,
		BinaryName:   binaryName,
		RootDir:      root,
		LastRunPath:  filepath.Join(root, "last_run.log"),
		DatabasePath: filepath.Join(root, "data", "apparat.db"),
		LogsDir:      filepath.Join(root, "logs"),
		IdentityDir:  filepath.Join(root, "identity"),
		CacheDir:     filepath.Join(root, "cache"),
		ArtifactsDir: filepath.Join(root, "artifacts"),
		BackupsDir:   filepath.Join(root, "backups"),
		RecoveryDir:  filepath.Join(root, "recovery"),
		Doctor:       doctor,
		SmokeTest:    smoke,
		Listen:       listen,
		Demo:         demo,
		Inspect:      inspect,
		Backup:       backup,
		Restore:      restore,
	}
	if implicit {
		cfg.LegacyRoots = []string{filepath.Join(base, "apparat"), filepath.Join(base, "apparatd")}
	}
	return cfg, nil
}

type stringMode Mode

func (mode *stringMode) Set(value string) error {
	parsed := Mode(value)
	{
		err := validateMode(parsed)
		if err != nil {
			return err
		}
	}
	*mode = stringMode(parsed)
	return nil
}

func (mode *stringMode) String() string {
	return string(*mode)
}

func validateMode(mode Mode) error {
	switch mode {
	case ModeGUI, ModeHeadless, ModeAuto:
		return nil
	default:
		return errors.New("mode must be gui, headless, or auto")
	}
}

func EnsureDirectories(cfg Config) error {
	for _, path := range []string{cfg.RootDir, filepath.Dir(cfg.DatabasePath), cfg.LogsDir, cfg.IdentityDir, cfg.CacheDir, cfg.ArtifactsDir, cfg.BackupsDir, cfg.RecoveryDir} {
		{
			err := os.MkdirAll(path, 0o700)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func defaultRootDir(env map[string]string) string {
	if runtime.GOOS == "android" {
		{
			value := env["APPARAT_ANDROID_FILES_DIR"]
			if value != "" {
				return filepath.Join(value, "apparat")
			}
		}
		return filepath.Join("/data/data/com.cjtrowbridge.apparat/files", "apparat")
	}
	{
		value := env["XDG_DATA_HOME"]
		if value != "" {
			return filepath.Join(value, "apparat")
		}
	}
	if runtime.GOOS == "windows" {
		{
			value := env["LOCALAPPDATA"]
			if value != "" {
				return filepath.Join(value, "Apparat")
			}
		}
	}
	{
		home, err := os.UserHomeDir()
		if err == nil && home != "" {
			return filepath.Join(home, ".local", "share", "apparat")
		}
	}
	return filepath.Join(os.TempDir(), "apparat")
}

func environ() map[string]string {
	result := map[string]string{}
	for _, item := range os.Environ() {
		for i := 0; i < len(item); i++ {
			if item[i] == '=' {
				result[item[:i]] = item[i+1:]
				break
			}
		}
	}
	return result
}
