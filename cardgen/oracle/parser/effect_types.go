package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// EffectKind identifies a resolving instruction. The parser owns the Oracle
// vocabulary which selects these values; consumers only map the typed value.
type EffectKind string

// Resolving effect kinds recognized by the parser.
const (
	EffectUnknown           EffectKind = ""
	EffectAddMana           EffectKind = "EffectAddMana"
	EffectAttach            EffectKind = "EffectAttach"
	EffectCantBeBlocked     EffectKind = "EffectCantBeBlocked"
	EffectCantBlock         EffectKind = "EffectCantBlock"
	EffectCast              EffectKind = "EffectCast"
	EffectCounter           EffectKind = "EffectCounter"
	EffectCreate            EffectKind = "EffectCreate"
	EffectDealDamage        EffectKind = "EffectDealDamage"
	EffectDestroy           EffectKind = "EffectDestroy"
	EffectDig               EffectKind = "EffectDig"
	EffectDiscard           EffectKind = "EffectDiscard"
	EffectDiscover          EffectKind = "EffectDiscover"
	EffectDouble            EffectKind = "EffectDouble"
	EffectDraw              EffectKind = "EffectDraw"
	EffectEnterTapped       EffectKind = "EffectEnterTapped"
	EffectEnterPrepared     EffectKind = "EffectEnterPrepared"
	EffectExile             EffectKind = "EffectExile"
	EffectFight             EffectKind = "EffectFight"
	EffectGain              EffectKind = "EffectGain"
	EffectGainControl       EffectKind = "EffectGainControl"
	EffectBecomeMonarch     EffectKind = "EffectBecomeMonarch"
	EffectGrantKeyword      EffectKind = "EffectGrantKeyword"
	EffectInvestigate       EffectKind = "EffectInvestigate"
	EffectImpulseExile      EffectKind = "EffectImpulseExile"
	EffectExplore           EffectKind = "EffectExplore"
	EffectLose              EffectKind = "EffectLose"
	EffectManifest          EffectKind = "EffectManifest"
	EffectManifestDread     EffectKind = "EffectManifestDread"
	EffectMill              EffectKind = "EffectMill"
	EffectManaSpendRider    EffectKind = "EffectManaSpendRider"
	EffectModifyPT          EffectKind = "EffectModifyPT"
	EffectGainPlayerCounter EffectKind = "EffectGainPlayerCounter"
	EffectPut               EffectKind = "EffectPut"
	EffectProliferate       EffectKind = "EffectProliferate"
	EffectRemoveCounter     EffectKind = "EffectRemoveCounter"
	EffectRegenerate        EffectKind = "EffectRegenerate"
	EffectReorderLibraryTop EffectKind = "EffectReorderLibraryTop"
	EffectReturn            EffectKind = "EffectReturn"
	EffectReveal            EffectKind = "EffectReveal"
	EffectSacrifice         EffectKind = "EffectSacrifice"
	EffectScry              EffectKind = "EffectScry"
	EffectSurveil           EffectKind = "EffectSurveil"
	EffectSearch            EffectKind = "EffectSearch"
	EffectShuffle           EffectKind = "EffectShuffle"
	EffectTap               EffectKind = "EffectTap"
	EffectTapOrUntap        EffectKind = "EffectTapOrUntap"
	EffectUntap             EffectKind = "EffectUntap"
	EffectTransform         EffectKind = "EffectTransform"
	EffectPreventDamage     EffectKind = "EffectPreventDamage"
)

const (
	// EffectLifeTotalCantChange models an immutable player life total.
	EffectLifeTotalCantChange EffectKind = "EffectLifeTotalCantChange"
	// EffectProtectionFromEverything models a player gaining protection from everything.
	EffectProtectionFromEverything EffectKind = "EffectProtectionFromEverything"
	// EffectPhaseOut models permanents phasing out.
	EffectPhaseOut EffectKind = "EffectPhaseOut"
	// EffectAdditionalLandPlays models the controller-scoped grant of one or more
	// extra land plays for the turn ("Play an additional land this turn.", "You
	// may play two additional lands this turn.").
	EffectAdditionalLandPlays EffectKind = "EffectAdditionalLandPlays"
	// EffectLoseGame models a player losing the game ("you lose the game"), as in
	// the unpaid consequence of a Pact upkeep cost (CR 104.3a).
	EffectLoseGame EffectKind = "EffectLoseGame"
	// EffectChooseNewTargets models re-choosing the targets of a spell or
	// ability on the stack ("You may choose new targets for target spell or
	// ability."), the Deflecting Swat / Redirect retarget family (CR 115.7).
	EffectChooseNewTargets EffectKind = "EffectChooseNewTargets"
	// EffectCastAsThoughFlash models the controller-scoped, turn-scoped timing
	// permission "You may cast spells this turn as though they had flash."
	// (Borne Upon a Wind, Emergence Zone), letting the controller cast spells at
	// instant speed for the rest of the turn (CR 702.8 / 601.3e).
	EffectCastAsThoughFlash EffectKind = "EffectCastAsThoughFlash"
	// EffectPlayFromLibraryTop models the controller-scoped, turn-scoped grant
	// "until end of turn, you may look at the top card of your library any time
	// and you may play cards from the top of your library." (Gwenom,
	// Remorseless), letting the controller privately look at and play (lands or
	// cast spells) the top card of their library for the rest of the turn. The
	// optional "If you cast a spell this way, pay life equal to its mana value
	// rather than pay its mana cost." rider sets PlayFromTopPayLife.
	EffectPlayFromLibraryTop EffectKind = "EffectPlayFromLibraryTop"
	// EffectCantCastSpells models the one-shot, turn-scoped player cast
	// prohibition "<players> can't cast spells this turn." (Silence: "Your
	// opponents can't cast spells this turn."), forbidding the affected players
	// from casting spells for the rest of the turn. The affected players are the
	// controller's opponents ("your opponents", "each opponent") or every player
	// ("players"). It reuses the same rule effect as the continuous static form
	// (RuleEffectCantCastSpells), applied with a this-turn duration.
	EffectCantCastSpells EffectKind = "EffectCantCastSpells"
	// EffectWinGame models a player winning the game ("you win the game"), as in
	// Felidar Sovereign and Thassa's Oracle (CR 104.2a). It mirrors
	// EffectLoseGame.
	EffectWinGame EffectKind = "EffectWinGame"
	// EffectSpellsCantBeCountered models the controller-scoped, turn-scoped
	// resolving buff "The next spell you cast this turn can't be countered."
	// (Mistrise Village) and the all-spells form "Spells you cast this turn
	// can't be countered." (Domri, Anarch of Bolas).
	EffectSpellsCantBeCountered EffectKind = "EffectSpellsCantBeCountered"
	// EffectEnterAsCopy models a self enters-the-battlefield replacement that has
	// the permanent enter as a copy of another permanent chosen as it enters
	// ("You may have this creature enter the battlefield as a copy of any creature
	// on the battlefield.", Clone), CR 706.
	EffectEnterAsCopy EffectKind = "EffectEnterAsCopy"
	// EffectBecomeCopy models an activated/resolving copy effect that has the
	// source permanent become a copy of a targeted permanent ("This land becomes
	// a copy of target land, except it has this ability.", Thespian's Stage;
	// "This artifact becomes a copy of target ... until end of turn.", Mirage
	// Mirror), CR 706. The copied-permanent target is carried as an ordinary
	// "target" selector; the duration and copiable exception riders are carried
	// in the BecomeCopy* fields.
	EffectBecomeCopy EffectKind = "EffectBecomeCopy"
	// EffectMassReanimationExchange models the symmetric mass-reanimation
	// sentence "Each player exiles all <type> cards from their graveyard, then
	// sacrifices all <type> they control, then puts all cards they exiled this
	// way onto the battlefield." (Living Death, Living End, Scrap Mastery). The
	// three clauses act atomically per player: the matching graveyard cards are
	// exiled first (so the cards sacrificed in the second step are not caught by
	// the third), then every matching permanent is sacrificed, then the
	// just-exiled cards enter the battlefield under their owners' control. The
	// card-type filter (creature or artifact) is carried in the effect's
	// Selection.
	EffectMassReanimationExchange EffectKind = "EffectMassReanimationExchange"
	// EffectPunisherLoseLife models the "punisher" family ("Each opponent loses
	// N life unless that player sacrifices a permanent of their choice or
	// discards a card."). The life amount is in Amount, the player group in
	// Context, the optional sacrifice filter in Selection, and PunisherSacrifice
	// / PunisherDiscard record which alternatives are offered.
	EffectPunisherLoseLife EffectKind = "EffectPunisherLoseLife"
	// EffectMoveCounters models moving counters off the source permanent onto a
	// single target permanent ("Move a +1/+1 counter from this creature onto
	// target creature.", "Move all counters from this permanent onto target
	// creature."). The source is the effect's own permanent (a self reference),
	// the destination is the single target, the counter kind is in CounterKind /
	// CounterKnown (unset for the kind-agnostic "all counters" form), and the
	// fixed count is in Amount. MoveCountersAll records the "all counters" form.
	EffectMoveCounters EffectKind = "EffectMoveCounters"
	// EffectMustAttack models the one-shot, turn-scoped forced-attack effect
	// "<group> attack this turn if able." (Bident of Thassa: "Creatures your
	// opponents control attack this turn if able."). The affected creature group
	// is carried in StaticSubject (creatures you control, creatures your
	// opponents control, or all creatures); lowering reads it to scope the
	// continuous RuleEffectMustAttack rule effect, applied with a this-turn
	// duration.
	EffectMustAttack EffectKind = "EffectMustAttack"
	// EffectRepeatProcess models a "Repeat the following process X times.
	// <body>" loop. Amount holds the repeat count (the spell's {X} via VariableX
	// or a fixed cardinal) and RepeatBody holds the sub-effect(s) executed each
	// iteration.
	EffectRepeatProcess EffectKind = "EffectRepeatProcess"
	// EffectCopyStackObject models the resolving effect "Copy target [activated
	// or ]triggered ability you control[. You may choose new targets for the
	// copy]." The targeted stack object is the activated or triggered ability to
	// copy; CopyMayChooseNewTargets records the optional retarget rider. The copy
	// is put on the stack and resolves independently (CR 707, 706.10).
	EffectCopyStackObject EffectKind = "EffectCopyStackObject"
	// EffectAmass models the amass keyword action (CR 701.44): "Amass Orcs N" /
	// "Amass Zombies N" / "Amass N". The controller puts N +1/+1 counters on an
	// Army they control, first creating a 0/0 black Army creature token of the
	// named subtype (AmassSubtype) if they control no Army. Amount holds N.
	EffectAmass EffectKind = "EffectAmass"
	// EffectRenown models the renown keyword action that the printed "Renown N"
	// keyword expands to (CR 702.111): the source permanent gets N +1/+1 counters
	// and becomes renowned, applied only once. Amount holds N. The runtime guard
	// (it skips an already-renowned permanent) subsumes the printed "if it isn't
	// renowned" intervening-if, so the expanded trigger body is the bare action.
	EffectRenown EffectKind = "EffectRenown"
	// EffectDevour models the Devour keyword's as-enters replacement (CR 702.81):
	// "As this creature enters, you may sacrifice any number of creatures. It
	// enters with N +1/+1 counters on it for each creature sacrificed." It is
	// produced only by the keyword expansion (expandDevourKeyword) and the
	// parseDevourEffect recognizer; the per-sacrificed-creature counter
	// multiplier N is carried in EntersDevourMultiplier.
	EffectDevour EffectKind = "EffectDevour"

	// EffectTribute models the Tribute keyword's as-enters replacement (CR
	// 702.110): "As this creature enters, an opponent of your choice may put N
	// +1/+1 counters on it." It is produced only by the keyword expansion
	// (expandTributeKeyword) and the parseTributeEffect recognizer; the counter
	// count N is carried in EntersTributeCount.
	EffectTribute EffectKind = "EffectTribute"

	// EffectChooseCreatureType models the resolution-time choice "Choose a
	// creature type." made as a spell or ability resolves (CR 614.12 analogue for
	// resolving abilities). It publishes the chosen subtype so a later effect in
	// the same resolution that counts permanents "of that type" can read it, as in
	// "Choose a creature type. Draw a card for each permanent you control of that
	// type." (Distant Melody). It lowers to a game.Choose instruction whose
	// ResolutionChoice is a ResolutionChoiceSubtype over creature types.
	EffectChooseCreatureType EffectKind = "EffectChooseCreatureType"
	// EffectNoMaximumHandSize models the controller-scoped, rest-of-game
	// continuous effect "You have no maximum hand size for the rest of the game."
	// (Sea Gate Restoration). As a resolving spell effect it removes the
	// controller's maximum hand size permanently; it lowers to an ApplyRule
	// carrying the continuous RuleEffectNoMaximumHandSize with a permanent
	// duration. The permanent static "You have no maximum hand size." form
	// (Reliquary Tower) is a static declaration, not this resolving effect.
	EffectNoMaximumHandSize EffectKind = "EffectNoMaximumHandSize"
	// EffectAdditionalCombatPhase models the extra-phase-insertion effect "After
	// this [main] phase, there is an additional combat phase[ followed by an
	// additional main phase]." (Aggravated Assault, Aurelia the Warleader, World
	// at War, Combat Celebrant). It inserts an additional combat phase into the
	// current turn, optionally followed by an additional main phase.
	// AdditionalCombatPhase and AdditionalMainPhase carry which phases are added.
	EffectAdditionalCombatPhase EffectKind = "EffectAdditionalCombatPhase"
	// EffectLookAtHand models the private hand-inspection effect "Look at target
	// player's hand." (Gitaxian Probe, Peek, Telepathy-adjacent). The source's
	// controller looks at the targeted player's hand; it conveys hidden
	// information only and changes no game state. The targeted player is carried
	// as a player target on the effect.
	EffectLookAtHand EffectKind = "EffectLookAtHand"
	// EffectRollDie models "roll a d<N>" (CR 706): it rolls a single fair die
	// with DieSides faces and publishes the rolled value as its resolved amount.
	// A later effect consumes the value via an "equal to the result"
	// dynamic-amount (EffectAmountDieRollResult). DieSides carries N.
	EffectRollDie EffectKind = "EffectRollDie"
	// EffectRemoveFromCombat models the resolving effect "Remove target <attacking
	// or blocking> creature from combat." (Reconnaissance), taking the referenced
	// creature out of combat so it stops being an attacker or blocker and deals
	// and is dealt no further combat damage (CR 506.4). The creature is carried as
	// an ordinary permanent target on the effect.
	EffectRemoveFromCombat EffectKind = "EffectRemoveFromCombat"
	// EffectExileIfLeaveBattlefield models the leaves-the-battlefield
	// self-replacement "If it would leave the battlefield, exile it instead of
	// putting it anywhere else." (Whip of Erebos, Kheru Lich Lord-adjacent). It
	// redirects any zone change off the battlefield for the affected object to
	// exile instead. The affected object is identified by Context: "it" refers
	// to a back-referenced object (EffectContextReferencedObject), while "this
	// creature" refers to the source. It lowers to a CreateReplacement bound to
	// that object.
	EffectExileIfLeaveBattlefield EffectKind = "EffectExileIfLeaveBattlefield"
	// EffectBecomeType models a targeted continuous type-adding effect ("Target
	// permanent becomes an artifact in addition to its other types until end of
	// turn.", Liquimetal Torque, Liquimetal Coating). The added card types are
	// carried in BecomeTypeAddTypes and the targeted permanent is left as an
	// ordinary target for the target machinery. It lowers to an ApplyContinuous
	// at LayerType that adds the types to the target for the recorded duration.
	EffectBecomeType EffectKind = "EffectBecomeType"
	// EffectPolymorph models a targeted resolving polymorph effect ("Until end of
	// turn, target creature loses all abilities and becomes a <color> <subtype>
	// with base power and toughness N/N.", Turn to Frog, Snakeform, Gift of
	// Tusks). The set colors, creature subtype, and base power/toughness are
	// carried in the Polymorph* fields; the targeted creature is left as an
	// ordinary target for the target machinery. It lowers to an ApplyContinuous
	// that removes all abilities and sets the creature's color, types, and base
	// power/toughness until end of turn.
	EffectPolymorph EffectKind = "EffectPolymorph"
	// EffectDelayedTrigger models an event-based delayed triggered ability
	// created for the rest of the turn ("Whenever you cast a spell this turn,
	// ...", Showdown of the Skalds chapter II) or for the next matching event
	// ("When you next cast a creature spell this turn, ...", Summon: Fenrir
	// chapter II). The preamble names a trigger event bounded by "this turn";
	// the body after the comma is the ability that fires on each match. The
	// whole sentence is reparsed as a nested triggered ability stored in
	// DelayedTriggerAbility, and DelayedTriggerOneShot records the "next"
	// surface form that fires only on the first match. It lowers to a
	// game.CreateDelayedTrigger carrying the nested ability's trigger pattern
	// and content, scoped to the turn.
	EffectDelayedTrigger EffectKind = "EffectDelayedTrigger"
	// EffectRingTempts models the fixed designation effect "The Ring tempts
	// you." (CR 701.51). The resolving controller gets the Ring emblem if they
	// don't already have it, advances it to the next of its four levels, and
	// chooses a creature they control to become (or remain) their Ring-bearer.
	// The wording is fully fixed, so the parser recognizes the whole sentence
	// and the effect carries no targets or amounts.
	EffectRingTempts EffectKind = "EffectRingTempts"
)

