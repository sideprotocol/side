// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: side/auction/auction.proto

package types

import (
	fmt "fmt"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	github_com_cosmos_gogoproto_types "github.com/cosmos/gogoproto/types"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type AssetType int32

const (
	AssetType_Bitcoin AssetType = 0
)

var AssetType_name = map[int32]string{
	0: "Bitcoin",
}

var AssetType_value = map[string]int32{
	"Bitcoin": 0,
}

func (x AssetType) String() string {
	return proto.EnumName(AssetType_name, int32(x))
}

func (AssetType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_f93c54ffdc046a31, []int{0}
}

type AuctionStatus int32

const (
	AuctionStatus_AuctionOpen  AuctionStatus = 0
	AuctionStatus_AuctionClose AuctionStatus = 1
)

var AuctionStatus_name = map[int32]string{
	0: "AuctionOpen",
	1: "AuctionClose",
}

var AuctionStatus_value = map[string]int32{
	"AuctionOpen":  0,
	"AuctionClose": 1,
}

func (x AuctionStatus) String() string {
	return proto.EnumName(AuctionStatus_name, int32(x))
}

func (AuctionStatus) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_f93c54ffdc046a31, []int{1}
}

type BidStatus int32

const (
	BidStatus_Bidding  BidStatus = 0
	BidStatus_Accepted BidStatus = 1
	BidStatus_Rejected BidStatus = 2
)

var BidStatus_name = map[int32]string{
	0: "Bidding",
	1: "Accepted",
	2: "Rejected",
}

var BidStatus_value = map[string]int32{
	"Bidding":  0,
	"Accepted": 1,
	"Rejected": 2,
}

func (x BidStatus) String() string {
	return proto.EnumName(BidStatus_name, int32(x))
}

func (BidStatus) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_f93c54ffdc046a31, []int{2}
}

type Bid struct {
	Id        uint64      `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	AuctionId uint64      `protobuf:"varint,2,opt,name=auction_id,json=auctionId,proto3" json:"auction_id,omitempty"`
	Bidder    string      `protobuf:"bytes,3,opt,name=bidder,proto3" json:"bidder,omitempty"`
	BidPrice  int64       `protobuf:"varint,4,opt,name=bid_price,json=bidPrice,proto3" json:"bid_price,omitempty"`
	BidAmount *types.Coin `protobuf:"bytes,5,opt,name=bid_amount,json=bidAmount,proto3" json:"bid_amount,omitempty"`
	Status    BidStatus   `protobuf:"varint,6,opt,name=status,proto3,enum=side.auction.BidStatus" json:"status,omitempty"`
}

func (m *Bid) Reset()         { *m = Bid{} }
func (m *Bid) String() string { return proto.CompactTextString(m) }
func (*Bid) ProtoMessage()    {}
func (*Bid) Descriptor() ([]byte, []int) {
	return fileDescriptor_f93c54ffdc046a31, []int{0}
}
func (m *Bid) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Bid) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Bid.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Bid) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Bid.Merge(m, src)
}
func (m *Bid) XXX_Size() int {
	return m.Size()
}
func (m *Bid) XXX_DiscardUnknown() {
	xxx_messageInfo_Bid.DiscardUnknown(m)
}

var xxx_messageInfo_Bid proto.InternalMessageInfo

func (m *Bid) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Bid) GetAuctionId() uint64 {
	if m != nil {
		return m.AuctionId
	}
	return 0
}

func (m *Bid) GetBidder() string {
	if m != nil {
		return m.Bidder
	}
	return ""
}

func (m *Bid) GetBidPrice() int64 {
	if m != nil {
		return m.BidPrice
	}
	return 0
}

func (m *Bid) GetBidAmount() *types.Coin {
	if m != nil {
		return m.BidAmount
	}
	return nil
}

func (m *Bid) GetStatus() BidStatus {
	if m != nil {
		return m.Status
	}
	return BidStatus_Bidding
}

type Auction struct {
	Id              uint64        `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	DepositedAsset  *types.Coin   `protobuf:"bytes,2,opt,name=deposited_asset,json=depositedAsset,proto3" json:"deposited_asset,omitempty"`
	Borrower        string        `protobuf:"bytes,3,opt,name=borrower,proto3" json:"borrower,omitempty"`
	LiquidatedPrice int64         `protobuf:"varint,4,opt,name=liquidated_price,json=liquidatedPrice,proto3" json:"liquidated_price,omitempty"`
	LiquidatedTime  time.Time     `protobuf:"bytes,5,opt,name=liquidated_time,json=liquidatedTime,proto3,stdtime" json:"liquidated_time"`
	ExpectedValue   int64         `protobuf:"varint,6,opt,name=expected_value,json=expectedValue,proto3" json:"expected_value,omitempty"`
	BiddedValue     int64         `protobuf:"varint,7,opt,name=bidded_value,json=biddedValue,proto3" json:"bidded_value,omitempty"`
	PaymentTxId     string        `protobuf:"bytes,8,opt,name=payment_tx_id,json=paymentTxId,proto3" json:"payment_tx_id,omitempty"`
	Status          AuctionStatus `protobuf:"varint,9,opt,name=status,proto3,enum=side.auction.AuctionStatus" json:"status,omitempty"`
}

