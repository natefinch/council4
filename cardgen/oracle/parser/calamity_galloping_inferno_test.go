package parser

import "testing"

const calamityGallopingInfernoOracle = "Haste\n" +
	"Whenever Calamity attacks while saddled, choose a nonlegendary creature that saddled it this turn and create a tapped and attacking token that's a copy of it. Sacrifice that token at the beginning of the next end step. Repeat this process once.\n" +
	"Saddle 1"

func TestParseCalamityComposableProcess(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(calamityGallopingInfernoOracle, Context{CardName: "Calamity, Galloping Inferno"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	var repeat *EffectSyntax
	for i := range document.Abilities {
		for j := range document.Abilities[i].Sentences {
			for k := range document.Abilities[i].Sentences[j].Effects {
				effect := &document.Abilities[i].Sentences[j].Effects[k]
				if effect.Kind == EffectRepeatProcess {
					repeat = effect
				}
			}
		}
	}
	if repeat == nil {
		t.Fatal("no repeat process parsed")
	}
	if !repeat.Exact || !repeat.Amount.Known || repeat.Amount.Value != 2 || len(repeat.RepeatBody) != 2 {
		t.Fatalf("repeat = %#v, want exact two-iteration, two-effect process", repeat)
	}
	create := repeat.RepeatBody[0]
	if create.Kind != EffectCreate || !create.Exact ||
		!create.TokenCopyOfChosenSaddleContributor ||
		!create.TokenCopyEntersTapped ||
		!create.TokenCopyAttacksWithSource {
		t.Fatalf("create kind=%v exact=%v chosen=%v tapped=%v same-attack=%v text=%q",
			create.Kind, create.Exact, create.TokenCopyOfChosenSaddleContributor,
			create.TokenCopyEntersTapped, create.TokenCopyAttacksWithSource, create.Text)
	}
	sacrifice := repeat.RepeatBody[1]
	if sacrifice.Kind != EffectSacrifice || !sacrifice.Exact ||
		!sacrifice.CreatedTokensReference ||
		sacrifice.DelayedTiming != DelayedTimingNextEndStep {
		t.Fatalf("sacrifice kind=%v exact=%v created=%v timing=%v text=%q",
			sacrifice.Kind, sacrifice.Exact, sacrifice.CreatedTokensReference,
			sacrifice.DelayedTiming, sacrifice.Text)
	}
}

func TestParseTrailingRepeatProcessIsCardNameIndependent(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever this creature attacks while saddled, choose a nonlegendary creature that saddled it this turn and create a tapped and attacking token that's a copy of it. Sacrifice that token at the beginning of the next end step. Repeat this process once.",
		Context{CardName: "Unrelated Mount"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	repeat := document.Abilities[0].Sentences[0].Effects[0]
	if repeat.Kind != EffectRepeatProcess || repeat.Amount.Value != 2 {
		t.Fatalf("repeat = %#v", repeat)
	}
}
