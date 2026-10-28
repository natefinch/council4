package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestStaticGrantRecordsOnlyApplicableOrigins(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatPermanent(g, game.Player1, frameDiffCreature("First recipient", 2, 2))
	second := addCombatPermanent(g, game.Player2, frameDiffCreature("Second recipient", 2, 2))
	trigger := &game.TriggeredAbility{
		MaxTriggersPerTurn: 1,
		Trigger:            game.TriggerCondition{Pattern: game.TriggerPattern{Event: game.EventCardDrawn}},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability(),
	}
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Attached grant", Types: []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerAbility, Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
			AddAbilities: []game.Ability{trigger},
		}}}},
	}})
	check := func(recipient, other *game.Permanent) {
		t.Helper()
		untouched := effectivePermanentValues(g, other)
		if len(untouched.abilities) != 0 || len(untouched.triggerGrantOrigins) != 0 {
			t.Fatalf("inapplicable grant retained %d abilities or %d provenance records",
				len(untouched.abilities), len(untouched.triggerGrantOrigins))
		}
		granted := effectivePermanentValues(g, recipient)
		if len(granted.abilities) != 1 || len(granted.triggerGrantOrigins) != 1 {
			t.Fatal("applicable grant lost its ability or provenance")
		}
		origin := granted.triggerOrigins[0]
		if !origin.Granted || origin.SourceID != source.ObjectID || origin.OriginID != source.CardInstanceID {
			t.Fatalf("applicable grant lost physical source identity: %+v", origin)
		}
	}
	source.AttachedTo = opt.Val(first.ObjectID)
	check(first, second)
	source.AttachedTo = opt.Val(second.ObjectID)
	check(second, first)
}
