package xdnd

import (
	"fmt"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// X11DisplayName returns the display name string (e.g., ":0")
// Returns empty string to use the default display
func X11DisplayName() string {
	return ""
}

// FindWindowByTitle finds an X11 window by searching for a window with the given title
// This is used to locate our GLFW window since we can't directly access its X11 handle
func FindWindowByTitle(conn *xgb.Conn, title string) (xproto.Window, error) {
	setup := xproto.Setup(conn)
	if len(setup.Roots) == 0 {
		return 0, fmt.Errorf("no screens")
	}
	root := setup.DefaultScreen(conn).Root

	// Intern the _NET_WM_NAME and WM_NAME atoms
	netWmNameCookie := xproto.InternAtom(conn, true, uint16(len("_NET_WM_NAME")), "_NET_WM_NAME")
	wmNameCookie := xproto.InternAtom(conn, true, uint16(len("WM_NAME")), "WM_NAME")
	utf8Cookie := xproto.InternAtom(conn, true, uint16(len("UTF8_STRING")), "UTF8_STRING")

	netWmNameReply, _ := netWmNameCookie.Reply()
	wmNameReply, _ := wmNameCookie.Reply()
	utf8Reply, _ := utf8Cookie.Reply()

	var netWmName, wmName, utf8String xproto.Atom
	if netWmNameReply != nil {
		netWmName = netWmNameReply.Atom
	}
	if wmNameReply != nil {
		wmName = wmNameReply.Atom
	}
	if utf8Reply != nil {
		utf8String = utf8Reply.Atom
	}

	// Search for the window
	return findWindowByTitleRecursive(conn, root, title, netWmName, wmName, utf8String)
}

func findWindowByTitleRecursive(conn *xgb.Conn, win xproto.Window, title string, netWmName, wmName, utf8String xproto.Atom) (xproto.Window, error) {
	// Check this window's title
	if checkWindowTitle(conn, win, title, netWmName, wmName, utf8String) {
		return win, nil
	}

	// Query children
	cookie := xproto.QueryTree(conn, win)
	reply, err := cookie.Reply()
	if err != nil {
		return 0, nil // Ignore errors, continue searching
	}

	// Search children
	for _, child := range reply.Children {
		found, err := findWindowByTitleRecursive(conn, child, title, netWmName, wmName, utf8String)
		if err == nil && found != 0 {
			return found, nil
		}
	}

	return 0, fmt.Errorf("window not found")
}

func checkWindowTitle(conn *xgb.Conn, win xproto.Window, title string, netWmName, wmName, utf8String xproto.Atom) bool {
	// Try _NET_WM_NAME first (UTF-8)
	if netWmName != 0 {
		cookie := xproto.GetProperty(conn, false, win, netWmName, utf8String, 0, 256)
		reply, err := cookie.Reply()
		if err == nil && reply.ValueLen > 0 {
			winTitle := string(reply.Value[:reply.ValueLen])
			if winTitle == title {
				return true
			}
		}
	}

	// Fall back to WM_NAME
	if wmName != 0 {
		cookie := xproto.GetProperty(conn, false, win, wmName, xproto.AtomString, 0, 256)
		reply, err := cookie.Reply()
		if err == nil && reply.ValueLen > 0 {
			winTitle := string(reply.Value[:reply.ValueLen])
			if winTitle == title {
				return true
			}
		}
	}

	return false
}
