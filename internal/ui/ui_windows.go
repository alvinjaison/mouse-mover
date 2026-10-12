//go:build windows

package ui

import (
	"fmt"
	"runtime"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/jarvis/mouse-mover/internal/config"
	"github.com/jarvis/mouse-mover/internal/engine"
)

// ---- Win32 constants ----

const (
	wsOverlappedWindow = 0x00CF0000
	wsVisible          = 0x10000000
	wsChild            = 0x40000000
	wsTabStop          = 0x00010000
	wsVScroll          = 0x00200000
	wsExClientEdge     = 0x00000200
	bsDefPushButton    = 0x00000001
	bsPushButton       = 0x00000000
	bsAutoCheckBox     = 0x00000003
	cbsDropDownList    = 0x0003
	esNumber           = 0x2000
	esAutoHScroll      = 0x0080
	ssLeft             = 0x0000
	ssEtchFrame        = 0x00000012
	wmDestroy          = 0x0002
	wmClose            = 0x0010
	wmCommand          = 0x0111
	wmTimer            = 0x0113
	wmSysCommand       = 0x0112
	wmApp              = 0x8000
	scMinimize         = 0xF020
	swHide             = 0
	swShow             = 5
	swRestore          = 9
	cbAddString        = 0x0143
	cbSetCurSel        = 0x014E
	cbGetCurSel        = 0x0147
	wmGetText          = 0x000D
	wmGetTextLength    = 0x000E
	wmSetText          = 0x000C
	bmGetCheck         = 0x00F0
	bmSetCheck         = 0x00F1
	bstChecked         = 1
	bstUnchecked       = 0
	mfString           = 0x0000
	mfGrayed           = 0x0001
	tpmBottomAlign     = 0x0020
	tpmRightButton     = 0x0002
	nifMessage         = 0x00000001
	nifIcon            = 0x00000002
	nifTip             = 0x00000004
	nimAdd             = 0x00000000
	nimModify          = 0x00000001
	nimDelete          = 0x00000002
	wmLButtonDblClk    = 0x0203
	wmRButtonUp        = 0x0205
	idiApplication     = 32512
	colorBtnFace       = 15
)

// control IDs
const (
	idStartBtn     = 101
	idStopBtn      = 102
	idPatternCB    = 103
	idDurCheck     = 110
	idDurCB        = 111
	idDurHrsEdit   = 112
	idDurMinEdit   = 113
	idIntCheck     = 120
	idIntCB        = 121
	idIntSecEdit   = 122
	idTimerUpdate  = 1
	idTrayMsg      = wmApp + 1
	idTrayShow     = 201
	idTrayStart    = 202
	idTrayStop     = 203
	idTrayExit     = 204
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	shell32  = windows.NewLazySystemDLL("shell32.dll")

	pRegisterClassEx  = user32.NewProc("RegisterClassExW")
	pCreateWindowEx   = user32.NewProc("CreateWindowExW")
	pDefWindowProc    = user32.NewProc("DefWindowProcW")
	pGetMessage       = user32.NewProc("GetMessageW")
	pTranslateMessage = user32.NewProc("TranslateMessage")
	pDispatchMessage  = user32.NewProc("DispatchMessageW")
	pPostQuitMessage  = user32.NewProc("PostQuitMessage")
	pShowWindow       = user32.NewProc("ShowWindow")
	pSetForegroundWnd = user32.NewProc("SetForegroundWindow")
	pDestroyWindow    = user32.NewProc("DestroyWindow")
	pSendMessage      = user32.NewProc("SendMessageW")
	pSetTimer         = user32.NewProc("SetTimer")
	pKillTimer        = user32.NewProc("KillTimer")
	pLoadIcon         = user32.NewProc("LoadIconW")
	pLoadCursor       = user32.NewProc("LoadCursorW")
	pCreatePopupMenu  = user32.NewProc("CreatePopupMenu")
	pAppendMenu       = user32.NewProc("AppendMenuW")
	pTrackPopupMenu   = user32.NewProc("TrackPopupMenu")
	pDestroyMenu      = user32.NewProc("DestroyMenu")
	pGetCursorPos2    = user32.NewProc("GetCursorPos")
	pEnableWindow     = user32.NewProc("EnableWindow")
	pMessageBox       = user32.NewProc("MessageBoxW")
	pGetModuleHandle  = kernel32.NewProc("GetModuleHandleW")
	pShellNotifyIcon  = shell32.NewProc("Shell_NotifyIconW")
)

