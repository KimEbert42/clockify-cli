package output

import (
	"io"

	"gopkg.in/yaml.v3"
)

type YAMLFormatter struct{}

func (f *YAMLFormatter) Format(w io.Writer, data any, headers []string) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	return enc.Encode(data)
}