// DigSourceKind identifies how an impulse "Put N <source> into your hand ..."
// clause refers back to the looked-at cards ("of them" / "of those cards"). It
// is recorded so the exactness recognizer reconstructs the clause byte-for-byte.
type DigSourceKind string

// Recognized impulse take-source phrasings.
const (
	DigSourceNone       DigSourceKind = ""
	DigSourceThem       DigSourceKind = "DigSourceThem"
	DigSourceThoseCards DigSourceKind = "DigSourceThoseCards"
)

// DigRemainderKind identifies where an impulse put clause sends the un-taken
// cards ("the rest" / "the other"). It is recorded so the exactness recognizer
// reconstructs the clause byte-for-byte; the library-bottom variants are kept
// distinct from one another only to preserve their printed ordering rider, and
// all of them lower to the same runtime library-bottom remainder.
type DigRemainderKind string

// Recognized impulse remainder destinations. DigRemainderGraveyard is the zero
// value so existing graveyard digs keep their wire form.
const (
	DigRemainderGraveyard           DigRemainderKind = ""
	DigRemainderLibraryBottom       DigRemainderKind = "DigRemainderLibraryBottom"
	DigRemainderLibraryBottomAny    DigRemainderKind = "DigRemainderLibraryBottomAny"
	DigRemainderLibraryBottomRandom DigRemainderKind = "DigRemainderLibraryBottomRandom"
)

// DigSyntax holds the structured fields of an impulse "Put N <source> into your
// hand and the <rest|other> <remainder>." clause. The parser sets it only on the
// EffectPut clause that follows an EffectDig "Look at the top N cards of your
// library." sentence; Put is false for every other effect. Remainder records the
// destination of the un-taken cards: the controller's graveyard (the default) or
// the bottom of their library, optionally with an "in any order" / "in a random
// order" rider that the runtime models as a single library-bottom placement.
type DigSyntax struct {
	// Put reports that this EffectPut clause is the put half of an impulse dig.
	Put bool `json:",omitempty"`
	// Source is the "of them" / "of those cards" back-reference phrasing.
	Source DigSourceKind `json:",omitempty"`
	// Singular reports the "the other" wording (exactly one card remains) rather
	// than "the rest". It is cosmetic: both route the remainder identically.
	Singular bool `json:",omitempty"`
	// Remainder is the destination of the un-taken cards. The zero value routes
	// them to the controller's graveyard.
	Remainder DigRemainderKind `json:",omitempty"`
	// Take is the number of cards put into the controller's hand. It is set only
	// on the single-effect draw-replacement dig form (parseDrawReplacementDig),
	// which folds both the look count (in the effect Amount) and this take count
	// into one effect. The two-sentence impulse form leaves it zero and carries
	// its take count in the EffectPut clause's own Amount.
	Take int `json:",omitempty"`
}

// HandLibraryPutSyntax marks the exact clause "Put N cards from your hand on
// top of your library in any order." The fixed amount remains in EffectSyntax;
// Present carries the player-chosen ordering semantics downstream without
// requiring consumers to inspect Oracle text.
type HandLibraryPutSyntax struct {
	Present bool `json:",omitempty"`
}

// HandDiscardSyntax marks an exact fixed-cardinality discard of cards from the
// resolving controller's hand. Present excludes targeted, opponent, typed-card,
// and variable-cardinality discard forms. AtRandom marks the "at random" variant
// ("Discard a card at random."), where the cards leave the hand by random
// selection rather than the player's choice.
type HandDiscardSyntax struct {
	Present  bool `json:",omitempty"`
	AtRandom bool `json:",omitempty"`
}

// HandChoiceDiscardSyntax holds the structured fields of the "You choose a
// [filter] card from it." middle sentence of a "Target player reveals their
// hand. You choose a [filter] card from it. That player discards that card."
// sequence (Duress, Thoughtseize, Coercion, Inquisition of Kozilek). Present
// marks the parsed sequence; ExcludeCreature and ExcludeLand carry the
// "noncreature" / "nonland" filter; MaxManaValue, when HasMaxManaValue is set,
// carries the "with mana value N or less" bound. ChooseSpan covers the middle
// sentence so the text-blind lowering can credit its tokens toward source
// coverage. The parser sets this only on the EffectDiscard half of the
// sequence; it is the zero value for every other effect.
type HandChoiceDiscardSyntax struct {
	Present         bool        `json:",omitempty"`
	ExcludeCreature bool        `json:",omitempty"`
	ExcludeLand     bool        `json:",omitempty"`
	HasMaxManaValue bool        `json:",omitempty"`
	MaxManaValue    int         `json:",omitempty"`
	ChooseSpan      shared.Span `json:"-"`
}

// SearchSplitSlot is one single-card destination slot of a split-destination
// library-search put clause. ToZone is the destination zone (hand or
// battlefield); EntersTapped reports the "tapped" rider on a battlefield slot.
type SearchSplitSlot struct {
	ToZone       zone.Type `json:",omitempty"`
	EntersTapped bool      `json:",omitempty"`
}

// SearchSplitSyntax holds the structured fields of a split-destination put
// clause "put one <slot> and the other <slot>." that distributes the cards found
// by a preceding "up to two" library search across two single-card destination
// slots. The parser sets it only on the EffectPut clause of such a search;
// Present is false for every other effect. First is the slot the "one" card
// fills and Second is the slot the "other" card fills, matching source order;
// both must be a hand or battlefield destination, modeling Cultivate and
// Kodama's Reach.
type SearchSplitSyntax struct {
	// Present reports that this EffectPut clause is a recognized split put.
	Present bool            `json:",omitempty"`
	First   SearchSplitSlot `json:",omitzero"`
	Second  SearchSplitSlot `json:",omitzero"`
}

// EffectDurationKind identifies a resolving effect's duration.
type EffectDurationKind string

// Resolving effect durations recognized by the parser.
const (
	EffectDurationNone                     EffectDurationKind = ""
	EffectDurationUntilEndOfTurn           EffectDurationKind = "EffectDurationUntilEndOfTurn"
	EffectDurationUntilYourNextTurn        EffectDurationKind = "EffectDurationUntilYourNextTurn"
	EffectDurationUntilEndOfYourNextTurn   EffectDurationKind = "EffectDurationUntilEndOfYourNextTurn"
	EffectDurationThisTurn                 EffectDurationKind = "EffectDurationThisTurn"
	EffectDurationThisCombat               EffectDurationKind = "EffectDurationThisCombat"
	EffectDurationWhileSourceOnBattlefield EffectDurationKind = "EffectDurationWhileSourceOnBattlefield"
	EffectDurationWhileYouControlSource    EffectDurationKind = "EffectDurationWhileYouControlSource"
	// EffectDurationWhileControlledCreatureEnchanted matches the
	// attachment-dependent wording "for as long as that creature is enchanted".
	// The effect expires when the affected creature is no longer enchanted.
	EffectDurationWhileControlledCreatureEnchanted EffectDurationKind = "EffectDurationWhileControlledCreatureEnchanted"
)

// DelayedTimingKind identifies a delayed resolving instruction suffix.
type DelayedTimingKind string

// Delayed timings recognized by resolving-effect grammar.
const (
	DelayedTimingNone        DelayedTimingKind = ""
	DelayedTimingNextEndStep DelayedTimingKind = "DelayedTimingNextEndStep"
	DelayedTimingNextUpkeep  DelayedTimingKind = "DelayedTimingNextUpkeep"
	DelayedTimingNextMain    DelayedTimingKind = "DelayedTimingNextMain"
)

// EffectDestinationPosition identifies an ordered position in a destination
// zone.
type EffectDestinationPosition string

// Ordered destination positions recognized by resolving-effect grammar.
const (
	EffectDestinationUnspecified EffectDestinationPosition = ""
	EffectDestinationTop         EffectDestinationPosition = "EffectDestinationTop"
	EffectDestinationBottom      EffectDestinationPosition = "EffectDestinationBottom"
)

// SearchControlRider identifies the player a search-and-put clause names as the
// new controller of the found permanent ("put it onto the battlefield ... under
// target player's control"). The zero value denotes no rider: the found card
// enters under the searching player's control.
type SearchControlRider string

// Recognized search-and-put controller riders.
const (
	SearchControlRiderNone           SearchControlRider = ""
	SearchControlRiderTargetPlayer   SearchControlRider = "SearchControlRiderTargetPlayer"
	SearchControlRiderTargetOpponent SearchControlRider = "SearchControlRiderTargetOpponent"
)

// EffectDynamicAmountKind identifies a rules-derived amount.
type EffectDynamicAmountKind string

// Dynamic resolving amounts recognized by the parser.
const (
	EffectDynamicAmountNone           EffectDynamicAmountKind = ""
	EffectDynamicAmountCount          EffectDynamicAmountKind = "EffectDynamicAmountCount"
	EffectDynamicAmountControllerLife EffectDynamicAmountKind = "EffectDynamicAmountControllerLife"
	// EffectDynamicAmountControllerSpeed is the resolving ability controller's
	// current speed ("your speed"), the Start your engines! subsystem amount
	// (CR 702.179). It is controller-scoped: "your" names the resolving
	// ability's controller, so the subject attaches no in-text referent. It
	// backs "where X is your speed" amounts such as The Speed Demon.
	EffectDynamicAmountControllerSpeed EffectDynamicAmountKind = "EffectDynamicAmountControllerSpeed"
	EffectDynamicAmountOpponentCount   EffectDynamicAmountKind = "EffectDynamicAmountOpponentCount"
	EffectDynamicAmountSourcePower     EffectDynamicAmountKind = "EffectDynamicAmountSourcePower"
	// EffectDynamicAmountSourceToughness is a referenced object's toughness
	// ("its toughness"), the toughness sibling of EffectDynamicAmountSourcePower.
	// It backs "gain/lose life equal to its toughness" riders whose subject is a
	// permanent named by an earlier clause.
	EffectDynamicAmountSourceToughness EffectDynamicAmountKind = "EffectDynamicAmountSourceToughness"
	// EffectDynamicAmountSourceManaValue is a referenced object's mana value
	// ("its mana value", "that permanent's mana value"). It backs "gain/lose
	// life equal to its mana value" riders whose subject is the permanent an
	// earlier clause destroyed, as in Feed the Swarm and Divine Offering.
	EffectDynamicAmountSourceManaValue EffectDynamicAmountKind = "EffectDynamicAmountSourceManaValue"
	// EffectDynamicAmountSourceCounterCount is the number of counters of one
	// recognized kind on a referenced object ("burden counter on The One Ring").
	EffectDynamicAmountSourceCounterCount EffectDynamicAmountKind = "EffectDynamicAmountSourceCounterCount"
	EffectDynamicAmountBasicLandTypes     EffectDynamicAmountKind = "EffectDynamicAmountBasicLandTypes"
	EffectDynamicAmountEventCardCount     EffectDynamicAmountKind = "EffectDynamicAmountEventCardCount"
	// EffectDynamicAmountLifeLostThisWay is the total life lost by the players
	// affected by an earlier life-loss effect in the same ability ("equal to the
	// life lost this way"). It scales a follow-on life gain such as the
	// "Each opponent loses N life. You gain life equal to the life lost this
	// way." drain pattern, reading the amount published by the preceding
	// life-loss instruction.
	EffectDynamicAmountLifeLostThisWay EffectDynamicAmountKind = "EffectDynamicAmountLifeLostThisWay"
	// EffectDynamicAmountGreatestPower is the greatest power among a battlefield
	// group ("the greatest power among <group>"). The group is carried in the
	// amount's Selection. EffectDynamicAmountGreatestToughness and
	// EffectDynamicAmountGreatestManaValue are the toughness and mana-value
	// siblings.
	EffectDynamicAmountGreatestPower     EffectDynamicAmountKind = "EffectDynamicAmountGreatestPower"
	EffectDynamicAmountGreatestToughness EffectDynamicAmountKind = "EffectDynamicAmountGreatestToughness"
	EffectDynamicAmountGreatestManaValue EffectDynamicAmountKind = "EffectDynamicAmountGreatestManaValue"
	// EffectDynamicAmountDevotion is the controller's devotion to one or two
	// colors ("your devotion to <color>", "your devotion to <color> and
	// <color>"), the number of mana symbols of those colors among the mana
	// costs of permanents the controller controls (CR 700.5). The colors are
	// carried in the amount's Colors. It backs "X is your devotion to <color>"
	// amounts such as Gray Merchant of Asphodel.
	EffectDynamicAmountDevotion EffectDynamicAmountKind = "EffectDynamicAmountDevotion"
	// EffectDynamicAmountGreatestDiscardedThisWay is the greatest number of
	// cards discarded by any one player during a preceding discard effect in the
	// same ability ("the greatest number of cards a player discarded this way").
	// It backs the Windfall family "Each player discards their hand, then draws
	// cards equal to the greatest number of cards a player discarded this way.",
	// reading the maximum per-player discard count published by the preceding
	// discard instruction.
	EffectDynamicAmountGreatestDiscardedThisWay EffectDynamicAmountKind = "EffectDynamicAmountGreatestDiscardedThisWay"
	// EffectDynamicAmountSpellsCastThisTurn is the number of spells the
	// controller has cast this turn ("for each spell you've cast this turn",
	// "equal to the number of spells you've cast this turn"). It backs the
	// storm-counter family such as Aetherflux Reservoir's "you gain 1 life for
	// each spell you've cast this turn." The triggering spell counts, since its
	// cast event precedes the resolving ability.
	EffectDynamicAmountSpellsCastThisTurn EffectDynamicAmountKind = "EffectDynamicAmountSpellsCastThisTurn"
	// EffectDynamicAmountTriggeringLifeChange is the amount of life gained or
	// lost by the event that triggered the enclosing life-change trigger ("that
	// much life" in "Whenever you gain life, target opponent loses that much
	// life."). It backs the life-drain mirror family (Sanguine Bond, Vito,
	// Exquisite Blood, Marauding Blight-Priest), reading the triggering event's
	// life quantity.
	EffectDynamicAmountTriggeringLifeChange EffectDynamicAmountKind = "EffectDynamicAmountTriggeringLifeChange"
	// EffectDynamicAmountTotalPower is the sum of power across a battlefield
	// group ("the total power of <group>"). The group is carried in the
	// amount's Selection. It backs "where X is the total power of creatures you
	// control" cost reductions (Ghalta, Primal Hunger) and the matching draw and
	// damage amounts. EffectDynamicAmountTotalToughness is the toughness sibling.
	EffectDynamicAmountTotalPower     EffectDynamicAmountKind = "EffectDynamicAmountTotalPower"
	EffectDynamicAmountTotalToughness EffectDynamicAmountKind = "EffectDynamicAmountTotalToughness"
	// EffectDynamicAmountColorCount is the number of distinct colors among a
	// battlefield group ("the number of colors among <group>", "color among
	// <group>"). The group is carried in the amount's Selection. It backs the
	// "+1/+1 for each color among permanents you control" self-buff family
	// (Faeburrow Elder).
	EffectDynamicAmountColorCount EffectDynamicAmountKind = "EffectDynamicAmountColorCount"
	// EffectDynamicAmountSacrificedPower is the power of the permanent
	// sacrificed to pay an activated ability's cost ("the sacrificed creature's
	// power"). Unlike EffectDynamicAmountSourcePower it has no in-text referent
	// span: the subject is the cost-sacrificed permanent, read at resolution
	// from last-known information. EffectDynamicAmountSacrificedToughness and
	// EffectDynamicAmountSacrificedManaValue are the toughness and mana-value
	// siblings. They back Altar of Dementia, whose sacrifice-cost ability mills
	// cards equal to the sacrificed creature's power.
	EffectDynamicAmountSacrificedPower     EffectDynamicAmountKind = "EffectDynamicAmountSacrificedPower"
	EffectDynamicAmountSacrificedToughness EffectDynamicAmountKind = "EffectDynamicAmountSacrificedToughness"
	EffectDynamicAmountSacrificedManaValue EffectDynamicAmountKind = "EffectDynamicAmountSacrificedManaValue"
	// EffectDynamicAmountSharedCreatureTypeCount is the number of other creatures
	// in a battlefield group that share at least one creature type with the
	// affected permanent ("for each other creature on the battlefield that shares
	// a creature type with it"). The group is carried in the amount's Selection.
	// It backs the shared-creature-type anthem family (Coat of Arms), a per-
	// affected-creature dynamic power/toughness bonus.
	EffectDynamicAmountSharedCreatureTypeCount EffectDynamicAmountKind = "EffectDynamicAmountSharedCreatureTypeCount"
	// EffectDynamicAmountTriggeringCombatDamage is the amount of combat damage
	// dealt by the event that triggered the enclosing combat-damage trigger
	// ("that many" in "Whenever a creature you control deals combat damage to a
	// player, create that many Treasure tokens."). It backs the "create that
	// many <predefined> tokens" family (Old Gnawbone), reading the triggering
	// event's damage quantity. Added last so existing kinds keep their values.
	EffectDynamicAmountTriggeringCombatDamage EffectDynamicAmountKind = "EffectDynamicAmountTriggeringCombatDamage"
	// EffectDynamicAmountDestroyedThisWay is the number of permanents destroyed
	// by the immediately preceding destroy effect in the same ability ("for each
	// permanent destroyed this way", "for each creature destroyed this way"). It
	// backs the mass-destroy payoff family (Fumigate, Multani's Decree, Death
	// Begets Life) whose gain-life or draw amount scales with how many permanents
	// the prior clause destroyed; the noun is descriptive of what was destroyed,
	// so the amount carries no selection and the lowerer reads the count the
	// destroy effect publishes. Added last so existing kinds keep their values.
	EffectDynamicAmountDestroyedThisWay EffectDynamicAmountKind = "EffectDynamicAmountDestroyedThisWay"
	// EffectDynamicAmountLifeLostThisTurn is the total life the controller has
	// lost so far this turn ("equal to the life you've lost this turn", "the
	// amount of life you lost this turn"). Damage to the controller counts,
	// because dealing damage to a player causes that player to lose that much
	// life (CR 120.3). It backs Children of Korlis. EffectDynamicAmount
	// LifeGainedThisTurn is the life-gained sibling ("the life you gained this
	// turn"). Both are controller-scoped: the "you" names the resolving
	// ability's controller, so they attach no in-text referent. Added last so
	// existing kinds keep their values.
	EffectDynamicAmountLifeLostThisTurn   EffectDynamicAmountKind = "EffectDynamicAmountLifeLostThisTurn"
	EffectDynamicAmountLifeGainedThisTurn EffectDynamicAmountKind = "EffectDynamicAmountLifeGainedThisTurn"
	// EffectDynamicAmountMaxOf is the greatest value among Operands, the
	// "whichever is greater" combinator over two rules-derived amounts ("equal
	// to the amount of life you gained this turn or the amount of life you lost
	// this turn, whichever is greater." — Willowdusk, Essence Seer). Each operand
	// is itself an EffectAmountSyntax. Added last so existing kinds keep their
	// values.
	EffectDynamicAmountMaxOf EffectDynamicAmountKind = "EffectDynamicAmountMaxOf"
	// EffectDynamicAmountTriggeringCounterCount is the number of counters added
	// by the event that triggered the enclosing counter-placement trigger ("that
	// many" in "Whenever you put one or more +1/+1 counters on a creature you
	// control, you may draw that many cards." — Terrasymbiosis). It reads the
	// triggering event's counter count; the lowerer accepts it only inside a
	// counter-placement trigger and fails closed elsewhere. Added last so
	// existing kinds keep their values.
	EffectDynamicAmountTriggeringCounterCount EffectDynamicAmountKind = "EffectDynamicAmountTriggeringCounterCount"
	// EffectDynamicAmountColorsOfManaSpent is the number of distinct colors of
	// mana spent to cast the source spell ("for each color of mana spent to cast
	// it" — the Converge count). It backs the Converge enters-with-counters
	// quantity (Crystalline Crawler) and is realized by the runtime, which
	// records the colors of mana spent as the spell's costs are paid and carries
	// the count to the entering permanent. Added last so existing kinds keep
	// their values.
	EffectDynamicAmountColorsOfManaSpent EffectDynamicAmountKind = "EffectDynamicAmountColorsOfManaSpent"
	// EffectDynamicAmountDieRollResult is the value produced by the immediately
	// preceding die roll in the same ability ("a number of Treasure tokens equal
	// to the result." — Ancient Copper Dragon and the Ancient Dragon dice cycle).
	// It carries no selection; the lowerer reads the count the EffectRollDie
	// effect publishes. Added last so existing kinds keep their values.
	EffectDynamicAmountDieRollResult EffectDynamicAmountKind = "EffectDynamicAmountDieRollResult"
	// EffectDynamicAmountTotalManaValue is the sum of mana value across a
	// battlefield group ("the total mana value of <group>"). The group is carried
	// in the amount's Selection. It backs "where X is the total mana value of
	// noncreature artifacts you control" cost reductions (Metalwork Colossus,
	// Earthquake Dragon, Excalibur, Sword of Eden). Added last so existing kinds
	// keep their values.
	EffectDynamicAmountTotalManaValue EffectDynamicAmountKind = "EffectDynamicAmountTotalManaValue"
	// EffectDynamicAmountTimesKicked is the number of times the spell was kicked
	// (its Multikicker count, CR 702.32), the "for each time it was kicked" amount
	// (Everflowing Chalice's enters-with-counters quantity, Wolfbriar Elemental's
	// Wolf-token count). It carries no selection; the runtime records the kick
	// count as the spell is cast. Added last so existing kinds keep their values.
	EffectDynamicAmountTimesKicked EffectDynamicAmountKind = "EffectDynamicAmountTimesKicked"
	// EffectDynamicAmountOpponentsAttackedThisCombat is the number of the
	// controller's opponents that creatures the controller controls are
	// attacking this combat ("for each opponent you attacked this combat" —
	// the Melee count, CR 702.72). Distinct opponents are counted once each,
	// read from the current combat's attack declarations as the ability
	// resolves. It is the controller-scoped, combat-state sibling of
	// EffectDynamicAmountOpponentCount and carries no in-text referent. Added
	// last so existing kinds keep their values.
	EffectDynamicAmountOpponentsAttackedThisCombat EffectDynamicAmountKind = "EffectDynamicAmountOpponentsAttackedThisCombat"
)

