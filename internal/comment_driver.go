package internal

import (
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

type Driver interface {
	DelegateMethodName(method *protogen.Method) (string, error)
	DelegateEnumName(enum *protogen.Enum) (string, error)
	DelegateEnumValueName(value *protogen.EnumValue) (string, error)
	DelegateMessageName(message *protogen.Message) (string, error)
	DelegateFieldName(field *protogen.Field) (string, error)
	DeprecatedField(field *protogen.Field) bool
	RequiredField(field *protogen.Field) bool
	DeprecatedMethod(method *protogen.Method) bool
	ValidateMessage(message *protogen.Message) bool
	ValidateField(field *protogen.Field) bool
	ReceiveRequired(field *protogen.Field) bool
	ReceiveEnumValueNames(value *protogen.EnumValue) []string
	Is(field *protogen.Field) (string, error)
	Min(field *protogen.Field) (string, bool, error)
	Max(field *protogen.Field) (string, bool, error)
	In(packageName string, field *protogen.Field) ([]string, error)
}

type commentDriver struct {
	inputs  map[*protogen.Message]struct{}
	outputs map[*protogen.Message]struct{}
}

func NewCommentDriver(inputs, outputs map[*protogen.Message]struct{}) Driver {
	return commentDriver{
		inputs:  inputs,
		outputs: outputs,
	}
}

func (c commentDriver) DelegateMethodName(method *protogen.Method) (string, error) {
	return c.delegate(method.Comments)
}

func (c commentDriver) DelegateEnumName(enum *protogen.Enum) (string, error) {
	return c.delegate(enum.Comments)
}

func (c commentDriver) DelegateEnumValueName(value *protogen.EnumValue) (string, error) {
	return c.delegate(value.Comments)
}

func (c commentDriver) DelegateMessageName(message *protogen.Message) (string, error) {
	return c.delegate(message.Comments)
}

func (c commentDriver) DelegateFieldName(field *protogen.Field) (string, error) {
	return c.delegate(field.Comments)
}

func (c commentDriver) DeprecatedField(field *protogen.Field) bool {
	return c.deprecated(field.Comments)
}

func (c commentDriver) DeprecatedMethod(method *protogen.Method) bool {
	return c.deprecated(method.Comments)
}

func (c commentDriver) RequiredField(field *protogen.Field) bool {
	return c.required(field)
}

func (c commentDriver) ValidateMessage(message *protogen.Message) bool {
	if _, ok := c.inputs[message]; ok {
		return true
	}

	for _, field := range message.Fields {
		if c.ValidateField(field) {
			return true
		}
	}

	return false
}

func (c commentDriver) ValidateField(field *protogen.Field) bool {
	return c.validate(field.Comments)
}

func (c commentDriver) ReceiveRequired(field *protogen.Field) bool {
	return c.receive(field.Comments)["required"] == "true"
}

func (c commentDriver) ReceiveEnumValueNames(value *protogen.EnumValue) []string {
	prefix := fmt.Sprintf("%s %s ", GenSvc, "receive")
	var values []string

	for _, comment := range c.comments(value.Comments, fmt.Sprintf("%s ", "receive")) {
		comment = strings.TrimPrefix(comment, prefix)
		comment = strings.Trim(comment, " ")
		rules := strings.Split(comment, " ")
		for _, rule := range rules {
			kv := strings.SplitN(rule, "=", 2)
			if kv[0] == "name" {
				values = append(values, kv[1])
			}
		}
	}

	return values
}

func (c commentDriver) required(field *protogen.Field) bool {
	return c.validations(field.Comments)["required"] == "true"
}

func (c commentDriver) Min(field *protogen.Field) (string, bool, error) {
	return c.number(field, "min")
}

func (c commentDriver) Max(field *protogen.Field) (string, bool, error) {
	return c.number(field, "max")
}

func (c commentDriver) In(packageName string, field *protogen.Field) ([]string, error) {
	value, ok := c.validations(field.Comments)["in"]
	if !ok {
		return nil, nil
	}

	values := strings.Split(value, ",")
	if c.builtins(values) {
		return values, nil
	}

	if field.Enum != nil {
		for i, value := range values {
			var matched bool
			enumName := field.Enum.GoIdent.GoName
			for _, ev := range field.Enum.Values {
				if string(ev.Desc.Name()) == value {
					valueName := ev.GoIdent.GoName
					values[i] = fmt.Sprintf("%s.%s", packageName, valueName)
					matched = true
				}
			}
			if !matched {
				return nil, fmt.Errorf("invalid value %s for enum %s", value, enumName)
			}
		}

		return values, nil
	}

	return nil, nil
}

func (c commentDriver) builtins(values []string) bool {
	var bools, ints, floats, strs int
	for _, value := range values {
		if c.contains([]string{"true", "false"}, value) {
			bools++
		}

		if len(value) >= 2 && strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
			strs++
		}

		if _, err := strconv.ParseInt(value, 10, 64); err == nil {
			floats++
		}

		if _, err := strconv.ParseFloat(value, 64); err == nil {
			floats++
		}
	}

	count := len(values)
	if bools == count || ints == count || floats == count || strs == count {
		return true
	}
	return false
}

