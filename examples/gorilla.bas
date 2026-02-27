' ============================================================================
'                    T U R B O   B A S I C   G O R I L L A S
' ============================================================================
'
' A faithful port of Microsoft's QBasic GORILLAS (Copyright (C) 1990)
' adapted for Turbo Basic (Borland) compatibility.
'
' YOUR MISSION: Hit your opponent with the exploding banana by varying the
' angle and power of your throw, taking into account wind speed, gravity,
' and the city skyline.
'
' SPEED: Game speed is governed by the constant SPEEDCONST (default 500).
' Increase it if the game runs too slowly; decrease if too fast.
'
' ============================================================================
' TURBO BASIC COMPATIBILITY NOTES
' ============================================================================
'
' The original QBasic GORILLAS relied on several QBasic-specific features
' that are not available in Turbo Basic.  The key changes made here are:
'
' 1. DECLARE statements removed
'    QBasic uses DECLARE SUB/FUNCTION for forward declarations.
'    Turbo Basic does not require (or support) them -- the compiler
'    resolves references to SUBs and FUNCTIONs automatically.
'
' 2. DEF FN replaced with FUNCTION ... END FUNCTION
'    QBasic's single-line DEF FnName(x) = expression is replaced
'    with a proper FUNCTION block, which is the Turbo Basic convention.
'
' 3. RND(1) replaced with RND
'    QBasic accepts RND(positive) to return the next random number.
'    Turbo Basic uses RND without an argument for the same result.
'
' 4. CALL keyword used for all SUB invocations
'    QBasic allows "SubName arg1, arg2" without CALL or parentheses.
'    Turbo Basic expects "CALL SubName(arg1, arg2)".  This port uses
'    the CALL syntax throughout for maximum compatibility.
'
' 5. AS ANY replaced with explicit types
'    QBasic DECLARE statements can use AS ANY for untyped parameters.
'    Since Turbo Basic has no DECLARE, the actual SUB/FUNCTION headers
'    always use the specific type (e.g., AS XYPoint).
'
' 6. SLEEP replaced with Rest
'    QBasic's SLEEP statement pauses for a given number of seconds.
'    Turbo Basic does not have SLEEP, so we call the Rest SUB instead.
'
' 7. REM $DYNAMIC metacommand
'    QBasic uses the comment-style metacommand '$DYNAMIC.
'    Turbo Basic uses REM $DYNAMIC -- functionally identical, it
'    makes all subsequent array declarations resizable via REDIM.
'
' ============================================================================
' TUTORIAL OVERVIEW
' ============================================================================
'
' This program demonstrates many fundamental BASIC programming concepts:
'
'   - TYPE ... END TYPE       User-defined data structures
'   - CONST                   Named constants for readability
'   - DIM SHARED              Module-level shared variables
'   - FUNCTION / SUB          Modular program design
'   - SCREEN / LINE / CIRCLE  Graphics programming (CGA and EGA)
'   - GET / PUT               Sprite capture and rendering
'   - PLAY                    Sound generation via MML strings
'   - ON ERROR GOTO           Structured error handling
'   - GOSUB / RETURN          Inline subroutines (for initialization)
'   - SELECT CASE             Multi-branch decision logic
'   - Projectile physics      Parabolic trajectory with wind and gravity
'
' ============================================================================

' ----------------------------------------------------------------------------
' Set default data type to INTEGER for faster arithmetic.
' All variables A-Z default to integer unless explicitly typed otherwise
' (e.g., with #, !, &, or $ suffixes, or AS clauses).
' ----------------------------------------------------------------------------
DEFINT A-Z

' ----------------------------------------------------------------------------
' Enable dynamic arrays so that DIM'd arrays can later be resized with REDIM.
' Without this, REDIM would cause a "Duplicate definition" error.
' ----------------------------------------------------------------------------
REM $DYNAMIC

' ============================================================================
' USER-DEFINED TYPES
' ============================================================================
' XYPoint stores integer X/Y screen coordinates.  It is used to record the
' upper-left corner of each building in the cityscape so that gorillas can
' be placed on rooftops and collision detection can reference building edges.
' ----------------------------------------------------------------------------
TYPE XYPoint
  XCoor AS INTEGER
  YCoor AS INTEGER
END TYPE

' ============================================================================
' CONSTANTS
' ============================================================================
' Named constants improve readability and make tuning easier.  They are
' evaluated at compile time and cost nothing at runtime.
' ----------------------------------------------------------------------------
CONST SPEEDCONST = 500          ' Base timing divisor -- increase to speed up
CONST TRUE = -1                 ' Boolean TRUE  (all bits set in two's complement)
CONST FALSE = NOT TRUE          ' Boolean FALSE (0)
CONST HITSELF = 1               ' Return value when a gorilla hits itself
CONST BACKATTR = 0              ' Background color attribute (black)
CONST OBJECTCOLOR = 1           ' Gorilla / banana color attribute
CONST WINDOWCOLOR = 14          ' Building window color (EGA yellow)
CONST SUNATTR = 3               ' Sun color attribute
CONST SUNHAPPY = FALSE          ' Sun draws a smile
CONST SUNSHOCK = TRUE           ' Sun draws an "O" mouth (shocked)
CONST RIGHTUP = 1               ' Gorilla arm pose: right arm raised
CONST LEFTUP = 2                ' Gorilla arm pose: left arm raised
CONST ARMSDOWN = 3              ' Gorilla arm pose: both arms down

' ============================================================================
' GLOBAL (SHARED) VARIABLES
' ============================================================================
' Variables declared DIM SHARED are visible to every SUB and FUNCTION in
' the program, behaving like global variables in other languages.
' ----------------------------------------------------------------------------

' --- Gorilla screen positions (two players) ---
DIM SHARED GorillaX(1 TO 2)    ' X pixel coordinate of each gorilla
DIM SHARED GorillaY(1 TO 2)    ' Y pixel coordinate of each gorilla
DIM SHARED LastBuilding         ' Index of the rightmost building drawn

' --- Math ---
DIM SHARED pi#                  ' Pi (double precision), computed once

' --- Sprite arrays for banana rotation frames ---
' These arrays store GET-captured bitmap data for the banana in four
' orientations.  The & suffix means LONG INTEGER elements.
' Initial size is 0; they are REDIMed in InitVars for CGA or EGA.
DIM SHARED LBan&(0)             ' Banana facing left
DIM SHARED RBan&(0)             ' Banana facing right
DIM SHARED UBan&(0)             ' Banana pointing up
DIM SHARED DBan&(0)             ' Banana pointing down

' --- Sprite arrays for gorilla arm poses ---
DIM SHARED GorD&(120)           ' Gorilla with both arms down
DIM SHARED GorL&(120)           ' Gorilla with left arm up
DIM SHARED GorR&(120)           ' Gorilla with right arm up

' --- Physics ---
DIM SHARED gravity#             ' Gravitational acceleration (m/s^2)
DIM SHARED Wind                 ' Wind speed and direction (negative = left)

' --- Screen geometry ---
DIM SHARED ScrHeight            ' Screen height in pixels (200 or 350)
DIM SHARED ScrWidth             ' Screen width in pixels  (320 or 640)
DIM SHARED Mode                 ' Graphics mode (1 = CGA, 9 = EGA)
DIM SHARED MaxCol               ' Maximum text column (40 or 80)

' --- Colors ---
DIM SHARED ExplosionColor       ' Color used for explosion circles
DIM SHARED SunColor             ' (Reserved) Sun palette color
DIM SHARED BackColor            ' Background palette color
DIM SHARED SunHit               ' Flag: TRUE if the sun was hit this turn

' --- Miscellaneous ---
DIM SHARED SunHt                ' Y threshold for sun collision detection
DIM SHARED GHeight              ' Gorilla sprite height in pixels
DIM SHARED MachSpeed AS SINGLE  ' Machine speed calibration value


' ############################################################################
'                           M A I N   P R O G R A M
' ############################################################################

' --- Set NumLock ON via BIOS keyboard flag byte ---
' Address 0000:0417 (decimal 1047) holds the keyboard shift-state flags.
' Bit 5 (value 32) is the NumLock toggle.  We force it on so that the
' numeric keypad produces digits for angle/velocity input.
DEF SEG = 0
KeyFlags = PEEK(1047)
IF (KeyFlags AND 32) = 0 THEN
  POKE 1047, KeyFlags OR 32
