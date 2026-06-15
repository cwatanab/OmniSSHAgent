package pageant

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"github.com/cwchiu/go-winapi"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/sys/windows"
)

const (
	cryptProtectMemoryCrossProcess = 0x1
	cryptProtectMemoryBlockSize    = 16
)

var (
	procCryptProtectMemory = syscall.NewLazyDLL("crypt32.dll").NewProc("CryptProtectMemory")
	procGetUserNameExA     = syscall.NewLazyDLL("secur32.dll").NewProc("GetUserNameExA")
)

const (
	className = "Pageant"
	id        = 0x804e50ba
	checkID   = 0x02e08fc7
)

var (
	procPostThreadMessage = syscall.NewLazyDLL("user32.dll").NewProc("PostThreadMessageW")
	procUnregisterClass   = syscall.NewLazyDLL("user32.dll").NewProc("UnregisterClassW")
	postQuitMessage       = func(threadID uint32) bool {
		ret, _, _ := procPostThreadMessage.Call(uintptr(threadID), uintptr(winapi.WM_QUIT), 0, 0)
		return ret != 0
	}
	unregisterClass = func(hInstance winapi.HINSTANCE) bool {
		classNameUTF16, err := syscall.UTF16PtrFromString(className)
		if err != nil {
			return false
		}
		ret, _, _ := procUnregisterClass.Call(uintptr(unsafe.Pointer(classNameUTF16)), uintptr(hInstance))
		return ret != 0
	}
)

type Pageant struct {
	agent.ExtendedAgent
	Debug     bool
	AppName   string
	CheckFunc func()
}

type copyDataStruct struct {
	dwData uintptr
	cbData uint32
	lpData uintptr
}

func (a *Pageant) myRegisterClass(hInstance winapi.HINSTANCE) winapi.ATOM {
	var wc winapi.WNDCLASSEX

	wc.CbSize = uint32(unsafe.Sizeof(winapi.WNDCLASSEX{}))
	wc.Style = 0
	wc.LpfnWndProc = syscall.NewCallback(a.wndProc)
	wc.CbClsExtra = 0
	wc.CbWndExtra = 0
	wc.HInstance = hInstance
	wc.HIcon = winapi.LoadIcon(hInstance, winapi.MAKEINTRESOURCE(132))
	wc.HCursor = winapi.LoadCursor(0, winapi.MAKEINTRESOURCE(winapi.IDC_CROSS))
	wc.HbrBackground = 0
	wc.LpszMenuName = nil
	wc.LpszClassName, _ = syscall.UTF16PtrFromString(className)

	return winapi.RegisterClassEx(&wc)
}

func (a *Pageant) wndProc(hWnd winapi.HWND, message uint32, wParam uintptr, lParam uintptr) uintptr {
	if message == winapi.WM_COPYDATA {
		err := a.handleCopyMessage((*copyDataStruct)(unsafe.Pointer(lParam)))
		if err != nil {
			log.Print(err)
			return 0
		}
		return 1
	}
	return winapi.DefWindowProc(hWnd, uint32(message), wParam, lParam)
}

type mMap []byte

func (m *mMap) header() *reflect.SliceHeader { return (*reflect.SliceHeader)(unsafe.Pointer(m)) }

func ptr2Array(addr uintptr, sz int) (m mMap) {
	dh := m.header()
	dh.Data = addr
	dh.Len = sz
	dh.Cap = dh.Len
	return
}

var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var openFileMappingW = kernel32.NewProc("OpenFileMappingW")
var fileMapAllAccess = uint32(0xF001F)
var zeroUint32 = uint32(0)

func (a *Pageant) handleCopyMessage(cdata *copyDataStruct) error {
	checkMode := false
	switch cdata.dwData {
	case id:
		break
	case checkID:
		checkMode = true
	default:
		return errors.New("ID is different")
	}
	if a.Debug {
		log.Println("Pageant: received message")
	}

	m := ptr2Array(cdata.lpData, int(cdata.cbData-1))
	mapname := string(m[:cdata.cbData-1])
	ret, _, _ := openFileMappingW.Call(
		uintptr(fileMapAllAccess),
		uintptr(zeroUint32),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(mapname))))
	h := syscall.Handle(ret)
	if h == 0 {
		return errors.New("err:OpenFileMappingW")
	}
	defer syscall.CloseHandle(h)
	addr, errno := syscall.MapViewOfFile(h, uint32(syscall.FILE_MAP_WRITE), 0, 0, 0)
	if addr == 0 {
		return fmt.Errorf("Failed: %s", os.NewSyscallError("MapViewOfFile", errno))
	}
	m = ptr2Array(addr, 4) // message size
	buf := bytes.NewBuffer(m)
	ln := int32(0)
	binary.Read(buf, binary.BigEndian, &ln)
	m = ptr2Array(addr, int(ln)+4) // read ssh-agent message

	out := bytes.Buffer{}
	if checkMode {
		if a.CheckFunc != nil {
			a.CheckFunc()
		}
		b := []byte(a.AppName)
		//out.WriteString(a.AppName)
		var length [4]byte
		binary.BigEndian.PutUint32(length[:], uint32(len(b)))
		out.Write(length[:])
		out.Write(b)
		m = ptr2Array(addr, out.Len())
		copy(m, out.Bytes()[:out.Len()])
		return nil
	}
	err := agent.ServeAgent(a,
		struct {
			io.Reader
			io.Writer
		}{bytes.NewBuffer(m), &out},
	)
	if err != nil && err != io.EOF {
		return fmt.Errorf("ServeAgent err:%w", err)
	}
	m = ptr2Array(addr, out.Len())
	copy(m, out.Bytes()[:out.Len()])
	return nil
}

