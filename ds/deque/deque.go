package deque

import (
	"errors"
	"fmt"
)

// Constants definition
const (
	SegmentCapacity = 128
)

// Define internal errors
var (
	ErrOutOffRange = errors.New("out off range")
)

// Deque is double-ended queue supports efficient data insertion from the head and tail, random access and iterator access.
type Deque struct {
	Pool    *Pool
	Segs    []*Segment
	BeginFi int
	EndFi   int
	SizeFi  int
}

// New creates a new deque
func New() *Deque {
	dq := &Deque{
		Pool: newPool(),
		Segs: make([]*Segment, 0),
	}
	return dq
}

// Size returns the amount of values in the deque
func (d *Deque) Size() int {
	return d.SizeFi
}

// Empty returns true if the deque is empty,otherwise returns false.
func (d *Deque) Empty() bool {
	return d.SizeFi == 0
}

func (d *Deque) segUsed() int {
	if d.SizeFi == 0 {
		return 0
	}
	if d.EndFi > d.BeginFi {
		return d.EndFi - d.BeginFi
	}
	return len(d.Segs) - d.BeginFi + d.EndFi
}

// PushFront pushed a value to the front of the deque
func (d *Deque) PushFront(value interface{}) {
	d.firstAvailableSegment().pushFront(value)
	d.SizeFi++
	if d.segUsed() >= len(d.Segs) {
		d.expand()
	}
}

// PushBack pushed a value to the back of deque
func (d *Deque) PushBack(value interface{}) {
	d.lastAvailableSegment().pushBack(value)
	d.SizeFi++
	if d.segUsed() >= len(d.Segs) {
		d.expand()
	}
}

// Insert inserts a value to the position pos of the deque
func (d *Deque) Insert(pos int, value interface{}) {
	if pos < 0 || pos > d.SizeFi {
		return
	}
	if pos == 0 {
		d.PushFront(value)
		return
	}
	if pos == d.SizeFi {
		d.PushBack(value)
		return
	}
	seg, pos := d.pos(pos)
	if seg < d.segUsed()-seg {
		// seg is closer to the front
		d.moveFrontInsert(seg, pos, value)
	} else {
		// seg is closer to the back
		d.moveBackInsert(seg, pos, value)
	}
	d.SizeFi++
	if d.segUsed() >= len(d.Segs) {
		d.expand()
	}
}

func (d *Deque) moveFrontInsert(seg, pos int, value interface{}) {
	if d.firstSegment().full() {
		if d.segUsed() >= len(d.Segs) {
			d.expand()
		}
		d.BeginFi = d.preIndex(d.BeginFi)
		d.Segs[d.BeginFi] = d.Pool.get()
		if pos == 0 {
			pos = SegmentCapacity - 1
		} else {
			seg++
			pos--
		}
	} else {
		if pos == 0 {
			seg--
			pos = SegmentCapacity - 1
		} else {
			if seg != 0 {
				pos--
			}
		}
	}
	for i := 0; i < seg; i++ {
		cur := d.segmentAt(i)
		next := d.segmentAt(i + 1)
		cur.pushBack(next.popFront())
	}
	d.segmentAt(seg).insert(pos, value)
}

func (d *Deque) moveBackInsert(seg, pos int, value interface{}) {
	// move back
	if d.lastSegment().full() {
		if d.segUsed() >= len(d.Segs) {
			d.expand()
		}
		d.Segs[d.EndFi] = d.Pool.get()
		d.EndFi = d.nextIndex(d.EndFi)
	}
	for i := d.segUsed() - 1; i > seg; i-- {
		cur := d.segmentAt(i)
		pre := d.segmentAt(i - 1)
		cur.pushFront(pre.popBack())
	}
	d.segmentAt(seg).insert(pos, value)
}

// Front returns the value at the first position of the deque
func (d *Deque) Front() interface{} {
	return d.firstSegment().front()
}

// Back returns the value at the last position of the deque
func (d *Deque) Back() interface{} {
	return d.lastSegment().back()
}

