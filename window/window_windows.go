package window

/*
#include <windows.h>
*/
import "C"
import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
	"github.com/zhuwei/go-sciter"
)

var isShowWorkArea bool = false //use for no frame form

func New(creationFlags sciter.WindowCreationFlag, rect *sciter.Rect) (*Window, error) {
	w := new(Window)
	w.creationFlags = creationFlags

	// Initialize OLE for DnD and printing support
	win.OleInitialize()

	// create window
	hwnd := sciter.CreateWindow(
		creationFlags,
		rect,
		syscall.NewCallback(delegateProc),
		0,
		sciter.BAD_HWINDOW)

	if hwnd == sciter.BAD_HWINDOW {
		return nil, fmt.Errorf("Sciter CreateWindow failed [%d]", win.GetLastError())
	}

	w.Sciter = sciter.Wrap(hwnd)
	return w, nil
}

func NewCenter(creationFlags sciter.WindowCreationFlag, width int32, height int32, showWorkArea bool) (*Window, error) {
	w := new(Window)
	w.creationFlags = creationFlags

	// Initialize OLE for DnD and printing support
	win.OleInitialize()

	//center
	left, top := getFormCenterPoint(width, height)
	rect := &sciter.Rect{left, top, left + width, top + height}

	// create window
	hwnd := sciter.CreateWindow(
		creationFlags,
		rect,
		syscall.NewCallback(delegateProc),
		0,
		sciter.BAD_HWINDOW)

	if hwnd == sciter.BAD_HWINDOW {
		return nil, fmt.Errorf("Sciter CreateWindow failed [%d]", win.GetLastError())
	}

	w.Sciter = sciter.Wrap(hwnd)

	isShowWorkArea = showWorkArea

	return w, nil
}

func (s *Window) Show() {
	// message handling
	hwnd := win.HWND(unsafe.Pointer(s.GetHwnd()))
	win.ShowWindow(hwnd, win.SW_SHOW)
	win.UpdateWindow(hwnd)
}

func (s *Window) SetTitle(title string) {
	// message handling
	hwnd := C.HWND(unsafe.Pointer(s.GetHwnd()))
	C.SetWindowTextW(hwnd, (*C.WCHAR)(unsafe.Pointer(sciter.StringToWcharPtr(title))))
}

func (s *Window) Run() {
	// for system drag-n-drop
	// win.OleInitialize()
	// defer win.OleUninitialize()
	s.run()
	// start main gui message loop
	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) > 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}
}

func (s *Window) RunWithHandler(handler func(*Window)) {
	// for system drag-n-drop
	// win.OleInitialize()
	// defer win.OleUninitialize()
	s.run()
	// start main gui message loop
	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) > 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
		handler(s)
	}
}

func getFormCenterPoint(w, h int32) (x int32, y int32) {

	rect := &win.RECT{0, 0, 0, 0}

	if win.SystemParametersInfo(win.SPI_GETWORKAREA, 0, unsafe.Pointer(rect), 0) {
		x = (rect.Right-rect.Left-w)/2 + rect.Left
		y = (rect.Bottom-rect.Top-h)/2 + rect.Top
	} else {
		x = 0
		y = 0
	}
	return x, y
}

// delegate Windows GUI messsage
func delegateProc(hWnd win.HWND, message uint, wParam uintptr, lParam uintptr, pParam uintptr, pHandled *int) int {
	switch message {
	case win.WM_DESTROY:
		// log.Println("closing window ...")
		win.PostQuitMessage(0)
		*pHandled = 1

	case win.WM_SIZE:
		//fmt.Println("WM_SIZE ...")

	case win.WM_SIZING:
		//fmt.Println("WM_SIZING ...")

	case win.WM_GETMINMAXINFO:
		//fmt.Println("WM_GETMINMAXINFO ...")

		if isShowWorkArea {
			nStyle := int64(win.GetWindowLong(hWnd, win.GWL_STYLE))
			//fmt.Println(nStyle & win.WS_CAPTION) //WS_CAPTION  WS_POPUP
			if (nStyle & win.WS_POPUP) != 0 {

				rect := &win.RECT{0, 0, 0, 0}
				win.GetWindowRect(hWnd, rect)

				r := &win.RECT{0, 0, 0, 0}

				if win.SystemParametersInfo(win.SPI_GETWORKAREA, 0, unsafe.Pointer(r), 0) {
					info := (*win.MINMAXINFO)(unsafe.Pointer(lParam))
					info.PtMinTrackSize.X = rect.Right - rect.Left
					info.PtMinTrackSize.Y = rect.Bottom - rect.Top

					info.PtMaxSize.X = r.Right - r.Left
					info.PtMaxSize.Y = r.Bottom - r.Top
					info.PtMaxPosition.X = r.Left
					info.PtMaxPosition.Y = r.Top
				}
			}
		}
	}
	return 0
}
