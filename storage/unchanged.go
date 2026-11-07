package storage

import (
	"bytes"
	"encoding/json"
)

// PeerUnchanged reports whether stored is the encoding of value, ignoring
// CreatedAt, which is set to the current time on every peer conversion.
//
// Storages use it to skip writing peers that did not change, e.g. users
// repeated in every update. So the stored CreatedAt is when this content was
// first stored, not when the peer was last seen.
func PeerUnchanged(stored []byte, value Peer) bool {
	var old Peer
	if err := json.Unmarshal(stored, &old); err != nil {
		return false
	}
	value.CreatedAt = old.CreatedAt
	data, err := json.Marshal(value)
	if err != nil {
		return false
	}
	return bytes.Equal(stored, data)
}
