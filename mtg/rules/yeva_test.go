package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCastAsThoughFlashFiltersSpellColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCastSpellsAsThoughFlash,
			AffectedPlayer: game.PlayerYou,
			SpellTypes:     []types.Card{types.Creature},
			SpellColors:    []color.Color{color.Green},
		}}}},
	}})
	greenCreature := &game.CardDef{CardFace: game.CardFace{
		Types:  []types.Card{types.Creature},
		Colors: []color.Color{color.Green},
	}}
	blueCreature := &game.CardDef{CardFace: game.CardFace{
		Types:  []types.Card{types.Creature},
		Colors: []color.Color{color.Blue},
	}}
	greenSorcery := &game.CardDef{CardFace: game.CardFace{
		Types:  []types.Card{types.Sorcery},
		Colors: []color.Color{color.Green},
	}}
	if !playerCanCastAsThoughFlash(g, game.Player1, greenCreature) {
		t.Fatal("green creature was not granted flash timing")
	}
	if playerCanCastAsThoughFlash(g, game.Player1, blueCreature) {
		t.Fatal("blue creature was granted green-only flash timing")
	}
	if playerCanCastAsThoughFlash(g, game.Player1, greenSorcery) {
		t.Fatal("green sorcery was granted creature-only flash timing")
	}
}
