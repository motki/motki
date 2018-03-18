package client

import (
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/motki/core/eveapi"
	"github.com/motki/core/proto"
)

// StructureClient retrieves information about player-owned Citadels in EVE.
type StructureClient struct {
	*bootstrap
}

// GetStructure returns public information about the given structure.
func (c *StructureClient) GetStructure(structureID int) (*eveapi.Structure, error) {
	if c.token == "" {
		return nil, ErrNotAuthenticated
	}
	conn, err := grpc.Dial(c.serverAddr, c.dialOpts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	service := proto.NewInfoServiceClient(conn)
	res, err := service.GetStructure(
		context.Background(),
		&proto.GetStructureRequest{Token: &proto.Token{Identifier: c.token}, StructureId: int64(structureID)})
	if err != nil {
		return nil, err
	}
	if res.Result.Status == proto.Status_FAILURE {
		return nil, errors.New(res.Result.Description)
	}
	return proto.ProtoToStructure(res.Structure), nil
}

// GetCorpStructures returns detailed information about corporation structures.
//
// This method requires that the user's corporation has opted-in to data collection.
func (c *StructureClient) GetCorpStructures() ([]*eveapi.CorporationStructure, error) {
	if c.token == "" {
		return nil, ErrNotAuthenticated
	}
	conn, err := grpc.Dial(c.serverAddr, c.dialOpts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	service := proto.NewCorporationServiceClient(conn)
	res, err := service.GetCorpStructures(
		context.Background(),
		&proto.GetCorpStructuresRequest{Token: &proto.Token{Identifier: c.token}})
	if err != nil {
		return nil, err
	}
	if res.Result.Status == proto.Status_FAILURE {
		return nil, errors.New(res.Result.Description)
	}
	var strucs []*eveapi.CorporationStructure
	for _, k := range res.Structures {
		strucs = append(strucs, proto.ProtoToCorpStructure(k))
	}
	return strucs, nil
}
