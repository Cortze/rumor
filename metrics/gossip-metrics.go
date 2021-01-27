package metrics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"
	"time"
    "strings"
    "net/http"

    "github.com/protolambda/rumor/p2p/gossip/database"
	"github.com/libp2p/go-libp2p-core/peer"
	pgossip "github.com/protolambda/rumor/p2p/gossip"
	"github.com/protolambda/rumor/p2p/track"
    "github.com/protolambda/zrnt/eth2/beacon"
)


type GossipMetrics struct {
    GossipMetrics   sync.Map
    TopicDatabase   database.TopicDatabase
//   PeerIdMap       map[string]peer.ID
//   NotChan         map[string]chan bool
}



func NewGossipMetrics(config *beacon.Spec) GossipMetrics{
    gm := GossipMetrics {
        TopicDatabase:  database.NewTopicDatabase(config),
        //PeerIdMap:      make(map[string]peer.ID),
        //c.NotChan   := make(map[string])
    }
    return gm
}


type GossipState struct {
	GsNode  pgossip.GossipSub
	CloseGS context.CancelFunc
	// string -> *pubsub.Topic
	Topics sync.Map
	// Metrics for Gossip Messages
}

// Base Struct for the topic name and the received messages on the different topics
type PeerMetrics struct {
	PeerId     peer.ID
	NodeId     string
	ClientType string
	Pubkey     string
	Addrs      string
	Ip         string
    Country    string
    City       string
    Latency    float64

	ConnectionEvents []ConnectionEvents
	// Counters for the different topics
	BeaconBlock          MessageMetrics
	BeaconAggregateProof MessageMetrics
	VoluntaryExit        MessageMetrics
	ProposerSlashing     MessageMetrics
	AttesterSlashing     MessageMetrics
	// Variables related to the SubNets (only needed for when Shards will be implemented)
}

func NewPeerMetrics(peerId peer.ID) PeerMetrics {
    pm := PeerMetrics {
        PeerId:     peerId,
        NodeId:     "",
        ClientType: "",
        Pubkey:     "",
        Addrs:      "",
        Ip:         "",
        Country:    "",
        City:       "",
        Latency:    0,

        ConnectionEvents:       make([]ConnectionEvents, 1),
        // Counters for the different topics
        BeaconBlock:            NewMessageMetrics(),
        BeaconAggregateProof:   NewMessageMetrics(),
        VoluntaryExit:          NewMessageMetrics(),
        ProposerSlashing:       NewMessageMetrics(),
        AttesterSlashing:       NewMessageMetrics(),
    }
    return pm
}

func NewMessageMetrics() MessageMetrics{
    mm := MessageMetrics {
        Cnt:    0,
        FirstMessageTime: 0,
        LastMessageTime: 0,
    }
    return mm
}

// Connection event model
type ConnectionEvents struct {
	ConnectionType string
	TimeMili       int64
}

// Information regarding the messages received on the beacon_lock topic
type MessageMetrics struct {
	Cnt              int64
	FirstMessageTime int64
	LastMessageTime  int64
}

// Function that Wraps/Marshals the content of the sync.Map to be exported as a json
func (c *GossipMetrics) MarshalMetrics() ([]byte, error) {
    tmpMap := make(map[string]PeerMetrics)
    c.GossipMetrics.Range(func(k, v interface{}) bool {
        tmpMap[k.(peer.ID).String()] = v.(PeerMetrics)
        return true
    })
    return json.Marshal(tmpMap)
}

// Function that Wraps/Marshals the content of the Entire Peerstore into a json
func (c *GossipMetrics) MarshalPeerStore(ep track.ExtendedPeerstore) ([]byte, error) {
    var peers []peer.ID
    peers = ep.Peers()
    peerData := make(map[string]*track.PeerAllData)
    for _, p := range peers {
        peerData[p.String()] = ep.GetAllData(p)
    }
    return json.Marshal(peerData)
}

// Get the Real Ip Address from the multi Address list
func GetFullAddress(MultiAddrs []string) string {
    var address string
    for _, element := range MultiAddrs {
        if strings.Contains(address, "192.168.") || strings.Contains(address, "127.0.0.0") {
            continue
        } else {
            address = element
        }
    }
    return address
}

