package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// cemeteryProwlerDef builds a reusable permanent carrying Cemetery Prowler's two
// linked building blocks: an enter-or-attack trigger that exiles a card from any
// graveyard under the exiled-with-source link, and a controller cast-cost static
// whose discount scales with the card types a spell shares with the cards exiled
// with the source. No card-name-specific behavior backs it.
func cemeteryProwlerDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Cemetery Prowler",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCostModifier,
				AffectedPlayer: game.PlayerYou,
				CostModifier: game.CostModifier{
					Kind:                          game.CostModifierSpell,
					SharedExiledCardTypeReduction: 1,
					ExiledLinkKey:                 game.LinkedKey("exiled-with-source"),
				},
			}},
		}},
	}}
}

// exileCardWithProwler resolves the enter-or-attack graveyard exile so the chosen
// card is recorded under the source's exiled-with-source link, matching the
// lowered trigger instruction.
func exileCardWithProwler(g *game.Game, obj *game.StackObject) {
	engine := NewEngine(nil)
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.ExileFromGraveyardChoice(
		game.ControllerReference(),
		game.Selection{},
		game.Fixed(1),
		true,
		game.LinkedKey("exiled-with-source"),
	)}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
}

func spellGenericReductionForCaster(g *game.Game, caster game.PlayerID, card *game.CardDef) int {
	total := 0
	for _, modifier := range staticCostModifiersForContext(g, caster, card, zone.Hand, nil) {
		total += modifier.GenericReduction
	}
	return total
}

// TestCemeteryProwlerCostReductionCountsSharedExiledTypes verifies the cost
// reduction equals the number of distinct card types the casting spell shares
// with the cards exiled with the source: a single exiled artifact creature lets
// an artifact creature spell cost {2} less, a plain creature spell {1} less, and
// an instant spell nothing.
func TestCemeteryProwlerCostReductionCountsSharedExiledTypes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	prowler := addCombatPermanent(g, game.Player1, cemeteryProwlerDef())
	obj := triggeredObjFor(prowler)

	addCfzGraveyardCard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Exiled Artifact Creature",
		Types: []types.Card{types.Artifact, types.Creature},
	}})
	exileCardWithProwler(g, obj)

	artifactCreature := &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Construct",
		Types:    []types.Card{types.Artifact, types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(4)}),
	}}
	if got := spellGenericReductionForCaster(g, game.Player1, artifactCreature); got != 2 {
		t.Fatalf("artifact creature reduction = %d, want 2", got)
	}

	creature := &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Bear",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(2)}),
	}}
	if got := spellGenericReductionForCaster(g, game.Player1, creature); got != 1 {
		t.Fatalf("creature reduction = %d, want 1", got)
	}

	instant := &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Bolt",
		Types:    []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
	}}
	if got := spellGenericReductionForCaster(g, game.Player1, instant); got != 0 {
		t.Fatalf("instant reduction = %d, want 0", got)
	}

	// The discount is the controller's only; an opponent's matching spell is
	// never reduced.
	if got := spellGenericReductionForCaster(g, game.Player2, artifactCreature); got != 0 {
		t.Fatalf("opponent reduction = %d, want 0", got)
	}
}

func TestSharedExiledTypeReductionOnce(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := cemeteryProwlerDef()
	modifier := &def.StaticAbilities[0].RuleEffects[0].CostModifier
	modifier.SharedExiledCardTypeReduction = 2
	modifier.SharedExiledCardTypeReductionOnce = true
	modifier.ExiledLinkObjectScoped = true
	source := addCombatPermanent(g, game.Player1, def)
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Types: []types.Card{types.Artifact, types.Creature},
	}})
	engine := NewEngine(nil)
	engine.resolveInstructionWithChoices(g, triggeredObjFor(source), &game.Instruction{
		Primitive: game.ExileFromHandChoice(
			game.ControllerReference(),
			game.Selection{ExcludedTypes: []types.Card{types.Land}},
			game.Fixed(1),
			game.LinkedKey("exiled-with-source"),
		),
	}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	artifactCreature := &game.CardDef{CardFace: game.CardFace{
		Types: []types.Card{types.Artifact, types.Creature},
	}}
	if got := spellGenericReductionForCaster(g, game.Player1, artifactCreature); got != 2 {
		t.Fatalf("flat shared-type reduction = %d, want 2", got)
	}
}

// TestCemeteryProwlerCostReductionEmptyWithoutExiledCards verifies the discount
// is zero before the trigger has exiled anything.
func TestCemeteryProwlerCostReductionEmptyWithoutExiledCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, cemeteryProwlerDef())

	creature := &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Bear",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(2)}),
	}}
	if got := spellGenericReductionForCaster(g, game.Player1, creature); got != 0 {
		t.Fatalf("reduction with no exiled cards = %d, want 0", got)
	}
}
