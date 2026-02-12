package main

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	ma "github.com/multiformats/go-multiaddr"
)

func main() {
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Node P2P Online"))
		})
		http.ListenAndServe(":8080", nil)
	}()

	var privKey crypto.PrivKey

	keyString := os.Getenv("P2P_KEY")

	if keyString == "" {
		fmt.Println("Variável P2P_KEY não encontrada. Gerando nova chave")
		var err error
		privKey, _, err = crypto.GenerateKeyPair(crypto.Ed25519, -1)
		if err != nil {
			panic(err)
		}

		rawKey, _ := crypto.MarshalPrivateKey(privKey)
		fmt.Printf("\n--- CHAVE RAILWAY ---\n%x\n---\n\n", rawKey)
	} else {
		fmt.Println("Usando chave fixa da variável de ambiente.")
		rawKey, err := hex.DecodeString(keyString)
		if err != nil {
			panic("Erro ao decodificar P2P_KEY: " + err.Error())
		}
		privKey, err = crypto.UnmarshalPrivateKey(rawKey)
		if err != nil {
			panic("Erro ao carregar P2P_KEY: " + err.Error())
		}
	}

	// Endereço do Proxy TCP do Railway
	externalAddr, _ := ma.NewMultiaddr("/dns4/yamanote.proxy.rlwy.net/tcp/50519")

	h, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/4001"),
		libp2p.AddrsFactory(func(addrs []ma.Multiaddr) []ma.Multiaddr {
			return []ma.Multiaddr{externalAddr}
		}),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("=== NODE INICIADO ===")
	fmt.Printf("ID do Nó: %s\n", h.ID().String())
	fmt.Printf("Endereço: /dns4/yamanote.proxy.rlwy.net/tcp/50519/p2p/%s\n", h.ID().String())

	// Mantém o container vivo
	select {}
}
