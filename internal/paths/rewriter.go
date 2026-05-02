package paths

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
)

// Rewriter resolves tsconfig path aliases in emitted JS files.
type Rewriter struct {
	cwd          string
	absOutDir    string
	sourceBase   string
	matchers     []pathMatcher
	nodeModCache sync.Map
}

type pathMatcher struct {
	pattern     string
	prefix      string
	hasWildcard bool
	targets     []string
}

// requireRe matches require("...") and dynamic import("...").
var requireRe = regexp.MustCompile(`(require\(["']|import\s*\(\s*["'])([^"'\n]+)(["'])`)

// fromRe matches static import/export ... from "..." including multiline.
var fromRe = regexp.MustCompile(`(?s)((?:import|export)[^"']*?from\s+["'])([^"']+)(["'])`)

type match struct {
	start, end  int
	prefix, specifier, quote string
}

func collectMatches(re *regexp.Regexp, content string) []match {
	raw := re.FindAllStringSubmatchIndex(content, -1)
	out := make([]match, 0, len(raw))
	for _, m := range raw {
		out = append(out, match{
			start:     m[0],
			end:       m[1],
			prefix:    content[m[2]:m[3]],
			specifier: content[m[4]:m[5]],
			quote:     content[m[6]:m[7]],
		})
	}
	return out
}

// New creates a Rewriter.
func New(cwd string, paths map[string][]string, outDir, rootDir string) *Rewriter {
	absOutDir := outDir
	if !filepath.IsAbs(outDir) {
		absOutDir = filepath.Join(cwd, outDir)
	}

	sourceBase := cwd
	if rootDir != "" {
		if filepath.IsAbs(rootDir) {
			sourceBase = rootDir
		} else {
			sourceBase = filepath.Join(cwd, rootDir)
		}
	}

	r := &Rewriter{cwd: cwd, absOutDir: absOutDir, sourceBase: sourceBase}
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
	if err := filepath.WalkDir(absOut, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".js") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return err
	}

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
		sem  = make(chan struct{}, runtime.NumCPU())
	)

	for _, f := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := r.rewriteFile(filePath); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(f)
	}
	wg.Wait()

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (r *Rewriter) rewriteFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	content := string(data)

	// Collect matches from both regexes, merge by start position.
	all := append(collectMatches(requireRe, content), collectMatches(fromRe, content)...)
	if len(all) == 0 {
		return nil
	}
	sort.Slice(all, func(i, j int) bool { return all[i].start < all[j].start })

	var sb strings.Builder
	lastPos := 0
	modified := false

	for _, m := range all {
		if m.start < lastPos {
			continue // skip overlapping (shouldn't happen in valid JS)
		}
		sb.WriteString(content[lastPos:m.start])

		resolved, ok := r.resolveAlias(m.specifier, filePath)
		if ok {
			sb.WriteString(m.prefix)
			sb.WriteString(resolved)
			sb.WriteString(m.quote)
			modified = true
		} else {
			sb.WriteString(content[m.start:m.end])
		}
		lastPos = m.end
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
			tplClean := filepath.FromSlash(tpl)
			var sourceTargetAbs string
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
	} else if specifier == m.pattern {
		return "", true
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