// EffectDynamicAmountForm identifies how a dynamic amount is introduced.
type EffectDynamicAmountForm string

// Dynamic amount forms recognized by the parser.
const (
	EffectDynamicAmountFormNone    EffectDynamicAmountForm = ""
	EffectDynamicAmountFormEqual   EffectDynamicAmountForm = "EffectDynamicAmountFormEqual"
	EffectDynamicAmountFormForEach EffectDynamicAmountForm = "EffectDynamicAmountFormForEach"
	EffectDynamicAmountFormWhereX  EffectDynamicAmountForm = "EffectDynamicAmountFormWhereX"
)

// EffectAmountSyntax is a fixed or rules-derived source-spanned amount.
type EffectAmountSyntax struct {
	Span       shared.Span `json:"-"`
	Text       string      `json:",omitempty"`
	Value      int         `json:",omitempty"`
	Known      bool        `json:",omitempty"`
	RangeKnown bool        `json:",omitempty"`
	Minimum    int         `json:",omitempty"`
	Maximum    int         `json:",omitempty"`
	VariableX  bool        `json:",omitempty"`
	// AnyNumber records the unbounded "any number of <noun>" count form, where
	// the resolving player may choose any quantity from none up to all eligible
	// objects. It is a positive marker set only by the literal "any number of"
	// wording, distinguishing that form from the otherwise identical empty amount
	// produced by "all", "the", or a bare plural noun.
	AnyNumber   bool                    `json:",omitempty"`
	DynamicKind EffectDynamicAmountKind `json:",omitempty"`
	DynamicForm EffectDynamicAmountForm `json:",omitempty"`
	Multiplier  int                     `json:",omitempty"`
	// Addend is a fixed integer added to a dynamic count after the multiplier,
	// modeling the "plus N" rider on a counted amount ("the number of cards in
	// your hand plus one.", Sea Gate Restoration). It is zero when no such rider
	// is present.
	Addend        int              `json:",omitempty"`
	ReferenceSpan shared.Span      `json:"-"`
	CounterKind   counter.Kind     `json:",omitempty"`
	Selection     *SelectionSyntax `json:",omitempty"`
	// Colors carries the colors of a devotion amount ("your devotion to
	// <color(s)>"). It is empty for every other amount kind.
	Colors []Color `json:",omitempty"`
	// Operands carries the sub-amounts of a EffectDynamicAmountMaxOf combinator
	// ("<A> or <B>, whichever is greater"). It is empty for every other amount
	// kind.
	Operands []EffectAmountSyntax `json:",omitempty"`
}

// EffectReplacementKind identifies how an instruction replaces an event.
type EffectReplacementKind string

// Resolving replacement modifiers recognized by the parser.
const (
	EffectReplacementNone           EffectReplacementKind = ""
	EffectReplacementInstead        EffectReplacementKind = "EffectReplacementInstead"
	EffectReplacementTwiceThatMany  EffectReplacementKind = "EffectReplacementTwiceThatMany"
	EffectReplacementThatMuchPlus   EffectReplacementKind = "EffectReplacementThatMuchPlus"
	EffectReplacementDoubleThat     EffectReplacementKind = "EffectReplacementDoubleThat"
	EffectReplacementThatManyPlus   EffectReplacementKind = "EffectReplacementThatManyPlus"
	EffectReplacementOneOfEach      EffectReplacementKind = "EffectReplacementOneOfEach"
	EffectReplacementTripleThat     EffectReplacementKind = "EffectReplacementTripleThat"
	EffectReplacementTwiceThatMuch  EffectReplacementKind = "EffectReplacementTwiceThatMuch"
	EffectReplacementPlusAdditional EffectReplacementKind = "EffectReplacementPlusAdditional"
)

// EffectReplacementSyntax is a source-spanned replacement modifier.
type EffectReplacementSyntax struct {
	Kind            EffectReplacementKind `json:",omitempty"`
	Span            shared.Span           `json:"-"`
	Amount          int                   `json:",omitempty"`
	EachCounterKind bool                  `json:",omitempty"`
}

// EffectManaSyntax describes exact add-mana output.
type EffectManaSyntax struct {
	Span    shared.Span `json:"-"`
	Symbols []string    `json:",omitempty"`
	// Colors are the typed mana colors recognized from Symbols, in order, when
	// every symbol is a basic color token ({W}{U}{B}{R}{G}{C}). They let a
	// consumer build add-mana content from typed values instead of re-parsing the
	// rendered symbol strings. Colors is populated only when ColorsKnown is true.
	Colors      []mana.Color `json:"-"`
	ColorsKnown bool         `json:",omitempty"`
	Choice      bool         `json:",omitempty"`
	AnyColor    bool         `json:",omitempty"`
	// ChosenColor reports the exact body "one mana of the chosen color", which
	// adds one mana of the color chosen as the source permanent entered (CR
	// 614.12) rather than a fixed or freely-chosen color.
	ChosenColor bool `json:",omitempty"`
	// ChosenColorFixed is the fixed alternative basic color of the composite body
	// "{C} or one mana of the chosen color." (the Gate/Thriving land cycle, e.g.
	// "Add {W} or one mana of the chosen color."). It is set together with
	// ChosenColor and ChosenColorFixedKnown; it is empty for the plain chosen
	// color body.
	ChosenColorFixed      mana.Color `json:"-"`
	ChosenColorFixedKnown bool       `json:",omitempty"`
	// ChosenColorDevotion reports the exact body "an amount of mana of that color
	// equal to your devotion to that color." (Nykthos, Shrine to Nyx). The
	// controller chooses a color as the ability resolves; the produced mana is
	// that color and its amount is the controller's devotion to that chosen color
	// (CR 700.5).
	ChosenColorDevotion bool `json:",omitempty"`
	// ChosenColorDynamic reports the body "an amount of mana of that color equal
	// to <dynamic amount>" whose quantity is a battlefield count carried by
	// EffectSyntax.Amount (Three Tree City: "...equal to the number of creatures
	// you control of the chosen type."). The controller chooses a color as the
	// ability resolves; the produced mana is that color and its amount is the
	// dynamic count. It pairs the chosen-color output with a dynamic amount the
	// fixed-shape ChosenColorDevotion body cannot express.
	ChosenColorDynamic bool `json:",omitempty"`
	// CommanderIdentity reports the exact body "one mana of any color in your
	// commander's color identity" (CR 903.4). The choosable colors are the
	// controller's commander color identity, resolved dynamically at activation.
	CommanderIdentity bool `json:",omitempty"`
	// DynamicColorless reports the exact body "an amount of {C} equal to ..."
	// whose quantity is carried by EffectSyntax.Amount.
	DynamicColorless bool `json:",omitempty"`
	LegacyBodyExact  bool `json:",omitempty"`
	// FilterPair reports the "filter land" output body
	// "{X}{X}, {X}{Y}, or {Y}{Y}.": the choice among the three two-mana
	// combinations of a fixed two-color pair (the filter-land cycle, e.g. Mystic
	// Gate's "Add {W}{W}, {W}{U}, or {U}{U}."). The pair's two distinct basic
	// colors are recorded in FilterColors as {X, Y}; the produced output is two
	// mana, each independently one of those two colors.
	FilterPair   bool         `json:",omitempty"`
	FilterColors []mana.Color `json:"-"`
	// LandsProduce reports the body "one mana of any color that a land <scope>
	// could produce" (Exotic Orchard, Reflecting Pool, Fellwar Stone). The
	// choosable colors are recomputed at resolution from the colors every
	// battlefield land matching LandsProduceScope could produce. LandsProduceScope
	// records which lands those are.
	LandsProduce      bool                  `json:",omitempty"`
	LandsProduceScope ManaLandsProduceScope `json:",omitempty"`
	// LandsProduceAnyType reports the "any type" wording (Reflecting Pool, Naga
	// Vitalist) rather than "any color" (Exotic Orchard, Harvester Druid). "Any
	// type" additionally offers colorless ({C}) when a matching land could
	// produce it; "any color" offers only colored mana. It is set together with
	// LandsProduce.
	LandsProduceAnyType bool `json:",omitempty"`
	// LinkedExileColors reports the body "one mana of any of the exiled card's
	// colors" (Chrome Mox). The choosable colors are recomputed at resolution
	// from the colors of the card the source permanent imprinted as it entered
	// (the card it exiled from hand); an absent or colorless imprint offers none.
	LinkedExileColors bool `json:",omitempty"`
	// ColorsAmongControlled reports the body "one mana of any color among
	// <permanents> you control" (Mox Amber's "legendary creatures and
	// planeswalkers you control", Plaza of Heroes' "legendary permanents you
	// control"). The choosable colors are recomputed at resolution as the union
	// of colors of the battlefield permanents the controller controls matching
	// ColorsAmongSelection.
	ColorsAmongControlled bool `json:",omitempty"`
	// ColorsAmongSelection carries the permanent filter of a ColorsAmongControlled
	// body. It is set together with ColorsAmongControlled.
	ColorsAmongSelection *SelectionSyntax `json:",omitempty"`
	// EachColorAmongControlled reports the body "For each color among
	// <permanents> you control, add one mana of that color" (Bloom Tender). It
	// produces one mana of EACH distinct color found among the permanents the
	// controller controls matching ColorsAmongSelection, recomputed at
	// resolution. Unlike ColorsAmongControlled, no color is chosen; one mana of
	// every color in the union is added.
	EachColorAmongControlled bool `json:",omitempty"`
	// AnyOneColorDynamic reports the body "X mana of any one color" (or "an
	// amount of mana of any one color") whose quantity is a dynamic amount
	// carried by EffectSyntax.Amount (Kami of Whispered Hopes: "Add X mana of
	// any one color, where X is this creature's power."). The controller chooses
	// any one color as the ability resolves; the produced mana is that color and
	// its amount is the dynamic value. It pairs a freely-chosen single color with
	// a generic dynamic amount (source power/toughness, devotion, a permanent
	// count, and so on).
	AnyOneColorDynamic bool `json:",omitempty"`
	// AnyColorCount reports the body "<N> mana of any one color" (Gilded Lotus:
	// "Add three mana of any one color."), N >= 2. The controller chooses a
	// single color as the ability resolves and adds that many mana of the one
	// chosen color. It is set together with AnyColor; the plain "one mana of any
	// color" body leaves it zero (one mana of the chosen color).
	AnyColorCount int `json:",omitempty"`
	// Instead reports a trailing "instead" on the add-mana body ("Add
	// {B}{B}{B}{B}{B} instead if there are seven or more cards in your
	// graveyard.", the Threshold conditional-mana cycle). It marks this mana
	// production as the conditional alternative to a sibling base production
	// rather than an additional output; lowering pairs the two into one ability
	// whose larger output replaces the base when the condition holds.
	Instead bool `json:",omitempty"`
	// TriggerLandProducedType reports the body "one mana of any type that land
	// produced" (Mirari's Wake, Zendikar Resurgent, Dictate of Karametra, Mana
	// Flare, Heartbeat of Spring). It is the mana-doubler output of a
	// tapped-for-mana trigger: one additional mana whose type is chosen among the
	// types the triggering land produced on that tap, recomputed at resolution
	// from the triggering tap rather than from a fixed color or the battlefield.
	TriggerLandProducedType bool `json:",omitempty"`
}

