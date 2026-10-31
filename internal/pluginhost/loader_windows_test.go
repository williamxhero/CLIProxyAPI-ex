//go:build windows

package pluginhost

import (
	"context"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

var testReentrantHostCallback uintptr

var testWindowsCloseShutdownCalls atomic.Int64

func testWindowsCloseShutdown() uintptr {
	testWindowsCloseShutdownCalls.Add(1)
	return 0
}

func TestWindowsClosedClientKeepsResidentHostState(t *testing.T) {
	host := New()
	id := windowsHostCallbackID.Add(1)
	instance := &hostCallbackInstance{}
	hostCtx := new(uintptr)
	*hostCtx = id
	hostAPI := &windowsHostAPI{
		abiVersion: pluginHostABIVersion,
		hostCtx:    uintptr(unsafe.Pointer(hostCtx)),
		call:       windowsHostCallCallback,
		freeBuffer: windowsHostFreeCallback,
	}
	windowsResidentHostState.Store(id, windowsResidentHostBlocks{hostCtx: hostCtx, hostAPI: hostAPI})
	windowsHostCallbackEntries.Store(id, dynamicHostCallbackEntry{host: host, pluginID: "alpha", instance: instance})
	shutdownsBefore := testWindowsCloseShutdownCalls.Load()
	client := &dynamicLibraryClient{
		hostCtx:  hostCtx,
		hostAPI:  hostAPI,
		host:     host,
		pluginID: "alpha",
		instance: instance,
		api: windowsPluginAPI{
			call:       1,
			freeBuffer: 1,
			shutdown:   syscall.NewCallback(testWindowsCloseShutdown),
		},
	}

	client.close()

	if _, okEntry := windowsHostCallbackEntries.Load(id); okEntry {
		t.Fatal("callback registration retained after close; repeated reloads would pin the host")
	}
	rawBlocks, okBlocks := windowsResidentHostState.Load(id)
	if !okBlocks {
		t.Fatal("resident host state missing after close; late plugin callbacks could dereference freed memory")
	}
	blocks, okBlocks := rawBlocks.(windowsResidentHostBlocks)
	if !okBlocks || blocks.hostCtx != hostCtx || blocks.hostAPI != hostAPI {
		t.Fatalf("resident host blocks = %+v, want pinned hostCtx/hostAPI", rawBlocks)
	}
	if client.api.call != 0 || client.api.freeBuffer != 0 || client.api.shutdown != 0 {
		t.Fatalf("plugin API table not cleared by close: %+v", client.api)
	}
	if gotShutdowns := testWindowsCloseShutdownCalls.Load(); gotShutdowns != shutdownsBefore+1 {
		t.Fatalf("plugin shutdown calls = %d, want %d", gotShutdowns, shutdownsBefore+1)
	}

	// A late callback from the resident plugin runtime dereferences the
	// pinned context, finds no registration, and fails safely.
	methodBytes, errMethod := syscall.BytePtrFromString("host.log")
	if errMethod != nil {
		t.Fatalf("BytePtrFromString() error = %v", errMethod)
	}
	if rc := windowsHostCall(hostCtx, methodBytes, nil, 0, nil); rc == 0 {
		t.Fatal("windowsHostCall() with detached context succeeded, want failure")
	}
	windowsResidentHostState.Delete(id)
}

func TestWindowsHostCallNativeCallbackReturnsErrorEnvelope(t *testing.T) {
	id := windowsHostCallbackID.Add(1)
	hostCtx := new(uintptr)
	*hostCtx = id
	windowsResidentHostState.Store(id, windowsResidentHostBlocks{hostCtx: hostCtx})
	windowsHostCallbackEntries.Store(id, dynamicHostCallbackEntry{host: New(), pluginID: "native-callback-test", instance: &hostCallbackInstance{}})
	t.Cleanup(func() {
		windowsHostCallbackEntries.Delete(id)
		windowsResidentHostState.Delete(id)
	})
	method, errMethod := syscall.BytePtrFromString("test.unsupported")
	if errMethod != nil {
		t.Fatal(errMethod)
	}
	request := []byte(`{"probe":true}`)
	mem, errAlloc := windows.LocalAlloc(windows.LMEM_FIXED|windows.LMEM_ZEROINIT, uint32(unsafe.Sizeof(windowsBuffer{})))
	if errAlloc != nil || mem == 0 {
		t.Fatalf("allocate native response descriptor: %v", errAlloc)
	}
	defer func() { _, _ = windows.LocalFree(windows.Handle(mem)) }()
	rc, _, _ := syscall.SyscallN(windowsHostCallCallback, uintptr(unsafe.Pointer(hostCtx)),
		uintptr(unsafe.Pointer(method)), uintptr(unsafe.Pointer(&request[0])), uintptr(len(request)), mem)
	runtime.KeepAlive(method)
	runtime.KeepAlive(request)
	if rc != 0 {
		t.Fatalf("native host callback status=%d, want an error envelope", rc)
	}
	var response windowsBuffer
	if errRead := windows.ReadProcessMemory(windows.CurrentProcess(), mem, (*byte)(unsafe.Pointer(&response)), unsafe.Sizeof(response), nil); errRead != nil {
		t.Fatal(errRead)
	}
	if response.ptr == 0 || response.len == 0 {
		t.Fatalf("native callback did not populate response descriptor: %+v", response)
	}
	defer func() { _, _, _ = syscall.SyscallN(windowsHostFreeCallback, response.ptr, response.len) }()
	body := make([]byte, response.len)
	if errRead := windows.ReadProcessMemory(windows.CurrentProcess(), response.ptr, &body[0], response.len, nil); errRead != nil {
		t.Fatal(errRead)
	}
	if !isPluginErrorEnvelope(body) || !strings.Contains(string(body), "unsupported host callback test.unsupported") {
		t.Fatalf("native callback error envelope=%s", body)
	}
}

func TestDynamicLibraryClientCallSurvivesReentrantCallbackStackGrowth(t *testing.T) {
	testReentrantHostCallback = syscall.NewCallback(testGrowHostCallbackStack)
	client := newGuardedPluginClient(&dynamicLibraryClient{api: windowsPluginAPI{
		call:       syscall.NewCallback(testReentrantPluginCall),
		freeBuffer: syscall.NewCallback(testReentrantPluginFree),
	}})
	t.Cleanup(client.Shutdown)

	got, errCall := client.Call(context.Background(), "model.route", []byte(`{}`))
	if errCall != nil {
		t.Fatalf("Call() error = %v", errCall)
	}
	want := `{"ok":true,"result":{"Handled":true}}`
	if string(got) != want {
		t.Fatalf("Call() response = %q, want %q", got, want)
	}
}

func testReentrantPluginCall(_, _ *byte, _ uintptr, response *windowsBuffer) uintptr {
	if testReentrantHostCallback == 0 || response == nil {
		return 1
	}
	_, _, _ = syscall.SyscallN(testReentrantHostCallback)

	raw := []byte(`{"ok":true,"result":{"Handled":true}}`)
	mem, errAlloc := windows.LocalAlloc(windows.LMEM_FIXED, uint32(len(raw)))
	if errAlloc != nil || mem == 0 {
		return 1
	}
	if errWrite := windows.WriteProcessMemory(windows.CurrentProcess(), mem, &raw[0], uintptr(len(raw)), nil); errWrite != nil {
		_, _ = windows.LocalFree(windows.Handle(mem))
		return 1
	}
	response.ptr = mem
	response.len = uintptr(len(raw))
	return 0
}

func testReentrantPluginFree(ptr, _ uintptr) uintptr {
	if ptr != 0 {
		_, _ = windows.LocalFree(windows.Handle(ptr))
	}
	return 0
}

func testGrowHostCallbackStack() uintptr {
	return uintptr(testGrowStack(64))
}

//go:noinline
func testGrowStack(depth int) int {
	var padding [1024]byte
	for index := range padding {
		padding[index] = byte(index + depth)
	}
	if depth == 0 {
		return int(padding[0])
	}
	return testGrowStack(depth-1) + int(padding[depth%len(padding)])
}

var (
	testLivenessFinalizedDuringCallback atomic.Bool
	testLivenessCallbackReadCorrupted   atomic.Bool
	testLivenessCurrentOwnerID          atomic.Uint64
	testLivenessFinalizedOwnerID        atomic.Uint64
)

type testLivenessRequestBuffer [2 * 1024 * 1024]byte

func TestDynamicLibraryClientCallKeepsRequestMemoryAlive(t *testing.T) {
	testLivenessFinalizedDuringCallback.Store(false)
	testLivenessCallbackReadCorrupted.Store(false)

	client := newGuardedPluginClient(&dynamicLibraryClient{api: windowsPluginAPI{
		call:       syscall.NewCallback(testLivenessPluginCall),
		freeBuffer: syscall.NewCallback(testReentrantPluginFree),
	}})
	t.Cleanup(client.Shutdown)

	req := new(testLivenessRequestBuffer)
	req[0] = 0xAA
	req[len(req)-1] = 0x55
	ownerID := testLivenessCurrentOwnerID.Add(1)
	runtime.SetFinalizer(req, func(_ *testLivenessRequestBuffer) {
		testLivenessFinalizedOwnerID.Store(ownerID)
	})

	got, errCall := client.Call(context.Background(), "test.liveness", req[:])
	if errCall != nil {
		t.Fatalf("Call() error = %v", errCall)
	}
	want := `{"ok":true}`
	if string(got) != want {
		t.Fatalf("Call() response = %q, want %q", got, want)
	}
	if testLivenessFinalizedDuringCallback.Load() {
		t.Fatalf("request memory was finalized during native callback execution")
	}
	if testLivenessCallbackReadCorrupted.Load() {
		t.Fatalf("request buffer was corrupted during native callback execution")
	}
}

func testLivenessPluginCall(methodPtr, requestPtr *byte, requestLen uintptr, response *windowsBuffer) uintptr {
	if methodPtr == nil || requestPtr == nil || requestLen == 0 || response == nil {
		return 1
	}

	for i := 0; i < 5; i++ {
		runtime.GC()
		runtime.Gosched()
	}

	currentID := testLivenessCurrentOwnerID.Load()
	if testLivenessFinalizedOwnerID.Load() == currentID {
		testLivenessFinalizedDuringCallback.Store(true)
		return 2
	}

	buf := unsafe.Slice(requestPtr, requestLen)
	if buf[0] != 0xAA || buf[requestLen-1] != 0x55 {
		testLivenessCallbackReadCorrupted.Store(true)
		return 3
	}

	raw := []byte(`{"ok":true}`)
	mem, errAlloc := windows.LocalAlloc(windows.LMEM_FIXED, uint32(len(raw)))
	if errAlloc != nil || mem == 0 {
		return 1
	}
	if errWrite := windows.WriteProcessMemory(windows.CurrentProcess(), mem, &raw[0], uintptr(len(raw)), nil); errWrite != nil {
		_, _ = windows.LocalFree(windows.Handle(mem))
		return 1
	}
	response.ptr = mem
	response.len = uintptr(len(raw))
	return 0
}

var testInvalidResponseFrees atomic.Int64

func testInvalidResponseCall(_, _ *byte, _ uintptr, response *windowsBuffer) uintptr {
	response.ptr = 1
	response.len = 1
	return 0
}

func testInvalidResponseFree(_, _ uintptr) uintptr {
	testInvalidResponseFrees.Add(1)
	return 0
}

func TestDynamicLibraryClientCallRejectsInvalidNativeResponseAndFreesIt(t *testing.T) {
	before := testInvalidResponseFrees.Load()
	client := &dynamicLibraryClient{api: windowsPluginAPI{
		call:       syscall.NewCallback(testInvalidResponseCall),
		freeBuffer: syscall.NewCallback(testInvalidResponseFree),
	}}
	got, err := client.Call(context.Background(), "test.invalid-response", nil)
	if err == nil || !strings.Contains(err.Error(), "read plugin response body") || got != nil {
		t.Fatalf("invalid native response was not rejected: got=%q err=%v", got, err)
	}
	if frees := testInvalidResponseFrees.Load() - before; frees != 1 {
		t.Fatalf("native response free calls=%d, want 1 even after a read failure", frees)
	}
}
