package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CounterPlacement describes counters a permanent enters with.
type CounterPlacement struct {
	Kind   counter.Kind
	Amount int
	// AmountFromX places a number of counters equal to the value of X chosen for
	// the entering permanent's spell (CR 107.3, CR 614). Amount is ignored when
	// it is set; a permanent that entered without a cast X (a copy or a
	// put-onto-the-battlefield effect) enters with zero such counters. It backs
	// "This creature enters with X +1/+1 counters on it." (Walking Ballista,
	// Hangarback Walker, Endless One).
	AmountFromX bool
	// Dynamic places a number of counters equal to a rules-derived amount
	// evaluated as the permanent enters (CR 608.2c), such as "for each creature
	// card in your graveyard." (Golgari Grave-Troll). When set, Amount and
	// AmountFromX are ignored and a non-positive evaluated amount places no
	// counters. The amount is immutable rules configuration shared across clones.
	Dynamic opt.V[*DynamicAmount]
}

// ConditionalCounterPlacement places Amount counters of Kind on an entering
// permanent only when its (copied) card types include IfType. It backs the
// conditional copiable counter riders of enters-as-copy replacements, such as
// Spark Double's "it enters with an additional +1/+1 counter on it if it's a
// creature" and "it enters with an additional loyalty counter on it if it's a
// planeswalker".
type ConditionalCounterPlacement struct {
	Kind   counter.Kind
	Amount int
	IfType types.Card
}

// PreventionShield prevents an amount of future damage to a player or
// permanent. When All is set the shield has no fixed capacity and prevents
// every qualifying event for its duration; when CombatOnly is set it prevents
// only combat damage. SourcePermanentID, when non-zero, prevents damage dealt
// BY that permanent rather than damage dealt TO PermanentID/Player. When OneShot
// is set the shield prevents a single qualifying event and then expires (the
// "next time ... prevent that damage" shields). SourceColors, when non-empty,
// restricts the shield to damage from a source of one of the listed colors.
//
// When RedirectToSourceController is set, a one-shot shield deals the amount it
// prevents to the prevented source's controller, as damage from the card
// identified by RedirectSourceID controlled by Controller ("If damage is
// prevented this way, Deflecting Palm deals that much damage to that source's
// controller.").
type PreventionShield struct {
	ID                         id.ID
	Controller                 PlayerID
	Player                     PlayerID
	PermanentID                id.ID
	SourcePermanentID          id.ID
	RedirectSourceID           id.ID
	SourceColors               []color.Color
	Amount                     int
	All                        bool
	CombatOnly                 bool
	Global                     bool
	OneShot                    bool
	RedirectToSourceController bool
	Duration                   EffectDuration
	CreatedTurn                int
	// SourceID is the card instance that created this shield, captured when the
	// shield is set up so a later effect of the same card can find its shields
	// after the creating spell has left the stack. It backs the "for each 1
	// damage prevented this way" payoff (Inkshield), where a delayed create-token
	// reads Prevented across the shields this card created.
	SourceID id.ID
	// Prevented accumulates the total combat and non-combat damage this shield
	// has prevented so far. A DynamicAmountDamagePreventedThisWay amount sums it
	// across the shields sharing the resolving card's SourceID, so a payoff that
	// scales by "damage prevented this way" reads the running tally. It is zero
	// for every shield whose card carries no such payoff.
	Prevented int
}

// ReplacementDecision records deterministic ordering for competing replacement
// or prevention effects.
type ReplacementDecision struct {
	Player       PlayerID
	Options      []string
	Selected     []int
	UsedFallback bool
}

// StateTriggerKey identifies one state-triggered ability latch (CR 603.8).
type StateTriggerKey struct {
	SourceObjectID id.ID
	SourceCardID   id.ID
	AbilityIndex   int
}

// ResolutionChoiceKind classifies a value chosen during spell or ability
// resolution (CR 608.2c, CR 609.3).
type ResolutionChoiceKind int

// Resolution choice kind values classify value-producing choices.
const (
	ResolutionChoiceNone ResolutionChoiceKind = iota
	ResolutionChoiceMana
	ResolutionChoiceCardType
	ResolutionChoicePlayer
	ResolutionChoiceCard
	// ResolutionChoiceSubtype chooses a subtype (such as a creature type) from
	// the subtypes defined for SubtypeOfType (CR 614.12, CR 205.3).
	ResolutionChoiceSubtype
	// ResolutionChoiceNumber chooses one integer in the inclusive MinNumber to
	// MaxNumber range.
	ResolutionChoiceNumber
	// ResolutionChoiceCardName chooses the name of a card matching
	// CardNameType. The result stores the public printed name rather than a card
	// instance, so it persists independently of zones and object identity.
	ResolutionChoiceCardName
)

// ResolutionChoiceColorSource identifies dynamic sources for color choice
// options.
type ResolutionChoiceColorSource int

