package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestConditionTypeSelectionsRuntime(t *testing.T) {
	t.Parallel()
	permanentTypes := []types.Card{types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker}
	tests := []struct {
		name      string
		selection game.Selection
		yes, no   []types.Card
	}{
		{"permanent card land", game.Selection{RequiredTypesAny: permanentTypes}, []types.Card{types.Land}, []types.Card{types.Instant}},
		{"permanent card artifact", game.Selection{RequiredTypesAny: permanentTypes}, []types.Card{types.Artifact}, []types.Card{types.Sorcery}},
		{"permanent card creature", game.Selection{RequiredTypesAny: permanentTypes}, []types.Card{types.Creature}, []types.Card{types.Instant, types.Sorcery}},
		{"permanent card battle", game.Selection{RequiredTypesAny: permanentTypes}, []types.Card{types.Battle}, nil},
		{"permanent card planeswalker", game.Selection{RequiredTypesAny: permanentTypes}, []types.Card{types.Planeswalker}, nil},
		{"permanent card enchantment", game.Selection{RequiredTypesAny: permanentTypes}, []types.Card{types.Enchantment}, nil},
		{"instant", game.Selection{RequiredTypes: []types.Card{types.Instant}}, []types.Card{types.Instant}, []types.Card{types.Sorcery}},
		{"sorcery", game.Selection{RequiredTypes: []types.Card{types.Sorcery}}, []types.Card{types.Sorcery}, []types.Card{types.Instant}},
		{"intersection", game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}}, []types.Card{types.Artifact, types.Creature}, []types.Card{types.Artifact}},
		{"union", game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}}, []types.Card{types.Artifact}, []types.Card{types.Land}},
		{"union creature", game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}}, []types.Card{types.Creature}, []types.Card{types.Land}},
		{"noncreature", game.Selection{ExcludedTypes: []types.Card{types.Creature}}, []types.Card{types.Instant}, []types.Card{types.Artifact, types.Creature}},
		{"nonland", game.Selection{ExcludedTypes: []types.Card{types.Land}}, []types.Card{types.Sorcery}, []types.Card{types.Land}},
		{"required with exclusion", game.Selection{RequiredTypes: []types.Card{types.Artifact}, ExcludedTypes: []types.Card{types.Creature}}, []types.Card{types.Artifact}, []types.Card{types.Artifact, types.Creature}},
		{"intersection or type", game.Selection{AnyOf: []game.Selection{
			{RequiredTypes: []types.Card{types.Artifact, types.Creature}},
			{RequiredTypes: []types.Card{types.Enchantment}},
		}}, []types.Card{types.Enchantment}, []types.Card{types.Creature}},
		{"required and union", game.Selection{RequiredTypes: []types.Card{types.Artifact}, RequiredTypesAny: []types.Card{types.Creature, types.Land}}, []types.Card{types.Artifact, types.Creature}, []types.Card{types.Creature}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			for _, sample := range []struct {
				types []types.Card
				want  bool
			}{{test.yes, true}, {test.no, false}} {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				def := &game.CardDef{CardFace: game.CardFace{Name: "Selection Subject", Types: sample.types}}
				permanent := addCombatPermanent(g, game.Player1, def)
				obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.PermanentTarget(permanent.ObjectID)}}
				condition := opt.Val(game.Condition{Object: opt.Val(game.TargetPermanentReference(0)), ObjectMatches: opt.Val(test.selection)})
				ctx := conditionContext{controller: game.Player1, obj: obj}
				if got := conditionSatisfied(g, ctx, condition); got != sample.want {
					t.Fatalf("live types %v: got %v, want %v", sample.types, got, sample.want)
				}
				snapshot := snapshotPermanent(g, permanent, zone.Battlefield)
				rememberLastKnown(g, &snapshot)
				g.Battlefield = nil
				if got := conditionSatisfied(g, ctx, condition); got != sample.want {
					t.Fatalf("target LKI types %v: got %v, want %v", sample.types, got, sample.want)
				}
				condition.Val.Object = opt.Val(game.EventPermanentReference())
				ctx.obj = nil
				ctx.event = &game.Event{Kind: game.EventPermanentDied, PermanentID: permanent.ObjectID}
				if got := conditionSatisfied(g, ctx, condition); got != sample.want {
					t.Fatalf("event LKI types %v: got %v, want %v", sample.types, got, sample.want)
				}
				cardID := addCardToGraveyard(g, game.Player1, def)
				card, ok := g.GetCardInstance(cardID)
				if !ok {
					t.Fatal("graveyard card instance missing")
				}
				subject := selectionSubject{kind: subjectCard, g: g, card: card}
				if got := matchSelection(&subject, &test.selection); got != sample.want {
					t.Fatalf("card characteristics %v: got %v, want %v", sample.types, got, sample.want)
				}
			}
		})
	}
}
