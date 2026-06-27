package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// chosenColorProtectionAnthem mirrors the static ability the executable backend
// generates for "All Slivers have protection from the chosen color." (Ward
// Sliver): a single LayerAbility continuous effect whose grant carries the
// chosen-color protection marker in AddAbilities. The grant resolves to a
// concrete color from the source's entry-time color choice.
func chosenColorProtectionAnthem(g *game.Game, controller game.PlayerID, chosen mana.Color) *game.Permanent {
	protection := game.ProtectionFromChosenColorStaticAbility()
	permanent := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Chosen Color Anthem",
		Types: []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.ObjectControlledGroup(
					game.SourcePermanentReference(),
					game.Selection{RequiredTypes: []types.Card{types.Creature}},
				),
				AddAbilities: []game.Ability{&protection},
			}},
		}},
	}})
	if chosen != "" {
		permanent.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
			game.EntryColorChoiceKey: {Kind: game.ResolutionChoiceMana, Color: chosen},
		}
	}
	return permanent
}

func bearPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
}

// TestChosenColorStaticProtectionResolvesFromEntryChoice verifies that a static
// "protection from the chosen color" grant resolves to protection from the
// concrete color the source chose as it entered: a controlled creature gains
// protection from the chosen color but not from any other color.
func TestChosenColorStaticProtectionResolvesFromEntryChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	chosenColorProtectionAnthem(g, game.Player1, mana.R)
	mine := bearPermanent(g, game.Player1)

	if !permanentHasGrantedProtectionFromColor(g, mine, color.Red) {
		t.Fatal("controlled creature lacks granted protection from the chosen color (red)")
	}
	if permanentHasGrantedProtectionFromColor(g, mine, color.Blue) {
		t.Fatal("controlled creature should not have protection from an unchosen color (blue)")
	}
}

// TestChosenColorStaticProtectionFailsClosedWithoutChoice verifies that when the
// granting source recorded no entry-time color choice the chosen-color marker is
// left in place and resolves to no concrete color, so the grant protects from
// nothing rather than defaulting to a color.
func TestChosenColorStaticProtectionFailsClosedWithoutChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	chosenColorProtectionAnthem(g, game.Player1, "")
	mine := bearPermanent(g, game.Player1)

	for _, c := range []color.Color{color.White, color.Blue, color.Black, color.Red, color.Green} {
		if permanentHasGrantedProtectionFromColor(g, mine, c) {
			t.Fatalf("creature gained protection from %v despite no recorded color choice", c)
		}
	}
}
