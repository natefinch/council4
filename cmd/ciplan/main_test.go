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
		{name: "empty", paths: []string{""}},
		{name: "skill executable", paths: []string{".agents/skills/card-authoring/scaffold.py"},
			want: plan{Code: true}},
		{name: "authored card", paths: []string{"mtg/cards/authored/terastodon.go"},
			want: plan{Code: true}},
		{name: "deleted card", paths: []string{"mtg/cards/t/terastodon.go"},
			want: plan{Code: true}},
		{name: "shared rule", paths: []string{"mtg/rules/continuous.go"},
			want: plan{Code: true}},
		{name: "model", paths: []string{"mtg/game/card.go"},
			want: plan{Code: true}},
		{name: "card data", paths: []string{"internal/carddata/corpus_policy.go"},
			want: plan{Code: true}},
		{name: "tooling", paths: []string{"magefiles/support_docs.go"},
			want: plan{Code: true}},
		{name: "CI", paths: []string{".github/workflows/ci.yml"},
			want: plan{Code: true}},
		{name: "corpus pin data", paths: []string{"testdata/card-preservation/corpus.json"},
			want: plan{Code: true}},
		{name: "unknown file", paths: []string{"new-tool.sh"}, want: plan{Code: true}},
		{name: "mixed docs and code", paths: []string{"README.md", "mtg/rules/damage.go", ""},
			want: plan{Code: true}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			if got := selectPlan(scenario.paths); got != scenario.want {
				t.Fatalf("plan = %+v, want %+v", got, scenario.want)
			}
		})
	}
}
