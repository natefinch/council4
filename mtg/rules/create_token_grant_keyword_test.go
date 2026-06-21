package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// createTokenThenHasteAbility models the lowered shape of "Create a 1/1 ... token.
// That token gains haste until end of turn.": a token creation that publishes the
// created token under a link key, followed by an until-end-of-turn keyword grant
// resolving its object reference to that linked token.
func createTokenThenHasteAbility() game.TriggeredAbility {
	const linkKey = game.LinkedKey("created-token")
	tokenDef := &game.CardDef{CardFace: game.CardFace{
		Name:      "Thopter",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Thopter},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
	return game.TriggeredAbility{
		Text: "create token then haste",
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhen,
			Pattern: game.TriggerPattern{
				Event:  game.EventPermanentEnteredBattlefield,
				Source: game.TriggerSourceSelf,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.CreateToken{
				Amount:        game.Fixed(1),
				Source:        game.TokenDef(tokenDef),
				PublishLinked: linkKey,
			}},
			{Primitive: game.ApplyContinuous{
				Object: opt.Val(game.LinkedObjectReference(string(linkKey))),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:       game.LayerAbility,
					AddKeywords: []game.Keyword{game.Haste},
				}},
				Duration: game.DurationUntilEndOfTurn,
			}},
		}}.Ability(),
	}
}

func addCreateTokenThenHasteSource(g *game.Game, controller game.PlayerID) *game.Permanent {
	def := &game.CardDef{CardFace: game.CardFace{
		Name:               "Token Maker",
		Types:              []types.Card{types.Artifact},
		TriggeredAbilities: []game.TriggeredAbility{createTokenThenHasteAbility()},
	}}
	return addCombatPermanent(g, controller, def)
}

func createdHasteThopter(g *game.Game) *game.Permanent {
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef != nil && permanent.TokenDef.Name == "Thopter" {
			return permanent
		}
	}
	return nil
}

// The created token gains haste until end of turn from the "that token gains
// haste" grant, and loses it at the cleanup step.
func TestCreateTokenThatTokenGainsHasteUntilEndOfTurn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCreateTokenThenHasteSource(g, game.Player1)

	emitEvent(g, game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: source.ObjectID})
	agents := [game.NumPlayers]PlayerAgent{}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("create-token trigger was not put on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	token := createdHasteThopter(g)
	if token == nil {
		t.Fatal("token was not created")
	}
	if !hasKeyword(g, token, game.Haste) {
		t.Fatal("created token does not have haste this turn")
	}

	expireCleanupDurations(g)
	if hasKeyword(g, token, game.Haste) {
		t.Fatal("created token still has haste after cleanup (next turn)")
	}
}
