package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// payEnlistCosts offers each enlist cost of the declared attackers (CR
// 702.154). Enlist's static ability is an optional cost to attack chosen and
// paid as attackers are declared, after the attackers are tapped (CR 508.1f,
// 508.1g, 508.1j). Each instance of enlist on each attacker is offered
// separately (CR 702.154d): the attacking player may tap one untapped creature
// they control that isn't attacking and that either has haste or has been under
// their control continuously since the turn began. Paying a cost queues that
// instance's linked "when you do" trigger (CR 603.12, 607.2h), which is put on
// the stack with the attack triggers. A creature tapped for one enlist cost is
// no longer untapped, so it can't pay another.
func (ce combatEngine) payEnlistCosts(g *game.Game, playerID game.PlayerID, attackers []game.AttackDeclaration, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	attacking := make(map[id.ID]bool, len(attackers))
	for _, declaration := range attackers {
		attacking[declaration.Attacker] = true
	}
	for _, declaration := range attackers {
		attacker, ok := permanentByObjectID(g, declaration.Attacker)
		if !ok {
			continue
		}
		for range enlistInstances(g, attacker) {
			candidates := enlistCandidates(g, playerID, attacking)
			tapped := ce.e.chooseAndTapPermanents(g, agents, log, chosenTapRequest{
				chooser:    playerID,
				candidates: candidates,
				maxChoices: 1,
				prompt:     "Enlist: choose a creature to tap",
				subject:    attacker,
			})
			if len(tapped) == 1 {
				queueEnlistTrigger(g, playerID, attacker, tapped[0])
			}
		}
	}
}

// enlistInstances counts the enlist instances an attacker has. A granted
// enlist that is not carried by a static ability body still counts once.
func enlistInstances(g *game.Game, attacker *game.Permanent) int {
	if !hasKeyword(g, attacker, game.Enlist) {
		return 0
	}
	count := 0
	for _, body := range permanentEffectiveAbilitiesView(g, attacker) {
		if _, ok := body.(*game.StaticAbility); !ok {
			continue
		}
		for _, ability := range game.BodyKeywordAbilities(body) {
			if game.KeywordAbilityKind(ability) == game.Enlist {
				count++
			}
		}
	}
	return max(count, 1)
}

// enlistCandidates returns the creatures playerID could tap to pay an enlist
// cost: untapped, not attacking, and either with haste or continuously
// controlled since the turn began (CR 702.154a).
func enlistCandidates(g *game.Game, playerID game.PlayerID, attacking map[id.ID]bool) []*game.Permanent {
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	var candidates []*game.Permanent
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) || permanent.Tapped || attacking[permanent.ObjectID] ||
			effectiveController(g, permanent) != playerID || !permanentHasType(g, permanent, types.Creature) {
			continue
		}
		if permanentSummoningSick(g, permanent) && !hasKeyword(g, permanent, game.Haste) && !suspendHasteApplies(g, permanent) {
			continue
		}
		candidates = append(candidates, permanent)
	}
	return candidates
}

// queueEnlistTrigger queues enlist's linked reflexive trigger: "When you do,
// this creature gets +X/+0 until end of turn, where X is the tapped creature's
// power." X is read as the trigger resolves, from the enlisted creature or its
// last known information.
func queueEnlistTrigger(g *game.Game, playerID game.PlayerID, attacker, enlisted *game.Permanent) {
	g.PendingReflexiveTriggers = append(g.PendingReflexiveTriggers, game.ReflexiveTrigger{
		SourceID:         attacker.CardInstanceID,
		SourceObjectID:   attacker.ObjectID,
		SourceTokenDef:   attacker.TokenDef,
		Controller:       playerID,
		Ability:          enlistTriggeredAbility(),
		CapturedObjectID: enlisted.ObjectID,
	})
}

func enlistTriggeredAbility() game.TriggeredAbility {
	return game.TriggeredAbility{
		Text: "Enlist",
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.ModifyPT{
				Object: game.SourcePermanentReference(),
				PowerDelta: game.Dynamic(game.DynamicAmount{
					Kind:   game.DynamicAmountObjectPower,
					Object: game.CapturedObjectReference(),
				}),
				ToughnessDelta: game.Fixed(0),
				Duration:       game.DurationUntilEndOfTurn,
			}}},
		}.Ability(),
		KeywordAbilities: game.SimpleKeywords(game.Enlist),
	}
}