// Resolution choice color source values identify dynamic color-choice sources.
const (
	ResolutionChoiceColorSourceStatic ResolutionChoiceColorSource = iota
	ResolutionChoiceColorSourceCommanderIdentity
	// ResolutionChoiceColorSourceFixedOrEntryChosen offers a fixed color (Colors)
	// together with the color chosen as the source permanent entered, read from
	// the stack object's seeded entry choice under EntryChoiceKey. It models the
	// composite "Add {C} or one mana of the chosen color." (the Gate/Thriving land
	// cycle).
	ResolutionChoiceColorSourceFixedOrEntryChosen
	// ResolutionChoiceColorSourceLandsProduce offers every color of mana that a
	// land matching PlayerRelation (relative to the choosing player) could
	// currently produce. It models "Add one mana of any color that a land you
	// control / an opponent controls could produce." (Reflecting Pool, Exotic
	// Orchard, Fellwar Stone). The candidate colors are recomputed from the
	// battlefield at resolution; an empty set leaves the ability unactivatable
	// (CR 605.1a).
	ResolutionChoiceColorSourceLandsProduce
	// ResolutionChoiceColorSourceLinkedExileColors offers each color of the card
	// linked to the source permanent under LinkID — the card imprinted by an
	// optional enter-the-battlefield exile from hand. It models "Add one mana of
	// any of the exiled card's colors." (Chrome Mox). The colors are recomputed
	// from the linked card at resolution (CR 106.6): a missing, declined, or
	// colorless imprint yields an empty set, leaving the ability unactivatable
	// (CR 605.1a, CR 202.2), while a multicolored imprint offers exactly its
	// colors. The link is scoped to the permanent's object identity so a
	// re-entered object has no imprint until it imprints again.
	ResolutionChoiceColorSourceLinkedExileColors
	// ResolutionChoiceColorSourceControlledPermanentColors offers every color
	// found among the colors of the permanents the choosing player controls that
	// match Selection. It models "Add one mana of any color among <permanents>
	// you control." (Mox Amber's "legendary creatures and planeswalkers you
	// control", Plaza of Heroes' "legendary permanents you control"). The
	// candidate colors are the union of the matching permanents' colors,
	// recomputed from the battlefield at resolution; a board with no matching
	// colored permanent yields an empty set and leaves the ability unactivatable
	// (CR 605.1a). Colorless ({C}) is never offered because a permanent's colors
	// are only the five colors (CR 105.2, CR 202.2).
	ResolutionChoiceColorSourceControlledPermanentColors
	// ResolutionChoiceColorSourceTriggerLandProduced offers every type of mana
	// the land that fired the resolving tapped-for-mana trigger produced on that
	// tap, read from the stack object's TriggerEvent.ProducedManaColors. It models
	// the mana-doubler body "add one mana of any type that land produced."
	// (Mirari's Wake, Zendikar Resurgent, Dictate of Karametra). The candidate
	// types are exactly the colors (including colorless) the triggering tap added;
	// an empty set leaves the trigger producing no mana (CR 605.1a).
	ResolutionChoiceColorSourceTriggerLandProduced
)

// ResolutionChoice describes a bounded value-producing choice made during
// resolution.
type ResolutionChoice struct {
	Kind ResolutionChoiceKind

	// Prompt overrides the default choice prompt.
	Prompt string

	// Player is the choosing player when UsePlayer is true; otherwise the stack
	// object's controller chooses.
	Player          PlayerID
	UsePlayer       bool
	PlayerReference *PlayerReference

	ColorSource ResolutionChoiceColorSource
	Colors      []mana.Color
	CardTypes   []types.Card
	// CardNameType restricts a ResolutionChoiceCardName to names of cards with
	// this printed card type.
	CardNameType   types.Card
	PlayerRelation PlayerRelation
	Zone           zone.Type

	// Selection constrains which permanents a dynamic color source reads. It is
	// consulted by ResolutionChoiceColorSourceControlledPermanentColors, which
	// offers the union of colors of the choosing player's permanents matching it.
	Selection *Selection

	// SubtypeOfType names the card type whose defined subtypes are the candidates
	// for a ResolutionChoiceSubtype choice, as in "choose a creature type."
	// (types.Creature). It is consulted only by ResolutionChoiceSubtype.
	SubtypeOfType types.Card

	// EntryChoiceKey names the source permanent's entry-time choice that a
	// dynamic color source reads (CR 614.12). It is consulted by
	// ResolutionChoiceColorSourceFixedOrEntryChosen.
	EntryChoiceKey ChoiceKey

	// LinkID names the linked object the choice's color source reads. It is
	// consulted by ResolutionChoiceColorSourceLinkedExileColors, which offers the
	// colors of the card linked to the source permanent under this key (the
	// imprinted exiled card).
	LinkID string

	// IncludeColorless additionally offers colorless ({C}) for a
	// ResolutionChoiceColorSourceLandsProduce choice when a matching land could
	// produce it. It distinguishes the "any type" wording (Reflecting Pool) from
	// "any color" (Exotic Orchard), which offers only colored mana.
	IncludeColorless bool
	MinNumber        int
	MaxNumber        int

	// AtRandom makes a ResolutionChoiceNumber choice uniformly at random over the
	// inclusive MinNumber..MaxNumber range using the engine RNG, with no player
	// prompt (Tibalt's Trickery's "Choose 1, 2, or 3 at random."). The chosen
	// value is published like any other number choice so a later instruction can
	// consume it. It is valid only with ResolutionChoiceNumber.
	AtRandom bool
}