// Function that iterates through the received peers and fills the missing information
func (c *GossipMetrics) FillMetrics(ep track.ExtendedPeerstore) {
    // to prevent the Filler from crashing (the url-service only accepts 45req/s)
    requestCounter := 0
    // Loop over the Peers on the GossipMetrics
    c.GossipMetrics.Range(func(key interface{}, value interface{}) bool{
        // Read the info that we have from him
        p, ok := c.GossipMetrics.Load(key)
        if ok {
            peerMetrics := p.(PeerMetrics)
            // start with the loop checking the info, First check the Peerstore Info
            if len(peerMetrics.ClientType) == 0 || peerMetrics.Latency != 0 || len(peerMetrics.Addrs) == 0 {
                peerData := ep.GetAllData(peerMetrics.PeerId)
                peerMetrics.NodeId      = peerData.NodeID.String()
                peerMetrics.ClientType  = peerData.UserAgent
                peerMetrics.Pubkey      = peerData.Pubkey

                address := GetFullAddress(peerData.Addrs)
                if len(address) > 0 {
                    peerMetrics.Addrs       = address
                    ip, country, city := getIpAndLocationFromAddrs(peerMetrics.Addrs)
                    // Increase the counter for preventing the DataFiller from crashing
                    requestCounter = requestCounter + 1
                    peerMetrics.Ip      = ip
                    peerMetrics.Country = country
                    peerMetrics.City    = city
                }

                peerMetrics.Latency = float64(peerData.Latency / 1*time.Millisecond)
                c.GossipMetrics.Store(peerMetrics.PeerId, peerMetrics)
            }

            if len(peerMetrics.Country) == 0 && len(peerMetrics.Ip) < 0 {
                ip, country, city := getIpAndLocationFromAddrs(peerMetrics.Addrs)
                // Increase the counter for preventing the DataFiller from crashing
                requestCounter = requestCounter + 1
                peerMetrics.Ip      = ip
                peerMetrics.Country = country
                peerMetrics.City    = city
                c.GossipMetrics.Store(peerMetrics.PeerId, peerMetrics)
            }
        }
        if requestCounter >= 45 { // Reminder 45 req/s
            time.Sleep(60 * time.Second)
            requestCounter = 0
        }
        // Keep with the loop on the Range function
        return true
    })

}

// Function that Exports the entire Metrics to a .json file (lets see if in the future we can add websockets or other implementations)
func (c *GossipMetrics) ExportMetrics(filePath string, peerstorePath string, ep track.ExtendedPeerstore) error {
    metrics, err := c.MarshalMetrics()
	if err != nil {
		fmt.Println("Error Marshalling the metrics")
	}
    fmt.Println("Loading Peerstore to Export them")
    peerstore, err := c.MarshalPeerStore(ep)
	if err != nil {
		fmt.Println("Error Marshalling the peerstore")
	}

    err = ioutil.WriteFile(filePath, metrics, 0644)
	if err != nil {
		fmt.Println("Error opening file: ", filePath)
		return err
	}
	err = ioutil.WriteFile(peerstorePath, peerstore, 0644)
	if err != nil {
		fmt.Println("Error opening file: ", peerstorePath)
		return err
	}
	return nil
}

// IP-API message structure
type IpApiMessage struct {
    Query       string
    Status      string
    Continent   string
    ContinentCode string
    Country     string
    CountryCode string
    Region      string
    RegionName  string
    City    string
    District    string
    Zip     string
    Lat         string
    Lon         string
    Timezone    string
    Offset      string
    Currency    string
    Isp     string
    Org     string
    As      string
    Asname  string
    Mobile  string
    Proxy   string
    Hosting string
}

