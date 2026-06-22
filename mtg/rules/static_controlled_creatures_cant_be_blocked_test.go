package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestControlledCreaturesCantBeBlockedStaticGrantsMassEvasion models the runtime
// behavior of "Creatures you control can't be blocked." (the unconditional
// mass-evasion static): every creature the source's controller controls must be
// unblockable by every legal blocker, while creatures controlled by an opponent
// remain blockable. The effect comes from a battlefield static ability whose rule
// effect is scoped to the controller's creatures, so no per-creature effect is
// placed on the game.
func TestControlledCreaturesCantBeBlockedStaticGrantsMassEvasion(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Mass Evasion Enchantment",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantBeBlocked,
				AffectedController: game.ControllerYou,
				PermanentTypes:     []types.Card{types.Creature},
			}},
		}},
	}})
	yourCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	otherYourCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	opponentBlocker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	for _, attacker := range []*game.Permanent{yourCreature, otherYourCreature} {
		if canBlockAttacker(g, blocker, attacker) {
			t.Fatal("controlled-creatures can't-be-blocked static let a creature you control be blocked")
		}
	}
	if !canBlockAttacker(g, opponentBlocker, opponentCreature) {
		t.Fatal("controlled-creatures can't-be-blocked static prevented blocking an opponent's creature")
	}
}

// TestControlledCreaturesWithCounterCantBeBlockedStaticFiltersAffected models
// "Creatures you control with +1/+1 counters on them can't be blocked." (Herald
// of Secret Streams): only the controller's creatures carrying a +1/+1 counter
// are unblockable, while a controlled creature without a counter stays blockable.
// The affected-permanent filter rides on the rule effect's AffectedSelection.
func TestControlledCreaturesWithCounterCantBeBlockedStaticFiltersAffected(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Evasion Enchantment",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantBeBlocked,
				AffectedController: game.ControllerYou,
				PermanentTypes:     []types.Card{types.Creature},
				AffectedSelection: game.Selection{
					MatchCounter:    true,
					RequiredCounter: counter.PlusOnePlusOne,
				},
			}},
		}},
	}})
	counteredCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	counteredCreature.Counters.Add(counter.PlusOnePlusOne, 1)
	plainCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	if canBlockAttacker(g, blocker, counteredCreature) {
		t.Fatal("counter-filtered can't-be-blocked static let a countered creature be blocked")
	}
	if !canBlockAttacker(g, blocker, plainCreature) {
		t.Fatal("counter-filtered can't-be-blocked static blocked a creature with no counter that should stay blockable")
	}
}

// TestControlledCreaturesColorFilteredCantBeBlockedStaticFiltersAffected models
// "Blue creatures you control can't be blocked." (Deepchannel Mentor): only the
// controller's blue creatures are unblockable, while a controlled non-blue
// creature stays blockable.
func TestControlledCreaturesColorFilteredCantBeBlockedStaticFiltersAffected(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Blue Evasion Enchantment",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantBeBlocked,
				AffectedController: game.ControllerYou,
				PermanentTypes:     []types.Card{types.Creature},
				AffectedSelection: game.Selection{
					ColorsAny: []color.Color{color.Blue},
				},
			}},
		}},
	}})
	blueCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:   "Blue Combat Creature",
		Types:  []types.Card{types.Creature},
		Colors: []color.Color{color.Blue},
	}})
	redCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:   "Red Combat Creature",
		Types:  []types.Card{types.Creature},
		Colors: []color.Color{color.Red},
	}})
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	if canBlockAttacker(g, blocker, blueCreature) {
		t.Fatal("color-filtered can't-be-blocked static let a blue creature be blocked")
	}
	if !canBlockAttacker(g, blocker, redCreature) {
		t.Fatal("color-filtered can't-be-blocked static blocked a non-blue creature that should stay blockable")
	}
}

// addCombatCreatureWithPowerToughness adds a battlefield creature with the given
// power and toughness so the power/toughness-comparison evasion filters can be
// exercised independently on each characteristic.
func addCombatCreatureWithPowerToughness(g *game.Game, controller game.PlayerID, power, toughness int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "PT Combat Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
	}})
}

// TestControlledCreaturesPowerOrToughnessFilteredCantBeBlockedStaticFiltersAffected
// models "Creatures you control with power or toughness 1 or less can't be
// blocked." (Tetsuko Umezawa, Fugitive): a controlled creature whose power OR
// whose toughness is 1 or less is unblockable, while a controlled creature with
// both power and toughness greater than 1 stays blockable. The disjunction rides
// on the rule effect's AffectedSelection.AnyOf.
func TestControlledCreaturesPowerOrToughnessFilteredCantBeBlockedStaticFiltersAffected(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "PT Evasion Enchantment",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantBeBlocked,
				AffectedController: game.ControllerYou,
				PermanentTypes:     []types.Card{types.Creature},
				AffectedSelection: game.Selection{
					AnyOf: []game.Selection{
						{Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 1})},
						{Toughness: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 1})},
					},
				},
			}},
		}},
	}})
	lowPower := addCombatCreatureWithPowerToughness(g, game.Player1, 1, 5)
	lowToughness := addCombatCreatureWithPowerToughness(g, game.Player1, 5, 1)
	bigCreature := addCombatCreatureWithPowerToughness(g, game.Player1, 3, 3)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	for _, attacker := range []*game.Permanent{lowPower, lowToughness} {
		if canBlockAttacker(g, blocker, attacker) {
			t.Fatal("power-or-toughness-filtered can't-be-blocked static let a creature with power or toughness 1 or less be blocked")
		}
	}
	if !canBlockAttacker(g, blocker, bigCreature) {
		t.Fatal("power-or-toughness-filtered can't-be-blocked static blocked a creature with power and toughness greater than 1 that should stay blockable")
	}
}

// TestControlledCreaturesSourcePowerFilteredCantBeBlockedStaticFiltersAffected
// models "Creatures you control with power greater than ~'s power can't be
// blocked." (Champion of Lambholt): a controlled creature whose power exceeds the
// static's source creature's power is unblockable, while a controlled creature
// whose power does not exceed it stays blockable. The source-relative comparison
// rides on the rule effect's AffectedSelection.PowerGreaterThanSource, read
// against the source permanent's own power.
func TestControlledCreaturesSourcePowerFilteredCantBeBlockedStaticFiltersAffected(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Source Power Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantBeBlocked,
				AffectedController: game.ControllerYou,
				PermanentTypes:     []types.Card{types.Creature},
				AffectedSelection: game.Selection{
					PowerGreaterThanSource: true,
				},
			}},
		}},
	}})
	stronger := addCombatCreatureWithPowerToughness(g, game.Player1, 3, 3)
	equalPower := addCombatCreatureWithPowerToughness(g, game.Player1, 2, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	if canBlockAttacker(g, blocker, stronger) {
		t.Fatal("source-power-filtered can't-be-blocked static let a creature with power greater than the source be blocked")
	}
	if !canBlockAttacker(g, blocker, equalPower) {
		t.Fatal("source-power-filtered can't-be-blocked static blocked a creature with power not greater than the source that should stay blockable")
	}
}
