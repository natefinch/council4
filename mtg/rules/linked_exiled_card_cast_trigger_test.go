package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned corpus 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Thought Prison (da1ea03f-747f-46c9-8ff5-7e85996df804), {5} Artifact:
// "Imprint — When this artifact enters, you may have target player reveal their
// hand. If you do, choose a nonland card from it and exile that card.
// Whenever a player casts a spell that shares a color or mana value with the
// exiled card, this artifact deals 2 damage to that player." Ruling 2004-12-01:
// the second ability triggers only once per spell; a spell sharing both a color
// and a mana value deals 2 damage, not 4. CR 2026-11-06
// (MagicCompRules 20261106.txt, SHA256
// 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829): 603.2 and
// 603.2c (the condition is part of the trigger event, which triggers once),
// 709.4 and 709.4b (an exiled split card has both halves' combined colors and
// mana value),
// 601.2i (cast triggers see the spell as it becomes cast), 607.2a (the linked
// ability refers only to cards its first ability put in exile that are still
// there), 400.7 (a card that leaves exile is a new object), 105.2c (a colorless
// object has no color), 202.3 (mana value). This test-local Thought Prison
// proves the shared trigger predicate (its {B} cost and opponent target reuse
// the linked hand-exile test helpers); the card stays unregistered.
const thoughtPrisonTestProducer game.LinkedProducerKey = "test/thought-prison/enter/exiles"

var thoughtPrisonEnterTestReference = game.ProgramReference{ID: "test/thought-prison/enter", Version: "1"}

func thoughtPrisonTestDef() *game.CardDef {
	enter := game.Mode{Targets: []game.TargetSpec{targetOpponentSpec()}, Program: thoughtPrisonEnterTestReference}.Ability()
	enter.LinkedProducers = []game.LinkedProducerKey{thoughtPrisonTestProducer}
	damage := game.Mode{Sequence: []game.Instruction{{Primitive: game.Damage{
		Amount:       game.Fixed(2),
		Recipient:    game.PlayerDamageRecipient(game.EventPlayerReference()),
		DamageSource: opt.Val(game.SourcePermanentReference()),
	}}}}.Ability()
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Thought Prison", ManaCost: opt.Val(cost.Mana{cost.B}), Types: []types.Card{types.Artifact},
		TriggeredAbilities: []game.TriggeredAbility{
			{Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
				Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
			}}, Content: enter},
			{Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event: game.EventSpellCast,
				SpellSharesWithLinkedExiledCard: game.LinkedExiledCardComparison{
					Producer: thoughtPrisonTestProducer, SharesColor: true, SameManaValue: true,
				},
			}}, Content: damage},
		},
	}}
}

func coloredInstant(name string, c color.Color, symbols ...cost.Symbol) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: name, ManaCost: opt.Val(cost.Mana(symbols)), Colors: []color.Color{c}, Types: []types.Card{types.Instant},
	}}
}

// castThoughtPrisonProbe casts def for player with its mana and reports whether
// exactly one triggered ability went on the stack above it. It resolves that ability and
// the spell, leaving the stack empty.
func castThoughtPrisonProbe(t *testing.T, e *Engine, g *game.Game, player game.PlayerID, def *game.CardDef, pay map[mana.Color]int) bool {
	t.Helper()
	cardID := addCardToHand(g, player, def)
	for c, amount := range pay {
		g.Players[player].ManaPool.Add(c, amount)
	}
	g.Turn.PriorityPlayer = player
	if !e.applyAction(g, player, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatalf("casting %s failed", def.Name)
	}
	triggered := e.putTriggeredAbilitiesOnStack(g)
	if triggered && g.Stack.Size() != 2 {
		t.Fatalf("%s put %d triggered abilities on the stack, want one", def.Name, g.Stack.Size()-1)
	}
	for !g.Stack.IsEmpty() {
		e.resolveTopOfStack(g, &TurnLog{})
	}
	return triggered
}

