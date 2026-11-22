package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: Petrified Wood-Kin, Guildpact #91 (Oracle ID 3bc9649c-4350-4c6a-8441-dc76a5f833af,
// pinned revision a56fbf32-f08f-4882-b8d4-41ccb60cba63). Oracle: “This spell can't be countered.
// Bloodthirst X (This creature enters with X +1/+1 counters on it, where X is the damage dealt to your opponents this turn.)
// Protection from instants”. Scryfall (https://api.scryfall.com/cards/a56fbf32-f08f-4882-b8d4-41ccb60cba63/rulings,
// retrieved 2026-11-22) returned two WotC rulings dated 2006-02-01: protection covers instant
// spells/sources and only works on the battlefield. Read the official 2026-11-06 Comprehensive
// Rules (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.docx), 102.1–4, 109.5,
// 113.6g–h, 120.1–4, 614.12, 702.16a–h, 702.54a–c, and 800.4i. The fixtures exercise only
// Bloodthirst X: at entry count actual player damage to the entrant's controller's opponents,
// including damage that resulted in poison rather than life loss.
func TestBloodthirstXCountsActualOpponentDamageAtEntry(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	log := &TurnLog{}
	infect := addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.Infect)
	markPlayerCombatDamage(g, infect, game.Player2, 2, log)
	if player := g.Players[game.Player2]; player.PoisonCounters != 2 || player.Life != 40 {
		t.Fatalf("infect damage gave %d poison and left %d life, want 2 poison and 40 life", player.PoisonCounters, player.Life)
	}
	source := addCombatCreaturePermanentWithPower(g, game.Player4, 3)
	dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player4, game.Player2, 1, false)

	setMainPhasePriority(g, game.Player1)
	g.Turn.Phase = game.PhasePostcombatMain
	cardID := addCardToHand(g, game.Player1, bloodthirstXFixture())
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("could not cast the Bloodthirst X fixture")
	}

	// Damage from another player's source while the creature spell is still on
	// the stack counts, but only the amount that survives prevention.
	g.PreventionShields = []game.PreventionShield{{ID: g.IDGen.Next(), Player: game.Player3, Amount: 1}}
	if dealt := dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player4, game.Player3, 3, false); dealt != 2 {
		t.Fatalf("damage after prevention = %d, want 2", dealt)
	}
	engine.resolveTopOfStack(g, log)
	permanent, ok := findPermanentByCardID(g, cardID)
	if !ok {
		t.Fatal("creature did not enter")
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 5 {
		t.Fatalf("Bloodthirst X counters = %d, want 5 (2 combat infect + 1 and 2 noncombat)", got)
	}
}

func TestBloodthirstXOnNoncastEntryUsesEnteringControllerAndCurrentTurn(t *testing.T) {
	for _, tc := range []struct {
		name     string
		nextTurn bool
		want     int
	}{
		{name: "same turn", want: 3},
		{name: "next turn", nextTurn: true, want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatCreaturePermanentWithPower(g, game.Player4, 4)
			dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player4, game.Player3, 3, false)
			if tc.nextTurn {
				engine.advanceToNextTurn(g)
			}
			dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player4, game.Player2, 4, false)
			target := addCombatCreaturePermanentWithPower(g, game.Player3, 4)
			dealPermanentDamage(g, source.CardInstanceID, source.ObjectID, game.Player4, target, 1, false)
			loseLife(g, game.Player3, 2)

			// Player2 gets a creature owned by Player3, on a turn whose active
			// player can differ from both. Owner/active-player scope is wrong.
			cardID := addCardToGraveyard(g, game.Player3, bloodthirstXFixture())
			resolveInstruction(engine, g, &game.StackObject{
				Controller: game.Player2,
				Targets:    []game.Target{currentCardTarget(t, g, cardID)},
			}, game.PutOnBattlefield{
				Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
				Recipient: opt.Val(game.ControllerReference()),
			}, &TurnLog{})
			permanent, ok := findPermanentByCardID(g, cardID)
			if !ok || permanent.Controller != game.Player2 {
				t.Fatal("creature did not enter under Player2 from Player3's graveyard")
			}
			if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != tc.want {
				t.Fatalf("Bloodthirst X counters = %d, want %d; self damage, permanent damage, and direct life loss don't count", got, tc.want)
			}
		})
	}
}

func bloodthirstXFixture() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:                 "Bloodthirst X fixture",
		Types:                []types.Card{types.Creature},
		Power:                opt.Val(game.PT{Value: 2}),
		Toughness:            opt.Val(game.PT{Value: 2}),
		ReplacementAbilities: []game.ReplacementAbility{game.BloodthirstXReplacement("Bloodthirst X")},
	}}
}
