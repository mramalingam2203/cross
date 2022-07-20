// Package cross defines contracts between Go programs.
package cross

import (
	"time"
	"unsafe"

	"github.com/blitz-frost/io"
)

const (
	InputNone    InputKind = 0 // needed when iterating in Unity, as C# functions return a single value
	InputKeyDown           = 1
	InputKeyUp             = 2
	InputScroll            = 3 // usually mouse wheel
	InputVector            = 4 // usually mouse or touch screen tracking
)

const (
	PacketVideo PacketKind = 0
	PacketAudio            = 1
	PacketInput            = 2
	PacketSync             = 3
)

type Client struct {
	Id    func() (Primary, error) // should probably separate identification from video settings
	Start func() error
}

type Engine struct {
	PrimaryAdd      func(Primary) error
	PrimaryRemove   func(uint64) error
	SecondaryAdd    func(Secondary) error
	SecondaryRemove func(uint64) error
	Start           func(uint64) error
	Stop            func(uint64) error
}

type InputKind byte

type InputPayload []byte

func MakeInputPayload() InputPayload {
	return make(InputPayload, 8)
}

func (x InputPayload) AppendKeyDown(key string) InputPayload {
	return x.appendKey(InputKeyDown, key)
}

func (x InputPayload) AppendKeyUp(key string) InputPayload {
	return x.appendKey(InputKeyUp, key)
}

func (x InputPayload) AppendScroll(delta int8) InputPayload {
	return append(x, byte(InputScroll), byte(delta))
}

func (x InputPayload) AppendVector(xPos, yPos uint16) InputPayload {
	b := *(*[2]byte)(unsafe.Pointer(&xPos))
	x = append(x, byte(InputVector), b[0], b[1])

	b = *(*[2]byte)(unsafe.Pointer(&yPos))
	return append(x, b[0], b[1])
}

func (x InputPayload) Data() []byte {
	return x[8:]
}

func (x InputPayload) IsEmpty() bool {
	if len(x) == 8 {
		return true
	}
	return false
}

// Reset can be used to compose a new InputPayload, without reallocation.
func (x InputPayload) Reset() InputPayload {
	return x[:8]
}

func (x InputPayload) Ts() time.Duration {
	return *(*time.Duration)(unsafe.Pointer(&x[0])) // int64
}

func (x InputPayload) TsSet(ts time.Duration) {
	b := *(*[8]byte)(unsafe.Pointer(&ts))
	copy(x, b[:])
}

func (x InputPayload) appendKey(kind InputKind, key string) InputPayload {
	x = append(x, byte(kind))
	x = append(x, byte(len(key)))
	x = append(x, key...)
	return x
}

type Packet []byte

func MakePacket(payloadSize int) Packet {
	x := make(Packet, 17+payloadSize)
	x.SizeSet(payloadSize)
	return x
}

func (x Packet) Id() uint64 {
	return *(*uint64)(unsafe.Pointer(&x[0]))
}

func (x Packet) IdSet(id uint64) {
	b := *(*[8]byte)(unsafe.Pointer(&id))
	copy(x, b[:])
}

func (x Packet) Kind() PacketKind {
	return PacketKind(x[8])
}

func (x Packet) KindSet(kind PacketKind) {
	x[8] = byte(kind)
}

func (x Packet) Payload() []byte {
	return x[17:]
}

// PayloadSet copies b into the packet, reallocating it if the copy doesn't fit.
func (x Packet) PayloadSet(b []byte) {
	x = append(x[:17], b...)
	x.SizeSet(len(b))
}

// Size returns the payload size.
func (x Packet) Size() int {
	return int(*(*uint64)(unsafe.Pointer(&x[9])))
}

func (x Packet) SizeSet(size int) {
	// store as uint64 for portability
	sz := uint64(size)
	b := *(*[8]byte)(unsafe.Pointer(&sz))
	copy(x[9:], b[:])
}

type PacketKind byte

// Primary defines primary client setup parameters for the rendering engine.
type Primary struct {
	Id           uint64
	RenderWidth  int32
	RenderHeight int32
	WebcamWidth  int32
	WebcamHeight int32
	MaxFps       float32
}

func (x Primary) AsSecondary() Secondary {
	return Secondary{
		Id:           x.Id,
		WebcamWidth:  x.WebcamWidth,
		WebcamHeight: x.WebcamHeight,
	}
}

// Secondary defines secondary client setup parameters for the rendering engine.
type Secondary struct {
	Id           uint64
	WebcamWidth  int32
	WebcamHeight int32
}

// TmpBuffer is used by websockets to receive RPC messages.
// A bit of a bandaid until RPC package gets reworked.
type TmpBuffer []byte

func (x *TmpBuffer) Write(b []byte) error {
	*x = append(*x, b...)
	return nil
}

func (x *TmpBuffer) WriteTo(w io.Writer) error {
	b := *x
	err := w.Write(b)
	*x = make([]byte, 0, len(b))
	return err
}

type VideoPayload []byte

func (x VideoPayload) Data() []byte {
	return x[16:]
}

func (x VideoPayload) Duration() time.Duration {
	return *(*time.Duration)(unsafe.Pointer(&x[8]))
}

func (x VideoPayload) DurationSet(t time.Duration) {
	b := *(*[8]byte)(unsafe.Pointer(&t))
	copy(x[8:], b[:])
}

func (x VideoPayload) Pts() time.Duration {
	return *(*time.Duration)(unsafe.Pointer(&x[0])) // int64
}

func (x VideoPayload) PtsSet(t time.Duration) {
	b := *(*[8]byte)(unsafe.Pointer(&t))
	copy(x, b[:])
}

func VideoPayloadSize(width, height int) int {
	return 16 + 4*width*height
}
