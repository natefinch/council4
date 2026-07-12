package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestNextChosenTypeFlashPermissionIsConsumedByMatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{})
	source.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryTypeChoiceKey: {Kind: game.ResolutionChoiceSubtype, Subtype: types.Elf},
	}
	resolveInstruction(NewEngine(nil), g, &game.StackObject{
		Controller:   game.Player1,
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
	}, game.ApplyRule{
		RuleEffects: []game.RuleEffect{{
			Kind:                   game.RuleEffectCastSpellsAsThoughFlash,
			AffectedPlayer:         game.PlayerYou,
			AppliesToNextSpellOnly: true,
			SpellChosenSubtypeFrom: game.EntryTypeChoiceKey,
		}},
		Duration: game.DurationThisTurn,
	}, &TurnLog{})
	if len(g.RuleEffects) != 1 ||
		len(g.RuleEffects[0].SpellSubtypes) != 1 ||
		g.RuleEffects[0].SpellSubtypes[0] != types.Elf ||
		g.RuleEffects[0].SpellChosenSubtypeFrom != "" {
		t.Fatalf("snapshotted effect = %#v, want Elf subtype", g.RuleEffects)
	}
	g.Battlefield = nil
	elf := &game.CardDef{CardFace: game.CardFace{
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Elf},
	}}
	goblin := &game.CardDef{CardFace: game.CardFace{
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Goblin},
	}}
	if !playerCanCastAsThoughFlash(g, game.Player1, elf) ||
		playerCanCastAsThoughFlash(g, game.Player1, goblin) {
		t.Fatal("chosen-type flash permission matched the wrong spells")
	}
	goblinID := addCardToHand(g, game.Player1, goblin)
	consumeNextSpellAsThoughFlashEffects(g, &game.StackObject{
		Controller: game.Player1,
		SourceID:   goblinID,
	})
	if len(g.RuleEffects) != 1 {
		t.Fatal("nonmatching spell consumed chosen-type flash permission")
	}
	elfID := addCardToHand(g, game.Player1, elf)
	consumeNextSpellAsThoughFlashEffects(g, &game.StackObject{
		Controller: game.Player1,
		SourceID:   elfID,
	})
	if len(g.RuleEffects) != 0 {
		t.Fatal("matching spell did not consume chosen-type flash permission")
	}
}