func (c commentDriver) number(field *protogen.Field, key string) (string, bool, error) {
	value, ok := c.validations(field.Comments)[key]
	if !ok {
		return "", false, nil
	}

	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return value, true, nil
	}

	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return value, true, nil
	}

	fieldName := field.GoName
	return "", false, fmt.Errorf("invalid %s value %q for field %s", key, value, fieldName)
}

func (c commentDriver) Is(field *protogen.Field) (string, error) {
	value, ok := c.validations(field.Comments)["is"]
	if !ok {
		return "", nil
	}

	if !c.contains([]string{"email", "uuid", "url"}, value) {
		return "", fmt.Errorf(`invalid validate "is" value %q`, value)
	}

	return value, nil
}

func (c commentDriver) delegate(commentSet protogen.CommentSet) (string, error) {
	prefix := fmt.Sprintf("%s delegate ", GenSvc)
	for _, comment := range c.comments(commentSet, "delegate ") {
		cond := strings.SplitN(strings.TrimPrefix(comment, prefix), "=", 2)
		if cond[0] != "name" {
			return "", fmt.Errorf("invalid key for delegate annotation: %s", cond[0])
		}

		return cond[1], nil
	}

	return "", nil
}

func (c commentDriver) deprecated(commentSet protogen.CommentSet) bool {
	return len(c.comments(commentSet, "deprecated")) > 0
}

func (c commentDriver) comments(commentSet protogen.CommentSet, prefix string) []string {
	comments := strings.Split(string(commentSet.Leading), "\n")
	for _, comment := range commentSet.LeadingDetached {
		comments = append(comments, string(comment))
	}

	var filtered []string
	prefix = fmt.Sprintf("%s %s", GenSvc, prefix)
	for _, comment := range comments {
		if strings.HasPrefix(comment, prefix) {
			filtered = append(filtered, comment)
		}
	}
	return filtered
}

func (c commentDriver) validate(commentSet protogen.CommentSet) bool {
	return len(c.comments(commentSet, "validate ")) > 0
}

func (c commentDriver) receive(commentSet protogen.CommentSet) map[string]string {
	return c.kvs(commentSet, "receive")
}

func (c commentDriver) validations(commentSet protogen.CommentSet) map[string]string {
	return c.kvs(commentSet, "validate")
}

func (c commentDriver) kvs(commentSet protogen.CommentSet, name string) map[string]string {
	prefix := fmt.Sprintf("%s %s ", GenSvc, name)
	values := make(map[string]string)
	for _, comment := range c.comments(commentSet, fmt.Sprintf("%s ", name)) {
		comment = strings.TrimPrefix(comment, prefix)
		comment = strings.Trim(comment, " ")
		rules := strings.Split(comment, " ")
		for _, rule := range rules {
			kv := strings.SplitN(rule, "=", 2)
			values[kv[0]] = kv[1]
		}
	}

	return values
}

func (c commentDriver) contains(values []string, v string) bool {
	for _, value := range values {
		if v == value {
			return true
		}
	}
	return false
}
