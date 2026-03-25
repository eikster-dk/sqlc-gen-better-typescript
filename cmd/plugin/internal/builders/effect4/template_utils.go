package effect4

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"unicode"
)

func executeTemplate(tmpl *template.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return cleanWhitespace(buf.String()), nil
}

func cleanWhitespace(content string) string {
	re := regexp.MustCompile(`\n{3,}`)
	content = re.ReplaceAllString(content, "\n\n")

	re = regexp.MustCompile(`[ \t]+\n`)
	content = re.ReplaceAllString(content, "\n")

	return strings.TrimSpace(content) + "\n"
}

func formatImports(imports Imports) string {
	if len(imports) == 0 {
		return ""
	}

	modules := make([]string, 0, len(imports))
	for mod := range imports {
		modules = append(modules, mod)
	}
	sort.Strings(modules)

	lines := make([]string, 0, len(modules))
	for _, mod := range modules {
		symbols := uniqueSorted(imports[mod])
		if len(symbols) == 0 {
			continue
		}

		lines = append(lines, fmt.Sprintf(`import { %s } from "%s"`, strings.Join(symbols, ", "), mod))
	}

	return strings.Join(lines, "\n")
}

func uniqueSorted(values []string) []string {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		set[value] = struct{}{}
	}

	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)

	return result
}

func toPascalCase(s string) string {
	words := strings.Split(s, "_")
	for i, word := range words {
		if word == "" {
			continue
		}
		runes := []rune(word)
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, "")
}

func toCamelCase(s string) string {
	pascal := toPascalCase(s)
	if pascal == "" {
		return ""
	}

	return strings.ToLower(pascal[:1]) + pascal[1:]
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}