// ---- Win32 structs ----

type wndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  uintptr
	LpszClassName uintptr
	HIconSm      uintptr
}

type msg struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

type notifyIconData struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
}

// ---- globals ----

var (
	gEng     *engine.Engine
	gCfg     config.Config
	gSaveCfg func(config.Config)

	gMainWnd uintptr
	gHIcon   uintptr

	// Status
	gStatusLbl    uintptr
	gLastMoveLbl  uintptr
	gElapsedLbl   uintptr
	gRemainingLbl uintptr

	// Pattern
	gPatternCB uintptr

	// Duration section
	gDurCheck  uintptr
	gDurCB     uintptr
	gDurHrsEd  uintptr
	gDurMinEd  uintptr
	gDurHrsLbl uintptr
	gDurMinLbl uintptr

	// Interval section
	gIntCheck  uintptr
	gIntCB     uintptr
	gIntSecEd  uintptr
	gIntSecLbl uintptr

	// Buttons
	gStartBtn uintptr
	gStopBtn  uintptr

	gNID notifyIconData
)

// Presets
var durPresets = []struct {
	label   string
	minutes int
}{
	{"30 minutes", 30},
	{"1 hour", 60},
	{"2 hours", 120},
	{"4 hours", 240},
	{"8 hours", 480},
}

var intPresets = []struct {
	label   string
	seconds int
}{
	{"10 seconds", 10},
	{"30 seconds", 30},
	{"1 minute", 60},
	{"2 minutes", 120},
	{"5 minutes", 300},
}

// ---- helpers ----

func mustUTF16Ptr(s string) uintptr {
	p, _ := syscall.UTF16PtrFromString(s)
	return uintptr(unsafe.Pointer(p))
}

func sendMsg(hwnd uintptr, m uint32, wp, lp uintptr) uintptr {
	r, _, _ := pSendMessage.Call(hwnd, uintptr(m), wp, lp)
	return r
}

func cbAddStr(hwnd uintptr, s string)     { sendMsg(hwnd, cbAddString, 0, mustUTF16Ptr(s)) }
func setWndText(hwnd uintptr, s string)   { sendMsg(hwnd, wmSetText, 0, mustUTF16Ptr(s)) }

func getWndText(hwnd uintptr) string {
	n := int(sendMsg(hwnd, wmGetTextLength, 0, 0))
	if n <= 0 {
		return ""
	}
	buf := make([]uint16, n+1)
	sendMsg(hwnd, wmGetText, uintptr(n+1), uintptr(unsafe.Pointer(&buf[0])))
	return syscall.UTF16ToString(buf)
}

func enableWnd(hwnd uintptr, on bool) {
	v := uintptr(0)
	if on {
		v = 1
	}
	pEnableWindow.Call(hwnd, v)
}

func showWnd(hwnd uintptr, on bool) {
	cmd := uintptr(swHide)
	if on {
		cmd = uintptr(swShow)
	}
	pShowWindow.Call(hwnd, cmd)
}

func isChecked(hwnd uintptr) bool {
	return sendMsg(hwnd, bmGetCheck, 0, 0) == bstChecked
}

func setChecked(hwnd uintptr, on bool) {
	v := uintptr(bstUnchecked)
	if on {
		v = bstChecked
	}
	sendMsg(hwnd, bmSetCheck, v, 0)
}

// ---- Run ----

