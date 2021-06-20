package internal

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func delegateAnnotation(commentSet protogen.CommentSet) (string, error) {
	prefix := fmt.Sprintf("%s delegate ", GenSvc)
	for _, comment := range mergeComments(commentSet) {
		if strings.HasPrefix(comment, prefix) {
			cond := strings.SplitN(strings.TrimPrefix(comment, prefix), "=", 2)
			if cond[0] != "name" {
				return "", fmt.Errorf("invalid key for delegate annotation: %s", cond[0])
			}

			return cond[1], nil
		}
	}

	return "", nil
}

func receiveAnnotations(commentSet protogen.CommentSet) ([]string, error) {
	prefix := fmt.Sprintf("%s receive ", GenSvc)
	var values []string
	for _, comment := range mergeComments(commentSet) {
		if strings.HasPrefix(comment, prefix) {
			cond := strings.SplitN(strings.TrimPrefix(comment, prefix), "=", 2)
			if cond[0] != "name" {
				return nil, fmt.Errorf("invalid key for delegate annotation: %s", cond[0])
			}

			values = append(values, cond[1])
		}
	}

	return values, nil
}