// ManaLandsProduceScope identifies which battlefield lands' producible colors
// feed a "mana of any color that a land ... could produce" mana ability.
type ManaLandsProduceScope int

// Mana lands-produce scope values name the relevant lands by controller.
const (
	// ManaLandsProduceNone marks a non-lands-produce mana body.
	ManaLandsProduceNone ManaLandsProduceScope = iota
	// ManaLandsProduceYou scopes to lands the controller controls, as in
	// Reflecting Pool's "a land you control could produce" body.
	ManaLandsProduceYou
	// ManaLandsProduceOpponent scopes to lands an opponent controls, as in the
	// Exotic Orchard and Fellwar Stone "a land an opponent controls could
	// produce" body.
	ManaLandsProduceOpponent
)

// EffectContextKind identifies the grammatical subject performing or receiving
// a resolving instruction.
type EffectContextKind string

// Resolving-effect contexts recognized by the parser.
const (
	EffectContextUnknown          EffectContextKind = ""
	EffectContextController       EffectContextKind = "EffectContextController"
	EffectContextTarget           EffectContextKind = "EffectContextTarget"
	EffectContextEachOpponent     EffectContextKind = "EffectContextEachOpponent"
	EffectContextEachPlayer       EffectContextKind = "EffectContextEachPlayer"
	EffectContextEventPlayer      EffectContextKind = "EffectContextEventPlayer"
	EffectContextSource           EffectContextKind = "EffectContextSource"
	EffectContextReferencedObject EffectContextKind = "EffectContextReferencedObject"
	EffectContextReferencedPlayer EffectContextKind = "EffectContextReferencedPlayer"
	// EffectContextReferencedObjectController marks an effect whose subject is the
	// controller of a referenced object ("Its controller creates …", "That
	// creature's controller creates …"). The recipient is the controller of the
	// object the subject reference resolves to.
	EffectContextReferencedObjectController EffectContextKind = "EffectContextReferencedObjectController"
	EffectContextPriorSubject               EffectContextKind = "EffectContextPriorSubject"
	// EffectContextControllerAndTarget marks an effect distributed to both the
	// controller and a single player target ("You and target opponent each draw
	// a card"). The target player is the effect's sole target; the controller is
	// the implicit co-recipient.
	EffectContextControllerAndTarget EffectContextKind = "EffectContextControllerAndTarget"
	// EffectContextEachOtherPlayer marks an effect whose subject is "each other
	// player" — every player except the controller. This denotes the same set as
	// "each opponent" (a player has no teammates in the supported formats), so it
	// resolves identically to OpponentsReference, but is kept distinct so the
	// parser reconstructs the original "Each other player" wording byte-for-byte.
	EffectContextEachOtherPlayer EffectContextKind = "EffectContextEachOtherPlayer"
	// EffectContextDefendingPlayer marks an effect whose subject is "defending
	// player" — the player being attacked in a combat trigger ("Whenever this
	// creature attacks, defending player sacrifices N permanents."). It resolves
	// to the attacking creature's defending player, identified by the triggering
	// attack event.
	EffectContextDefendingPlayer EffectContextKind = "EffectContextDefendingPlayer"
)

// DamageRecipientReferenceKind identifies a damage recipient that is the
// controller or owner of a referenced object (the prior removal target), as in
// "deals 2 damage to that land's controller" or "deals 2 damage to its owner",
// or the source's own controller ("deals 2 damage to you").
// It is None for every other recipient (a target, a group, or any target).
type DamageRecipientReferenceKind uint8

// Damage recipient reference kinds.
const (
	DamageRecipientReferenceNone DamageRecipientReferenceKind = iota
	DamageRecipientReferenceController
	DamageRecipientReferenceOwner
	// DamageRecipientReferenceYou marks the source's own controller as the
	// damage recipient, the literal "you" recipient of "deals N damage to you".
	DamageRecipientReferenceYou
	// DamageRecipientReferenceThatPlayer marks the triggering event's player as
	// the damage recipient, the "that player" recipient of "Whenever an opponent
	// draws a card, ~ deals N damage to that player." (Underworld Dreams,
	// Megrim, Manabarbs). The "that player" reference binds to the event player
	// (ReferenceBindingEventPlayer); lowering resolves it to EventPlayerReference.
	DamageRecipientReferenceThatPlayer
	// DamageRecipientReferenceItself marks every member of an "each <group>"
	// subject as dealing damage to itself, the per-member self recipient of
	// "Each creature deals damage to itself equal to its power." (Wave of
	// Reckoning). It is recorded only as EachSourceDamageRecipient, paired with
	// the subject group in EachSourceDamageGroup; lowering deals each member its
	// own power.
	DamageRecipientReferenceItself
)

// SignedAmountSyntax is one signed half of a power/toughness change.
type SignedAmountSyntax struct {
	Span     shared.Span `json:"-"`
	Value    int         `json:",omitempty"`
	Known    bool        `json:",omitempty"`
	Negative bool        `json:",omitempty"`
	// VariableX marks a side written as the variable "X" (as in "+X/+0"), whose
	// magnitude is supplied by the effect's dynamic amount rather than a fixed
	// Value. Known stays false for an X side.
	VariableX bool `json:",omitempty"`
}

// GraveyardZoneExileKind identifies a recognized whole-graveyard exile and whose
// graveyard it targets, distinguishing it from single-card graveyard exile and
// the unmodeled multi-graveyard forms.
type GraveyardZoneExileKind string

// Whole-graveyard exile owner relations.
const (
	GraveyardZoneExileNone GraveyardZoneExileKind = ""
	// GraveyardZoneExileTargetPlayer is "Exile target player's graveyard." — a
	// player-targeted zone wipe of any one player's graveyard.
	GraveyardZoneExileTargetPlayer GraveyardZoneExileKind = "GraveyardZoneExileTargetPlayer"
	// GraveyardZoneExileTargetOpponent is "Exile target opponent's graveyard." —
	// the same wipe restricted to an opponent's graveyard.
	GraveyardZoneExileTargetOpponent GraveyardZoneExileKind = "GraveyardZoneExileTargetOpponent"
	// GraveyardZoneExileAll is "Exile all graveyards." (and the synonymous "Exile
	// each player's graveyard.") — a non-targeted wipe of every player's
	// graveyard at once.
	GraveyardZoneExileAll GraveyardZoneExileKind = "GraveyardZoneExileAll"
)

// AttackDefenderKind identifies the defender a created attacking token is
// told to attack in the trailing "... attacking <defender>" clause (CR 508.4),
// e.g. "create a 1/1 ... token that's tapped and attacking that player or a
// planeswalker they control." (Adeline, Resplendent Cathar). The runtime
// EntryAttacking model is defender-agnostic — it puts the token into combat
// against a legal defender — so the kind records only the recognized Oracle
// wording for byte-exact clause reconstruction; it does not change lowering.
type AttackDefenderKind string

// Created-attacking-token defender relations.
const (
	AttackDefenderNone AttackDefenderKind = ""
	// AttackDefenderThatPlayerOrPlaneswalker is "... attacking that player or
	// a planeswalker they control." — the modern per-opponent templating.
	AttackDefenderThatPlayerOrPlaneswalker AttackDefenderKind = "AttackDefenderThatPlayerOrPlaneswalker"
	// AttackDefenderThatPlayer is "... attacking that player."
	AttackDefenderThatPlayer AttackDefenderKind = "AttackDefenderThatPlayer"
	// AttackDefenderThatOpponent is "... attacking that opponent."
	AttackDefenderThatOpponent AttackDefenderKind = "AttackDefenderThatOpponent"
)

// SelectionController identifies a selected object's controller.
type SelectionController string

// Selection controller relations.
const (
	SelectionControllerAny      SelectionController = ""
	SelectionControllerYou      SelectionController = "SelectionControllerYou"
	SelectionControllerOpponent SelectionController = "SelectionControllerOpponent"
	SelectionControllerNotYou   SelectionController = "SelectionControllerNotYou"
)

// SelectionKind identifies the broad object selected by a phrase.
type SelectionKind string

// Selection kinds recognized by resolving-effect grammar.
const (
	SelectionUnknown                          SelectionKind = ""
	SelectionAny                              SelectionKind = "SelectionAny"
	SelectionPlayer                           SelectionKind = "SelectionPlayer"
	SelectionOpponent                         SelectionKind = "SelectionOpponent"
	SelectionArtifact                         SelectionKind = "SelectionArtifact"
	SelectionCreature                         SelectionKind = "SelectionCreature"
	SelectionEnchantment                      SelectionKind = "SelectionEnchantment"
	SelectionLand                             SelectionKind = "SelectionLand"
	SelectionPermanent                        SelectionKind = "SelectionPermanent"
	SelectionCard                             SelectionKind = "SelectionCard"
	SelectionSpell                            SelectionKind = "SelectionSpell"
	SelectionActivatedAbility                 SelectionKind = "SelectionActivatedAbility"
	SelectionTriggeredAbility                 SelectionKind = "SelectionTriggeredAbility"
	SelectionActivatedOrTriggeredAbility      SelectionKind = "SelectionActivatedOrTriggeredAbility"
	SelectionSpellActivatedOrTriggeredAbility SelectionKind = "SelectionSpellActivatedOrTriggeredAbility"
	SelectionTriggeredAbilityOrSpell          SelectionKind = "SelectionTriggeredAbilityOrSpell"
	SelectionPlaneswalker                     SelectionKind = "SelectionPlaneswalker"
	SelectionBattle                           SelectionKind = "SelectionBattle"
	SelectionCommander                        SelectionKind = "SelectionCommander"
)

// SelectionSyntax is a typed, source-spanned noun phrase.
type SelectionSyntax struct {
	Span       shared.Span         `json:"-"`
	Text       string              `json:",omitempty"`
	Kind       SelectionKind       `json:",omitempty"`
	Controller SelectionController `json:",omitempty"`
	// OpponentEach records that an opponent-controlled selection used the
	// distributive "each opponent controls" wording rather than the plural "your
	// opponents control". It is meaningful only when Controller is
	// SelectionControllerOpponent and exists solely so the byte-exact recipient
	// reconstruction can rebuild the verbatim phrasing; both wordings lower to
	// the same opponent controller.
	OpponentEach bool `json:",omitempty"`
	All          bool `json:",omitempty"`
	Another      bool `json:",omitempty"`
	Other        bool `json:",omitempty"`
	Attacking    bool `json:",omitempty"`
	Blocking     bool `json:",omitempty"`
	Tapped       bool `json:",omitempty"`
	Untapped     bool `json:",omitempty"`
	// NonToken records a "nontoken" selector qualifier ("nontoken creature");
	// TokenOnly records a "token" qualifier ("token creature"). They are mutually
	// exclusive and lower to Selection.NonToken / Selection.TokenOnly.
	NonToken      bool `json:",omitempty"`
	TokenOnly     bool `json:",omitempty"`
	Colorless     bool `json:",omitempty"`
	Multicolored  bool `json:",omitempty"`
	BasicLandType bool `json:",omitempty"`
	// Historic records a "historic" card qualifier ("target historic card from
	// your graveyard"). A historic card is an artifact, a legendary, or a Saga
	// (CR 702.61b), a cross-category disjunction no single type/supertype/subtype
	// field can express, so it is kept as its own flag and lowers to a
	// Selection.AnyOf of those three alternatives.
	Historic bool `json:",omitempty"`
	// ConjunctiveTypes records that a multi-member RequiredTypesAny names card
	// types the permanent must carry all at once ("artifact creature") rather
	// than any one of them ("artifact or creature"). It lowers the type set to
	// the conjunctive permanent-type target filter.
	ConjunctiveTypes bool `json:",omitempty"`
	// PlayerOrPlaneswalker marks the combined "player or planeswalker" /
	// "opponent or planeswalker" combined damage target. Kind stays
	// SelectionPlayer or SelectionOpponent for the player half; this flag records
	// the additional planeswalker-permanent half the merged Kind cannot express.
	PlayerOrPlaneswalker bool `json:",omitempty"`
	// MatchManaValue, MatchPower, and MatchToughness record whether their paired
	// ManaValue/Power/Toughness comparison below is active. They are grouped with
	// the other booleans to keep the struct compact.
	MatchManaValue bool `json:",omitempty"`
	MatchPower     bool `json:",omitempty"`
	MatchToughness bool `json:",omitempty"`
	// MatchTotalManaValue records whether the paired TotalManaValue comparison is
	// active. Unlike MatchManaValue, which bounds each matched card's own mana
	// value, this bounds the combined mana value of the whole chosen set ("Return
	// up to two creature cards with total mana value 4 or less ..."). It is kept
	// distinct so a set-sum constraint never lowers to a per-card filter.
	MatchTotalManaValue bool        `json:",omitempty"`
	Keyword             KeywordKind `json:",omitempty"`
	// ExcludedKeyword records a "without <keyword>" selector qualifier (e.g.
	// "each creature without flying"); it is mutually exclusive with Keyword.
	ExcludedKeyword    KeywordKind       `json:",omitempty"`
	Zone               zone.Type         `json:",omitempty"`
	RequiredTypesAny   []CardType        `json:",omitempty"`
	ExcludedTypes      []CardType        `json:",omitempty"`
	SourceTypes        []CardType        `json:",omitempty"`
	Supertypes         []Supertype       `json:",omitempty"`
	ExcludedSupertypes []Supertype       `json:",omitempty"`
	ColorsAny          []Color           `json:",omitempty"`
	ExcludedColors     []Color           `json:",omitempty"`
	SubtypesAny        []types.Sub       `json:",omitempty"`
	ExcludedSubtypes   []types.Sub       `json:",omitempty"`
	Alternatives       []SelectionSyntax `json:",omitempty"`
	// InclusiveOneOfEach records that the selection joined two or more singular
	// articled card nouns with "and/or" ("a Saga card and/or a land card"),
	// meaning the controller may choose up to one card of each named type rather
	// than a single card matching any one of them. The merged RequiredTypesAny /
	// SubtypesAny still carry the named types; this flag tells the lowering to
	// realize one independent optional pick per named type. It is meaningful only
	// for the put-from-among-onto-battlefield shape and is ignored elsewhere.
	InclusiveOneOfEach bool        `json:",omitempty"`
	ManaValue          compare.Int `json:",omitzero"`
	Power              compare.Int `json:",omitzero"`
	Toughness          compare.Int `json:",omitzero"`
	// TotalManaValue holds the set-sum mana value comparison bound when
	// MatchTotalManaValue is set ("with total mana value N or less"). It bounds
	// the combined mana value of the chosen cards, not any individual card.
	TotalManaValue compare.Int `json:",omitzero"`
	// CounterRequired records a "with a <kind> counter on it/them" qualifier;
	// CounterKind names the counter the matched permanent must carry. When the
	// qualifier names no kind ("with a counter on it"), CounterAny is set instead
	// and the match requires a counter of any kind.
	CounterRequired bool         `json:",omitempty"`
	CounterKind     counter.Kind `json:",omitempty"`
	CounterAny      bool         `json:",omitempty"`
	// SubtypeFromEntryChoice records a trailing "of the chosen type" qualifier on
	// a count subject ("the number of creatures you control of the chosen type"),
	// requiring each matched permanent to share the creature subtype the source
	// permanent chose as it entered (Three Tree City). It lowers to the runtime
	// Selection.SubtypeFromSourceEntryChoice predicate.
	SubtypeFromEntryChoice bool `json:",omitempty"`
	// SubtypeFromChosenType records a trailing "of that type" qualifier on a count
	// subject ("each permanent you control of that type"), requiring each matched
	// permanent to share the creature subtype chosen earlier in the same
	// resolution by a "Choose a creature type." effect (Distant Melody). It lowers
	// to the runtime Selection.SubtypeFromChosenType predicate (which reads
	// game.SpellChosenTypeChoiceKey).
	SubtypeFromChosenType bool `json:",omitempty"`
	// SubtypeFromChosenTypeExcluded records a trailing "that aren't of the chosen
	// type" qualifier on a mass group ("Destroy all creatures that aren't of the
	// chosen type." — Kindred Dominance), requiring each matched permanent to NOT
	// share the creature subtype chosen earlier in the same resolution by a "Choose
	// a creature type." effect. It lowers to the runtime
	// game.SubtypeChoiceResolutionExcluded predicate (which reads
	// game.SpellChosenTypeChoiceKey).
	SubtypeFromChosenTypeExcluded bool `json:",omitempty"`
	// ManaValueX records that the MatchManaValue comparison bound is the spell's
	// chosen {X} rather than a fixed number ("with mana value X or less"). When
	// set, ManaValue holds the operator (LessOrEqual) with no fixed Value; the
	// bound is resolved from the spell's X as the effect resolves. It backs the
	// X-bounded library-search tutors (Green Sun's Zenith, Chord of Calling,
	// Wargate).
	ManaValueX bool `json:",omitempty"`
	// ManaValueDynamic records a "with mana value less than or equal to the
	// amount of life you (lost|gained) this turn" bound (Betor, Ancestor's
	// Voice), whose upper bound is a turn-event life total rather than a fixed
	// number. It is independent of MatchManaValue/ManaValue (which model fixed
	// and X bounds): only the dedicated graveyard-card target reconstruction
	// renders it, so every other selection context fails closed. The recognized
	// kinds are EffectDynamicAmountLifeLostThisTurn and
	// EffectDynamicAmountLifeGainedThisTurn; the empty value means no dynamic
	// bound.
	ManaValueDynamic EffectDynamicAmountKind `json:",omitempty"`
	// RequiredName carries the verbatim card name of a "named <Name>" selector
	// qualifier ("a card named Trustworthy Scout"). It is captured from the
	// source tokens after "named" so the byte-exact search reconstruction can
	// rebuild the qualifier; the runtime matches a library card by this name.
	RequiredName string `json:",omitempty"`
	// EnteredThisTurn records a trailing "that entered this turn" relative clause
	// ("each green creature that entered this turn"), restricting the match to
	// permanents that entered the battlefield during the current turn. It lowers
	// to Selection.EnteredThisTurn.
	EnteredThisTurn bool `json:",omitempty"`
	// PowerLessThanSource records a trailing "with lesser power" relative clause
	// ("target attacking creature with lesser power", Mentor), restricting the
	// match to permanents whose power is strictly less than the ability's source
	// permanent's power. PowerGreaterThanSource is the "with greater power"
	// sibling. They are relative to the source, so unlike Power/MatchPower they
	// carry no fixed comparison; they lower to Selection.PowerLessThanSource /
	// Selection.PowerGreaterThanSource.
	PowerLessThanSource    bool `json:",omitempty"`
	PowerGreaterThanSource bool `json:",omitempty"`
}

