package protocol

import (
	"encoding/binary"
	"errors"
	"strings"
)

// 板块文件协议（地域/板块/概念/指数）。参考 pytdx get_block_info + block_reader。
const (
	TypeBlockMeta uint16 = 0x02C5 // 板块文件元信息（大小）
	TypeBlockInfo uint16 = 0x06B9 // 板块文件内容（分块）

	// 通达信板块文件名。
	BlockFileZS = "block_zs.dat" // 指数板块（行业）
	BlockFileGN = "block_gn.dat" // 概念板块
	BlockFileFG = "block_fg.dat" // 风格板块（含地域）
	BlockFile   = "block.dat"    // 默认板块

	blockChunk = 0x7530 // 单次下载块大小 30000
)

// BlockMetaResp 板块文件元信息。
type BlockMetaResp struct{ Size uint32 }

// BlockInfoResp 板块文件分块内容。
type BlockInfoResp struct{ Data []byte }

// Block 一个板块及其成分（Codes 为 7 字符，首字符为市场标志：1=沪 0=深）。
type Block struct {
	Name  string
	Type  uint16
	Codes []string
}

type block struct{}

// MBlock 板块协议单例。
var MBlock block

// FrameMeta 构造板块元信息请求帧。
func (block) FrameMeta(file string) *Frame {
	data := make([]byte, 0x2a-2) // 40 字节，文件名 null 填充
	copy(data, file)
	return &Frame{Control: Control01, Type: TypeBlockMeta, Data: data}
}

// FrameInfo 构造板块内容分块请求帧。
func (block) FrameInfo(start, size uint32, file string) *Frame {
	data := make([]byte, 8+0x6e-10) // start(4)+size(4)+filename(100)
	binary.LittleEndian.PutUint32(data[0:4], start)
	binary.LittleEndian.PutUint32(data[4:8], size)
	copy(data[8:], file)
	return &Frame{Control: Control01, Type: TypeBlockInfo, Data: data}
}

// DecodeMeta 解析元信息：前 4 字节为文件大小。
func (block) DecodeMeta(bs []byte) (*BlockMetaResp, error) {
	if len(bs) < 4 {
		return nil, errors.New("block meta 数据不足")
	}
	return &BlockMetaResp{Size: Uint32(bs[:4])}, nil
}

// DecodeInfo 解析分块内容：去掉前 4 字节（块长度），其余为文件内容。
func (block) DecodeInfo(bs []byte) (*BlockInfoResp, error) {
	if len(bs) < 4 {
		return &BlockInfoResp{Data: nil}, nil
	}
	return &BlockInfoResp{Data: bs[4:]}, nil
}

// ParseBlockFile 解析完整板块文件 → 板块列表。
// 格式：偏移 384 处 uint16 板块数；每块 = 名称(9,GBK) + 成分数(2) + 类型(2) + 成分(400×7)，定长 2813。
func ParseBlockFile(data []byte) []*Block {
	if len(data) < 386 {
		return nil
	}
	pos := 384
	num := int(Uint16(data[pos : pos+2]))
	pos += 2
	out := make([]*Block, 0, num)
	for i := 0; i < num; i++ {
		if pos+13 > len(data) {
			break
		}
		name := strings.TrimRight(string(UTF8ToGBK(data[pos:pos+9])), "\x00")
		pos += 9
		stockCount := int(Uint16(data[pos : pos+2]))
		blockType := Uint16(data[pos+2 : pos+4])
		pos += 4
		begin := pos
		codes := make([]string, 0, stockCount)
		for j := 0; j < stockCount; j++ {
			if pos+7 > len(data) {
				break
			}
			c := strings.TrimRight(string(data[pos:pos+7]), "\x00")
			if c != "" {
				codes = append(codes, c)
			}
			pos += 7
		}
		pos = begin + 400*7 // 每块成分区固定 2800 字节
		out = append(out, &Block{Name: name, Type: blockType, Codes: codes})
	}
	return out
}
