package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Jose-Prince/UWULexer/lib"
)

type LexFileData struct {
	Header string
	Footer string
	// The key represents the regex expanded to only have valid regex items
	// The value is the go code to execute when the regex matches
	Rule map[string]lib.DummyInfo
}

// Example Lex file:
// {
//     package main
// }
//
// let delim = [' ''\t''\n']
// let ws = {delim}+
// let letra = ['A'-'Z''a'-'z']
// let digito = ['0'-'9']
// let id = {letra}({letra}|{digito})*
// let numero = {digito}+(\.{digito}+)?
// let literal = \"({letra}|{digito})*\"
// let operator = '+'|'-'|'*'|'/'
// let oprel = '=='|'<='|'>='|'<'|'>'
//
// rule gettoken =
// 	     {ws}	       { continue } (* Ignora white spaces, tabs y nueva línea)
// 	   | {id}          { return ID }
// 	   | {numero}      { return NUM }
//     | {literal}     { return LIT }
//     | {operator}    { return OP }
//     | {oprel}       { return OPREL }
//     | '='           { return ASSIGN }
//     | '('           { return LPAREN }
//     | ')'           { return RPAREN }
//     | '{'           { return LBRACE }
//     | '}'           { return RBRACE }
//     | eof           { return nil }
//
// {
//     fmt.Println("Footer!")
// }

// El LexFileData del archivo de arriba sería:
// {
// 	Header: "package main"
// 	Footer: "fmt.Println(\"Footer!\")"
// 	Rule: {
// 		"[\t\n ]+": {Code: "continue", Priority: 1},
// 		"[A-Za-z]([A-Za-z]|[0-9])*": {Code: "return ID", Priority: 2},
//		...etc etc que hueva escribir todos xD
// 	}
// }

func LexParser(yalexFile string) LexFileData { // string represents the error
	file, err := os.Open(yalexFile)
	if err != nil {
		fmt.Println("Error opening the file:", err)
		return LexFileData{}
	}
	defer file.Close()

	// Identifies priority
	var index uint
	index = 1

	var info lib.DummyInfo

	scanner := bufio.NewScanner(file)
	var header, footer strings.Builder
	dummyRules := make(map[string]string)
	rules := make(map[string]lib.DummyInfo)
	state := 0 // 0: Reading header, 1: Reading rules, 2: Reading footer

	// Regex to identify
	ruleDeclaration := regexp.MustCompile(`(?i)\b(rule)\b`) // Ignores line "rule gettoken ="
	ruleRegex := regexp.MustCompile(`^\s*let\s+([^\s=]+)\s*=\s*(.*)`)
	regexBrackets := regexp.MustCompile(`'(?:[^']*)'|{([^}]*)}`) // Identifies what is inside {}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Header identification
		if line == "{" && state == 0 {
			continue
		} else if line == "}" && state == 0 {
			state = 1
			continue
		} else if state == 0 {
			header.WriteString(line + "\n")
			continue
		}

		if ruleDeclaration.MatchString(line) {
			continue
		}

		// Rules identification
		match := ruleRegex.FindStringSubmatch(line)

		if len(match) > 2 {
			resolvedValue := resolveRule(match[2], dummyRules)
			dummyRules[match[1]] = resolvedValue
			continue
		}

		if line == "" {
			continue
		}

		bracketsMatches := regexBrackets.FindAllStringSubmatch(line, -1)

		if len(bracketsMatches) == 1 {
			line = strings.Replace(line, bracketsMatches[0][0], "", 1)
			line = strings.TrimSpace(line)
			line = strings.Trim(line, "|")
			line = strings.TrimSpace(line)
			if line[0] == '\'' && line[len(line)-1] == '\'' {
				line = line[1 : len(line)-1]
			}

			code := strings.TrimSpace(bracketsMatches[0][1])

			info.Code = code
			info.Priority = index
			info.Regex = line

			rules[line] = info

			index++
			continue
		} else if len(bracketsMatches) > 2 {
			line = strings.Replace(line, bracketsMatches[len(bracketsMatches)-1][0], "", 1)
			line = strings.TrimSpace(line)
			line = strings.Trim(line, "|")
			line = strings.TrimSpace(line)
			if line[0] == '\'' && line[len(line)-1] == '\'' {
				line = line[1 : len(line)-1]
			}

			code := strings.TrimSpace(bracketsMatches[len(bracketsMatches)-1][1])

			info.Code = code
			info.Priority = index
			info.Regex = line

			rules[line] = info

			index++
			continue

		} else if len(bracketsMatches) > 1 {
			line = strings.Replace(line, bracketsMatches[1][0], "", 1)
			line = strings.TrimSpace(line)
			line = strings.Trim(line, "|")
			line = strings.TrimSpace(line)

			line = line[1 : len(line)-1]

			regexValue := dummyRules[line]
			if regexValue == "" {
				regexValue = line
			}
			code := strings.TrimSpace(bracketsMatches[1][1])
			info.Code = code
			info.Priority = index
			info.Regex = regexValue

			rules[regexValue] = info

			index++
			continue
		}

		// Footer identification
		if line == "{" && state == 1 {
			state = 2
			continue
		} else if state == 2 {
			if line == "}" {
				continue
			}
			footer.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error scaning the file:", err)
	}

	fileData := LexFileData{
		Header: header.String(),
		Footer: footer.String(),
		Rule:   rules,
	}

	return fileData
}

// Replace rules into other rules
func resolveRule(rule string, rules map[string]string) string {
	re := regexp.MustCompile(`\{([^\}]+)\}`)
	matches := re.FindAllStringSubmatch(rule, -1)

	if len(matches) == 0 {
		return rule
	}

	for _, match := range matches {
		ruleName := match[1]
		if value, exists := rules[ruleName]; exists {
			rule = strings.ReplaceAll(rule, match[0], value)
		} else {
			rule = strings.ReplaceAll(rule, match[0], ruleName)
		}
	}

	return resolveRule(rule, rules)
}