// TestThoughtPrisonLinkedExiledCardCastTrigger plays Thought Prison's text: its
// enters ability exiles a red mana value 3 card from the opponent's hand, then
// spells sharing a color or the mana value trigger 2 damage to their caster
// once, an unrelated spell triggers nothing, and nothing triggers after the
// exiled card left exile and returned as a new object. A clone keeps the link.
func TestThoughtPrisonLinkedExiledCardCastTrigger(t *testing.T) {
	t.Parallel()
	g, e := linkedHandExileEngine(t, revealHandExileEnterProgram(thoughtPrisonEnterTestReference, nonlandSelection(), true,
		cardprogram.CardPlacement{PublishLinkedProducer: thoughtPrisonTestProducer}))
	addCardToHand(g, game.Player2, revealHandLand())
	exiled := addCardToHand(g, game.Player2, coloredInstant("Red Three", color.Red, cost.O(2), cost.R))
	castLinkedHandExileCreature(t, e, g, thoughtPrisonTestDef())
	e.resolveTopOfStackWithChoices(g, chooseRevealedCard(t, "Red Three"), &TurnLog{})
	requireCardInZone(t, g, exiled, zone.Exile)

	probes := []struct {
		name    string
		caster  game.PlayerID
		def     *game.CardDef
		pay     map[mana.Color]int
		trigger bool
	}{
		{"shared color", game.Player1, coloredInstant("Red One", color.Red, cost.R), map[mana.Color]int{mana.R: 1}, true},
		{"shared mana value", game.Player2, coloredInstant("Blue Three", color.Blue, cost.O(2), cost.U), map[mana.Color]int{mana.U: 3}, true},
		{"color and mana value once", game.Player1, coloredInstant("Red Three Again", color.Red, cost.O(2), cost.R), map[mana.Color]int{mana.R: 3}, true},
		{"unrelated", game.Player1, coloredInstant("Green Two", color.Green, cost.O(1), cost.G), map[mana.Color]int{mana.G: 2}, false},
	}
	for _, probe := range probes {
		life := g.Players[probe.caster].Life
		if got := castThoughtPrisonProbe(t, e, g, probe.caster, probe.def, probe.pay); got != probe.trigger {
			t.Fatalf("%s: triggered = %v, want %v", probe.name, got, probe.trigger)
		}
		want := life
		if probe.trigger {
			want -= 2
		}
		if g.Players[probe.caster].Life != want {
			t.Fatalf("%s: caster life = %d, want %d", probe.name, g.Players[probe.caster].Life, want)
		}
	}

	branch := g.Clone()
	moveExiledOutAndBack(t, g, exiled)
	if castThoughtPrisonProbe(t, e, g, game.Player1, coloredInstant("Red One", color.Red, cost.R), map[mana.Color]int{mana.R: 1}) {
		t.Fatal("a spell triggered against a card that left exile and returned as a new object")
	}
	if !castThoughtPrisonProbe(t, e, branch, game.Player1, coloredInstant("Red One", color.Red, cost.R), map[mana.Color]int{mana.R: 1}) {
		t.Fatal("the cloned game lost the linked exiled card")
	}
}

// TestThoughtPrisonComparesExiledSplitCardCombinedCharacteristics proves the
// exiled card is compared with its characteristics in exile: Assault // Battery
// is red and green with mana value 5 there (CR 709.4, 709.4b), so a green spell
// and a mana value 5 spell trigger while a blue mana value 1 spell, matching
// only Assault's own mana value, does not.
func TestThoughtPrisonComparesExiledSplitCardCombinedCharacteristics(t *testing.T) {
	t.Parallel()
	g, e := linkedHandExileEngine(t, revealHandExileEnterProgram(thoughtPrisonEnterTestReference, nonlandSelection(), true,
		cardprogram.CardPlacement{PublishLinkedProducer: thoughtPrisonTestProducer}))
	split := registryCard(t, "Assault")
	if split.Layout != game.LayoutSplit || !split.Alternate.Exists || split.Alternate.Val.Name != "Battery" {
		t.Fatalf("Assault is not registered as the split card Assault // Battery: %+v", split)
	}
	exiled := addCardToHand(g, game.Player2, split)
	castLinkedHandExileCreature(t, e, g, thoughtPrisonTestDef())
	e.resolveTopOfStackWithChoices(g, chooseRevealedCard(t, "Assault"), &TurnLog{})
	requireCardInZone(t, g, exiled, zone.Exile)

	for _, probe := range []struct {
		name    string
		def     *game.CardDef
		pay     map[mana.Color]int
		trigger bool
	}{
		{"alternate half color", coloredInstant("Green Two", color.Green, cost.O(1), cost.G), map[mana.Color]int{mana.G: 2}, true},
		{"combined mana value", coloredInstant("Blue Five", color.Blue, cost.O(4), cost.U), map[mana.Color]int{mana.U: 5}, true},
		{"front half mana value only", coloredInstant("Blue One", color.Blue, cost.U), map[mana.Color]int{mana.U: 1}, false},
	} {
		if got := castThoughtPrisonProbe(t, e, g, game.Player1, probe.def, probe.pay); got != probe.trigger {
			t.Fatalf("%s: triggered = %v, want %v", probe.name, got, probe.trigger)
		}
	}
}

// TestThoughtPrisonWithoutLinkedExiledCardNeverTriggers proves that declining
// the optional imprint leaves nothing linked, so no spell triggers the ability.
func TestThoughtPrisonWithoutLinkedExiledCardNeverTriggers(t *testing.T) {
	t.Parallel()
	g, e := linkedHandExileEngine(t, revealHandExileEnterProgram(thoughtPrisonEnterTestReference, nonlandSelection(), true,
		cardprogram.CardPlacement{PublishLinkedProducer: thoughtPrisonTestProducer}))
	kept := addCardToHand(g, game.Player2, coloredInstant("Red Three", color.Red, cost.O(2), cost.R))
	castLinkedHandExileCreature(t, e, g, thoughtPrisonTestDef())
	e.resolveTopOfStackWithChoices(g, chooseRevealedCard(t, ""), &TurnLog{})
	requireCardInZone(t, g, kept, zone.Hand)
	if castThoughtPrisonProbe(t, e, g, game.Player1, coloredInstant("Red Three Again", color.Red, cost.O(2), cost.R), map[mana.Color]int{mana.R: 3}) {
		t.Fatal("a spell triggered with no linked exiled card")
	}
}

func moveExiledOutAndBack(t *testing.T, g *game.Game, cardID id.ID) {
	t.Helper()
	if !moveCardBetweenZonesInBatch(g, game.Player2, cardID, zone.Exile, zone.Graveyard, false, 0) ||
		!moveCardBetweenZonesInBatch(g, game.Player2, cardID, zone.Graveyard, zone.Exile, false, 0) {
		t.Fatal("could not move the exiled card out and back")
	}
}
