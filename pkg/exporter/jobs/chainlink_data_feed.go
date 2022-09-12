package jobs

import (
	"context"
	"time"

	"github.com/onrik/ethrpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ChainlinkDataFeed exposes metrics for ethereum chainlink data feed contract
type ChainlinkDataFeed struct {
	client                   *ethrpc.EthRPC
	log                      logrus.FieldLogger
	ChainlinkDataFeedBalance prometheus.GaugeVec
	addresses                []*AddressChainlinkDataFeed
}

type AddressChainlinkDataFeed struct {
	From     string `yaml:"from"`
	To       string `yaml:"to"`
	Contract string `yaml:"contract"`
	Name     string `yaml:"name"`
}

const (
	NameChainlinkDataFeed = "chainlink_data_feed"
)

func (n *ChainlinkDataFeed) Name() string {
	return NameChainlinkDataFeed
}

// NewChainlinkDataFeed returns a new ChainlinkDataFeed instance.
func NewChainlinkDataFeed(client *ethrpc.EthRPC, log logrus.FieldLogger, namespace string, constLabels map[string]string, addresses []*AddressChainlinkDataFeed) ChainlinkDataFeed {
	namespace += "_" + NameChainlinkDataFeed

	instance := ChainlinkDataFeed{
		client:    client,
		log:       log.WithField("module", NameChainlinkDataFeed),
		addresses: addresses,
		ChainlinkDataFeedBalance: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "balance",
				Help:        "The balance of a ethereum chainlink data feed contract.",
				ConstLabels: constLabels,
			},
			[]string{"name", "contract", "from", "to"},
		),
	}

	prometheus.MustRegister(instance.ChainlinkDataFeedBalance)

	return instance
}

func (n *ChainlinkDataFeed) Start(ctx context.Context) {
	n.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 15):
			n.tick(ctx)
		}
	}
}

//nolint:unparam // context will be used in the future
func (n *ChainlinkDataFeed) tick(ctx context.Context) {
	for _, address := range n.addresses {
		err := n.getBalance(address)

		if err != nil {
			n.log.WithError(err).WithField("address", address).Error("Failed to get chainlink data feed balance")
		}
	}
}

func (n *ChainlinkDataFeed) getBalance(address *AddressChainlinkDataFeed) error {
	balanceStr, err := n.client.EthCall(ethrpc.T{
		To:   address.Contract,
		From: "0x0000000000000000000000000000000000000000",
		// call latestAnswer() which is 0x50d25bcd
		Data: "0x50d25bcd000000000000000000000000",
	}, "latest")
	if err != nil {
		return err
	}

	n.ChainlinkDataFeedBalance.WithLabelValues(address.Name, address.Contract, address.From, address.To).Set(hexStringToFloat64(balanceStr))

	return nil
}