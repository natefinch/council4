package rules

import (
	"errors"
	"slices"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

func (c *programContext) Destroy(primitive game.Destroy) (cardprogram.Destruction, error) {
	if err := c.valid(); err != nil {
		return cardprogram.Destruction{}, err
	}
	if err := game.ValidatePrimitive(primitive, c.specs); err != nil {
		return cardprogram.Destruction{}, err
	}
	previous, _ := c.game.ChoiceContext()
	c.engine.setReplacementChoiceContext(c.game, c.agents, c.log)
	defer c.game.SetChoiceContext(previous)
	resolver := newEffectResolver(c.engine, c.game, c.obj, c.agents, c.log)
	targets := resolver.resolveObjectGroup(primitive.Object, primitive.Group)
	outcome := cardprogram.Destruction{Available: !targets.single || targets.resolved}
	seen := make(map[id.ID]bool, len(targets.permanents))
	permanents := make([]*game.Permanent, 0, len(targets.permanents))
	for _, permanent := range targets.permanents {
		if seen[permanent.ObjectID] {
			continue
		}
		seen[permanent.ObjectID] = true
		permanents = append(permanents, permanent)
		outcome.Members = append(outcome.Members, cardprogram.Departure{
			Before: currentProgramSnapshot(c.game, permanent), Attempted: true,
		})
	}
	batch := &destroyBatch{game: c.game, simultaneousID: c.game.IDGen.Next()}
	destroyed, replacements := planDestroyPermanents(c.game, permanents, primitive.PreventRegeneration, batch.simultaneousID)
	results := applyPlannedDestroyBatchResults(c.game, destroyed, replacements, batch)
	for index := range outcome.Members {
		member := &outcome.Members[index]
		for _, result := range results {
			if result.permanent.ObjectID != member.Before.ObjectID {
				continue
			}
			member.Moved, member.Destination = result.moved, result.destination
			if result.moved {
				member.Arrivals = slices.Clone(result.arrivals)
				for _, arrival := range result.arrivals {
					card, exists := c.game.GetCardInstance(arrival.CardID)
					from, located := cardZone(c.game, arrival.CardID)
					if exists && located && card.ZoneVersion == arrival.ZoneVersion &&
						(from == zone.Graveyard || from == zone.Exile && !c.game.Players[card.Owner].Exile.IsFaceDown(card.ID)) {
						member.Cards = append(member.Cards, c.mintFoundCard(card, from))
					}
				}
			}
			break
		}
	}
	return outcome, nil
}

func (c *programContext) CreateTokens(def *game.CardDef, recipients ...cardprogram.TokenRecipient) ([]game.ObjectSnapshot, error) {
	if err := c.valid(); err != nil {
		return nil, err
	}
	previous, _ := c.game.ChoiceContext()
	c.engine.setReplacementChoiceContext(c.game, c.agents, c.log)
	defer c.game.SetChoiceContext(previous)
	if def == nil {
		return nil, errors.New("token creation requires a definition")
	}
	var counts [game.NumPlayers]int
	for _, recipient := range recipients {
		if recipient.Player < 0 || int(recipient.Player) >= game.NumPlayers || recipient.Count < 0 {
			return nil, errors.New("token creation requires valid players and nonnegative counts")
		}
		counts[recipient.Player] += recipient.Count
	}
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		return nil, errors.New("invalid token definition: " + issues[0].Message)
	}
	if err := c.engine.programs.ValidateCard(def); err != nil {
		return nil, err
	}
	var plans []tokenCreationPlan
	for _, player := range triggerAPNAPPlayers(c.game) {
		if counts[player] > 0 && isPlayerAlive(c.game, player) {
			plans = append(plans, planTokenCreation(c.game, player, def, counts[player])...)
		}
	}
	total := 0
	for _, plan := range plans {
		total += plan.count
	}
	batchID := tokenCreationSimultaneousID(c.game, total)
	created, ok := createPlannedTokenEntries(c.engine, c.game, plans, batchID, false, nil, c.agents, c.log)
	if !ok {
		return nil, errors.New("validated token creation failed")
	}
	result := make([]game.ObjectSnapshot, 0, total)
	for _, permanent := range created {
		result = append(result, currentProgramSnapshot(c.game, permanent))
	}
	return result, nil
}

func currentProgramSnapshot(g *game.Game, permanent *game.Permanent) game.ObjectSnapshot {
	snapshot := publicProgramSnapshot(snapshotPermanent(g, permanent, zone.Battlefield))
	for i := range snapshot.ZoneCards {
		card, _ := g.GetCardInstance(snapshot.ZoneCards[i].CardID)
		snapshot.ZoneCards[i].ZoneVersion = card.ZoneVersion
	}
	return snapshot
}

func publicProgramSnapshot(snapshot game.ObjectSnapshot) game.ObjectSnapshot {
	// These are immutable template pointers in engine snapshots, but authoring
	// results expose copied values only. Component identities remain in ZoneCards.
	snapshot.TokenDef, snapshot.CopiableDef, snapshot.MergedCards = nil, nil, nil
	return snapshot
}
