package kinesis

import (
	"context"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peer"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
)

type KinesisPeer struct {
	*peer.SafePeer
	region       string
	streamName   string
	partitionKey string
	client       *kinesis.Client
}

func NewKinesisPeer(region string, streamName string, partitionKey string, gs rua.GameServer) *KinesisPeer {
	kp := &KinesisPeer{
		SafePeer:     peer.NewSafePeer(gs),
		region:       region,
		streamName:   streamName,
		partitionKey: partitionKey,
		client:       nil,
	}

	kp.SafePeer.
		OnWriteSafe(func(data []byte) error {
			_, err := kp.client.PutRecord(context.TODO(), &kinesis.PutRecordInput{
				Data:         data,
				PartitionKey: &partitionKey,
				StreamName:   &streamName,
			})
			return err
		}).
		OnCloseSafe(func() error {
			return nil
		}).
		OnStart(func() {
			cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
			if err != nil {
				kp.Logger().Error(err)
				kp.GameServer().RemovePeer(kp.Id())
				return
			}

			kp.client = kinesis.NewFromConfig(cfg)
		}).
		WithTag("kinesis")

	return kp
}