func Run(cfg config.Config, eng *engine.Engine, saveCfg func(config.Config)) error {
	runtime.LockOSThread()

	gEng = eng
	gCfg = cfg
	gSaveCfg = saveCfg

	hInst, _, _ := pGetModuleHandle.Call(0)
	gHIcon, _, _ = pLoadIcon.Call(0, uintptr(idiApplication))
	hCursor, _, _ := pLoadCursor.Call(0, uintptr(32512))

	className := mustUTF16Ptr("MouseMoverWndClass")
	wc := wndClassEx{
		CbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		Style:         0x0003,
		LpfnWndProc:   windows.NewCallback(wndProc),
		HInstance:     hInst,
		HIcon:         gHIcon,
		HCursor:       hCursor,
		HbrBackground: colorBtnFace + 1,
		LpszClassName: className,
		HIconSm:      gHIcon,
	}

	r, _, err := pRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if r == 0 {
		return fmt.Errorf("RegisterClassEx: %w", err)
	}

	gMainWnd, _, err = pCreateWindowEx.Call(
		0, className, mustUTF16Ptr("Mouse Mover"),
		uintptr(wsOverlappedWindow),
		100, 100, 470, 530,
		0, 0, hInst, 0,
	)
	if gMainWnd == 0 {
		return fmt.Errorf("CreateWindowEx: %w", err)
	}

	createControls(hInst)
	applyConfig()
	addTrayIcon()

	eng.SetOnStopped(func() {
		pSendMessage.Call(gMainWnd, wmTimer, idTimerUpdate, 0)
	})

	if !cfg.StartMinimized {
		pShowWindow.Call(gMainWnd, uintptr(swShow))
	}
	pSetTimer.Call(gMainWnd, idTimerUpdate, 1000, 0)

	var m msg
	for {
		r, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if r == 0 || int32(r) == -1 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}
	return nil
}

