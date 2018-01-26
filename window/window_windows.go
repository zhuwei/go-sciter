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

func GetSystemMetrics() (int32, int32) {
	return win.GetSystemMetrics(win.SM_CXSCREEN), win.GetSystemMetrics(win.SM_CYSCREEN)
}

func GetWorkArea() (int32, int32, int32, int32) {
	s := "Shell_TrayWnd"
	hWnd := win.FindWindow(StrInt(s), nil)
	rect := &win.RECT{0, 0, 0, 0}
	win.GetWindowRect(hWnd, rect)
	return rect.Left, rect.Top, rect.Right, rect.Bottom
}

func StrPtr(s string) uintptr {
	return uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(s)))
}

func StrInt(s string) *uint16 {
	return (*uint16)(unsafe.Pointer(syscall.StringToUTF16Ptr(s)))
}

func getFormCenterPoint(w, h int32) (int32, int32) {

	left, top, right, bottom := GetWorkArea()
	screenW, screenH := GetSystemMetrics()

	var screenLeft int32
	var screenTop int32
	screenRight := screenW
	screenBottom := screenH
	//除开任务栏宽高
	if left == 0 && top == 0 && bottom == screenH {
		screenLeft = right //居左
	} else if left == 0 && top == 0 && right == screenW {
		screenTop = bottom //居上
	} else if top == 0 && right == screenW && bottom == screenH {
		screenRight = left //居右
	} else if left == 0 && right == screenW && bottom == screenH {
		screenBottom = top //居下
	}

	//计算非任务栏区域的中心坐标
	return (screenRight-screenLeft-w)/2 + screenLeft, (screenBottom-screenTop-h)/2 + screenTop
}

// delegate Windows GUI messsage
func delegateProc(hWnd win.HWND, message uint, wParam uintptr, lParam uintptr, pParam uintptr, pHandled *int) int {
	switch message {
	case win.WM_DESTROY:
		// log.Println("closing window ...")
		win.PostQuitMessage(0)
		*pHandled = 1
	case win.WM_SIZE:
		fmt.Println("WM_SIZE ...")

		if isShowWorkArea {
			nStyle := int64(win.GetWindowLong(hWnd, win.GWL_STYLE))
			//fmt.Println(nStyle & win.WS_CAPTION) //WS_CAPTION  WS_POPUP
			//判断窗体样式是不是无边框
			if (nStyle & win.WS_POPUP) != 0 {

				screenW, screenH := GetSystemMetrics()

				rect := &win.RECT{0, 0, 0, 0}
				win.GetWindowRect(hWnd, rect)

				//判断窗体是否全屏
				if rect.Left == 0 && rect.Top == 0 && screenW == rect.Right && screenH == rect.Bottom {

					var screenLeft int32
					var screenTop int32
					screenRight := screenW
					screenBottom := screenH

					left, top, right, bottom := GetWorkArea()
					if left == 0 && top == 0 && bottom == screenH {
						screenLeft = right //居左
					} else if left == 0 && top == 0 && right == screenW {
						screenTop = bottom //居上
					} else if top == 0 && right == screenW && bottom == screenH {
						screenRight = left //居右
					} else if left == 0 && right == screenW && bottom == screenH {
						screenBottom = top //居下
					}
					//修改坐标和大小, 让任务栏显示出来
					win.SetWindowPos(hWnd, win.HWND_TOP, screenLeft, screenTop, screenRight-screenLeft, screenBottom-screenTop, win.SWP_SHOWWINDOW)
				}
			}
		}

	case win.WM_SIZING:
		fmt.Println("WM_SIZING ...")
	case win.WM_GETMINMAXINFO:
		fmt.Println("WM_GETMINMAXINFO ...")
	}
	return 0
}