// TargetCardinalitySyntax is an inclusive target-count range.
type TargetCardinalitySyntax struct {
	Min int `json:",omitempty"`
	Max int `json:",omitempty"`
}

// TargetSyntax is one typed target production.
type TargetSyntax struct {
	Span shared.Span `json:"-"`
	// ChoiceSpan is the exact leading "Choose" verb for target declarations
	// whose selection occurs before resolution.
	ChoiceSpan  shared.Span             `json:"-"`
	Text        string                  `json:",omitempty"`
	Cardinality TargetCardinalitySyntax `json:",omitzero"`
	Selection   SelectionSyntax         `json:",omitzero"`
	Exact       bool                    `json:",omitempty"`
	// Order is the target's dense source-order rank, used downstream to bind
	// references to their closest preceding target without byte offsets.
	Order shared.SourceOrder `json:"-"`
}

// EffectConnectionKind identifies how a resolving instruction is coordinated
// with the preceding instruction in the same sentence.
type EffectConnectionKind string

// Resolving-instruction connections recognized by the parser.
const (
	EffectConnectionNone EffectConnectionKind = ""
	EffectConnectionAnd  EffectConnectionKind = "EffectConnectionAnd"
	EffectConnectionThen EffectConnectionKind = "EffectConnectionThen"
	// EffectConnectionOtherwise marks an effect introduced by a leading
	// "Otherwise," that runs only when the immediately preceding effect's gate
	// condition is false ("draw a card if its power is 3 or greater. Otherwise,
	// put two +1/+1 counters on it."). The lowering gates this effect on the
	// negation of the preceding effect's condition so exactly one branch runs.
	EffectConnectionOtherwise EffectConnectionKind = "EffectConnectionOtherwise"
	// EffectConnectionOr marks a resolving effect joined to the preceding effect
	// by a sentence-level "or" ("Put a +1/+1 counter on target creature or that
	// creature gains trample."). The two effects are alternatives: the
	// controller chooses exactly one to carry out. Lowering realizes the choice
	// as a modal ability whose modes are the alternative effects.
	EffectConnectionOr EffectConnectionKind = "EffectConnectionOr"
)

// EffectPlayerKind identifies the player who performs an effect and whose zone
// it acts on when that player is not the resolving ability's controller.
type EffectPlayerKind string

// Resolving-effect player relations recognized by the parser.
const (
	EffectPlayerNone        EffectPlayerKind = ""
	EffectPlayerTargetOwner EffectPlayerKind = "EffectPlayerTargetOwner"
)

// EffectCardSourceKind identifies a card source consumed by an effect.
type EffectCardSourceKind string

// Resolving-effect card sources recognized by the parser.
const (
	EffectCardSourceNone                   EffectCardSourceKind = ""
	EffectCardSourceTopOfPlayerLibrary     EffectCardSourceKind = "EffectCardSourceTopOfPlayerLibrary"
	EffectCardSourcePriorInstructionResult EffectCardSourceKind = "EffectCardSourcePriorInstructionResult"
)

// EntersAsCopyConditionalCounter is a conditional copiable counter rider on an
// enters-as-copy replacement: Amount counters of Kind are placed on the copy
// only when the copy has IfType among its card types ("it enters with an
// additional +1/+1 counter on it if it's a creature"; Spark Double).
type EntersAsCopyConditionalCounter struct {
	Kind   counter.Kind `json:",omitempty"`
	Amount int          `json:",omitempty"`
	IfType types.Card   `json:",omitempty"`
}

