package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mattn/go-isatty"

	"github.com/zetamatta/go-texts/mbcs"
)

type Rule struct {
	Target  string
	Sources []string
	Code    []string
}

var rxMacro = regexp.MustCompile(`\$[\(\{]\w+[\{\)]`)
var rxPattern = regexp.MustCompile(`^(\.\w+)(\.\w+)$`)

var makefilePath = flag.String("f", "Makefile", "path of Makefile")

func parseMakefile(macro map[string]string) (map[string]*Rule, string, error) {
	fd, err := os.Open(*makefilePath)
	if err != nil {
		return nil, "", err
	}
	defer fd.Close()

	rules := make(map[string]*Rule)
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
			if val, ok := macro[name]; ok {
				return val
			} else {
				return "%" + name + "%"
			}
		})

		if text[0] == '\t' {
			if current == nil {
				return nil, "", fmt.Errorf("no current target")
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
				return nil, firstentry, fmt.Errorf("none or multi targets")
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
		return nil, firstentry, fmt.Errorf("Syntax Error: %s", text)
	}
	return rules, firstentry, sc.Err()
}

func dumpCode(rules map[string]*Rule, rule *Rule, indent int, w io.Writer) {
	if len(rule.Code) <= 0 && len(rule.Sources) >= 1 {
		newtarget := filepath.Ext(rule.Sources[0]) + filepath.Ext(rule.Target)
		if r, ok := rules[newtarget]; ok {
			rule = r
		}
	}
	noextTarget := rule.Target[:len(rule.Target)-len(filepath.Ext(rule.Target))]
	for _, code1 := range rule.Code {
		code1 = strings.Replace(code1, "$@", rule.Target, -1)
		code1 = strings.Replace(code1, "$*", noextTarget, -1)
		if len(rule.Sources) >= 1 {
			code1 = strings.Replace(code1, "$<", rule.Sources[0], -1)
			code1 = strings.Replace(code1, "$^", strings.Join(rule.Sources, " "), -1)
		}
		for i := 0; i < indent; i++ {
			fmt.Fprint(w, " ")
		}
		fmt.Fprintln(w, code1)
	}
}

func dumpEntry(rules map[string]*Rule, name string, w io.Writer) {
	fmt.Fprintf(w, ":\"%s\"\n", name)
	rule := rules[name]
	if len(rule.Sources) > 0 {
		for _, source1 := range rule.Sources {
			if _, ok := rules[source1]; ok {
				fmt.Fprintf(w, "  call :\"%s\"\n", source1)
			}
		}
		fmt.Fprintf(w, "  call :test %s %s\n", rule.Target, strings.Join(rule.Sources, " "))
		fmt.Fprintf(w, "  if errorlevel 1 (\n")
		dumpCode(rules, rule, 4, w)
		fmt.Fprintf(w, "  )\n")
	} else {
		dumpCode(rules, rule, 2, w)
	}
	fmt.Fprintln(w, "  exit /b")
}

func dumpTools(w io.Writer) {
	io.WriteString(w, `
:test
  if "%~2" == "" exit /b 1
  setlocal
  for /F %%I in ('forfiles /M "%~1" /C "cmd /c echo @fdate-@ftime"') do (
    set "TARGET=%%I"
  )
:each_source
  for /F %%I in ('forfiles /M "%~2" /C "cmd /c echo @fdate-@ftime"') do (
    set "SOURCE=%%I"
  )
  if "%SOURCE%" gtr "%TARGET%" exit /b 1
  shift
  if not "%~2" == "" goto :each_source
  endlocal & exit /b 0
`)
}

func main1(args []string) error {
	macro := map[string]string{}
	for _, arg := range args {
		if pos := strings.IndexRune(arg, '='); pos >= 0 {
			macro[arg[:pos]] = arg[pos+1:]
		}
	}
	rules, firstentry, err := parseMakefile(macro)
	if err != nil {
		return err
	}

	var w io.Writer = os.Stdout
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		var r io.Reader
		r, w = io.Pipe()
		go func() {
			sc := bufio.NewScanner(r)
			for sc.Scan() {
				text, err := mbcs.UtoA(sc.Text(), mbcs.ACP, true)
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
				} else {
					os.Stdout.Write(text)
					os.Stdout.Write([]byte{'\r', '\n'})
				}
			}
			if err := sc.Err(); err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
			}
			os.Stdout.Sync()
		}()
	}

	fmt.Fprintf(w, "@echo off\n")
	fmt.Fprintf(w, "call :\"%%1\"\n")
	fmt.Fprintf(w, "exit /b\n")
	fmt.Fprintf(w, ":\"\"\n")
	fmt.Fprintf(w, "  call :\"%s\"\n", firstentry)
	fmt.Fprintf(w, "  exit /b\n")

	for key := range rules {
		fmt.Fprintln(w)
		dumpEntry(rules, key, w)
	}
	dumpTools(w)
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
