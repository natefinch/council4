package game

import (
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// The constructors below build a ChooseFromZone envelope for each historical
// choose-from-zone family. They are the single compile-time source of truth for
// the field mapping that the retired wrapper primitives (ExileFromHand,
// ExileFromGraveyard, PutFromHand, ReturnFromGraveyard) once carried, so a
// lowerer expresses a family-shaped effect by naming the family rather than
// hand-assembling the envelope. The runtime, validator, and renderer all operate
// on the resulting envelope.

// ExileFromHandChoice builds the envelope for "(you may) exile a <filter> card
// from your hand": the resolving player chooses up to amount matching hand cards
// and each moves to exile. When publishLinked is set the chosen cards are
// imprinted on the source permanent's object identity (Chrome Mox), so a
// re-entered object starts without an imprint.
func ExileFromHandChoice(player PlayerReference, selection Selection, amount Quantity, publishLinked LinkedKey) ChooseFromZone {
	return ChooseFromZone{
		Player:      player,
		SourceZone:  zone.Hand,
		Filter:      selection,
		Quantity:    amount,
		Count:       ChooseExactly,
		Destination: ChooseDestination{Zone: zone.Exile},
		Riders: ChooseRiders{
			PublishLinked:       publishLinked,
			PublishObjectScoped: true,
		},
		Prompt: "Choose a card to exile",
	}
}

// ExileFromGraveyardChoice builds the envelope for "(you may) exile a <filter>
// card from your graveyard": the resolving player chooses up to amount matching
// graveyard cards and each moves to exile. When allOwners is set the candidate
// pool spans every player's graveyard (Cemetery Prowler). When publishLinked is
// set each exiled card is remembered under that source-keyed set for a later
// ability to read.
func ExileFromGraveyardChoice(player PlayerReference, selection Selection, amount Quantity, allOwners bool, publishLinked LinkedKey) ChooseFromZone {
	return ChooseFromZone{
		Player:      player,
		SourceZone:  zone.Graveyard,
		AllOwners:   allOwners,
		Filter:      selection,
		Quantity:    amount,
		Count:       ChooseExactly,
		Destination: ChooseDestination{Zone: zone.Exile},
		Riders:      ChooseRiders{PublishLinked: publishLinked},
		Prompt:      "Choose a card to exile",
	}
}

// PutFromHandChoice builds the envelope for "put a <filter> card from your hand
// onto the battlefield": the resolving player chooses matching hand cards and
// each enters the battlefield under that player's control, tapped when
// entersTapped is set and attacking when entersAttacking is set (CR 508.4, the
// "tapped and attacking" rider). When anyNumber is set the player may choose any
// number of matching cards from none up to all of them ("put any number of
// creature cards from your hand onto the battlefield", Ghalta); amount is ignored
// for that form. Otherwise the player chooses exactly amount matching cards.
func PutFromHandChoice(player PlayerReference, selection Selection, amount Quantity, entersTapped, entersAttacking, anyNumber bool) ChooseFromZone {
	count := ChooseExactly
	if anyNumber {
		// The any-number form ignores Quantity; the validator requires a zero
		// amount, so force it here regardless of what the caller passed.
		count = ChooseAnyNumber
		amount = Quantity{}
	}
	return ChooseFromZone{
		Player:      player,
		SourceZone:  zone.Hand,
		Filter:      selection,
		Quantity:    amount,
		Count:       count,
		Destination: ChooseDestination{Zone: zone.Battlefield},
		Riders:      ChooseRiders{EntersTapped: entersTapped, EntersAttacking: entersAttacking},
		Prompt:      "Choose a card to put onto the battlefield",
	}
}

// ReturnFromGraveyardChoice builds the envelope for "Return a <filter> card from
// your graveyard to your hand/the battlefield": the resolving player chooses up
// to amount matching graveyard cards (any number when anyNumber is set, capped by
// maxTotalManaValue) and each returns to its owner's hand or, for a battlefield
// destination, enters under the player's control tapped when entryTapped is set.
// fromLinked restricts the candidate pool to an earlier-produced linked set.
func ReturnFromGraveyardChoice(player PlayerReference, selection Selection, amount Quantity, destination zone.Type, entryTapped bool, maxTotalManaValue opt.V[int], anyNumber bool, fromLinked LinkedKey) ChooseFromZone {
	dest := zone.Hand
	prompt := "Choose a card to return to your hand"
	if destination == zone.Battlefield {
		dest = zone.Battlefield
		prompt = "Choose a card to return to the battlefield"
	}
	count := ChooseExactly
	if anyNumber {
		count = ChooseAnyNumber
	}
	return ChooseFromZone{
		Player:      player,
		SourceZone:  zone.Graveyard,
		Filter:      selection,
		Quantity:    amount,
		Count:       count,
		Destination: ChooseDestination{Zone: dest},
		Riders: ChooseRiders{
			EntersTapped:      entryTapped,
			MaxTotalManaValue: maxTotalManaValue,
			FromLinked:        fromLinked,
		},
		Prompt: prompt,
	}
}
