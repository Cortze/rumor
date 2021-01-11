package gossip

import(
    "strings"
)

var BeaconBlock string = "/eth2/b5303f2a/beacon_block/ssz_snappy"
var BeaconAggregateProof string = "/eth2/b5303f2a/beacon_aggregate_and_proof/ssz_snappy"
var VoluntaryExit string = "/eth2/b5303f2a/voluntary_exit/ssz_snappy"
var ProposerSlashing string = "/eth2/b5303f2a/proposer_slashing/ssz_snappy"
var AttesterSlashing string = "/eth2/b5303f2a/attester_slashing/ssz_snappy"

var MedallaForkDigest string = "b5303f2a"


func GenerateEth2Topics(network string, topic string, encoding string) string {
    var forkDigest string
    if network == "mainnet" { // If network mainnet forkDigest = b5303f2a
        forkDigest = MedallaForkDigest
    } else {
        if strings.Contains(network, "0x") { // if the given network is a forkDigest check if it has the proper format for gossip topics
            forkDigest = strings.Replace(network, "0x", "", 1)
        } else {
            forkDigest = network
        }
    }
    topicComposedName := "/eth2/" + forkDigest + "/" + topic + "/" + encoding
    return topicComposedName
}
