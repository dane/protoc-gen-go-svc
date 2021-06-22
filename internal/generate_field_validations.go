package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var EnumRegex = regexp.MustCompile(`\A[A-Z_]+\z`)

func generateFieldValidations(pkgName string, field *protogen.Field) ([]string, error) {
	prefix := fmt.Sprintf("%s validate ", GenSvc)
	var validations []string

	var minInt, maxInt int
	var minFloat, maxFloat float64
	var minIntSet, maxIntSet, minFloatSet, maxFloatSet bool

	for _, comment := range validateAnnotations(field.Comments) {
		conditions := strings.Split(strings.TrimPrefix(comment, prefix), " ")
		for _, condition := range conditions {
			cond := strings.SplitN(condition, "=", 2)
			key := cond[0]
			value := cond[1]

			switch key {
			case "required":
				if err := validation.Validate(&value, validation.In("true", "false")); err != nil {
					return nil, err
				}

				validations = append(validations, "validation.Required")
			case "min":
				if is.Int.Validate(&value) == nil {
					if v, err := strconv.Atoi(value); err != nil {
						return nil, err
					} else {
						minInt = v
					}
					minIntSet = true
					continue
				}

				if is.Float.Validate(&value) == nil {
					if v, err := strconv.ParseFloat(value, 64); err != nil {
						return nil, err
					} else {
						minFloat = v
					}
					minFloatSet = true
					continue
				}

				return nil, validation.Validate(&value, is.Int, is.Float)
			case "max":
				if is.Int.Validate(&value) == nil {
					if v, err := strconv.Atoi(value); err != nil {
						return nil, err
					} else {
						maxInt = v
					}
					maxIntSet = true
					continue
				}

				if is.Float.Validate(&value) == nil {
					if v, err := strconv.ParseFloat(value, 64); err != nil {
						return nil, err
					} else {
						maxFloat = v
					}
					maxFloatSet = true
					continue
				}

				return nil, validation.Validate(&value, is.Int, is.Float)
			case "is":
				switch value {
				case "email":
					validations = append(validations, "is.EmailFormat")
				case "uuid":
					validations = append(validations, "is.UUID")
				case "url":
					validations = append(validations, "is.URL")
				default:
					return nil, validation.Validate(&value, validation.In("email", "uuid", "url"))
				}
			case "in":
				values := strings.Split(value, ",")

				if validation.Validate(values, validation.Each(is.Int)) == nil {
					validations = append(validations, fmt.Sprintf("validation.In(%s)", value))
					continue
				}

				if validation.Validate(values, validation.Each(is.Float)) == nil {
					validations = append(validations, fmt.Sprintf("validation.In(%s)", value))
					continue
				}

				if validation.Validate(values, validation.Each(validation.In("true", "false"))) == nil {
					validations = append(validations, fmt.Sprintf("validation.In(%s)", value))
					continue
				}

				if validation.Validate(values, validation.Each(IsQuoted)) == nil {
					validations = append(validations, fmt.Sprintf("validation.In(%s)", value))
					continue
				}

				if validation.Validate(values, validation.Each(validation.Match(EnumRegex))) == nil {
					if err := validation.Validate(field.Desc.Kind(), validation.In(protoreflect.EnumKind)); err != nil {
						return nil, err
					}

					var enums []interface{}
					refs := make(map[string]*protogen.EnumValue)
					for _, v := range field.Enum.Values {
						name := fmt.Sprintf("%s", v.Desc.Name())
						enums = append(enums, name)
						refs[name] = v
					}

					var goEnums []string
					for _, v := range values {
						if err := validation.Validate(v, validation.In(enums...)); err != nil {
							return nil, err
						}

						goEnums = append(goEnums, fmt.Sprintf("%s.%s", pkgName, refs[v].GoIdent.GoName))
					}

					validations = append(validations, fmt.Sprintf("validation.In(%s)", strings.Join(goEnums, ",")))
					continue
				}

				return nil, validation.Validate(&value,
					validation.Each(is.Int),
					validation.Each(is.Float),
					validation.Each(validation.In("true", "false")),
					validation.Each(IsQuoted),
				)

			}
		}
	}

	if minIntSet || maxIntSet {
		switch field.Desc.Kind() {
		case protoreflect.StringKind:
			validations = append(validations, fmt.Sprintf("validation.Length(%d, %d)", minInt, maxInt))
		case protoreflect.Int64Kind:
			if minIntSet {
				validations = append(validations, fmt.Sprintf("validation.Min(%d)", minInt))
			}

			if maxIntSet {
				validations = append(validations, fmt.Sprintf("validation.Max(%d)", maxInt))
			}
		default:
			println("AT: int kind failed")
			return nil, validation.Validate(field.Desc.Kind(), validation.In(protoreflect.StringKind, protoreflect.Int64Kind))
		}
	}

	if minFloatSet || maxFloatSet {
		switch field.Desc.Kind() {
		case protoreflect.FloatKind:
			if minFloatSet {
				validations = append(validations, fmt.Sprintf("validation.Min(%d)", minFloat))
			}

			if maxFloatSet {
				validations = append(validations, fmt.Sprintf("validation.Max(%d)", maxFloat))
			}
		default:
			println(fmt.Sprintf("min set: %v | max set: %v\n", minFloatSet, maxFloatSet))
			return nil, validation.Validate(field.Desc.Kind(), validation.In(protoreflect.FloatKind))
		}
	}

	var mustValidate bool
	if field.Message != nil {
		if path, ok := messagesByImportPath[field.Message.GoIdent.GoImportPath]; ok {
			if message, ok := path[messageName(field.Message)]; ok {
				mustValidate = message.MustValidate
			}
		}
	}

	if mustValidate {
		validations = append(validations, fmt.Sprintf(
			"validation.By(func(interface{}) error { return v.Validate%s(in.%s) })",
			messageName(field.Message),
			field.GoName,
		),
		)
	}

	return validations, nil
}

var IsQuoted = validation.NewStringRule(isQuoted, "must be double quoted strings")

func isQuoted(s string) bool {
	if len(s) > 1 {
		return strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)
	}
	return false
}
