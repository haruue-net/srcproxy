package srcproxy

const (
	Version = 0x01
)

const (
	OpCodeRelayRequest = 0x01
)

type RelayRequest struct {
	Version uint8
	OpCode  uint8

	AuthLen uint8 `struc:"uint8,sizeof=Auth"`
	Auth    []byte

	DstAddrLen uint8 `struc:"uint8,sizeof=DstAddr"`
	DstAddr    []byte
	DstPort    uint16

	SrcAddrLen uint8 `struc:"uint8,sizeof=SrcAddr"`
	SrcAddr    []byte
	SrcPort    uint16
}
