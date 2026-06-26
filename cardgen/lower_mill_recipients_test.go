package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// millSpell builds a minimal sorcery card carrying the given mill oracle text so
// a recipient/amount test can lower it through the executable backend.
func millSpell(name, oracle string) *ScryfallCard {
	return &ScryfallCard{
		Name:       name,
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Sorcery",
		OracleText: oracle,
	}
}

// millCreature builds a minimal creature card carrying the given mill oracle
// text so a triggered-recipient test can lower it.
func millCreature(name, oracle string) *ScryfallCard {
	power, toughness := "2", "2"
	return &ScryfallCard{
		Name:       name,
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Creature — Wizard",
		OracleText: oracle,
		Power:      &power,
		Toughness:  &toughness,
	}
}

func soleSpellMill(t *testing.T, face loweredFaceAbilities) game.Mill {
	t.Helper()
	if !face.SpellAbility.Exists || len(face.SpellAbility.Val.Modes) == 0 {
		t.Fatal("expected a spell ability with at least one mode")
	}
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("spell sequence length = %d, want 1", len(sequence))
	}
	mill, ok := sequence[0].Primitive.(game.Mill)
	if !ok {
		t.Fatalf("spell primitive = %#v, want Mill", sequence[0].Primitive)
	}
	return mill
}

func soleTriggeredMill(t *testing.T, face loweredFaceAbilities) game.Mill {
	t.Helper()
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	modes := face.TriggeredAbilities[0].Content.Modes
	if len(modes) == 0 {
		t.Fatal("triggered ability has no modes")
	}
	sequence := modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("triggered sequence length = %d, want 1", len(sequence))
	}
	mill, ok := sequence[0].Primitive.(game.Mill)
	if !ok {
		t.Fatalf("triggered primitive = %#v, want Mill", sequence[0].Primitive)
	}
	return mill
}

// TestLowerMillSpelledFixedAmounts proves a spelled-out mill count above the
// legacy four-card digit ceiling ("mills five cards") now reconstructs and
// lowers to an exact fixed Mill, the broadening unlocked by exempting EffectMill
// from exactLegacyFixedAmountSyntax. The same gate previously forced
// effect.Exact = false and failed the spell closed.
func TestLowerMillSpelledFixedAmounts(t *testing.T) {
	t.Parallel()

	t.Run("target player spelled amount", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, millSpell("Test Tome Scour", "Target player mills five cards."))
		mode := face.SpellAbility.Val.Modes[0]
		if len(mode.Targets) != 1 {
			t.Fatalf("targets = %d, want 1", len(mode.Targets))
		}
		mill := soleSpellMill(t, face)
		if mill.Amount.IsDynamic() || mill.Amount.Value() != 5 {
			t.Fatalf("mill amount = %#v, want fixed 5", mill.Amount)
		}
		if mill.Player != game.TargetPlayerReference(0) {
			t.Fatalf("mill player = %v, want TargetPlayerReference(0)", mill.Player)
		}
	})

	t.Run("controller self mill seven", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, millSpell("Test Self Mill", "Mill seven cards."))
		mill := soleSpellMill(t, face)
		if mill.Amount.IsDynamic() || mill.Amount.Value() != 7 {
			t.Fatalf("mill amount = %#v, want fixed 7", mill.Amount)
		}
		if mill.Player != game.ControllerReference() {
			t.Fatalf("mill player = %v, want ControllerReference", mill.Player)
		}
	})
}

// TestLowerMillEachPlayerGroups proves the group recipients ("each player",
// "each opponent") lower to the player-group Mill form rather than a single
// player.
func TestLowerMillEachPlayerGroups(t *testing.T) {
	t.Parallel()

	t.Run("each player", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, millSpell("Test Each Player Mill", "Each player mills thirteen cards."))
		mill := soleSpellMill(t, face)
		if mill.PlayerGroup != game.AllPlayersReference() {
			t.Fatalf("mill group = %v, want AllPlayersReference", mill.PlayerGroup)
		}
		if mill.Amount.Value() != 13 {
			t.Fatalf("mill amount = %#v, want fixed 13", mill.Amount)
		}
	})

	t.Run("each opponent", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, millSpell("Test Each Opponent Mill", "Each opponent mills five cards."))
		mill := soleSpellMill(t, face)
		if mill.PlayerGroup != game.OpponentsReference() {
			t.Fatalf("mill group = %v, want OpponentsReference", mill.PlayerGroup)
		}
	})
}

