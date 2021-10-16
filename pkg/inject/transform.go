package inject

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	jsonpatch "github.com/evanphx/json-patch"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	yamlconvert "sigs.k8s.io/yaml"
)

type Inputs []Input
type Input struct {
	ArgNumber  int
	Identifier string
	Reader     io.Reader
}

type Transformer struct {
	PatchGenerator PatchGenerator
	Inputs         Inputs
	Output         io.Writer
}

func ArgumentsToInputs(args []string) (Inputs, error) {
	inputs := Inputs{}
	for i, arg := range args {
		if arg == "-" {
			input := Input{
				ArgNumber:  i,
				Identifier: arg,
				Reader:     os.Stdin,
			}
			inputs = append(inputs, input)
			continue
		}

		file, err := os.Open(arg)
		if err != nil {
			return nil, fmt.Errorf("failed to open input(%d): %s, error: %w", i, arg, err)
		}

		input := Input{
			ArgNumber:  i,
			Identifier: arg,
			Reader:     file,
		}
		inputs = append(inputs, input)
	}

	return inputs, nil
}
func (t *Transformer) Transform() error {
	first := true
	for _, v := range t.Inputs {
		if !first {
			_, err := t.Output.Write([]byte("---\n"))
			if err != nil {
				return fmt.Errorf("failed to write to standard output stream, error: %v", err)
			}
		}

		first = false
		err := t.transformInput(v)

		if err != nil {
			return fmt.Errorf("transformation failed for input: %s(%d), error: %w", v.Identifier, v.ArgNumber, err)
		}

	}

	return nil
}

func (t *Transformer) transformInput(input Input) error {
	reader := yaml.NewYAMLReader(bufio.NewReaderSize(input.Reader, 4096))

	first := true
	for {
		bytes, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		obj, err := parseTypeMetaSkeleton(bytes)
		if err != nil {
			return err
		}

		if obj == nil {
			if !first {
				_, err = t.Output.Write([]byte("---\n"))
				if err != nil {
					return fmt.Errorf("failed to write to standard output stream, error: %v", err)
				}
			}

			_, err = t.Output.Write(bytes)
			if err != nil {
				return fmt.Errorf("failed to write to standard output stream, error: %v", err)
			}

			first = false
			fmt.Fprintf(os.Stderr, "unknown TypeMeta in input: %s(%d), writing to output as-is\n", input.Identifier, input.ArgNumber)
			continue
		}

		err = yaml.Unmarshal(bytes, obj)
		if err != nil {
			return err
		}

		patchObj, err := t.PatchGenerator.Generate(obj, "")
		if err != nil {
			return fmt.Errorf("failed to generate patch for kind: %T, error: %w", obj, err)
		}

		patchJSON, err := json.Marshal(patchObj)
		if err != nil {
			return err
		}

		patch, err := jsonpatch.DecodePatch(patchJSON)
		if err != nil {
			return err
		}

		origJSON, err := yamlconvert.YAMLToJSON(bytes)
		if err != nil {
			return err
		}

		injectedJSON, err := patch.Apply(origJSON)
		if err != nil {
			return err
		}

		injectedYAML, err := yamlconvert.JSONToYAML(injectedJSON)
		if err != nil {
			return err
		}

		if !first {
			_, err = t.Output.Write([]byte("---\n"))
			if err != nil {
				return fmt.Errorf("failed to write to standard output stream, error: %v", err)
			}
		}

		_, err = t.Output.Write(injectedYAML)
		if err != nil {
			return fmt.Errorf("failed to write to standard output stream, error: %v", err)
		}

		first = false
	}

	return nil
}

func parseTypeMetaSkeleton(data []byte) (interface{}, error) {
	var metainfo metav1.TypeMeta
	err := yaml.Unmarshal(data, &metainfo)
	if err != nil {
		return nil, err
	}

	switch metainfo.Kind {
	case "Deployment":
		return &appsv1.Deployment{}, nil
	case "Pod":
		return &corev1.Pod{}, nil
	case "List":
		return &corev1.List{}, nil
	}

	return nil, nil
}