// At returns the value at position pos of the deque
func (d *Deque) At(pos int) interface{} {
	if pos < 0 || pos >= d.Size() {
		return nil
	}
	seg, pos := d.pos(pos)
	return d.segmentAt(seg).at(pos)
}

// Set sets the value of the deque's position pos with value val
func (d *Deque) Set(pos int, val interface{}) error {
	if pos < 0 || pos >= d.SizeFi {
		return ErrOutOffRange
	}
	seg, pos := d.pos(pos)
	d.segmentAt(seg).set(pos, val)
	return nil
}

// PopFront returns the value at the first position of the deque and removes it
func (d *Deque) PopFront() interface{} {
	if d.SizeFi == 0 {
		return nil
	}
	s := d.Segs[d.BeginFi]
	v := s.popFront()
	if s.size() == 0 {
		d.putToPool(s)
		d.Segs[d.BeginFi] = nil
		d.BeginFi = d.nextIndex(d.BeginFi)
	}
	d.SizeFi--
	d.shrinkIfNeeded()
	return v
}

// PopBack returns the value at the lase position of the deque and removes it
func (d *Deque) PopBack() interface{} {
	if d.SizeFi == 0 {
		return nil
	}
	seg := d.preIndex(d.EndFi)
	s := d.Segs[seg]
	v := s.popBack()

	if s.size() == 0 {
		d.putToPool(s)
		d.Segs[seg] = nil
		d.EndFi = seg
	}

	d.SizeFi--
	d.shrinkIfNeeded()
	return v
}

// EraseAt erases the element at the position pos
func (d *Deque) EraseAt(pos int) {
	if pos < 0 || pos >= d.SizeFi {
		return
	}
	seg, pos := d.pos(pos)
	d.segmentAt(seg).eraseAt(pos)
	if seg < d.SizeFi-seg-1 {
		for i := seg; i > 0; i-- {
			cur := d.segmentAt(i)
			pre := d.segmentAt(i - 1)
			cur.pushFront(pre.popBack())
		}
		if d.firstSegment().empty() {
			d.putToPool(d.firstSegment())
			d.Segs[d.BeginFi] = nil
			d.BeginFi = d.nextIndex(d.BeginFi)
			d.shrinkIfNeeded()
		}
	} else {
		for i := seg; i < d.segUsed()-1; i++ {
			cur := d.segmentAt(i)
			next := d.segmentAt(i + 1)
			cur.pushBack(next.popFront())
		}
		if d.lastSegment().empty() {
			d.putToPool(d.lastSegment())
			d.Segs[d.preIndex(d.EndFi)] = nil
			d.EndFi = d.preIndex(d.EndFi)
			d.shrinkIfNeeded()
		}
	}
	d.SizeFi--
}

// EraseRange erases elements in range [firstPos, lastPos)
func (d *Deque) EraseRange(firstPos, lastPos int) {
	if firstPos < 0 || firstPos >= lastPos || lastPos > d.SizeFi {
		return
	}
	num := lastPos - firstPos
	if d.SizeFi-firstPos < lastPos {
		// move back
		for pos := firstPos; pos+num < d.SizeFi; pos++ {
			d.Set(pos, d.At(pos+num))
		}
		for ; num > 0; num-- {
			d.PopBack()
		}
	} else {
		// move front
		for pos := lastPos - 1; pos-num >= 0; pos-- {
			d.Set(pos, d.At(pos-num))
		}
		for ; num > 0; num-- {
			d.PopFront()
		}
	}
}

// Clear erases all elements in the deque
func (d *Deque) Clear() {
	d.EraseRange(0, d.SizeFi)
}

func (d *Deque) putToPool(s *Segment) {
	s.clear()
	d.Pool.put(s)

	if d.Pool.size()*6/5 > d.segUsed() {
		d.Pool.shrinkToSize(d.segUsed() / 5)
	}
}

