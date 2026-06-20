package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func toxicDelugeDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Toxic Deluge",
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B}),
		Types:    []types.Card{types.Sorcery},
		AdditionalCosts: []cost.Additional{{
			Kind:        cost.AdditionalPayLife,
			Text:        "pay X life",
			AmountFromX: true,
		}},
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{{Primitive: toxicDelugeContinuous()}},
		}.Ability()),
	}}
}

func toxicDelugeContinuous() game.ApplyContinuous {
	dynamic := opt.Val(game.DynamicAmount{Kind: game.DynamicAmountX, Multiplier: -1})
	return game.ApplyContinuous{
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:                 game.LayerPowerToughnessModify,
			Group:                 game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			PowerDeltaDynamic:     dynamic,
			ToughnessDeltaDynamic: dynamic,
		}},
		Duration: game.DurationUntilEndOfTurn,
	}
}

func TestCastPayXLifeSpellChoicesAndPayment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, toxicDelugeDef())
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	setSorcerySpeedTurn(g, game.Player1)

	legal := engine.legalActions(g, game.Player1)
	for _, x := range []int{0, 39, 40} {
		if !containsAction(legal, action.CastSpell(spellID, nil, x, nil)) {
			t.Fatalf("pay-X-life spell was not legal for X=%d at 40 life", x)
		}
	}
	if containsAction(legal, action.CastSpell(spellID, nil, 41, nil)) {
		t.Fatal("pay-X-life spell was legal for X=41 at 40 life")
	}

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 39, nil)) {
		t.Fatal("applyAction(cast pay-X-life spell with X=39) = false")
	}
	if got := g.Players[game.Player1].Life; got != 1 {
		t.Fatalf("life after payment = %d, want 1", got)
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.XValue != 39 {
		t.Fatalf("stack object = %#v, want chosen X=39", obj)
	}
}

func TestCastPayXLifeAllowsPayingEntireLifeTotal(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].Life = 3
	spellID := addCardToHand(g, game.Player1, toxicDelugeDef())
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	setSorcerySpeedTurn(g, game.Player1)

	act := action.CastSpell(spellID, nil, 3, nil)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("paying exactly the current life total must be a legal payment")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("cast paying exactly the current life total failed")
	}
	if got := g.Players[game.Player1].Life; got != 0 {
		t.Fatalf("life after payment = %d, want 0", got)
	}
}

func TestCastPayXLifeWithZeroPaysNoLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, toxicDelugeDef())
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	setSorcerySpeedTurn(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast pay-X-life spell with X=0 failed")
	}
	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("life after X=0 payment = %d, want 40", got)
	}
	obj, _ := g.Stack.Peek()
	if obj.XValue != 0 {
		t.Fatalf("stack X value = %d, want 0", obj.XValue)
	}
}

func TestAlternativeSpellCostRetainsPayXLifeCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	card := toxicDelugeDef()
	cardID := addCardToHand(g, game.Player1, card)
	_, _, ok := paymentOrch.paySpellCosts(g, payment.SpellRequest{
		PlayerID:   game.Player1,
		CardID:     cardID,
		SourceZone: zone.Hand,
		Card:       card,
		XValue:     4,
		Alternative: opt.Val(cost.Alternative{
			Label:    "Test alternative",
			ManaCost: opt.Val(cost.Mana{}),
		}),
	})
	if !ok {
		t.Fatal("alternative spell cost with pay X life was not payable")
	}
	if got := g.Players[game.Player1].Life; got != 36 {
		t.Fatalf("life after alternative cast payment = %d, want 36", got)
	}
}

func TestToxicDelugeResolutionSnapshotsGroupAndXUntilCleanup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCombatPermanent(g, game.Player1, creatureWithPT("First", 6, 6))
	second := addCombatPermanent(g, game.Player2, creatureWithPT("Second", 7, 7))
	addEffectSpellToStack(g, game.Player1, toxicDelugeContinuous(), nil)
	obj, _ := g.Stack.Peek()
	obj.XValue = 3

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, first); got != 3 {
		t.Fatalf("first creature power = %d, want 3", got)
	}
	if got := effectivePower(g, second); got != 4 {
		t.Fatalf("second creature power = %d, want 4", got)
	}
	for _, effect := range g.ContinuousEffects {
		if effect.PowerDelta != -3 || effect.ToughnessDelta != -3 ||
			effect.PowerDeltaDynamic.Exists || effect.ToughnessDeltaDynamic.Exists {
			t.Fatalf("runtime effect = %#v, want snapshotted -3/-3", effect)
		}
	}
	later := addCombatPermanent(g, game.Player3, creatureWithPT("Later", 5, 5))
	if got := effectivePower(g, later); got != 5 {
		t.Fatalf("later entrant power = %d, want unaffected 5", got)
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if got := effectivePower(g, first); got != 6 {
		t.Fatalf("first creature power after cleanup = %d, want 6", got)
	}
	if got := effectivePower(g, second); got != 7 {
		t.Fatalf("second creature power after cleanup = %d, want 7", got)
	}
}

func TestToxicDelugeSpellCopyUsesChosenXWithoutRepayingLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatPermanent(g, game.Player2, creatureWithPT("Copy Victim", 8, 8))
	sourceID := addEffectSpellToStack(g, game.Player1, toxicDelugeContinuous(), nil)
	original, _ := g.Stack.Peek()
	original.XValue = 4
	spell := g.CardInstances[sourceID].Def
	createStormCopies(g, original, spell, 1)

	copyObj, _ := g.Stack.Peek()
	if !copyObj.Copy || copyObj.XValue != 4 {
		t.Fatalf("copy = %#v, want copied X=4", copyObj)
	}
	before := g.Players[game.Player1].Life
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := effectivePower(g, creature); got != 4 {
		t.Fatalf("creature power after copy = %d, want 4", got)
	}
	if got := g.Players[game.Player1].Life; got != before {
		t.Fatalf("life after resolving copy = %d, want unchanged %d", got, before)
	}
}

func creatureWithPT(name string, power, toughness int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
	}}
}
