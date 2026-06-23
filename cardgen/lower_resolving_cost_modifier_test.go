package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerResolvingSpellCostModifier(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText      string
		affectedPlayer  game.PlayerRelation
		genericIncrease int
		excludedTypes   []types.Card
		duration        game.EffectDuration
	}{
		"opponents noncreature increase until your next turn": {
			oracleText:      "Noncreature spells your opponents cast cost {2} more to cast until your next turn.",
			affectedPlayer:  game.PlayerOpponent,
			genericIncrease: 2,
			excludedTypes:   []types.Card{types.Creature},
			duration:        game.DurationUntilYourNextTurn,
		},
		"all spells increase leading duration": {
			oracleText:      "Until your next turn, spells your opponents cast cost {1} more to cast.",
			affectedPlayer:  game.PlayerOpponent,
			genericIncrease: 1,
			duration:        game.DurationUntilYourNextTurn,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Resolving Modifier",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			})
			if !face.SpellAbility.Exists {
				t.Fatalf("expected a spell ability, got %#v", face)
			}
			seq := face.SpellAbility.Val.Modes[0].Sequence
			if len(seq) != 1 {
				t.Fatalf("instructions = %d, want 1", len(seq))
			}
			apply, ok := seq[0].Primitive.(game.ApplyRule)
			if !ok {
				t.Fatalf("primitive = %T, want ApplyRule", seq[0].Primitive)
			}
			if apply.Duration != test.duration {
				t.Fatalf("duration = %v, want %v", apply.Duration, test.duration)
			}
			if len(apply.RuleEffects) != 1 {
				t.Fatalf("rule effects = %d, want 1", len(apply.RuleEffects))
			}
			effect := apply.RuleEffects[0]
			if effect.Kind != game.RuleEffectCostModifier {
				t.Fatalf("rule effect kind = %v, want cost modifier", effect.Kind)
			}
			if effect.AffectedPlayer != test.affectedPlayer {
				t.Fatalf("affected player = %v, want %v", effect.AffectedPlayer, test.affectedPlayer)
			}
			if effect.CostModifier.GenericIncrease != test.genericIncrease {
				t.Fatalf("generic increase = %d, want %d", effect.CostModifier.GenericIncrease, test.genericIncrease)
			}
			if len(effect.ExcludedSpellTypes) != len(test.excludedTypes) {
				t.Fatalf("excluded spell types = %#v, want %#v", effect.ExcludedSpellTypes, test.excludedTypes)
			}
			for i, want := range test.excludedTypes {
				if effect.ExcludedSpellTypes[i] != want {
					t.Fatalf("excluded spell type %d = %v, want %v", i, effect.ExcludedSpellTypes[i], want)
				}
			}
		})
	}
}

func TestLowerGraveyardReturnThenCounterChoice(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Reanimate Counter",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target creature card from your graveyard to the battlefield. Put a +1/+1 counter or a loyalty counter on it.",
	})
	if !face.SpellAbility.Exists {
		t.Fatalf("expected a spell ability, got %#v", face)
	}
	seq := face.SpellAbility.Val.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("instructions = %d, want 2", len(seq))
	}
	put, ok := seq[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("instruction[0] = %T, want PutOnBattlefield", seq[0].Primitive)
	}
	if put.PublishLinked == "" {
		t.Fatalf("PutOnBattlefield must publish a link key, got empty")
	}
	add, ok := seq[1].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("instruction[1] = %T, want AddCounter", seq[1].Primitive)
	}
	if add.Object != game.LinkedObjectReference(string(put.PublishLinked)) {
		t.Fatalf("AddCounter.Object = %#v, want linked reference %q", add.Object, put.PublishLinked)
	}
	wantKinds := []counter.Kind{counter.PlusOnePlusOne, counter.Loyalty}
	if len(add.KindChoices) != len(wantKinds) {
		t.Fatalf("KindChoices = %#v, want %#v", add.KindChoices, wantKinds)
	}
	for i, want := range wantKinds {
		if add.KindChoices[i] != want {
			t.Fatalf("KindChoices[%d] = %v, want %v", i, add.KindChoices[i], want)
		}
	}
}
