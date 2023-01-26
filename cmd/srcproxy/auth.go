package main

import (
	"crypto/sha256"
	"errors"
	"net"
)

type AccessControl struct {
	acl map[sha256Digest]AccessControlEntry
}

type sha256Digest = [sha256.Size]byte

func NewAccessControl(acl []AccessControlEntry) (a *AccessControl) {
	a = &AccessControl{}
	if acl == nil {
		return
	}
	a.acl = make(map[sha256Digest]AccessControlEntry)
	for _, entry := range acl {
		a.acl[sha256.Sum256([]byte(entry.Auth))] = entry
	}
	return
}

func (a *AccessControl) CheckSrc(origSrc *net.TCPAddr, auth []byte) (modifiedSrc *net.TCPAddr, err error) {
	if a.acl == nil {
		// acl is disabled, allow all
		modifiedSrc = origSrc
		return
	}
	if len(auth) != sha256.Size {
		err = errors.New("invalid auth: length mismatch")
		return
	}
	var digest sha256Digest
	copy(digest[:], auth)
	entry, ok := a.acl[digest]
	if !ok {
		err = errors.New("invalid auth: not found in acl")
		return
	}
	if entry.AllowedSrcIPs == nil {
		// allow all src
		modifiedSrc = origSrc
		return
	}
	for _, prefix := range entry.AllowedSrcIPs {
		if prefix.Contains(origSrc.IP) {
			modifiedSrc = origSrc
			return
		}
	}
	err = errors.New("src not in the server allowed src list")
	return
}