func createControls(hInst uintptr) {
	btnCls := mustUTF16Ptr("BUTTON")
	stCls := mustUTF16Ptr("STATIC")
	cbCls := mustUTF16Ptr("COMBOBOX")
	edCls := mustUTF16Ptr("EDIT")

	y := int32(10)

	// Title
	pCreateWindowEx.Call(0, stCls, mustUTF16Ptr("Mouse Mover"),
		uintptr(wsChild|wsVisible|ssLeft), 20, uintptr(y), 400, 24, gMainWnd, 0, hInst, 0)
	y += 30

	// Status
	gStatusLbl, _, _ = pCreateWindowEx.Call(0, stCls, mustUTF16Ptr("Status: Stopped"),
		uintptr(wsChild|wsVisible|ssLeft), 20, uintptr(y), 400, 20, gMainWnd, 0, hInst, 0)
	y += 30

	// Pattern
	pCreateWindowEx.Call(0, stCls, mustUTF16Ptr("Pattern:"),
		uintptr(wsChild|wsVisible|ssLeft), 20, uintptr(y+3), 60, 20, gMainWnd, 0, hInst, 0)
	gPatternCB, _, _ = pCreateWindowEx.Call(0, cbCls, 0,
		uintptr(wsChild|wsVisible|wsTabStop|wsVScroll|cbsDropDownList),
		85, uintptr(y), 130, 150, gMainWnd, idPatternCB, hInst, 0)
	cbAddStr(gPatternCB, "Nudge")
	cbAddStr(gPatternCB, "Jiggle")
	cbAddStr(gPatternCB, "Square")
	sendMsg(gPatternCB, cbSetCurSel, 0, 0)
	y += 35

	// ---- Custom Duration section ----
	gDurCheck, _, _ = pCreateWindowEx.Call(0, btnCls, mustUTF16Ptr("Custom Duration (auto-stop after)"),
		uintptr(wsChild|wsVisible|wsTabStop|bsAutoCheckBox),
		20, uintptr(y), 250, 20, gMainWnd, idDurCheck, hInst, 0)
	y += 25

	// Duration combo (presets)
	gDurCB, _, _ = pCreateWindowEx.Call(0, cbCls, 0,
		uintptr(wsChild|wsTabStop|wsVScroll|cbsDropDownList),
		40, uintptr(y), 130, 200, gMainWnd, idDurCB, hInst, 0)
	for _, p := range durPresets {
		cbAddStr(gDurCB, p.label)
	}
	cbAddStr(gDurCB, "Custom...")
	sendMsg(gDurCB, cbSetCurSel, 1, 0) // 1 hour

	// Custom hrs + min edits
	gDurHrsEd, _, _ = pCreateWindowEx.Call(uintptr(wsExClientEdge), edCls, mustUTF16Ptr("1"),
		uintptr(wsChild|wsTabStop|esNumber|esAutoHScroll),
		180, uintptr(y+1), 40, 22, gMainWnd, idDurHrsEdit, hInst, 0)
	gDurHrsLbl, _, _ = pCreateWindowEx.Call(0, stCls, mustUTF16Ptr("hrs"),
		uintptr(wsChild|ssLeft), 224, uintptr(y+3), 25, 20, gMainWnd, 0, hInst, 0)
	gDurMinEd, _, _ = pCreateWindowEx.Call(uintptr(wsExClientEdge), edCls, mustUTF16Ptr("0"),
		uintptr(wsChild|wsTabStop|esNumber|esAutoHScroll),
		255, uintptr(y+1), 40, 22, gMainWnd, idDurMinEdit, hInst, 0)
	gDurMinLbl, _, _ = pCreateWindowEx.Call(0, stCls, mustUTF16Ptr("min"),
		uintptr(wsChild|ssLeft), 299, uintptr(y+3), 25, 20, gMainWnd, 0, hInst, 0)
	y += 35

	// ---- Custom Interval section ----
	gIntCheck, _, _ = pCreateWindowEx.Call(0, btnCls, mustUTF16Ptr("Custom Interval (move every)"),
		uintptr(wsChild|wsVisible|wsTabStop|bsAutoCheckBox),
		20, uintptr(y), 250, 20, gMainWnd, idIntCheck, hInst, 0)
	y += 25

	// Interval combo (presets)
	gIntCB, _, _ = pCreateWindowEx.Call(0, cbCls, 0,
		uintptr(wsChild|wsTabStop|wsVScroll|cbsDropDownList),
		40, uintptr(y), 130, 200, gMainWnd, idIntCB, hInst, 0)
	for _, p := range intPresets {
		cbAddStr(gIntCB, p.label)
	}
	cbAddStr(gIntCB, "Custom...")
	sendMsg(gIntCB, cbSetCurSel, 1, 0) // 30 seconds

	// Custom seconds edit
	gIntSecEd, _, _ = pCreateWindowEx.Call(uintptr(wsExClientEdge), edCls, mustUTF16Ptr("30"),
		uintptr(wsChild|wsTabStop|esNumber|esAutoHScroll),
		180, uintptr(y+1), 50, 22, gMainWnd, idIntSecEdit, hInst, 0)
	gIntSecLbl, _, _ = pCreateWindowEx.Call(0, stCls, mustUTF16Ptr("sec"),
		uintptr(wsChild|ssLeft), 234, uintptr(y+3), 25, 20, gMainWnd, 0, hInst, 0)
	y += 40

	// ---- Buttons ----
	gStartBtn, _, _ = pCreateWindowEx.Call(0, btnCls, mustUTF16Ptr("Start"),
		uintptr(wsChild|wsVisible|wsTabStop|bsDefPushButton),
		20, uintptr(y), 100, 35, gMainWnd, idStartBtn, hInst, 0)
	gStopBtn, _, _ = pCreateWindowEx.Call(0, btnCls, mustUTF16Ptr("Stop"),
		uintptr(wsChild|wsVisible|wsTabStop|bsPushButton),
		140, uintptr(y), 100, 35, gMainWnd, idStopBtn, hInst, 0)
	enableWnd(gStopBtn, false)
	y += 50

	// ---- Info labels ----
	gLastMoveLbl, _, _ = pCreateWindowEx.Call(0, stCls, mustUTF16Ptr("Last move: \u2014"),
		uintptr(wsChild|wsVisible|ssLeft), 20, uintptr(y), 400, 20, gMainWnd, 0, hInst, 0)
	y += 25
	gElapsedLbl, _, _ = pCreateWindowEx.Call(0, stCls, mustUTF16Ptr("Elapsed: \u2014"),
		uintptr(wsChild|wsVisible|ssLeft), 20, uintptr(y), 400, 20, gMainWnd, 0, hInst, 0)
	y += 25
	gRemainingLbl, _, _ = pCreateWindowEx.Call(0, stCls, mustUTF16Ptr("Remaining: \u2014"),
		uintptr(wsChild|wsVisible|ssLeft), 20, uintptr(y), 400, 20, gMainWnd, 0, hInst, 0)
}