func (d *Deque) firstAvailableSegment() *Segment {
	if d.firstSegment() != nil && !d.firstSegment().full() {
		return d.firstSegment()
	}
	if d.segUsed() >= len(d.Segs) {
		d.expand()
	}
	if d.firstSegment() == nil || d.firstSegment().full() {
		d.BeginFi = d.preIndex(d.BeginFi)
		s := d.Pool.get()
		d.Segs[d.BeginFi] = s
		return s
	}
	return d.firstSegment()
}

func (d *Deque) lastAvailableSegment() *Segment {
	if d.lastSegment() != nil && !d.lastSegment().full() {
		return d.lastSegment()
	}
	if d.segUsed() >= len(d.Segs) {
		d.expand()
	}
	if d.lastSegment() == nil || d.lastSegment().full() {
		s := d.Pool.get()
		d.Segs[d.EndFi] = s
		d.EndFi = d.nextIndex(d.EndFi)
		return s
	}
	return d.lastSegment()
}

func (d *Deque) firstSegment() *Segment {
	if len(d.Segs) == 0 {
		return nil
	}
	return d.Segs[d.BeginFi]
}

func (d *Deque) lastSegment() *Segment {
	if len(d.Segs) == 0 {
		return nil
	}
	return d.Segs[d.preIndex(d.EndFi)]
}

func (d *Deque) segmentAt(seg int) *Segment {
	return d.Segs[(seg+d.BeginFi)%cap(d.Segs)]
}

func (d *Deque) pos(position int) (int, int) {
	if position <= d.firstSegment().size()-1 {
		return 0, position
	}
	position -= d.firstSegment().size()
	return position/SegmentCapacity + 1, position % SegmentCapacity
}

func (d *Deque) expand() {
	newCapacity := d.segUsed() * 2
	if newCapacity == 0 {
		newCapacity = 1
	}
	seg := make([]*Segment, newCapacity)
	for i := 0; i < d.segUsed(); i++ {
		seg[i] = d.Segs[(d.BeginFi+i)%d.segUsed()]
	}
	u := d.segUsed()

	d.BeginFi = 0
	d.EndFi = u
	d.Segs = seg
}

//shrinkIfNeeded shrinks the deque if it has too many unused space.
func (d *Deque) shrinkIfNeeded() {
	if int(float64(d.segUsed()*2)*1.2) < cap(d.Segs) {
		newCapacity := cap(d.Segs) / 2
		seg := make([]*Segment, newCapacity)
		for i := 0; i < d.segUsed(); i++ {
			seg[i] = d.Segs[(d.BeginFi+i)%cap(d.Segs)]
		}
		u := d.segUsed()
		d.BeginFi = 0
		d.EndFi = u
		d.Segs = seg
	}
}

func (d *Deque) nextIndex(index int) int {
	return (index + 1) % cap(d.Segs)
}

func (d *Deque) preIndex(index int) int {
	return (index - 1 + cap(d.Segs)) % cap(d.Segs)
}

// String returns a string representation of the deque
func (d *Deque) String() string {
	str := "["
	for i := 0; i < d.Size(); i++ {
		if str != "[" {
			str += " "
		}
		str += fmt.Sprintf("%v", d.At(i))
	}
	str += "]"

	return str
}

// Begin returns an iterator of the deque with the first position
func (d *Deque) Begin() *DequeIterator {
	return d.First()
}

// End returns an iterator of the deque with the position d.Size()
func (d *Deque) End() *DequeIterator {
	return d.IterAt(d.Size())
}

// First returns an iterator of the deque with the first position
func (d *Deque) First() *DequeIterator {
	return d.IterAt(0)
}

// Last returns an iterator of the deque with the last position
func (d *Deque) Last() *DequeIterator {
	return d.IterAt(d.Size() - 1)
}

// IterAt returns an iterator of the deque with the position pos
func (d *Deque) IterAt(pos int) *DequeIterator {
	return &DequeIterator{dq: d, position: pos}
}
