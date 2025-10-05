package hotenv

import (
	"bufio"
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
)

type config struct {
	m map[string]string
}

var (
	cfg        atomic.Value // holds config
	initOnce   sync.Once
	stopOnce   sync.Once
	cancelFunc context.CancelFunc

	// defaults
	defaultPath     = "/app/secrets/.env"
	defaultDebounce = 800 * time.Millisecond

	// options (set before first Get/Init)
	optFallbackToProcessEnv atomic.Bool // default true
	optLogger               = log.Printf
)

// --------- Public API ----------

// Getenv returns the value for key. If not present, it returns def (if provided) or "".
// The first call lazily starts a watcher on SECRETS_FILE (or /secrets/.env).
func Getenv(key string, def ...string) string {
	ensureStarted("")
	v := get(key)
	if v == "" && len(def) > 0 {
		return def[0]
	}
	return v
}

// Init starts the watcher explicitly with a given path. Call at program start if you prefer.
// If path == "", it uses SECRETS_FILE or the default path.
// Safe to call multiple times; only the first has an effect.
func Init(path string) {
	ensureStarted(path)
}

// Stop stops the background watcher (useful for tests/shutdown).
func Stop() {
	stopOnce.Do(func() {
		if cancelFunc != nil {
			cancelFunc()
		}
	})
}

// WithFallbackToProcessEnv controls whether os.Getenv is consulted
// when a key is missing from the file. Default: true.
func WithFallbackToProcessEnv(enabled bool) {
	optFallbackToProcessEnv.Store(enabled)
}

// WithLogger lets you override the logger (printf-style). Call before Init/Getenv.
func WithLogger(fn func(format string, v ...any)) {
	if fn != nil {
		optLogger = fn
	}
}

// WithDefaultPath lets you override the implicit file path used by lazy init.
// Call before Init/Getenv.
func WithDefaultPath(path string) {
	if path != "" {
		defaultPath = path
	}
}

// --------- Internals ----------

func ensureStarted(path string) {
	initOnce.Do(func() {
		if path == "" {
			if p := os.Getenv("SECRETS_FILE"); p != "" {
				path = p
			} else {
				path = defaultPath
			}
		}
		// initial load
		if c, err := loadEnvFile(path); err == nil {
			cfg.Store(c)
		} else {
			optLogger("hotenv: initial load failed: %v (continuing with empty config)", err)
			cfg.Store(config{m: map[string]string{}})
		}
		// start watcher
		ctx, cancel := context.WithCancel(context.Background())
		cancelFunc = cancel
		go watchAndReload(ctx, path, defaultDebounce)
	})
}

func get(key string) string {
	// 1) file-based
	if cur, ok := cfg.Load().(config); ok {
		if v := cur.m[key]; v != "" {
			return v
		}
	}
	// 2) optional process env fallback
	if optFallbackToProcessEnv.Load() {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return ""
}

func watchAndReload(ctx context.Context, filePath string, debounce time.Duration) {
	dir := filepath.Dir(filePath)

	w, err := fsnotify.NewWatcher()
	if err != nil {
		optLogger("hotenv: watcher init failed: %v", err)
		return
	}
	defer w.Close()

	if err := w.Add(dir); err != nil {
		optLogger("hotenv: watch add failed: %v", err)
		return
	}

	var timerMu sync.Mutex
	var timer *time.Timer
	trigger := func() {
		timerMu.Lock()
		defer timerMu.Unlock()
		if timer != nil {
			_ = timer.Stop()
		}
		timer = time.AfterFunc(debounce, func() {
			if c, err := loadEnvFile(filePath); err == nil {
				cfg.Store(c)
				optLogger("hotenv: reloaded (%d keys)", len(c.m))
			} else {
				optLogger("hotenv: reload failed: %v", err)
			}
		})
	}

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-w.Events:
			if !ok {
				return
			}
			// Any change in dir (K8s does atomic swaps) -> reload
			if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename|fsnotify.Chmod) != 0 {
				trigger()
			}
		case err := <-w.Errors:
			optLogger("hotenv: watch error: %v", err)
		}
	}
}

// loadEnvFile supports:
// - KEY=VALUE (one line)
// - blank lines and # comments
// - multi-line values quoted with '...' or "..." (closing quote can be on a later line)
func loadEnvFile(path string) (config, error) {
	out := make(map[string]string)

	f, err := os.Open(path)
	if err != nil {
		return config{m: out}, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	var key, value string
	var inMultiline bool
	var quote rune

	for sc.Scan() {
		line := sc.Text()
		if !inMultiline {
			trim := strings.TrimSpace(line)
			if trim == "" || strings.HasPrefix(trim, "#") {
				continue
			}
			kv := strings.SplitN(trim, "=", 2)
			if len(kv) != 2 {
				continue
			}
			key = strings.TrimSpace(kv[0])
			value = strings.TrimSpace(kv[1])

			// quoted single-line or start of multi-line
			if len(value) >= 2 {
				start := rune(value[0])
				end := rune(value[len(value)-1])
				if (start == '\'' || start == '"') && end == start {
					out[key] = strings.TrimSuffix(strings.TrimPrefix(value, string(start)), string(start))
					key, value = "", ""
					continue
				}
				if (start == '\'' || start == '"') && end != start {
					inMultiline = true
					quote = start
					value = strings.TrimPrefix(value, string(start)) + "\n"
					continue
				}
			}
			// unquoted single-line
			out[key] = value
			key, value = "", ""
		} else {
			// collecting multi-line until closing quote
			if strings.HasSuffix(line, string(quote)) {
				value += strings.TrimSuffix(line, string(quote))
				out[key] = value
				inMultiline = false
				key, value = "", ""
			} else {
				value += line + "\n"
			}
		}
	}
	// ignore scanner error or return it
	if err := sc.Err(); err != nil {
		return config{m: out}, err
	}
	return config{m: out}, nil
}
