// Code generated by protoc-gen-gogo.
// source: common.proto
// DO NOT EDIT!

/*
	Package api is a generated protocol buffer package.

	It is generated from these files:
		common.proto

	It has these top-level messages:
		Process
		User
*/
package api

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/golang/protobuf/ptypes/empty"

import strings "strings"
import github_com_gogo_protobuf_proto "github.com/gogo/protobuf/proto"
import sort "sort"
import strconv "strconv"
import reflect "reflect"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type State int32

const (
	State_UNKNOWN State = 0
	State_CREATED State = 1
	State_RUNNING State = 2
	State_STOPPED State = 3
	State_PAUSED  State = 4
)

var State_name = map[int32]string{
	0: "UNKNOWN",
	1: "CREATED",
	2: "RUNNING",
	3: "STOPPED",
	4: "PAUSED",
}
var State_value = map[string]int32{
	"UNKNOWN": 0,
	"CREATED": 1,
	"RUNNING": 2,
	"STOPPED": 3,
	"PAUSED":  4,
}

func (x State) String() string {
	return proto.EnumName(State_name, int32(x))
}
func (State) EnumDescriptor() ([]byte, []int) { return fileDescriptorCommon, []int{0} }

type Process struct {
	Pid        uint32   `protobuf:"varint,1,opt,name=pid,proto3" json:"pid,omitempty"`
	Args       []string `protobuf:"bytes,2,rep,name=args" json:"args,omitempty"`
	Env        []string `protobuf:"bytes,3,rep,name=env" json:"env,omitempty"`
	User       *User    `protobuf:"bytes,4,opt,name=user" json:"user,omitempty"`
	Cwd        string   `protobuf:"bytes,5,opt,name=cwd,proto3" json:"cwd,omitempty"`
	Terminal   bool     `protobuf:"varint,6,opt,name=terminal,proto3" json:"terminal,omitempty"`
	State      State    `protobuf:"varint,7,opt,name=state,proto3,enum=containerd.v1.State" json:"state,omitempty"`
	ExitStatus uint32   `protobuf:"varint,8,opt,name=exit_status,json=exitStatus,proto3" json:"exit_status,omitempty"`
}

func (m *Process) Reset()                    { *m = Process{} }
func (*Process) ProtoMessage()               {}
func (*Process) Descriptor() ([]byte, []int) { return fileDescriptorCommon, []int{0} }

type User struct {
	Uid            uint32   `protobuf:"varint,1,opt,name=uid,proto3" json:"uid,omitempty"`
	Gid            uint32   `protobuf:"varint,2,opt,name=gid,proto3" json:"gid,omitempty"`
	AdditionalGids []uint32 `protobuf:"varint,3,rep,packed,name=additionalGids" json:"additionalGids,omitempty"`
}

func (m *User) Reset()                    { *m = User{} }
func (*User) ProtoMessage()               {}
func (*User) Descriptor() ([]byte, []int) { return fileDescriptorCommon, []int{1} }

