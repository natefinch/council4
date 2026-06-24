package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// immerwolfProhibitionSource builds a permanent whose static ability prevents the
// controller's non-Human Werewolves from transforming, mirroring Immerwolf's
// "Non-Human Werewolves you control can't transform." rider.
func immerwolfProhibitionSource(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Immerwolf",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Wolf},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantTransform,
				AffectedController: game.ControllerYou,
				AffectedSelection: game.Selection{
					SubtypesAny:     []types.Sub{types.Werewolf},
					ExcludedSubtype: types.Sub("Human"),
				},
			}},
		}},
	}})
}

func transformWerewolf() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Village Day",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sub("Human"), types.Werewolf},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})}, Layout: game.LayoutTransform,

		Back: opt.Val(game.CardFace{Name: "Village Night",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Werewolf},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4})}),
	}
}

func werewolfPermanent(g *game.Game, controller game.PlayerID, face game.FaceIndex) *game.Permanent {
	cardID := addCardInstance(g, controller, transformWerewolf())
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
		Face:           face,
		Transformed:    face == game.FaceBack,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

func transformPermanentInstruction(engine *Engine, g *game.Game, permanent *game.Permanent) {
	obj := &game.StackObject{Controller: permanent.Controller, Targets: []game.Target{game.PermanentTarget(permanent.ObjectID)}}
	resolveInstruction(engine, g, obj, game.Transform{Object: game.TargetPermanentReference(0)}, nil)
}

// TestCantTransformStaticPreventsNightWerewolfTransform verifies the prohibition
// keeps a non-Human (night-side) Werewolf you control on its current face when a
// transform is attempted.
func TestCantTransformStaticPreventsNightWerewolfTransform(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	immerwolfProhibitionSource(g, game.Player1)
	night := werewolfPermanent(g, game.Player1, game.FaceBack)

	transformPermanentInstruction(engine, g, night)

	if night.Face != game.FaceBack {
		t.Fatalf("prohibited night Werewolf transformed to %v, want back", night.Face)
	}
}

// TestCantTransformStaticAllowsHumanWerewolfTransform verifies the "non-Human"
// exclusion: the day-side face still carries Human, so the prohibition does not
// apply and the transform succeeds.
func TestCantTransformStaticAllowsHumanWerewolfTransform(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	immerwolfProhibitionSource(g, game.Player1)
	day := werewolfPermanent(g, game.Player1, game.FaceFront)

	transformPermanentInstruction(engine, g, day)

	if day.Face != game.FaceBack {
		t.Fatalf("Human Werewolf face = %v, want back (exclusion should allow transform)", day.Face)
	}
}

// TestCantTransformStaticIgnoresOpponentWerewolf verifies the "you control" scope:
// an opponent's night Werewolf is unaffected by the controller-scoped prohibition.
func TestCantTransformStaticIgnoresOpponentWerewolf(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	immerwolfProhibitionSource(g, game.Player1)
	enemyNight := werewolfPermanent(g, game.Player2, game.FaceBack)

	transformPermanentInstruction(engine, g, enemyNight)

	if enemyNight.Face != game.FaceFront {
		t.Fatalf("opponent Werewolf face = %v, want front (prohibition should not apply)", enemyNight.Face)
	}
}
