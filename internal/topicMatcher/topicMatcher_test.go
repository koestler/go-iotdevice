package topicMatcher

import "testing"

func TestMatcher(t *testing.T) {
	t.Run("invalid template", func(t *testing.T) {
		m, err0 := CreateMatcherSingleVariable("some/thing/not/containing/placeholder", "%RegisterName%")
		if err0 == nil {
			t.Errorf("expected an error during createMatcherSingleVariable, got: %s", err0)
		}

		_, err1 := m.ParseTopic("foobar")
		if err1 == nil {
			t.Errorf("expected an error during parseTopic")
		}
	})

	t.Run("valid A", func(t *testing.T) {
		m, err0 := CreateMatcherSingleVariable("prefix/real/control0/%RegisterName%", "%RegisterName%")
		if err0 != nil {
			t.Errorf("did not expect any error, got: %s", err0)
		}

		v, err1 := m.ParseTopic("prefix/real/control0/my-reg")
		if err1 != nil {
			t.Errorf("did not expect any error, got: %s", err1)
		}

		if expect, got := "my-reg", v; expect != got {
			t.Errorf("expect: '%s' but got '%s'", expect, got)
		}
	})

	t.Run("valid B", func(t *testing.T) {
		m, err0 := CreateMatcherSingleVariable("prefix/real/control0/%RegisterName%/foobar", "%RegisterName%")
		if err0 != nil {
			t.Errorf("did not expect any error, got: %s", err0)
		}

		v, err1 := m.ParseTopic("prefix/real/control0/my-reg/foobar")
		if err1 != nil {
			t.Errorf("did not expect any error, got: %s", err1)
		}

		if expect, got := "my-reg", v; expect != got {
			t.Errorf("expect: '%s' but got '%s'", expect, got)
		}
	})

	t.Run("invalid A", func(t *testing.T) {
		m, err0 := CreateMatcherSingleVariable("prefix/real/control0/%RegisterName%", "%RegisterName%")
		if err0 != nil {
			t.Errorf("did not expect any error, got: %s", err0)
		}

		_, err1 := m.ParseTopic("prefix/cmnd/control0/my-reg")
		if err1 == nil {
			t.Errorf("expecxted an error")
		}
	})
}
