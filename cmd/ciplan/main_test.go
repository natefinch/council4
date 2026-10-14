package main

import "testing"

func TestValidationMatrix(t *testing.T) {
	t.Parallel()
	for _, scenario := range []struct {
		name  string
		paths []string
		want  plan
	}{
		{name: "documentation", paths: []string{"README.md", ".agents/skills/impl-loop/SKILL.md"}},
		{name: "skill executable", paths: []string{".agents/skills/card-authoring/scaffold.py"},
			want: plan{Code: true}},
		{name: "authored card", paths: []string{"mtg/cards/authored/terastodon.go"},
			want: plan{Code: true, Preservation: true}},
		{name: "deleted card", paths: []string{"mtg/cards/t/terastodon.go"},
			want: plan{Code: true, Preservation: true}},
		{name: "shared rule", paths: []string{"mtg/rules/continuous.go"},
			want: plan{Code: true, Preservation: true, Slow: true}},
		{name: "model", paths: []string{"mtg/game/card.go"},
			want: plan{Code: true, Preservation: true, Slow: true}},
		{name: "compiler", paths: []string{"cardgen/lower.go"},
			want: plan{Code: true, Compiler: true, Preservation: true, Slow: true}},
		{name: "tooling", paths: []string{"magefiles/support_docs.go"},
			want: plan{Code: true, Compiler: true, Preservation: true, Slow: true}},
		{name: "CI", paths: []string{".github/workflows/ci.yml"},
			want: plan{Code: true, Compiler: true, Preservation: true, Slow: true}},
		{name: "explicit dataset-only upgrade", paths: []string{
			"testdata/card-preservation/corpus.json", "testdata/card-preservation/baseline.json", "README.md",
		}, want: plan{Code: true, Compiler: true, Preservation: true, Slow: true, CorpusUpgrade: true}},
		{name: "upgrade cannot hide rule change", paths: []string{
			"testdata/card-preservation/corpus.json", "testdata/card-preservation/baseline.json", "mtg/rules/damage.go",
		}, want: plan{Code: true, Compiler: true, Preservation: true, Slow: true}},
		{name: "baseline-only change", paths: []string{"testdata/card-preservation/baseline.json"},
			want: plan{Code: true, Compiler: true, Preservation: true, Slow: true}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			if got := selectPlan(scenario.paths); got != scenario.want {
				t.Fatalf("plan = %+v, want %+v", got, scenario.want)
			}
		})
	}
}
