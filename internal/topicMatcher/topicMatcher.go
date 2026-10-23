package topicMatcher

import (
	"fmt"
	"regexp"
	"strings"
)

type TopicMatcher struct {
	m *regexp.Regexp
}

func CreateMatcherSingleVariable(topicTemplate, variablePlaceholder string) (matcher TopicMatcher, err error) {
	// create regexp to match against
	regNameExpr := "([^\\/]+)"

	if !strings.Contains(topicTemplate, variablePlaceholder) {
		err = fmt.Errorf("cannot find variablePlacholder='%s' in topic='%s'", variablePlaceholder, topicTemplate)
		return
	}

	// must not have anything before / after
	expr := "^" + strings.Replace(regexp.QuoteMeta(topicTemplate), variablePlaceholder, regNameExpr, 1) + "$"

	if m, e := regexp.Compile(expr); e != nil {
		err = fmt.Errorf("cannot compile regexp: %s", e)
	} else {
		matcher.m = m
	}

	return
}

func (tm TopicMatcher) ParseTopic(topic string) (variable string, err error) {
	if tm.m == nil {
		return "", fmt.Errorf("invalid matcher")
	}

	matches := tm.m.FindStringSubmatch(topic)
	if matches == nil {
		err = fmt.Errorf("topic='%s' does not match", topic)
	} else {
		variable = matches[1]
	}

	return

}
