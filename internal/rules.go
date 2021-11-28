package internal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dane/protoc-gen-go-svc/gen/svc"
)

func NewRules(f *Field, validate *svc.Validate) ([]string, error) {
	var rules []string
	if validate.GetRequired() {
		rules = append(rules, "validation.Required")
	}

	if f.Type == MessageType {
		var prefix string
		if f.Message.IsExternal {
			prefix = "External"
		}
		rules = append(rules, fmt.Sprintf("validation.By(v.By%s%s)", prefix, f.Message.Name))
	}

	switch validate.GetIs() {
	case svc.Validate_UUID:
		rules = append(rules, "is.UUID")
	case svc.Validate_EMAIL:
		rules = append(rules, "is.Email")
	case svc.Validate_URL:
		rules = append(rules, "is.URL")
	}

	var in []string
	for _, value := range validate.GetIn() {
		switch f.Type {
		case BooleanType:
			if value != "true" && value != "false" {
				return nil, NewErrInvalidRuleIn(f, value)
			}
		case Int64Type:
			if _, err := strconv.ParseInt(value, 10, 64); err != nil {
				return nil, NewErrInvalidRuleIn(f, value)
			}
		case Uint64Type:
			if _, err := strconv.ParseUint(value, 10, 64); err != nil {
				return nil, NewErrInvalidRuleIn(f, value)
			}
		case Float64Type:
			if _, err := strconv.ParseFloat(value, 64); err != nil {
				return nil, NewErrInvalidRuleIn(f, value)
			}
		case StringType:
			// no-op
		case BytesType:
			value = fmt.Sprintf("[]byte(%q)", value)
		case EnumType:
			ev, ok := f.EnumValueByName[value]
			if !ok {
				return nil, NewErrInvalidRuleIn(f, value)
			}
			if f.IsPrivate {
				value = fmt.Sprintf("privatepb.%s", ev.Name)
			} else {
				value = fmt.Sprintf("publicpb.%s", ev.Name)
			}
		}
		in = append(in, value)
	}

	if len(in) > 0 {
		rules = append(rules, fmt.Sprintf("validation.In(%s)", strings.Join(in, ",")))
	}

	if validate.GetMin() != nil || validate.GetMax() != nil {
		switch f.Type {
		case Int64Type:
			if value := validate.GetMin(); value != nil {
				rules = append(rules, fmt.Sprintf("validation.Min(%d)", value.GetInt64()))
			}

			if value := validate.GetMax(); value != nil {
				rules = append(rules, fmt.Sprintf("validation.Max(%d)", value.GetInt64()))
			}
		case Uint64Type:
			if value := validate.GetMin(); value != nil {
				rules = append(rules, fmt.Sprintf("validation.Min(%d)", value.GetUint64()))
			}

			if value := validate.GetMax(); value != nil {
				rules = append(rules, fmt.Sprintf("validation.Max(%d)", value.GetUint64()))
			}
		case Float64Type:
			if value := validate.GetMin(); value != nil {
				rules = append(rules, fmt.Sprintf("validation.Min(%f)", value.GetDouble()))
			}

			if value := validate.GetMax(); value != nil {
				rules = append(rules, fmt.Sprintf("validation.Max(%f)", value.GetDouble()))
			}
		case StringType:
			min := validate.GetMin().GetInt64()
			max := validate.GetMax().GetInt64()

			if value := validate.GetMin().GetUint64(); int64(value) > min {
				min = int64(value)
			}

			if value := validate.GetMax().GetUint64(); int64(value) > max {
				max = int64(value)
			}

			rules = append(rules, fmt.Sprintf("validation.Length(%d, %d)", min, max))
		default:
			if validate.GetMin() != nil {
				return nil, NewErrInvalidRuleForField(f, "min")
			}

			if validate.GetMax() != nil {
				return nil, NewErrInvalidRuleForField(f, "max")
			}
		}
	}

	return rules, nil
}
