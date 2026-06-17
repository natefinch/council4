# Card-Support Backlog

Every eligible Scryfall corpus card is evaluated with two signals and routed to the layer that blocks it:

- **Parser signal** (parser-only): `cardgen.ParseCardFaces` + `parser.DocumentCoverage` — is the card parser-complete, and which uncovered components remain?
- **Lowering signal** (full compile): compilecards' canonical report — did the card generate, and if not, which distinct diagnostic summaries blocked lowering? compilecards is the authority; an independent per-card recompile reconciles against it.

This produces two ranked, actionable queues. Regenerate with `mage cardBacklog`.

## Headline

- Eligible cards: 31835
- Supported (generated): 6638
- Parser-complete: 15225
- **Lowering backlog** (parser-complete, not generated): 8598
- **Parser backlog** (not parser-complete, not generated): 16599

Partition check: 6638 supported + 8598 lowering-backlog + 16599 parser-backlog = 31835 eligible. ✓

11 generated cards are not parser-complete. The lowerer fully generates them, but the parser-coverage harness does not span all their must-cover tokens (the residue tracked in `parser-coverage.md`). They are counted as **supported**, not routed to either backlog queue:

- Death Frenzy
- Hollowhenge Overlord
- Crescent Island Temple
- The Royal Scions
- Unagi's Spray
- Faebloom Trick
- Bladewing, Deathless Tyrant
- Tribune of Rot
- Rockslide Sorcerer
- Umara Mystic
- Cadira, Caller of the Small

### Reconciliation guard

Generated membership is read from compilecards' canonical report. An independent per-card recompile cross-checks it; the run fails if they diverge.

- Authoritative generated (compilecards report): 6638
- Independent per-card recompile generated: 6638
- Divergences: 0 — the two pipelines agree. ✓

## Lowering queue

Parser-complete cards that do not yet lower, bucketed by distinct lowering diagnostic summary and ranked by affected-card count. Parsing is already done for these cards, so they are the lowest-risk backlog: this is `unsupported-reasons.md` restricted to the parser-complete subset.

