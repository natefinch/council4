package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestConvokeMakesGenericSpellPayableAndTapsCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, convokeSpell(cost.Mana{cost.O(2)}))
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
	spellID := addCardToHand(g, game.Player1, convokeSpell(cost.Mana{cost.O(1)}))
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
	spellID := addCardToHand(g, game.Player1, convokeSpell(cost.Mana{cost.O(1)}))
	creature := addCombatCreaturePermanent(g, game.Player1)
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
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
	spellID := addCardToHand(g, game.Player1, convokeSpell(cost.Mana{cost.O(1)}))
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
	spellID := addCardToHand(g, game.Player1, convokeSpell(cost.Mana{cost.O(1)}))
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
	spellID := addCardToHand(g, game.Player1, convokeSpell(cost.Mana{cost.G, cost.G}))
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
	spellID := addCardToHand(g, game.Player1, convokeSpell(cost.Mana{cost.O(1), cost.G}))
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

func convokeSpell(manaCost cost.Mana) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Convoke Spell",
		Types:           []types.Card{types.Sorcery},
		ManaCost:        opt.Val(manaCost),
		SpellAbility:    opt.Val(game.AbilityContent{}),
		StaticAbilities: []game.StaticAbility{game.ConvokeStaticBody}},
	}
}

func greenManaCreature() *game.CardDef {
	pt := game.PT{Value: 1}
	return &game.CardDef{CardFace: game.CardFace{Name: "Green Mana Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		ManaAbilities: []game.ManaAbility{{
			AdditionalCosts: []cost.Additional{{
				Kind: cost.AdditionalTap,
				Text: "{T}",
			}},
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.AddMana{ManaColor: mana.G, Amount: game.Fixed(1)}}},
			}.Ability(),
		}}},
	}
}

func greenConvokeCreature() *game.CardDef {
	pt := game.PT{Value: 1}
	return &game.CardDef{CardFace: game.CardFace{Name: "Green Convoke Creature",
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.Green},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt)},
	}
}
