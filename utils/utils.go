package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

func HumanReadableBytes(bytes string) string {
	if bytes == "" {
		return "-"
	}
	b, err := strconv.ParseInt(bytes, 10, 64)
	if err != nil {
		return bytes
	}
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func TrailingSlash(in string) string {
	out := in
	if !strings.HasSuffix(in, "/") {
		out = fmt.Sprintf("%s/", in)
	}
	return out
}

func Colorise(in string) string {
	cyan := color.New(color.FgCyan).SprintFunc()
	return cyan(in)
}

func ColourStatusCode(code int) string {
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	mag := color.New(color.FgMagenta).SprintFunc()

	out := ""

	if 100 <= code && code < 200 {
		out = mag(code)
	} else if 200 <= code && code < 300 {
		out = green(code)
	} else if 300 <= code && code < 400 {
		out = cyan(code)
	} else if 400 <= code && code < 500 {
		out = yellow(code)
	} else if 500 <= code && code < 600 {
		out = red(code)
	}

	return out
}

func ColourHTTPMethod(method string) string {
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	mag := color.New(color.FgMagenta).SprintFunc()

	out := ""

	if method == "GET" {
		out = cyan(method)
	} else if method == "POST" {
		out = mag(method)
	} else if method == "HEAD" {
		out = blue(method)
	} else if method == "OPTIONS" {
		out = blue(method)
	} else {
		out = yellow(method)
	}

	return out
}

func ColourWebDAVMethod(method string) string {
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	mag := color.New(color.FgMagenta).SprintFunc()

	out := ""

	if method == "HEAD" || method == "PROPFIND" {
		out = cyan(method)
	} else if method == "MKCOL" || method == "MOVE" || method == "DELETE" || method == "COPY" || method == "PROPPATCH" || method == "PUT" || method == "GET" {
		out = mag(method)
	} else {
		out = yellow(method)
	}

	return out
}
