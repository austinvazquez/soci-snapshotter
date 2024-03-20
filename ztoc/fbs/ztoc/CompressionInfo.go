// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package ztoc

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type CompressionInfo struct {
	_tab flatbuffers.Table
}

func GetRootAsCompressionInfo(buf []byte, offset flatbuffers.UOffsetT) *CompressionInfo {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &CompressionInfo{}
	x.Init(buf, n+offset)
	return x
}

func FinishCompressionInfoBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.Finish(offset)
}

func GetSizePrefixedRootAsCompressionInfo(buf []byte, offset flatbuffers.UOffsetT) *CompressionInfo {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &CompressionInfo{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func FinishSizePrefixedCompressionInfoBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.FinishSizePrefixed(offset)
}

func (rcv *CompressionInfo) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *CompressionInfo) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *CompressionInfo) CompressionAlgorithm() CompressionAlgorithm {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return CompressionAlgorithm(rcv._tab.GetInt8(o + rcv._tab.Pos))
	}
	return 1
}

func (rcv *CompressionInfo) MutateCompressionAlgorithm(n CompressionAlgorithm) bool {
	return rcv._tab.MutateInt8Slot(4, int8(n))
}

func (rcv *CompressionInfo) MaxSpanId() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *CompressionInfo) MutateMaxSpanId(n int32) bool {
	return rcv._tab.MutateInt32Slot(6, n)
}

func (rcv *CompressionInfo) SpanDigests(j int) []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.ByteVector(a + flatbuffers.UOffsetT(j*4))
	}
	return nil
}

func (rcv *CompressionInfo) SpanDigestsLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *CompressionInfo) Checkpoints(j int) byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.GetByte(a + flatbuffers.UOffsetT(j*1))
	}
	return 0
}

func (rcv *CompressionInfo) CheckpointsLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *CompressionInfo) CheckpointsBytes() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *CompressionInfo) MutateCheckpoints(j int, n byte) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.MutateByte(a+flatbuffers.UOffsetT(j*1), n)
	}
	return false
}

func CompressionInfoStart(builder *flatbuffers.Builder) {
	builder.StartObject(4)
}
func CompressionInfoAddCompressionAlgorithm(builder *flatbuffers.Builder, compressionAlgorithm CompressionAlgorithm) {
	builder.PrependInt8Slot(0, int8(compressionAlgorithm), 1)
}
func CompressionInfoAddMaxSpanId(builder *flatbuffers.Builder, maxSpanId int32) {
	builder.PrependInt32Slot(1, maxSpanId, 0)
}
func CompressionInfoAddSpanDigests(builder *flatbuffers.Builder, spanDigests flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(spanDigests), 0)
}
func CompressionInfoStartSpanDigestsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func CompressionInfoAddCheckpoints(builder *flatbuffers.Builder, checkpoints flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(3, flatbuffers.UOffsetT(checkpoints), 0)
}
func CompressionInfoStartCheckpointsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(1, numElems, 1)
}
func CompressionInfoEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