END IF
DEF SEG

' --- Initialize variables, detect graphics mode, load sprite data ---
GOSUB InitVars

' --- Show the title screen ---
CALL Intro

' --- Prompt for player names, number of games, and gravity ---
CALL GetInputs(Name1$, Name2$, NumGames)

' --- Draw gorilla intro animation (with optional music) ---
CALL GorillaIntro(Name1$, Name2$)

' --- Play the game! ---
CALL PlayGame(Name1$, Name2$, NumGames)

' --- Restore the original NumLock state before exiting ---
DEF SEG = 0
POKE 1047, KeyFlags
DEF SEG
END


' ############################################################################
'                         D A T A   S T A T E M E N T S
' ############################################################################
' These DATA lines encode pre-computed bitmap images for the banana sprite
' in four rotations, for both CGA (low-res) and EGA (hi-res) modes.
' The values are read into the LBan&/RBan&/UBan&/DBan& arrays during
' InitVars via RESTORE + READ loops.
' ----------------------------------------------------------------------------

' --- CGA banana bitmaps (3 longs each) ---
CGABanana:
  'BananaLeft
  DATA 327686, -252645316, 60
  'BananaDown
  DATA 196618, -1057030081, 49344
  'BananaUp
  DATA 196618, -1056980800, 63
  'BananaRight
  DATA 327686,  1010580720, 240

' --- EGA banana bitmaps (9 longs each) ---
EGABanana:
  'BananaLeft
  DATA 458758,202116096,471604224,943208448,943208448,943208448,471604224,202116096,0
  'BananaDown
  DATA 262153,-2134835200,-2134802239,-2130771968,-2130738945,8323072,8323199,4063232,4063294
  'BananaUp
  DATA 262153,4063232,4063294,8323072,8323199,-2130771968,-2130738945,-2134835200,-2134802239
  'BananaRight
  DATA 458758,-1061109760,-522133504,1886416896,1886416896,1886416896,-522133504,-1061109760,0


' ############################################################################
'                    G O S U B   S U B R O U T I N E S
' ############################################################################
' These labeled blocks are reached only via GOSUB from the main program.
' They share the main program's variable scope (unlike SUB/FUNCTION blocks).
' ============================================================================

' ----------------------------------------------------------------------------
' InitVars
' ----------------------------------------------------------------------------
' Initializes global variables, detects the best graphics mode, loads
' banana sprite data, and calibrates machine speed.
'
' TUTORIAL: This routine uses ON ERROR GOTO for graceful degradation.
' It first tries EGA mode (SCREEN 9, 640x350, 16 colors).  If the hardware
' does not support it, the error handler drops down to CGA mode
' (SCREEN 1, 320x200, 4 colors).  A second error trap catches 64K EGA
' cards that support mode 9 but not the full palette.
' ----------------------------------------------------------------------------
InitVars:
  ' Compute pi once using the arctangent identity: pi = 4 * arctan(1)
  pi# = 4 * ATN(1#)

  ' --- Attempt EGA mode first; fall back to CGA on error ---
  ON ERROR GOTO ScreenModeError
  Mode = 9
  SCREEN Mode

  ' --- Check for 64K EGA cards (they support SCREEN 9 but not PALETTE) ---
  ON ERROR GOTO PaletteError
  IF Mode = 9 THEN PALETTE 4, 0

  ' --- Disable error trapping now that mode detection is complete ---
  ON ERROR GOTO 0

  ' --- Calibrate machine speed for timing loops ---
  MachSpeed = CalcDelay

  ' --- Configure mode-specific settings ---
  IF Mode = 9 THEN
    ' ---- EGA mode: 640x350, 16 colors ----
    ScrWidth = 640
    ScrHeight = 350
    GHeight = 25              ' Gorilla height in pixels

    ' Load EGA banana bitmaps (9 elements each)
    RESTORE EGABanana
    REDIM LBan&(8), RBan&(8), UBan&(8), DBan&(8)

    FOR i = 0 TO 8
      READ LBan&(i)
    NEXT i
    FOR i = 0 TO 8
      READ DBan&(i)
    NEXT i
    FOR i = 0 TO 8
      READ UBan&(i)
    NEXT i
    FOR i = 0 TO 8
      READ RBan&(i)
    NEXT i

    SunHt = 39               ' Sun collision Y threshold

  ELSE
    ' ---- CGA mode: 320x200, 4 colors ----
    ScrWidth = 320
    ScrHeight = 200
    GHeight = 12

    ' Load CGA banana bitmaps (3 elements each)
    RESTORE CGABanana
    REDIM LBan&(2), RBan&(2), UBan&(2), DBan&(2)
    REDIM GorL&(20), GorD&(20), GorR&(20)

    FOR i = 0 TO 2
      READ LBan&(i)
    NEXT i
    FOR i = 0 TO 2
      READ DBan&(i)
    NEXT i
    FOR i = 0 TO 2
      READ UBan&(i)
    NEXT i
    FOR i = 0 TO 2
      READ RBan&(i)
    NEXT i

    ' CGA is slower per pixel, so increase the speed multiplier
    MachSpeed = MachSpeed * 1.3
    SunHt = 20
  END IF
RETURN


' ############################################################################
'                       E R R O R   H A N D L E R S
' ############################################################################
' These labels are targets for ON ERROR GOTO.  They handle hardware
' capability detection at startup.
' ============================================================================

' ----------------------------------------------------------------------------
' ScreenModeError
' ----------------------------------------------------------------------------
' Reached if SCREEN 9 (EGA) fails.  Falls back to SCREEN 1 (CGA).
' If CGA also fails, the program exits with an error message.
' RESUME retries the SCREEN statement with Mode = 1.
' ----------------------------------------------------------------------------
ScreenModeError:
  IF Mode = 1 THEN
    ' Even CGA failed -- no compatible graphics hardware
    CLS
    LOCATE 10, 5
    PRINT "Sorry, you must have CGA, EGA color, or VGA graphics to play GORILLA.BAS"
    END
  ELSE
    Mode = 1
    RESUME               ' Retry SCREEN with Mode = 1 (CGA)
  END IF

' ----------------------------------------------------------------------------
' PaletteError
' ----------------------------------------------------------------------------
' Reached if PALETTE fails on a 64K EGA card.  These cards support
' SCREEN 9 but not custom palettes, so we fall back to CGA mode.
' RESUME NEXT skips the PALETTE statement and continues execution.
' ----------------------------------------------------------------------------
PaletteError:
  Mode = 1               ' 64K EGA cards will run in CGA mode
  RESUME NEXT


' ############################################################################
'                F U N C T I O N S   A N D   S U B R O U T I N E S
' ############################################################################
' In Turbo Basic, FUNCTION and SUB blocks are compiled independently and
' can appear in any order after the main module.  Each block has its own
' local variable scope; only DIM SHARED variables are visible across blocks.
' ============================================================================


' ============================================================================
' FUNCTION FnRan (x)
' ============================================================================
' Returns a random integer between 1 and x (inclusive).
'
' TUTORIAL: In QBasic, this was a single-line DEF FN:
'   DEF FnRan(x) = INT(RND(1) * x) + 1
' Turbo Basic does not support DEF FN reliably across SUBs, so we use
' a proper FUNCTION block instead.
'
' RND returns a SINGLE-precision random number in [0, 1).
' INT() truncates toward negative infinity, giving [0, x-1].
' Adding 1 shifts the range to [1, x].
' ----------------------------------------------------------------------------
FUNCTION FnRan (x)
  FnRan = INT(RND * x) + 1
END FUNCTION


' ============================================================================
' FUNCTION CalcDelay! ()
' ============================================================================
' Calibrates machine speed by counting loop iterations for half a second.
' Returns a SINGLE-precision count that scales all timing in the game.
'
' TUTORIAL: Early PCs varied hugely in speed.  Rather than relying on
' a fixed delay constant, the game measures how fast the CPU can loop
' and uses that as a baseline.  The Rest SUB divides the desired pause
' by (MachSpeed / SPEEDCONST) to produce timing that adapts to the hardware.
' ----------------------------------------------------------------------------
FUNCTION CalcDelay!
  s! = TIMER
  DO
    i! = i! + 1
  LOOP UNTIL TIMER - s! >= .5
  CalcDelay! = i!
