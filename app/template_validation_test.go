package main

import (
	"strings"
	"testing"
	"time"
)

func TestDocumentedTemplateTokensValidateAndFormat(t *testing.T) {
	stamp := time.Date(2051, 7, 13, 22, 45, 9, 0, time.UTC)
	for _, tc := range []struct{ template, want string }{
		{"{time}", "22:45:09"}, {"{date}", "07-13-51"},
		{"{yyyy}", "2051"}, {"{yy}", "51"}, {"{MM}", "07"}, {"{dd}", "13"},
		{"{HH}", "22"}, {"{mm}", "45"}, {"{ss}", "09"}, {"{ddd}", "Thu"},
		{"{HH}:{mm}:{ss}", "22:45:09"}, {"{yy}-{MM}-{dd}", "51-07-13"},
	} {
		t.Run(tc.template, func(t *testing.T) {
			s := defaults()
			s.Template = tc.template
			if err := s.validate(); err != nil {
				t.Fatal(err)
			}
			if got := Format(stamp, s); got != tc.want {
				t.Fatalf("got %q; want %q", got, tc.want)
			}
		})
	}
	for _, invalid := range []string{"", "plain text", "{HH", "{hh}", "{unknown}", "{time}" + strings.Repeat("x", 2000)} {
		s := defaults()
		s.Template = invalid
		if err := s.validate(); err == nil {
			t.Errorf("accepted invalid template %q", invalid)
		}
	}
}