// TestLowerMillDefendingPlayer proves "defending player mills N." on a combat
// trigger binds the recipient to DefendingPlayerReference, the attacked player
// carried by the triggering attack/blocked event (Flint Golem, Nemesis of
// Reason).
func TestLowerMillDefendingPlayer(t *testing.T) {
	t.Parallel()

	t.Run("becomes blocked", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, millCreature("Test Flint Golem",
			"Whenever this creature becomes blocked, defending player mills three cards."))
		if got := face.TriggeredAbilities[0].Trigger.Pattern.Event; got != game.EventAttackerBecameBlocked {
			t.Fatalf("trigger event = %v, want EventAttackerBecameBlocked", got)
		}
		mill := soleTriggeredMill(t, face)
		if mill.Amount.Value() != 3 {
			t.Fatalf("mill amount = %#v, want fixed 3", mill.Amount)
		}
		if mill.Player != game.DefendingPlayerReference() {
			t.Fatalf("mill player = %v, want DefendingPlayerReference", mill.Player)
		}
	})

	t.Run("attacks spelled ten", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, millCreature("Test Nemesis",
			"Whenever this creature attacks, defending player mills ten cards."))
		mill := soleTriggeredMill(t, face)
		if mill.Amount.Value() != 10 {
			t.Fatalf("mill amount = %#v, want fixed 10", mill.Amount)
		}
		if mill.Player != game.DefendingPlayerReference() {
			t.Fatalf("mill player = %v, want DefendingPlayerReference", mill.Player)
		}
	})
}

// TestLowerMillEventPermanentController proves "that permanent's controller
// mills N." / "its controller mills N." binds the recipient to the controller
// of the triggering event permanent (Mesmeric Orb, Chronic Flooding).
func TestLowerMillEventPermanentController(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mesmeric Orb",
		Layout:     "normal",
		ManaCost:   "{2}",
		TypeLine:   "Artifact",
		OracleText: "Whenever a permanent becomes untapped, that permanent's controller mills a card.",
	})
	mill := soleTriggeredMill(t, face)
	if mill.Amount.Value() != 1 {
		t.Fatalf("mill amount = %#v, want fixed 1", mill.Amount)
	}
	want := game.ObjectControllerReference(game.EventPermanentReference())
	if mill.Player != want {
		t.Fatalf("mill player = %v, want ObjectControllerReference(EventPermanentReference)", mill.Player)
	}
}

// TestLowerMillEventQuantityThatMany proves a "that many" triggering-event
// anaphor resolves to the firing event's quantity: the life lost on a life-loss
// trigger (Mindcrank) and the combat damage dealt on a combat-damage trigger
// (Crosstown Courier), each milled by the event's own player.
func TestLowerMillEventQuantityThatMany(t *testing.T) {
	t.Parallel()

	t.Run("life loss that many (Mindcrank)", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test Mindcrank",
			Layout:     "normal",
			ManaCost:   "{2}",
			TypeLine:   "Artifact",
			OracleText: "Whenever an opponent loses life, that player mills that many cards.",
		})
		if got := face.TriggeredAbilities[0].Trigger.Pattern.Event; got != game.EventLifeLost {
			t.Fatalf("trigger event = %v, want EventLifeLost", got)
		}
		mill := soleTriggeredMill(t, face)
		dynamic := mill.Amount.DynamicAmount()
		if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountEventLifeChange {
			t.Fatalf("mill amount = %#v, want dynamic event life change", mill.Amount)
		}
		if mill.Player != game.EventPlayerReference() {
			t.Fatalf("mill player = %v, want EventPlayerReference", mill.Player)
		}
	})

	t.Run("combat damage that many (Crosstown Courier)", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, millCreature("Test Crosstown Courier",
			"Whenever this creature deals combat damage to a player, that player mills that many cards."))
		mill := soleTriggeredMill(t, face)
		dynamic := mill.Amount.DynamicAmount()
		if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountEventDamage {
			t.Fatalf("mill amount = %#v, want dynamic event damage", mill.Amount)
		}
		if mill.Player != game.EventPlayerReference() {
			t.Fatalf("mill player = %v, want EventPlayerReference", mill.Player)
		}
	})
}

// TestLowerMillControllerLifeStillUnsupported pins the deliberately deferred
// "mills cards equal to your life total" form (Space-Time Anomaly) as fail-
// closed, guarding against an accidental future miscompilation of a controller-
// life mill amount the runtime does not yet model for this recipient shape.
func TestLowerMillControllerLifeStillUnsupported(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, millSpell("Test Space-Time Anomaly",
		"Target player mills cards equal to your life total."))
}
