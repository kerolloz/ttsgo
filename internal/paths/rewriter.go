// Package paths implements a post-emit path alias rewriter.
package paths

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

// Rewriter resolves tsconfig path aliases in emitted JS files.
type Rewriter struct {
	cwd          string
	absOutDir    string
	sourceBase   string
	matchers     []pathMatcher
	nodeModCache sync.Map // string → bool
}

type pathMatcher struct {
	pattern     string
	prefix      string
	hasWildcard bool
	targets     []string
}

// importRe matches all JS import/require/export-from specifiers.
var importRe = regexp.MustCompile(
	`(require\(["']|(?:import|export)[^"'\n]*?from\s+["'])([^"'\n]+)(["'])`,
)

// New creates a Rewriter.
func New(cwd string, paths map[string][]string, outDir, rootDir string) *Rewriter {
	absOutDir := outDir
	if !filepath.IsAbs(outDir) {
		absOutDir = filepath.Join(cwd, outDir)
	}

	var sourceBase string
	if rootDir != "" {
		if filepath.IsAbs(rootDir) {
			sourceBase = rootDir
		} else {
			sourceBase = filepath.Join(cwd, rootDir)
		}
	} else {
		sourceBase = cwd
	}

	r := &Rewriter{
		cwd:        cwd,
		absOutDir:  absOutDir,
		sourceBase: sourceBase,
	}
	r.buildMatchers(paths)
	return r
}

func (r *Rewriter) buildMatchers(paths map[string][]string) {
	for pattern, targets := range paths {
		m := pathMatcher{pattern: pattern, targets: targets}
		if strings.HasSuffix(pattern, "/*") {
			m.hasWildcard = true
			m.prefix = strings.TrimSuffix(pattern, "/*")
		} else if strings.HasSuffix(pattern, "*") {
			m.hasWildcard = true
			m.prefix = strings.TrimSuffix(pattern, "*")
		} else {
			m.hasWildcard = false
			m.prefix = pattern
		}
		r.matchers = append(r.matchers, m)
	}
}

// RewriteDir walks outDir and rewrites all .js files in parallel.
func (r *Rewriter) RewriteDir(outDir string) error {
	if len(r.matchers) == 0 {
		return nil
	}

	absOut := outDir
	if !filepath.IsAbs(outDir) {
		absOut = filepath.Join(r.cwd, outDir)
	}

	var files []string
	err := filepath.WalkDir(absOut, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".js") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(files))
	// Use NumCPU for the semaphore to better utilize hardware
	sem := make(chan struct{}, runtime.NumCPU())

	for _, f := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := r.rewriteFile(filePath); err != nil {
				errs <- err
			}
		}(f)
	}
	wg.Wait()
	close(errs)

	for e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

// rewriteFile rewrites path aliases in a single JS file.
func (r *Rewriter) rewriteFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	content := string(data)
	modified := false

	// Optimized: avoid double-regex by using FindAllStringSubmatchIndex
	matches := importRe.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return nil
	}

	var sb strings.Builder
	lastPos := 0
	for _, match := range matches {
		sb.WriteString(content[lastPos:match[0]])
		
		matchPrefix := content[match[2]:match[3]]
		specifier := content[match[4]:match[5]]
		closingQuote := content[match[6]:match[7]]

		resolved, ok := r.resolveAlias(specifier, filePath)
		if ok {
			sb.WriteString(matchPrefix)
			sb.WriteString(resolved)
			sb.WriteString(closingQuote)
			modified = true
		} else {
			sb.WriteString(content[match[0]:match[1]])
		}
		lastPos = match[1]
	}
	sb.WriteString(content[lastPos:])

	if !modified {
		return nil
	}
	return os.WriteFile(filePath, []byte(sb.String()), 0644)
}

func (r *Rewriter) resolveAlias(specifier, fromFile string) (string, bool) {
	if strings.HasPrefix(specifier, ".") || strings.HasPrefix(specifier, "/") {
		return "", false
	}

	if r.isNodeModule(specifier) {
		return "", false
	}

	for _, m := range r.matchers {
		remainder, matched := matchPattern(m, specifier)
		if !matched {
			continue
		}

		for _, tpl := range m.targets {
			var sourceTargetAbs string
			tplClean := filepath.FromSlash(tpl)
			if m.hasWildcard {
				tplBase := strings.TrimSuffix(strings.TrimSuffix(tplClean, "/*"), "*")
				sourceTargetAbs = filepath.Join(r.cwd, tplBase, remainder)
			} else {
				sourceTargetAbs = filepath.Join(r.cwd, tplClean)
			}

			rel, err := filepath.Rel(r.sourceBase, sourceTargetAbs)
			if err != nil || strings.HasPrefix(rel, "..") {
				rel, err = filepath.Rel(r.cwd, sourceTargetAbs)
				if err != nil {
					continue
				}
			}
			outputTargetAbs := filepath.Join(r.absOutDir, rel)

			fromDir := filepath.Dir(fromFile)
			relPath, err := filepath.Rel(fromDir, outputTargetAbs)
			if err != nil {
				continue
			}

			relPath = filepath.ToSlash(relPath)
			if !strings.HasPrefix(relPath, ".") {
				relPath = "./" + relPath
			}
			return relPath, true
		}
	}
	return "", false
}

func matchPattern(m pathMatcher, specifier string) (string, bool) {
	if m.hasWildcard {
		if specifier == m.prefix {
			return "", true
		}
		if strings.HasPrefix(specifier, m.prefix+"/") {
			return strings.TrimPrefix(specifier, m.prefix+"/"), true
		}
	} else {
		if specifier == m.pattern {
			return "", true
		}
	}
	return "", false
}

func (r *Rewriter) isNodeModule(specifier string) bool {
	parts := strings.SplitN(specifier, "/", 3)
	pkgName := parts[0]
	if strings.HasPrefix(pkgName, "@") && len(parts) >= 2 {
		pkgName = parts[0] + "/" + parts[1]
	}

	if v, ok := r.nodeModCache.Load(pkgName); ok {
		return v.(bool)
	}

	_, err := os.Stat(filepath.Join(r.cwd, "node_modules", pkgName))
	exists := err == nil
	r.nodeModCache.Store(pkgName, exists)
	return exists
}