func (m *Auction) Reset()         { *m = Auction{} }
func (m *Auction) String() string { return proto.CompactTextString(m) }
func (*Auction) ProtoMessage()    {}
func (*Auction) Descriptor() ([]byte, []int) {
	return fileDescriptor_f93c54ffdc046a31, []int{1}
}
func (m *Auction) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Auction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Auction.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Auction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Auction.Merge(m, src)
}
func (m *Auction) XXX_Size() int {
	return m.Size()
}
func (m *Auction) XXX_DiscardUnknown() {
	xxx_messageInfo_Auction.DiscardUnknown(m)
}

var xxx_messageInfo_Auction proto.InternalMessageInfo

func (m *Auction) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Auction) GetDepositedAsset() *types.Coin {
	if m != nil {
		return m.DepositedAsset
	}
	return nil
}

func (m *Auction) GetBorrower() string {
	if m != nil {
		return m.Borrower
	}
	return ""
}

func (m *Auction) GetLiquidatedPrice() int64 {
	if m != nil {
		return m.LiquidatedPrice
	}
	return 0
}

func (m *Auction) GetLiquidatedTime() time.Time {
	if m != nil {
		return m.LiquidatedTime
	}
	return time.Time{}
}

func (m *Auction) GetExpectedValue() int64 {
	if m != nil {
		return m.ExpectedValue
	}
	return 0
}

func (m *Auction) GetBiddedValue() int64 {
	if m != nil {
		return m.BiddedValue
	}
	return 0
}

func (m *Auction) GetPaymentTxId() string {
	if m != nil {
		return m.PaymentTxId
	}
	return ""
}

func (m *Auction) GetStatus() AuctionStatus {
	if m != nil {
		return m.Status
	}
	return AuctionStatus_AuctionOpen
}

func init() {
	proto.RegisterEnum("side.auction.AssetType", AssetType_name, AssetType_value)
	proto.RegisterEnum("side.auction.AuctionStatus", AuctionStatus_name, AuctionStatus_value)
	proto.RegisterEnum("side.auction.BidStatus", BidStatus_name, BidStatus_value)
	proto.RegisterType((*Bid)(nil), "side.auction.Bid")
	proto.RegisterType((*Auction)(nil), "side.auction.Auction")
}

func init() { proto.RegisterFile("side/auction/auction.proto", fileDescriptor_f93c54ffdc046a31) }

