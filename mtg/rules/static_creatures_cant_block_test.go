package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestCreaturesCantBlockSourceShieldStaticRestrictsBlockers models the runtime
// behavior of "Creatures with power less than this creature's power can't block
// it." (Wandering Wolf, Aura Gnarlid, Den Protector): a creature whose power is
// less than the source's power can't block the source, while a creature with
// power not less than the source's may. Because the restriction shields only the
// source (BlockedSource), an affected weak blocker can still block any other
// attacker.
func TestCreaturesCantBlockSourceShieldStaticRestrictsBlockers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Source Shield Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantBlock,
				AffectedController: game.ControllerAny,
				PermanentTypes:     []types.Card{types.Creature},
				AffectedSelection:  game.Selection{PowerLessThanSource: true},
				BlockedSource:      true,
			}},
		}},
	}})
	weakBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	strongBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	otherAttacker := addCombatCreaturePermanentWithPower(g, game.Player1, 5)

	if canBlockAttacker(g, weakBlocker, source) {
		t.Fatal("source-shield can't-block static let a weaker creature block the source")
	}
	if !canBlockAttacker(g, strongBlocker, source) {
		t.Fatal("source-shield can't-block static stopped a creature with power not less than the source from blocking it")
	}
	if !canBlockAttacker(g, weakBlocker, otherAttacker) {
		t.Fatal("source-shield can't-block static stopped a weak creature from blocking an attacker other than the source")
	}
}

// TestCreaturesCantBlockControlledCreaturesStaticRestrictsBlockers models
// "Creatures with power less than this creature's power can't block creatures you
// control." (Champion of Lambholt): a creature whose power is less than the
// source's power can't block any creature the source's controller controls, but
// it may still block creatures controlled by others, and a creature with power
// not less than the source's may block freely.
func TestCreaturesCantBlockControlledCreaturesStaticRestrictsBlockers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Lambholt Champion",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantBlock,
				AffectedController: game.ControllerAny,
				PermanentTypes:     []types.Card{types.Creature},
				AffectedSelection:  game.Selection{PowerLessThanSource: true},
				BlockedSelection: game.Selection{
					Controller:    game.ControllerYou,
					RequiredTypes: []types.Card{types.Creature},
				},
			}},
		}},
	}})
	yourCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	weakBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	strongBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 4)

	if canBlockAttacker(g, weakBlocker, yourCreature) {
		t.Fatal("controlled-creatures can't-block static let a weaker creature block a creature you control")
	}
	if !canBlockAttacker(g, weakBlocker, opponentCreature) {
		t.Fatal("controlled-creatures can't-block static stopped a weak creature from blocking a creature you don't control")
	}
	if !canBlockAttacker(g, strongBlocker, yourCreature) {
		t.Fatal("controlled-creatures can't-block static stopped a creature with power not less than the source from blocking your creature")
	}
}
