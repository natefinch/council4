package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: Chaoslace Oracle ID 08842aa3-f923-46e9-a106-f542331e9cc1,
// 4ED printing/revision 476180df-8b88-4ead-b6c5-6ccb3e8a2cfd; Oracle text:
// "Target spell or permanent becomes red. (Its mana symbols remain unchanged.)"
// Scryfall rulings URI https://api.scryfall.com/cards/476180df-8b88-4ead-b6c5-6ccb3e8a2cfd/rulings returned no rulings.
// Official CR effective 2026-11-06: https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.pdf
// Read 105.3 and 202.2 (a color-changing effect replaces the object's colors but not its mana symbols),
// 601.2c/f/h and 608.2b (choose/pay for the spell and check that its target remains legal),
// 611.2a/c (a no-duration resolving effect lasts until end of game and fixes its affected set),
// 613.1e (color-changing effects apply in layer 5), and 400.7a (the effect on a permanent spell
// continues to apply to the permanent it becomes).

func TestChaoslaceTargetPermanentBecomesRed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	targetDef := g.CardInstances[target.CardInstanceID].Def
	targetDef.Colors = []color.Color{color.Blue}
	targetDef.ManaCost = opt.Val(cost.Mana{cost.U})

	chaoslace := addCardToHand(g, game.Player1, authored.Chaoslace.New())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(chaoslace, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)) {
		t.Fatal("casting Chaoslace for {R} targeting a permanent failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := permanentEffectiveColors(g, target); !slices.Equal(got, []color.Color{color.Red}) {
		t.Fatalf("target permanent colors = %v, want [Red]", got)
	}
	if !slices.Equal(targetDef.ManaCost.Val, cost.Mana{cost.U}) {
		t.Fatalf("target permanent mana cost = %v, want {U}", targetDef.ManaCost.Val)
	}
}

func TestChaoslaceTargetSpellBecomesRed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := pushStackSpell(g, game.Player2, "Blue instant", types.Instant)
	targetDef := g.CardInstances[target.SourceID].Def
	targetDef.Colors = []color.Color{color.Blue}
	targetDef.ManaCost = opt.Val(cost.Mana{cost.U})

	chaoslace := addCardToHand(g, game.Player1, authored.Chaoslace.New())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(chaoslace, []game.Target{game.StackObjectTarget(target.ID)}, 0, nil)) {
		t.Fatal("casting Chaoslace for {R} targeting a spell failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	got, ok := stackObjectColors(g, target)
	if !ok {
		t.Fatal("target spell characteristics were not available after Chaoslace resolved")
	}
	if !slices.Equal(targetDef.ManaCost.Val, cost.Mana{cost.U}) {
		t.Fatalf("target spell mana cost = %v, want {U}", targetDef.ManaCost.Val)
	}
	if !slices.Equal(got, []color.Color{color.Red}) {
		t.Fatalf("target spell colors = %v, want [Red]", got)
	}
}
