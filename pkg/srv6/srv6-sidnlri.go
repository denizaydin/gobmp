package srv6

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/base"
	"github.com/sbezverk/gobmp/pkg/tools"
)

// SIDNLRI defines SRv6 SID NLRI onject
// no RFC yet
type SIDNLRI struct {
	ProtocolID    base.ProtoID
	Identifier    []byte
	LocalNode     *base.NodeDescriptor `json:"local_node_descriptor,omitempty"`
	SRv6SID       *SIDDescriptor       `json:"sid_descriptor,omitempty"`
	LocalNodeHash string
}

// GetAllAttribute returns a slice with all attribute types found in SRv6 SID NLRI object
func (sr *SIDNLRI) GetAllAttribute() []uint16 {
	attrs := make([]uint16, 0)
	for _, attr := range sr.LocalNode.SubTLV {
		attrs = append(attrs, attr.Type)
	}

	return attrs
}

// GetSRv6SIDProtocolID returns a string representation of LinkNLRI ProtocolID field
func (sr *SIDNLRI) GetSRv6SIDProtocolID() string {
	return base.ProtocolIDString(sr.ProtocolID)
}

// GetSRv6SIDLSID returns a value of Local Node Descriptor TLV BGP-LS Identifier
func (sr *SIDNLRI) GetSRv6SIDLSID() uint32 {
	return sr.LocalNode.GetLSID()
}

// GetSRv6SIDIGPRouterID returns a value of a local node Descriptor TLV IGP Router ID
func (sr *SIDNLRI) GetSRv6SIDIGPRouterID() string {
	return sr.LocalNode.GetIGPRouterID()
}

// GetSRv6SIDASN returns Autonomous System Number used to uniqely identify BGP-LS domain
func (sr *SIDNLRI) GetSRv6SIDASN() uint32 {
	return sr.LocalNode.GetASN()
}

// GetSRv6SIDMTID returns Multi-Topology identifiers
func (sr *SIDNLRI) GetSRv6SIDMTID() uint16 {
	if sr.SRv6SID == nil {
		return 0
	}
	if sr.SRv6SID.MultiTopologyIdentifier == nil {
		return 0
	}

	return sr.SRv6SID.MultiTopologyIdentifier.GetMTID()[0]
}

// GetSRv6SID returns a slice of SIDs
func (sr *SIDNLRI) GetSRv6SID() string {
	return net.IP(sr.SRv6SID.SID).To16().String()
}

// UnmarshalSRv6SIDNLRI builds SRv6SIDNLRI NLRI object
func UnmarshalSRv6SIDNLRI(b []byte) (*SIDNLRI, error) {
	if glog.V(6) {
		glog.Infof("SRv6 SID NLRI Raw: %s", tools.MessageHex(b))
	}
	if len(b) == 0 {
		return nil, fmt.Errorf("NLRI length is 0")
	}
	sr := SIDNLRI{}
	p := 0
	sr.ProtocolID = base.ProtoID(b[p])
	p++
	// Skip reserved bytes
	//	p += 3
	sr.Identifier = make([]byte, 8)
	copy(sr.Identifier, b[p:p+8])
	p += 8
	// Get Node Descriptor's length, skip Node Descriptor Type
	l := binary.BigEndian.Uint16(b[p+2 : p+4])
	ln, err := base.UnmarshalNodeDescriptor(b[p : p+int(l)])
	if err != nil {
		return nil, err
	}
	sr.LocalNode = ln
	sr.LocalNodeHash = fmt.Sprintf("%x", md5.Sum(b[p:p+int(l)]))
	// Skip Node Descriptor Type and Length 4 bytes
	p += 4
	p += int(l)
	srd, err := UnmarshalSRv6SIDDescriptor(b[p:])
	if err != nil {
		return nil, err
	}
	sr.SRv6SID = srd

	return &sr, nil
}