// EffectSyntax is one typed resolving instruction. Text and Tokens remain
// lossless metadata; all meaning consumed downstream is carried by typed fields.
type EffectSyntax struct {
	Kind           EffectKind           `json:",omitempty"`
	Context        EffectContextKind    `json:",omitempty"`
	Connection     EffectConnectionKind `json:",omitempty"`
	ConnectionSpan shared.Span          `json:"-"`
	Span           shared.Span          `json:"-"`
	VerbSpan       shared.Span          `json:"-"`
	ClauseSpan     shared.Span          `json:"-"`
	Text           string               `json:",omitempty"`
	Tokens         []shared.Token       `json:"-"`
	Player         EffectPlayerKind     `json:",omitempty"`
	CardSource     EffectCardSourceKind `json:",omitempty"`
	// RequirePermanentCard gates a linked-card effect on the referenced card
	// being a permanent card.
	RequirePermanentCard bool               `json:",omitempty"`
	Duration             EffectDurationKind `json:",omitempty"`
	// CantCastSpellsAllPlayers reports that an EffectCantCastSpells clause
	// affects every player ("Players can't cast spells this turn.") rather than
	// only the controller's opponents ("Your opponents can't cast spells this
	// turn."). It is meaningful only when Kind is EffectCantCastSpells.
	CantCastSpellsAllPlayers bool `json:",omitempty"`
	// CantCastSpellsRequiredTypes narrows an EffectCantCastSpells clause to spells
	// of the named card type ("Target player can't cast creature spells this
	// turn."); CantCastSpellsExcludedTypes exempts spells of the named card type
	// ("Your opponents can't cast noncreature spells this turn."). Both are
	// meaningful only when Kind is EffectCantCastSpells, hold at most one type,
	// and are mutually exclusive.
	CantCastSpellsRequiredTypes []CardType `json:",omitempty"`
	CantCastSpellsExcludedTypes []CardType `json:",omitempty"`
	// PreventDamageTo and PreventDamageBy mark an EffectPreventDamage clause
	// that prevents all combat damage for the turn to and/or from a single
	// referenced or targeted permanent ("Prevent all combat damage that would
	// be dealt to and dealt by that creature this turn." — Maze of Ith).
	PreventDamageTo bool `json:",omitempty"`
	PreventDamageBy bool `json:",omitempty"`
	// PreventDamageGlobal marks an EffectPreventDamage clause that prevents all
	// combat damage that would be dealt this turn, with no recipient or source
	// object ("Prevent all combat damage that would be dealt this turn." — Spike
	// Weaver, Holy Day). It is mutually exclusive with PreventDamageTo and
	// PreventDamageBy.
	PreventDamageGlobal bool `json:",omitempty"`
	// SpellsCantBeCounteredNextOnly reports that an EffectSpellsCantBeCountered
	// clause limits the buff to the single next spell the controller casts ("The
	// next spell you cast this turn can't be countered.") rather than every spell
	// cast this turn ("Spells you cast this turn can't be countered.").
	SpellsCantBeCounteredNextOnly bool              `json:",omitempty"`
	DelayedTiming                 DelayedTimingKind `json:",omitempty"`
	Selection                     SelectionSyntax   `json:",omitzero"`
	// DamageRecipientPair holds the two recipient groups of a dual-recipient
	// fixed group-damage effect ("deals N damage to each X and each Y"). It is
	// populated only when the recipient is exactly two "each <group>" phrases
	// joined by "and"; it is empty for every other recipient. The single
	// merged Selection cannot represent two distinct groups, so lowering emits
	// one damage instruction per recipient in Oracle order instead.
	DamageRecipientPair []SelectionSyntax `json:",omitempty"`
	// DamageRecipientReference marks a damage recipient that is the controller or
	// owner of a referenced object (the prior removal target), as in "deals 2
	// damage to that land's controller". It is None for every other recipient.
	DamageRecipientReference DamageRecipientReferenceKind `json:",omitempty"`
	// EachSourceDamageGroup holds the source group of an "each <group> deals N
	// damage to its controller/owner" effect ("Each creature deals 1 damage to
	// its controller."), where every member of the group is the damage source
	// dealing to the player who controls (or owns) it. It is populated only when
	// the subject begins with "each", parses to a recognized group, and the
	// recipient is exactly the bare "its controller"/"its owner"; it is empty for
	// every other shape. EachSourceDamageRecipient records the per-source
	// recipient role.
	EachSourceDamageGroup     SelectionSyntax              `json:",omitzero"`
	EachSourceDamageRecipient DamageRecipientReferenceKind `json:",omitempty"`
	// HasSelfDamageRider reports a "... and N damage to you" rider appended to a
	// single-target deal-damage clause ("deals A damage to any target and B
	// damage to you"). SelfDamageRiderValue holds the fixed self-damage amount
	// B; the recipient is the source's own controller. Lowering emits a second
	// Damage instruction to that controller after the primary target damage.
	HasSelfDamageRider   bool `json:",omitempty"`
	SelfDamageRiderValue int  `json:",omitempty"`
	// TargetControllerDamageRiderRecipient marks a "... and B damage to that
	// creature's controller/owner" rider appended to a single-target deal-damage
	// clause ("deals A damage to target creature and B damage to that creature's
	// controller"). It names whether the rider hits the primary target's
	// controller or owner and is None when no such rider is present.
	// TargetControllerDamageRiderValue holds the fixed rider amount B. Lowering
	// emits a second Damage instruction to that player after the primary target
	// damage.
	TargetControllerDamageRiderRecipient DamageRecipientReferenceKind `json:",omitempty"`
	TargetControllerDamageRiderValue     int                          `json:",omitempty"`
	// HasSecondTargetDamageRider reports a "... and B damage to <second target>"
	// rider appended to a single-target deal-damage clause whose second clause
	// names its own target ("deals A damage to target creature and B damage to
	// target player or planeswalker"). SecondTargetDamageRiderValue holds the
	// fixed amount B; the recipient is the clause's second target. Lowering
	// emits a second Damage instruction to that target after the primary one.
	HasSecondTargetDamageRider   bool               `json:",omitempty"`
	SecondTargetDamageRiderValue int                `json:",omitempty"`
	Amount                       EffectAmountSyntax `json:",omitzero"`
	PowerDelta                   SignedAmountSyntax `json:",omitzero"`
	ToughnessDelta               SignedAmountSyntax `json:",omitzero"`
	// TokenPower/TokenToughness/TokenPTKnown hold a created token's fixed
	// power/toughness (e.g. "1/1"). Known is false for tokens with no printed
	// power/toughness (named artifact tokens like Treasure).
	TokenPower     int  `json:",omitempty"`
	TokenToughness int  `json:",omitempty"`
	TokenPTKnown   bool `json:",omitempty"`
	// TokenPTVariableX reports a created token whose printed power and toughness
	// are both the variable "X" ("an X/X ... token"), rather than fixed integers.
	// The value of X is supplied by a binding clause elsewhere in the ability
	// ("where X is <dynamic>"); downstream layers resolve it into TokenPTDynamic.
	// It is false for fixed-power/toughness tokens.
	TokenPTVariableX bool `json:",omitempty"`
	// TokenPTDynamic names the rules-derived amount a variable "X/X" token's
	// power and toughness each equal, bound from the ability's "where X is
	// <dynamic>" clause. It is set only when TokenPTVariableX is true and the
	// binding resolves to a recognized dynamic amount; lowering reads it to size
	// the token at creation. It stays empty for fixed-power/toughness tokens.
	TokenPTDynamic EffectDynamicAmountKind `json:",omitempty"`
	// TokenGrantedAbility is the full quoted ability a created token enters with
	// ("... creature token with \"When this token dies, you gain 1 life.\""),
	// parsed once through the same pipeline so downstream layers lower it from
	// the typed inner document. It is nil for tokens with no quoted-ability
	// rider.
	TokenGrantedAbility *StaticGrantedAbilitySyntax `json:"-"`
	// GainGrantedAbility is the full quoted ability a resolving ability grant
	// confers on its subject ("This creature gains \"Whenever this creature deals
	// combat damage to a player, that player loses the game.\""), parsed once
	// through the same pipeline so downstream layers lower it from the typed inner
	// document. It is nil for gain effects that confer no quoted ability.
	GainGrantedAbility *StaticGrantedAbilitySyntax `json:"-"`
	// DelayedTriggerAbility is the nested triggered ability of an
	// EffectDelayedTrigger effect, reparsed from the sentence with its "this
	// turn" window stripped ("Whenever you cast a spell this turn, put a +1/+1
	// counter on target creature you control." -> "Whenever you cast a spell,
	// put a +1/+1 counter on target creature you control."), so downstream
	// layers lower the delayed trigger from the typed inner document. It is nil
	// for effects that are not event-based delayed triggers.
	DelayedTriggerAbility *StaticGrantedAbilitySyntax `json:"-"`
	// DelayedTriggerOneShot records that an EffectDelayedTrigger fires only on
	// the first matching event ("When you next cast a creature spell this turn,
	// ...", "The next time you cast ..."), as opposed to repeating on every
	// match for the rest of the turn ("Whenever you cast a spell this turn,
	// ..."). It is meaningful only when Kind is EffectDelayedTrigger.
	DelayedTriggerOneShot bool `json:",omitempty"`
	// TokenKeywords lists every creature keyword a created token enters with, in
	// source order ("with menace and reach" -> [Menace, Reach]). The first
	// keyword is also recorded on Selection.Keyword (a "with <keyword>" selector
	// qualifier); each additional keyword is a bare conjoined keyword. The
	// create-token exactness recognizer joins them to reconstruct the "with
	// <keyword> and <keyword> ..." rider, and lowering grants one static ability
	// per keyword. It is empty for tokens with no keyword rider.
	TokenKeywords []KeywordKind `json:",omitempty"`
	// TokenName holds a created creature token's explicit Oracle name ("...
	// creature token named Koma's Coil." -> "Koma's Coil"), captured verbatim
	// from the source so the create-token exactness recognizer can reconstruct
	// the trailing "named <Name>" tail. It is empty for tokens named only by
	// their subtypes (the default).
	TokenName string `json:",omitempty"`
	// AttackDefender names the defender in a created attacking token's
	// trailing "... attacking <defender>" clause (CR 508.4), e.g. "that player or
	// a planeswalker they control." It is set only when Selection.Attacking is
	// true and the clause carries an explicit defender; the bare "... attacking."
	// clause leaves it AttackDefenderNone. The runtime EntryAttacking model
	// is defender-agnostic, so this field only drives byte-exact clause
	// reconstruction in the create-token exactness recognizer.
	AttackDefender AttackDefenderKind `json:",omitempty"`
	// AttackDefenderSpan covers the trailing "... attacking <defender>"
	// defender phrase whenever AttackDefender is set. The semantic-reference
	// scan removes this span so an anaphoric "that player" inside the defender
	// (which the defender-agnostic runtime ignores) does not surface as a
	// dangling reference. It is the zero span when no defender is recognized.
	AttackDefenderSpan shared.Span `json:"-"`
	// TokenCopyOfTarget reports that the created token is a copy of the effect's
	// single target object ("Create a token that's a copy of target creature you
	// control."). The copy source is the effect's lone target, captured in
	// Targets; the token has no printed power/toughness of its own.
	TokenCopyOfTarget bool `json:",omitempty"`
	// TokenCopyOfReference reports that the created token is a copy of the
	// effect's single explicit reference rather than a grammatical target
	// ("Create a token that's a copy of this creature[ instead]."). The copy
	// source is the effect's lone reference, captured in References; the token
	// has no printed power/toughness of its own. An optional trailing " instead"
	// (recorded separately in Replacement) is part of the recognized clause.
	TokenCopyOfReference bool `json:",omitempty"`
	// TokenCopyOfAttached reports that the created token is a copy of the
	// permanent the source is attached to ("Create a token that's a copy of
	// equipped creature" / "enchanted creature"), as on Equipment and Auras. The
	// copy source resolves at runtime to the attached permanent; the token has no
	// printed power/toughness of its own.
	TokenCopyOfAttached bool `json:",omitempty"`
	// TokenCopyOfTriggeringSet reports that the created token is a copy of one of
	// the permanents that triggered this ability, chosen by the controller
	// ("create a token that's a copy of one of them." on a "Whenever one or more
	// ... enter" trigger, Twilight Diviner). The copy source resolves at runtime
	// to a controller-chosen member of the triggering event batch; the token has
	// no printed power/toughness of its own. The "them" pronoun is the effect's
	// lone non-modifier reference.
	TokenCopyOfTriggeringSet bool `json:",omitempty"`
	// TokenCopyDropLegendary reports a copy-token "except <it/the token> isn't
	// legendary" modifier: the created token copies the source but drops the
	// Legendary supertype so it does not force the legend rule on the original.
	TokenCopyDropLegendary bool `json:",omitempty"`
	// TokenCopyEntersTapped reports a copy-token "tapped" entry modifier ("Create
	// a tapped token that's a copy of ...", "Create two tapped tokens that are
	// copies of ..."): every created copy enters the battlefield tapped.
	TokenCopyEntersTapped bool `json:",omitempty"`
	// TokenCopyGrantKeywords lists keyword abilities a copy-token gains from a
	// folded "[That token/It] gains <keyword>." rider sentence following the
	// create effect, in source order. It is empty when no such rider is folded.
	TokenCopyGrantKeywords []KeywordKind `json:",omitempty"`
	// TokenCopyGrantRiderSpan covers the folded "[That token/It] gains
	// <keyword>." rider sentence so lowering credits its tokens toward source
	// coverage. It is set only when TokenCopyGrantKeywords is non-empty.
	TokenCopyGrantRiderSpan shared.Span `json:"-"`
	// TokenChoice reports a create-token effect that offers a choice among two or
	// more complete named-token specs ("create a Food token or a Treasure token",
	// "create your choice of a Clue token, a Food token, or a Treasure token").
	// The alternatives are the Selection.SubtypesAny entries in source order; the
	// effect creates exactly one of them, not a single multi-subtype token. It is
	// false for a single-token create and for any multi-subtype creature token.
	TokenChoice   bool                      `json:",omitempty"`
	StaticSubject EffectStaticSubjectSyntax `json:",omitzero"`
	CounterKind   counter.Kind              `json:",omitempty"`
	CounterKnown  bool                      `json:",omitempty"`
	// CounterRecipientAttached reports that a counter-placement effect places its
	// counters on the permanent the source is attached to ("... on enchanted
	// creature"), the Aura recipient the runtime models with its source
	// attached-permanent reference. It is set only for the bare "enchanted
	// creature" recipient; any other wording leaves it false so lowering fails
	// closed.
	CounterRecipientAttached bool `json:",omitempty"`
	// CounterRecipientSingleChoice reports that a non-target counter-placement
	// effect places its counters on a single permanent the controller chooses
	// from a battlefield group ("put a vigilance counter on a creature you
	// control", Ajani Fells the Godsire), rather than on every member of an
	// "each <group>" recipient. The distributive "each" and singular "a"/"an"/
	// "another" forms compile to identical selectors, so this flag carries the
	// determiner distinction to lowering. It is set only when the exact singular
	// recipient reconstruction matches the source text.
	CounterRecipientSingleChoice bool `json:",omitempty"`
	// RegenerateAttached reports that an EffectRegenerate effect regenerates the
	// permanent the source is attached to ("Regenerate enchanted creature." /
	// "Regenerate equipped creature."), the Aura or Equipment recipient the
	// runtime models with its source attached-permanent reference. It is set only
	// for the bare attached-creature recipient; any other wording leaves it false
	// so lowering fails closed.
	RegenerateAttached bool `json:",omitempty"`
	// MoveCountersAll reports the kind-agnostic "move all counters" form of an
	// EffectMoveCounters effect, where every counter on the source moves to the
	// destination regardless of kind ("Move all counters from this permanent onto
	// target creature."). It is false for a specific-kind move ("Move a +1/+1
	// counter ..."), whose kind is carried in CounterKind / CounterKnown.
	MoveCountersAll bool `json:",omitempty"`
	// MoveCountersDistribute reports the "move any number of <kind> counters from
	// <source> onto other creatures" form of an EffectMoveCounters effect, where
	// the controller distributes the source's counters among a group of other
	// creatures rather than a single target ("move any number of +1/+1 counters
	// from this creature onto other creatures."). It is false for the
	// single-target move forms.
	MoveCountersDistribute bool `json:",omitempty"`
	// MoveThoseCounters reports the counter-salvage form of an EffectPut effect,
	// "put those counters on <destination>", where "those counters" back-refers
	// to the counters a triggering permanent had as it left a zone ("Whenever a
	// creature you control leaves the battlefield, if it had counters on it, put
	// those counters on target creature you control."). The counters are read
	// from the triggering event permanent's last-known information and placed on
	// the destination (a single/optional target permanent or the source itself).
	MoveThoseCounters bool `json:",omitempty"`
	// MoveCountersFromTarget reports the two-target counter-move form, where the
	// counters are read from a first chosen target permanent and placed onto a
	// second chosen target permanent ("Move a counter from target permanent you
	// control onto a second target permanent." — Nesting Grounds). It is false
	// for the self-source single-target move and the distributed group form.
	MoveCountersFromTarget bool `json:",omitempty"`
	// MoveCountersAnyKind reports the kind-unspecified single counter move ("Move
	// a counter ..."), where the controller moves one counter of any kind present
	// on the source. It is false for a named-kind move ("Move a +1/+1 counter
	// ...") and the kind-agnostic "all counters" move.
	MoveCountersAnyKind bool      `json:",omitempty"`
	FromZone            zone.Type `json:",omitempty"`
	// GraveyardZoneExile records a recognized whole-graveyard exile ("Exile
	// target player's graveyard."), naming whose graveyard is exiled. It is
	// GraveyardZoneExileNone for every other effect, including single-card
	// graveyard exile ("Exile target card from a graveyard.") and the unmodeled
	// multi-graveyard forms ("Exile all graveyards.").
	GraveyardZoneExile GraveyardZoneExileKind    `json:",omitempty"`
	ToZone             zone.Type                 `json:",omitempty"`
	Destination        EffectDestinationPosition `json:",omitempty"`
	EntersTapped       bool                      `json:",omitempty"`
	EntersTappedSelf   bool                      `json:",omitempty"`
	EntersWithCounters bool                      `json:",omitempty"`
	// EntersWithCountersGroup reports a static enters-with-counters replacement
	// that adds a counter to a group of OTHER permanents as they enter, e.g.
	// "Each other creature you control enters with an additional vigilance
	// counter on it." (Tayam, Luminous Enigma). It is distinct from the self
	// form (EntersWithCounters with a self subject). The affected permanents are
	// the controller's permanents matched by Selection; the counter kind is in
	// CounterKind / CounterKnown and exactly one counter is added.
	EntersWithCountersGroup bool `json:",omitempty"`
	// EntersTappedGroup reports a static enters-tapped replacement that taps a
	// group of OTHER permanents as they enter, e.g. "Creatures your opponents
	// control enter tapped." (Authority of the Consuls). It is distinct from the
	// self form EntersTappedSelf ("This land enters tapped."). The controller
	// scope and affected permanent types are carried in the sibling fields.
	EntersTappedGroup bool `json:",omitempty"`
	// EntersTappedGroupScope identifies whose entering permanents are tapped by an
	// EntersTappedGroup replacement. It is EntersTappedGroupControllerNone for
	// every other effect.
	EntersTappedGroupScope EntersTappedGroupControllerScope `json:",omitempty"`
	// EntersTappedGroupTypes restricts an EntersTappedGroup replacement to entering
	// permanents that have any of these card types. It is empty when the
	// replacement taps every entering permanent ("Permanents ... enter tapped.").
	EntersTappedGroupTypes []types.Card `json:",omitempty"`
	// EntersColorChoice reports a self entry replacement of the form "As this
	// <permanent> enters, choose a color." or "... choose a color other than
	// <color>." The enters verb is shared by several entry constructs, so this is
	// set only for those exact color-choice clauses (not a non-color choice).
	EntersColorChoice bool `json:",omitempty"`
	// EntersColorChoiceExclude is the single forbidden basic color of an "As this
	// <permanent> enters, choose a color other than <color>." clause (the
	// Gate/Thriving land cycle). It is empty for the unconstrained "choose a
	// color." form.
	EntersColorChoiceExclude mana.Color `json:",omitempty"`
	// EntersTypeChoice reports a self entry replacement of the form "As this
	// <permanent> enters, choose a creature type." The enters verb is shared by
	// several entry constructs, so this is set only for that exact clause.
	EntersTypeChoice bool `json:",omitempty"`
	// EntersDevour reports the Devour keyword's as-enters replacement (CR
	// 702.81): as this creature enters its controller may sacrifice any number of
	// creatures, and it enters with EntersDevourMultiplier +1/+1 counters on it
	// for each creature sacrificed. It is set only by parseDevourEffect.
	EntersDevour bool `json:",omitempty"`
	// EntersDevourMultiplier is the per-sacrificed-creature +1/+1 counter count N
	// of a Devour replacement ("Devour N"). It is zero for every other effect.
	EntersDevourMultiplier int `json:",omitempty"`
	// EntersDevourType and EntersDevourSubtype carry the structured sacrifice
	// filter of a typed Devour replacement ("Devour artifact N", "Devour land N",
	// "Devour Food N"): the permanents the controller may sacrifice as the
	// creature enters. EntersDevourType names a base card type and
	// EntersDevourSubtype a subtype; both are zero for the plain creature form,
	// which sacrifices creatures. Set only by parseDevourEffect.
	EntersDevourType    types.Card `json:",omitempty"`
	EntersDevourSubtype types.Sub  `json:",omitempty"`
	// EntersTribute reports the Tribute keyword's as-enters replacement (CR
	// 702.110): as this creature enters, an opponent of the controller's choice
	// may put EntersTributeCount +1/+1 counters on it. It is set only by
	// parseTributeEffect.
	EntersTribute bool `json:",omitempty"`
	// EntersTributeCount is the +1/+1 counter count N of a Tribute replacement
	// ("Tribute N"). It is zero for every other effect.
	EntersTributeCount int `json:",omitempty"`
	// EntersAsCopy reports a self enters-the-battlefield replacement that has the
	// permanent enter as a copy of another permanent chosen as it enters ("You
	// may have this creature enter the battlefield as a copy of any creature on
	// the battlefield.", Clone). The copied-permanent filter is carried in
	// Selection; EntersAsCopyOptional, EntersAsCopyNotLegendary, and
	// EntersAsCopyAddTypes carry the "you may" form and the recognized copiable
	// riders. It is set only by the dedicated copy-replacement recognizer.
	EntersAsCopy bool `json:",omitempty"`
	// EntersAsCopyOptional reports the "You may have ..." form of an EntersAsCopy
	// replacement. It is false for the mandatory "this creature enters as a copy
	// of ..." form.
	EntersAsCopyOptional bool `json:",omitempty"`
	// EntersAsCopyNotLegendary reports the "except it isn't legendary" copiable
	// rider on an EntersAsCopy replacement.
	EntersAsCopyNotLegendary bool `json:",omitempty"`
	// EntersAsCopyAddTypes lists the card types added by the "except it's an
	// <type> in addition to its other types" copiable rider on an EntersAsCopy
	// replacement (Phyrexian Metamorph). It is empty for every other replacement.
	EntersAsCopyAddTypes []types.Card `json:",omitempty"`
	// EntersAsCopyAddSubtypes lists the subtypes added by the "except it's a
	// <subtype> in addition to its other types" copiable rider on an EntersAsCopy
	// replacement (Mockingbird's "a Bird", Synth Infiltrator's "a Synth"). It is
	// empty for every other replacement.
	EntersAsCopyAddSubtypes []types.Sub `json:",omitempty"`
	// EntersAsCopyConditionalCounters lists the conditional copiable counter
	// riders of an EntersAsCopy replacement ("it enters with an additional +1/+1
	// counter on it if it's a creature", "... loyalty counter ... if it's a
	// planeswalker"; Spark Double). It is empty for every other replacement.
	EntersAsCopyConditionalCounters []EntersAsCopyConditionalCounter `json:",omitempty"`
	// BecomeCopyUntilEndOfTurn reports the "until end of turn" duration on an
	// EffectBecomeCopy resolving copy (Mirage Mirror). When false the copy lasts
	// permanently (Thespian's Stage).
	BecomeCopyUntilEndOfTurn bool `json:",omitempty"`
	// BecomeCopyRetainsThisAbility reports the "except it has this ability"
	// copiable rider on an EffectBecomeCopy, which keeps the source's own copy
	// ability so it can become a copy again (Thespian's Stage).
	BecomeCopyRetainsThisAbility bool `json:",omitempty"`
	// BecomeCopyAddKeywords lists the keywords granted by an "except it has
	// <keyword>" copiable rider on an EffectBecomeCopy. It is empty otherwise.
	BecomeCopyAddKeywords []KeywordKind `json:",omitempty"`
	// BecomeTypeAddTypes lists the card types added by an EffectBecomeType
	// targeted type-change ("Target permanent becomes an artifact in addition to
	// its other types until end of turn.", Liquimetal Torque). The types are
	// added to the target without removing its existing types.
	BecomeTypeAddTypes []types.Card `json:",omitempty"`
	// BecomeTypeUntilEndOfTurn reports the "until end of turn" duration on an
	// EffectBecomeType targeted type-change. It is always set for the recognized
	// Liquimetal form; a permanent form is not yet recognized.
	BecomeTypeUntilEndOfTurn bool `json:",omitempty"`
	// Polymorph* carry the structured payload of an EffectPolymorph resolving
	// polymorph ("Until end of turn, target creature loses all abilities and
	// becomes a <color> <subtype> with base power and toughness N/N."). The
	// effect always removes all abilities and sets the base power/toughness; the
	// colors and subtypes SET the creature's color and creature type. The
	// duration is always until end of turn for the recognized form.
	PolymorphColors        []Color     `json:"-"`
	PolymorphColorless     bool        `json:",omitempty"`
	PolymorphSubtypes      []types.Sub `json:"-"`
	PolymorphBasePower     int         `json:",omitempty"`
	PolymorphBaseToughness int         `json:",omitempty"`
	// EntersAsCopyUntilEndOfTurn reports the temporary "become a copy of <filter>
	// until end of turn" form of an EntersAsCopy replacement (Cursed Mirror),
	// where the copy effect lasts until end of turn instead of as long as the
	// permanent remains on the battlefield. It is false for the permanent
	// enter-as-copy forms (Clone, Spark Double).
	EntersAsCopyUntilEndOfTurn bool `json:",omitempty"`
	// EntersAsCopyAddKeywords lists the keywords granted by the "except it has
	// <keyword>" copiable rider on an EntersAsCopy replacement (Cursed Mirror's
	// "except it has haste"). It is empty for every other replacement.
	EntersAsCopyAddKeywords []KeywordKind `json:",omitempty"`
	UnderYourControl        bool          `json:",omitempty"`
	CastAsAdventure         bool          `json:",omitempty"`
	// CastWithoutPayingManaCost reports a cast effect carrying the free-cast
	// rider "... without paying its mana cost" ("(You may) cast <spell> from
	// <zone> without paying its mana cost."). It is false for every other cast
	// effect, including ones that pay an alternative or normal cost.
	CastWithoutPayingManaCost bool `json:",omitempty"`
	Negated                   bool `json:",omitempty"`
	// FallbackOnInability marks an effect whose subject is a "who can't" relative
	// clause ("Each player who can't discards a card."): it applies only to
	// players who couldn't satisfy the immediately preceding required action. It
	// also suppresses the spurious negation the "can't" qualifier would otherwise
	// trigger, so the effect keeps its plain (non-negated) classification.
	FallbackOnInability bool `json:",omitempty"`
	Optional            bool `json:",omitempty"`
	// Divided reports a "deals N damage divided as you choose among <targets>"
	// effect: a fixed total split among the chosen targets, at least one each.
	Divided      bool             `json:",omitempty"`
	OptionalSpan shared.Span      `json:"-"`
	Symbol       string           `json:",omitempty"`
	Mana         EffectManaSyntax `json:",omitzero"`
	// SourceSpellCostReduction marks the EffectCast effect of the exact
	// single-clause ability "This spell costs {N} less to cast for each
	// <countable battlefield object>." It is a typed cast cost modifier rather
	// than a resolving effect: lowering reads SourceSpellCostReductionAmount (the
	// per-object generic reduction N) together with this effect's typed Amount
	// (the per-object battlefield count and its selection) and never inspects the
	// source text. It is set only when the ability matches that exact shape.
	SourceSpellCostReduction       bool `json:",omitempty"`
	SourceSpellCostReductionAmount int  `json:",omitempty"`
	// SourceSpellCostReductionDynamic marks the EffectCast effect of the exact
	// single-clause ability "This spell costs {X} less to cast, where X is
	// <dynamic amount>." (e.g. The Great Henge: the greatest power among creatures
	// you control). The reduction amount is the effect's typed Amount itself; the
	// per-object SourceSpellCostReductionAmount is unused for this form. It is set
	// only when the ability matches that exact shape and the dynamic amount is one
	// lowering can evaluate at cost time.
	SourceSpellCostReductionDynamic bool                    `json:",omitempty"`
	Replacement                     EffectReplacementSyntax `json:",omitzero"`
	References                      []Reference             `json:",omitempty"`
	SubjectReferences               []Reference             `json:",omitempty"`
	Targets                         []TargetSyntax          `json:",omitempty"`
	SubjectTargets                  []TargetSyntax          `json:",omitempty"`
	Payment                         EffectPaymentSyntax     `json:",omitzero"`
	Exact                           bool                    `json:",omitempty"`
	// KeywordGrantChoice marks an EffectGain keyword grant whose keyword list is a
	// disjunction ("gains banding, first strike, or trample") rather than a
	// conjunction. The disjunction means the controller chooses exactly one of the
	// listed keywords at resolution; lowering reads this to emit a keyword-choice
	// grant instead of granting every listed keyword. It is set only for gain
	// effects whose body is a recognized disjunctive keyword list.
	KeywordGrantChoice bool `json:",omitempty"`
	// RevealUntilThenPut marks each effect of the recognized closed "reveal
	// cards from the top of <library> until <player> reveal a <type> card, then
	// put those cards into <zone>" sequence. The parser keeps the three-effect
	// shape [Reveal, match-Reveal, Put], sets this marker on each, captures the
	// until predicate on the match-Reveal's Selection, and the destination on
	// the Put's ToZone. Lowering reads these typed fields to emit a single
	// RevealUntil primitive; the marker is false for every other effect.
	RevealUntilThenPut      bool   `json:",omitempty"`
	RequiresOrderedLowering bool   `json:",omitempty"`
	HasUnrecognizedSibling  bool   `json:",omitempty"`
	UnsupportedDetail       string `json:",omitempty"`
	// Order is the effect's dense source-order rank (of Span); VerbOrder is the
	// rank of VerbSpan. Downstream stages compare these ranks to order effects
	// and bind references to effect verbs without inspecting byte offsets.
	Order     shared.SourceOrder `json:"-"`
	VerbOrder shared.SourceOrder `json:"-"`
	// LifeObject reports that a gain/lose effect's grammatical object is the
	// player's life (e.g. "gain 3 life", "loses that much life"), as opposed to
	// a keyword or quoted ability ("gains shadow", "loses protection from
	// black"). It lets consumers route only true life changes to the life
	// lowerer rather than misclassifying keyword/ability grants and losses.
	LifeObject bool `json:",omitempty"`
	// PreventRegeneration reports that a destroy effect is followed by a
	// regeneration rider ("It/They can't be regenerated."). The rider is a
	// separate zero-effect sentence whose pronoun refers to the destroyed
	// permanents; the parser folds it onto the destroy effect so lowering
	// emits a destruction that bypasses regeneration shields.
	PreventRegeneration bool `json:",omitempty"`
	// RegenerationRiderSpan covers the rider sentence's semantic tokens so the
	// lowerer can credit them toward source coverage. It is set only when
	// PreventRegeneration is true.
	RegenerationRiderSpan shared.Span `json:"-"`
	// CopyMayChooseNewTargets reports that a copy-stack-object effect is
	// followed by the optional "You may choose new targets for the copy[ies]."
	// rider. The rider is a separate sentence whose "the copy" subject denotes
	// the copy this effect creates; the parser folds it onto the copy effect so
	// lowering emits a CopyStackObject that may re-choose the copy's targets.
	CopyMayChooseNewTargets bool `json:",omitempty"`
	// CopyChooseNewTargetsRiderSpan covers the folded "You may choose new
	// targets for the copy[ies]." rider sentence so the lowerer can credit it
	// toward source coverage. It is set only when CopyMayChooseNewTargets is true.
	CopyChooseNewTargetsRiderSpan shared.Span `json:"-"`
	// Dig holds the structured fields of an impulse "Put N of them into your
	// hand and the rest into your graveyard." clause. It is set only on the
	// EffectPut half of an impulse dig sequence (Dig.Put true); the look half is
	// classified EffectDig with the looked-at count in Amount.
	Dig DigSyntax `json:",omitzero"`
	// HandLibraryPut marks an exact own-hand-to-library-top clause whose selected
	// cards are ordered by the resolving player.
	HandLibraryPut HandLibraryPutSyntax `json:",omitzero"`
	// HandDiscard marks an exact fixed-cardinality discard chosen from the
	// resolving controller's hand.
	HandDiscard HandDiscardSyntax `json:",omitzero"`
	// SearchSplit holds the structured fields of a split-destination put clause
	// "put one <slot> and the other <slot>" that distributes the cards found by a
	// preceding "up to two" library search across two single-card destination
	// slots. It is set only on the EffectPut half of such a search
	// (SearchSplit.Present true).
	SearchSplit SearchSplitSyntax `json:",omitzero"`
	// ManaSpendRider holds the structured fields of a mana-spend rider sentence
	// ("When that mana is spent to cast a creature spell that shares a creature
	// type with your commander, scry N"). It is set only on a synthesized
	// EffectManaSpendRider effect that replaces the sentence's generic cast/scry
	// effects when the exact rider wording is recognized; it is nil otherwise.
	ManaSpendRider *ManaSpendRiderSyntax `json:",omitempty"`
	// SearchSharedSubtype reports the "that share a land type" correlation rider
	// on a multi-card library search ("up to two basic land cards that share a
	// land type"). It is set only on the EffectSearch clause carrying the rider;
	// the runtime requires every found card to share a land subtype with the
	// others (CR 701.19), modeling Myriad Landscape.
	SearchSharedSubtype bool `json:",omitempty"`
	// SearchDestination carries the ordered destination of an exact library
	// search whose found card stays in the library. It is currently set only for
	// the singular "then shuffle and put that card on top" family.
	SearchDestination EffectDestinationPosition `json:",omitempty"`
	// SearchControl carries the controller rider on a search-and-put-onto-the-
	// battlefield clause ("put it onto the battlefield tapped under target
	// player's control", Yavimaya Dryad): the found permanent enters under the
	// named target player's control rather than the searching player's. The zero
	// value (no rider) means the found card enters under the searcher's control.
	SearchControl SearchControlRider `json:",omitempty"`
	// DiscardEntireHand marks a "discard their hand" clause ("Each player
	// discards their hand", "Discard your hand", "Target player discards their
	// hand"): the affected player discards every card in hand rather than a fixed
	// count. The discarded amount is unknown at parse time.
	DiscardEntireHand bool `json:",omitempty"`
	// CounteredSpellExileReplacement marks the exact counter rider "If that
	// spell is countered this way, exile it instead of putting it into its
	// owner's graveyard." (CR 614 replacement). It pairs with a preceding
	// counter effect so lowering can emit a single counter-and-exile primitive.
	CounteredSpellExileReplacement bool `json:",omitempty"`
	// ExileUntilSourceLeaves marks the exact O-Ring exile clause "exile <target>
	// until <this permanent> leaves the battlefield." (Banisher Priest, Banishing
	// Light, Journey to Nowhere-style enchantments). The exiled permanent is
	// linked to the source so a paired leaves-the-battlefield trigger returns it;
	// the trailing self-reference is the duration anchor, not a target.
	ExileUntilSourceLeaves bool `json:",omitempty"`
	// ReturnExiledCard marks the explicit O-Ring leaves-the-battlefield clause
	// "return the exiled card to the battlefield under its owner's control."
	// (Oblivion Ring, Journey to Nowhere, Fiend Hunter). The returned card is the
	// one a sibling enters-the-battlefield exile removed, identified by the
	// source link rather than a target. Lowering emits a linked battlefield
	// return; it is false for every other return shape.
	ReturnExiledCard bool `json:",omitempty"`
	// ExileEntireHand marks the exact involuntary whole-hand exile clause "Exile
	// all cards from your hand." (Wormfang Behemoth). The exiled cards are linked
	// to the source so a paired leaves-the-battlefield "return the exiled cards"
	// trigger returns the set; it is false for every other exile shape.
	ExileEntireHand bool `json:",omitempty"`
	// ReturnExiledCardsToHand marks the exact leaves-the-battlefield clause
	// "Return the exiled cards to their owner's hand." (Wormfang Behemoth). The
	// returned cards are the set a sibling ExileEntireHand exiled, identified by
	// the source link rather than a target; lowering emits the linked return to
	// hand. It is false for every other return shape.
	ReturnExiledCardsToHand bool `json:",omitempty"`
	// Additional marks a draw clause whose counted cards carry the "additional"
	// qualifier ("draw two additional cards", "draw an additional card"), as on
	// draw-step triggers like Sylvan Library. Drawing N additional cards is
	// mechanically a plain draw of N cards, so consumers treat it as one; the
	// flag exists only so exact reconstruction can restore the "additional"
	// word. It is false for every plain draw.
	Additional bool `json:",omitempty"`
	// DoublePower and DoubleToughness mark an EffectDouble whose object is "the
	// power[ and toughness] of <group>" ("double the power and toughness of each
	// creature you control until end of turn", Unnatural Growth). Each affected
	// permanent gains +X to the named characteristic, where X is its own current
	// value, doubling it (CR 107.16). The affected group is carried in
	// StaticSubject; both flags are false for every other double effect (double
	// life, double counters, double mana).
	DoublePower     bool `json:",omitempty"`
	DoubleToughness bool `json:",omitempty"`
	// DoubleSourceCounters marks an EffectDouble whose object is "the number of
	// <kind> counters on <self>" ("double the number of +1/+1 counters on this
	// creature", Mossborn Hydra). The source permanent gains additional counters
	// of DoubleSourceCounterKind equal to its current count of that kind,
	// doubling it. It is false for power/toughness and other double effects.
	DoubleSourceCounters    bool         `json:",omitempty"`
	DoubleSourceCounterKind counter.Kind `json:",omitempty"`
	// DoubleCountersTarget marks a counter-doubling EffectDouble whose object is
	// a "target ..." permanent ("double the number of +1/+1 counters on target
	// creature", Gilder Bairn) rather than the source itself, so lowering doubles
	// the sentence's target. DoubleCountersAllKinds marks the "each kind of
	// counter" form ("double the number of each kind of counter on target
	// artifact, creature, or land", Vorel of the Hull Clade), doubling every
	// counter kind present instead of the single DoubleSourceCounterKind. Both
	// are false for the original self single-kind form.
	DoubleCountersTarget   bool `json:",omitempty"`
	DoubleCountersAllKinds bool `json:",omitempty"`
	// UnderOwnersControl marks a battlefield-destination effect carrying the
	// rider "under their owners' control" / "under its owner's control" (Open
	// the Vaults, Planar Birth, Living Death), where each moved card enters under
	// the control of its own owner rather than the resolving player. It is false
	// for the bare and "under your control" forms.
	UnderOwnersControl bool `json:",omitempty"`
	// TokenCopyOfForEach reports a per-each copy-token create whose copy source
	// is each member of a controlled battlefield group (Second Harvest). The
	// iterated group is carried in TokenCopyForEachGroup.
	TokenCopyOfForEach bool `json:",omitempty"`
	// TokenCopyForEachGroup carries the controlled battlefield group iterated by
	// a TokenCopyOfForEach create. Nil unless TokenCopyOfForEach is set.
	TokenCopyForEachGroup *SelectionSyntax `json:",omitempty"`
	// PunisherSacrifice and PunisherDiscard mark the alternatives offered by an
	// EffectPunisherLoseLife effect ("... unless that player sacrifices a
	// permanent of their choice or discards a card."): PunisherSacrifice records
	// that a sacrifice alternative (filtered by Selection) is offered, and
	// PunisherDiscard records that a discard-a-card alternative is offered. Both
	// are false for every other effect.
	PunisherSacrifice bool `json:",omitempty"`
	PunisherDiscard   bool `json:",omitempty"`
	// RepeatBody holds the sub-effect(s) of an EffectRepeatProcess loop ("Repeat
	// the following process X times. <body>"). It is nil for every other effect.
	RepeatBody []EffectSyntax `json:",omitempty"`
	// ReturnAsEnchantment reports that a return-to-battlefield effect is followed
	// by an "It's an enchantment." rider (the Enduring enchantment-creature
	// cycle), so the returned permanent enters as an Enchantment (losing its
	// creature type). The rider is a separate zero-effect sentence whose pronoun
	// refers to the returned card; the parser folds it onto the return effect.
	ReturnAsEnchantment bool `json:",omitempty"`
	// ReturnAsEnchantmentRiderSpan covers the rider sentence's semantic tokens so
	// the lowerer can credit them toward source coverage. It is set only when
	// ReturnAsEnchantment is true.
	ReturnAsEnchantmentRiderSpan shared.Span `json:"-"`
	// PlayFromTopPayLife reports that an EffectPlayFromLibraryTop grant is
	// followed by the "If you cast a spell this way, pay life equal to its mana
	// value rather than pay its mana cost." rider, so spells cast from the top of
	// the library via the grant pay life equal to their mana value instead of
	// their mana cost. The rider is a separate zero-effect sentence the parser
	// folds onto the grant effect.
	PlayFromTopPayLife bool `json:",omitempty"`
	// PlayFromTopPayLifeRiderSpan covers the pay-life rider sentence's semantic
	// tokens so the lowerer can credit them toward source coverage. It is set
	// only when PlayFromTopPayLife is true.
	PlayFromTopPayLifeRiderSpan shared.Span `json:"-"`
	// AmassSubtype is the creature subtype named by an EffectAmass keyword action
	// ("Amass Orcs N" -> Orc, "Amass Zombies N" -> Zombie). The amassed Army
	// token enters as a 0/0 black creature with this subtype and Army. The older
	// untyped "Amass N" form names no subtype in its text and defaults to Zombie.
	// It is empty for every other effect.
	AmassSubtype types.Sub `json:",omitempty"`
	// AdditionalCombatPhase reports an "After this [main] phase, there is an
	// additional combat phase[ followed by an additional main phase]." effect
	// (Aggravated Assault, Aurelia the Warleader, World at War): it inserts an
	// extra combat phase into the current turn. AdditionalMainPhase reports the
	// optional "followed by an additional main phase" tail that inserts an extra
	// main phase after that combat phase. Both are false for every other effect;
	// AdditionalMainPhase is set only together with AdditionalCombatPhase.
	AdditionalCombatPhase bool `json:",omitempty"`
	AdditionalMainPhase   bool `json:",omitempty"`
	// DieSides is the number of faces of the die rolled by an EffectRollDie
	// effect ("roll a d20" sets DieSides to 20). It is zero for every other
	// effect kind.
	DieSides int `json:",omitempty"`
	// RevealChooseDiscard marks each effect of the recognized "Target player
	// reveals their hand. You choose a [filter] card from it. That player
	// discards that card." sequence: it is set on both the EffectReveal half and
	// the EffectDiscard half. The resolving controller chooses the discarded
	// card; the filter and the middle "You choose..." sentence's coverage span
	// live on HandChoiceDiscard (set on the EffectDiscard half). It is false for
	// every other effect.
	RevealChooseDiscard bool `json:",omitempty"`
	// HandChoiceDiscard holds the filter and coverage span of a reveal-choose-
	// discard sequence's middle "You choose a [filter] card from it." sentence.
	// It is set only on the EffectDiscard half (HandChoiceDiscard.Present true).
	HandChoiceDiscard HandChoiceDiscardSyntax `json:",omitzero"`
}

