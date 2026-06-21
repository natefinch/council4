package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func reanimateInstructions(gain bool) []game.Instruction {
	const (
		linkedKey game.LinkedKey = "reanimated-card"
		resultKey game.ResultKey = "reanimation-move"
	)
	amount := game.Dynamic(game.DynamicAmount{
		Kind:       game.DynamicAmountObjectManaValue,
		Multiplier: 1,
		Object:     game.LinkedObjectReference(string(linkedKey)),
	})
	var life game.Primitive = game.LoseLife{Amount: amount, Player: game.ControllerReference()}
	if gain {
		life = game.GainLife{Amount: amount, Player: game.ControllerReference()}
	}
	return []game.Instruction{
		{
			Primitive: game.PutOnBattlefield{
				Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
				Recipient:     opt.Val(game.ControllerReference()),
				PublishLinked: linkedKey,
			},
			PublishResult: resultKey,
		},
		{
			Primitive: life,
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       resultKey,
				Succeeded: game.TriTrue,
			}),
		},
	}
}

func addReanimateSpell(g *game.Game, target game.Target, gain bool) id.ID {
	sourceID := addInstructionSpellToStackForController(g, game.Player1, reanimateInstructions(gain), []game.Target{target})
	card, _ := g.GetCardInstance(sourceID)
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection: opt.Val(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		}),
	}}
	return sourceID
}

func reanimatedPermanent(g *game.Game, cardID id.ID) (*game.Permanent, bool) {
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return permanent, true
		}
	}
	return nil, false
}

func TestReanimateMovesTargetUnderCasterControlAndLosesManaValue(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	oldPermanent := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:     "Returned Creature",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(5)}),
	}})
	cardID := oldPermanent.CardInstanceID
	oldObjectID := oldPermanent.ObjectID
	if !movePermanentToZone(g, oldPermanent, zone.Graveyard) {
		t.Fatal("moving target to graveyard")
	}
	addReanimateSpell(g, currentCardTarget(t, g, cardID), false)
	before := g.Players[game.Player1].Life

	engine.resolveTopOfStack(g, &TurnLog{})

	permanent, ok := reanimatedPermanent(g, cardID)
	if !ok {
		t.Fatal("target card was not put onto the battlefield")
	}
	if permanent.ObjectID == oldObjectID {
		t.Fatal("reanimated permanent reused its old object identity")
	}
	if permanent.Controller != game.Player1 || permanent.Owner != game.Player2 {
		t.Fatalf("controller/owner = %v/%v, want Player1/Player2", permanent.Controller, permanent.Owner)
	}
	if got := before - g.Players[game.Player1].Life; got != 5 {
		t.Fatalf("life lost = %d, want 5", got)
	}
}

func TestReanimateIllegalTargetDoesNotApplyLifeRider(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:     "Gone Creature",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(4)}),
	}})
	addReanimateSpell(g, currentCardTarget(t, g, cardID), false)
	if !moveCardBetweenZones(g, game.Player2, cardID, zone.Graveyard, zone.Exile) {
		t.Fatal("moving target before resolution")
	}
	before := g.Players[game.Player1].Life
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if got := g.Players[game.Player1].Life; got != before {
		t.Fatalf("caster life = %d, want unchanged %d", got, before)
	}
	if len(log.Resolves) != 1 || log.Resolves[0].Result != "countered by rules" {
		t.Fatalf("resolve log = %+v, want countered by rules", log.Resolves)
	}
}

func TestReanimateDestinationReplacementSkipsLifeRider(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:     "Diverted Creature",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(6)}),
	}})
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID:            g.IDGen.Next(),
		Description:   "exile entering graveyard cards instead",
		MatchEvent:    game.EventZoneChanged,
		MatchFromZone: true,
		FromZone:      zone.Graveyard,
		MatchToZone:   true,
		ToZone:        zone.Battlefield,
		ReplaceToZone: zone.Exile,
	})
	addReanimateSpell(g, currentCardTarget(t, g, cardID), false)
	before := g.Players[game.Player1].Life

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player2].Exile.Contains(cardID) {
		t.Fatal("replacement did not divert target card to exile")
	}
	if _, ok := reanimatedPermanent(g, cardID); ok {
		t.Fatal("diverted card became a permanent")
	}
	if got := g.Players[game.Player1].Life; got != before {
		t.Fatalf("caster life = %d, want unchanged %d", got, before)
	}
}

