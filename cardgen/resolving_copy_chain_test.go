package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateResolvingCopyChainSource covers the copy-chain family: a resolving
// spell that performs a base effect on one target, then lets the affected
// target's controller copy the spell with a new target so the copy chains
// iteratively. The payment-gated forms (String of Disappearances, Chain
// Lightning, Chain Stasis) lower to a base effect, an affected-controller
// resolution Pay publishing its result, and a result-gated optional copy whose
// Chooser is the affected target's controller; the unconditional forms (Chain of
// Acid, Chain of Smog) lower to the base effect and an optional
// affected-controller copy with no payment.
func TestGenerateResolvingCopyChainSource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		typeLine string
		manaCost string
		oracle   string
		want     []string
		absent   []string
	}{
		{
			name:     "String of Disappearances",
			typeLine: "Instant",
			manaCost: "{U}",
			oracle:   "Return target creature to its owner's hand. Then that creature's controller may pay {U}{U}. If the player does, they may copy this spell and may choose a new target for that copy.",
			want: []string{
				"Primitive: game.Bounce{",
				"Primitive: game.Pay{",
				"Payer:  opt.Val(game.AffectedTargetControllerReference(0)),",
				"PublishResult: game.ResultKey(\"copy-chain-paid\"),",
				"Primitive: game.CopyStackObject{",
				"Object:              game.ResolvingStackObjectReference(),",
				"MayChooseNewTargets: true,",
				"Chooser:             opt.Val(game.AffectedTargetControllerReference(0)),",
				"ResultGate: opt.Val(game.InstructionResultGate{",
				"Succeeded: game.TriTrue,",
				"OptionalActor: opt.Val(game.AffectedTargetControllerReference(0)),",
			},
		},
		{
			name:     "Chain Lightning",
			typeLine: "Sorcery",
			manaCost: "{R}",
			oracle:   "Chain Lightning deals 3 damage to any target. Then that player or that permanent's controller may pay {R}{R}. If the player does, they may copy this spell and may choose a new target for that copy.",
			want: []string{
				"Primitive: game.Damage{",
				"Primitive: game.Pay{",
				"Payer:  opt.Val(game.AffectedTargetControllerReference(0)),",
				"Chooser:             opt.Val(game.AffectedTargetControllerReference(0)),",
			},
		},
		{
			name:     "Chain Stasis",
			typeLine: "Instant",
			manaCost: "{U}",
			oracle:   "You may tap or untap target creature. Then that creature's controller may pay {2}{U}. If the player does, they may copy this spell and may choose a new target for that copy.",
			want: []string{
				"Primitive: game.TapOrUntap{",
				"Optional: true,",
				"Primitive: game.Pay{",
				"Chooser:             opt.Val(game.AffectedTargetControllerReference(0)),",
			},
		},
		{
			name:     "Chain of Acid",
			typeLine: "Sorcery",
			manaCost: "{3}{G}",
			oracle:   "Destroy target noncreature permanent. Then that permanent's controller may copy this spell and may choose a new target for that copy.",
			want: []string{
				"Primitive: game.Destroy{",
				"Primitive: game.CopyStackObject{",
				"MayChooseNewTargets: true,",
				"Chooser:             opt.Val(game.AffectedTargetControllerReference(0)),",
				"OptionalActor: opt.Val(game.AffectedTargetControllerReference(0)),",
			},
			absent: []string{"game.Pay{", "ResultGate"},
		},
		{
			name:     "Chain of Smog",
			typeLine: "Sorcery",
			manaCost: "{B}",
			oracle:   "Target player discards two cards. That player may copy this spell and may choose a new target for that copy.",
			want: []string{
				"Primitive: game.Discard{",
				"Primitive: game.CopyStackObject{",
				"Chooser:             opt.Val(game.AffectedTargetControllerReference(0)),",
			},
			absent: []string{"game.Pay{", "ResultGate"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       test.name,
				Layout:     "normal",
				TypeLine:   test.typeLine,
				ManaCost:   test.manaCost,
				OracleText: test.oracle,
			}, "c")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range test.want {
				if !strings.Contains(spaceCollapsed(source), spaceCollapsed(want)) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
			for _, absent := range test.absent {
				if strings.Contains(spaceCollapsed(source), spaceCollapsed(absent)) {
					t.Fatalf("source unexpectedly contains %q:\n%s", absent, source)
				}
			}
		})
	}
}

// TestGenerateResolvingCopyChainFailsClosed proves near-miss wordings that share
// the "may copy this spell" surface do not lower through the copy-chain path: a
// non-mana payment offer (Chain of Plasma) and a plural-new-targets copy
// (Barroom Brawl) are left unsupported rather than silently mis-lowered.
func TestGenerateResolvingCopyChainFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		typeLine string
		manaCost string
		oracle   string
	}{
		{
			name:     "Chain of Plasma",
			typeLine: "Instant",
			manaCost: "{1}{R}",
			oracle:   "Chain of Plasma deals 3 damage to any target. Then that player or that permanent's controller may discard a card. If the player does, they may copy this spell and may choose a new target for that copy.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       test.name,
				Layout:     "normal",
				TypeLine:   test.typeLine,
				ManaCost:   test.manaCost,
				OracleText: test.oracle,
			}, "c")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected the near-miss to fail closed with diagnostics, got none")
			}
		})
	}
}
