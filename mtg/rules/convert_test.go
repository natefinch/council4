package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestConvertTriggerFlipsFace drives the real trigger→stack→resolve path to prove
// a front-face "convert it" triggered ability (lowered to game.Transform on the
// source permanent) flips the double-faced card to its back face, exactly like the
// transform keyword action it is flavored after.
func TestConvertTriggerFlipsFace(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardInstance(g, game.Player1, convertRobotWithFrontTrigger())
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          game.Player1,
		Controller:     game.Player1,
		Face:           game.FaceFront,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	g.Events = append(g.Events, game.Event{
		Kind:       game.EventSpellCast,
		Controller: game.Player1,
		CardTypes:  []types.Card{types.Instant},
	})

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("putTriggeredAbilitiesOnStack() = false, want front-face convert trigger")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Face != game.FaceFront {
		t.Fatalf("trigger stack object = %+v, want front face", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if permanent.Face != game.FaceBack || !permanent.Transformed {
		t.Fatalf("permanent face/transformed = %v/%v, want back/true", permanent.Face, permanent.Transformed)
	}
	if got := effectivePower(g, permanent); got != 4 {
		t.Fatalf("effective power = %d, want 4 from converted back face", got)
	}
}

func convertRobotWithFrontTrigger() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Convert Front",

		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		TriggeredAbilities: []game.TriggeredAbility{
			{
				Trigger: game.TriggerCondition{
					Type:    game.TriggerWhenever,
					Pattern: game.TriggerPattern{Event: game.EventSpellCast, Controller: game.TriggerControllerYou},
				},
				Content: game.Mode{
					Sequence: []game.Instruction{{Primitive: game.Transform{Object: game.SourcePermanentReference()}}},
				}.Ability(),
			},
		}}, Layout: game.LayoutTransform,

		Back: opt.Val(game.CardFace{Name: "Convert Back", Types: []types.Card{types.Artifact, types.Creature}, Power: opt.Val(game.PT{Value: 4}), Toughness: opt.Val(game.PT{Value: 4})}),
	}
}