func applyConfig() {
	// Pattern
	patIdx := 0
	switch gCfg.Pattern {
	case "jiggle":
		patIdx = 1
	case "square":
		patIdx = 2
	}
	sendMsg(gPatternCB, cbSetCurSel, uintptr(patIdx), 0)

	// Duration: check if non-default (default=0 means indefinite)
	if gCfg.DurationMinutes > 0 {
		setChecked(gDurCheck, true)
		idx := len(durPresets) // Custom
		for i, p := range durPresets {
			if gCfg.DurationMinutes == p.minutes {
				idx = i
				break
			}
		}
		sendMsg(gDurCB, cbSetCurSel, uintptr(idx), 0)
		if idx == len(durPresets) {
			setWndText(gDurHrsEd, strconv.Itoa(gCfg.DurationMinutes/60))
			setWndText(gDurMinEd, strconv.Itoa(gCfg.DurationMinutes%60))
		}
	}

	// Interval: check if non-default
	if gCfg.MoveIntervalSeconds > 0 && gCfg.MoveIntervalSeconds != config.DefaultMoveIntervalS {
		setChecked(gIntCheck, true)
		idx := len(intPresets) // Custom
		for i, p := range intPresets {
			if gCfg.MoveIntervalSeconds == p.seconds {
				idx = i
				break
			}
		}
		sendMsg(gIntCB, cbSetCurSel, uintptr(idx), 0)
		if idx == len(intPresets) {
			setWndText(gIntSecEd, strconv.Itoa(gCfg.MoveIntervalSeconds))
		}
	}

	updateSectionVisibility()
}

func updateSectionVisibility() {
	durOn := isChecked(gDurCheck)
	showWnd(gDurCB, durOn)
	durCustom := durOn && int(sendMsg(gDurCB, cbGetCurSel, 0, 0)) == len(durPresets)
	showWnd(gDurHrsEd, durCustom)
	showWnd(gDurHrsLbl, durCustom)
	showWnd(gDurMinEd, durCustom)
	showWnd(gDurMinLbl, durCustom)

	intOn := isChecked(gIntCheck)
	showWnd(gIntCB, intOn)
	intCustom := intOn && int(sendMsg(gIntCB, cbGetCurSel, 0, 0)) == len(intPresets)
	showWnd(gIntSecEd, intCustom)
	showWnd(gIntSecLbl, intCustom)
}