// ResolutionChoiceResult stores the selected value from a ResolutionChoice.
type ResolutionChoiceResult struct {
	Kind     ResolutionChoiceKind
	Color    mana.Color
	CardType types.Card
	Subtype  types.Sub
	Player   PlayerID
	CardID   id.ID
	CardName string
	Number   int
}

// ResolutionPayment describes an optional cost that may be paid during
// resolution (CR 608.2c, CR 117.12).
type ResolutionPayment struct {
	Prompt   string
	Payer    opt.V[PlayerReference]
	ManaCost opt.V[cost.Mana]
	// DynamicGenericManaCost is a generic mana amount evaluated as the payment
	// instruction resolves. Negative values are treated as zero.
	DynamicGenericManaCost opt.V[*DynamicAmount]
	// ManaCostMultiplier repeats the fixed ManaCost by an amount evaluated as
	// the payment instruction resolves. Negative values are treated as zero.
	ManaCostMultiplier opt.V[*DynamicAmount]
	AdditionalCosts    []cost.Additional
	XValue             int
}

// ReplacementEffect is a runtime replacement effect that changes a future event
// before it happens (CR 614). This first generic slice covers zone destination
// changes and enters-the-battlefield modifiers; specialized replacement paths
// such as commander replacement and regeneration remain rules-owned.
type ReplacementEffect struct {
	ID             id.ID
	Controller     PlayerID
	SourceObjectID id.ID
	SourceCardID   id.ID
	Description    string

	Duration    EffectDuration
	CreatedTurn int

	MatchEvent EventKind

	ControllerFilter TriggerControllerFilter

	MatchFromZone bool
	FromZone      zone.Type
	MatchToZone   bool
	ToZone        zone.Type

	// Condition gates this replacement against the in-flight event.
	Condition opt.V[Condition]

	ReplaceToZone      zone.Type
	ShuffleIntoLibrary bool
	RevealSource       bool
	TokenMultiplier    int
	// TokenAddend adds a fixed number of extra tokens after TokenMultiplier is
	// applied, backing "create those tokens plus an additional <Type> token."
	// (Xorn). TokenRequiredSubtypes, when non-empty, restricts a token-creation
	// replacement to tokens carrying all of the listed subtypes (Xorn's Treasure
	// filter); an empty filter matches every created token (Doubling Season).
	// TokenRequiredTypes, when non-empty, additionally restricts the replacement
	// to tokens carrying all of the listed card types ("one or more artifact
	// tokens", Worldwalker Helm; "one or more creature tokens", Queen Allenal).
	TokenAddend           int
	TokenRequiredSubtypes []types.Sub
	TokenRequiredTypes    []types.Card
	// TokenAddendDef, when non-nil, makes the addend create TokenAddend copies of
	// this predefined token rather than copies of the triggering token (Tippy-Toe:
	// "create those tokens plus an additional Food token"). The addend tokens are
	// created directly alongside the matched tokens, so they neither re-trigger
	// this replacement nor multiply with TokenMultiplier.
	TokenAddendDef *CardDef
	// TokenReplaceDef, when non-nil, replaces each token the matched creation
	// event would create with one copy of this definition, backing the identity
	// substitution "If one or more <type> tokens would be created under your
	// control, that many <other token> are created instead." (Divine Visitation:
	// each created creature token becomes a 4/4 Angel). The would-create count is
	// preserved (one substitute per original token), so the substitution carries
	// TokenMultiplier 1 and no addend; TokenRequiredTypes / TokenRequiredSubtypes
	// restrict which created tokens it replaces. It is nil for every multiplying
	// or additive token-creation replacement.
	TokenReplaceDef *CardDef
	// SpellCopyAddend adds this many copies when the controller would copy a
	// spell one or more times. SpellCopyAdditionalMayChooseNewTargets lets the
	// controller retarget each additional copy.
	SpellCopyAddend                        int
	SpellCopyAdditionalMayChooseNewTargets bool
	CounterMultiplier                      int
	CounterAddend                          int
	MatchCounterKind                       bool
	CounterKindFilter                      counter.Kind
	// CounterRecipientSelection restricts the counter recipient to a permanent
	// whose characteristics satisfy this canonical Selection, matched through the
	// shared matchSelection so the recipient filter reads the same vocabulary as
	// targets, triggers, and cost modifiers. It carries the conjunctive
	// "creature" recipient of a typed counter-doubling replacement
	// (CounterPlacementReplacement) via RequiredTypes and the union recipient of
	// "an artifact or creature you control" (Ozolith, the Shattered Spire) via
	// RequiredTypesAny. It is nil for replacements with no recipient-type filter.
	// Recipient controller scope stays outside the Selection on
	// CounterUseRecipientController.
	CounterRecipientSelection     *Selection
	CounterRecipientAnyPermanent  bool
	CounterUseRecipientController bool
	// CounterRecipientSelf restricts the recipient to the replacement's own
	// source permanent ("If one or more +1/+1 counters would be put on Mowu, ...",
	// Mowu, Loyal Companion). When set, registration binds the replacement's
	// AffectedObjectID to the source's object ID so it matches only counters that
	// would be put on that one permanent. It is false for every group or broad
	// counter-placement replacement.
	CounterRecipientSelf bool
	// CounterRecipientControllerPlayer widens a permanent-recipient
	// counter-placement replacement's recipient union to include the
	// replacement's controller as a player ("... on a creature or planeswalker
	// you control or on yourself", Lae'zel, Vlaakith's Champion). When set, the
	// replacement also applies to counters put on a player, gated to the
	// controller by the shared CounterUseRecipientController/ControllerFilter
	// "you" scope. It is false for permanent-only recipient replacements.
	CounterRecipientControllerPlayer bool
	DamageMultiplier                 int
	DamageAddend                     int
	DamageSourceColors               []color.Color
	DamageExcludeSource              bool
	// DamageSourceTypes restricts a damage replacement to sources that have all
	// of the listed card types ("a creature you control"). DamageRecipientOpponent
	// restricts it to damage dealt to an opponent of the replacement's controller
	// or a permanent that opponent controls. DamageNoncombatOnly restricts it to
	// noncombat damage. Each is empty/false when the replacement is unrestricted.
	DamageSourceTypes                 []types.Card
	DamageRecipientOpponent           bool
	DamageRecipientOpponentPlayerOnly bool
	DamageNoncombatOnly               bool

	// DamagePreventAmount is the fixed amount a continuous static damage
	// prevention caps from each matching damage event ("prevent N of that
	// damage.", Sphere of Law, Urza's Armor). When positive the replacement is a
	// prevention rather than a multiplier/addend replacement: it reduces each
	// matching event by up to this many and emits a damage-prevented event.
	// DamageRecipientController restricts the prevention to damage dealt to the
	// replacement's controller (the player "you"); DamageSourceControllerOpponent
	// further restricts it to a source controlled by an opponent of that
	// controller ("a source an opponent controls"). Both are false/zero on every
	// multiplicative or additive damage replacement.
	DamagePreventAmount            int
	DamageRecipientController      bool
	DamageSourceControllerOpponent bool

	// DamagePreventAll prevents all of a matching damage event ("If damage would
	// be dealt to <permanent>, prevent that damage ...", Jared Carthalion,
	// Panther Habit, Anti-Venom). Unlike DamagePreventAmount it caps the whole
	// event rather than a fixed number, and it registers/matches independently of
	// the additive and multiplicative damage replacements.
	DamagePreventAll bool
	// DamagePreventedBecomesPlusOneCounters puts that many +1/+1 counters on the
	// prevention's recipient permanent after a DamagePreventAll event ("... and
	// put that many +1/+1 counters on it."). It is false for a plain prevention.
	DamagePreventedBecomesPlusOneCounters bool
	// DamagePreventedRemovesPlusOneCounter removes a single +1/+1 counter from the
	// prevention's recipient permanent after a DamagePreventAll event ("... prevent
	// that damage. Remove a +1/+1 counter from this creature.", the Phantom
	// mechanic). Exactly one counter is removed per prevented damage event,
	// independent of the damage amount; it does nothing when no +1/+1 counter is
	// present. It is false for a plain prevention.
	DamagePreventedRemovesPlusOneCounter bool
	// DamageRecipientSelf scopes a DamagePreventAll prevention to the
	// replacement's own source permanent ("If damage would be dealt to
	// <this permanent>, ..."). Registration sets AffectedObjectID to the source.
	DamageRecipientSelf bool
	// DamageRecipientAttached scopes a DamagePreventAll prevention to the permanent
	// the replacement's source is attached to ("If equipped creature would be
	// dealt damage, ...", Panther Habit). Registration sets AffectedObjectID to
	// the attached permanent.
	DamageRecipientAttached bool

	// DamageCombatOnly restricts a damage prevention to combat damage ("Prevent
	// all combat damage that would be dealt to ...", Goldbug). It is false for a
	// prevention that applies to all damage.
	DamageCombatOnly bool
	// DamageRecipientSelection scopes a DamagePreventAll prevention to a group of
	// recipient permanents matching this canonical Selection, matched through the
	// shared matchSelection so the recipient filter reads the same vocabulary as
	// targets and triggers ("attacking Humans you control"). The Selection's
	// controller relation is resolved relative to the replacement's controller. It
	// is nil for a prevention with a fixed self/attached/player recipient scope.
	DamageRecipientSelection *Selection

	// LifeGainMultiplier multiplies a single "you would gain life" event by the
	// replacement's controller before the life is gained (CR 614), backing "If
	// you would gain life, you gain twice that much life instead." (Boon
	// Reflection, Rhox Faithmender, Alhammarret's Archive). LifeGainAddend then
	// adds a fixed amount, backing "you gain that much life plus N instead."
	// (Angel of Vitality, Heron of Hope). A multiplier of zero or one with a zero
	// addend leaves life gain unchanged. Both apply only when the gaining player
	// is the replacement's controller.
	LifeGainMultiplier int
	LifeGainAddend     int

	// LifeLossMultiplier multiplies a single "would lose life" event before the
	// life is lost (CR 614), backing "they lose twice that much life instead."
	// (Bloodletter of Aclazotz). LifeLossAddend then adds a fixed amount. A
	// multiplier of zero or one with a zero addend leaves life loss unchanged.
	// LifeLossRecipientOpponent restricts the replacement to opponents of the
	// replacement's controller (false matches any player), and
	// LifeLossDuringControllerTurn restricts it to the controller's own turn.
	LifeLossMultiplier           int
	LifeLossAddend               int
	LifeLossRecipientOpponent    bool
	LifeLossDuringControllerTurn bool

	EntersTapped       bool
	EntersWithCounters []CounterPlacement

	// EntryColorChoice marks an enters-the-battlefield replacement that prompts
	// the controller to choose a color as the permanent enters (CR 614.12), such
	// as "As this artifact enters, choose a color." The chosen color is stored on
	// the permanent under EntryColorChoiceKey for later abilities to read.
	EntryColorChoice bool

	// EntryColorChoiceExclude is a single forbidden color removed from the
	// entry-time color prompt, as in "As this land enters, choose a color other
	// than white." (the Gate/Thriving land cycle). It is empty when the choice is
	// unconstrained. It is only meaningful when EntryColorChoice is true.
	EntryColorChoiceExclude mana.Color

	// EntryTypeChoice marks an enters-the-battlefield replacement that prompts the
	// controller to choose a creature type as the permanent enters (CR 614.12),
	// as in "As this creature enters, choose a creature type." The chosen subtype
	// is stored on the permanent under EntryTypeChoiceKey for later abilities to
	// read.
	EntryTypeChoice bool
	// EntryCardTypeChoice chooses among a bounded set of card types as the
	// permanent enters and stores the result under EntryCardTypeChoiceKey.
	EntryCardTypeChoice bool

	// AttachCardNameChoiceType and AttachSubtypeChoiceType mark a replacement
	// that makes persistent choices as its source becomes attached. Each nonempty
	// type requests one choice and stores it on the attachment under the
	// corresponding Attachment*ChoiceKey. Reattaching replaces prior values.
	AttachCardNameChoiceType types.Card
	AttachSubtypeChoiceType  types.Card

	// EntryDevourMultiplier marks a Devour as-enters replacement (CR 702.81) and
	// carries its per-sacrificed-creature +1/+1 counter count N. As the permanent
	// enters, its controller may sacrifice any number of other creatures they
	// control and the permanent enters with N counters for each one sacrificed.
	// It is zero for every non-Devour replacement.
	EntryDevourMultiplier int

	// EntryDevourType and EntryDevourSubtype refine a Devour replacement to a
	// typed permanent variant (CR 702.81): the controller may sacrifice any
	// number of permanents matching this card type (artifact, land) or subtype
	// (Food) instead of creatures. Both are zero for the plain creature form,
	// which sacrifices creatures.
	EntryDevourType    types.Card
	EntryDevourSubtype types.Sub

	// EntryTributeCount marks a Tribute as-enters replacement (CR 702.110) and
	// carries its +1/+1 counter count N. As the permanent enters, a chosen
	// opponent may put N counters on it; doing so sets the permanent's TributePaid
	// flag. It is zero for every non-Tribute replacement.
	EntryTributeCount int

	// EntersTappedOthers marks a continuous static enters-tapped replacement that
	// taps a group of OTHER permanents as they enter (Authority of the Consuls),
	// as opposed to the self form printed on the entering permanent. It is
	// registered into Game.ReplacementEffects while its source is on the
	// battlefield and matched against every entering permanent that satisfies
	// ControllerFilter and EntersTappedTypes.
	EntersTappedOthers bool
	// EntersUntapped clears the entering permanent's tapped state.
	EntersUntapped bool
	// EntersUntappedOthers marks a continuous group replacement rather than a
	// self replacement.
	EntersUntappedOthers bool

	// EntersTappedSelection restricts an EntersTappedOthers replacement to
	// entering permanents whose characteristics satisfy this canonical Selection,
	// matched through the shared matchSelection. Its RequiredTypesAny carries the
	// "any of these card types" recipient filter; it is nil when every entering
	// permanent is tapped ("Permanents ... enter tapped.").
	EntersTappedSelection *Selection

	// EntersWithCountersOthers marks a continuous static enters-with-counters
	// replacement that adds the EntersWithCounters placements to a group of OTHER
	// permanents as they enter ("Each other creature you control enters with an
	// additional vigilance counter on it." — Tayam, Luminous Enigma), as opposed
	// to the self form printed on the entering permanent. It is registered into
	// Game.ReplacementEffects while its source is on the battlefield and matched
	// against every entering permanent that satisfies EntersWithCountersRecipient.
	EntersWithCountersOthers bool

	// EntersWithCountersRecipient restricts an EntersWithCountersOthers
	// replacement to entering permanents matched by this selection (controller
	// scope, card types, subtypes, and source exclusion for the "other" form). It
	// is nil for the self form.
	EntersWithCountersRecipient *Selection

	// EntersBecomesCharacteristic marks a continuous static group ETB
	// characteristic replacement that changes the characteristics of a group of
	// permanents (including the source itself if it qualifies) as they enter (CR
	// 614), as in "As a historic permanent you control enters, it becomes a 7/7
	// Dinosaur creature in addition to its other types." (Displaced Dinosaurs). It
	// is registered into Game.ReplacementEffects while its source is on the
	// battlefield and matched against every entering permanent that satisfies
	// ControllerFilter and EntersBecomesSelection. The change is applied through
	// layer-appropriate continuous effects tied to the entering permanent, so it
	// persists even if the source later leaves the battlefield.
	EntersBecomesCharacteristic bool

	// EntersBecomesSelection restricts an EntersBecomesCharacteristic replacement
	// to entering permanents whose characteristics satisfy this canonical
	// Selection (the "historic permanent" filter is the AnyOf of artifact,
	// legendary, and Saga). It is nil when every entering permanent qualifies.
	EntersBecomesSelection *Selection

	// EntersBecomesAddTypes, EntersBecomesAddSubtypes, and EntersBecomesAddColors
	// list the card types, creature subtypes, and colors an
	// EntersBecomesCharacteristic entrant gains "in addition to its other types".
	// They are only consulted when EntersBecomesCharacteristic is true.
	EntersBecomesAddTypes    []types.Card
	EntersBecomesAddSubtypes []types.Sub
	EntersBecomesAddColors   []color.Color

	// EntersBecomesBasePower and EntersBecomesBaseToughness set an
	// EntersBecomesCharacteristic entrant's base power and toughness to a fixed
	// size (Displaced Dinosaurs' 7/7). They are set together and only consulted
	// when EntersBecomesCharacteristic is true.
	EntersBecomesBasePower     opt.V[int]
	EntersBecomesBaseToughness opt.V[int]

	// CreateOneOfEachTokens replaces the creation of a token whose name matches
	// one of these definitions with the creation of one of each listed token
	// (Academy Manufactor: "If you would create a Clue, Food, or Treasure token,
	// instead create one of each."). It is empty for every other replacement.
	CreateOneOfEachTokens []*CardDef

	// DrawFromEmptyLibraryWins replaces an attempt to draw a card from an empty
	// library by the controller with that controller winning the game (CR 104.2,
	// CR 614). It backs "If you would draw a card while your library has no cards
	// in it, you win the game instead." (Laboratory Maniac, Jace, Wielder of
	// Mysteries). It is registered while its source is on the battlefield and
	// consulted when the controller would otherwise lose to a failed draw.
	DrawFromEmptyLibraryWins bool

	// EntersAsCopy marks a self enters-the-battlefield replacement that has the
	// permanent enter as a copy of another permanent chosen as it enters (CR 706,
	// CR 614), as in "You may have this creature enter the battlefield as a copy
	// of any creature on the battlefield." (Clone). The controller chooses one
	// permanent matching EntersAsCopySelection; the entering permanent's copiable
	// values are overlaid with the chosen permanent's via a layer-1 continuous
	// effect that lasts as long as the copy is on the battlefield.
	EntersAsCopy bool

	// EntersAsCopyOptional marks the "You may have ..." form of an EntersAsCopy
	// replacement: the controller is first asked whether to copy at all and the
	// permanent enters as itself if they decline. It is false for the mandatory
	// "this creature enters as a copy of ..." form.
	EntersAsCopyOptional bool

	// EntersAsCopySelection restricts which permanents may be copied by an
	// EntersAsCopy replacement ("any creature on the battlefield", "any nonland
	// permanent on the battlefield", "a creature you control"). It is nil for
	// every other replacement and only consulted when EntersAsCopy is true.
	EntersAsCopySelection *Selection

	// EntersAsCopyNotLegendary applies the "except it isn't legendary" copiable
	// rider (CR 706.9c) by dropping the legendary supertype from the copied
	// values. It is only meaningful when EntersAsCopy is true.
	EntersAsCopyNotLegendary bool

	// EntersAsCopyRetainName keeps the entering permanent's own defined name
	// instead of the copied permanent's name as a copiable exception.
	EntersAsCopyRetainName bool

	// EntersAsCopyAddOtherAbilities adds the entering permanent's other defined
	// abilities to the copied values, excluding every enters-as-copy replacement
	// so the exception does not recursively add itself.
	EntersAsCopyAddOtherAbilities bool

	// EntersAsCopyAddTypes applies the "except it's an <type> in addition to its
	// other types" copiable rider (Phyrexian Metamorph) by adding these card
	// types to the copied values. It is empty for every other replacement and
	// only consulted when EntersAsCopy is true.
	EntersAsCopyAddTypes []types.Card

	// EntersAsCopyAddSubtypes applies the "except it's a <subtype> in addition to
	// its other types" copiable rider (Mockingbird's Bird, Synth Infiltrator's
	// Synth) by adding these subtypes to the copied values. It is empty for every
	// other replacement and only consulted when EntersAsCopy is true.
	EntersAsCopyAddSubtypes []types.Sub

	// EntersAsCopyConditionalCounters applies the conditional copiable counter
	// riders of an enters-as-copy replacement, placing additional counters on the
	// copy based on the copied card's types (Spark Double: "+1/+1 counter if it's
	// a creature" and "loyalty counter if it's a planeswalker"). It is empty for
	// every other replacement and only consulted when EntersAsCopy is true.
	EntersAsCopyConditionalCounters []ConditionalCounterPlacement

	// EntersAsCopyUntilEndOfTurn scopes an enters-as-copy replacement's copy
	// effect to end of turn rather than as long as the permanent stays on the
	// battlefield (Cursed Mirror's "become a copy ... until end of turn"). It is
	// only consulted when EntersAsCopy is true.
	EntersAsCopyUntilEndOfTurn bool

	// EntersAsCopyAddKeywords lists keywords granted to the copy by the "except it
	// has <keyword>" rider of an enters-as-copy replacement (Cursed Mirror's
	// haste). It is empty for every other replacement and only consulted when
	// EntersAsCopy is true.
	EntersAsCopyAddKeywords []Keyword

	// EntersAsCopyTapped taps the permanent as it enters the battlefield as its
	// chosen copy (Vesuva's "enter tapped as a copy of any land"). It is only
	// consulted when EntersAsCopy is true, and applies after the optional copy
	// choice is confirmed so a declined copy enters untapped.
	EntersAsCopyTapped bool

	// EntersAsCopyBasePower and EntersAsCopyBaseToughness apply the "except it's
	// N/N" copiable rider (Quicksilver Gargantuan's "except it's 7/7") by
	// overriding the copied values' power and toughness with a fixed size (CR
	// 706.2). They are set together and only consulted when EntersAsCopy is true.
	EntersAsCopyBasePower     opt.V[int]
	EntersAsCopyBaseToughness opt.V[int]

	// EntersAsCopyMaxManaValueFromManaSpent restricts the permanents that may be
	// copied to those whose mana value is at most the amount of mana spent to
	// cast this permanent (Mockingbird's "with mana value less than or equal to
	// the amount of mana spent to cast this creature"). A permanent that did not
	// enter from a cast spell spent no mana, so the bound is zero. It is only
	// consulted when EntersAsCopy is true.
	EntersAsCopyMaxManaValueFromManaSpent bool

	// EntersAsCopyAddAbilities lists abilities granted to the copy by the "except
	// it has \"<quoted ability>\"" copiable rider (Estrid's Invocation's granted
	// upkeep self-blink ability). They are appended to the copy's copiable
	// abilities as it enters, so they themselves become copiable (CR 706.2). It is
	// nil for every other replacement and only consulted when EntersAsCopy is true.
	EntersAsCopyAddAbilities []Ability

	// DrawCardMultiplier replaces a single "draw a card" event by the controller
	// with drawing this many cards instead (CR 614). It backs the draw-doubling
	// replacement "If you would draw a card, draw two cards instead." A value of
	// zero or one leaves draws unchanged. It is registered while its source is on
	// the battlefield.
	DrawCardMultiplier int

	// DrawCardExceptFirstInDrawStep exempts the controller's first draw in each
	// of their own draw steps from DrawCardMultiplier ("If you would draw a card
	// except the first one you draw in each of your draw steps, draw two cards
	// instead.", Teferi's Ageless Insight). It is only meaningful when
	// DrawCardMultiplier is greater than one.
	DrawCardExceptFirstInDrawStep bool

	// DrawCardDigLook replaces a single "draw a card" event by the controller
	// with looking at the top DrawCardDigLook cards of their library, putting
	// DrawCardDigTake of them into their hand, and routing the rest to
	// DrawCardDigRemainder (CR 614). It backs "If you would draw a card, instead
	// look at the top three cards of your library, then put one into your hand
	// and the rest into your graveyard." (Underrealm Lich). A value of zero
	// leaves draws unchanged. It is registered while its source is on the
	// battlefield and consulted each time the controller would draw.
	DrawCardDigLook int

	// DrawCardDigTake is the number of looked-at cards a DrawCardDigLook
	// replacement puts into the controller's hand. It is only meaningful when
	// DrawCardDigLook is greater than zero.
	DrawCardDigTake int

	// DrawCardDigRemainder is the destination of the un-taken cards of a
	// DrawCardDigLook replacement: the controller's graveyard (the default) or
	// the bottom of their library. It is only meaningful when DrawCardDigLook is
	// greater than zero.
	DrawCardDigRemainder DigRemainder

	// ContinuousZoneRedirect marks a continuous static replacement that redirects
	// a card (or permanent) headed for a graveyard to a different zone (CR 614),
	// as on "If a card would be put into a graveyard from anywhere, exile it
	// instead." (Leyline of the Void, Samurai of the Pale Curtain). It is
	// registered into Game.ReplacementEffects while its source is on the
	// battlefield and matched against every card the event moves, using
	// RedirectOwnerFilter and RedirectTypeFilter. The self form printed on a
	// single card ("If THIS would be put into a graveyard, ...") is not marked
	// and is handled by the static self-zone replacement path instead.
	ContinuousZoneRedirect bool

	// RedirectOwnerFilter restricts a ContinuousZoneRedirect replacement by the
	// owner of the moving card relative to the replacement's controller: You
	// watches the controller's own graveyard, Opponent an opponent's, and Any
	// every player's. It is only meaningful when ContinuousZoneRedirect is true.
	RedirectOwnerFilter TriggerControllerFilter

	// RedirectTypeFilter restricts a ContinuousZoneRedirect replacement to moving
	// cards that have any of these card types ("an instant or sorcery card").
	// It is empty when every card is redirected. It is only meaningful when
	// ContinuousZoneRedirect is true.
	RedirectTypeFilter []types.Card

	// RedirectControlFilter restricts a ContinuousZoneRedirect replacement by the
	// controller of the dying permanent relative to the replacement's controller,
	// for "would die" forms ("If a creature an opponent controls would die, exile
	// it instead."): You watches the controller's own permanents, Opponent an
	// opponent's, and Any every controller's. It is only meaningful when
	// ContinuousZoneRedirect is true, and applies in addition to RedirectOwnerFilter.
	RedirectControlFilter TriggerControllerFilter

	// RedirectCounter is the named counter placed on a card a
	// ContinuousZoneRedirect exiles ("instead exile it with a void counter on
	// it." — Dauthi Voidwalker). It is only meaningful when
	// ContinuousZoneRedirect is true and ReplaceToZone is zone.Exile; the
	// runtime places one counter of this kind on the redirected card once it
	// reaches exile, mirroring the MoveCard.Counter exile rider. It is absent
	// for redirects that place no counter (Leyline of the Void).
	RedirectCounter opt.V[counter.Kind]

	// AffectedObjectID restricts the replacement to events about a single
	// permanent identified by its object ID. When non-zero, the replacement
	// matches only an event whose moving permanent is exactly this object,
	// backing a dynamically created replacement bound to one specific permanent
	// ("If it would leave the battlefield, exile it instead of putting it
	// anywhere else." applied to a just-reanimated creature — Whip of Erebos).
	// It is zero for every printed or unscoped replacement, which match by their
	// other filters alone.
	AffectedObjectID id.ID

	// AffectedObjectMustBeCreature further restricts an AffectedObjectID-bound
	// zone-change redirect to events whose moving permanent is a creature. It
	// backs the "a creature dealt damage this way would die this turn, exile it
	// instead." burn rider (Yamabushi's Flame, Demonfire) bound to an "any
	// target" spell's single target: the redirect must not exile a player or
	// planeswalker that the same spell killed. It is false for the "that
	// creature [or planeswalker]" rider, which targets a creature (or a
	// deliberately-included planeswalker) and needs no creature gate.
	AffectedObjectMustBeCreature bool

	// AffectedCardID restricts the replacement to events about the permanent
	// created when a single card instance enters the battlefield, identified by
	// the card's stable instance ID. A permanent spell gains a fresh object ID as
	// it resolves onto the battlefield, so an object-ID binding taken from the
	// stack object cannot match the entering permanent; the card instance ID is
	// preserved across the stack-to-battlefield move and identifies it. It backs a
	// one-shot replacement created for a future-cast spell ("When you next cast a
	// creature spell this turn, that creature enters with an additional +1/+1
	// counter on it." — Summon: Fenrir chapter II). It is zero for every
	// replacement that is not bound to one specific card instance.
	AffectedCardID id.ID
}

