package kinesis

import (
	"context"
	"sync"

	"github.com/DiscreteTom/rua"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
)

func NewKinesisPeer(region string, gs rua.GameServer, streamName string, partitionKey string) *rua.BasicPeer {
	lock := sync.Mutex{}
	var client *kinesis.Client

	return rua.NewBasicPeer(gs).
		WithTag("kinesis").
		OnWrite(func(data []byte, p *rua.BasicPeer) error {
			// prevent concurrent write
			lock.Lock()
			defer lock.Unlock()

			_, err := client.PutRecord(context.TODO(), &kinesis.PutRecordInput{
				Data:         data,
				PartitionKey: &partitionKey,
				StreamName:   &streamName,
			})
			return err
		}).
		OnClose(func(p *rua.BasicPeer) error {
			// wait after write finished
			lock.Lock()
			defer lock.Unlock()
			return nil
		}).
		OnStart(func(p *rua.BasicPeer) {
			cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
			if err != nil {
				p.GetLogger().Error(err)
				p.GetGameServer().RemovePeer(p.GetId())
				return
			}

			client = kinesis.NewFromConfig(cfg)
		})
}