| Rank | Reason | Affected (parser-complete) cards | Sole blockers | Example cards |
| --- | --- | --- | --- | --- |
| 1 | unsupported ordered effect sequence | 1666 | 1406 | Mind Extraction; Hunt the Hunter; Blur; Fear of Falling; Talisman of Progress |
| 2 | unsupported activation cost | 556 | 355 | Greta, Sweettooth Scourge; Ruthless Knave; Thunderherd Migration; Harvest Pyre; Krovikan Sorcerer |
| 3 | unsupported static ability | 510 | 231 | Nissa, Worldsoul Speaker; Static Orb; Waterknot; Marang River Prowler; Kykar, Zephyr Awakener |
| 4 | unsupported damage spell | 500 | 408 | Torrent of Fire; Arc Mage; Cinder Elemental; Armed Response; Fire Covenant |
| 5 | unsupported static declaration group | 494 | 368 | Weakstone; Dungeon Delver; Magma Sliver; Chamber of Manipulation; Etchings of the Chosen |
| 6 | unsupported ability content | 487 | 318 | Ulvenwald Captive // Ulvenwald Abomination; Scorching Missile; Consuming Sinkhole; Vihaan, Goldwaker; Call Forth the Tempest |
| 7 | unsupported token creation | 465 | 316 | Greta, Sweettooth Scourge; Ruthless Knave; Seedship Agrarian; Nimble Thopterist; Summoning Station |
| 8 | unsupported counter placement | 440 | 275 | Sword-Swallowing Seraph; Greater Werewolf; Ent-Draught Basin; Misinformation; Drill Too Deep |
| 9 | unsupported return spell | 377 | 303 | Palinchron; Pharika's Mender; Selesnya Sanctuary; Dragon Fangs; Salvage Scuttler |
| 10 | unsupported optional effect | 344 | 277 | Courageous Outrider; Dazzling Sphinx; Mindclaw Shaman; Sword of Light and Shadow; Ponder |
| 11 | unsupported temporary keyword spell | 311 | 266 | Unyielding Krumar; Bladed Sentinel; Shadowcloak Vampire; Jareth, Leonine Titan; Viashino Lashclaw |
| 12 | unsupported static declaration operation | 279 | 199 | Food Fight; Ghostly Touch; Wingrattle Scarecrow; Executioner's Hood; Cryptolith Rite |
| 13 | unsupported destroy spell | 253 | 218 | Stand Up for Yourself; Dakmor Lancer; Rock Soldiers; Summon: Primal Odin; Unliving Psychopath |
| 14 | unsupported power/toughness spell | 241 | 172 | Nissa, Worldsoul Speaker; Park Bleater; Zariel, Archduke of Avernus; Elspeth, Sun's Champion; Dawnhart Wardens |
| 15 | unsupported enters-tapped replacement | 236 | 92 | Etchings of the Chosen; Gravetiller Wurm; Sedge Sliver; Witch Enchanter // Witch-Blessed Meadow; Sol Grail |
| 16 | unsupported life spell | 229 | 186 | South Wind Avatar; Shizo, Death's Storehouse; Summon: Primal Odin; Miren, the Moaning Well; Mourning Thrull |
| 17 | unsupported exile spell | 211 | 133 | Ravnica at War; Disposal Mummy; Crypt Creeper; Reaver Ambush; Baffling End |
| 18 | unsupported search effect | 203 | 162 | Blighted Woodland; Dragonstorm Forecaster; Ramosian Commander; Avatar of Growth; Misty Rainforest |
| 19 | unsupported permanent zone-change trigger effect | 169 | 157 | Squadron Hawk; Solemn Simulacrum; Screaming Seahawk; Remembrance; Ecologist's Terrarium |
| 20 | unsupported regenerate spell | 134 | 106 | Patchwork Gnomes; Darkling Stalker; Ranger en-Vec; Troll Ascetic; Rusted Slasher |
| 21 | unsupported ability word | 132 | 84 | Dawnbringer Cleric; Bloodthorn Flail; Astarion, the Decadent; Solar Tide; Maha, Its Feathers Night |
| 22 | unsupported Oracle construct | 119 | 0 | Marang River Prowler; Dawnbringer Cleric; Kykar, Zephyr Awakener; Astarion, the Decadent; Irreverent Revelers |
| 23 | unsupported activation ability word | 115 | 93 | Half-Elf Monk; Zulaport Chainmage; Jiwari, the Earth Aflame; Illuminor Szeras; Bagel and Schmear |
| 24 | unsupported multiple spell abilities | 111 | 106 | Instill Infection; Refocus; Drag Under; Deadly Visit; Hard Evidence |
| 25 | unsupported tap spell | 109 | 73 | Waterknot; Coeurl; Torrent Elemental; Locked in the Cemetery; Freed from the Real |
| 26 | unsupported unknown ability | 102 | 0 | Dawnbringer Cleric; Kykar, Zephyr Awakener; Astarion, the Decadent; Irreverent Revelers; Fangkeeper's Familiar |
| 27 | unsupported sacrifice spell | 99 | 72 | Puppet Conjurer; Stenchskipper; Kibo, Uktabi Prince; Anowon, the Ruin Sage; Yukora, the Prisoner |
| 28 | unsupported activation references | 99 | 71 | Planebound Accomplice; Puresight Merrow; Titans' Nest; Spurnmage Advocate; Shackles |
| 29 | unsupported untap spell | 98 | 57 | Palinchron; Summoning Station; Freed from the Real; Battered Golem; Peregrine Drake |
| 30 | unsupported mana symbol | 95 | 56 | Sol Grail; Pit of Offerings; Cabal Stronghold; Sacrifice; Command Tower |
| 31 | unsupported discard spell | 94 | 76 | Tourach, Dread Cantor; Black Cat; Chilling Apparition; Tormented Thoughts; Urborg Mindsucker |
| 32 | unsupported triggered ability effect | 86 | 69 | Trapjaw Tyrant; Hunting Cheetah; Marina Vendrell's Grimoire; Displacer Kitten; Centaur Rootcaster |
| 33 | unsupported draw spell | 84 | 55 | Theft of Dreams; Master of the Feast; Friendly Teddy; Marketback Walker; Marina Vendrell's Grimoire |
| 34 | unsupported gain-control spell | 73 | 51 | Unwilling Recruit; Donate; Slave of Bolas; Legacy's Allure; Dominating Vampire |
| 35 | unsupported counter spell | 68 | 63 | Spell Blast; Lifeforce; Drown in the Loch; Disdainful Stroke; Vigilant Martyr |
| 36 | unsupported type line | 60 | 59 | Playable Delusionary Hydra; Notorious Sliver War; City's Blessing // Elemental; Demonic Tourist Laser; Night Brushwagg Ringmaster |
| 37 | unsupported mixed keyword ability | 60 | 35 | Kykar, Zephyr Awakener; Chief Engineer; Irreverent Revelers; Sky Tether; Arbalest Engineers |
| 38 | unsupported enters-with-counters replacement | 52 | 24 | Flycatcher Giraffid; Marketback Walker; Malefic Scythe; Avatar of the Resolute; Pentavus |
| 39 | unsupported mill spell | 49 | 41 | Mesmeric Orb; Flint Golem; Persistent Petitioners; Sibsig Host; Towering-Wave Mystic |
| 40 | unsupported phase/step trigger phrase effect | 48 | 39 | Savior of the Small; Alesha, Who Laughs at Fate; Tidal Force; Monastery Siege; Cautious Survivor |
| 41 | unsupported activation condition | 35 | 32 | Wizard Replica; Fabled Passage; Martyr of Frost; Judge's Familiar; Patron Wizard |
| 42 | unsupported parameterized keyword | 28 | 15 | Proven Combatant; Eldrazi Ravager; Jarl of the Forsaken; Shepherd of the Cosmos; Steadfast Sentinel |
| 43 | unsupported group power/toughness spell | 27 | 18 | Adventuring Gear; Keldon Mantle; Flowstone Embrace; Fiery Mantle; Planar Despair |
| 44 | unsupported manifest spell | 22 | 20 | Merfolk Observer; Orcish Spy; Dewdrop Spy; They Came from the Pipes; Gitaxian Probe |
| 45 | unsupported card layout | 20 | 20 | Nezumi Graverobber // Nighteyes the Desecrator; Faithful Squire // Kaiso, Memory of Loyalty; Jushi Apprentice // Tomoya the Revealer; Cunning Bandit // Azamuki, Treachery Incarnate; Nezumi Shortfang // Stabwhisker the Odious |
| 46 | unsupported fight spell | 17 | 10 | Pheres-Band Brawler; Clash of Titans; Wicked Wolf; Scab-Clan Giant; Gargos, Vicious Watcher |
| 47 | unsupported permanent zone-change trigger | 15 | 9 | Marina Vendrell's Grimoire; Bringer of the Last Gift; Sigardian Savior; Zacama, Primal Calamity; The One Ring |
| 48 | incomplete executable lowering | 14 | 11 | Swooping Protector; Hedron Crawler; Crumbling Vestige; Warden of Geometries; Wingshield Agent |
| 49 | unsupported mana effect | 14 | 11 | City of Shadows; Blinkmoth Urn; Pristine Talisman; Conduit of Storms // Conduit of Emrakul; Black Market |
| 50 | unsupported Enchant ability | 12 | 6 | Aura Graft; Robe of Mirrors; Tallowisp; Fear; Enfeeblement |
| 51 | unsupported triggered ability | 11 | 10 | Alaborn Zealot; Cinder Wall; Wall of Junk; Tephraderm; Elder Land Wurm |
| 52 | validation failed: oracle-without-abilities | 7 | 7 | Mishra's Warform; Icehide Golem; Cyberman; Forest Dryad; Morph |
| 53 | unsupported static declaration duration | 7 | 5 | Roller Coaster; Deathleaper, Terror Weapon; Protean Raider; Kiddie Coaster; Sakashima's Protege |
| 54 | unsupported delayed effect | 7 | 4 | Rienne, Angel of Rebirth; Giant Caterpillar; Library of Lat-Nam; Tiana, Ship's Caretaker; Resurrection Orb |
| 55 | unsupported loyalty ability | 7 | 0 | Kasmina, Enigma Sage; Sorin, Grim Nemesis; Chandra, Flamecaller; Tezzeret, Cruel Captain; Chandra, Torch of Defiance |
| 56 | unsupported explore spell | 6 | 5 | Jadelight Spelunker; Hakbal of the Surging Soul; Map; Seeker of Sunlight; Tomb Robber |
| 57 | unsupported optional replacement effect | 4 | 4 | Callidus Assassin; Cursed Mirror; Arsenal Thresher; Altered Ego |
| 58 | unsupported static declaration shell | 4 | 2 | Inquisitor Greyfax; Bloodcrusher of Khorne; Vexilus Praetor; Frostcliff Siege |
| 59 | unsupported counter-placement replacement | 3 | 1 | Michelangelo, Weirdness to 11; Hardened Scales; Conclave Mentor |
| 60 | unsupported keyword ability | 3 | 1 | Underworld Breach; The Master of Keys; Sami, Wildcat Captain |

