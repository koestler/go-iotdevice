package topicMatcher

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type TopicMatcher struct {
	m *regexp.Regexp
}

func CreateMatcherSingleVariable(topicTemplate, variablePlaceholder string) (matcher TopicMatcher, err error) {
	// create regexp to match against
	regNameExpr := `([^/]+)`

	before, after, found := strings.Cut(topicTemplate, variablePlaceholder)
	if !found || variablePlaceholder == "" {
		err = fmt.Errorf("cannot find variablePlaceholder='%s' in topic='%s'", variablePlaceholder, topicTemplate)
		return
	}

	// must not have anything before / after
	expr := "^" + regexp.QuoteMeta(before) + regNameExpr + regexp.QuoteMeta(after) + "$"

	// all user input is quoted, so the expression always compiles
	matcher.m = regexp.MustCompile(expr)

	return
}

func (tm TopicMatcher) ParseTopic(topic string) (variable string, err error) {
	if tm.m == nil {
		return "", errors.New("invalid matcher")
	}

	matches := tm.m.FindStringSubmatch(topic)
	if matches == nil {
		err = fmt.Errorf("topic='%s' does not match", topic)
	} else {
		variable = matches[1]
	}

	return

}
