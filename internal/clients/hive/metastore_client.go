package hive

import gohive "github.com/beltran/gohive/v2"

func NewHMSClient(host string, port int) (*gohive.HiveMetastoreClient, error) {
	config := gohive.NewMetastoreConnectConfiguration()
	client, err := gohive.ConnectToMetastore(host, port, "KERBEROS", config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
