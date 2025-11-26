package xdnd

import (
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// Atoms holds all the X11 atoms needed for XDND protocol
type Atoms struct {
	// XDND protocol atoms
	XdndAware     xproto.Atom
	XdndSelection xproto.Atom
	XdndEnter     xproto.Atom
	XdndPosition  xproto.Atom
	XdndStatus    xproto.Atom
	XdndLeave     xproto.Atom
	XdndDrop      xproto.Atom
	XdndFinished  xproto.Atom
	XdndTypeList  xproto.Atom

	// XDND actions
	XdndActionCopy xproto.Atom

	// Standard selection atoms
	TARGETS xproto.Atom

	// MIME type atoms
	AudioMidi   xproto.Atom
	TextUriList xproto.Atom

	// Connection for reverse lookups
	conn *xgb.Conn
}

// InternAtoms creates and interns all required atoms
func InternAtoms(conn *xgb.Conn) (*Atoms, error) {
	atoms := &Atoms{conn: conn}

	// List of atom names to intern
	atomNames := []struct {
		name string
		dest *xproto.Atom
	}{
		{"XdndAware", &atoms.XdndAware},
		{"XdndSelection", &atoms.XdndSelection},
		{"XdndEnter", &atoms.XdndEnter},
		{"XdndPosition", &atoms.XdndPosition},
		{"XdndStatus", &atoms.XdndStatus},
		{"XdndLeave", &atoms.XdndLeave},
		{"XdndDrop", &atoms.XdndDrop},
		{"XdndFinished", &atoms.XdndFinished},
		{"XdndTypeList", &atoms.XdndTypeList},
		{"XdndActionCopy", &atoms.XdndActionCopy},
		{"TARGETS", &atoms.TARGETS},
		{"audio/midi", &atoms.AudioMidi},
		{"text/uri-list", &atoms.TextUriList},
	}

	// Intern each atom
	for _, an := range atomNames {
		cookie := xproto.InternAtom(conn, false, uint16(len(an.name)), an.name)
		reply, err := cookie.Reply()
		if err != nil {
			return nil, err
		}
		*an.dest = reply.Atom
	}

	return atoms, nil
}

// GetMimeAtom returns the atom for a MIME type string
func (a *Atoms) GetMimeAtom(mime string) xproto.Atom {
	switch mime {
	case "audio/midi":
		return a.AudioMidi
	case "text/uri-list":
		return a.TextUriList
	default:
		return 0
	}
}

// GetAtomName returns the name of an atom (for debugging)
func (a *Atoms) GetAtomName(atom xproto.Atom) string {
	if a.conn == nil {
		return "?"
	}
	cookie := xproto.GetAtomName(a.conn, atom)
	reply, err := cookie.Reply()
	if err != nil {
		return "?"
	}
	return reply.Name
}
