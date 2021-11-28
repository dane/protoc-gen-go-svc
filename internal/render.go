package internal

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"text/template"
)

func render(file io.Writer, name, tmpl string, data interface{}) error {
	funcs := template.FuncMap{
		"public_from_private_config":            newPublicFromPrivateConfig(""),
		"deprecated_public_from_private_config": newPublicFromPrivateConfig("Deprecated"),
		"partial":                               partial,
		"type_of":                               typeOf,
	}

	tpl, err := template.New(name).Funcs(funcs).Parse(tmpl)
	if err != nil {
		return err
	}

	for _, partial := range Partials {
		tpl, err = tpl.Parse(partial)
		if err != nil {
			return err
		}
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return err
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	_, err = file.Write(formatted)
	return err
}

type PartialNamer interface {
	PartialName() string
}

type publicFromPrivateConfig struct {
	*Message
	Prefix      string
	partialName string
}

func (p publicFromPrivateConfig) PartialName() string {
	return p.partialName
}

func newPublicFromPrivateConfig(prefix string) func(*Message) publicFromPrivateConfig {
	return func(msg *Message) publicFromPrivateConfig {
		return publicFromPrivateConfig{
			partialName: "to-public-from-private",
			Message:     msg,
			Prefix:      prefix,
		}
	}
}

func partial(data interface{}) (string, error) {
	var name string
	if v, ok := data.(PartialNamer); ok {
		name = v.PartialName()
	}

	tmpl := fmt.Sprintf(`{{ template %q . }}`, name)
	var buf bytes.Buffer
	if err := render(&buf, "partial", tmpl, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func typeOf(f *Field) string {
	switch f.Type {
	case StringType:
		return "string"
	case Int64Type:
		return "int64"
	case Uint64Type:
		return "uint64"
	case Float64Type:
		return "float64"
	case BooleanType:
		return "bool"
	case BytesType:
		return "[]byte"
	case MessageType:
		if f.Message.IsPrivate {
			return fmt.Sprintf("*privatepb.%s", f.Message.Name)
		} else if f.Message.IsExternal {
			return fmt.Sprintf("*%s.%s", f.Message.PackageName, f.Message.Name)
		}
		return fmt.Sprintf("*publicpb.%s", f.Message.Name)
	case EnumType:
		if f.IsPrivate {
			return fmt.Sprintf("privatepb.%s", f.EnumName)
		}
		return fmt.Sprintf("publicpb.%s", f.EnumName)
	case OneOfType:
		// TODO: Support deprecating a oneof field. It will likely
		// require a mutator function of each oneof messages.
	}

	return ""
}