func TestReanimateManaValueUsesFrontFaceAndXAsZero(t *testing.T) {
	tests := []struct {
		name string
		def  *game.CardDef
		want int
	}{
		{
			name: "X is zero",
			def: &game.CardDef{CardFace: game.CardFace{
				Name:     "Variable Creature",
				Types:    []types.Card{types.Creature},
				ManaCost: opt.Val(cost.Mana{cost.X, cost.O(2), cost.B}),
			}},
			want: 3,
		},
		{
			name: "modal double faced card uses front",
			def: &game.CardDef{
				CardFace: game.CardFace{
					Name:     "Front Creature",
					Types:    []types.Card{types.Creature},
					ManaCost: opt.Val(cost.Mana{cost.O(2)}),
				},
				Layout: game.LayoutModalDFC,
				Back: opt.Val(game.CardFace{
					Name:     "Back Creature",
					Types:    []types.Card{types.Creature},
					ManaCost: opt.Val(cost.Mana{cost.O(7)}),
				}),
			},
			want: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			cardID := addCardToGraveyard(g, game.Player2, test.def)
			addReanimateSpell(g, currentCardTarget(t, g, cardID), false)
			before := g.Players[game.Player1].Life

			engine.resolveTopOfStack(g, &TurnLog{})

			if got := before - g.Players[game.Player1].Life; got != test.want {
				t.Fatalf("life lost = %d, want %d", got, test.want)
			}
		})
	}
}

func TestReanimateGainLifeSiblingUsesMovedCardManaValue(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:     "Restorative Creature",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(3)}),
	}})
	addReanimateSpell(g, currentCardTarget(t, g, cardID), true)
	before := g.Players[game.Player1].Life

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life - before; got != 3 {
		t.Fatalf("life gained = %d, want 3", got)
	}
}

func TestReanimateReplacesStaleLinkedPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	stale := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Previously Returned Creature",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(7)}),
	}})
	cardID := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:     "Current Creature",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(2)}),
	}})
	sourceID := addReanimateSpell(g, currentCardTarget(t, g, cardID), false)
	rememberLinkedObject(g, game.LinkedObjectKey{
		SourceID: sourceID,
		LinkID:   "reanimated-card",
	}, permanentLinkedObjectRef(stale))
	before := g.Players[game.Player1].Life

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := before - g.Players[game.Player1].Life; got != 2 {
		t.Fatalf("life lost = %d, want current returned card's mana value 2", got)
	}
}

// TestReanimatePermanentCardWithinManaValueBoundReturnsToBattlefield covers the
// generic reanimation shape from Sevinne's Reclamation: a targeted permanent
// card in the controller's graveyard (here a low-cost artifact, exercising the
// permanent type union beyond creatures) is returned to the battlefield under
// the caster's control.
func TestReanimatePermanentCardWithinManaValueBoundReturnsToBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Salvaged Relic",
		Types:    []types.Card{types.Artifact},
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
	}})

	instructions := []game.Instruction{{
		Primitive: game.PutOnBattlefield{
			Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
			Recipient: opt.Val(game.ControllerReference()),
		},
	}}
	sourceID := addInstructionSpellToStackForController(g, game.Player1, instructions, []game.Target{currentCardTarget(t, g, cardID)})
	card, _ := g.GetCardInstance(sourceID)
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection: opt.Val(game.Selection{
			RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle},
			Controller:       game.ControllerYou,
			ManaValue:        opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 1}),
		}),
	}}

	engine.resolveTopOfStack(g, &TurnLog{})

	permanent, ok := reanimatedPermanent(g, cardID)
	if !ok {
		t.Fatal("targeted permanent card was not put onto the battlefield")
	}
	if permanent.Controller != game.Player1 {
		t.Fatalf("controller = %v, want Player1", permanent.Controller)
	}
	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("reanimated card still in graveyard")
	}
}
