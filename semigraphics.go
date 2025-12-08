package tview

import "github.com/gdamore/tcell/v3"

// Semigraphics provides easy access to Unicode characters for drawing.
// Using strings with \u escapes to keep the source ASCII-safe.
const (
	// General Punctuation U+2000-U+206F
	SemigraphicsHorizontalEllipsis = "\u2026" // …

	// Box Drawing U+2500-U+257F
	BoxDrawingsLightHorizontal                    = "\u2500" // ─
	BoxDrawingsHeavyHorizontal                    = "\u2501" // ━
	BoxDrawingsLightVertical                      = "\u2502" // │
	BoxDrawingsHeavyVertical                      = "\u2503" // ┃
	BoxDrawingsLightTripleDashHorizontal          = "\u2504" // ┄
	BoxDrawingsHeavyTripleDashHorizontal          = "\u2505" // ┅
	BoxDrawingsLightTripleDashVertical            = "\u2506" // ┆
	BoxDrawingsHeavyTripleDashVertical            = "\u2507" // ┇
	BoxDrawingsLightQuadrupleDashHorizontal       = "\u2508" // ┈
	BoxDrawingsHeavyQuadrupleDashHorizontal       = "\u2509" // ┉
	BoxDrawingsLightQuadrupleDashVertical         = "\u250a" // ┊
	BoxDrawingsHeavyQuadrupleDashVertical         = "\u250b" // ┋
	BoxDrawingsLightDownAndRight                  = "\u250c" // ┌
	BoxDrawingsDownLightAndRightHeavy             = "\u250d" // ┍
	BoxDrawingsDownHeavyAndRightLight             = "\u250e" // ┎
	BoxDrawingsHeavyDownAndRight                  = "\u250f" // ┏
	BoxDrawingsLightDownAndLeft                   = "\u2510" // ┐
	BoxDrawingsDownLightAndLeftHeavy              = "\u2511" // ┑
	BoxDrawingsDownHeavyAndLeftLight              = "\u2512" // ┒
	BoxDrawingsHeavyDownAndLeft                   = "\u2513" // ┓
	BoxDrawingsLightUpAndRight                    = "\u2514" // └
	BoxDrawingsUpLightAndRightHeavy               = "\u2515" // ┕
	BoxDrawingsUpHeavyAndRightLight               = "\u2516" // ┖
	BoxDrawingsHeavyUpAndRight                    = "\u2517" // ┗
	BoxDrawingsLightUpAndLeft                     = "\u2518" // ┘
	BoxDrawingsUpLightAndLeftHeavy                = "\u2519" // ┙
	BoxDrawingsUpHeavyAndLeftLight                = "\u251a" // ┚
	BoxDrawingsHeavyUpAndLeft                     = "\u251b" // ┛
	BoxDrawingsLightVerticalAndRight              = "\u251c" // ├
	BoxDrawingsVerticalLightAndRightHeavy         = "\u251d" // ┝
	BoxDrawingsUpHeavyAndRightDownLight           = "\u251e" // ┞
	BoxDrawingsDownHeavyAndRightUpLight           = "\u251f" // ┟
	BoxDrawingsVerticalHeavyAndRightLight         = "\u2520" // ┠
	BoxDrawingsDownLightAndRightUpHeavy           = "\u2521" // ┡
	BoxDrawingsUpLightAndRightDownHeavy           = "\u2522" // ┢
	BoxDrawingsHeavyVerticalAndRight              = "\u2523" // ┣
	BoxDrawingsLightVerticalAndLeft               = "\u2524" // ┤
	BoxDrawingsVerticalLightAndLeftHeavy          = "\u2525" // ┥
	BoxDrawingsUpHeavyAndLeftDownLight            = "\u2526" // ┦
	BoxDrawingsDownHeavyAndLeftUpLight            = "\u2527" // ┧
	BoxDrawingsVerticalHeavyAndLeftLight          = "\u2528" // ┨
	BoxDrawingsDownLightAndLeftUpHeavy            = "\u2529" // ┩
	BoxDrawingsUpLightAndLeftDownHeavy            = "\u252a" // ┪
	BoxDrawingsHeavyVerticalAndLeft               = "\u252b" // ┫
	BoxDrawingsLightDownAndHorizontal             = "\u252c" // ┬
	BoxDrawingsLeftHeavyAndRightDownLight         = "\u252d" // ┭
	BoxDrawingsRightHeavyAndLeftDownLight         = "\u252e" // ┮
	BoxDrawingsDownLightAndHorizontalHeavy        = "\u252f" // ┯
	BoxDrawingsDownHeavyAndHorizontalLight        = "\u2530" // ┰
	BoxDrawingsRightLightAndLeftDownHeavy         = "\u2531" // ┱
	BoxDrawingsLeftLightAndRightDownHeavy         = "\u2532" // ┲
	BoxDrawingsHeavyDownAndHorizontal             = "\u2533" // ┳
	BoxDrawingsLightUpAndHorizontal               = "\u2534" // ┴
	BoxDrawingsLeftHeavyAndRightUpLight           = "\u2535" // ┵
	BoxDrawingsRightHeavyAndLeftUpLight           = "\u2536" // ┶
	BoxDrawingsUpLightAndHorizontalHeavy          = "\u2537" // ┷
	BoxDrawingsUpHeavyAndHorizontalLight          = "\u2538" // ┸
	BoxDrawingsRightLightAndLeftUpHeavy           = "\u2539" // ┹
	BoxDrawingsLeftLightAndRightUpHeavy           = "\u253a" // ┺
	BoxDrawingsHeavyUpAndHorizontal               = "\u253b" // ┻
	BoxDrawingsLightVerticalAndHorizontal         = "\u253c" // ┼
	BoxDrawingsLeftHeavyAndRightVerticalLight     = "\u253d" // ┽
	BoxDrawingsRightHeavyAndLeftVerticalLight     = "\u253e" // ┾
	BoxDrawingsVerticalLightAndHorizontalHeavy    = "\u253f" // ┿
	BoxDrawingsUpHeavyAndDownHorizontalLight      = "\u2540" // ╀
	BoxDrawingsDownHeavyAndUpHorizontalLight      = "\u2541" // ╁
	BoxDrawingsVerticalHeavyAndHorizontalLight    = "\u2542" // ╂
	BoxDrawingsLeftUpHeavyAndRightDownLight       = "\u2543" // ╃
	BoxDrawingsRightUpHeavyAndLeftDownLight       = "\u2544" // ╄
	BoxDrawingsLeftDownHeavyAndRightUpLight       = "\u2545" // ╅
	BoxDrawingsRightDownHeavyAndLeftUpLight       = "\u2546" // ╆
	BoxDrawingsDownLightAndUpHorizontalHeavy      = "\u2547" // ╇
	BoxDrawingsUpLightAndDownHorizontalHeavy      = "\u2548" // ╈
	BoxDrawingsRightLightAndLeftVerticalHeavy     = "\u2549" // ╉
	BoxDrawingsLeftLightAndRightVerticalHeavy     = "\u254a" // ╊
	BoxDrawingsHeavyVerticalAndHorizontal         = "\u254b" // ╋
	BoxDrawingsLightDoubleDashHorizontal          = "\u254c" // ╌
	BoxDrawingsHeavyDoubleDashHorizontal          = "\u254d" // ╍
	BoxDrawingsLightDoubleDashVertical            = "\u254e" // ╎
	BoxDrawingsHeavyDoubleDashVertical            = "\u254f" // ╏
	BoxDrawingsDoubleHorizontal                   = "\u2550" // ═
	BoxDrawingsDoubleVertical                     = "\u2551" // ║
	BoxDrawingsDownSingleAndRightDouble           = "\u2552" // ╒
	BoxDrawingsDownDoubleAndRightSingle           = "\u2553" // ╓
	BoxDrawingsDoubleDownAndRight                 = "\u2554" // ╔
	BoxDrawingsDownSingleAndLeftDouble            = "\u2555" // ╕
	BoxDrawingsDownDoubleAndLeftSingle            = "\u2556" // ╖
	BoxDrawingsDoubleDownAndLeft                  = "\u2557" // ╗
	BoxDrawingsUpSingleAndRightDouble             = "\u2558" // ╘
	BoxDrawingsUpDoubleAndRightSingle             = "\u2559" // ╙
	BoxDrawingsDoubleUpAndRight                   = "\u255a" // ╚
	BoxDrawingsUpSingleAndLeftDouble              = "\u255b" // ╛
	BoxDrawingsUpDoubleAndLeftSingle              = "\u255c" // ╜
	BoxDrawingsDoubleUpAndLeft                    = "\u255d" // ╝
	BoxDrawingsVerticalSingleAndRightDouble       = "\u255e" // ╞
	BoxDrawingsVerticalDoubleAndRightSingle       = "\u255f" // ╟
	BoxDrawingsDoubleVerticalAndRight             = "\u2560" // ╠
	BoxDrawingsVerticalSingleAndLeftDouble        = "\u2561" // ╡
	BoxDrawingsVerticalDoubleAndLeftSingle        = "\u2562" // ╢
	BoxDrawingsDoubleVerticalAndLeft              = "\u2563" // ╣
	BoxDrawingsDownSingleAndHorizontalDouble      = "\u2564" // ╤
	BoxDrawingsDownDoubleAndHorizontalSingle      = "\u2565" // ╥
	BoxDrawingsDoubleDownAndHorizontal            = "\u2566" // ╦
	BoxDrawingsUpSingleAndHorizontalDouble        = "\u2567" // ╧
	BoxDrawingsUpDoubleAndHorizontalSingle        = "\u2568" // ╨
	BoxDrawingsDoubleUpAndHorizontal              = "\u2569" // ╩
	BoxDrawingsVerticalSingleAndHorizontalDouble  = "\u256a" // ╪
	BoxDrawingsVerticalDoubleAndHorizontalSingle  = "\u256b" // ╫
	BoxDrawingsDoubleVerticalAndHorizontal        = "\u256c" // ╬
	BoxDrawingsLightArcDownAndRight               = "\u256d" // ╭
	BoxDrawingsLightArcDownAndLeft                = "\u256e" // ╮
	BoxDrawingsLightArcUpAndLeft                  = "\u256f" // ╯
	BoxDrawingsLightArcUpAndRight                 = "\u2570" // ╰
	BoxDrawingsLightDiagonalUpperRightToLowerLeft = "\u2571" // ╱
	BoxDrawingsLightDiagonalUpperLeftToLowerRight = "\u2572" // ╲
	BoxDrawingsLightDiagonalCross                 = "\u2573" // ╳
	BoxDrawingsLightLeft                          = "\u2574" // ╴
	BoxDrawingsLightUp                            = "\u2575" // ╵
	BoxDrawingsLightRight                         = "\u2576" // ╶
	BoxDrawingsLightDown                          = "\u2577" // ╷
	BoxDrawingsHeavyLeft                          = "\u2578" // ╸
	BoxDrawingsHeavyUp                            = "\u2579" // ╹
	BoxDrawingsHeavyRight                         = "\u257a" // ╺
	BoxDrawingsHeavyDown                          = "\u257b" // ╻
	BoxDrawingsLightLeftAndHeavyRight             = "\u257c" // ╼
	BoxDrawingsLightUpAndHeavyDown                = "\u257d" // ╽
	BoxDrawingsHeavyLeftAndLightRight             = "\u257e" // ╾
	BoxDrawingsHeavyUpAndLightDown                = "\u257f" // ╿

	// Block Elements U+2580–U+259F
	BlockUpperHalfBlock                              = "\u2580" // ▀
	BlockLowerOneEighthBlock                         = "\u2581" // ▁
	BlockLowerOneQuarterBlock                        = "\u2582" // ▂
	BlockLowerThreeEighthsBlock                      = "\u2583" // ▃
	BlockLowerHalfBlock                              = "\u2584" // ▄
	BlockLowerFiveEighthsBlock                       = "\u2585" // ▅
	BlockLowerThreeQuartersBlock                     = "\u2586" // ▆
	BlockLowerSevenEighthsBlock                      = "\u2587" // ▇
	BlockFullBlock                                   = "\u2588" // █
	BlockLeftSevenEighthsBlock                       = "\u2589" // ▉
	BlockLeftThreeQuartersBlock                      = "\u258A" // ▊
	BlockLeftFiveEighthsBlock                        = "\u258B" // ▋
	BlockLeftHalfBlock                               = "\u258C" // ▌
	BlockLeftThreeEighthsBlock                       = "\u258D" // ▍
	BlockLeftOneQuarterBlock                         = "\u258E" // ▎
	BlockLeftOneEighthBlock                          = "\u258F" // ▏
	BlockRightHalfBlock                              = "\u2590" // ▐
	BlockLightShade                                  = "\u2591" // ░
	BlockMediumShade                                 = "\u2592" // ▒
	BlockDarkShade                                   = "\u2593" // ▓
	BlockUpperOneEighthBlock                         = "\u2594" // ▔
	BlockRightOneEighthBlock                         = "\u2595" // ▕
	BlockQuadrantLowerLeft                           = "\u2596" // ▖
	BlockQuadrantLowerRight                          = "\u2597" // ▗
	BlockQuadrantUpperLeft                           = "\u2598" // ▘
	BlockQuadrantUpperLeftAndLowerLeftAndLowerRight  = "\u2599" // ▙
	BlockQuadrantUpperLeftAndLowerRight              = "\u259A" // ▚
	BlockQuadrantUpperLeftAndUpperRightAndLowerLeft  = "\u259B" // ▛
	BlockQuadrantUpperLeftAndUpperRightAndLowerRight = "\u259C" // ▜
	BlockQuadrantUpperRight                          = "\u259D" // ▝
	BlockQuadrantUpperRightAndLowerLeft              = "\u259E" // ▞
	BlockQuadrantUpperRightAndLowerLeftAndLowerRight = "\u259F" // ▟
)

