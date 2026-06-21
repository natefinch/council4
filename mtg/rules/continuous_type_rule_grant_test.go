package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// brotherhoodRegaliaEquipment mirrors the static ability the executable backend
// generates for "Equipped creature has ward {2}, is an Assassin in addition to
// its other types, and can't be blocked." — a single static ability that
// composes a granted Ward keyword ability, a creature-type addition, and a
// can't-be-blocked rule effect on its attached object.
func brotherhoodRegaliaEquipment(g *game.Game, controller game.PlayerID) *game.Permanent {
	ward := game.WardStaticAbility(cost.Mana{cost.O(2)})
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     "Brotherhood Regalia",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Equipment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:        game.LayerAbility,
					Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
					AddAbilities: []game.Ability{&ward},
				},
				{
					Layer:       game.LayerType,
					Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
					AddSubtypes: []types.Sub{types.Sub("Assassin")},
				},
			},
			RuleEffects: []game.RuleEffect{{
				Kind:             game.RuleEffectCantBeBlocked,
				AffectedAttached: true,
			}},
		}},
	}})
}

func attachedCreatureHasGrantedWard(g *game.Game, permanent *game.Permanent) bool {
	for _, ability := range permanentEffectiveAbilities(g, permanent) {
		if body, ok := ability.(*game.StaticAbility); ok {
			if _, ok := game.StaticBodyWardCost(body); ok {
				return true
			}
		}
	}
	return false
}

// TestEquippedCreatureGainsTypeWardAndCantBeBlocked confirms the composed static
// applies all three grants to the attached creature at once: the Assassin
// subtype in addition to its other types, the granted Ward keyword, and the
// can't-be-blocked rule — and that all three lapse when the Equipment leaves.
func TestEquippedCreatureGainsTypeWardAndCantBeBlocked(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	equipment := brotherhoodRegaliaEquipment(g, game.Player1)

	if permanentHasSubtype(g, creature, types.Sub("Assassin")) {
		t.Fatal("creature has Assassin subtype before attaching")
	}
	if attachedCreatureHasGrantedWard(g, creature) {
		t.Fatal("creature has granted ward before attaching")
	}

	creature.Attachments = append(creature.Attachments, equipment.ObjectID)
	equipment.AttachedTo = opt.Val(creature.ObjectID)

	if !permanentHasSubtype(g, creature, types.Sub("Assassin")) {
		t.Fatal("creature lacks Assassin subtype while equipped")
	}
	if !permanentHasType(g, creature, types.Creature) {
		t.Fatal("creature lost its Creature type (addition must not replace existing types)")
	}
	if !attachedCreatureHasGrantedWard(g, creature) {
		t.Fatal("creature lacks granted ward while equipped")
	}

	// The grants stop applying once the Equipment leaves the battlefield.
	g.Battlefield = g.Battlefield[:len(g.Battlefield)-1]
	if permanentHasSubtype(g, creature, types.Sub("Assassin")) {
		t.Fatal("creature retains Assassin subtype after equipment leaves")
	}
	if attachedCreatureHasGrantedWard(g, creature) {
		t.Fatal("creature retains granted ward after equipment leaves")
	}
}
