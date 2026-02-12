package p2p

import (
	"context"
	"discordia/internal/domain"
	"encoding/json"
	"sync"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
)

type Libp2pService struct {
	host  host.Host
	topic *pubsub.Topic
	sub   *pubsub.Subscription
	dht   *dht.IpfsDHT
}

func NewLibp2pService(ctx context.Context, h host.Host, ps *pubsub.PubSub, roomName string) *Libp2pService {
	// 1. Inicia a DHT em modo cliente/servidor (ajuda a rede a se achar)
	kademliaDHT, _ := dht.New(ctx, h)
	kademliaDHT.Bootstrap(ctx)

	// 2. Conecta aos nós de Bootstrap públicos (os "faróis")
	var wg sync.WaitGroup
	for _, addr := range dht.DefaultBootstrapPeers {
		pi, _ := peer.AddrInfoFromP2pAddr(addr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.Connect(ctx, *pi)
		}()
	}
	wg.Wait()

	// 3. Usa a DHT para anunciar nossa sala na internet
	routingDiscovery := routing.NewRoutingDiscovery(kademliaDHT)
	util.Advertise(ctx, routingDiscovery, roomName)

	// 4. Procura por outras pessoas na mesma sala globalmente
	go func() {
		for {
			peerChan, _ := routingDiscovery.FindPeers(ctx, roomName)
			for p := range peerChan {
				if p.ID == h.ID() {
					continue
				}
				h.Connect(ctx, p)
			}
		}
	}()

	topic, _ := ps.Join(roomName)
	sub, _ := topic.Subscribe()

	return &Libp2pService{host: h, topic: topic, sub: sub, dht: kademliaDHT}
}

// Publish transforma a struct em JSON e joga na rede
func (s *Libp2pService) Publish(msg *domain.Message) error {
	bytes, _ := json.Marshal(msg)
	return s.topic.Publish(context.Background(), bytes)
}

// Subscribe fica ouvindo a rede e chama um callback quando chega algo
func (s *Libp2pService) Subscribe(handler func(*domain.Message)) {
	go func() {
		for {
			msg, err := s.sub.Next(context.Background())
			if err != nil {
				return
			}
			// Não processa mensagens próprias
			if msg.ReceivedFrom == s.host.ID() {
				continue
			}
			var m domain.Message
			if err := json.Unmarshal(msg.Data, &m); err == nil {
				handler(&m)
			}
		}
	}()
}