func readUIConfig() (config.Config, error) {
	// Duration
	durationMin := 0 // 0 = indefinite
	if isChecked(gDurCheck) {
		sel := int(sendMsg(gDurCB, cbGetCurSel, 0, 0))
		if sel >= 0 && sel < len(durPresets) {
			durationMin = durPresets[sel].minutes
		} else {
			h, err1 := strconv.Atoi(getWndText(gDurHrsEd))
			m, err2 := strconv.Atoi(getWndText(gDurMinEd))
			if err1 != nil || err2 != nil {
				return config.Config{}, fmt.Errorf("enter valid hours and minutes for duration")
			}
			durationMin = h*60 + m
			if durationMin < 1 {
				return config.Config{}, fmt.Errorf("custom duration must be at least 1 minute")
			}
		}
	}

	// Interval
	intervalSec := config.DefaultMoveIntervalS
	if isChecked(gIntCheck) {
		sel := int(sendMsg(gIntCB, cbGetCurSel, 0, 0))
		if sel >= 0 && sel < len(intPresets) {
			intervalSec = intPresets[sel].seconds
		} else {
			v, err := strconv.Atoi(getWndText(gIntSecEd))
			if err != nil || v < config.MinMoveIntervalSecs {
				return config.Config{}, fmt.Errorf("enter a valid interval (%d-%d seconds)", config.MinMoveIntervalSecs, config.MaxMoveIntervalSecs)
			}
			intervalSec = v
		}
	}

	// Pattern
	patSel := int(sendMsg(gPatternCB, cbGetCurSel, 0, 0))
	patterns := []string{"nudge", "jiggle", "square"}
	pat := "nudge"
	if patSel >= 0 && patSel < len(patterns) {
		pat = patterns[patSel]
	}

	cfg := config.Config{
		DurationMinutes:     durationMin,
		MoveIntervalSeconds: intervalSec,
		Pattern:             pat,
		DeltaPixels:         gCfg.DeltaPixels,
		StartMinimized:      gCfg.StartMinimized,
		AutoStart:           gCfg.AutoStart,
	}
	if cfg.DeltaPixels == 0 {
		cfg.DeltaPixels = 2
	}

	if err := cfg.Validate(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func doStart() {
	cfg, err := readUIConfig()
	if err != nil {
		showError(err.Error())
		return
	}
	gCfg = cfg
	if gSaveCfg != nil {
		gSaveCfg(cfg)
	}

	pat, _ := engine.ParsePattern(cfg.Pattern)
	gEng.Start(cfg.Duration(), cfg.MoveInterval(), pat, int32(cfg.DeltaPixels))
	updateStatus()
}

func doStop() {
	gEng.Stop()
	updateStatus()
}

func updateStatus() {
	s := gEng.Status()
	if s.Running {
		setWndText(gStatusLbl, "Status: Running")
		enableWnd(gStartBtn, false)
		enableWnd(gStopBtn, true)
	} else {
		setWndText(gStatusLbl, "Status: Stopped")
		enableWnd(gStartBtn, true)
		enableWnd(gStopBtn, false)
	}

	if !s.LastMove.IsZero() {
		setWndText(gLastMoveLbl, "Last move: "+s.LastMove.Format("15:04:05"))
	} else {
		setWndText(gLastMoveLbl, "Last move: \u2014")
	}

	if s.Running && s.Elapsed > 0 {
		setWndText(gElapsedLbl, "Elapsed: "+fmtDuration(s.Elapsed))
	} else {
		setWndText(gElapsedLbl, "Elapsed: \u2014")
	}

	if s.Running && s.Remaining > 0 {
		setWndText(gRemainingLbl, "Remaining: "+fmtDuration(s.Remaining))
	} else if s.Running {
		setWndText(gRemainingLbl, "Remaining: indefinite")
	} else {
		setWndText(gRemainingLbl, "Remaining: \u2014")
	}

	updateTrayTip(s.Running)
}

func fmtDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %02dm %02ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func showError(text string) {
	pMessageBox.Call(gMainWnd, mustUTF16Ptr(text), mustUTF16Ptr("Mouse Mover"), 0x10)
}

// ---- Tray ----

func addTrayIcon() {
	gNID = notifyIconData{
		CbSize:           uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:             gMainWnd,
		UID:              1,
		UFlags:           nifMessage | nifIcon | nifTip,
		UCallbackMessage: idTrayMsg,
		HIcon:            gHIcon,
	}
	copy(gNID.SzTip[:], utf16FromString("Mouse Mover \u2014 Stopped"))
	pShellNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(&gNID)))
}

