package gossip

import (
	// ----- temporary -----
	"fmt"
	// --- End temporary ---

	"strings"
)

// From given topicName and forkVersion, return the string of the gossipsub topic with the propper format
func Eth2TopicBuilder(topicName string, forkVersion string) (string, error) {

	fmt.Println("Parsed topicName:", topicName)
	fmt.Println("Parsed forkVersion:", forkVersion)

	if topicName == "" || forkVersion == "" {
		return "", fmt.Errorf("Topic Name or Fork version are empty. TopicName:", topicName, "Fork Version:", forkVersion)
	}
	// Check if there is any blank space on the topic name
	if strings.Contains(topicName, " ") {
		return "", fmt.Errorf("Parsed topic has a blank space, please remove the blank space")
	}
	// Check if the fork_digest_version has 0xXXXXXX, if true remove it
	if strings.Contains(forkVersion, "0x") {
		forkVersion = strings.Replace(forkVersion, "0x", "", -1)
	}
	// Compose the full name of the topic with /eth2/FORK_DIGEST/TOPIC_NAME/ssz_snappy
	eth2topicName := "/eth2/" + forkVersion + "/" + topicName + "/ssz_snappy"
	return eth2topicName, nil
}
