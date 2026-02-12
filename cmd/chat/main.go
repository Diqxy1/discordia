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
	privKeyData, err := repo.GetIdentity()

	var privKey crypto.PrivKey

	if err != nil {
		privKey, _, err = crypto.GenerateKeyPair(crypto.Ed25519, -1)
		if err != nil {
			panic(err)
		}
		rawKey, _ := crypto.MarshalPrivateKey(privKey)
		repo.SaveIdentity(rawKey)
		fmt.Println("Nova identidade gerada e salva.")
	} else {
		privKey, err = crypto.UnmarshalPrivateKey(privKeyData)
		if err != nil {
			panic(err)
		}
		fmt.Println("Identidade carregada do banco.")
	}

	h, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/0",
			"/ip4/0.0.0.0/udp/0/quic-v1",
		),
		libp2p.NATPortMap(),
		libp2p.EnableHolePunching(),
		libp2p.EnableRelayService(),
		libp2p.EnableAutoRelayWithPeerSource(func(ctx context.Context, num int) <-chan peer.AddrInfo {
			return make(<-chan peer.AddrInfo)
		}),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Seu endereço completo:\n")
	for _, addr := range h.Addrs() {
		fmt.Printf("%s/p2p/%s\n", addr.String(), h.ID().String())
	}

	ps, _ := pubsub.NewGossipSub(context.Background(), h)

	network := p2p.NewLibp2pService(ctx, h, ps, "sala-1")

	chat := usecase.NewChatUseCase(repo, network)

	mdns.NewMdnsService(h, "sala-1", &discoveryNotifee{h: h}).Start()

	isCloud := os.Getenv("FLY_APP_NAME") != ""

	if isCloud {
		fmt.Println("Rodando em modo Node (Nuvem).")
		select {}
	}

	fmt.Printf("Seu ID: %s | Digite seu nome: ", h.ID().String()[:6])
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	username := scanner.Text()

	history, _ := repo.GetAll()
	for _, m := range history {
		fmt.Printf("[%s] %s: %s (Histórico)\n", m.Timestamp.Format("15:04"), m.Sender, m.Content)
	}

	chat.StartReceiving(func(m domain.Message) {
		fmt.Printf("\n\x1b[34m[%s]:\x1b[0m %s\n> ", m.Sender, m.Content)
	})

	fmt.Println("\nConectado! Digite sua mensagem ou /connect <endereço> para add um amigo:")
	for scanner.Scan() {
		input := scanner.Text()
		if input == "" {
			continue
		}

		// COMANDO PARA CONECTAR MANUALMENTE
		if strings.HasPrefix(input, "/connect ") {
			addrStr := strings.TrimPrefix(input, "/connect ")

			// Transforma string em endereço
			maddr, err := multiaddr.NewMultiaddr(addrStr)
			if err != nil {
				fmt.Println("Endereço inválido:", err)
				continue
			}

			// Extrai o ID do peer e tenta conectar
			info, err := peer.AddrInfoFromP2pAddr(maddr)
			if err != nil {
				fmt.Println("Erro no endereço:", err)
				continue
			}

			if err := h.Connect(ctx, *info); err != nil {
				fmt.Println("Falha ao conectar:", err)
			} else {
				fmt.Println("Conectado com sucesso!")
			}
			fmt.Print("> ")
			continue
		}

		// Fluxo normal de mensagem
		chat.SendMessage(username, input)
		fmt.Print("> ")
	}
}

type discoveryNotifee struct{ h host.Host }

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) { n.h.Connect(context.Background(), pi) }
