//
// Copyright (c) 2019, Manticore Software LTD (http://manticoresearch.com)
// All rights reserved
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License. You should have
// received a copy of the GPL license along with this program; if you
// did not, you can find it at http://www.gnu.org/
//

package manticore

import (
	"encoding/binary"
	"math"
	"time"
)

type apibuf []byte

func (buf *apibuf) putByte(val uint8) {
	*buf = append(*buf, val)
}

func (buf *apibuf) putWord(val uint16) {
	tmp := make([]byte, 2)
	binary.BigEndian.PutUint16(tmp, val)
	*buf = append(*buf, tmp...)
}

func (buf *apibuf) putUint(val uint32) {
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, val)
	*buf = append(*buf, tmp...)
}

func (buf *apibuf) putInt(val int32) {
	buf.putUint(uint32(val))
}

func (buf *apibuf) putDword(val uint32) {
	buf.putUint(val)
}

func (buf *apibuf) putUint64(val uint64) {
	tmp := make([]byte, 8)
	binary.BigEndian.PutUint64(tmp, val)
	*buf = append(*buf, tmp...)
}

func (buf *apibuf) putInt64(val int64) {
	buf.putUint64(uint64(val))
}

func (buf *apibuf) putDocid(val DocID) {
	buf.putUint64(uint64(val))
}

func (buf *apibuf) putLen(val int) {
	buf.putUint(uint32(val))
}

func (buf *apibuf) putFloat(val float32) {
	buf.putUint(math.Float32bits(val))
}

func (buf *apibuf) putDuration(val time.Duration) {
	buf.putUint(uint32(val / time.Millisecond))
}

func (buf *apibuf) putBoolByte(val bool) {
	if val {
		buf.putByte(1)
	} else {
		buf.putByte(0)
	}
}

func (buf *apibuf) putBoolDword(val bool) {
	if val {
		buf.putDword(1)
	} else {
		buf.putDword(0)
	}
}

func (buf *apibuf) putBytes(val []byte) {
	*buf = append(*buf, val...)
}

func (buf *apibuf) putString(str string) {
	buf.putLen(len(str))
	bytes := []byte(str)
	buf.putBytes(bytes)
}

func (buf *apibuf) getByte() byte {
	val := (*buf)[0]
	*buf = (*buf)[1:]
	return val
}

func (buf *apibuf) getWord() uint16 {
	val := binary.BigEndian.Uint16(*buf)
	*buf = (*buf)[2:]
	return val
}

func (buf *apibuf) getDword() uint32 {
	val := binary.BigEndian.Uint32(*buf)
	*buf = (*buf)[4:]
	return val
}

func (buf *apibuf) getInt() int {
	return int(buf.getDword())
}

func (buf *apibuf) getUint64() uint64 {
	val := binary.BigEndian.Uint64(*buf)
	*buf = (*buf)[8:]
	return val
}

func (buf *apibuf) getInt64() int64 {
	return int64(buf.getUint64())
}

func (buf *apibuf) getDocid() DocID {
	return DocID(buf.getUint64())
}

func (buf *apibuf) getFloat() float32 {
	return math.Float32frombits(buf.getDword())
}

func (buf *apibuf) getByteBool() bool {
	return buf.getByte() != 0
}

func (buf *apibuf) getIntBool() bool {
	return buf.getDword() != 0
}

func (buf *apibuf) getString() string {
	lng := buf.getInt()
	result := string((*buf)[:lng])
	*buf = (*buf)[lng:]
	return result
}

// zerocopy (return slice to original buffer)
func (buf *apibuf) getRefBytes() []byte {
	lng := buf.getInt()
	result := (*buf)[:lng]
	*buf = (*buf)[lng:]
	return result
}

// full-copy
func (buf *apibuf) getBytes() []byte {
	src := buf.getRefBytes()
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

func (buf *apibuf) apiCommand(uCommand eSearchdcommand) int {
	buf.putWord(uint16(uCommand))
	buf.putWord(uint16(searchdcommandv[uCommand]))
	iPlace := len(*buf)
	buf.putUint(0) // space for future len encoding
	return iPlace
}

func (buf *apibuf) finishAPIPacket(iPlace int) {
	uLen := uint32(len(*buf) - iPlace - 4)
	binary.BigEndian.PutUint32((*buf)[iPlace:], uLen)
}

// ensure buf is capable for given size
func (buf *apibuf) resizeBuf(size, maxsize int) {
	if cap(*buf) > maxsize {
		*buf = nil
	}
	if len(*buf) < size {
		*buf = append(*buf, make([]byte, size-len(*buf))...)
	}
	*buf = (*buf)[:size]
}
