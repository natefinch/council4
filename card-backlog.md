# Card-Support Backlog

Every eligible Scryfall corpus card is evaluated with two signals and routed to the layer that blocks it:

- **Parser signal** (parser-only): `cardgen.ParseCardFaces` + `parser.DocumentCoverage` — is the card parser-complete, and which uncovered components remain?
- **Lowering signal** (full compile): compilecards' canonical report — did the card generate, and if not, which distinct diagnostic summaries blocked lowering? compilecards is the authority; an independent per-card recompile reconciles against it.

This produces two ranked, actionable queues. Regenerate with `mage cardBacklog`.

## Headline

- Eligible cards: 31838
- Supported (generated): 9592
- Parser-complete: 16591
- **Lowering backlog** (parser-complete, not generated): 7048
- **Parser backlog** (not parser-complete, not generated): 15198

Partition check: 9592 supported + 7048 lowering-backlog + 15198 parser-backlog = 31838 eligible. ✓

49 generated cards are not parser-complete. The lowerer fully generates them, but the parser-coverage harness does not span all their must-cover tokens (the residue tracked in `parser-coverage.md`). They are counted as **supported**, not routed to either backlog queue:

- Deadly Rollick
- Pongify
- Rapid Hybridization
- Chaos Warp
- Pillage
- Dark Banishing
- Shu General
- Oxidize
- Smothering Tithe
- Kodama's Reach
- Plague Wind
- Ancient Spider
- Flesh to Dust
- Wrath of God
- Selfless Glyphweaver // Deadly Vanity
- Obliterate
- Afterlife
- Flawless Maneuver
- Longbow Archer
- Rhystic Study
- Sever Soul
- Esper Sentinel
- Vinebred Brawler
- Retribution of the Meek
- Goblin Fire Fiend
- Terminate
- Jokulhaups
- Tunnel
- Cultivate
- Crumble
- Tel-Jilad Archers
- Shatterstorm
- Putrefy
- Fear of Being Hunted
- Smother
- Spite // Malice
- Giant Solifuge
- Dungeoneer's Pack
- Fissure
- Inescapable Brute
- Fierce Guardianship
- Damnation
- Gaea's Protector
- Bloomvine Regent // Claim Territory
- Perish
- Death Bomb
- Zhang Fei, Fierce Warrior
- Reprisal
- Lu Bu, Master-at-Arms

### Reconciliation guard

Generated membership is read from compilecards' canonical report. An independent per-card recompile cross-checks it; the run fails if they diverge.

- Authoritative generated (compilecards report): 9592
- Independent per-card recompile generated: 9592
- Divergences: 0 — the two pipelines agree. ✓

## Lowering queue

Parser-complete cards that do not yet lower, bucketed by distinct lowering diagnostic summary and ranked by affected-card count. Parsing is already done for these cards, so they are the lowest-risk backlog: this is `unsupported-reasons.md` restricted to the parser-complete subset.

