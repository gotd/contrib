package storage

import "bytes"

// PeerUnchanged reports whether stored is the encoding of value in enc,
// ignoring CreatedAt, which is set to the current time on every peer conversion.
// A value stored in another format is reported as changed, so it is rewritten.
//
// Storages use it to skip writing peers that did not change, e.g. users
// repeated in every update. So the stored CreatedAt is when this content was
// first stored, not when the peer was last seen.
func PeerUnchanged(stored []byte, value Peer, enc PeerEncoding) bool {
	var old Peer
	if err := DecodePeer(stored, &old); err != nil {
		return false
	}
	value.CreatedAt = old.CreatedAt
	data, err := enc.Encode(value)
	if err != nil {
		return false
	}
	return bytes.Equal(stored, data)
}
