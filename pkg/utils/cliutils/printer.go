package cliutils

import (
	"fmt"
	"github.com/solo-io/solo-kit/pkg/utils/protoutils"
	"io"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

// Printer represents a function that prints a value to io.Writer, usually using
// a table
type Printer func(interface{}, io.Writer) error

// Print - prints the given proto.Message to io.Writer using the specified output format
func Print(output, template string, m proto.Message, tblPrn Printer, w io.Writer) error {
	switch strings.ToLower(output) {
	case "yaml":
		return PrintYAML(m, w)
	case "json":
		return PrintJSON(m, w)
	case "template":
		return PrintTemplate(m, template, w)
	default:
		return tblPrn(m, w)
	}
}


// PrintJSON - prints the given proto.Message to io.Writer in JSON
func PrintJSON(m proto.Message, w io.Writer) error {
	b, err := protoutils.MarshalBytes(m)
	if err != nil {
		return errors.Wrap(err, "unable to convert to JSON")
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}

// PrintYAML - prints the given proto.Message to io.Writer in YAML
func PrintYAML(m proto.Message, w io.Writer) error {
	jsn, err := protoutils.MarshalBytes(m)
	if err != nil {
		return errors.Wrap(err, "unable to marshal")
	}
	b, err := yaml.JSONToYAML(jsn)
	if err != nil {
		return errors.Wrap(err, "unable to convert to YAML")
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}

// PrintTemplate prints the give value using the provided Go template to io.Writer
func PrintTemplate(data interface{}, tmpl string, w io.Writer) error {
	t, err := template.New("output").Parse(tmpl)
	if err != nil {
		return errors.Wrap(err, "unable to parse template")
	}
	return t.Execute(w, data)
}
