package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"syscall"

	morphclient "git.frostfs.info/TrueCloudLab/frostfs-node/pkg/morph/client"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/actor"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/gas"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/nep17"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm/opcode"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	cfgRPCEndpoint   = "rpc_endpoint"
	cfgRPCEndpointWS = "rpc_endpoint_ws"
	cfgWallet        = "wallet"
	cfgClientWallet  = "client"
	cfgPassword      = "password"
	cfgContract      = "contract"
	cfgListenAddress = "listen_address"
)

type Server struct {
	serverAcc    *wallet.Account
	serverAct    *actor.Actor
	serverKey    *keys.PublicKey
	gasAct       *nep17.Token
	contractHash util.Uint160
	log          *zap.Logger
	serverCli    *rpcclient.Client
}

type User struct {
	userAcc *wallet.Account
	userCli *rpcclient.Client
}

type Opcode struct {
	code  opcode.Opcode
	param []byte
}

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	if len(os.Args) != 2 {
		Exception(fmt.Errorf("invalid args: %v", os.Args))
	}

	viper.GetViper().SetConfigType("yml")

	f, err := os.Open(os.Args[1])
	Exception(err)
	Exception(viper.GetViper().ReadConfig(f))
	Exception(f.Close())

	server, err := NewServer(ctx)
	Exception(err)
	Exception(server.Listen(ctx))
}

func NewServer(ctx context.Context) (*Server, error) {
	newClient, err := rpcclient.New(ctx, viper.GetString(cfgRPCEndpoint), rpcclient.Options{})
	if err != nil {
		return nil, err
	}

	newWallet, err := wallet.NewWalletFromFile(viper.GetString(cfgWallet))
	if err != nil {
		return nil, err
	}

	account := newWallet.GetAccount(newWallet.GetChangeAddress())
	if err = account.Decrypt(viper.GetString(cfgPassword), newWallet.Scrypt); err != nil {
		return nil, err
	}

	backendKey := account.PrivateKey().PublicKey()

	actor, err := actor.NewSimple(newClient, account)
	if err != nil {
		return nil, err
	}

	contractHash, err := util.Uint160DecodeStringLE(viper.GetString(cfgContract))
	if err != nil {
		return nil, err
	}

	neoClient, err := morphclient.New(ctx, account.PrivateKey(),
		morphclient.WithEndpoints(morphclient.Endpoint{Address: viper.GetString(cfgRPCEndpointWS), Priority: 1}),
	)
	if err != nil {
		return nil, fmt.Errorf("new morph client: %w", err)
	}

	if err = neoClient.EnableNotarySupport(); err != nil {
		return nil, err
	}

	log, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	return &Server{
		serverAcc:    account,
		serverAct:    actor,
		serverKey:    backendKey,
		serverCli:    newClient,
		contractHash: contractHash,
		gasAct:       nep17.New(actor, gas.Hash),
		log:          log,
	}, nil
}

func (server *Server) Listen(ctx context.Context) error {
	http.DefaultServeMux.HandleFunc("/sendMessage", func(w http.ResponseWriter, r *http.Request) {
		server.log.Info("REQUEST SEND Message")
		if r.Method != http.MethodPost {
			server.log.Error("invalid method", zap.String("method", r.Method))
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var requestBody struct {
			Key         string `json:"key"`
			MessageHash string `json:"messageHash"`
		}

		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			server.log.Error("failed to parse request body", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		server.log.Info("Decoded JSON inputs", zap.String("key", requestBody.Key), zap.String("messageHash", requestBody.MessageHash))

		key := requestBody.Key

		messageHash := requestBody.MessageHash

		contractID, _ := util.Uint160DecodeStringLE(viper.GetString(cfgContract))

		invresult, err := server.serverAct.Wait(server.serverAct.SendCall(contractID, "sendMessage", key, messageHash))

		server.log.Info("Contract call result", zap.String("state", invresult.VMState.String()), zap.Any("stack", invresult.Stack))
		if err != nil {
			server.log.Error("CONTRACT call error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		server.log.Info("Contract stack", zap.Any("stack", invresult.Stack))

		if invresult.VMState.String() != "HALT" {
			server.log.Error("CONTRACT execution failed", zap.String("state", invresult.VMState.String()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(invresult.Stack) == 0 {
			server.log.Error("CONTRACT stack is empty")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		stackResult := invresult.Stack[0]

		success, err := stackResult.TryBool()

		if len(invresult.Stack) > 0 {
			server.log.Info("Stack result", zap.Any("stackResult", invresult.Stack[0]))
		}
		if err != nil {
			server.log.Error("FAILED to parse result as bool", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Print(success)

		if success {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"status": "success"}`)); err != nil {
				server.log.Error("FAILED to write response", zap.Error(err))
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			if _, err := w.Write([]byte(`{"status": "invalid send or found dublicate"}`)); err != nil {
				server.log.Error("FAILED to write response", zap.Error(err))
			}
		}
	})

	http.DefaultServeMux.HandleFunc("/checkMessage", func(w http.ResponseWriter, r *http.Request) {
		server.log.Info("REQUEST CHECK Message")

		if r.Method != http.MethodPost {
			server.log.Error("invalid method", zap.String("method", r.Method))
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var requestBody struct {
			Key         string `json:"key"`
			MessageHash string `json:"messageHash"`
		}

		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			server.log.Error("failed to parse request body", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		server.log.Info("Decoded JSON inputs", zap.String("key", requestBody.Key), zap.String("messageHash", requestBody.MessageHash))

		key := requestBody.Key

		messageHash := requestBody.MessageHash

		contractID, _ := util.Uint160DecodeStringLE(viper.GetString(cfgContract))

		invresult, err := server.serverAct.Wait(server.serverAct.SendCall(contractID, "findAndCheckMessage", key, messageHash))
		server.log.Info("Contract call result", zap.String("state", invresult.VMState.String()), zap.Any("stack", invresult.Stack))
		if err != nil {
			server.log.Error("CONTRACT call error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		server.log.Info("Contract stack", zap.Any("stack", invresult.Stack))

		if invresult.VMState.String() != "HALT" {
			server.log.Error("CONTRACT execution failed", zap.String("state", invresult.VMState.String()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(invresult.Stack) == 0 {
			server.log.Error("CONTRACT stack is empty")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		stackResult := invresult.Stack[0]
		success, err := stackResult.TryBool()
		if len(invresult.Stack) > 0 {
			server.log.Info("Stack result", zap.Any("stackResult", invresult.Stack[0]))
		}
		if err != nil {
			server.log.Error("FAILED to parse result as bool", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Print(success)

		if success {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"status": "success"}`)); err != nil {
				server.log.Error("FAILED to write response", zap.Error(err))
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			if _, err := w.Write([]byte(`{"status": "message not found or invalid"}`)); err != nil {
				server.log.Error("FAILED to write response", zap.Error(err))
			}
		}
	})

	http.DefaultServeMux.HandleFunc("/balance", func(w http.ResponseWriter, r *http.Request) {
		server.log.Info("REQUEST BALANCE")

		res, err := server.gasAct.BalanceOf(server.serverAcc.ScriptHash())
		if err != nil {
			server.log.Error("ERROR balance", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err = w.Write([]byte(strconv.FormatInt(res.Int64(), 10))); err != nil {
			server.log.Error("ERROR write response", zap.Error(err))
		}
	})

	return http.ListenAndServe(viper.GetString(cfgListenAddress), nil)
}

func Exception(err error) {
	if err == nil {
		return
	}

	debug.PrintStack()

	fmt.Fprintf(os.Stderr, "Error: %v\n", err)

	os.Exit(1)
}
