package reader

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

type rawHeader struct {
	Header    [4]byte // FORM
	FileSize  uint32
	SubHeader [4]byte // AUDO
	Size      uint32
	Tracks    uint32 // Track count
}

type TrackInfo struct {
	Offset int64
	Size   uint32
}

type Reader struct {
	r            io.ReadSeekCloser
	trackOffsets []int64
	ti           []TrackInfo
}

func New(r io.ReadSeekCloser) (rdr Reader, err error) {
	rdr = Reader{
		r: r,
	}

	err = rdr.readHeader()
	if err != nil {
		return rdr, err
	}

	rdr.calculateTracks()

	return rdr, nil
}

func (r *Reader) readHeader() (err error) {
	var header rawHeader
	err = binary.Read(r.r, binary.LittleEndian, &header)
	if err != nil {
		return err
	}

	if string(header.Header[:]) != `FORM` {
		return fmt.Errorf(`invalid header: %[1]v (%[1]s)`, header.Header)
	}

	if string(header.SubHeader[:]) != `AUDO` {
		return fmt.Errorf(`invalid sub-header: %[1]v (%[1]s)`, header.SubHeader)
	}

	trackOffsets := make([]uint32, header.Tracks)
	err = binary.Read(r.r, binary.LittleEndian, &trackOffsets)
	if err != nil {
		return err
	}

	for _, t := range trackOffsets {
		tOffset := t + 4 // 4 = header length?
		r.trackOffsets = append(r.trackOffsets, int64(tOffset))
	}

	return nil
}

// calculateTracks calculates the sizes of tracks from next offset.
// size = nextOffset - currentOffset
// the last track size will be wrong, so just use a large number
func (r *Reader) calculateTracks() {
	for idx, t := range r.trackOffsets {

		ti := TrackInfo{
			Offset: t,
			Size:   0,
		}

		if idx != (len(r.trackOffsets) - 1) {
			size := r.trackOffsets[idx+1] - t

			if size >= math.MaxUint32 {
				size = math.MaxUint32
			}

			ti.Size = uint32(size)
		} else {
			// Last track
			ti.Size = math.MaxUint32 // read until EOF
		}

		r.ti = append(r.ti, ti)
	}
}

func (r Reader) Tracks() (l []TrackInfo) {
	return r.ti
}
