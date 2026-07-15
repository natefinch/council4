package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ingaDrawAnchor is a permanent carrying Inga and Esika's headline draw ability:
// "Whenever you cast a creature spell, if three or more mana from creatures was
// spent to cast it, draw a card." It filters cast events to creature spells and
// gates the draw on the creature-mana intervening condition.
func ingaDrawAnchor() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Inga Draw Anchor",
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:            game.EventSpellCast,
					Controller:       game.TriggerControllerYou,
					RequireCardTypes: []types.Card{types.Creature},
				},
				InterveningCondition: opt.Val(game.Condition{
					Aggregates: []game.AggregateComparison{{
						Aggregate: game.AggregateEventSpellManaFromCreaturesSpentToCast,
						Op:        compare.GreaterOrEqual,
						Value:     3,
					}},
				}),
			},
			Content: game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
				}},
			}.Ability(),
		}},
	}}
}

// castThreeGCreatureFromDorks sets up n creature dorks and one Forest land, casts
// a {G}{G}{G} creature paying from the available sources (which, when n >= 3,
// spends only creature mana), and resolves any triggers. It returns whether the
// player drew the expected library card.
func castCreatureAndCheckDraw(t *testing.T, creatureDorks, rocks int) bool {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, ingaDrawAnchor())
	for range creatureDorks {
		creatureManaDork(g, game.Player1, "Dork", mana.G)
	}
	for range rocks {
		noncreatureManaRock(g, game.Player1, "Rock", mana.G)
	}
	drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Reward"}})

	def := creatureSpellDef("Cast Beast", types.Beast)
	def.ManaCost = opt.Val(cost.Mana{cost.G, cost.G, cost.G})
	def.Power = opt.Val(game.PT{Value: 3})
	def.Toughness = opt.Val(game.PT{Value: 3})
	spellID := addCardToHand(g, game.Player1, def)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast creature spell) = false, want true")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		for {
			obj, ok := g.Stack.Peek()
			if !ok || obj.Kind != game.StackTriggeredAbility {
				break
			}
			engine.resolveTopOfStack(g, &TurnLog{})
		}
	}
	return g.Players[game.Player1].Hand.Contains(drawn)
}

// TestIngaDrawFiresWithThreeCreatureMana proves the full chain: casting a
// creature spell paid entirely with three creature mana meets the intervening
// condition and draws a card.
func TestIngaDrawFiresWithThreeCreatureMana(t *testing.T) {
	t.Parallel()
	if !castCreatureAndCheckDraw(t, 3, 0) {
		t.Fatal("expected a draw when three creature mana paid for the creature spell")
	}
}

// TestIngaDrawSkippedWithTwoCreatureMana proves the threshold is strict: paying
// with only two creature mana (and one non-creature rock) leaves the condition
// unmet, so no card is drawn.
func TestIngaDrawSkippedWithTwoCreatureMana(t *testing.T) {
	t.Parallel()
	if castCreatureAndCheckDraw(t, 2, 1) {
		t.Fatal("did not expect a draw when only two creature mana paid for the creature spell")
	}
}

// TestIngaDrawSkippedWithNoCreatureMana proves paying entirely from non-creature
// sources never meets the condition.
func TestIngaDrawSkippedWithNoCreatureMana(t *testing.T) {
	t.Parallel()
	if castCreatureAndCheckDraw(t, 0, 3) {
		t.Fatal("did not expect a draw when no creature mana paid for the creature spell")
	}
}

// TestIngaDrawSkippedForNoncreatureSpell proves the creature-spell filter: even
// when three creature mana pays for the spell, casting a noncreature spell does
// not fire the trigger, so no card is drawn.
func TestIngaDrawSkippedForNoncreatureSpell(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, ingaDrawAnchor())
	creatureManaDork(g, game.Player1, "Dork", mana.G)
	creatureManaDork(g, game.Player1, "Dork", mana.G)
	creatureManaDork(g, game.Player1, "Dork", mana.G)
	drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Reward"}})

	sorcery := &game.CardDef{CardFace: game.CardFace{
		Name:     "Three-Mana Sorcery",
		ManaCost: opt.Val(cost.Mana{cost.G, cost.G, cost.G}),
		Types:    []types.Card{types.Sorcery},
	}}
	spellID := addCardToHand(g, game.Player1, sorcery)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast noncreature spell) = false, want true")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("creature-spell trigger fired for a noncreature spell")
	}
	if g.Players[game.Player1].Hand.Contains(drawn) {
		t.Fatal("did not expect a draw when casting a noncreature spell")
	}
}