var fileDescriptor_f93c54ffdc046a31 = []byte{
	// 576 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x53, 0xc1, 0x6f, 0xd3, 0x3e,
	0x14, 0x8e, 0xdb, 0xfd, 0xba, 0xe6, 0xb5, 0xeb, 0x22, 0xeb, 0x27, 0x08, 0x9d, 0xc8, 0xca, 0x24,
	0xa4, 0x32, 0x21, 0x47, 0xdb, 0x38, 0x70, 0x6d, 0x26, 0x21, 0xed, 0x80, 0x40, 0x61, 0xe2, 0xc0,
	0xa5, 0x4a, 0x62, 0x13, 0x8c, 0x9a, 0x38, 0xd4, 0xce, 0xe8, 0xfe, 0x8b, 0x9d, 0xf8, 0x9b, 0x76,
	0xdc, 0x81, 0x03, 0x27, 0x40, 0xdb, 0x3f, 0x82, 0xec, 0xb8, 0xd9, 0x26, 0x24, 0x4e, 0xed, 0xfb,
	0xbe, 0xef, 0xd9, 0xef, 0x7d, 0x9f, 0x03, 0x63, 0xc9, 0x29, 0x0b, 0x93, 0x3a, 0x53, 0x5c, 0x94,
	0xeb, 0x5f, 0x52, 0x2d, 0x85, 0x12, 0x78, 0xa8, 0x39, 0x62, 0xb1, 0xf1, 0xff, 0xb9, 0xc8, 0x85,
	0x21, 0x42, 0xfd, 0xaf, 0xd1, 0x8c, 0x77, 0x73, 0x21, 0xf2, 0x05, 0x0b, 0x4d, 0x95, 0xd6, 0x1f,
	0x43, 0xc5, 0x0b, 0x26, 0x55, 0x52, 0x54, 0x56, 0x10, 0x64, 0x42, 0x16, 0x42, 0x86, 0x69, 0x22,
	0x59, 0x78, 0x76, 0x90, 0x32, 0x95, 0x1c, 0x84, 0x99, 0xe0, 0xf6, 0x92, 0xbd, 0xef, 0x08, 0xba,
	0x11, 0xa7, 0x78, 0x04, 0x1d, 0x4e, 0x7d, 0x34, 0x41, 0xd3, 0x8d, 0xb8, 0xc3, 0x29, 0x7e, 0x0c,
	0x60, 0x6f, 0x9e, 0x73, 0xea, 0x77, 0x0c, 0xee, 0x5a, 0xe4, 0x84, 0xe2, 0x07, 0xd0, 0x4b, 0x39,
	0xa5, 0x6c, 0xe9, 0x77, 0x27, 0x68, 0xea, 0xc6, 0xb6, 0xc2, 0x3b, 0xe0, 0xa6, 0x9c, 0xce, 0xab,
	0x25, 0xcf, 0x98, 0xbf, 0x31, 0x41, 0xd3, 0x6e, 0xdc, 0x4f, 0x39, 0x7d, 0xab, 0x6b, 0xfc, 0x12,
	0x40, 0x93, 0x49, 0x21, 0xea, 0x52, 0xf9, 0xff, 0x4d, 0xd0, 0x74, 0x70, 0xf8, 0x88, 0x34, 0x03,
	0x12, 0x3d, 0x20, 0xb1, 0x03, 0x92, 0x63, 0xc1, 0xcb, 0x58, 0x9f, 0x34, 0x33, 0x5a, 0x1c, 0x42,
	0x4f, 0xaa, 0x44, 0xd5, 0xd2, 0xef, 0x4d, 0xd0, 0x74, 0x74, 0xf8, 0x90, 0xdc, 0xf5, 0x86, 0x44,
	0x9c, 0xbe, 0x33, 0x74, 0x6c, 0x65, 0x7b, 0xdf, 0xba, 0xb0, 0x39, 0x6b, 0xd8, 0xbf, 0x56, 0x8b,
	0x60, 0x9b, 0xb2, 0x4a, 0x48, 0xae, 0x18, 0x9d, 0x27, 0x52, 0x32, 0x65, 0xf6, 0xfb, 0xe7, 0x2c,
	0xa3, 0xb6, 0x63, 0xa6, 0x1b, 0xf0, 0x18, 0xfa, 0xa9, 0x58, 0x2e, 0xc5, 0xd7, 0xd6, 0x81, 0xb6,
	0xc6, 0xcf, 0xc0, 0x5b, 0xf0, 0x2f, 0x35, 0xa7, 0x89, 0xbe, 0xe0, 0xae, 0x15, 0xdb, 0xb7, 0x78,
	0xe3, 0xc8, 0x6b, 0xb8, 0x03, 0xcd, 0x75, 0x76, 0xd6, 0x96, 0x31, 0x69, 0x82, 0x25, 0xeb, 0x60,
	0xc9, 0xe9, 0x3a, 0xd8, 0xa8, 0x7f, 0xf9, 0x73, 0xd7, 0xb9, 0xf8, 0xb5, 0x8b, 0xe2, 0xd1, 0x6d,
	0xb3, 0xa6, 0xf1, 0x53, 0x18, 0xb1, 0x55, 0xc5, 0x32, 0x7d, 0xd8, 0x59, 0xb2, 0xa8, 0x99, 0xb1,
	0xab, 0x1b, 0x6f, 0xad, 0xd1, 0xf7, 0x1a, 0xc4, 0x4f, 0x60, 0x68, 0xe2, 0x5a, 0x8b, 0x36, 0x8d,
	0x68, 0xd0, 0x60, 0x8d, 0x64, 0x0f, 0xb6, 0xaa, 0xe4, 0xbc, 0x60, 0xa5, 0x9a, 0xab, 0x95, 0x7e,
	0x01, 0x7d, 0xb3, 0xe4, 0xc0, 0x82, 0xa7, 0xab, 0x13, 0x8a, 0x8f, 0xda, 0x50, 0x5c, 0x13, 0xca,
	0xce, 0xfd, 0x50, 0xac, 0xfd, 0xf7, 0x83, 0xd9, 0xf7, 0xc1, 0x35, 0x0e, 0x9e, 0x9e, 0x57, 0x0c,
	0x0f, 0x60, 0x33, 0xe2, 0x4a, 0xbf, 0x46, 0xcf, 0xd9, 0x3f, 0x84, 0xad, 0x7b, 0x2d, 0x78, 0x1b,
	0x06, 0x16, 0x78, 0x53, 0xb1, 0xd2, 0x73, 0xb0, 0x07, 0x43, 0x0b, 0x1c, 0x2f, 0x84, 0x64, 0x1e,
	0xda, 0x7f, 0x01, 0x6e, 0x9b, 0x7d, 0x73, 0x1a, 0xa5, 0xbc, 0xcc, 0x3d, 0x07, 0x0f, 0xa1, 0x3f,
	0xcb, 0x32, 0x56, 0x29, 0x46, 0x3d, 0xa4, 0xab, 0x98, 0x7d, 0x36, 0x16, 0x78, 0x9d, 0xe8, 0xd5,
	0xe5, 0x75, 0x80, 0xae, 0xae, 0x03, 0xf4, 0xfb, 0x3a, 0x40, 0x17, 0x37, 0x81, 0x73, 0x75, 0x13,
	0x38, 0x3f, 0x6e, 0x02, 0xe7, 0xc3, 0xf3, 0x9c, 0xab, 0x4f, 0x75, 0x4a, 0x32, 0x51, 0x84, 0x7a,
	0x19, 0xe3, 0x7e, 0x26, 0x16, 0xa6, 0x08, 0x57, 0xed, 0x87, 0xaa, 0xce, 0x2b, 0x26, 0xd3, 0x9e,
	0xa1, 0x8f, 0xfe, 0x04, 0x00, 0x00, 0xff, 0xff, 0xed, 0x4a, 0xd6, 0x87, 0xc5, 0x03, 0x00, 0x00,
}