## Parser queue

Cards that are not parser-complete (and do not lower), bucketed by owning component family and normalized uncovered-span cluster, ranked by occurrence. This is the grammar-recognition backlog.

| Rank | Component | Cluster | Count | Example cards |
| --- | --- | --- | --- | --- |
| 1 | condition | if you do | 1028 | Keldon Raider; Witch's Mark; Duplicity; Minion Reflector; Brawl-Bash Ogre |
| 2 | condition | if able | 170 | Impetuous Devils; The Foretold Soldier; Nacatl Hunt-Pride; Culling Mark; Legion Warboss |
| 3 | effect | you may choose new targets for the copy. | 123 | Verrak, Warped Sengir; Abstruse Archaic; League Guildmage; Psychic Rebuttal; Melek, Izzet Paragon |
| 4 | effect | this ability triggers only once each turn. | 103 | Fang, Fearless l'Cie; Twilight Diviner; Nanoform Sentinel; Cloaked Cadet; G'raha Tia |
| 5 | condition | if this spell was kicked | 99 | Strength of Night; Goblin Barrage; Colossal Growth; Overload; Stall for Time |
| 6 | effect | crew N (tap any number of creatures you control with total power N or more: this vehicle becomes an artifact creature until end of turn.) | 98 | War Balloon; Flywheel Racer; Mukotai Soulripper; Skybox Ferry; Sidequest: Card Collection // Magicked Card |
| 7 | effect | it's still a land. | 92 | Llanowar Loamspeaker; Hall of Storm Giants; Restless Vinestalk; Vastwood Animist; Faerie Conclave |
| 8 | trigger | when you cast this spell | 85 | Empyrial Storm; The Fourteenth Doctor; Ulamog, the Ceaseless Hunger; Temporal Extortion; Malicious Affliction |
| 9 | effect | it can't be regenerated. | 82 | Polymorph; Pongify; Rapid Hybridization; Phage the Untouchable; Pillage |
| 10 | effect | you may pay{1}. | 80 | Ruthless Sniper; Smolder Initiate; Wooden Sphere; Rebellion of the Flamekin; Azorius Aethermage |
| 11 | effect | level N | 68 | Stormchaser's Talent; Builder's Talent; Leader's Talent; Sorcerer Class; Fortune Teller's Talent |
| 12 | effect | crew N | 67 | Mindlink Mech; Mighty Servant of Leuk-o; _____ _____ Rocketship; The Lunar Whale; Rocketeer Boostbuggy |
| 13 | trigger | whenever this creature enters or attacks | 63 | Omnivorous Flytrap; Graveyard Trespasser // Graveyard Glutton; Inferno Titan; Sigarda's Vanguard; Cemetery Illuminator |
| 14 | effect | changeling (this card is every creature type.) | 61 | Barkform Harvester; Guardian Gladewalker; Wings of Velis Vel; Firdoch Core; Mirror Entity |
| 15 | condition | if you search your library this way | 57 | Vraska's Scorn; Ashiok's Forerunner; Niambi, Faithful Healer; Claim Jumper; Fang-Druid Summoner |
| 16 | effect | prevent the next N damage that would be dealt to any target this turn. | 56 | Heal; Master Apothecary; Barrenton Medic; Rakalite; Militant Monk |
| 17 | effect | partner (you can have two commanders if both have partner.) | 55 | Ghost of Ramirez DePietro; Silas Renn, Seeker Adept; Krark, the Thumbless; Vial Smasher the Fierce; Francisco, Fowl Marauder |
| 18 | effect | enchant creature you control | 52 | One with the Kami; Endless Evil; Inferno Fist; Pitiless Fists; Dying Wish |
| 19 | trigger | when you cycle this card | 49 | Deem Worthy; Rampaging War Mammoth; Shefet Monitor; Agonasaur Rex; Windcaller Aven |
| 20 | trigger | whenever you cast a spell that targets this creature | 46 | Akroan Line Breaker; Lagonna-Band Trailblazer; Hero of Iroas; War-Wing Siren; Triton Cavalry |
| 21 | effect | target creature can't block this turn. | 45 | Nacatl Hunt-Pride; Mardu Roughrider; Bola Warrior; Unstoppable Ogre; Goblin Shortcutter |
| 22 | trigger | whenever you cast your second spell each turn | 45 | Sunstar Lightsmith; Kraum, Violent Cacophony; Monk of the Open Hand; Illvoi Operative; Cori Mountain Stalwart |
| 23 | effect | flip a coin. | 44 | Goblin Lyre; Ydwen Efreet; Ral, Monsoon Mage // Ral, Leyline Prodigy; Krark, the Thumbless; Risky Move |
| 24 | effect | you may play that card this turn. | 41 | Party Thrasher; Professional Face-Breaker; Geistflame Reservoir; Dark-Dweller Oracle; Tablet of Discovery |
| 25 | condition | if it's a land card | 40 | Lantern of Revealing; Unexpected Results; Traveling Botanist; Countryside Crusher; Raiders' Karve |
| 26 | effect | the ring tempts you. | 40 | Uruk-hai Berserker; Horses of the Bruinen; Fiery Inscription; Relentless Rohirrim; Shortcut to Mushrooms |
| 27 | effect | you become the monarch. | 40 | Custodi Lich; Court of Ire; Palace Jailer; Grave Venerations; Court of Garenbrig |
| 28 | effect | they can't be regenerated. | 39 | Plague Wind; Wave of Terror; Wrath of God; Obliterate; Tsabo's Decree |
| 29 | effect | you may pay{2}. | 39 | Minion Reflector; Esoteric Duplicator; Kavaron Harrier; Unassuming Sage; Terra, Herald of Hope |
| 30 | condition | if you lose the flip | 38 | Goblin Bomb; Goblin Lyre; Ydwen Efreet; Ral, Monsoon Mage // Ral, Leyline Prodigy; Krark, the Thumbless |
| 31 | effect | activate only during your turn. | 38 | Ruthless Waterbender; Humble Defector; Yes Man, Personal Securitron; Hoofprints of the Stag; June, Bounty Hunter |
| 32 | effect | swampwalk (this creature can't be blocked as long as defending player controls a swamp.) | 38 | Whispering Shade; Dirtwater Wraith; Slithery Stalker; Lost Soul; Quag Vampires |
| 33 | trigger | whenever this creature or another ally you control enters | 38 | Kazandu Blademaster; Murasa Pyromancer; Grovetender Druids; Tuktuk Grunts; Firemantle Mage |
| 34 | condition | if this creature was kicked | 37 | Kavu Primarch; Faerie Squadron; Aether Figment; Skyclave Shade; Skyclave Sentinel |
| 35 | effect | fear (this creature can't be blocked except by artifact creatures and/or black creatures.) | 37 | Guiltfeeder; Commander Greven il-Vec; Dread; Lingering Tormentor; Undercity Shade |
| 36 | effect | start your engines ! (if you have no speed, it starts at 1. it increases once on each of your turns when an opponent loses life. max speed is 4.) | 37 | Momentum Breaker; Point the Way; Lightwheel Enhancements; Burnout Bashtronaut; Goblin Surveyor |
| 37 | effect | shadow (this creature can block or be blocked by only creatures with shadow.) | 36 | Dauthi Horror; Dauthi Mercenary; Thalakos Seer; Augur il-Vec; Soltari Guerrillas |
| 38 | condition | if you win the flip | 35 | Goblin Bomb; Goblin Lyre; Ral, Monsoon Mage // Ral, Leyline Prodigy; Krark, the Thumbless; Plasma Caster |
| 39 | effect | affinity for artifacts (this spell costs{1}less to cast for each artifact you control.) | 35 | Refurbished Familiar; Into Thin Air; Assert Authority; Valkyrie Aerial Unit; Furnace Dragon |
| 40 | effect | choose one. | 35 | Wail of the Forgotten; Jeska's Will; See Double; Inscription of Insight; Will of the Abzan |
| 41 | effect | daybound (if a player casts no spells during their own turn, it becomes night next turn.) | 34 | Graveyard Trespasser // Graveyard Glutton; Shady Traveler // Stalking Predator; Tovolar's Huntmaster // Tovolar's Packleader; Oakshade Stalker // Moonlit Ambusher; Brutal Cathar // Moonrage Brute |
| 42 | effect | fuse (you may cast one or both halves of this card from your hand.) | 34 | Breaking // Entering; Flesh // Blood; Protect // Serve; Turn // Burn; Ready // Willing |
| 43 | effect | islandwalk (this creature can't be blocked as long as defending player controls an island.) | 34 | Goblin Flotilla; Merrow Harbinger; Pale Bears; Stonybrook Banneret; Stormtide Leviathan |
| 44 | effect | nightbound (if a player casts at least two spells during their own turn, it becomes day next turn.) | 34 | Graveyard Trespasser // Graveyard Glutton; Shady Traveler // Stalking Predator; Tovolar's Huntmaster // Tovolar's Packleader; Oakshade Stalker // Moonlit Ambusher; Brutal Cathar // Moonrage Brute |
| 45 | condition | if a player cast two or more spells last turn | 33 | Ulrich of the Krallenhorde // Ulrich, Uncontested Alpha; Lambholt Elder // Silverpelt Werewolf; Instigator Gang // Wildblood Pack; Daybreak Ranger // Nightfall Predator; Hinterland Logger // Timber Shredder |
| 46 | trigger | whenever this creature attacks or blocks | 33 | Loafing Giant; Rotting Giant; Wicker Warcrawler; Hamlet Captain; Carrion Rats |
| 47 | condition | if it's a creature card | 32 | Search for Survivors; Elven Farsight; Hauntwoods Shrieker; Sapling of Colfenor; Domri Rade |
| 48 | effect | prevent all combat damage that would be dealt this turn. | 32 | Leery Fogbeast; Jaheira's Respite; Pollen Lullaby; Fog; Sunstone |
| 49 | effect | this creature can block only creatures with flying. | 32 | Welkin Tern; Devoted Grafkeeper // Departed Soulkeeper; Cloud Elemental; Vaporkin; Cloud Sprite |
| 50 | effect | any player may activate this ability. | 31 | Flailing Manticore; Vintara Elephant; Xantcha, Sleeper Agent; Casey Jones, Asphalt Hooligan; Deadly Designs |
| 51 | effect | choose a background (you can have a background as a second commander.) | 31 | Halsin, Emerald Archdruid; Karlach, Fury of Avernus; Jaheira, Friend of the Forest; Shadowheart, Dark Justiciar; Erinis, Gloom Stalker |
| 52 | effect | venture into the dungeon. | 31 | Nadaar, Selfless Paladin; Bar the Gate; Veteran Dungeoneer; Dungeon Map; Radiant Solar |
| 53 | trigger | whenever this creature attacks and isn't blocked | 31 | Murk Dwellers; Guiltfeeder; Farrel's Zealot; Pygmy Hippo; Abyssal Nightstalker |
| 54 | effect | rebound (if you cast this spell from your hand, exile it as it resolves. at the beginning of your next upkeep, you may cast this card from exile without paying its mana cost.) | 30 | World at War; Faithless Salvaging; Staggershock; Profound Journey; Ephemerate |
| 55 | effect | roll a d 20. | 30 | Earth-Cult Elemental; Arcane Investigator; Herald of Hadar; Treasure Chest; Thunderwave |
| 56 | trigger | whenever you cast or copy an instant or sorcery spell | 30 | Zaffai, Thunder Conductor; Karok Wrangler; Extus, Oriq Overlord // Awaken the Blood Avatar; Prismari Apprentice; Leonin Lightscribe |
| 57 | condition | if you win | 28 | Research the Deep; Woodland Guidance; Titan's Revenge; Captivating Glance; Sentry Oak |
| 58 | effect | do this only once each turn. | 28 | Lucy MacLean, Positively Armed; Calix, Guided by Fate; Donal, Herald of Wings; Corruption of Towashi; Nykthos Paragon |
| 59 | effect | flanking (whenever a creature without flanking blocks this creature, the blocking creature gets -1/-1 until end of turn.) | 28 | Burning Shield Askari; Knight; Benalish Cavalry; Mtenda Herder; Zhalfirin Commander |
| 60 | effect | pay N life. | 28 | Brutal Cathar // Moonrage Brute; Prismari, the Inspiration; Sedgemoor Witch; Inner Sanctum; Invasion of Karsus // Refraction Elemental |