// get IP, location country and City from the multiaddress of the peer on the peerstore
func getIpAndLocationFromAddrs(multiAddrs string) (ip string, country string, city string) {
    ip = strings.TrimPrefix(multiAddrs, "/ip4/")
    ipSlices := strings.Split(ip, "/")
    ip = ipSlices[0]
    url := "http://ip-api.com/json/" + ip
    resp, err := http.Get(url)
    if err != nil {
        fmt.Println("There was an error getting the Location of the peer from the IP-API, please, check that there 40requests/minute limit hasn't been exceed")
    }
    defer resp.Body.Close()
    bodyBytes, _ := ioutil.ReadAll(resp.Body)

    // Convert response body to Todo struct
    var ipApiResp IpApiMessage
    json.Unmarshal(bodyBytes, &ipApiResp)

    country = ipApiResp.Country
    city = ipApiResp.City
    // return the received values from the received message
    return ip, country, city

}

// Add new peer with all the information from the peerstore to the metrics db
// returns: Alredy (Bool)
func (c *GossipMetrics) AddNewPeer(peerId peer.ID) {
    _, ok := c.GossipMetrics.Load(peerId)
    if !ok {
        // We will just add the info that we have (the peerId)
        peerMetrics := NewPeerMetrics(peerId)
        // Include it to the Peer DB
        c.GossipMetrics.Store(peerId, peerMetrics)
        // return that wasn't already on the peerstore
    }
}

// Add a connection Event to the given peer
func (c *GossipMetrics) AddConnectionEvent(peerId peer.ID, connectionType string) {
	newConnection := ConnectionEvents{
		ConnectionType: connectionType,
		TimeMili:       GetTimeMiliseconds(),
	}
	pMetrics , ok := c.GossipMetrics.Load(peerId)
    if ok {
        peerMetrics := pMetrics.(PeerMetrics)
        peerMetrics.ConnectionEvents = append(peerMetrics.ConnectionEvents, newConnection)
        c.GossipMetrics.Store(peerId, peerMetrics)
    } else {
        // Might be possible to add 
        errors.New("Counld't add Event, Peer is not in the list")
	}
}


// Increments the counter of the topic
func (c *MessageMetrics) IncrementCnt() int64 {
	c.Cnt++
	return c.Cnt
}

// Stamps linux_time(millis) on the FirstMessageTime/LastMessageTime from given args: time (int64), flag string("first"/"last")
func (c *MessageMetrics) StampTime(flag string) {
	unixMillis := GetTimeMiliseconds()

	switch flag {
	case "first":
		c.FirstMessageTime = unixMillis
	case "last":
		c.LastMessageTime = unixMillis
	default:
		fmt.Println("Metrics Package -> StampTime.flag wrongly parsed")
	}
}

func GetTimeMiliseconds() int64 {
	now := time.Now()
	//secs := now.Unix()
	nanos := now.UnixNano()
	millis := nanos / 1000000

	return millis
}

// Function that Manages the metrics updates for the incoming messages
func (c *GossipMetrics) IncomingMessageManager(peerId peer.ID, topicName string) error {
    pMetrics , _ := c.GossipMetrics.Load(peerId)
    peerMetrics := pMetrics.(PeerMetrics)
	messageMetrics, err := GetMessageMetrics(&peerMetrics, topicName)
	if err != nil {
		return errors.New("Topic Name no supported")
	}
	fmt.Println(err, topicName, messageMetrics)
	if messageMetrics.Cnt == 0 {
		messageMetrics.StampTime("first")
	}

	messageMetrics.IncrementCnt()
	messageMetrics.StampTime("last")

    // Store back the Loaded/Modified Variable
    c.GossipMetrics.Store(peerId, peerMetrics)

	return nil
}

func GetMessageMetrics(c *PeerMetrics, topicName string) (mesMetr *MessageMetrics, err error) {
	// All this could be inside a different function
	switch topicName {
	case pgossip.BeaconBlock:
		return &c.BeaconBlock, nil
	case pgossip.BeaconAggregateProof:
		return &c.BeaconAggregateProof, nil
	case pgossip.VoluntaryExit:
		return &c.VoluntaryExit, nil
	case pgossip.ProposerSlashing:
		return &c.ProposerSlashing, nil
	case pgossip.AttesterSlashing:
		return &c.AttesterSlashing, nil
	default: //TODO: - Not returning BeaconBlock as Default
		return &c.BeaconBlock, err
	}
}