END FUNCTION


' ============================================================================
' FUNCTION Scl (n!)
' ============================================================================
' Scales a coordinate value for CGA (320x200) vs EGA (640x350).
' In EGA mode, returns the value rounded to the nearest integer.
' In CGA mode, halves the value (since CGA has half the resolution).
'
' TUTORIAL: The game was designed for EGA coordinates.  Rather than
' maintaining two separate sets of drawing code, this single function
' converts all coordinates at runtime.  A clever trick: fractional values
' like 2.9 scale to 1 in CGA (2.9 - 1 = 1.9, /2 = ~1) but to 3 in EGA,
' giving fine-grained control over both resolutions from one codebase.
' ----------------------------------------------------------------------------
FUNCTION Scl (n!)
  IF n! <> INT(n!) THEN
    IF Mode = 1 THEN n! = n! - 1
  END IF
  IF Mode = 1 THEN
    Scl = CINT(n! / 2 + .1)
  ELSE
    Scl = CINT(n!)
  END IF
END FUNCTION


' ============================================================================
' FUNCTION GetNum# (Row, Col)
' ============================================================================
' Displays a text cursor and accepts numeric input (with decimal point)
' at the given screen row and column.  Returns the entered value as a
' DOUBLE-precision number.
'
' TUTORIAL: This is a custom input routine that provides:
'   - Character-by-character validation (only digits and one decimal)
'   - Visual feedback with a blinking underscore cursor
'   - Backspace support for corrections
'   - Range limiting (max 360, suitable for angles)
' It replaces BASIC's built-in INPUT statement to prevent invalid entries.
' ----------------------------------------------------------------------------
FUNCTION GetNum# (Row, Col)
  Result$ = ""
  Done = FALSE

  ' Clear the keyboard buffer to discard stale keypresses
  WHILE INKEY$ <> "": WEND

  DO WHILE NOT Done

    ' Show current input with underscore cursor
    LOCATE Row, Col
    PRINT Result$; CHR$(95); "    ";

    Kbd$ = INKEY$
    SELECT CASE Kbd$
      CASE "0" TO "9"
        ' Append digit
        Result$ = Result$ + Kbd$
      CASE "."
        ' Allow one decimal point
        IF INSTR(Result$, ".") = 0 THEN
          Result$ = Result$ + Kbd$
        END IF
      CASE CHR$(13)
        ' Enter key -- validate and accept
        IF VAL(Result$) > 360 THEN
          Result$ = ""          ' Too large, reset
        ELSE
          Done = TRUE
        END IF
      CASE CHR$(8)
        ' Backspace -- remove last character
        IF LEN(Result$) > 0 THEN
          Result$ = LEFT$(Result$, LEN(Result$) - 1)
        END IF
      CASE ELSE
        ' Reject non-numeric keys with a beep
        IF LEN(Kbd$) > 0 THEN
          BEEP
        END IF
    END SELECT
  LOOP

  ' Display final value and return it
  LOCATE Row, Col
  PRINT Result$; " ";
  GetNum# = VAL(Result$)
END FUNCTION


' ============================================================================
' SUB Center (Row, Text$)
' ============================================================================
' Prints a string horizontally centered on the given screen row.
'
' TUTORIAL: MaxCol is the current screen width in text columns (40 or 80).
' Dividing by 2 gives the center column; subtracting half the string length
' positions the text so its midpoint aligns with the screen center.
' The trailing semicolon suppresses the automatic newline after PRINT.
' ----------------------------------------------------------------------------
SUB Center (Row, Text$)
  Col = MaxCol \ 2
  LOCATE Row, Col - (LEN(Text$) / 2 + .5)
  PRINT Text$;
END SUB