func (m *Bid) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Bid) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Bid) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Status != 0 {
		i = encodeVarintAuction(dAtA, i, uint64(m.Status))
		i--
		dAtA[i] = 0x30
	}
	if m.BidAmount != nil {
		{
			size, err := m.BidAmount.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintAuction(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x2a
	}
	if m.BidPrice != 0 {
		i = encodeVarintAuction(dAtA, i, uint64(m.BidPrice))
		i--
		dAtA[i] = 0x20
	}
	if len(m.Bidder) > 0 {
		i -= len(m.Bidder)
		copy(dAtA[i:], m.Bidder)
		i = encodeVarintAuction(dAtA, i, uint64(len(m.Bidder)))
		i--
		dAtA[i] = 0x1a
	}
	if m.AuctionId != 0 {
		i = encodeVarintAuction(dAtA, i, uint64(m.AuctionId))
		i--
		dAtA[i] = 0x10
	}
	if m.Id != 0 {
		i = encodeVarintAuction(dAtA, i, uint64(m.Id))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *Auction) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Auction) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Auction) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Status != 0 {
		i = encodeVarintAuction(dAtA, i, uint64(m.Status))
		i--
		dAtA[i] = 0x48
	}
	if len(m.PaymentTxId) > 0 {
		i -= len(m.PaymentTxId)
		copy(dAtA[i:], m.PaymentTxId)
		i = encodeVarintAuction(dAtA, i, uint64(len(m.PaymentTxId)))
		i--
		dAtA[i] = 0x42
	}
	if m.BiddedValue != 0 {
		i = encodeVarintAuction(dAtA, i, uint64(m.BiddedValue))
		i--
		dAtA[i] = 0x38
	}
	if m.ExpectedValue != 0 {
		i = encodeVarintAuction(dAtA, i, uint64(m.ExpectedValue))
		i--
		dAtA[i] = 0x30
	}
	n2, err2 := github_com_cosmos_gogoproto_types.StdTimeMarshalTo(m.LiquidatedTime, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdTime(m.LiquidatedTime):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintAuction(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0x2a
	if m.LiquidatedPrice != 0 {
		i = encodeVarintAuction(dAtA, i, uint64(m.LiquidatedPrice))
		i--
		dAtA[i] = 0x20
	}
	if len(m.Borrower) > 0 {
		i -= len(m.Borrower)
		copy(dAtA[i:], m.Borrower)
		i = encodeVarintAuction(dAtA, i, uint64(len(m.Borrower)))
		i--
		dAtA[i] = 0x1a
	}
	if m.DepositedAsset != nil {
		{
			size, err := m.DepositedAsset.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintAuction(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if m.Id != 0 {
		i = encodeVarintAuction(dAtA, i, uint64(m.Id))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintAuction(dAtA []byte, offset int, v uint64) int {
	offset -= sovAuction(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Bid) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Id != 0 {
		n += 1 + sovAuction(uint64(m.Id))
	}
	if m.AuctionId != 0 {
		n += 1 + sovAuction(uint64(m.AuctionId))
	}
	l = len(m.Bidder)
	if l > 0 {
		n += 1 + l + sovAuction(uint64(l))
	}
	if m.BidPrice != 0 {
		n += 1 + sovAuction(uint64(m.BidPrice))
	}
	if m.BidAmount != nil {
		l = m.BidAmount.Size()
		n += 1 + l + sovAuction(uint64(l))
	}
	if m.Status != 0 {
		n += 1 + sovAuction(uint64(m.Status))
	}
	return n
}

func (m *Auction) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Id != 0 {
		n += 1 + sovAuction(uint64(m.Id))
	}
	if m.DepositedAsset != nil {
		l = m.DepositedAsset.Size()
		n += 1 + l + sovAuction(uint64(l))
	}
	l = len(m.Borrower)
	if l > 0 {
		n += 1 + l + sovAuction(uint64(l))
	}
	if m.LiquidatedPrice != 0 {
		n += 1 + sovAuction(uint64(m.LiquidatedPrice))
	}
	l = github_com_cosmos_gogoproto_types.SizeOfStdTime(m.LiquidatedTime)
	n += 1 + l + sovAuction(uint64(l))
	if m.ExpectedValue != 0 {
		n += 1 + sovAuction(uint64(m.ExpectedValue))
	}
	if m.BiddedValue != 0 {
		n += 1 + sovAuction(uint64(m.BiddedValue))
	}
	l = len(m.PaymentTxId)
	if l > 0 {
		n += 1 + l + sovAuction(uint64(l))
	}
	if m.Status != 0 {
		n += 1 + sovAuction(uint64(m.Status))
	}
	return n
}

func sovAuction(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozAuction(x uint64) (n int) {
	return sovAuction(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Bid) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAuction
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Bid: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Bid: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			m.Id = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuction
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Id |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AuctionId", wireType)
			}
			m.AuctionId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuction
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.AuctionId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Bidder", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuction
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAuction
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAuction
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Bidder = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BidPrice", wireType)
			}
			m.BidPrice = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuction
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BidPrice |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BidAmount", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuction
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthAuction
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthAuction
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.BidAmount == nil {
				m.BidAmount = &types.Coin{}
			}
			if err := m.BidAmount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Status", wireType)
			}
			m.Status = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuction
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Status |= BidStatus(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipAuction(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAuction
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
func (m *Auction) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAuction
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Auction: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Auction: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			m.Id = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuction
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Id |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DepositedAsset", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuction
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthAuction
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthAuction
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.DepositedAsset == nil {
				m.DepositedAsset = &types.Coin{}
			}
			if err := m.DepositedAsset.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Borrower", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuction
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAuction
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAuction
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Borrower = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LiquidatedPrice", wireType)
			}
			m.LiquidatedPrice = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuction
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.LiquidatedPrice |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LiquidatedTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuction
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthAuction
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthAuction
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdTimeUnmarshal(&m.LiquidatedTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExpectedValue", wireType)
			}
			m.ExpectedValue = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuction
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ExpectedValue |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BiddedValue", wireType)
			}
			m.BiddedValue = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuction
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BiddedValue |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PaymentTxId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuction
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAuction
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAuction
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PaymentTxId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 9:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Status", wireType)
			}
			m.Status = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuction
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Status |= AuctionStatus(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipAuction(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAuction
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
func skipAuction(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowAuction
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
					return 0, ErrIntOverflowAuction
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowAuction
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
			if length < 0 {
				return 0, ErrInvalidLengthAuction
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupAuction
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthAuction
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthAuction        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowAuction          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupAuction = fmt.Errorf("proto: unexpected end of group")
)
