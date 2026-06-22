package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
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
