package wallet

// General constants for HD wallet implementation
const (
	HMACKeyStandard    = "ChainNet seed"
	HardenedIndex      = 0x80000000
	HardenedKeyPrefix  = 0x00
	HDPurposeBIP44     = 44
	HDChainNetCoinType = 0

	GapLimit = 20
)

type coinType uint32

const (
	TypeBitcoin               coinType = 0x80000000
	TypeTestnet                        = 0x80000001
	TypeLitecoin                       = 0x80000002
	TypeDogecoin                       = 0x80000003
	TypeReddcoin                       = 0x80000004
	TypeDash                           = 0x80000005
	TypePeercoin                       = 0x80000006
	TypeNamecoin                       = 0x80000007
	TypeFeathercoin                    = 0x80000008
	TypeCounterparty                   = 0x80000009
	TypeBlackcoin                      = 0x8000000a
	TypeNuShares                       = 0x8000000b
	TypeNuBits                         = 0x8000000c
	TypeMazacoin                       = 0x8000000d
	TypeViacoin                        = 0x8000000e
	TypeClearingHouse                  = 0x8000000f
	TypeRubycoin                       = 0x80000010
	TypeGroestlcoin                    = 0x80000011
	TypeDigitalcoin                    = 0x80000012
	TypeCannacoin                      = 0x80000013
	TypeDigiByte                       = 0x80000014
	TypeOpenAssets                     = 0x80000015
	TypeMonacoin                       = 0x80000016
	TypeClams                          = 0x80000017
	TypePrimecoin                      = 0x80000018
	TypeNeoscoin                       = 0x80000019
	TypeJumbucks                       = 0x8000001a
	TypeziftrCOIN                      = 0x8000001b
	TypeVertcoin                       = 0x8000001c
	TypeNXT                            = 0x8000001d
	TypeBurst                          = 0x8000001e
	TypeMonetaryUnit                   = 0x8000001f
	TypeZoom                           = 0x80000020
	TypeVpncoin                        = 0x80000021
	TypeCanadaeCoin                    = 0x80000022
	TypeShadowCash                     = 0x80000023
	TypeParkByte                       = 0x80000024
	TypePandacoin                      = 0x80000025
	TypeStartCOIN                      = 0x80000026
	TypeMOIN                           = 0x80000027
	TypeArgentum                       = 0x8000002D
	TypeGlobalCurrencyReserve          = 0x80000031
	TypeNovacoin                       = 0x80000032
	TypeAsiacoin                       = 0x80000033
	TypeBitcoindark                    = 0x80000034
	TypeDopecoin                       = 0x80000035
	TypeTemplecoin                     = 0x80000036
	TypeAIB                            = 0x80000037
	TypeEDRCoin                        = 0x80000038
	TypeSyscoin                        = 0x80000039
	TypeSolarcoin                      = 0x8000003a
	TypeSmileycoin                     = 0x8000003b
	TypeEther                          = 0x8000003c
	TypeEtherClassic                   = 0x8000003d
	TypeOpenChain                      = 0x80000040
	TypeOKCash                         = 0x80000045
	TypeDogecoinDark                   = 0x8000004d
	TypeElectronicGulden               = 0x8000004e
	TypeClubCoin                       = 0x8000004f
	TypeRichCoin                       = 0x80000050
	TypePotcoin                        = 0x80000051
	TypeQuarkcoin                      = 0x80000052
	TypeTerracoin                      = 0x80000053
	TypeGridcoin                       = 0x80000054
	TypeAuroracoin                     = 0x80000055
	TypeIXCoin                         = 0x80000056
	TypeGulden                         = 0x80000057
	TypeBitBean                        = 0x80000058
	TypeBata                           = 0x80000059
	TypeMyriadcoin                     = 0x8000005a
	TypeBitSend                        = 0x8000005b
	TypeUnobtanium                     = 0x8000005c
	TypeMasterTrader                   = 0x8000005d
	TypeGoldBlocks                     = 0x8000005e
	TypeSaham                          = 0x8000005f
	TypeChronos                        = 0x80000060
	TypeUbiquoin                       = 0x80000061
	TypeEvotion                        = 0x80000062
	TypeSaveTheOcean                   = 0x80000063
	TypeBigUp                          = 0x80000064
	TypeGameCredits                    = 0x80000065
	TypeDollarcoins                    = 0x80000066
	TypeZayedcoin                      = 0x80000067
	TypeDubaicoin                      = 0x80000068
	TypeStratis                        = 0x80000069
	TypeShilling                       = 0x8000006a
	TypePiggyCoin                      = 0x80000076
	TypeMonero                         = 0x80000080
	TypeNavCoin                        = 0x80000082
	TypeFactomFactoids                 = 0x80000083
	TypeFactomEntryCredits             = 0x80000084
	TypeZcash                          = 0x80000085
	TypeLisk                           = 0x80000086
	TypeFactomIdentity                 = 0x80000119
	TypeChainNet                       = 0x80000120
)

type changeType uint32

const (
	ExternalChangeType changeType = iota // ExternalChangeType for addresses shared with others
	InternalChangeType                   // InternalChangeType for not shared/visible addresses
)
