package utilities

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
)

//GetIP retrieves the IP address from an http.Request
func GetIP(r *http.Request) string {
	if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
		return ipProxy
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

//SrvConfiguration represents a container with the server's configuration
type SrvConfiguration struct {
	AuthDbConnectionString string              `json:"authDb"`
	QueueConnectionUrl     string              `json:"queueUrl"`
	StorageDbParams        cassandraParameters `json:"storageDb"`
}
type cassandraParameters struct {
	ClusterUrls string `json:"clusterUrls"`
	Keyspace    string `json:"keyspace"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

//LoadJSONConfig returns a SrvConfiguration struct from a json file
func LoadJSONConfig() (*SrvConfiguration, error) {
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return nil, err
	}

	var config SrvConfiguration
	json.Unmarshal(file, &config)

	return &config, nil
}