func init() {
	proto.RegisterType((*Process)(nil), "containerd.v1.Process")
	proto.RegisterType((*User)(nil), "containerd.v1.User")
	proto.RegisterEnum("containerd.v1.State", State_name, State_value)
}
func (this *Process) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 12)
	s = append(s, "&api.Process{")
	s = append(s, "Pid: "+fmt.Sprintf("%#v", this.Pid)+",\n")
	s = append(s, "Args: "+fmt.Sprintf("%#v", this.Args)+",\n")
	s = append(s, "Env: "+fmt.Sprintf("%#v", this.Env)+",\n")
	if this.User != nil {
		s = append(s, "User: "+fmt.Sprintf("%#v", this.User)+",\n")
	}
	s = append(s, "Cwd: "+fmt.Sprintf("%#v", this.Cwd)+",\n")
	s = append(s, "Terminal: "+fmt.Sprintf("%#v", this.Terminal)+",\n")
	s = append(s, "State: "+fmt.Sprintf("%#v", this.State)+",\n")
	s = append(s, "ExitStatus: "+fmt.Sprintf("%#v", this.ExitStatus)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *User) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 7)
	s = append(s, "&api.User{")
	s = append(s, "Uid: "+fmt.Sprintf("%#v", this.Uid)+",\n")
	s = append(s, "Gid: "+fmt.Sprintf("%#v", this.Gid)+",\n")
	s = append(s, "AdditionalGids: "+fmt.Sprintf("%#v", this.AdditionalGids)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringCommon(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func extensionToGoStringCommon(m github_com_gogo_protobuf_proto.Message) string {
	e := github_com_gogo_protobuf_proto.GetUnsafeExtensionsMap(m)
	if e == nil {
		return "nil"
	}
	s := "proto.NewUnsafeXXX_InternalExtensions(map[int32]proto.Extension{"
	keys := make([]int, 0, len(e))
	for k := range e {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	ss := []string{}
	for _, k := range keys {
		ss = append(ss, strconv.Itoa(k)+": "+e[int32(k)].GoString())
	}
	s += strings.Join(ss, ",") + "})"
	return s
}
func (m *Process) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Process) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Pid != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintCommon(dAtA, i, uint64(m.Pid))
	}
	if len(m.Args) > 0 {
		for _, s := range m.Args {
			dAtA[i] = 0x12
			i++
			l = len(s)
			for l >= 1<<7 {
				dAtA[i] = uint8(uint64(l)&0x7f | 0x80)
				l >>= 7
				i++
			}
			dAtA[i] = uint8(l)
			i++
			i += copy(dAtA[i:], s)
		}
	}
	if len(m.Env) > 0 {
		for _, s := range m.Env {
			dAtA[i] = 0x1a
			i++
			l = len(s)
			for l >= 1<<7 {
				dAtA[i] = uint8(uint64(l)&0x7f | 0x80)
				l >>= 7
				i++
			}
			dAtA[i] = uint8(l)
			i++
			i += copy(dAtA[i:], s)
		}
	}
	if m.User != nil {
		dAtA[i] = 0x22
		i++
		i = encodeVarintCommon(dAtA, i, uint64(m.User.Size()))
		n1, err := m.User.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	if len(m.Cwd) > 0 {
		dAtA[i] = 0x2a
		i++
		i = encodeVarintCommon(dAtA, i, uint64(len(m.Cwd)))
		i += copy(dAtA[i:], m.Cwd)
	}
	if m.Terminal {
		dAtA[i] = 0x30
		i++
		if m.Terminal {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i++
	}
	if m.State != 0 {
		dAtA[i] = 0x38
		i++
		i = encodeVarintCommon(dAtA, i, uint64(m.State))
	}
	if m.ExitStatus != 0 {
		dAtA[i] = 0x40
		i++
		i = encodeVarintCommon(dAtA, i, uint64(m.ExitStatus))
	}
	return i, nil
}

func (m *User) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *User) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Uid != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintCommon(dAtA, i, uint64(m.Uid))
	}
	if m.Gid != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintCommon(dAtA, i, uint64(m.Gid))
	}
	if len(m.AdditionalGids) > 0 {
		dAtA3 := make([]byte, len(m.AdditionalGids)*10)
		var j2 int
		for _, num := range m.AdditionalGids {
			for num >= 1<<7 {
				dAtA3[j2] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j2++
			}
			dAtA3[j2] = uint8(num)
			j2++
		}
		dAtA[i] = 0x1a
		i++
		i = encodeVarintCommon(dAtA, i, uint64(j2))
		i += copy(dAtA[i:], dAtA3[:j2])
	}
	return i, nil
}