func initInstance(hInstance winapi.HINSTANCE, nCmdShow int) (winapi.HWND, error) {
	classNameUTF16, err := syscall.UTF16PtrFromString(className)
	if err != nil {
		return 0, err
	}

	hWnd := winapi.CreateWindowEx(
		winapi.WS_EX_TRANSPARENT|winapi.WS_EX_TOOLWINDOW|winapi.WS_EX_TOPMOST|winapi.WS_EX_NOACTIVATE,
		classNameUTF16,
		classNameUTF16,
		winapi.WS_POPUP,
		0, 0, 0, 0,
		0, 0, hInstance, nil)
	if hWnd == 0 {
		return 0, errors.New("cannot create window")
	}

	winapi.ShowWindow(hWnd, winapi.SW_SHOW)
	return hWnd, nil
}

func (a *Pageant) RunAgent(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hInstance := winapi.GetModuleHandle(nil)
	registeredClass := a.myRegisterClass(hInstance) != 0

	hWnd, err := initInstance(hInstance, winapi.SW_SHOW)
	if err != nil {
		log.Printf("pageant: %v\n", err)
		return
	}
	if registeredClass {
		defer unregisterClass(hInstance)
	}
	defer winapi.DestroyWindow(hWnd)

	threadID := windows.GetCurrentThreadId()
	stopWatcher := startCancelWatcher(ctx, threadID)
	defer stopWatcher()

	msg := (*winapi.MSG)(unsafe.Pointer(winapi.GlobalAlloc(0, unsafe.Sizeof(winapi.MSG{}))))
	defer winapi.GlobalFree(winapi.HGLOBAL(unsafe.Pointer(msg)))
	for winapi.GetMessage(msg, 0, 0, 0) != 0 {
		winapi.TranslateMessage(msg)
		winapi.DispatchMessage(msg)
		if msg.Message == winapi.WM_QUIT {
			break
		}
	}

	runtime.KeepAlive(&msg)
}

func startCancelWatcher(ctx context.Context, threadID uint32) func() {
	if ctx == nil {
		ctx = context.Background()
	}
	exitCh := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			postQuitMessage(threadID)
		case <-exitCh:
		}
	}()
	return func() {
		close(exitCh)
	}
}

const NameUserPrincipal = 8

func getUserName() string {
	var nameLength uint32 = 0
	// GetUserNameExA returns FALSE with ERROR_MORE_DATA when buffer is NULL, but we only need the required size in nameLength.
	// We ignore the return value here.
	_, _, _ = procGetUserNameExA.Call(
		uintptr(NameUserPrincipal),
		0,
		uintptr(unsafe.Pointer(&nameLength)),
	)
	if nameLength > 0 {
		nameBuf := make([]byte, nameLength)
		ret, _, _ := procGetUserNameExA.Call(
			uintptr(NameUserPrincipal),
			uintptr(unsafe.Pointer(&nameBuf[0])),
			uintptr(unsafe.Pointer(&nameLength)),
		)
		if ret != 0 {
			// Find terminating null byte and convert to string
			name := string(nameBuf[:nameLength-1])
			if idx := strings.Index(name, "@"); idx != -1 {
				name = name[:idx]
			}
			if name != "" {
				return name
			}
		}
	}
	return os.Getenv("USERNAME")
}

func ObfuscatedPipeName() (string, error) {
	username := getUserName()

	input := "Pageant"
	cryptlen := ((len(input) + 1) + cryptProtectMemoryBlockSize - 1) / cryptProtectMemoryBlockSize * cryptProtectMemoryBlockSize
	cryptdata := make([]byte, cryptlen)
	copy(cryptdata, []byte(input))

	ret, _, _ := procCryptProtectMemory.Call(
		uintptr(unsafe.Pointer(&cryptdata[0])),
		uintptr(cryptlen),
		uintptr(cryptProtectMemoryCrossProcess),
	)
	if ret == 0 {
		return "", fmt.Errorf("CryptProtectMemory failed")
	}

	h := sha256.New()
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(cryptdata)))
	h.Write(lenBuf[:])
	h.Write(cryptdata)
	hashStr := hex.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("pageant.%s.%s", username, hashStr), nil
}
