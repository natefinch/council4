package rules

import (
	"fmt"
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

const commanderTotalCardCount = 100

// CommanderLegalityError describes one conservative Commander deck legality
// violation.
type CommanderLegalityError struct {
	Player game.PlayerID
	Reason string
}

func (e CommanderLegalityError) Error() string {
	return fmt.Sprintf("player %d: %s", e.Player, e.Reason)
}

// ValidateCommanderConfigs returns commander legality errors for each configured player.
func ValidateCommanderConfigs(configs [game.NumPlayers]game.PlayerConfig) []CommanderLegalityError {
	var errs []CommanderLegalityError
	for i, config := range configs {
		errs = append(errs, validateCommanderConfig(game.PlayerID(i), config)...)
	}
	return errs
}

func validateCommanderConfig(playerID game.PlayerID, config game.PlayerConfig) []CommanderLegalityError {
	var errs []CommanderLegalityError
	add := func(reason string) {
		errs = append(errs, CommanderLegalityError{Player: playerID, Reason: reason})
	}

	commanders := config.CommanderDefs()
	if len(commanders) == 0 {
		add("missing commander")
		return errs
	}
	if len(commanders) > 2 {
		add(fmt.Sprintf("has %d commanders, want 1 or 2", len(commanders)))
	}
	wantDeckCards := commanderTotalCardCount - len(commanders)
	if len(config.Deck) != wantDeckCards {
		add(fmt.Sprintf("deck has %d cards, want %d", len(config.Deck), wantDeckCards))
	}
	commanderNames := make(map[string]bool)
	var commanderColors []color.Color
	for _, commander := range commanders {
		if commander == nil {
			add("commander list contains nil card")
			continue
		}
		if commanderNames[commander.Name] {
			add(fmt.Sprintf("duplicate commander %q", commander.Name))
		}
		commanderNames[commander.Name] = true
		commanderColors = append(commanderColors, commander.ColorIdentity.Colors()...)
	}
	if len(commanders) == 1 && commanders[0] != nil && !isLegalSingleCommander(commanders[0]) {
		add("commander must be a legendary creature or have permission to be your commander")
	}
	if len(commanders) == 2 && commanders[0] != nil && commanders[1] != nil && !validCommanderPair(commanders[0], commanders[1]) {
		add("commanders must both have partner or form a choose-a-Background pair")
	}
	identity := color.NewIdentity(commanderColors...)
	seen := make(map[string]bool)
	for _, card := range config.Deck {
		if card == nil {
			add("deck contains nil card")
			continue
		}
		if !identity.ContainsAll(card.ColorIdentity) {
			add(fmt.Sprintf("%q has color identity outside commander's color identity", card.Name))
		}
		if commanderNames[card.Name] {
			add(fmt.Sprintf("commander %q is also present in deck", card.Name))
		}
		if card.HasSupertype(types.Basic) {
			continue
		}
		if seen[card.Name] {
			add(fmt.Sprintf("duplicate nonbasic card %q", card.Name))
			continue
		}
		seen[card.Name] = true
	}
	return errs
}

func isLegendaryCreature(card *game.CardDef) bool {
	return card.HasSupertype(types.Legendary) && card.HasType(types.Creature)
}

func isLegalSingleCommander(card *game.CardDef) bool {
	return isLegendaryCreature(card) || card.CanBeCommander
}

func isBackground(card *game.CardDef) bool {
	return card.HasSupertype(types.Legendary) &&
		card.HasType(types.Enchantment) &&
		card.HasSubtype(types.Background)
}

func validCommanderPair(first, second *game.CardDef) bool {
	firstPartnerName := partnerWithName(first)
	secondPartnerName := partnerWithName(second)
	if firstPartnerName != "" && secondPartnerName != "" &&
		firstPartnerName == second.Name && secondPartnerName == first.Name {
		return true
	}
	firstPartner, firstQuality := partnerQuality(first)
	secondPartner, secondQuality := partnerQuality(second)
	if firstPartner && secondPartner &&
		(firstQuality == "" && secondQuality == "" ||
			firstQuality != "" && strings.EqualFold(firstQuality, secondQuality)) {
		return true
	}
	return isLegendaryCreature(first) && first.HasKeyword(game.ChooseABackground) && isBackground(second) ||
		isLegendaryCreature(second) && second.HasKeyword(game.ChooseABackground) && isBackground(first)
}

func partnerWithName(card *game.CardDef) string {
	if !isLegendaryCreature(card) || !card.HasKeyword(game.PartnerWith) {
		return ""
	}
	const prefix = "Partner with "
	for line := range strings.SplitSeq(card.OracleText, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		name := strings.TrimPrefix(line, prefix)
		if reminder := strings.Index(name, " ("); reminder >= 0 {
			name = name[:reminder]
		}
		return strings.TrimSuffix(strings.TrimSpace(name), ".")
	}
	return ""
}

func partnerQuality(card *game.CardDef) (bool, string) {
	if !isLegendaryCreature(card) || !card.HasKeyword(game.Partner) {
		return false, ""
	}
	if card.OracleText == "" {
		return true, ""
	}
	for line := range strings.SplitSeq(card.OracleText, "\n") {
		line = strings.TrimSpace(line)
		if line == "Partner" || strings.HasPrefix(line, "Partner (") {
			return true, ""
		}
		for _, prefix := range []string{"Partner—", "Partner — "} {
			if quality, ok := strings.CutPrefix(line, prefix); ok {
				if reminder := strings.Index(quality, " ("); reminder >= 0 {
					quality = quality[:reminder]
				}
				return true, strings.TrimSpace(quality)
			}
		}
	}
	return false, ""
}

func isCommanderCardID(g *game.Game, cardID id.ID) bool {
	if cardID == 0 {
		return false
	}
	if g.CommanderIDs[cardID] {
		return true
	}
	for _, player := range g.Players {
		if player.CommanderInstanceID == cardID {
			return true
		}
	}
	return false
}

func isCommanderOwnedBy(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	if !isCommanderCardID(g, cardID) {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	return ok && card.Owner == playerID
}

func recordCommanderCast(g *game.Game, playerID game.PlayerID, cardID id.ID) {
	player, ok := playerByID(g, playerID)
	if !ok || !isCommanderOwnedBy(g, playerID, cardID) {
		return
	}
	player.RecordCommanderCast(cardID)
}

func permanentContainsCardMatching(permanent *game.Permanent, matches func(id.ID) bool) bool {
	if permanent == nil {
		return false
	}
	if matches(permanent.CardInstanceID) {
		return true
	}
	for _, merged := range permanent.MergedCards {
		if matches(merged.CardInstanceID) {
			return true
		}
	}
	return false
}

func permanentContainsCommander(g *game.Game, permanent *game.Permanent) bool {
	return permanentContainsCardMatching(permanent, func(cardID id.ID) bool {
		return isCommanderCardID(g, cardID)
	})
}

// commanderPermanent returns the battlefield permanent that currently
// represents the commander, whether the commander is the permanent's own card
// or a card merged beneath it by Mutate.
func commanderPermanent(g *game.Game, commanderID id.ID) (*game.Permanent, bool) {
	for _, permanent := range g.Battlefield {
		if permanentContainsCardMatching(permanent, func(cardID id.ID) bool {
			return cardID == commanderID
		}) {
			return permanent, true
		}
	}
	return nil, false
}

func commanderReplacementDestination(g *game.Game, cardID id.ID, destination zone.Type) zone.Type {
	if !isCommanderCardID(g, cardID) {
		return destination
	}
	switch destination {
	case zone.Graveyard, zone.Exile, zone.Hand, zone.Library:
		return zone.Command
	default:
		return destination
	}
}
