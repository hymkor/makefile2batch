package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/mattn/go-isatty"

	"github.com/zetamatta/go-texts/mbcs"
)

type Rule struct {
	Target  string
	Sources []string
	Code    []string
}

type MakeRules struct {
	Rules        map[string]*Rule
	DefaultEntry string
}

var rxMacro = regexp.MustCompile(`\$[\(\{]\w+[\{\)]`)
var rxPattern = regexp.MustCompile(`^(\.\w+)(\.\w+)$`)
var rxDefMacro = regexp.MustCompile(`^(\w+)=(.*)$`)

var makefilePath = flag.String("f", "Makefile", "path of Makefile")

func parse(makefile string, cmdlineMacro map[string]string) (*MakeRules, error) {
	fd, err := os.Open(makefile)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	rules := make(map[string]*Rule)
	macro := map[string]string{
		"MAKEDIR": "%~dp0.",
		"MAKE":    "%~dfnx0",
	}
	var current *Rule
	firstentry := ""

	sc := bufio.NewScanner(mbcs.NewAutoDetectReader(fd, mbcs.ACP))
	contline := ""
	for sc.Scan() {
		text := sc.Text()
		if len(text) <= 0 {
			continue
		}
		if text[0] == '#' {
			continue
		}

		text = contline + text
		if strings.HasSuffix(text, "\\") {
			contline = text[:len(text)-1]
			continue
		} else {
			contline = ""
		}

		text = rxMacro.ReplaceAllStringFunc(text, func(src string) string {
			name := src[2 : len(src)-1]
			if val, ok := cmdlineMacro[name]; ok {
				return val
			} else if val, ok := macro[name]; ok {
				return val
			} else {
				return "%" + name + "%"
			}
		})

		if m := rxDefMacro.FindStringSubmatch(text); m != nil {
			macro[m[1]] = strings.TrimSpace(m[2])
			continue
		}

		if text[0] == '\t' {
			if current == nil {
				return nil, fmt.Errorf("no current target")
			}
			text = text[1:]
			if len(text) > 0 && text[0] == '-' {
				text = text[1:]
			}
			current.Code = append(current.Code, text)
			continue
		}

		if strings.HasPrefix(text, "if") {
			continue
		}
		if strings.HasPrefix(text, "else") {
			continue
		}
		if strings.HasPrefix(text, "end") {
			continue
		}

		if pos := strings.IndexRune(text, ':'); pos >= 0 {
			targets := strings.Fields(text[:pos])
			if len(targets) != 1 {
				return &MakeRules{nil, firstentry}, fmt.Errorf("none or multi targets")
			}
			sources := strings.Fields(text[pos+1:])
			current = &Rule{
				Target:  targets[0],
				Sources: sources,
				Code:    []string{},
			}
			rules[targets[0]] = current
			if firstentry == "" && targets[0][0] != '.' {
				firstentry = targets[0]
			}
			continue
		}
		if pos := strings.IndexRune(text, '='); pos >= 0 {
			macro[text[:pos]] = text[pos+1:]
			continue
		}
		return &MakeRules{nil, firstentry}, fmt.Errorf("Syntax Error: %s", text)
	}
	return &MakeRules{rules, firstentry}, sc.Err()
}

func (this *MakeRules) dumpCode(rule *Rule, indent int, w io.Writer) {
	rules := this.Rules
	if len(rule.Code) <= 0 && len(rule.Sources) >= 1 {
		newtarget := filepath.Ext(rule.Sources[0]) + filepath.Ext(rule.Target)
		if r, ok := rules[newtarget]; ok {
			rule = r
		}
	}
	noextTarget := rule.Target[:len(rule.Target)-len(filepath.Ext(rule.Target))]
	indents := strings.Repeat(" ", indent)
	fmt.Fprintf(w, "%s@echo on\n", indents)
	for _, code1 := range rule.Code {
		code1 = strings.Replace(code1, "$@", rule.Target, -1)
		code1 = strings.Replace(code1, "$*", noextTarget, -1)
		if len(rule.Sources) >= 1 {
			code1 = strings.Replace(code1, "$<", rule.Sources[0], -1)
			code1 = strings.Replace(code1, "$^", strings.Join(rule.Sources, " "), -1)
		}
		fmt.Fprintf(w, "%s%s\n", indents, code1)
	}
	fmt.Fprintf(w, "%s@echo off\n", indents)
}

func (this *MakeRules) DumpEntry(name string, w io.Writer) bool {
	rules := this.Rules
	useTest := false
	fmt.Fprintf(w, ":\"%s\"\n", name)
	rule := rules[name]
	if len(rule.Sources) > 0 {
		for _, source1 := range rule.Sources {
			if _, ok := rules[source1]; ok {
				fmt.Fprintf(w, "  call :\"%s\"\n", source1)
			}
		}
		useTest = true
		fmt.Fprintf(w, "  call :test %s %s && exit /b\n", rule.Target, strings.Join(rule.Sources, " "))
	} else {
		fmt.Fprintf(w, "  if exist \"%s\" exit /b\n", rule.Target)
	}
	this.dumpCode(rule, 2, w)
	fmt.Fprintln(w, "  exit /b")
	return useTest
}

func dumpTools(w io.Writer) {
	io.WriteString(w, `
:test
  if not exist "%~1" exit /b 1
  if "%~2" == "" exit /b 0
  setlocal
  for /F "tokens=2,3" %%I in ('where /R . /T "%~1"') do set TARGET=%%I_%%J
  echo %TARGET% | findstr _[0-9]: > nul && set TARGET=%TARGET:_=_0%

:each_source
  for /F "tokens=2,3" %%I in ('where /R . /T "%~2"') do set SOURCE=%%I_%%J
  echo %SOURCE% | findstr _[0-9]: > nul && set SOURCE=%SOURCE:_=_0%
  if "%SOURCE%" gtr "%TARGET%" exit /b 1
  shift
  if not "%~2" == "" goto :each_source
  endlocal & exit /b 0`)
}

func main1(args []string) error {
	macro := map[string]string{}
	for _, arg := range args {
		if pos := strings.IndexRune(arg, '='); pos >= 0 {
			macro[arg[:pos]] = arg[pos+1:]
		}
	}
	makerules, err := parse(*makefilePath, macro)
	if err != nil {
		return err
	}

	var w io.Writer = os.Stdout
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		w = mbcs.NewWriter(os.Stdout, mbcs.ACP)
		os.Stdout.Sync()
	}

	fmt.Fprintln(w, "@echo off")
	fmt.Fprintln(w, "rem ***")
	fmt.Fprintln(w, "rem *** Code generated by `makefile2batch`; DO NOT EDIT.")
	fmt.Fprintln(w, "rem *** ( https://github.com/zetamatta/makefile2batch )")
	fmt.Fprintln(w, "rem ***")
	fmt.Fprintln(w, "setlocal")
	fmt.Fprintln(w, `set "PROMPT=$$ "`)
	fmt.Fprintln(w, `call :"%1"`)
	fmt.Fprintln(w, `endlocal`)
	fmt.Fprintln(w, `exit /b`)
	fmt.Fprintln(w, `:""`)
	fmt.Fprintf(w, "  call :\"%s\"\n", makerules.DefaultEntry)
	fmt.Fprintln(w, "  exit /b")

	useTest := false
	keys := make([]string, 0, len(makerules.Rules))
	for key := range makerules.Rules {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintln(w)
		if makerules.DumpEntry(key, w) {
			useTest = true
		}
	}
	if useTest {
		dumpTools(w)
	}
	fmt.Fprintln(w)
	return nil
}

func main() {
	flag.Parse()
	if err := main1(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