func encodeFixed64Common(dAtA []byte, offset int, v uint64) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	dAtA[offset+4] = uint8(v >> 32)
	dAtA[offset+5] = uint8(v >> 40)
	dAtA[offset+6] = uint8(v >> 48)
	dAtA[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Common(dAtA []byte, offset int, v uint32) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintCommon(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Process) Size() (n int) {
	var l int
	_ = l
	if m.Pid != 0 {
		n += 1 + sovCommon(uint64(m.Pid))
	}
	if len(m.Args) > 0 {
		for _, s := range m.Args {
			l = len(s)
			n += 1 + l + sovCommon(uint64(l))
		}
	}
	if len(m.Env) > 0 {
		for _, s := range m.Env {
			l = len(s)
			n += 1 + l + sovCommon(uint64(l))
		}
	}
	if m.User != nil {
		l = m.User.Size()
		n += 1 + l + sovCommon(uint64(l))
	}
	l = len(m.Cwd)
	if l > 0 {
		n += 1 + l + sovCommon(uint64(l))
	}
	if m.Terminal {
		n += 2
	}
	if m.State != 0 {
		n += 1 + sovCommon(uint64(m.State))
	}
	if m.ExitStatus != 0 {
		n += 1 + sovCommon(uint64(m.ExitStatus))
	}
	return n
}

func (m *User) Size() (n int) {
	var l int
	_ = l
	if m.Uid != 0 {
		n += 1 + sovCommon(uint64(m.Uid))
	}
	if m.Gid != 0 {
		n += 1 + sovCommon(uint64(m.Gid))
	}
	if len(m.AdditionalGids) > 0 {
		l = 0
		for _, e := range m.AdditionalGids {
			l += sovCommon(uint64(e))
		}
		n += 1 + sovCommon(uint64(l)) + l
	}
	return n
}

func sovCommon(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozCommon(x uint64) (n int) {
	return sovCommon(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *Process) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&Process{`,
		`Pid:` + fmt.Sprintf("%v", this.Pid) + `,`,
		`Args:` + fmt.Sprintf("%v", this.Args) + `,`,
		`Env:` + fmt.Sprintf("%v", this.Env) + `,`,
		`User:` + strings.Replace(fmt.Sprintf("%v", this.User), "User", "User", 1) + `,`,
		`Cwd:` + fmt.Sprintf("%v", this.Cwd) + `,`,
		`Terminal:` + fmt.Sprintf("%v", this.Terminal) + `,`,
		`State:` + fmt.Sprintf("%v", this.State) + `,`,
		`ExitStatus:` + fmt.Sprintf("%v", this.ExitStatus) + `,`,
		`}`,
	}, "")
	return s
}
func (this *User) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&User{`,
		`Uid:` + fmt.Sprintf("%v", this.Uid) + `,`,
		`Gid:` + fmt.Sprintf("%v", this.Gid) + `,`,
		`AdditionalGids:` + fmt.Sprintf("%v", this.AdditionalGids) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringCommon(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *Process) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCommon
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Process: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Process: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Pid", wireType)
			}
			m.Pid = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Pid |= (uint32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Args", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCommon
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Args = append(m.Args, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Env", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCommon
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Env = append(m.Env, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field User", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthCommon
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.User == nil {
				m.User = &User{}
			}
			if err := m.User.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Cwd", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCommon
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Cwd = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Terminal", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Terminal = bool(v != 0)
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field State", wireType)
			}
			m.State = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.State |= (State(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExitStatus", wireType)
			}
			m.ExitStatus = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ExitStatus |= (uint32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipCommon(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCommon
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *User) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCommon
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: User: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: User: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Uid", wireType)
			}
			m.Uid = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Uid |= (uint32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Gid", wireType)
			}
			m.Gid = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Gid |= (uint32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowCommon
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= (int(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthCommon
				}
				postIndex := iNdEx + packedLen
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				for iNdEx < postIndex {
					var v uint32
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowCommon
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= (uint32(b) & 0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.AdditionalGids = append(m.AdditionalGids, v)
				}
			} else if wireType == 0 {
				var v uint32
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowCommon
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= (uint32(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.AdditionalGids = append(m.AdditionalGids, v)
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field AdditionalGids", wireType)
			}
		default:
			iNdEx = preIndex
			skippy, err := skipCommon(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCommon
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipCommon(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowCommon
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthCommon
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowCommon
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipCommon(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthCommon = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowCommon   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("common.proto", fileDescriptorCommon) }

var fileDescriptorCommon = []byte{
	// 372 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x5c, 0x91, 0x4f, 0xef, 0xd2, 0x30,
	0x1c, 0xc6, 0x29, 0x1b, 0x7f, 0x2c, 0x42, 0x96, 0xea, 0xa1, 0x41, 0x33, 0x17, 0x0e, 0xba, 0x70,
	0x18, 0x11, 0x5f, 0x01, 0xca, 0x42, 0x8c, 0xc9, 0x58, 0x3a, 0x16, 0x8f, 0x66, 0xac, 0x75, 0x69,
	0xc2, 0x56, 0xb2, 0x76, 0xa0, 0x37, 0x5f, 0x1e, 0x47, 0x8f, 0x1e, 0x65, 0x89, 0x77, 0x5f, 0x82,
	0x69, 0x31, 0x9a, 0x1f, 0xb7, 0xe7, 0xf9, 0x7c, 0x9f, 0xef, 0xf6, 0x7d, 0x52, 0xf8, 0x38, 0x17,
	0x65, 0x29, 0xaa, 0xe0, 0x58, 0x0b, 0x25, 0xd0, 0x38, 0x17, 0x95, 0xca, 0x78, 0xc5, 0x6a, 0x1a,
	0x9c, 0x5e, 0x4f, 0x9f, 0x15, 0x42, 0x14, 0x07, 0xb6, 0x30, 0xc3, 0x7d, 0xf3, 0x79, 0xc1, 0xca,
	0xa3, 0xfa, 0x7a, 0xcb, 0xce, 0x7e, 0x01, 0x38, 0x88, 0x6b, 0x91, 0x33, 0x29, 0x91, 0x03, 0xad,
	0x23, 0xa7, 0x18, 0x78, 0xc0, 0x1f, 0x13, 0x2d, 0x11, 0x82, 0x76, 0x56, 0x17, 0x12, 0x77, 0x3d,
	0xcb, 0x7f, 0x44, 0x8c, 0xd6, 0x29, 0x56, 0x9d, 0xb0, 0x65, 0x90, 0x96, 0xe8, 0x15, 0xb4, 0x1b,
	0xc9, 0x6a, 0x6c, 0x7b, 0xc0, 0x1f, 0x2d, 0x9f, 0x04, 0x0f, 0x7e, 0x1f, 0xa4, 0x92, 0xd5, 0xc4,
	0x04, 0xf4, 0x6a, 0x7e, 0xa6, 0xb8, 0xe7, 0x01, 0xbd, 0x9a, 0x9f, 0x29, 0x9a, 0xc2, 0xa1, 0x62,
	0x75, 0xc9, 0xab, 0xec, 0x80, 0xfb, 0x1e, 0xf0, 0x87, 0xe4, 0x9f, 0x47, 0x73, 0xd8, 0x93, 0x2a,
	0x53, 0x0c, 0x0f, 0x3c, 0xe0, 0x4f, 0x96, 0x4f, 0xef, 0xbe, 0x9b, 0xe8, 0x19, 0xb9, 0x45, 0xd0,
	0x0b, 0x38, 0x62, 0x5f, 0xb8, 0xfa, 0xa4, 0x5d, 0x23, 0xf1, 0xd0, 0x54, 0x80, 0x1a, 0x25, 0x86,
	0xcc, 0x08, 0xb4, 0xd3, 0xbf, 0x27, 0x34, 0xff, 0x3b, 0x36, 0x9c, 0x6a, 0x52, 0x70, 0x8a, 0xbb,
	0x37, 0x52, 0x70, 0x8a, 0x5e, 0xc2, 0x49, 0x46, 0x29, 0x57, 0x5c, 0x54, 0xd9, 0x61, 0xc3, 0xa9,
	0x34, 0x65, 0xc7, 0xe4, 0x8e, 0xce, 0x37, 0xb0, 0x67, 0x8e, 0x40, 0x23, 0x38, 0x48, 0xa3, 0x0f,
	0xd1, 0xf6, 0x63, 0xe4, 0x74, 0xb4, 0x79, 0x47, 0xc2, 0xd5, 0x2e, 0x5c, 0x3b, 0x40, 0x1b, 0x92,
	0x46, 0xd1, 0xfb, 0x68, 0xe3, 0x74, 0xb5, 0x49, 0x76, 0xdb, 0x38, 0x0e, 0xd7, 0x8e, 0x85, 0x20,
	0xec, 0xc7, 0xab, 0x34, 0x09, 0xd7, 0x8e, 0xfd, 0xf6, 0xf9, 0xe5, 0xea, 0x76, 0x7e, 0x5c, 0xdd,
	0xce, 0xef, 0xab, 0x0b, 0xbe, 0xb5, 0x2e, 0xb8, 0xb4, 0x2e, 0xf8, 0xde, 0xba, 0xe0, 0x67, 0xeb,
	0x82, 0x7d, 0xdf, 0xbc, 0xd4, 0x9b, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x68, 0x7a, 0x1a, 0x7a,
	0xe5, 0x01, 0x00, 0x00,
}