| Rank | Reason | Affected (parser-complete) cards | Sole blockers | Example cards |
| --- | --- | --- | --- | --- |
| 1 | unsupported ordered effect sequence | 1471 | 1271 | Mind Extraction; Hunt the Hunter; Fear of Falling; Heartwarming Redemption; Fracturing Gust |
| 2 | unsupported optional effect | 689 | 603 | Courageous Outrider; Squadron Hawk; Dazzling Sphinx; Brawl-Bash Ogre; Ragefire Hellkite |
| 3 | unsupported ability content | 533 | 404 | Ulvenwald Captive // Ulvenwald Abomination; Scorching Missile; Junk Jet; Consuming Sinkhole; Howling Gale |
| 4 | unsupported static ability | 466 | 231 | Nissa, Worldsoul Speaker; Static Orb; Marang River Prowler; Starnheim Courser; Kykar, Zephyr Awakener |
| 5 | unsupported static declaration operation | 329 | 254 | Food Fight; Compulsory Rest; Ghostly Touch; Wingrattle Scarecrow; Executioner's Hood |
| 6 | unsupported activation cost | 295 | 168 | Thunderherd Migration; Krovikan Sorcerer; Etchings of the Chosen; Meteor Storm; City of Shadows |
| 7 | unsupported static declaration group | 277 | 218 | Dungeon Delver; Magma Sliver; Chamber of Manipulation; Etchings of the Chosen; Sedge Sliver |
| 8 | unsupported damage spell | 242 | 206 | Torrent of Fire; Armed Response; Combo Attack; Huatli, Dinosaur Knight; Outrage Shaman |
| 9 | unsupported counter placement | 230 | 119 | Sword-Swallowing Seraph; Greater Werewolf; Ent-Draught Basin; Huatli, Dinosaur Knight; Sporoloth Ancient |
| 10 | unsupported return spell | 222 | 179 | Dragon Fangs; Odunos River Trawler; Dragon Scales; Venser's Diffusion; Scapegoat |
| 11 | unsupported token creation | 215 | 160 | Witch's Mark; Goblin Gathering; Junk Jet; Zariel, Archduke of Avernus; Queen Brahne |
| 12 | unsupported power/toughness spell | 191 | 141 | Nissa, Worldsoul Speaker; Park Bleater; Zariel, Archduke of Avernus; Elspeth, Sun's Champion; Dawnhart Wardens |
| 13 | unsupported multiple spell abilities | 158 | 150 | Instill Infection; Blur; Refocus; Drag Under; Deadly Visit |
| 14 | unsupported exile spell | 158 | 104 | Ravnica at War; Trapjaw Tyrant; Moonring Mirror; Consuming Sinkhole; Pit of Offerings |
| 15 | unsupported destroy spell | 150 | 127 | Coils of the Medusa; Summon: Primal Odin; Unliving Psychopath; Bounty Agent; Silent Assassin |
| 16 | unsupported regenerate spell | 146 | 128 | Patchwork Gnomes; Darkling Stalker; Ranger en-Vec; Troll Ascetic; Rusted Slasher |
| 17 | unsupported Oracle construct | 137 | 0 | Marang River Prowler; Dawnbringer Cleric; Kykar, Zephyr Awakener; Astarion, the Decadent; Desperate Castaways |
| 18 | unsupported search effect | 135 | 109 | Dragonstorm Forecaster; Avatar of Growth; Green Sun's Zenith; Protean Hulk; Eye of Ugin |
| 19 | unsupported ability word | 132 | 84 | Dawnbringer Cleric; Bloodthorn Flail; Astarion, the Decadent; Solar Tide; Maha, Its Feathers Night |
| 20 | unsupported temporary keyword spell | 131 | 106 | Jareth, Leonine Titan; Vulture, Scheming Scavenger; Witch's Clinic; Pale Wayfarer; Swift Warden |
| 21 | unsupported life spell | 128 | 110 | South Wind Avatar; Miren, the Moaning Well; Mourning Thrull; Wall of Reverence; Wolverine Riders |
| 22 | unsupported enters-tapped replacement | 112 | 57 | Sedge Sliver; Stenn, Paranoid Partisan; Choco-Comet; Rest in Peace; Nevermore |
| 23 | unsupported unknown ability | 108 | 0 | Dawnbringer Cleric; Kykar, Zephyr Awakener; Astarion, the Decadent; Irreverent Revelers; Fangkeeper's Familiar |
| 24 | unsupported enters-with-counters replacement | 99 | 56 | Flycatcher Giraffid; Marketback Walker; Malefic Scythe; Glinting Creeper; Callous Sell-Sword // Burn Together |
| 25 | unsupported activation ability word | 94 | 83 | Half-Elf Monk; Zulaport Chainmage; Illuminor Szeras; Bagel and Schmear; Drana's Chosen |
| 26 | unsupported library placement | 91 | 79 | Misinformation; Fallow Earth; Chittering Rats; Footbottom Feast; Bone Harvest |
| 27 | unsupported sacrifice spell | 91 | 77 | Nefarox, Overlord of Grixis; Puppet Conjurer; Stenchskipper; Kibo, Uktabi Prince; Anowon, the Ruin Sage |
| 28 | unsupported activation references | 87 | 70 | Planebound Accomplice; Puresight Merrow; Titans' Nest; Spurnmage Advocate; Shackles |
| 29 | unsupported untap spell | 87 | 65 | Palinchron; Summoning Station; Freed from the Real; Battered Golem; Peregrine Drake |
| 30 | unsupported draw spell | 72 | 50 | Theft of Dreams; Marketback Walker; Fatigue; Shinestriker; Thought Sponge |
| 31 | unsupported tap spell | 68 | 46 | Waterknot; Torrent Elemental; Locked in the Cemetery; Freed from the Real; Amazing Acrobatics |
| 32 | unsupported gain-control spell | 66 | 50 | Unwilling Recruit; Donate; Slave of Bolas; Gilt-Leaf Archdruid; Legacy's Allure |
| 33 | unsupported keyword or ability grant | 62 | 53 | Shizo, Death's Storehouse; Summon: Primal Odin; Elvish Pathcutter; Furystoke Giant; Soraya the Falconer |
| 34 | unsupported mana symbol | 62 | 51 | Pit of Offerings; Cabal Stronghold; Sacrifice; Brightstone Ritual; Lotus Field |
| 35 | unsupported type line | 60 | 59 | Playable Delusionary Hydra; Notorious Sliver War; City's Blessing // Elemental; Demonic Tourist Laser; Night Brushwagg Ringmaster |
| 36 | unsupported discard spell | 55 | 47 | Tourach, Dread Cantor; Black Cat; Chilling Apparition; Tormented Thoughts; Urborg Mindsucker |
| 37 | unsupported mixed keyword ability | 53 | 33 | Chief Engineer; Irreverent Revelers; Sky Tether; Arbalest Engineers; Umaro, Raging Yeti |
| 38 | unsupported counter spell | 50 | 47 | Spell Blast; Drown in the Loch; Disdainful Stroke; Vigilant Martyr; Hydromorph Gull |
| 39 | unsupported activation condition | 44 | 40 | Metathran Aerostat; Wizard Replica; Martyr of Frost; Judge's Familiar; Patron Wizard |
| 40 | unsupported group power/toughness spell | 36 | 28 | Bloodline Culling; Battle Frenzy; Adventuring Gear; Keldon Mantle; Flowstone Embrace |
| 41 | unsupported mill spell | 35 | 28 | Mesmeric Orb; Flint Golem; Persistent Petitioners; Towering-Wave Mystic; Reef Pirates |
| 42 | unsupported can't-be-blocked effect | 31 | 24 | Frostpeak Yeti; Gingerbrute; Harbor Bandit; Spincrusher; Private Eye |
| 43 | unsupported parameterized keyword | 29 | 19 | Proven Combatant; Eldrazi Ravager; Jarl of the Forsaken; Shepherd of the Cosmos; Steadfast Sentinel |
| 44 | unsupported Enchant ability | 27 | 9 | Aura Graft; Robe of Mirrors; Corrupted Roots; Nettlevine Blight; Quiet Disrepair |
| 45 | unsupported manifest spell | 25 | 23 | Merfolk Observer; Orcish Spy; Dewdrop Spy; They Came from the Pipes; Gitaxian Probe |
| 46 | unsupported mana effect | 23 | 20 | City of Shadows; Blinkmoth Urn; Pristine Talisman; Skycloud Egg; Conduit of Storms // Conduit of Emrakul |
| 47 | unsupported keyword or ability loss | 22 | 17 | Cephalid Snitch; Canopy Claws; Thundercloud Elemental; Scarwood Hag; Gravity Well |
| 48 | unsupported fight spell | 21 | 15 | Pheres-Band Brawler; Surly Badgersaur; Clash of Titans; Wicked Wolf; Scab-Clan Giant |
| 49 | unsupported card layout | 20 | 20 | Nezumi Graverobber // Nighteyes the Desecrator; Faithful Squire // Kaiso, Memory of Loyalty; Jushi Apprentice // Tomoya the Revealer; Cunning Bandit // Azamuki, Treachery Incarnate; Nezumi Shortfang // Stabwhisker the Odious |
| 50 | unsupported loyalty ability | 12 | 0 | Kasmina, Enigma Sage; Sarkhan, Fireblood; Sorin, Grim Nemesis; Chandra, Flamecaller; Tezzeret, Cruel Captain |
| 51 | unsupported triggered ability | 11 | 10 | Alaborn Zealot; Cinder Wall; Wall of Junk; Tephraderm; Elder Land Wurm |
| 52 | unsupported divided damage spell | 9 | 7 | Fire Covenant; Rock Slide; Hail of Arrows; Roil's Retribution; Rolling Thunder |
| 53 | unsupported optional replacement effect | 8 | 8 | Callidus Assassin; Cursed Mirror; The Mimeoplasm; Worldheart Phoenix; Arsenal Thresher |
| 54 | unsupported static declaration duration | 8 | 5 | Roller Coaster; Jin Sakai, Ghost of Tsushima; Deathleaper, Terror Weapon; Protean Raider; Kiddie Coaster |
| 55 | unsupported delayed effect | 8 | 4 | Rienne, Angel of Rebirth; Library of Lat-Nam; Biolume Egg // Biolume Serpent; Tiana, Ship's Caretaker; Resurrection Orb |
| 56 | validation failed: oracle-without-abilities | 7 | 7 | Mishra's Warform; Icehide Golem; Cyberman; Forest Dryad; Morph |
| 57 | unsupported explore spell | 7 | 6 | Jadelight Spelunker; Hakbal of the Surging Soul; Legion Vanguard; Map; Seeker of Sunlight |
| 58 | unsupported static declaration shell | 5 | 3 | Inquisitor Greyfax; Bloodcrusher of Khorne; Cryptothrall; Vexilus Praetor; Frostcliff Siege |
| 59 | unsupported phase/step trigger phrase | 4 | 3 | Thumbscrews; Ivory Crane Netsuke; Numbing Dose; Complex Automaton |
| 60 | incomplete executable lowering | 4 | 2 | Swooping Protector; Wingshield Agent; Sleep-Cursed Faerie; Disciplined Duelist |

