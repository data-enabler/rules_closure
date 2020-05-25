package closure_js

import (
	"flag"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

type grepExtern struct {
	token *regexp.Regexp // the search regex
	label string         // the label to add to deps on matches
}

func newExternGrep(token, label string) grepExtern {
	const b = `[^a-zA-Z0-9]`
	return grepExtern{
		token: regexp.MustCompile(b + regexp.QuoteMeta(token) + b),
		label: label,
	}
}

func (ge grepExtern) matches(file []byte) bool {
	return ge.token.Match(file)
}

// jsConfig contains configuration values related to JS rules.
type jsConfig struct {
	// grepExterns is a (crude) mechanism that finds specified tokens in js or
	// jsx files, and if found includes the given label in deps.
	grepExterns []grepExtern
	// rulePerFile is true if gazelle should generate a rule for each js file.
	// By default, it generates a rule per directory, named for the directory.
	rulePerFile bool
}

func newJsConfig() *jsConfig {
	gc := &jsConfig{}
	return gc
}

func getJsConfig(c *config.Config) *jsConfig {
	return c.Exts[jsName].(*jsConfig)
}

func (gc *jsConfig) clone() *jsConfig {
	gcCopy := *gc
	return &gcCopy
}

func (_ *jsLang) KnownDirectives() []string {
	return []string{
		"js_grep_extern",
		"js_rule_per_file",
	}
}

func (_ *jsLang) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {
	c.Exts[jsName] = newJsConfig()
}

func (_ *jsLang) CheckFlags(fs *flag.FlagSet, c *config.Config) error {
	return nil
}

func (_ *jsLang) Configure(c *config.Config, rel string, f *rule.File) {
	var gc *jsConfig
	if raw, ok := c.Exts[jsName]; !ok {
		gc = newJsConfig()
	} else {
		gc = raw.(*jsConfig).clone()
	}
	c.Exts[jsName] = gc

	if f != nil {
		for _, d := range f.Directives {
			switch d.Key {
			case "js_grep_extern":
				fields := strings.Fields(d.Value)
				if len(fields) != 2 {
					log.Println("expected 2 fields: `js_grep_extern (token) (extern label)")
					continue
				}
				gc.grepExterns = append(gc.grepExterns, newExternGrep(fields[0], fields[1]))
			case "js_rule_per_file":
				fields := strings.Fields(d.Value)
				switch len(fields) {
				case 0:
					gc.rulePerFile = true
				case 1:
					b, err := strconv.ParseBool(fields[0])
					if err != nil {
						log.Println("`js_rule_per_file` invalid argument:", err)
					} else {
						gc.rulePerFile = b
					}
				default:
					log.Println("expected 0 or 1 arguments: `js_rule_per_file [true|false]`")
				}
			}
		}
	}
}