' ============================================================================
' SUB Rest (t#)
' ============================================================================
' Pauses execution for approximately t# time units.
'
' TUTORIAL: This is the game's universal delay function.  It converts the
' requested duration into a real-time delay using the machine speed
' calibration from CalcDelay.  The formula:
'   actual_delay = MachSpeed * t# / SPEEDCONST
' means that on a fast machine (large MachSpeed), the TIMER threshold is
' proportionally larger, keeping perceived speed consistent.
'
' Note: Replaces QBasic's SLEEP statement, which Turbo Basic lacks.
' ----------------------------------------------------------------------------
SUB Rest (t#)
  s# = TIMER
  t2# = MachSpeed * t# / SPEEDCONST
  DO
  LOOP UNTIL TIMER - s# > t2#
END SUB


' ============================================================================
' SUB SetScreen
' ============================================================================
' Configures palette colors appropriate for the current graphics mode.
'
' TUTORIAL: EGA mode supports 16 simultaneous colors from a palette of 64.
' The PALETTE statement remaps logical color numbers to physical colors.
' CGA mode has a fixed 4-color palette, so only the background color and
' palette selection (via COLOR) can be changed.
' ----------------------------------------------------------------------------
SUB SetScreen
  IF Mode = 9 THEN
    ExplosionColor = 2
    BackColor = 1
    PALETTE 0, 1          ' Background: dark blue
    PALETTE 1, 46         ' Object color: brown (gorilla/banana)
    PALETTE 2, 44         ' Explosion color: red
    PALETTE 3, 54         ' Sun color: bright yellow
    PALETTE 5, 7          ' Building color 1
    PALETTE 6, 4          ' Building color 2
    PALETTE 7, 3          ' Building color 3
    PALETTE 9, 63         ' Display text: bright white
  ELSE
    ExplosionColor = 2
    BackColor = 0
    COLOR BackColor, 2    ' CGA: black background, palette 2
  END IF
END SUB


' ============================================================================
' SUB SparklePause
' ============================================================================
' Displays a sparkling border animation on text screens until a key is
' pressed.  Used for the intro and game-over screens.
'
' TUTORIAL: The animation uses a rotating pattern string "* ... *" and
' MID$ with a sliding offset to create the illusion of stars orbiting the
' screen border.  The vertical sparkles use modular arithmetic to alternate
' between '*' and ' ' at each row, producing a chase-light effect.
' ----------------------------------------------------------------------------
SUB SparklePause
  COLOR 4, 0
  A$ = "*    *    *    *    *    *    *    *    *    *    *    *    *    *    *    *    *    "

  ' Clear keyboard buffer
  WHILE INKEY$ <> "": WEND

  WHILE INKEY$ = ""
    FOR A = 1 TO 5
      ' --- Horizontal sparkle rows (top and bottom) ---
      LOCATE 1, 1
      PRINT MID$(A$, A, 80);
      LOCATE 22, 1
      PRINT MID$(A$, 6 - A, 80);

      ' --- Vertical sparkle columns (left and right edges) ---
      FOR b = 2 TO 21
        c = (A + b) MOD 5
        IF c = 1 THEN
          LOCATE b, 80
          PRINT "*";
          LOCATE 23 - b, 1
          PRINT "*";
        ELSE
          LOCATE b, 80
          PRINT " ";
          LOCATE 23 - b, 1
          PRINT " ";
        END IF
      NEXT b
    NEXT A
  WEND
END SUB


' ============================================================================
' SUB Intro
' ============================================================================
' Displays the title screen with game instructions and plays a short tune.
'
' TUTORIAL: SCREEN 0 is text mode (80x25).  The MB prefix in the PLAY
' string means "Music Background" -- the tune plays while the program
' continues.  T160 sets tempo, O1 sets octave 1, L8 sets eighth notes.
' SparklePause provides the "press any key" wait with visual flair.
' ----------------------------------------------------------------------------
SUB Intro
  SCREEN 0
  WIDTH 80, 25
  MaxCol = 80
  COLOR 15, 0
  CLS

  CALL Center(4, "Q B a s i c    G O R I L L A S")
  COLOR 7
  CALL Center(6, "Copyright (C) Microsoft Corporation 1990")
  CALL Center(8, "Your mission is to hit your opponent with the exploding")
  CALL Center(9, "banana by varying the angle and power of your throw, taking")
  CALL Center(10, "into account wind speed, gravity, and the city skyline.")
  CALL Center(11, "The wind speed is shown by a directional arrow at the bottom")
  CALL Center(12, "of the playing field, its length relative to its strength.")
  CALL Center(24, "Press any key to continue")

  ' Play a short fanfare in the background
  PLAY "MBT160O1L8CDEDCDL4ECC"
  CALL SparklePause

  ' After intro, if we fell back to CGA, adjust column count
  IF Mode = 1 THEN MaxCol = 40
END SUB


' ============================================================================
' SUB GetInputs (Player1$, Player2$, NumGames)
' ============================================================================
' Prompts for player names, number of rounds, and gravity setting.
'
' TUTORIAL: LINE INPUT reads an entire line including commas and spaces,
' unlike INPUT which stops at commas.  This is important because player
' names might contain special characters.  VAL() converts a string to a
' number, returning 0 for non-numeric input.
'
' Parameters (all passed by reference -- modified values return to caller):
'   Player1$  - Name of player 1 (default "Player 1")
'   Player2$  - Name of player 2 (default "Player 2")
'   NumGames  - Total points to play to (default 3)
' ----------------------------------------------------------------------------
SUB GetInputs (Player1$, Player2$, NumGames)
  COLOR 7, 0
  CLS

  LOCATE 8, 15
  LINE INPUT "Name of Player 1 (Default = 'Player 1'): "; Player1$
  IF Player1$ = "" THEN
    Player1$ = "Player 1"
  ELSE
    Player1$ = LEFT$(Player1$, 10)    ' Truncate to 10 characters
  END IF

  LOCATE 10, 15
  LINE INPUT "Name of Player 2 (Default = 'Player 2'): "; Player2$
  IF Player2$ = "" THEN
    Player2$ = "Player 2"
  ELSE
    Player2$ = LEFT$(Player2$, 10)
  END IF

  ' --- Number of games (validated loop) ---
  DO
    LOCATE 12, 56: PRINT SPACE$(25);   ' Clear previous input
    LOCATE 12, 13
    INPUT "Play to how many total points (Default = 3)"; game$
    NumGames = VAL(LEFT$(game$, 2))
  LOOP UNTIL NumGames > 0 AND LEN(game$) < 3 OR LEN(game$) = 0
  IF NumGames = 0 THEN NumGames = 3

  ' --- Gravity setting ---
  DO
    LOCATE 14, 53: PRINT SPACE$(28);
    LOCATE 14, 17
    INPUT "Gravity in Meters/Sec (Earth = 9.8)"; grav$
    gravity# = VAL(grav$)
  LOOP UNTIL gravity# > 0 OR LEN(grav$) = 0
  IF gravity# = 0 THEN gravity# = 9.8
END SUB


' ============================================================================
' SUB GorillaIntro (Player1$, Player2$)
' ============================================================================
' Draws the gorillas for the first time (capturing their sprites into
' arrays via GET), then optionally plays an animated intro sequence.
'
' TUTORIAL: The gorilla sprites are drawn procedurally (not loaded from
' a file).  DrawGorilla uses LINE, CIRCLE, and PSET to construct the
' gorilla pixel by pixel, then GET captures the rectangular region into
' a LONG INTEGER array.  Three poses are captured: arms down, left arm
' up, and right arm up.  These arrays are later PUT onto the screen to
' animate the gorillas without redrawing them each frame.
'
' The VIEW PRINT statement restricts text scrolling to a region of the
' screen, preventing CLS 2 from erasing the entire display.
' ----------------------------------------------------------------------------
SUB GorillaIntro (Player1$, Player2$)
  LOCATE 16, 34: PRINT "--------------"
  LOCATE 18, 34: PRINT "V = View Intro"
  LOCATE 19, 34: PRINT "P = Play Game"
  LOCATE 21, 35: PRINT "Your Choice?"

  ' Wait for a keypress
  DO WHILE Char$ = ""
    Char$ = INKEY$
  LOOP

  ' Set gorilla drawing position based on graphics mode
  IF Mode = 1 THEN
    x = 125
    y = 100
  ELSE
    x = 278
    y = 175
  END IF

  SCREEN Mode
  CALL SetScreen

  IF Mode = 1 THEN CALL Center(5, "Please wait while gorillas are drawn.")

  ' Restrict text viewport so CLS 2 only clears the drawing area
  VIEW PRINT 9 TO 24

  ' Temporarily hide the gorilla color during drawing (EGA only)
  IF Mode = 9 THEN PALETTE OBJECTCOLOR, BackColor

  ' Draw and capture all three gorilla poses
  CALL DrawGorilla(x, y, ARMSDOWN)
  CLS 2
  CALL DrawGorilla(x, y, LEFTUP)
  CLS 2
  CALL DrawGorilla(x, y, RIGHTUP)
  CLS 2

  ' Restore full text viewport
  VIEW PRINT 1 TO 25

  ' Reveal the gorilla color
  IF Mode = 9 THEN PALETTE OBJECTCOLOR, 46

  ' --- Optional intro animation ---
  IF UCASE$(Char$) = "V" THEN
    CALL Center(2, "Q B A S I C   G O R I L L A S")
    CALL Center(5, "             STARRING:               ")
    P$ = Player1$ + " AND " + Player2$
    CALL Center(7, P$)

    ' Show gorillas side by side
    PUT (x - 13, y), GorD&, PSET
    PUT (x + 47, y), GorD&, PSET
    CALL Rest(1)

    ' Dance animation with musical accompaniment
    PUT (x - 13, y), GorL&, PSET
    PUT (x + 47, y), GorR&, PSET
    PLAY "t120o1l16b9n0baan0bn0bn0baaan0b9n0baan0b"
    CALL Rest(.3)

    PUT (x - 13, y), GorR&, PSET
    PUT (x + 47, y), GorL&, PSET
    PLAY "o2l16e-9n0e-d-d-n0e-n0e-n0e-d-d-d-n0e-9n0e-d-d-n0e-"
    CALL Rest(.3)

    PUT (x - 13, y), GorL&, PSET
    PUT (x + 47, y), GorR&, PSET
    PLAY "o2l16g-9n0g-een0g-n0g-n0g-eeen0g-9n0g-een0g-"
    CALL Rest(.3)

    PUT (x - 13, y), GorR&, PSET
    PUT (x + 47, y), GorL&, PSET
    PLAY "o2l16b9n0baan0g-n0g-n0g-eeen0o1b9n0baan0b"
    CALL Rest(.3)

    ' Final rapid dance
    FOR i = 1 TO 4
      PUT (x - 13, y), GorL&, PSET
      PUT (x + 47, y), GorR&, PSET
      PLAY "T160O0L32EFGEFDC"
      CALL Rest(.1)
      PUT (x - 13, y), GorR&, PSET
      PUT (x + 47, y), GorL&, PSET
      PLAY "T160O0L32EFGEFDC"
      CALL Rest(.1)
    NEXT
  END IF
END SUB


' ============================================================================
' SUB DoSun (Mouth)
' ============================================================================
' Draws the sun at the top center of the screen.
'
' TUTORIAL: The sun is drawn using CIRCLE (body) and LINE (rays) with
' trigonometric positioning.  Two mouth states are supported:
'   - SUNHAPPY (FALSE): a smile arc drawn with CIRCLE's start/stop angles
'   - SUNSHOCK (TRUE):  a small filled circle "O" mouth
' The sun reacts when hit by a banana -- a fun cosmetic touch.
'
' Parameters:
'   Mouth - TRUE for shocked face, FALSE for happy face
' ----------------------------------------------------------------------------
SUB DoSun (Mouth)
  ' Sun position: top center of screen
  x = ScrWidth \ 2
  y = Scl(25)

  ' Clear the old sun (erase with background color)
  LINE (x - Scl(22), y - Scl(18))-(x + Scl(22), y + Scl(18)), BACKATTR, BF

  ' --- Draw sun body (filled circle) ---
  CIRCLE (x, y), Scl(12), SUNATTR
  PAINT (x, y), SUNATTR

  ' --- Draw rays (8 lines radiating outward) ---
  LINE (x - Scl(20), y)-(x + Scl(20), y), SUNATTR                           ' Horizontal
  LINE (x, y - Scl(15))-(x, y + Scl(15)), SUNATTR                           ' Vertical
  LINE (x - Scl(15), y - Scl(10))-(x + Scl(15), y + Scl(10)), SUNATTR       ' Diagonal \
  LINE (x - Scl(15), y + Scl(10))-(x + Scl(15), y - Scl(10)), SUNATTR       ' Diagonal /
  LINE (x - Scl(8), y - Scl(13))-(x + Scl(8), y + Scl(13)), SUNATTR         ' Steep /
  LINE (x - Scl(8), y + Scl(13))-(x + Scl(8), y - Scl(13)), SUNATTR         ' Steep \
  LINE (x - Scl(18), y - Scl(5))-(x + Scl(18), y + Scl(5)), SUNATTR         ' Shallow \
  LINE (x - Scl(18), y + Scl(5))-(x + Scl(18), y - Scl(5)), SUNATTR         ' Shallow /

  ' --- Draw mouth ---
  IF Mouth THEN
    ' Shocked: small "O" mouth
    CIRCLE (x, y + Scl(5)), Scl(2.9), 0
    PAINT (x, y + Scl(5)), 0, 0
  ELSE
    ' Happy: smile arc (210 to 330 degrees)
    CIRCLE (x, y), Scl(8), 0, (210 * pi# / 180), (330 * pi# / 180)
  END IF

  ' --- Draw eyes (two small dots) ---
  CIRCLE (x - 3, y - 2), 1, 0
  CIRCLE (x + 3, y - 2), 1, 0
  PSET (x - 3, y - 2), 0
  PSET (x + 3, y - 2), 0
END SUB


' ============================================================================
' SUB DrawBan (xc#, yc#, r, bc)
' ============================================================================
' Draws or erases the banana sprite at the given position and rotation.
'
' TUTORIAL: The banana has four rotation frames (0-3) stored in the
' LBan&, UBan&, DBan&, RBan& arrays.  PUT with PSET draws the sprite
' opaquely (replacing background pixels).  PUT with XOR toggles pixels,
' which erases the sprite when called a second time at the same position.
'
' Parameters:
'   xc#, yc# - Screen coordinates (double precision for smooth motion)
'   r         - Rotation frame (0=left, 1=up, 2=down, 3=right)
'   bc        - TRUE to draw, FALSE to erase
' ----------------------------------------------------------------------------
SUB DrawBan (xc#, yc#, r, bc)
  SELECT CASE r
    CASE 0
      IF bc THEN PUT (xc#, yc#), LBan&, PSET ELSE PUT (xc#, yc#), LBan&, XOR
    CASE 1
      IF bc THEN PUT (xc#, yc#), UBan&, PSET ELSE PUT (xc#, yc#), UBan&, XOR
    CASE 2
      IF bc THEN PUT (xc#, yc#), DBan&, PSET ELSE PUT (xc#, yc#), DBan&, XOR
    CASE 3
      IF bc THEN PUT (xc#, yc#), RBan&, PSET ELSE PUT (xc#, yc#), RBan&, XOR
  END SELECT
END SUB


' ============================================================================
' SUB DrawGorilla (x, y, arms)
' ============================================================================
' Procedurally draws a gorilla sprite and captures it with GET.
'
' TUTORIAL: This is the most complex drawing routine in the game.  It
' constructs a gorilla from geometric primitives:
'   - LINE with BF (Box Fill) for the head, body, and neck
'   - CIRCLE with partial arcs for the legs, chest, and arms
'   - PSET for individual nose pixels
'
' After drawing, GET captures the rectangular bounding box into one of
' three arrays (GorD&, GorL&, GorR&) depending on the arm pose.
' These arrays are later used with PUT for fast sprite rendering.
'
' CIRCLE arc angles are specified in radians.  The partial arc syntax is:
'   CIRCLE (x, y), radius, color, start_angle, end_angle
'
' Parameters:
'   x, y - Top-left reference point for the gorilla
'   arms - ARMSDOWN (3), LEFTUP (2), or RIGHTUP (1)
' ----------------------------------------------------------------------------
SUB DrawGorilla (x, y, arms)
  DIM i AS SINGLE         ' Loop counter needs single precision for Scl()

  ' --- Head ---
  LINE (x - Scl(4), y)-(x + Scl(2.9), y + Scl(6)), OBJECTCOLOR, BF
  LINE (x - Scl(5), y + Scl(2))-(x + Scl(4), y + Scl(4)), OBJECTCOLOR, BF

  ' --- Eyes and brow (black line across face) ---
  LINE (x - Scl(3), y + Scl(2))-(x + Scl(2), y + Scl(2)), 0

  ' --- Nose (EGA only -- too small for CGA resolution) ---
  IF Mode = 9 THEN
    FOR i = -2 TO -1
      PSET (x + i, y + 4), 0
      PSET (x + i + 3, y + 4), 0
    NEXT i
  END IF

  ' --- Neck ---
  LINE (x - Scl(3), y + Scl(7))-(x + Scl(2), y + Scl(7)), OBJECTCOLOR

  ' --- Body (torso + lower body) ---
  LINE (x - Scl(8), y + Scl(8))-(x + Scl(6.9), y + Scl(14)), OBJECTCOLOR, BF
  LINE (x - Scl(6), y + Scl(15))-(x + Scl(4.9), y + Scl(20)), OBJECTCOLOR, BF

  ' --- Legs (drawn as partial circle arcs) ---
  FOR i = 0 TO 4
    CIRCLE (x + Scl(i), y + Scl(25)), Scl(10), OBJECTCOLOR, 3 * pi# / 4, 9 * pi# / 8
    CIRCLE (x + Scl(-6) + Scl(i - .1), y + Scl(25)), Scl(10), OBJECTCOLOR, 15 * pi# / 8, pi# / 4
  NEXT

  ' --- Chest detail (curved lines) ---
  CIRCLE (x - Scl(4.9), y + Scl(10)), Scl(4.9), 0, 3 * pi# / 2, 0
  CIRCLE (x + Scl(4.9), y + Scl(10)), Scl(4.9), 0, pi#, 3 * pi# / 2

  ' --- Arms (different poses, captured with GET) ---
  FOR i = -5 TO -1
    SELECT CASE arms
      CASE RIGHTUP
        ' Right arm raised, left arm lowered
        CIRCLE (x + Scl(i - .1), y + Scl(14)), Scl(9), OBJECTCOLOR, 3 * pi# / 4, 5 * pi# / 4
        CIRCLE (x + Scl(4.9) + Scl(i), y + Scl(4)), Scl(9), OBJECTCOLOR, 7 * pi# / 4, pi# / 4
        GET (x - Scl(15), y - Scl(1))-(x + Scl(14), y + Scl(28)), GorR&
      CASE LEFTUP
        ' Left arm raised, right arm lowered
        CIRCLE (x + Scl(i - .1), y + Scl(4)), Scl(9), OBJECTCOLOR, 3 * pi# / 4, 5 * pi# / 4
        CIRCLE (x + Scl(4.9) + Scl(i), y + Scl(14)), Scl(9), OBJECTCOLOR, 7 * pi# / 4, pi# / 4
        GET (x - Scl(15), y - Scl(1))-(x + Scl(14), y + Scl(28)), GorL&
      CASE ARMSDOWN
        ' Both arms down
        CIRCLE (x + Scl(i - .1), y + Scl(14)), Scl(9), OBJECTCOLOR, 3 * pi# / 4, 5 * pi# / 4
        CIRCLE (x + Scl(4.9) + Scl(i), y + Scl(14)), Scl(9), OBJECTCOLOR, 7 * pi# / 4, pi# / 4
        GET (x - Scl(15), y - Scl(1))-(x + Scl(14), y + Scl(28)), GorD&
    END SELECT
  NEXT i
END SUB


' ============================================================================
' SUB DoExplosion (x#, y#)
' ============================================================================
' Animates an explosion at the given coordinates.
'
' TUTORIAL: The explosion is drawn as an expanding then contracting
' series of concentric circles.  The expansion phase draws circles in
' ExplosionColor; the contraction phase redraws them in BACKATTR to
' erase them, creating a brief flash effect.  PLAY generates a sound
' effect simultaneously (MB = Music Background).
'
' Parameters:
'   x#, y# - Explosion center coordinates
' ----------------------------------------------------------------------------
SUB DoExplosion (x#, y#)
  PLAY "MBO0L32EFGEFDC"
  Radius = ScrHeight / 50
  IF Mode = 9 THEN Inc# = .5 ELSE Inc# = .41

  ' --- Expand ---
  FOR c# = 0 TO Radius STEP Inc#
    CIRCLE (x#, y#), c#, ExplosionColor
  NEXT c#

  ' --- Contract (erase) ---
  FOR c# = Radius TO 0 STEP (-1 * Inc#)
    CIRCLE (x#, y#), c#, BACKATTR
    FOR i = 1 TO 100
    NEXT i
    CALL Rest(.005)
  NEXT c#
END SUB


' ============================================================================
' FUNCTION ExplodeGorilla (x#, y#)
' ============================================================================
' Animates a gorilla exploding after a direct hit.  Returns the player
' number of the gorilla that was hit.
'
' TUTORIAL: This is a multi-phase animation:
'   1. Initial flash with elliptical CIRCLEs (aspect ratio -1.57 makes
'      them wider than tall, simulating a horizontal blast).
'   2. Expanding colored rings that alternate between two colors.
'   3. Contracting erase to leave a gap where the gorilla was.
' The player hit is determined by which half of the screen the shot
' landed on (left half = player 1, right half = player 2).
'
' Parameters:
'   x#, y# - Impact coordinates
' Returns:
'   Player number (1 or 2) of the hit gorilla
' ----------------------------------------------------------------------------
FUNCTION ExplodeGorilla (x#, y#)
  YAdj = Scl(12)
  XAdj = Scl(5)
  SclX# = ScrWidth / 320
  SclY# = ScrHeight / 200

  ' Determine which gorilla was hit based on screen half
  IF x# < ScrWidth / 2 THEN PlayerHit = 1 ELSE PlayerHit = 2

  PLAY "MBO0L16EFGEFDC"

  ' --- Phase 1: Initial explosion flash ---
  FOR i = 1 TO 8 * SclX#
    CIRCLE (GorillaX(PlayerHit) + 3.5 * SclX# + XAdj, GorillaY(PlayerHit) + 7 * SclY# + YAdj), i, ExplosionColor, , , -1.57
    LINE (GorillaX(PlayerHit) + 7 * SclX#, GorillaY(PlayerHit) + 9 * SclY# - i)-(GorillaX(PlayerHit), GorillaY(PlayerHit) + 9 * SclY# - i), ExplosionColor
  NEXT i

  ' --- Phase 2: Expanding colored rings ---
  FOR i = 1 TO 16 * SclX#
    IF i < (8 * SclX#) THEN CIRCLE (GorillaX(PlayerHit) + 3.5 * SclX# + XAdj, GorillaY(PlayerHit) + 7 * SclY# + YAdj), (8 * SclX# + 1) - i, BACKATTR, , , -1.57
    CIRCLE (GorillaX(PlayerHit) + 3.5 * SclX# + XAdj, GorillaY(PlayerHit) + YAdj), i, i MOD 2 + 1, , , -1.57
  NEXT i

  ' --- Phase 3: Contracting erase ---
  FOR i = 24 * SclX# TO 1 STEP -1
    CIRCLE (GorillaX(PlayerHit) + 3.5 * SclX# + XAdj, GorillaY(PlayerHit) + YAdj), i, BACKATTR, , , -1.57
    FOR Count = 1 TO 200
    NEXT
  NEXT i

  ExplodeGorilla = PlayerHit
END FUNCTION


' ============================================================================
' SUB MakeCityScape (BCoor() AS XYPoint)
' ============================================================================
' Generates a random city skyline and draws it on screen.
'
' TUTORIAL: The cityscape is the game's "terrain".  Buildings are drawn
' as filled rectangles (LINE with BF) with randomly colored windows.
' The skyline follows one of six randomly chosen slope patterns:
'
'   Slope 1: Heights increase left to right (upward slope)
'   Slope 2: Heights decrease left to right (downward slope)
'   Slope 3-5: "V" shape - increases to center, then decreases (most common)
'   Slope 6: Inverted "V" - decreases to center, then increases
'
' Building coordinates are stored in the BCoor() array for later use
' by PlaceGorillas and collision detection.
'
' After drawing buildings, this routine also:
'   - Generates a random wind speed and direction
'   - Draws a wind indicator arrow at the bottom of the screen
'
' NOTE: In the original QBasic source, CASE 4 in the building height
' adjustment loop is unreachable because CASE 3 TO 5 catches it first.
' This is preserved faithfully from the original.
'
' Parameters:
'   BCoor() - Array of XYPoint to receive building coordinates
' ----------------------------------------------------------------------------
SUB MakeCityScape (BCoor() AS XYPoint)
  x = 2

  ' --- Choose a random slope pattern for the skyline ---
  Slope = FnRan(6)
  SELECT CASE Slope
    CASE 1: NewHt = 15           ' Upward slope
    CASE 2: NewHt = 130          ' Downward slope
    CASE 3 TO 5: NewHt = 15     ' "V" slope (most common)
    CASE 6: NewHt = 130          ' Inverted "V" slope
  END SELECT

  ' --- Mode-dependent building parameters ---
  IF Mode = 9 THEN
    BottomLine = 335             ' Y coordinate of ground level
    HtInc = 10                   ' Height increment per building
    DefBWidth = 37               ' Base building width
    RandomHeight = 120           ' Random height variation
    WWidth = 3                   ' Window width in pixels
    WHeight = 6                  ' Window height in pixels
    WDifV = 15                   ' Vertical window spacing
    WDifh = 10                   ' Horizontal window spacing
  ELSE
    BottomLine = 190
    HtInc = 6
    NewHt = NewHt * 20 \ 35     ' Scale initial height for CGA
    DefBWidth = 18
    RandomHeight = 54
    WWidth = 1
    WHeight = 2
    WDifV = 5
    WDifh = 4
  END IF

  CurBuilding = 1

  ' --- Draw buildings left to right across the screen ---
  DO
    ' Adjust height based on slope pattern
    SELECT CASE Slope
      CASE 1
        NewHt = NewHt + HtInc
      CASE 2
        NewHt = NewHt - HtInc
      CASE 3 TO 5
        ' "V" pattern: increase until midpoint, then decrease
        IF x > ScrWidth \ 2 THEN
          NewHt = NewHt - 2 * HtInc
        ELSE
          NewHt = NewHt + 2 * HtInc
        END IF
      ' NOTE: CASE 4 below is unreachable (caught by CASE 3 TO 5 above).
      ' This is a bug in the original QBasic source, preserved for fidelity.
      CASE 4
        IF x > ScrWidth \ 2 THEN
          NewHt = NewHt + 2 * HtInc
        ELSE
          NewHt = NewHt - 2 * HtInc
        END IF
    END SELECT

    ' --- Calculate building dimensions ---
    BWidth = FnRan(DefBWidth) + DefBWidth
    IF x + BWidth > ScrWidth THEN BWidth = ScrWidth - x - 2

    BHeight = FnRan(RandomHeight) + NewHt
    IF BHeight < HtInc THEN BHeight = HtInc

    ' Prevent buildings from overlapping the gorilla/sun area
    IF BottomLine - BHeight <= MaxHeight + GHeight THEN BHeight = MaxHeight + GHeight - 5

    ' Store building coordinates for PlaceGorillas
    BCoor(CurBuilding).XCoor = x
    BCoor(CurBuilding).YCoor = BottomLine - BHeight

    ' Random building color (EGA: 3 colors, CGA: fixed)
    IF Mode = 9 THEN BuildingColor = FnRan(3) + 4 ELSE BuildingColor = 2

    ' Draw building outline then fill
    LINE (x - 1, BottomLine + 1)-(x + BWidth + 1, BottomLine - BHeight - 1), BACKATTR, B
    LINE (x, BottomLine)-(x + BWidth, BottomLine - BHeight), BuildingColor, BF

    ' --- Draw windows on the building ---
    c = x + 3
    DO
      FOR i = BHeight - 3 TO 7 STEP -WDifV
        ' Randomly light or darken windows
        IF Mode <> 9 THEN
          WinColr = (FnRan(2) - 2) * -3
        ELSEIF FnRan(4) = 1 THEN
          WinColr = 8             ' Dark window (25% chance)
        ELSE
          WinColr = WINDOWCOLOR   ' Lit window (75% chance)
        END IF
        LINE (c, BottomLine - i)-(c + WWidth, BottomLine - i + WHeight), WinColr, BF
      NEXT
      c = c + WDifh
    LOOP UNTIL c >= x + BWidth - 3

    x = x + BWidth + 2
    CurBuilding = CurBuilding + 1

  LOOP UNTIL x > ScrWidth - HtInc

  LastBuilding = CurBuilding - 1

  ' --- Generate random wind ---
  ' Base wind: -4 to +4.  33% chance of stronger gusts.
  Wind = FnRan(10) - 5
  IF FnRan(3) = 1 THEN
    IF Wind > 0 THEN
      Wind = Wind + FnRan(10)
    ELSE
      Wind = Wind - FnRan(10)
    END IF
  END IF

  ' --- Draw wind indicator arrow ---
  IF Wind <> 0 THEN
    WindLine = Wind * 3 * (ScrWidth \ 320)
    LINE (ScrWidth \ 2, ScrHeight - 5)-(ScrWidth \ 2 + WindLine, ScrHeight - 5), ExplosionColor
    IF Wind > 0 THEN ArrowDir = -2 ELSE ArrowDir = 2
    LINE (ScrWidth / 2 + WindLine, ScrHeight - 5)-(ScrWidth / 2 + WindLine + ArrowDir, ScrHeight - 5 - 2), ExplosionColor
    LINE (ScrWidth / 2 + WindLine, ScrHeight - 5)-(ScrWidth / 2 + WindLine + ArrowDir, ScrHeight - 5 + 2), ExplosionColor
  END IF
END SUB


' ============================================================================
' SUB PlaceGorillas (BCoor() AS XYPoint)
' ============================================================================
' Places gorilla sprites on rooftops near the left and right edges.
'
' TUTORIAL: Each gorilla is placed on the 2nd or 3rd building from its
' respective edge of the screen.  The X position is centered on the
' building width; the Y position is above the building top by the
' gorilla's height (YAdj).  PUT with PSET draws the pre-captured
' sprite array onto the screen.
'
' Parameters:
'   BCoor() - Building coordinate array (from MakeCityScape)
' ----------------------------------------------------------------------------
SUB PlaceGorillas (BCoor() AS XYPoint)
  IF Mode = 9 THEN
    XAdj = 14
    YAdj = 30
  ELSE
    XAdj = 7
    YAdj = 16
  END IF
  SclX# = ScrWidth / 320
  SclY# = ScrHeight / 200

  ' Place gorillas on buildings near opposite edges
  FOR i = 1 TO 2
    IF i = 1 THEN BNum = FnRan(2) + 1 ELSE BNum = LastBuilding - FnRan(2)

    BWidth = BCoor(BNum + 1).XCoor - BCoor(BNum).XCoor
    GorillaX(i) = BCoor(BNum).XCoor + BWidth / 2 - XAdj
    GorillaY(i) = BCoor(BNum).YCoor - YAdj
    PUT (GorillaX(i), GorillaY(i)), GorD&, PSET
  NEXT i
END SUB


' ============================================================================
' FUNCTION DoShot (PlayerNum, x, y)
' ============================================================================
' Handles one player's turn: accepts angle/velocity input, fires the
' banana, and determines if a gorilla was hit.
'
' TUTORIAL: This function orchestrates a single turn:
'   1. Display input prompts at the correct screen edge
'   2. Get angle and velocity from the player via GetNum#
'   3. Mirror the angle for player 2 (180 - angle)
'   4. Clear the input area
'   5. Call PlotShot to animate the trajectory
'   6. If a hit occurred, trigger VictoryDance
'
' Parameters:
'   PlayerNum - Current player (1 or 2)
'   x, y      - Gorilla position of the throwing player
' Returns:
'   TRUE if a gorilla was hit, FALSE otherwise
' ----------------------------------------------------------------------------
FUNCTION DoShot (PlayerNum, x, y)
  ' Position input prompts on the correct side of the screen
  IF PlayerNum = 1 THEN
    LocateCol = 1
  ELSE
    IF Mode = 9 THEN
      LocateCol = 66
    ELSE
      LocateCol = 26
    END IF
  END IF

  ' --- Get angle input ---
  LOCATE 2, LocateCol
  PRINT "Angle:";
  Angle# = GetNum#(2, LocateCol + 7)

  ' --- Get velocity input ---
  LOCATE 3, LocateCol
  PRINT "Velocity:";
  Velocity = GetNum#(3, LocateCol + 10)

  ' Mirror angle for player 2 (they face left)
  IF PlayerNum = 2 THEN
    Angle# = 180 - Angle#
  END IF

  ' --- Erase input text ---
  FOR i = 1 TO 4
    LOCATE i, 1
    PRINT SPACE$(30 \ (80 \ MaxCol));
    LOCATE i, (50 \ (80 \ MaxCol))
    PRINT SPACE$(30 \ (80 \ MaxCol));
  NEXT

  ' --- Fire the shot ---
  SunHit = FALSE
  PlayerHit = PlotShot(x, y, Angle#, Velocity, PlayerNum)

  IF PlayerHit = 0 THEN
    DoShot = FALSE            ' Miss -- no one hit
  ELSE
    DoShot = TRUE
    ' If the thrower hit themselves, switch to the other player for victory
    IF PlayerHit = PlayerNum THEN PlayerNum = 3 - PlayerNum
    CALL VictoryDance(PlayerNum)
  END IF
END FUNCTION


' ============================================================================
' FUNCTION PlotShot (StartX, StartY, Angle#, Velocity, PlayerNum)
' ============================================================================
' Animates the banana's parabolic trajectory and detects collisions.
'
' TUTORIAL: This is the physics engine of the game.  The banana follows
' a parabolic arc determined by:
'
'   x(t) = x0 + Vx*t + 0.5*wind*t^2    (horizontal: velocity + wind)
'   y(t) = y0 - Vy*t + 0.5*g*t^2        (vertical: velocity - gravity)
'
' where Vx = V*cos(angle), Vy = V*sin(angle), and t increments by 0.1.
' The negative sign on Vy accounts for the screen's inverted Y axis
' (0 at top, increasing downward).
'
' Collision detection uses POINT() to read the color of the pixel at
' the banana's position.  If it's not black (background), something
' was hit:
'   - OBJECTCOLOR (1) = gorilla hit -> ExplodeGorilla
'   - SUNATTR (3)     = sun hit -> cosmetic reaction only
'   - Any other color = building hit -> DoExplosion
'
' The banana rotates through 4 frames as it flies, creating a tumbling
' visual effect.  DrawBan handles both drawing and erasing.
'
' Parameters:
'   StartX, StartY - Gorilla position of the thrower
'   Angle#         - Launch angle in degrees (converted to radians here)
'   Velocity       - Launch speed
'   PlayerNum      - Which player is throwing (1 or 2)
' Returns:
'   0 if no gorilla hit, or the player number of the gorilla that was hit
' ----------------------------------------------------------------------------
FUNCTION PlotShot (StartX, StartY, Angle#, Velocity, PlayerNum)
  ' Convert angle from degrees to radians
  Angle# = Angle# / 180 * pi#
  Radius = Mode MOD 7

  ' Decompose velocity into horizontal and vertical components
  InitXVel# = COS(Angle#) * Velocity
  InitYVel# = SIN(Angle#) * Velocity

  oldx# = StartX
  oldy# = StartY

  ' --- Show gorilla throwing pose ---
  IF PlayerNum = 1 THEN
    PUT (StartX, StartY), GorL&, PSET    ' Left arm up (throwing right)
  ELSE
    PUT (StartX, StartY), GorR&, PSET    ' Right arm up (throwing left)
  END IF

  ' Throw sound effect
  PLAY "MBo0L32A-L64CL16BL64A+"
  CALL Rest(.1)

  ' Restore gorilla to neutral pose
  PUT (StartX, StartY), GorD&, PSET

  adjust = Scl(4)                         ' CGA/EGA coordinate adjustment

  xedge = Scl(9) * (2 - PlayerNum)       ' Leading edge of banana for checks

  ' --- Initialize trajectory state ---
  Impact = FALSE
  ShotInSun = FALSE
  OnScreen = TRUE
  PlayerHit = 0
  NeedErase = FALSE

  ' Calculate starting position (above gorilla's head)
  StartXPos = StartX
  StartYPos = StartY - adjust - 3

  IF PlayerNum = 2 THEN
    StartXPos = StartXPos + Scl(25)
    direction = Scl(4)
  ELSE
    direction = Scl(-4)
  END IF

  ' Handle zero/near-zero velocity (banana drops on self)
  IF Velocity < 2 THEN
    x# = StartX
    y# = StartY
    pointval = OBJECTCOLOR
  END IF

  ' =====================================================
  ' Main trajectory loop
  ' =====================================================
  DO WHILE (NOT Impact) AND OnScreen

    CALL Rest(.02)

    ' Erase the previous banana frame
    IF NeedErase THEN
      NeedErase = FALSE
      CALL DrawBan(oldx#, oldy#, oldrot, FALSE)
    END IF

    ' --- Calculate new position using projectile motion equations ---
    ' Horizontal: initial velocity + wind acceleration
    x# = StartXPos + (InitXVel# * t#) + (.5 * (Wind / 5) * t# ^ 2)
    ' Vertical: initial velocity (upward) + gravity (downward)
    y# = StartYPos + ((-1 * (InitYVel# * t#)) + (.5 * gravity# * t# ^ 2)) * (ScrHeight / 350)

    ' Check if banana has left the screen
    IF (x# >= ScrWidth - Scl(10)) OR (x# <= 3) OR (y# >= ScrHeight - 3) THEN
      OnScreen = FALSE
    END IF

    ' --- Collision detection (only while on screen and below top edge) ---
    IF OnScreen AND y# > 0 THEN
      LookY = 0
      LookX = Scl(8 * (2 - PlayerNum))

      DO
        pointval = POINT(x# + LookX, y# + LookY)
        IF pointval = 0 THEN
          ' Background -- no collision
          Impact = FALSE
          IF ShotInSun = TRUE THEN
            IF ABS(ScrWidth \ 2 - x#) > Scl(20) OR y# > SunHt THEN ShotInSun = FALSE
          END IF
        ELSEIF pointval = SUNATTR AND y# < SunHt THEN
          ' Hit the sun -- cosmetic reaction only
          IF NOT SunHit THEN CALL DoSun(SUNSHOCK)
          SunHit = TRUE
          ShotInSun = TRUE
        ELSE
          ' Hit a building or gorilla
          Impact = TRUE
        END IF
        LookX = LookX + direction
        LookY = LookY + Scl(6)
      LOOP UNTIL Impact OR LookX <> Scl(4)

      ' Draw banana if not inside the sun and not impacting
      IF NOT ShotInSun AND NOT Impact THEN
        rot = (t# * 10) MOD 4            ' Rotate through 4 frames
        CALL DrawBan(x#, y#, rot, TRUE)
        NeedErase = TRUE
      END IF

      oldx# = x#
      oldy# = y#
      oldrot = rot

    END IF

    t# = t# + .1                          ' Advance time step

  LOOP

  ' --- Handle impact ---
  IF pointval <> OBJECTCOLOR AND Impact THEN
    ' Hit a building -- show explosion
    CALL DoExplosion(x# + adjust, y# + adjust)
  ELSEIF pointval = OBJECTCOLOR THEN
    ' Hit a gorilla -- show gorilla explosion
    PlayerHit = ExplodeGorilla(x#, y#)
  END IF

  PlotShot = PlayerHit
END FUNCTION


' ============================================================================
' SUB VictoryDance (Player)
' ============================================================================
' Animates the winning gorilla waving its arms in celebration.
'
' TUTORIAL: The dance alternates between left-arm-up and right-arm-up
' sprites using PUT with PSET, accompanied by a short musical phrase.
' The MF prefix in PLAY means "Music Foreground" -- execution pauses
' until each note sequence finishes, synchronizing the animation with
' the music.
'
' Parameters:
'   Player - The winning player number (1 or 2)
' ----------------------------------------------------------------------------
SUB VictoryDance (Player)
  FOR i# = 1 TO 4
    PUT (GorillaX(Player), GorillaY(Player)), GorL&, PSET
    PLAY "MFO0L32EFGEFDC"
    CALL Rest(.2)
    PUT (GorillaX(Player), GorillaY(Player)), GorR&, PSET
    PLAY "MFO0L32EFGEFDC"
    CALL Rest(.2)
  NEXT
END SUB


' ============================================================================
' SUB UpdateScores (Record(), PlayerNum, Results)
' ============================================================================
' Updates the score array after a hit.
'
' TUTORIAL: If a player hit themselves (Results = HITSELF), the point
' goes to the OTHER player (3 - PlayerNum).  Otherwise, the throwing
' player gets the point.  ABS(PlayerNum - 3) is equivalent to
' (3 - PlayerNum) for values 1 and 2.
'
' Parameters:
'   Record()  - Score array (1 TO 2), passed by reference
'   PlayerNum - The player who threw the banana
'   Results   - Hit result (HITSELF or other)
' ----------------------------------------------------------------------------
SUB UpdateScores (Record(), PlayerNum, Results)
  IF Results = HITSELF THEN
    Record(ABS(PlayerNum - 3)) = Record(ABS(PlayerNum - 3)) + 1
  ELSE
    Record(PlayerNum) = Record(PlayerNum) + 1
  END IF
END SUB


' ============================================================================
' SUB PlayGame (Player1$, Player2$, NumGames)
' ============================================================================
' Main game loop: plays the specified number of rounds, alternating
' turns between players until a gorilla is hit each round.
'
' TUTORIAL: Each round consists of:
'   1. Clear screen and generate a new random cityscape
'   2. Place gorillas on rooftops
'   3. Draw the sun
'   4. Alternate turns (DoShot) until someone is hit
'   5. Update and display scores
' After all rounds, display the final score screen.
'
' The variable J alternates between 0 and 1 each turn (J = 1 - J),
' so Tosser cycles between player 1 and player 2.  Tossee is the
' opponent (3 - Tosser).
'
' Parameters:
'   Player1$, Player2$ - Player names
'   NumGames           - Total rounds to play
' ----------------------------------------------------------------------------
SUB PlayGame (Player1$, Player2$, NumGames)
  DIM BCoor(0 TO 30) AS XYPoint   ' Building coordinates
  DIM TotalWins(1 TO 2)            ' Running score

  J = 1

  FOR i = 1 TO NumGames

    CLS
    RANDOMIZE (TIMER)               ' Re-seed RNG each round

    ' Build the city and place combatants
    CALL MakeCityScape(BCoor())
    CALL PlaceGorillas(BCoor())
    CALL DoSun(SUNHAPPY)

    Hit = FALSE

    ' --- Turn loop: alternate shots until someone is hit ---
    DO WHILE Hit = FALSE
      J = 1 - J                     ' Toggle turn

      ' Display player names
      LOCATE 1, 1
      PRINT Player1$
      LOCATE 1, (MaxCol - 1 - LEN(Player2$))
      PRINT Player2$

      ' Display current scores
      CALL Center(23, LTRIM$(STR$(TotalWins(1))) + ">Score<" + LTRIM$(STR$(TotalWins(2))))

      Tosser = J + 1                ' Current thrower (1 or 2)
      Tossee = 3 - J                ' Opponent

      ' Fire! Returns TRUE if a gorilla was hit.
      Hit = DoShot(Tosser, GorillaX(Tosser), GorillaY(Tosser))

      ' Reset the sun if it was hit during this shot
      IF SunHit THEN CALL DoSun(SUNHAPPY)

      ' Update scores on a hit
      IF Hit = TRUE THEN CALL UpdateScores(TotalWins(), Tosser, Hit)
    LOOP

    ' Brief pause between rounds (Turbo Basic: use Rest instead of SLEEP)
    CALL Rest(1)
  NEXT i

  ' --- Game Over screen ---
  SCREEN 0
  WIDTH 80, 25
  COLOR 7, 0
  MaxCol = 80
  CLS

  CALL Center(8, "GAME OVER!")
  CALL Center(10, "Score:")
  LOCATE 11, 30: PRINT Player1$; TAB(50); TotalWins(1)
  LOCATE 12, 30: PRINT Player2$; TAB(50); TotalWins(2)
  CALL Center(24, "Press any key to continue")
  CALL SparklePause
  COLOR 7, 0
  CLS
END SUB