// ManaSpendConditionKind identifies the exact spend condition of a mana-spend
// rider. The set is closed; only fully modeled conditions are recognized.
type ManaSpendConditionKind string

// Mana-spend rider conditions recognized by the parser.
const (
	ManaSpendConditionUnknown ManaSpendConditionKind = ""
	// ManaSpendCastCommanderCreatureType is "spent to cast a creature spell that
	// shares a creature type with your commander".
	ManaSpendCastCommanderCreatureType ManaSpendConditionKind = "ManaSpendCastCommanderCreatureType"
	// ManaSpendCastChosenCreatureType is "spent only to cast a creature spell of
	// the chosen type".
	ManaSpendCastChosenCreatureType ManaSpendConditionKind = "ManaSpendCastChosenCreatureType"
	// ManaSpendCastLegendarySpell is "spent only to cast a legendary spell".
	ManaSpendCastLegendarySpell ManaSpendConditionKind = "ManaSpendCastLegendarySpell"
	// ManaSpendCastOrActivateChosenCreatureType is "spent only to cast a creature
	// spell of the chosen type or activate an ability of a creature source of the
	// chosen type" (Secluded Courtyard).
	ManaSpendCastOrActivateChosenCreatureType ManaSpendConditionKind = "ManaSpendCastOrActivateChosenCreatureType"
	// ManaSpendCastCreatureSpell is "spent on a creature spell" (Arena of Glory,
	// Generator Servant). It is an unrestricted bonus rider that grants the spell
	// a keyword until end of turn.
	ManaSpendCastCreatureSpell ManaSpendConditionKind = "ManaSpendCastCreatureSpell"
)

