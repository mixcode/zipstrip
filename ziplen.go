package main

import (
	"fmt"
	"io"
)

var (
	ErrNoZIP = fmt.Errorf("Cannot find ZIP entry")
)

//
// Note: for ZIP file spec, see https://www.pkware.com/documents/APPNOTE/APPNOTE_6.2.0.txt
//

// Find ZIP End-of-central-directory record.
func findZipSignature(b []byte) (offset int, ok bool) {
	for i := 0; i < len(b)-4; i++ {
		if b[i] == 'P' && b[i+1] == 'K' && b[i+2] == 0x05 && b[i+3] == 0x06 {
			return i, true
		}
	}
	return -1, false
}

// Get exact end-of-ZIP position, within last maxTruncate bytes.
func zipLength(r io.ReadSeeker, maxTruncate int64) (sz int64, err error) {
	// get file length
	fileSz, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return
	}

	const lenEOCD = 0x16   // length of a ZIP end-of-central-directory entry
	maxTruncate += lenEOCD // maxTruncate == ZIP eocd search range

	// Find ZIP End-of-central-directory record at the end of file
	offset := fileSz
	targetSz := int64(0x40) // size of data to search
	if targetSz > maxTruncate {
		targetSz = maxTruncate
	}
	ok := false
	eocd := 0

	for offset > 0 {
		newOffset := fileSz - targetSz
		if newOffset < 0 {
			newOffset = 0
		}
		readSz := offset - newOffset + 4
		offset = newOffset
		if offset+readSz > fileSz {
			readSz = fileSz - offset
		}

		buf := make([]byte, readSz)
		_, err = r.Seek(offset, io.SeekStart)
		if err != nil {
			return
		}
		_, err = io.ReadFull(r, buf)
		if err != nil {
			return
		}
		eocd, ok = findZipSignature(buf)
		if ok {
			break
		}

		if targetSz == maxTruncate {
			break
		}
		targetSz *= 2 // increase target read size
		if targetSz > maxTruncate {
			targetSz = maxTruncate
		}
	}
	if !ok {
		return 0, ErrNoZIP
	}

	// Read the End-of-central-directory entry
	buf := make([]byte, lenEOCD)
	_, err = r.Seek(offset+int64(eocd), io.SeekStart)
	if err != nil {
		return
	}
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}
	// the last 16bit contains the comment length in little-endian
	lenComment := int(buf[lenEOCD-1])<<8 | int(buf[lenEOCD-2])

	// actual size
	return offset + int64(eocd+lenEOCD+lenComment), nil
}
