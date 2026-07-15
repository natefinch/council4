package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// handleIterativeLibraryProcess runs the generic iterative library processor
// shared by Tainted Pact, Demonic Consultation, and Tibalt's Trickery. It
// processes cards from the top of the player's library one at a time until the
// configured name predicate fires or the library empties.
//
// The processed-name history lives entirely in local state for one call, so
// independent copies of the same spell never share history and nothing is
// shuffled. When the library empties before the predicate fires the process
// ends with every processed card left exiled.
func handleIterativeLibraryProcess(r *effectResolver, prim game.IterativeLibraryProcess) effectResolved {
	res := effectResolved{accepted: true}
	if prim.PublishLinked != "" {
		clearLinkedObjects(r.game, linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked)))
	}
	playerID, ok := resolvePlayerReference(r.game, r.obj, prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}

	if prim.Stop == game.IterativeLibraryStopDifferentNameNonland {
		return r.iterativeDifferentNameNonland(prim, playerID, player, res)
	}

	// Name a card up front for the chosen-name predicate (Demonic Consultation).
	// A chosen absent name never matches, so the whole library is exiled.
	var chosenName string
	haveChosenName := false
	chosenAbsent := false
	if prim.ChooseName {
		chosenName, chosenAbsent, haveChosenName =
			r.engine.chooseLibraryCardName(r.game, r.agents, r.log, playerID, prim.AllowAbsentName)
	}

	// Exile a fixed count from the top before the loop, without revealing or
	// offering them to hand (Demonic Consultation's "top six cards").
	if pre := r.quantity(prim.PreExile); pre > 0 {
		exileTopOfLibraryCards(r.game, playerID, pre, opt.V[counter.Kind]{}, playerID, false)
		res.succeeded = true
	}

	seenNames := map[string]bool{}
	for {
		cardID, topOK := player.Library.Top()
		if !topOK {
			// Empty library: the process simply ends. For the chosen-name stop
			// the named card was never found, so the whole library stays exiled.
			break
		}
		name := searchCardName(r.game, cardID)
		if prim.Reveal {
			emitCardRevealEvent(r.game, r.obj, playerID, cardID, zone.Library)
		}

		switch prim.Stop {
		case game.IterativeLibraryStopChosenName:
			if haveChosenName && !chosenAbsent && name == chosenName {
				moveProcessedCard(r.game, playerID, cardID, zone.Library, zone.Hand)
				res.succeeded = true
				return res
			}
			moveProcessedCard(r.game, playerID, cardID, zone.Library, zone.Exile)
			res.succeeded = true

		case game.IterativeLibraryStopDuplicateName:
			// Exile the card first: only then is it "exiled this way" and able to
			// match a later duplicate.
			dest := moveProcessedCard(r.game, playerID, cardID, zone.Library, zone.Exile)
			res.succeeded = true
			if seenNames[name] {
				// A duplicate name ends the process; the duplicate stays exiled.
				return res
			}
			seenNames[name] = true
			if prim.OptionalTake &&
				r.engine.chooseMay(r.game, r.agents, playerID, "Put the exiled card into your hand?", r.log) {
				moveProcessedCard(r.game, playerID, cardID, dest, zone.Hand)
				return res
			}

		default:
			// Validation guarantees a known stop condition; terminate safely if
			// an unexpected one ever reaches here.
			return res
		}
	}
	return res
}

// iterativeDifferentNameNonland runs the different-name-nonland stop (Tibalt's
// Trickery). It exiles cards from the top of the player's library one at a time,
// applying the commander-zone replacement, until it exiles a nonland card whose
// front-face name differs from the DifferentNameFrom spell's captured cast-face
// name; lands and same-name nonlands are exiled and the loop continues. When the
// library empties first the process simply ends. PublishLinked records the found
// card first, followed by the other cards exiled by this process, so later
// instructions can independently cast the found card and dispose of the group.
func (r *effectResolver) iterativeDifferentNameNonland(prim game.IterativeLibraryProcess, playerID game.PlayerID, player *game.Player, res effectResolved) effectResolved {
	referenceName, haveName := iterativeStopReferenceName(r.game, r.obj, prim.DifferentNameFrom)
	var exiledThisWay []id.ID
	var found id.ID
	for {
		cardID, topOK := player.Library.Top()
		if !topOK {
			// Empty library: the process ends with every processed card exiled.
			break
		}
		name := searchCardName(r.game, cardID)
		nonland := libraryCardIsNonland(r.game, cardID)
		// Exile the card first: only an exiled card is "exiled this way" and part
		// of the remainder. The commander-zone replacement may divert it instead.
		dest := moveProcessedCard(r.game, playerID, cardID, zone.Library, zone.Exile)
		if dest == zone.Exile {
			exiledThisWay = append(exiledThisWay, cardID)
		}
		if nonland && haveName && name != referenceName {
			// A different-named nonland ends the process. A commander diverted to
			// the command zone can no longer be cast from exile, so only an
			// actually-exiled card becomes the free-cast candidate.
			if dest == zone.Exile {
				found = cardID
				res.succeeded = true
			}
			break
		}
	}
	if prim.PublishLinked != "" {
		key := linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
		if found != 0 {
			publishIterativeCard(r.game, key, found)
		}
		for _, cardID := range exiledThisWay {
			if cardID != found {
				publishIterativeCard(r.game, key, cardID)
			}
		}
	}
	return res
}

func publishIterativeCard(g *game.Game, key game.LinkedObjectKey, cardID id.ID) {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return
	}
	rememberLinkedObject(g, key, game.LinkedObjectRef{
		CardID:          cardID,
		CardZoneVersion: card.ZoneVersion,
	})
}

