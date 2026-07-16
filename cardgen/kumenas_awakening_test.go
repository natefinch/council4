package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// TestLowerKumenasAwakeningConditionalDraw proves Kumena's Awakening lowers fully
// with no card-name logic: the Ascend keyword becomes the reusable
// AscendStaticBody static, and the upkeep body becomes one triggered ability
// whose two complementary-gated draw instructions model the "instead" conditional
// — without the city's blessing every player draws one card, and with it only the
// controller draws one. Exactly one instruction resolves because the gates are
// the controller's live city's-blessing flag and its negation.
func TestLowerKumenasAwakeningConditionalDraw(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Kumena's Awakening",
		Layout:   "normal",
		TypeLine: "Enchantment",
		OracleText: "Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)\n" +
			"At the beginning of your upkeep, each player draws a card. If you have the city's blessing, instead only you draw a card.",
	})

	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	if got := face.StaticAbilities[0].VarName; got != "game.AscendStaticBody" {
		t.Fatalf("ascend static VarName = %q, want game.AscendStaticBody", got)
	}

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Pattern.Event != game.EventBeginningOfStep || trigger.Trigger.Pattern.Step != game.StepUpkeep {
		t.Fatalf("trigger pattern = %+v, want beginning of upkeep step", trigger.Trigger.Pattern)
	}
	if len(trigger.Content.Modes) != 1 {
		t.Fatalf("modes = %d, want 1", len(trigger.Content.Modes))
	}

	got := trigger.Content.Modes[0].Sequence
	want := []game.Instruction{
		{
			Primitive: game.Draw{
				Amount:      game.Fixed(1),
				PlayerGroup: game.AllPlayersReference(),
			},
			Condition: opt.Val(game.EffectCondition{
				Condition: opt.Val(game.Condition{
					Text:                      "If you have the city's blessing",
					Negate:                    true,
					ControllerHasCityBlessing: true,
				}),
			}),
		},
		{
			Primitive: game.Draw{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			},
			Condition: opt.Val(game.EffectCondition{
				Condition: opt.Val(game.Condition{
					Text:                      "If you have the city's blessing",
					ControllerHasCityBlessing: true,
				}),
			}),
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("upkeep sequence =\n%#v\nwant\n%#v", got, want)
	}
}
