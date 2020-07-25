package csgo

/*
	Provides access to CS:GO Game Coordinator functionality.
*/
import (
	"./protocol/protobuf"
	"github.com/Philipp15b/go-steam"
	. "github.com/Philipp15b/go-steam/protocol/gamecoordinator"
	"github.com/golang/protobuf/proto"
)

const AppId = 730

// To use any methods of this, you'll need to SetPlaying(true) and wait for
// the GCReadyEvent.
type CSGO struct {
	client *steam.Client
	isConnected bool
}

// Creates a new CSGO instance and registers it as a packet handler
func New(client *steam.Client) *CSGO {
	c := &CSGO{client, false}
	client.GC.RegisterPacketHandler(c)
	go c.handler()

	return c
}

func (c *CSGO) handler()  {
	for event := range c.client.Events() {
		switch e := event.(type) {
		case *steam.DisconnectedEvent:
			c.isConnected = false
		}
	}
}

func (c *CSGO) IsInCSGO() bool {
	return c.isConnected
}

func (c *CSGO) SetPlaying(playing bool) {
	if playing {
		c.client.GC.SetGamesPlayed(AppId)
		c.sendHello()
	} else {
		c.client.GC.SetGamesPlayed()
	}

	c.client.Disconnect()
}

func (c *CSGO) sendHello() {
	c.client.GC.Write(NewGCMsgProtobuf(AppId, uint32(protobuf.EGCBaseClientMsg_k_EMsgGCClientHello), &protobuf.CMsgClientHello{}))
}

func (c *CSGO) RequestItemData(s, a, d, m uint64) {
	c.client.GC.Write(NewGCMsgProtobuf(AppId, uint32(protobuf.ECsgoGCMsg_k_EMsgGCCStrike15_v2_Client2GCEconPreviewDataBlockRequest), &protobuf.CMsgGCCStrike15V2_Client2GCEconPreviewDataBlockRequest{
		ParamS: proto.Uint64(s),
		ParamA: proto.Uint64(a),
		ParamD: proto.Uint64(d),
		ParamM: proto.Uint64(m),
	}))
}

type GCReadyEvent struct{}

type GCItemDataEvent struct {
	ItemInfo *protobuf.CEconItemPreviewDataBlock `json:"iteminfo"`
}

func (c *CSGO) HandleGCPacket(packet *GCPacket) {
	if packet.AppId != AppId {
		return
	}

	switch protobuf.EGCBaseClientMsg(packet.MsgType) {
	case protobuf.EGCBaseClientMsg_k_EMsgGCClientWelcome:
		c.isConnected = true
		c.client.Emit(&GCReadyEvent{})
	case protobuf.EGCBaseClientMsg(protobuf.ECsgoGCMsg_k_EMsgGCCStrike15_v2_Client2GCEconPreviewDataBlockResponse):
		itemInfo := &protobuf.CEconItemPreviewDataBlock{}
		packet.ReadProtoMsg(itemInfo)
		c.client.Emit(&GCItemDataEvent{ItemInfo: itemInfo})
	}
}
