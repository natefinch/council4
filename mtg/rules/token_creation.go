package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

type tokenCreationPlan struct {
	controller game.PlayerID
	def        *game.CardDef
	count      int
	face       game.FaceIndex
}

func planTokenCreation(g *game.Game, controller game.PlayerID, token *game.CardDef, amount int, faces ...game.FaceIndex) []tokenCreationPlan {
	if token == nil || amount <= 0 {
		return nil
	}
	face := game.FaceFront
	if len(faces) > 0 {
		face = faces[0]
	}
	if _, available := token.FaceDefView(face); !available {
		return nil
	}
	plans := []tokenCreationPlan{{controller: controller, def: token, count: amount, face: face}}
	// Replacement-created products inherit this event's applied-effect lineage
	// (CR 614.5). Distinct sources retain distinct replacement identities.
	applied := make(map[id.ID]bool)
	for {
		var replacements []game.ReplacementEffect
		for index := range g.ReplacementEffects {
			replacement := &g.ReplacementEffects[index]
			if applied[replacement.ID] {
				continue
			}
			if slices.ContainsFunc(plans, func(plan tokenCreationPlan) bool {
				return tokenCreationPlanMatchesReplacement(g, plan, replacement)
			}) {
				replacements = append(replacements, *replacement)
			}
		}
		if len(replacements) == 0 {
			return plans
		}
		replacement := replacements[0]
		if len(replacements) > 1 {
			decision := chooseReplacementDecision(g, controller, replacementEffectLabels(replacements))
			replacement = selectedReplacementEffect(replacements, decision)
		}
		applied[replacement.ID] = true
		var modified []tokenCreationPlan
		extra := false
		for _, plan := range plans {
			if !tokenCreationPlanMatchesReplacement(g, plan, &replacement) {
				modified = appendTokenCreationPlan(modified, plan)
				continue
			}
			switch {
			case replacement.TokenReplaceDef != nil:
				modified = appendTokenCreationPlan(modified, tokenCreationPlan{
					controller: plan.controller, def: replacement.TokenReplaceDef, count: plan.count,
				})
			case len(replacement.CreateOneOfEachTokens) > 0:
				for _, def := range replacement.CreateOneOfEachTokens {
					modified = appendTokenCreationPlan(modified, tokenCreationPlan{
						controller: plan.controller, def: def, count: plan.count,
					})
				}
			default:
				if replacement.TokenMultiplier > 1 {
					plan.count *= replacement.TokenMultiplier
				}
				if replacement.TokenAddendDef != nil {
					extra = replacement.TokenAddend > 0
				} else {
					plan.count = max(0, plan.count+replacement.TokenAddend)
				}
				modified = appendTokenCreationPlan(modified, plan)
			}
		}
		if extra {
			modified = appendTokenCreationPlan(modified, tokenCreationPlan{
				controller: controller, def: replacement.TokenAddendDef, count: replacement.TokenAddend,
			})
		}
		plans = modified
	}
}

func appendTokenCreationPlan(plans []tokenCreationPlan, member tokenCreationPlan) []tokenCreationPlan {
	if member.count <= 0 {
		return plans
	}
	for index, plan := range plans {
		if plan.controller == member.controller && plan.def == member.def && plan.face == member.face {
			plans[index].count += member.count
			return plans
		}
	}
	return append(plans, member)
}

func tokenCreationPlanMatchesReplacement(g *game.Game, plan tokenCreationPlan, replacement *game.ReplacementEffect) bool {
	visible, available := plan.def.FaceDefView(plan.face)
	if !available || plan.count <= 0 {
		return false
	}
	event := game.Event{Kind: game.EventTokenCreated, Controller: plan.controller, Player: plan.controller, Amount: plan.count}
	return tokenCreationReplacementMatches(g, replacement, event, visible)
}
