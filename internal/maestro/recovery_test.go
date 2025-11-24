package maestro

import "testing"

func TestClassifyError(t *testing.T) {
	cases := map[string]ErrorType{
		"syntax error near token": ErrorSyntax,
		"build failed":            ErrorBuild,
		"test failed at":          ErrorTest,
	}
	for s, want := range cases {
		if got := ClassifyError(s); got != want {
			t.Fatalf("%s: want %s got %s", s, want, got)
		}
	}
}
