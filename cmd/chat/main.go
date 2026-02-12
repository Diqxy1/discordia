package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"discordia/internal/domain"
	"discordia/internal/p2p"
	"discordia/internal/repository"
	"discordia/internal/usecase"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
)

func main() {
	ctx := context.Background()
	repo := repository.NewBadgerRepo("./data")

	// Carregar ou gerar identidade
	privKeyData, err := repo.GetIdentity()
	var privKey crypto.PrivKey
	if err != nil {
		privKey, _, _ = crypto.GenerateKeyPair(crypto.Ed25519, -1)
		rawKey, _ := crypto.MarshalPrivateKey(privKey)
		repo.SaveIdentity(rawKey)
	} else {
		privKey, _ = crypto.UnmarshalPrivateKey(privKeyData)
	}

	h, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"), // Porta aleatória local
	)
	if err != nil {
		panic(err)
	}

	ps, _ := pubsub.NewGossipSub(ctx, h)
	network := p2p.NewLibp2pService(ctx, h, ps, "sala-1")
	chat := usecase.NewChatUseCase(repo, network)
	mdns.NewMdnsService(h, "sala-1", &discoveryNotifee{h: h}).Start()

	fmt.Printf("Seu ID: %s\nDigite seu nome: ", h.ID().String()[:6])
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	username := scanner.Text()

	chat.StartReceiving(func(m domain.Message) {
		fmt.Printf("\n\x1b[34m[%s]:\x1b[0m %s\n> ", m.Sender, m.Content)
	})

	fmt.Println("\nDigite /connect <endereço> para add o Node:")
	for scanner.Scan() {
		input := scanner.Text()
		if strings.HasPrefix(input, "/connect ") {
			addrStr := strings.TrimPrefix(input, "/connect ")
			maddr, _ := multiaddr.NewMultiaddr(addrStr)
			info, _ := peer.AddrInfoFromP2pAddr(maddr)
			if err := h.Connect(ctx, *info); err != nil {
				fmt.Println("Erro:", err)
			} else {
				fmt.Println("Conectado ao Node!")
			}
			continue
		}
		chat.SendMessage(username, input)
		fmt.Print("> ")
	}
}

type discoveryNotifee struct{ h host.Host }

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) { n.h.Connect(context.Background(), pi) }