## Parser queue

Cards that are not parser-complete (and do not lower), bucketed by owning component family and normalized uncovered-span cluster, ranked by occurrence. This is the grammar-recognition backlog.

| Rank | Component | Cluster | Count | Example cards |
| --- | --- | --- | --- | --- |
| 1 | condition | if able | 165 | Impetuous Devils; The Foretold Soldier; Nacatl Hunt-Pride; Culling Mark; Legion Warboss |
| 2 | effect | you may choose new targets for the copy. | 123 | Verrak, Warped Sengir; Abstruse Archaic; League Guildmage; Psychic Rebuttal; Melek, Izzet Paragon |
| 3 | condition | if this spell was kicked | 99 | Strength of Night; Goblin Barrage; Colossal Growth; Overload; Stall for Time |
| 4 | effect | crew N (tap any number of creatures you control with total power N or more: this vehicle becomes an artifact creature until end of turn.) | 98 | War Balloon; Flywheel Racer; Mukotai Soulripper; Skybox Ferry; Sidequest: Card Collection // Magicked Card |
| 5 | effect | it's still a land. | 92 | Llanowar Loamspeaker; Hall of Storm Giants; Restless Vinestalk; Vastwood Animist; Faerie Conclave |
| 6 | trigger | when you cast this spell | 85 | Empyrial Storm; The Fourteenth Doctor; Ulamog, the Ceaseless Hunger; Temporal Extortion; Malicious Affliction |
| 7 | effect | you may pay{1}. | 80 | Ruthless Sniper; Smolder Initiate; Wooden Sphere; Rebellion of the Flamekin; Azorius Aethermage |
| 8 | effect | level N | 68 | Stormchaser's Talent; Builder's Talent; Leader's Talent; Sorcerer Class; Fortune Teller's Talent |
| 9 | effect | crew N | 67 | Mindlink Mech; Mighty Servant of Leuk-o; _____ _____ Rocketship; The Lunar Whale; Rocketeer Boostbuggy |
| 10 | effect | it can't be regenerated. | 65 | Polymorph; Phage the Untouchable; Fatal Blow; Wooden Stake; Shivan Emissary |
| 11 | trigger | whenever this creature enters or attacks | 63 | Omnivorous Flytrap; Graveyard Trespasser // Graveyard Glutton; Inferno Titan; Sigarda's Vanguard; Cemetery Illuminator |
| 12 | effect | changeling (this card is every creature type.) | 61 | Barkform Harvester; Guardian Gladewalker; Wings of Velis Vel; Firdoch Core; Mirror Entity |
| 13 | condition | if you search your library this way | 57 | Vraska's Scorn; Ashiok's Forerunner; Niambi, Faithful Healer; Claim Jumper; Fang-Druid Summoner |
| 14 | effect | prevent the next N damage that would be dealt to any target this turn. | 56 | Heal; Master Apothecary; Barrenton Medic; Rakalite; Militant Monk |
| 15 | effect | partner (you can have two commanders if both have partner.) | 55 | Ghost of Ramirez DePietro; Silas Renn, Seeker Adept; Krark, the Thumbless; Vial Smasher the Fierce; Francisco, Fowl Marauder |
| 16 | effect | enchant creature you control | 52 | One with the Kami; Endless Evil; Inferno Fist; Pitiless Fists; Dying Wish |
| 17 | trigger | when you cycle this card | 49 | Deem Worthy; Rampaging War Mammoth; Shefet Monitor; Agonasaur Rex; Windcaller Aven |
| 18 | trigger | whenever you cast a spell that targets this creature | 46 | Akroan Line Breaker; Lagonna-Band Trailblazer; Hero of Iroas; War-Wing Siren; Triton Cavalry |
| 19 | effect | target creature can't block this turn. | 45 | Nacatl Hunt-Pride; Mardu Roughrider; Bola Warrior; Unstoppable Ogre; Goblin Shortcutter |
| 20 | effect | flip a coin. | 44 | Goblin Lyre; Ydwen Efreet; Ral, Monsoon Mage // Ral, Leyline Prodigy; Krark, the Thumbless; Risky Move |
| 21 | effect | you may play that card this turn. | 41 | Party Thrasher; Professional Face-Breaker; Geistflame Reservoir; Dark-Dweller Oracle; Tablet of Discovery |
| 22 | condition | if it's a land card | 40 | Lantern of Revealing; Unexpected Results; Traveling Botanist; Countryside Crusher; Raiders' Karve |
| 23 | effect | the ring tempts you. | 40 | Uruk-hai Berserker; Horses of the Bruinen; Fiery Inscription; Relentless Rohirrim; Shortcut to Mushrooms |
| 24 | effect | you become the monarch. | 40 | Custodi Lich; Court of Ire; Palace Jailer; Grave Venerations; Court of Garenbrig |
| 25 | effect | you may pay{2}. | 39 | Minion Reflector; Esoteric Duplicator; Kavaron Harrier; Unassuming Sage; Terra, Herald of Hope |
| 26 | condition | if you lose the flip | 38 | Goblin Bomb; Goblin Lyre; Ydwen Efreet; Ral, Monsoon Mage // Ral, Leyline Prodigy; Krark, the Thumbless |
| 27 | effect | swampwalk (this creature can't be blocked as long as defending player controls a swamp.) | 38 | Whispering Shade; Dirtwater Wraith; Slithery Stalker; Lost Soul; Quag Vampires |
| 28 | condition | if this creature was kicked | 37 | Kavu Primarch; Faerie Squadron; Aether Figment; Skyclave Shade; Skyclave Sentinel |
| 29 | effect | fear (this creature can't be blocked except by artifact creatures and/or black creatures.) | 37 | Guiltfeeder; Commander Greven il-Vec; Dread; Lingering Tormentor; Undercity Shade |
| 30 | effect | start your engines ! (if you have no speed, it starts at 1. it increases once on each of your turns when an opponent loses life. max speed is 4.) | 37 | Momentum Breaker; Point the Way; Lightwheel Enhancements; Burnout Bashtronaut; Goblin Surveyor |
| 31 | effect | shadow (this creature can block or be blocked by only creatures with shadow.) | 36 | Dauthi Horror; Dauthi Mercenary; Thalakos Seer; Augur il-Vec; Soltari Guerrillas |
| 32 | condition | if you win the flip | 35 | Goblin Bomb; Goblin Lyre; Ral, Monsoon Mage // Ral, Leyline Prodigy; Krark, the Thumbless; Plasma Caster |
| 33 | effect | affinity for artifacts (this spell costs{1}less to cast for each artifact you control.) | 35 | Refurbished Familiar; Into Thin Air; Assert Authority; Valkyrie Aerial Unit; Furnace Dragon |
| 34 | effect | choose one. | 35 | Wail of the Forgotten; Jeska's Will; See Double; Inscription of Insight; Will of the Abzan |
| 35 | effect | daybound (if a player casts no spells during their own turn, it becomes night next turn.) | 34 | Graveyard Trespasser // Graveyard Glutton; Shady Traveler // Stalking Predator; Tovolar's Huntmaster // Tovolar's Packleader; Oakshade Stalker // Moonlit Ambusher; Brutal Cathar // Moonrage Brute |
| 36 | effect | fuse (you may cast one or both halves of this card from your hand.) | 34 | Breaking // Entering; Flesh // Blood; Protect // Serve; Turn // Burn; Ready // Willing |
| 37 | effect | islandwalk (this creature can't be blocked as long as defending player controls an island.) | 34 | Goblin Flotilla; Merrow Harbinger; Pale Bears; Stonybrook Banneret; Stormtide Leviathan |
| 38 | effect | nightbound (if a player casts at least two spells during their own turn, it becomes day next turn.) | 34 | Graveyard Trespasser // Graveyard Glutton; Shady Traveler // Stalking Predator; Tovolar's Huntmaster // Tovolar's Packleader; Oakshade Stalker // Moonlit Ambusher; Brutal Cathar // Moonrage Brute |
| 39 | condition | if a player cast two or more spells last turn | 33 | Ulrich of the Krallenhorde // Ulrich, Uncontested Alpha; Lambholt Elder // Silverpelt Werewolf; Instigator Gang // Wildblood Pack; Daybreak Ranger // Nightfall Predator; Hinterland Logger // Timber Shredder |
| 40 | trigger | whenever this creature attacks or blocks | 33 | Loafing Giant; Rotting Giant; Wicker Warcrawler; Hamlet Captain; Carrion Rats |
| 41 | condition | if it's a creature card | 32 | Search for Survivors; Elven Farsight; Hauntwoods Shrieker; Sapling of Colfenor; Domri Rade |
| 42 | effect | prevent all combat damage that would be dealt this turn. | 32 | Leery Fogbeast; Jaheira's Respite; Pollen Lullaby; Fog; Sunstone |
| 43 | effect | this creature can block only creatures with flying. | 32 | Welkin Tern; Devoted Grafkeeper // Departed Soulkeeper; Cloud Elemental; Vaporkin; Cloud Sprite |
| 44 | effect | any player may activate this ability. | 31 | Flailing Manticore; Vintara Elephant; Xantcha, Sleeper Agent; Casey Jones, Asphalt Hooligan; Deadly Designs |
| 45 | effect | choose a background (you can have a background as a second commander.) | 31 | Halsin, Emerald Archdruid; Karlach, Fury of Avernus; Jaheira, Friend of the Forest; Shadowheart, Dark Justiciar; Erinis, Gloom Stalker |
| 46 | effect | they can't be regenerated. | 31 | Wave of Terror; Tsabo's Decree; Kirtar's Wrath; Reign of Terror; Spreading Plague |
| 47 | effect | venture into the dungeon. | 31 | Nadaar, Selfless Paladin; Bar the Gate; Veteran Dungeoneer; Dungeon Map; Radiant Solar |
| 48 | trigger | whenever this creature attacks and isn't blocked | 31 | Murk Dwellers; Guiltfeeder; Farrel's Zealot; Pygmy Hippo; Abyssal Nightstalker |
| 49 | effect | rebound (if you cast this spell from your hand, exile it as it resolves. at the beginning of your next upkeep, you may cast this card from exile without paying its mana cost.) | 30 | World at War; Faithless Salvaging; Staggershock; Profound Journey; Ephemerate |
| 50 | effect | roll a d 20. | 30 | Earth-Cult Elemental; Arcane Investigator; Herald of Hadar; Treasure Chest; Thunderwave |
| 51 | condition | if you win | 28 | Research the Deep; Woodland Guidance; Titan's Revenge; Captivating Glance; Sentry Oak |
| 52 | effect | do this only once each turn. | 28 | Lucy MacLean, Positively Armed; Calix, Guided by Fate; Donal, Herald of Wings; Corruption of Towashi; Nykthos Paragon |
| 53 | effect | flanking (whenever a creature without flanking blocks this creature, the blocking creature gets -1/-1 until end of turn.) | 28 | Burning Shield Askari; Knight; Benalish Cavalry; Mtenda Herder; Zhalfirin Commander |
| 54 | effect | pay N life. | 28 | Brutal Cathar // Moonrage Brute; Prismari, the Inspiration; Sedgemoor Witch; Inner Sanctum; Invasion of Karsus // Refraction Elemental |
| 55 | effect | you choose a nonland card from it. | 28 | Gix's Caress; Grief; Drill Bit; Unmask; Memory Theft |
| 56 | trigger | when you unlock this door | 28 | Cramped Vents // Access Maze; Painter's Studio // Defaced Gallery; Underwater Tunnel // Slimy Aquarium; Moldering Gym // Weight Room; Glassworks // Shattered Yard |
| 57 | effect | aftermath (cast this spell only from your graveyard. then exile it.) | 27 | Heaven // Earth; Struggle // Survive; Claim // Fame; Farm // Market; Appeal // Authority |
| 58 | effect | ascend (if you control ten or more permanents, you get the city's blessing for the rest of the game.) | 27 | Radiant Destiny; Skymarcher Aspirant; Detective of the Month; Wayward Swordtooth; Arch of Orazca |
| 59 | effect | doctor's companion (you can have two commanders if the other is the doctor.) | 27 | Nyssa of Traken; Donna Noble; Barbara Wright; Bill Potts; Susan Foreman |
| 60 | effect | take an extra turn after this one. | 27 | Twice Upon a Time // Unlikely Meeting; Temporal Extortion; The Legend of Kuruk // Avatar Kuruk; Alchemist's Gambit; Chance for Glory |