// iterativeStopReferenceName resolves the name a different-name-nonland stop
// compares against: the cast-face name of the referenced target spell. It
// prefers the name captured in TargetNameLKI when the spell was countered, and
// falls back to the live stack object for a target that could not be countered
// and remains on the stack. It returns ok=false when no name is known, in which
// case no card can differ and the whole library is exiled.
func iterativeStopReferenceName(g *game.Game, obj *game.StackObject, ref game.ObjectReference) (string, bool) {
	if obj == nil || ref.Kind() != game.ObjectReferenceTargetStackObject {
		return "", false
	}
	index := ref.TargetIndex()
	if name, ok := obj.TargetNameLKI[index]; ok {
		return name, true
	}
	objectID, ok := effectStackObjectID(g, obj, index)
	if !ok {
		return "", false
	}
	stackObject, ok := stackObjectByID(g, objectID)
	if !ok {
		return "", false
	}
	return stackSpellName(g, stackObject)
}

// libraryCardIsNonland reports whether a library card is a nonland by its front
// face, the characteristics a card presents outside the battlefield and stack
// (CR 712.4a). A missing card instance is treated as not a stopping card so the
// process continues, matching the other library-dig handlers.
func libraryCardIsNonland(g *game.Game, cardID id.ID) bool {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	return !cardFaceOrDefault(card, game.FaceFront).HasType(types.Land)
}

// moveProcessedCard moves a card from its current zone (from) into the intended
// destination, applying the commander-zone replacement (CR 903.9) that lets a
// commander's owner divert it to the command zone instead of exile, hand,
// graveyard, or library. It removes the card from the source zone, adds it to
// the actual destination zone, reports the zone change, and returns the zone the
// card entered so a follow-up move (Tainted Pact's optional take-to-hand) starts
// from the right place.
func moveProcessedCard(g *game.Game, playerID game.PlayerID, cardID id.ID, from, intended zone.Type) zone.Type {
	sourceOwner := playerID
	if card, ok := g.GetCardInstance(cardID); from == zone.Command && ok {
		sourceOwner = card.Owner
	}
	if sourceCards, ok := destinationZone(g, sourceOwner, from); ok {
		sourceCards.Remove(cardID)
	}
	destination := commanderReplacementDestination(g, cardID, intended)
	zoneOwner := playerID
	if card, ok := g.GetCardInstance(cardID); destination == zone.Command && ok {
		zoneOwner = card.Owner
	}
	destinationCards, ok := destinationZone(g, zoneOwner, destination)
	if !ok {
		return from
	}
	destinationCards.Add(cardID)
	emitZoneChangeEvent(g, game.Event{
		Player:   playerID,
		CardID:   cardID,
		FromZone: from,
		ToZone:   destination,
		Amount:   1,
	})
	return destination
}

// absentLibraryNameLabel is the user-visible label for the sentinel option that
// lets a player deliberately name a card that is not in their library. Selection
// is resolved by option index, not by this label, so it never collides with a
// real card even one literally sharing this text.
const absentLibraryNameLabel = "A card name not in this library"

// chooseLibraryCardName asks the player to name a card for a chosen-name
// iterative process. The bounded, deterministic option set is the distinct card
// names currently in the player's library, sorted so the offered order never
// leaks the library's hidden ordering. When allowAbsent is set, an extra
// sentinel option (absentLibraryNameLabel) is appended at a distinct index; the
// caller detects it structurally by index and treats it as a name the process
// can never match, so the whole library is exiled. Agents that do not answer
// fall back to the first option.
//
// It returns ok=false only when there is nothing to choose: the library holds no
// identifiable card and the absent-name sentinel is not offered, in which case
// the process finds nothing. When allowAbsent is set the choice is always
// offered (absent is chosen for an empty library), so ok is true.
//
// Known limitation (see issue #3044, "name any card" registry/free-text choice):
// a true "name a card" choice lets the player name any card in existence from memory. The engine has
// no card-name registry or free-text choice input, so the concrete real names it
// can offer are the distinct names currently in the player's own library. That
// exposes the library's hidden name set to the choosing agent. The absent-name
// sentinel restores the strategically important "name a card you don't have"
// line (exile the whole library) without widening this leak; closing the leak
// itself needs a registry/free-text choice the engine does not yet have.
func (e *Engine) chooseLibraryCardName(
	g *game.Game,
	agents [game.NumPlayers]PlayerAgent,
	log *TurnLog,
	playerID game.PlayerID,
	allowAbsent bool,
) (name string, absent bool, ok bool) {
	player, present := playerByID(g, playerID)
	if !present {
		return "", false, false
	}
	seen := map[string]bool{}
	var names []string
	for _, cardID := range player.Library.All() {
		cardName := searchCardName(g, cardID)
		if cardName == "" || seen[cardName] {
			continue
		}
		seen[cardName] = true
		names = append(names, cardName)
	}
	if len(names) == 0 && !allowAbsent {
		return "", false, false
	}
	slices.Sort(names)
	options := make([]game.ChoiceOption, 0, len(names)+1)
	for i, cardName := range names {
		options = append(options, game.ChoiceOption{Index: i, Label: cardName})
	}
	// The sentinel occupies its own index past the real names, so it is
	// identified structurally by position rather than by matching its label.
	absentIndex := len(names)
	if allowAbsent {
		options = append(options, game.ChoiceOption{Index: absentIndex, Label: absentLibraryNameLabel})
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           "Choose a card name.",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}
	selected := e.chooseChoice(g, agents, request, log)
	chosen := 0
	if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(options) {
		chosen = selected[0]
	}
	if allowAbsent && chosen == absentIndex {
		return "", true, true
	}
	return names[chosen], false, true
}
