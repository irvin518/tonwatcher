package toninterface

import (
	"context"
	"log"

	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
)

var api ton.APIClientWrapped

func Init(url string, env string) error {
	log.Printf("begin to new connect")
	client := liteclient.NewConnectionPool()
	log.Printf("new connect end")
	ctx := context.Background()
	log.Printf("begin to get config %s", url)
	cfg, err := liteclient.GetConfigFromUrl(ctx, url)
	if err != nil {
		log.Printf("get config error: %s", err.Error())
		return err
	}

	log.Printf("get config finish %s", err)
	err = client.AddConnectionsFromConfig(ctx, cfg)
	if err != nil {
		log.Printf("connection error: %s", err.Error())
		return err
	}
	log.Printf("AddConnectionsFromConfig")
	policy := ton.ProofCheckPolicySecure
	if env == "test" {
		policy = ton.ProofCheckPolicyFast
	}
	api = ton.NewAPIClient(client, policy).WithRetry()
	api.SetTrustedBlockFromConfig(cfg)
	log.Printf("api init finish")
	return err
}

func Api() ton.APIClientWrapped {
	return api
}