// SemigraphicJoints maps pairs of semigraphics strings to the resulting joint.
// All combinations for light and double lines are included.
var SemigraphicJoints = map[string]string{
	// ─ + │ = ┼
	BoxDrawingsLightHorizontal + BoxDrawingsLightVertical: BoxDrawingsLightVerticalAndHorizontal,
	// ─ + ┌ = ┬
	BoxDrawingsLightHorizontal + BoxDrawingsLightDownAndRight: BoxDrawingsLightDownAndHorizontal,
	// ─ + ┐ = ┬
	BoxDrawingsLightHorizontal + BoxDrawingsLightDownAndLeft: BoxDrawingsLightDownAndHorizontal,
	// ─ + └ = ┴
	BoxDrawingsLightHorizontal + BoxDrawingsLightUpAndRight: BoxDrawingsLightUpAndHorizontal,
	// ─ + ┘ = ┴
	BoxDrawingsLightHorizontal + BoxDrawingsLightUpAndLeft: BoxDrawingsLightUpAndHorizontal,
	// ─ + ├ = ┼
	BoxDrawingsLightHorizontal + BoxDrawingsLightVerticalAndRight: BoxDrawingsLightVerticalAndHorizontal,
	// ─ + ┤ = ┼
	BoxDrawingsLightHorizontal + BoxDrawingsLightVerticalAndLeft: BoxDrawingsLightVerticalAndHorizontal,
	// ─ + ┬ = ┬
	BoxDrawingsLightHorizontal + BoxDrawingsLightDownAndHorizontal: BoxDrawingsLightDownAndHorizontal,
	// ─ + ┴ = ┴
	BoxDrawingsLightHorizontal + BoxDrawingsLightUpAndHorizontal: BoxDrawingsLightUpAndHorizontal,
	// ─ + ┼ = ┼
	BoxDrawingsLightHorizontal + BoxDrawingsLightVerticalAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,

	// │ + ┌ = ├
	BoxDrawingsLightVertical + BoxDrawingsLightDownAndRight: BoxDrawingsLightVerticalAndRight,
	// │ + ┐ = ┤
	BoxDrawingsLightVertical + BoxDrawingsLightDownAndLeft: BoxDrawingsLightVerticalAndLeft,
	// │ + └ = ├
	BoxDrawingsLightVertical + BoxDrawingsLightUpAndRight: BoxDrawingsLightVerticalAndRight,
	// │ + ┘ = ┤
	BoxDrawingsLightVertical + BoxDrawingsLightUpAndLeft: BoxDrawingsLightVerticalAndLeft,
	// │ + ├ = ├
	BoxDrawingsLightVertical + BoxDrawingsLightVerticalAndRight: BoxDrawingsLightVerticalAndRight,
	// │ + ┤ = ┤
	BoxDrawingsLightVertical + BoxDrawingsLightVerticalAndLeft: BoxDrawingsLightVerticalAndLeft,
	// │ + ┬ = ┼
	BoxDrawingsLightVertical + BoxDrawingsLightDownAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,
	// │ + ┴ = ┼
	BoxDrawingsLightVertical + BoxDrawingsLightUpAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,
	// │ + ┼ = ┼
	BoxDrawingsLightVertical + BoxDrawingsLightVerticalAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,

	// ┌ + ┐ = ┬
	BoxDrawingsLightDownAndRight + BoxDrawingsLightDownAndLeft: BoxDrawingsLightDownAndHorizontal,
	// ┌ + └ = ├
	BoxDrawingsLightDownAndRight + BoxDrawingsLightUpAndRight: BoxDrawingsLightVerticalAndRight,
	// ┌ + ┘ = ┼
	BoxDrawingsLightDownAndRight + BoxDrawingsLightUpAndLeft: BoxDrawingsLightVerticalAndHorizontal,
	// ┌ + ├ = ├
	BoxDrawingsLightDownAndRight + BoxDrawingsLightVerticalAndRight: BoxDrawingsLightVerticalAndRight,
	// ┌ + ┤ = ┼
	BoxDrawingsLightDownAndRight + BoxDrawingsLightVerticalAndLeft: BoxDrawingsLightVerticalAndHorizontal,
	// ┌ + ┬ = ┬
	BoxDrawingsLightDownAndRight + BoxDrawingsLightDownAndHorizontal: BoxDrawingsLightDownAndHorizontal,
	// ┌ + ┴ = ┼
	BoxDrawingsLightDownAndRight + BoxDrawingsLightUpAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,
	// ┌ + ┼ = ┼
	BoxDrawingsLightDownAndRight + BoxDrawingsLightVerticalAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,

	// ┐ + └ = ┼
	BoxDrawingsLightDownAndLeft + BoxDrawingsLightUpAndRight: BoxDrawingsLightVerticalAndHorizontal,
	// ┐ + ┘ = ┤
	BoxDrawingsLightDownAndLeft + BoxDrawingsLightUpAndLeft: BoxDrawingsLightVerticalAndLeft,
	// ┐ + ├ = ┼
	BoxDrawingsLightDownAndLeft + BoxDrawingsLightVerticalAndRight: BoxDrawingsLightVerticalAndHorizontal,
	// ┐ + ┤ = ┤
	BoxDrawingsLightDownAndLeft + BoxDrawingsLightVerticalAndLeft: BoxDrawingsLightVerticalAndLeft,
	// ┐ + ┬ = ┬
	BoxDrawingsLightDownAndLeft + BoxDrawingsLightDownAndHorizontal: BoxDrawingsLightDownAndHorizontal,
	// ┐ + ┴ = ┼
	BoxDrawingsLightDownAndLeft + BoxDrawingsLightUpAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,
	// ┐ + ┼ = ┼
	BoxDrawingsLightDownAndLeft + BoxDrawingsLightVerticalAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,

	// └ + ┘ = ┴
	BoxDrawingsLightUpAndRight + BoxDrawingsLightUpAndLeft: BoxDrawingsLightUpAndHorizontal,
	// └ + ├ = ├
	BoxDrawingsLightUpAndRight + BoxDrawingsLightVerticalAndRight: BoxDrawingsLightVerticalAndRight,
	// └ + ┤ = ┼
	BoxDrawingsLightUpAndRight + BoxDrawingsLightVerticalAndLeft: BoxDrawingsLightVerticalAndHorizontal,
	// └ + ┬ = ┼
	BoxDrawingsLightUpAndRight + BoxDrawingsLightDownAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,
	// └ + ┴ = ┴
	BoxDrawingsLightUpAndRight + BoxDrawingsLightUpAndHorizontal: BoxDrawingsLightUpAndHorizontal,
	// └ + ┼ = ┼
	BoxDrawingsLightUpAndRight + BoxDrawingsLightVerticalAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,

	// ┘ + ├ = ┼
	BoxDrawingsLightUpAndLeft + BoxDrawingsLightVerticalAndRight: BoxDrawingsLightVerticalAndHorizontal,
	// ┘ + ┤ = ┤
	BoxDrawingsLightUpAndLeft + BoxDrawingsLightVerticalAndLeft: BoxDrawingsLightVerticalAndLeft,
	// ┘ + ┬ = ┼
	BoxDrawingsLightUpAndLeft + BoxDrawingsLightDownAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,
	// ┘ + ┴ = ┴
	BoxDrawingsLightUpAndLeft + BoxDrawingsLightUpAndHorizontal: BoxDrawingsLightUpAndHorizontal,
	// ┘ + ┼ = ┼
	BoxDrawingsLightUpAndLeft + BoxDrawingsLightVerticalAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,

	// ├ + ┤ = ┼
	BoxDrawingsLightVerticalAndRight + BoxDrawingsLightVerticalAndLeft: BoxDrawingsLightVerticalAndHorizontal,
	// ├ + ┬ = ┼
	BoxDrawingsLightVerticalAndRight + BoxDrawingsLightDownAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,
	// ├ + ┴ = ┼
	BoxDrawingsLightVerticalAndRight + BoxDrawingsLightUpAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,
	// ├ + ┼ = ┼
	BoxDrawingsLightVerticalAndRight + BoxDrawingsLightVerticalAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,

	// ┤ + ┬ = ┼
	BoxDrawingsLightVerticalAndLeft + BoxDrawingsLightDownAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,
	// ┤ + ┴ = ┼
	BoxDrawingsLightVerticalAndLeft + BoxDrawingsLightUpAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,
	// ┤ + ┼ = ┼
	BoxDrawingsLightVerticalAndLeft + BoxDrawingsLightVerticalAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,

	// ┬ + ┴ = ┼
	BoxDrawingsLightDownAndHorizontal + BoxDrawingsLightUpAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,
	// ┬ + ┼ = ┼
	BoxDrawingsLightDownAndHorizontal + BoxDrawingsLightVerticalAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,

	// ┴ + ┼ = ┼
	BoxDrawingsLightUpAndHorizontal + BoxDrawingsLightVerticalAndHorizontal: BoxDrawingsLightVerticalAndHorizontal,

	// ═ + ║ = ╬
	BoxDrawingsDoubleHorizontal + BoxDrawingsDoubleVertical: BoxDrawingsDoubleVerticalAndHorizontal,
}

// PrintJoinedSemigraphics prints a semigraphics string into the screen at the given
// position with the given style, joining it with any existing semigraphics.
func PrintJoinedSemigraphics(screen tcell.Screen, x, y int, str string, style tcell.Style) {
	previous, _, _ := screen.Get(x, y)

	var result string
	if str == previous {
		result = str
	} else {
		if str < previous {
			previous, str = str, previous
		}
		result = SemigraphicJoints[previous+str]
	}
	if result == "" {
		result = str
	}

	// We only print something if we have something.
	screen.Put(x, y, result, style)
}
