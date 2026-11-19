package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned Quirion Dryad Oracle ID f42b6510-7322-4733-85ea-7ae6bd6ca8fc,
// printing a20d20f7-bc4d-42fa-9f5b-5bb330eb7f38 (M21 #198), corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. The pinned
// and live Scryfall Oracle text (retrieved 2026-11-19) is: "Whenever you cast a
// spell that's white, blue, black, or red, put a +1/+1 counter on this creature."
// Both WotC rulings retrieved 2026-11-19 from
// https://api.scryfall.com/cards/a20d20f7-bc4d-42fa-9f5b-5bb330eb7f38/rulings
// (two entries, both 2020-06-23) were reviewed: a multicolored spell triggers
// only once; the trigger resolves before its spell and still resolves if that
// spell is countered. Raw Scryfall responses are retained under
// .cardwork/quirion-dryad/. Wizards Comprehensive Rules effective 2026-11-06, official
// TXT https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// linked from https://magic.wizards.com/en/rules; verified local copy
// .cardwork/quirion-dryad/MagicCompRules-20261106.txt (SHA-256
// 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829). Read
// 202.2-202.2f, 601.2-601.2i, and 603.1-603.3: match the spell's actual stack
// colors, trigger once per cast event, then put the trigger above the spell.
func TestQuirionDryadCastColorTrigger(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	dryadDef, err := cards.NewDefaultRegistry().LookupOracleID(authored.QuirionDryad.OracleID)
	if err != nil {
		t.Fatalf("look up registered Quirion Dryad: %v", err)
	}
	dryadID := addCardToHand(g, game.Player1, dryadDef)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(dryadID, nil, 0, nil)) {
		t.Fatal("could not cast Quirion Dryad for {1}{G}")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana after casting Quirion Dryad = %d, want 0", got)
	}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Quirion Dryad triggered from its own spell before entering the battlefield")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	var dryad *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == dryadID {
			dryad = permanent
			break
		}
	}
	if dryad == nil {
		t.Fatal("Quirion Dryad did not enter the battlefield")
	}

	lifeBeforeSpell := g.Players[game.Player1].Life
	spell := &game.CardDef{CardFace: game.CardFace{
		Name:     "White test instant",
		ManaCost: opt.Val(cost.Mana{cost.W}),
		Colors:   []color.Color{color.White},
		Types:    []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
			Amount: game.Fixed(1), Player: game.ControllerReference(),
		}}}}.Ability()),
	}}
	spellID := addCardToHand(g, game.Player1, spell)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("could not cast the white instant for {W}")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana after casting the white instant = %d, want 0", got)
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Quirion Dryad did not trigger for a white spell")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack after casting the white instant = %d objects, want spell and trigger", got)
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if got := dryad.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("Quirion Dryad has %d +1/+1 counters, want 1", got)
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if got := g.Players[game.Player1].Life; got != lifeBeforeSpell+1 {
		t.Fatalf("life after the white instant resolves = %d, want %d", got, lifeBeforeSpell+1)
	}
}
