package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestTokenOnlyTargetPredicateAllowsTokensRejectsNontokens covers a
// "target token you control" restriction (Caretaker's Talent copies target
// token). The TokenOnly predicate must accept a token permanent and reject an
// otherwise-eligible nontoken permanent.
func TestTokenOnlyTargetPredicateAllowsTokensRejectsNontokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Token Copier",
		Types: []types.Card{types.Enchantment},
		ActivatedAbilities: []game.ActivatedAbility{{
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target token you control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						Controller: game.ControllerYou,
						TokenOnly:  true,
					},
				}},
				Sequence: []game.Instruction{{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}}},
			}.Ability(),
		}}},
	})

	tokenDef := &game.CardDef{CardFace: game.CardFace{Name: "Cat Token", Types: []types.Card{types.Creature}}}
	token := addCombatPermanent(g, game.Player1, tokenDef)
	token.Token = true
	token.TokenDef = tokenDef
	nonToken := addCreaturePermanent(g, game.Player1)

	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if !containsAction(legal, action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PermanentTarget(token.ObjectID)}, 0)) {
		t.Fatal("token-only target predicate did not allow token permanent")
	}
	if containsAction(legal, action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PermanentTarget(nonToken.ObjectID)}, 0)) {
		t.Fatal("token-only target predicate allowed nontoken permanent")
	}
}