func updateTrayTip(running bool) {
	tip := "Mouse Mover \u2014 Stopped"
	if running {
		tip = "Mouse Mover \u2014 Running"
	}
	copy(gNID.SzTip[:], utf16FromString(tip))
	gNID.UFlags = nifTip
	pShellNotifyIcon.Call(nimModify, uintptr(unsafe.Pointer(&gNID)))
}

func removeTrayIcon() {
	pShellNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(&gNID)))
}

func showTrayMenu() {
	hMenu, _, _ := pCreatePopupMenu.Call()
	if hMenu == 0 {
		return
	}
	defer pDestroyMenu.Call(hMenu)

	s := gEng.Status()
	sf := uintptr(mfString)
	stf := uintptr(mfString)
	if s.Running {
		sf |= mfGrayed
	} else {
		stf |= mfGrayed
	}

	pAppendMenu.Call(hMenu, uintptr(mfString), idTrayShow, mustUTF16Ptr("Show"))
	pAppendMenu.Call(hMenu, sf, idTrayStart, mustUTF16Ptr("Start"))
	pAppendMenu.Call(hMenu, stf, idTrayStop, mustUTF16Ptr("Stop"))
	pAppendMenu.Call(hMenu, uintptr(mfString), idTrayExit, mustUTF16Ptr("Exit"))

	pSetForegroundWnd.Call(gMainWnd)
	var pt struct{ X, Y int32 }
	pGetCursorPos2.Call(uintptr(unsafe.Pointer(&pt)))
	pTrackPopupMenu.Call(hMenu, uintptr(tpmBottomAlign|tpmRightButton),
		uintptr(pt.X), uintptr(pt.Y), 0, gMainWnd, 0)
}

func utf16FromString(s string) []uint16 {
	u, _ := syscall.UTF16FromString(s)
	return u
}

// ---- Window proc ----

func wndProc(hwnd uintptr, umsg uint32, wParam, lParam uintptr) uintptr {
	switch umsg {
	case wmCommand:
		id := int(wParam & 0xFFFF)
		switch id {
		case idStartBtn, idTrayStart:
			doStart()
			return 0
		case idStopBtn, idTrayStop:
			doStop()
			return 0
		case idTrayShow:
			pShowWindow.Call(gMainWnd, uintptr(swRestore))
			pSetForegroundWnd.Call(gMainWnd)
			return 0
		case idTrayExit:
			gEng.Stop()
			removeTrayIcon()
			pDestroyWindow.Call(gMainWnd)
			return 0
		case idDurCheck, idIntCheck, idDurCB, idIntCB:
			updateSectionVisibility()
			return 0
		}

	case wmTimer:
		if wParam == idTimerUpdate {
			updateStatus()
		}
		return 0

	case wmSysCommand:
		if wParam&0xFFF0 == scMinimize {
			pShowWindow.Call(hwnd, uintptr(swHide))
			return 0
		}

	case wmClose:
		// Actually quit: stop engine, remove tray, destroy window.
		gEng.Stop()
		removeTrayIcon()
		pDestroyWindow.Call(hwnd)
		return 0

	case idTrayMsg:
		switch lParam {
		case wmLButtonDblClk:
			pShowWindow.Call(gMainWnd, uintptr(swRestore))
			pSetForegroundWnd.Call(gMainWnd)
		case wmRButtonUp:
			showTrayMenu()
		}
		return 0

	case wmDestroy:
		pKillTimer.Call(hwnd, idTimerUpdate)
		gEng.Stop()
		removeTrayIcon()
		pPostQuitMessage.Call(0)
		return 0
	}

	r, _, _ := pDefWindowProc.Call(hwnd, uintptr(umsg), wParam, lParam)
	return r
}
