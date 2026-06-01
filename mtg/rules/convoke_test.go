package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestConvokeMakesGenericSpellPayableAndTapsCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, convokeSpell(mana.Cost{mana.GenericMana(2)}))
	first := addCombatCreaturePermanent(g, game.Player1)
	second := addCombatCreaturePermanent(g, game.Player1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("convoke spell cast failed")
	}
	if !first.Tapped || !second.Tapped {
		t.Fatal("convoke did not tap creatures for generic mana")
	}
}

func TestConvokeCanUseSummoningSickCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, convokeSpell(mana.Cost{mana.GenericMana(1)}))
	creature := addCombatCreaturePermanent(g, game.Player1)
	creature.SummoningSick = true
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("convoke spell cast with summoning-sick creature failed")
	}
	if !creature.Tapped {
		t.Fatal("summoning-sick creature was not tapped for convoke")
	}
}

func TestConvokeDoesNotTapCreaturesWhenManaCanPay(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, convokeSpell(mana.Cost{mana.GenericMana(1)}))
	creature := addCombatCreaturePermanent(g, game.Player1)
	forest := addBasicLandPermanent(g, game.Player1, game.LandSubtypeForest)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("convoke spell cast failed")
	}
	if creature.Tapped {
		t.Fatal("convoke tapped creature even though mana could pay")
	}
	if !forest.Tapped {
		t.Fatal("mana source was not used for normal payment")
	}
}

func TestConvokeIgnoresTappedCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, convokeSpell(mana.Cost{mana.GenericMana(1)}))
	creature := addCombatCreaturePermanent(g, game.Player1)
	creature.Tapped = true
	setMainPhasePriority(g, game.Player1)

	if engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("convoke spell cast with only tapped creature, want failure")
	}
}

func TestConvokePaymentPlanValidityChecksConvokeTapsWithoutManaTaps(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, convokeSpell(mana.Cost{mana.GenericMana(1)}))
	creature := addCombatCreaturePermanent(g, game.Player1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("convoke-only spell cast failed")
	}
	if !creature.Tapped {
		t.Fatal("convoke-only payment did not tap creature")
	}
}

func TestConvokePaysColoredSymbolsWithCreatureColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, convokeSpell(mana.Cost{mana.ColoredMana(mana.Green), mana.ColoredMana(mana.Green)}))
	first := addCombatPermanent(g, game.Player1, greenConvokeCreature())
	second := addCombatPermanent(g, game.Player1, greenConvokeCreature())
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("convoke spell with colored symbols failed")
	}
	if !first.Tapped || !second.Tapped {
		t.Fatal("green creatures were not tapped for colored convoke symbols")
	}
}

func TestConvokeDoesNotDoubleUseManaCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, convokeSpell(mana.Cost{mana.GenericMana(1), mana.ColoredMana(mana.Green)}))
	manaCreature := addCombatPermanent(g, game.Player1, greenManaCreature())
	otherCreature := addCombatCreaturePermanent(g, game.Player1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("convoke spell cast failed")
	}
	if !otherCreature.Tapped {
		t.Fatal("non-mana creature was not tapped for convoke")
	}
	if !manaCreature.Tapped {
		t.Fatal("mana creature was not tapped for green mana")
	}
}

func setMainPhasePriority(g *game.Game, playerID game.PlayerID) {
	g.Turn.ActivePlayer = playerID
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = playerID
}

func convokeSpell(cost mana.Cost) *game.CardDef {
	return &game.CardDef{
		Name:     "Convoke Spell",
		Types:    []game.CardType{game.TypeSorcery},
		ManaCost: optCost(cost),
		Abilities: []game.AbilityDef{
			{Kind: game.StaticAbility, Keywords: []game.Keyword{game.Convoke}},
			{Kind: game.SpellAbility},
		},
	}
}

func greenManaCreature() *game.CardDef {
	pt := game.PT{Value: 1}
	return &game.CardDef{
		Name:      "Green Mana Creature",
		Types:     []game.CardType{game.TypeCreature},
		Power:     optPT(pt),
		Toughness: optPT(pt),
		Abilities: []game.AbilityDef{{
			Kind:          game.ActivatedAbility,
			IsManaAbility: true,
			AdditionalCosts: []game.AdditionalCost{{
				Kind: game.AdditionalCostTap,
				Text: "{T}",
			}},
			Effects: []game.Effect{{Type: game.EffectAddMana, ManaColor: mana.Green, Amount: 1}},
		}},
	}
}

func greenConvokeCreature() *game.CardDef {
	pt := game.PT{Value: 1}
	return &game.CardDef{
		Name:      "Green Convoke Creature",
		Types:     []game.CardType{game.TypeCreature},
		Colors:    []mana.Color{mana.Green},
		Power:     optPT(pt),
		Toughness: optPT(pt),
	}
}