// EntryTypeChoiceKey is the ChoiceKey under which an entry-time creature-type
// choice is stored on a Permanent's EntryChoices map. Abilities referencing "the
// chosen type" read the result from this key.
const EntryTypeChoiceKey = ChoiceKey("oracle-entry-type")

// EntryCardTypeChoiceKey stores an entry-time card-type choice.
const EntryCardTypeChoiceKey = ChoiceKey("oracle-entry-card-type")

// EntryColorChoiceKey is the ChoiceKey under which an entry-time color choice is
// stored on a Permanent's EntryChoices map. Mana abilities that add "one mana of
// the chosen color" read the result from this key.
const EntryColorChoiceKey = ChoiceKey("oracle-entry-color")

// AttachmentCardNameChoiceKey stores the most recent card name chosen as an
// attachment became attached.
const AttachmentCardNameChoiceKey = ChoiceKey("oracle-attachment-card-name")

// AttachmentSubtypeChoiceKey stores the most recent subtype chosen as an
// attachment became attached.
const AttachmentSubtypeChoiceKey = ChoiceKey("oracle-attachment-subtype")

// SpellChosenTypeChoiceKey is the ChoiceKey under which a resolution-time
// creature-type choice made by a resolving spell or ability is published (a
// Choose instruction's PublishChoice). Later effects in the same resolution that
// reference "of that type" read the chosen subtype from this key, as in "Choose a
// creature type. Draw a card for each permanent you control of that type."
// (Distant Melody).
const SpellChosenTypeChoiceKey = ChoiceKey("oracle-chosen-type")
