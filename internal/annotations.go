package internal

import (
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func delegateMethodName(method *protogen.Method) (string, error) {
	return delegate(method.Comments)
}

func delegateMessageName(message *protogen.Message) (string, error) {
	return delegate(message.Comments)
}

func delegateFieldName(field *protogen.Field) (string, error) {
	return delegate(field.Comments)
}

func deprecatedField(field *protogen.Field) bool {
	return deprecated(field.Comments)
}

func deprecatedMethod(method *protogen.Method) bool {
	return deprecated(method.Comments)
}

func validateMessage(message *protogen.Message) bool {
	if _, ok := inputs[message]; ok {
		return true
	}

	for _, field := range message.Fields {
		if validateField(field) {
			return true
		}
	}

	return false
}

func validateField(field *protogen.Field) bool {
	return validate(field.Comments)
}

func required(field *protogen.Field) bool {
	return validations(field.Comments)["required"] == "true"
}

func min(field *protogen.Field) (string, bool, error) {
	return number(field, "min")
}

func max(field *protogen.Field) (string, bool, error) {
	return number(field, "max")
}

func in(packageName string, field *protogen.Field) ([]string, error) {
	value, ok := validations(field.Comments)["in"]
	if !ok {
		return nil, nil
	}

	values := strings.Split(value, ",")
	if builtins(values) {
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

func builtins(values []string) bool {
	var bools, ints, floats, strs int
	for _, value := range values {
		if contains([]string{"true", "false"}, value) {
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

func number(field *protogen.Field, key string) (string, bool, error) {
	value, ok := validations(field.Comments)[key]
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

func is(field *protogen.Field) (string, error) {
	value, ok := validations(field.Comments)["is"]
	if !ok {
		return "", nil
	}

	if !contains([]string{"email", "uuid", "url"}, value) {
		return "", fmt.Errorf(`invalid validate "is" value %q`, value)
	}

	return value, nil
}

func delegate(commentSet protogen.CommentSet) (string, error) {
	prefix := fmt.Sprintf("%s delegate ", GenSvc)
	for _, comment := range comments(commentSet, "delegate ") {
		cond := strings.SplitN(strings.TrimPrefix(comment, prefix), "=", 2)
		if cond[0] != "name" {
			return "", fmt.Errorf("invalid key for delegate annotation: %s", cond[0])
		}

		return cond[1], nil
	}

	return "", nil
}

func deprecated(commentSet protogen.CommentSet) bool {
	return len(comments(commentSet, "deprecated")) > 0
}

func comments(commentSet protogen.CommentSet, prefix string) []string {
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

func validate(commentSet protogen.CommentSet) bool {
	return len(comments(commentSet, "validate ")) > 0
}

func validations(commentSet protogen.CommentSet) map[string]string {
	prefix := fmt.Sprintf("%s validate ", GenSvc)
	values := make(map[string]string)
	for _, comment := range comments(commentSet, "validate ") {
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

func contains(values []string, v string) bool {
	for _, value := range values {
		if v == value {
			return true
		}
	}
	return false
}