// ManaSpendRiderEffectKind identifies the exact resolving effect of a mana-spend
// rider. The set is closed; only fully modeled effects are recognized.
type ManaSpendRiderEffectKind string

// Mana-spend rider effects recognized by the parser.
const (
	ManaSpendRiderEffectUnknown ManaSpendRiderEffectKind = ""
	// ManaSpendRiderEffectScry is "scry N".
	ManaSpendRiderEffectScry ManaSpendRiderEffectKind = "ManaSpendRiderEffectScry"
	// ManaSpendRiderEffectCantBeCountered is "that spell can't be countered".
	ManaSpendRiderEffectCantBeCountered ManaSpendRiderEffectKind = "ManaSpendRiderEffectCantBeCountered"
	// ManaSpendRiderEffectGainsHasteUntilEndOfTurn is "it gains haste until end of
	// turn", granting the qualifying creature spell haste through end of turn.
	ManaSpendRiderEffectGainsHasteUntilEndOfTurn ManaSpendRiderEffectKind = "ManaSpendRiderEffectGainsHasteUntilEndOfTurn"
)

// ManaSpendRiderSyntax is the typed syntax of a recognized mana-spend rider.
type ManaSpendRiderSyntax struct {
	Span          shared.Span              `json:"-"`
	ConditionSpan shared.Span              `json:"-"`
	EffectSpan    shared.Span              `json:"-"`
	Condition     ManaSpendConditionKind   `json:",omitempty"`
	Effect        ManaSpendRiderEffectKind `json:",omitempty"`
	Restricted    bool                     `json:",omitempty"`
	ScryAmount    int                      `json:",omitempty"`
}

// EffectPaymentPayerKind identifies who may pay a cost embedded in an effect.
type EffectPaymentPayerKind string

// Embedded-effect payers recognized by the parser.
const (
	EffectPaymentPayerUnknown          EffectPaymentPayerKind = ""
	EffectPaymentPayerController       EffectPaymentPayerKind = "EffectPaymentPayerController"
	EffectPaymentPayerTargetController EffectPaymentPayerKind = "EffectPaymentPayerTargetController"
	EffectPaymentPayerEventPlayer      EffectPaymentPayerKind = "EffectPaymentPayerEventPlayer"
)

// EffectPaymentForm identifies the Oracle grammar that offers a resolution
// payment. Distinct forms can normalize to the same runtime Pay/result gate
// while preserving whether the consequence itself is optional.
type EffectPaymentForm string

// Embedded-effect payment forms recognized by the parser.
const (
	EffectPaymentFormUnknown             EffectPaymentForm = ""
	EffectPaymentFormUnless              EffectPaymentForm = "EffectPaymentFormUnless"
	EffectPaymentFormMayPayThenIfDo      EffectPaymentForm = "EffectPaymentFormMayPayThenIfDo"
	EffectPaymentFormMayPayThenIfDoesNot EffectPaymentForm = "EffectPaymentFormMayPayThenIfDoesNot"
)

// EffectPaymentSyntax is a source-spanned typed resolution payment.
type EffectPaymentSyntax struct {
	Span              shared.Span            `json:"-"`
	Form              EffectPaymentForm      `json:",omitempty"`
	Payer             EffectPaymentPayerKind `json:",omitempty"`
	ManaCost          cost.Mana              `json:",omitempty"`
	GenericManaAmount EffectAmountSyntax     `json:",omitzero"`
	// AdditionalCost is a non-mana resolution payment cost (such as "sacrifice a
	// land", "discard a card", or the fixed life portion of "pay {mana} and N
	// life") recognized in a "you may <cost>. If you do, ..." sequence. It is nil
	// for mana-only payments. ManaCost and AdditionalCost are both set for a
	// combined mana+life payment; otherwise exactly one is set.
	AdditionalCost         *Cost `json:",omitempty"`
	SuccessConditionNodeID int   `json:"-"`
	FailureConditionNodeID int   `json:"-"`
	// Order is the payment's dense source-order rank, used downstream to test
	// condition containment without byte offsets.
	Order shared.SourceOrder `json:"-"`
}

// EffectStaticSubjectKind identifies the group affected by a static resolving
// effect production.
type EffectStaticSubjectKind string

// EntersTappedGroupControllerScope identifies whose entering permanents a static
// "<permanents> enter tapped" replacement taps.
type EntersTappedGroupControllerScope string

// Enters-tapped group controller scopes recognized by the replacement grammar.
const (
	EntersTappedGroupControllerNone      EntersTappedGroupControllerScope = ""
	EntersTappedGroupControllerYou       EntersTappedGroupControllerScope = "EntersTappedGroupControllerYou"
	EntersTappedGroupControllerOpponents EntersTappedGroupControllerScope = "EntersTappedGroupControllerOpponents"
	EntersTappedGroupControllerEach      EntersTappedGroupControllerScope = "EntersTappedGroupControllerEach"
)

// Static effect subjects recognized by resolving-effect grammar.
const (
	EffectStaticSubjectNone                           EffectStaticSubjectKind = ""
	EffectStaticSubjectAttachedObject                 EffectStaticSubjectKind = "EffectStaticSubjectAttachedObject"
	EffectStaticSubjectAllCreatures                   EffectStaticSubjectKind = "EffectStaticSubjectAllCreatures"
	EffectStaticSubjectAllOtherCreatures              EffectStaticSubjectKind = "EffectStaticSubjectAllOtherCreatures"
	EffectStaticSubjectAttackingCreatures             EffectStaticSubjectKind = "EffectStaticSubjectAttackingCreatures"
	EffectStaticSubjectBlockingCreatures              EffectStaticSubjectKind = "EffectStaticSubjectBlockingCreatures"
	EffectStaticSubjectControlledPermanents           EffectStaticSubjectKind = "EffectStaticSubjectControlledPermanents"
	EffectStaticSubjectControlledLands                EffectStaticSubjectKind = "EffectStaticSubjectControlledLands"
	EffectStaticSubjectControlledCreatures            EffectStaticSubjectKind = "EffectStaticSubjectControlledCreatures"
	EffectStaticSubjectOtherControlledCreatures       EffectStaticSubjectKind = "EffectStaticSubjectOtherControlledCreatures"
	EffectStaticSubjectControlledWalls                EffectStaticSubjectKind = "EffectStaticSubjectControlledWalls"
	EffectStaticSubjectControlledArtifacts            EffectStaticSubjectKind = "EffectStaticSubjectControlledArtifacts"
	EffectStaticSubjectControlledSagas                EffectStaticSubjectKind = "EffectStaticSubjectControlledSagas"
	EffectStaticSubjectControlledTokens               EffectStaticSubjectKind = "EffectStaticSubjectControlledTokens"
	EffectStaticSubjectOpponentControlledCreatures    EffectStaticSubjectKind = "EffectStaticSubjectOpponentControlledCreatures"
	EffectStaticSubjectControlledCreatureSubtype      EffectStaticSubjectKind = "EffectStaticSubjectControlledCreatureSubtype"
	EffectStaticSubjectOtherControlledCreatureSubtype EffectStaticSubjectKind = "EffectStaticSubjectOtherControlledCreatureSubtype"
	EffectStaticSubjectAllCreatureSubtype             EffectStaticSubjectKind = "EffectStaticSubjectAllCreatureSubtype"
	EffectStaticSubjectOtherCreatureSubtype           EffectStaticSubjectKind = "EffectStaticSubjectOtherCreatureSubtype"
	EffectStaticSubjectControlledAttackingCreatures   EffectStaticSubjectKind = "EffectStaticSubjectControlledAttackingCreatures"
	EffectStaticSubjectControlledCreatureTokens       EffectStaticSubjectKind = "EffectStaticSubjectControlledCreatureTokens"
	EffectStaticSubjectBattlefieldCreatureTokens      EffectStaticSubjectKind = "EffectStaticSubjectBattlefieldCreatureTokens"
	EffectStaticSubjectControlledLegendaryCreatures   EffectStaticSubjectKind = "EffectStaticSubjectControlledLegendaryCreatures"
	EffectStaticSubjectControlledUntappedCreatures    EffectStaticSubjectKind = "EffectStaticSubjectControlledUntappedCreatures"
	EffectStaticSubjectControlledModifiedCreatures    EffectStaticSubjectKind = "EffectStaticSubjectControlledModifiedCreatures"
	EffectStaticSubjectOtherControlledTappedCreatures EffectStaticSubjectKind = "EffectStaticSubjectOtherControlledTappedCreatures"

	EffectStaticSubjectControlledArtifactCreatures      EffectStaticSubjectKind = "EffectStaticSubjectControlledArtifactCreatures"
	EffectStaticSubjectOtherControlledArtifactCreatures EffectStaticSubjectKind = "EffectStaticSubjectOtherControlledArtifactCreatures"
	EffectStaticSubjectControlledNontokenCreatures      EffectStaticSubjectKind = "EffectStaticSubjectControlledNontokenCreatures"
	EffectStaticSubjectOtherControlledNontokenCreatures EffectStaticSubjectKind = "EffectStaticSubjectOtherControlledNontokenCreatures"

	// EffectStaticSubjectAllLands names every land on the battlefield regardless
	// of controller ("Each land ...", "All lands ..."). It is the affected group
	// of the continuous land-type-adding statics printed on cards such as
	// Yavimaya, Cradle of Growth and Urborg, Tomb of Yawgmoth.
	EffectStaticSubjectAllLands EffectStaticSubjectKind = "EffectStaticSubjectAllLands"

	// EffectStaticSubjectControlledCreaturesChosenType and its "other" sibling
	// name the controlled creatures whose creature type matches the source
	// permanent's entry-time creature-type choice ("creatures you control of the
	// chosen type ..."), the affected group of chosen-type anthems such as
	// Patchwork Banner and Adaptive Automaton.
	EffectStaticSubjectControlledCreaturesChosenType      EffectStaticSubjectKind = "EffectStaticSubjectControlledCreaturesChosenType"
	EffectStaticSubjectOtherControlledCreaturesChosenType EffectStaticSubjectKind = "EffectStaticSubjectOtherControlledCreaturesChosenType"

	// EffectStaticSubjectOpponentControlledPermanents names every permanent your
	// opponents control ("Permanents your opponents control lose hexproof and
	// indestructible until end of turn."), the affected group of resolving
	// keyword removals such as Shadowspear's activated ability.
	EffectStaticSubjectOpponentControlledPermanents EffectStaticSubjectKind = "EffectStaticSubjectOpponentControlledPermanents"

	// EffectStaticSubjectOtherAttackingCreatures names every attacking creature
	// except the source ("each other attacking creature gets +1/+0 until end of
	// turn."), the affected group of the Battle cry triggered ability.
	EffectStaticSubjectOtherAttackingCreatures EffectStaticSubjectKind = "EffectStaticSubjectOtherAttackingCreatures"

	// EffectStaticSubjectOtherControlledPermanents names every permanent you
	// control except the source ("Other permanents you control have
	// indestructible.", Avacyn, Angel of Hope), the self-excluded sibling of
	// EffectStaticSubjectControlledPermanents.
	EffectStaticSubjectOtherControlledPermanents EffectStaticSubjectKind = "EffectStaticSubjectOtherControlledPermanents"

	// EffectStaticSubjectControlledNonlegendaryCreatures names the nonlegendary
	// creatures you control ("Nonlegendary creatures you control get +1/+1.",
	// Flowering of the White Tree), the excluded-supertype sibling of
	// EffectStaticSubjectControlledLegendaryCreatures. Downstream maps it onto a
	// Selection that excludes the Legendary supertype.
	EffectStaticSubjectControlledNonlegendaryCreatures EffectStaticSubjectKind = "EffectStaticSubjectControlledNonlegendaryCreatures"

	// EffectStaticSubjectControlledCommanderCreatures names the commander
	// creatures you control ("Commander creatures you control get +2/+2 and have
	// indestructible.", Bastion Protector), the affected group of commander
	// anthems. Downstream maps it onto a Selection that requires the creature
	// card type and that the permanent is a commander.
	EffectStaticSubjectControlledCommanderCreatures EffectStaticSubjectKind = "EffectStaticSubjectControlledCommanderCreatures"

	// EffectStaticSubjectControlledCommanders names the commander permanents you
	// control ("Commanders you control have hexproof.", Guardian Augmenter)
	// regardless of card type. Downstream maps it onto a Selection that requires
	// only that the permanent is a commander.
	EffectStaticSubjectControlledCommanders EffectStaticSubjectKind = "EffectStaticSubjectControlledCommanders"
)

// EffectStaticSubjectSyntax is a source-spanned typed static-effect subject.
type EffectStaticSubjectSyntax struct {
	Kind         EffectStaticSubjectKind `json:",omitempty"`
	Span         shared.Span             `json:"-"`
	Subtype      types.Sub               `json:",omitempty"`
	SubtypeText  string                  `json:",omitempty"`
	SubtypeKnown bool                    `json:",omitempty"`
	// ExcludedSubtype marks the Subtype as a "non-<subtype>" exclusion rather
	// than a required subtype ("Non-Human creatures you control get ..."). When
	// set, the affected group matches creatures that do NOT carry Subtype.
	ExcludedSubtype bool `json:",omitempty"`

	// Colors, Colorless, and Multicolored carry an optional color filter
	// constraining the affected creature group ("Other red creatures you
	// control ..."). Colors lists single-color words matched disjunctively;
	// Colorless and Multicolored are the color-family qualifiers. They are
	// mutually exclusive shapes downstream maps onto a Selection color filter.
	Colors       []Color `json:",omitempty"`
	Colorless    bool    `json:",omitempty"`
	Multicolored bool    `json:",omitempty"`

	// Keyword and ExcludedKeyword carry an optional single keyword filter
	// constraining the affected creature group ("Creatures with flying ...",
	// "Creatures without flying ..."). At most one is set: Keyword requires the
	// named keyword be present, ExcludedKeyword requires it be absent. They map
	// downstream onto a Selection keyword predicate.
	Keyword         KeywordKind `json:",omitempty"`
	ExcludedKeyword KeywordKind `json:",omitempty"`

	// CounterRequired records a "with a <kind> counter on it/them" qualifier
	// constraining the affected creature group to members carrying that counter
	// ("Each creature you control with a +1/+1 counter on it has ..."); CounterKind
	// names the required counter. They map downstream onto a Selection
	// MatchCounter predicate. When the qualifier names no kind ("with a counter
	// on it", Rishkar), CounterAny is set instead and the group matches members
	// carrying a counter of any kind, mapping onto a MatchAnyCounter predicate.
	CounterRequired bool         `json:",omitempty"`
	CounterKind     counter.Kind `json:",omitempty"`
	CounterAny      bool         `json:",omitempty"`

	// ChosenColorFromEntry records a trailing "of the chosen color" qualifier on
	// the affected group ("Creatures you control of the chosen color get ..."),
	// constraining each matched permanent to share the color the source
	// permanent chose as it entered (Heraldic Banner). It maps downstream onto
	// the runtime Selection.ColorChoice = ColorChoiceSourceEntry predicate.
	ChosenColorFromEntry bool `json:",omitempty"`

	// Power and Toughness carry an optional numeric power/toughness comparison
	// constraining the affected group ("Creatures you control with power 1 or
	// less ..."). MatchPower and MatchToughness record whether each comparison
	// is active. PowerOrToughness marks the disjunctive "with power or toughness
	// N or less" form (Tetsuko Umezawa), where a member matches if EITHER its
	// power OR its toughness satisfies the bound; both Power and Toughness then
	// carry the same comparison and lowering emits a Selection.AnyOf of the two
	// single-characteristic alternatives.
	Power            compare.Int `json:",omitzero"`
	MatchPower       bool        `json:",omitempty"`
	Toughness        compare.Int `json:",omitzero"`
	MatchToughness   bool        `json:",omitempty"`
	PowerOrToughness bool        `json:",omitempty"`

	// PowerLessThanSource and PowerGreaterThanSource constrain the affected group
	// by comparing each member's power to the static's SOURCE permanent's power
	// ("Creatures you control with power greater than <source>'s power ...",
	// Champion of Lambholt). They are source-relative, so they carry no fixed
	// comparison and lower to the runtime Selection.PowerLessThanSource /
	// Selection.PowerGreaterThanSource predicates.
	PowerLessThanSource    bool `json:",omitempty"`
	PowerGreaterThanSource bool `json:",omitempty"`

	// ExcludedTypes lists card types a matched permanent must NOT carry, set by
	// a "non-<type>" prefix on the group noun ("Nonland permanents you control
	// are artifacts ...", Encroaching Mycosynth). It lowers onto the runtime
	// Selection.ExcludedTypes predicate.
	ExcludedTypes []CardType `json:",omitempty"`
}
