package dfx

// the app id is the string the desktop pairs a window with its launcher entry by: the wayland app id, and the x11
// class and instance names. glfw takes them only through string window hints, which cimgui-go's backend does not
// expose, so this reaches glfw directly. glfw is linked statically by cimgui-go's glfw backend, so the symbol
// resolves without a header; the constants are glfw 3.4's.

/*
#include <stdlib.h>

#define DFX_GLFW_X11_CLASS_NAME    0x00024001
#define DFX_GLFW_X11_INSTANCE_NAME 0x00024002
#define DFX_GLFW_WAYLAND_APP_ID    0x00026001

extern void glfwWindowHintString(int hint, const char* value);

static void dfxWindowHintAppId(const char* id) {
	glfwWindowHintString(DFX_GLFW_WAYLAND_APP_ID, id);
	glfwWindowHintString(DFX_GLFW_X11_CLASS_NAME, id);
	glfwWindowHintString(DFX_GLFW_X11_INSTANCE_NAME, id);
}
*/
import "C"

import "unsafe"

// setAppID applies the app id hints to the next window glfw creates. it must run after glfw is initialized, which
// the backend's construction does, and before CreateWindow. tests replace it, as they replace createBackend.
var setAppID = func(id string) {
	cid := C.CString(id)
	defer C.free(unsafe.Pointer(cid))
	C.dfxWindowHintAppId(cid)
}
