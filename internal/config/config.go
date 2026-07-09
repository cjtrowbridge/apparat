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
	root := env["APPARAT_RUNTIME_DIR"]
	if root == "" {
		root = filepath.Join(defaultRootDir(env), binaryName)
	}
	flags := flag.NewFlagSet("apparat", flag.ContinueOnError)
	flags.Var((*stringMode)(&mode), "mode", "runtime mode: gui, headless, or auto")
	flags.StringVar(&root, "runtime-dir", root, "runtime data directory")
	doctor := false
	smoke := false
	flags.BoolVar(&doctor, "doctor", false, "run runtime doctor")
	flags.BoolVar(&smoke, "smoke-test", false, "run non-window smoke test")
	if err := flags.Parse(options.Args); err != nil {
		return Config{}, err
	}
	if err := validateMode(mode); err != nil {
		return Config{}, err
	}
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
	}
	return cfg, nil
}

type stringMode Mode

func (mode *stringMode) Set(value string) error {
	parsed := Mode(value)
	if err := validateMode(parsed); err != nil {
		return err
	}
	*mode = stringMode(parsed)
	return nil
}

func (mode *stringMode) String() string { return string(*mode) }

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
		if err := os.MkdirAll(path, 0o700); err != nil {
			return err
		}
	}
	return nil
}

func defaultRootDir(env map[string]string) string {
	if runtime.GOOS == "android" {
		if value := env["APPARAT_ANDROID_FILES_DIR"]; value != "" {
			return filepath.Join(value, "apparat")
		}
		return filepath.Join("/data/data/com.cjtrowbridge.apparat/files", "apparat")
	}
	if value := env["XDG_DATA_HOME"]; value != "" {
		return filepath.Join(value, "apparat")
	}
	if runtime.GOOS == "windows" {
		if value := env["LOCALAPPDATA"]; value != "" {
			return filepath.Join(value, "Apparat")
		}
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".local", "share", "apparat")
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
