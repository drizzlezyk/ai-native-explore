//go:build debug

package main

import "testing"

func TestNeedsFallbackUser(t *testing.T) {
	t.Run("no auth cases do not require user", func(t *testing.T) {
		cases := []TestCase{{Name: "public", AuthRequired: false}}

		if needsFallbackUser(cases) {
			t.Fatalf("expected public-only cases not to require fallback user")
		}
	})

	t.Run("auth case without username requires user", func(t *testing.T) {
		cases := []TestCase{{Name: "private", AuthRequired: true}}

		if !needsFallbackUser(cases) {
			t.Fatalf("expected auth case without username to require fallback user")
		}
	})

	t.Run("auth case with explicit username does not require fallback user", func(t *testing.T) {
		cases := []TestCase{{Name: "github", AuthRequired: true, AuthUsername: "octocat"}}

		if needsFallbackUser(cases) {
			t.Fatalf("expected auth case with explicit username not to require fallback user")
		}
	})
}
