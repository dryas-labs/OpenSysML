package grpc

import (
	"connectrpc.com/connect"
	"context"
	pb "github.com/Open-MBEE/OpenSysML/api/proto"
	"testing"
)

func TestDryasDerivedNativeQueries(t *testing.T) {
	s := mustNewService(t, 4)
	h := mustParse(t, s, `package D { port def P; part def A { port power : P; } part def B :> A; part def C :> B; part def Other; }`)
	ctx := context.Background()
	inherited, err := s.DryasDescribeInherited(ctx, &pb.GetSymbolRequest{ModelHash: h, SymbolId: "D::C"})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range inherited.Elements {
		if e.Properties["name"] == "power" {
			found = true
			if e.Properties["inheritedFrom"] != "D::A" || e.Properties["type"] != "D::P" {
				t.Fatalf("wrong native facts: %v", e)
			}
		}
	}
	if !found {
		t.Fatal("inherited port missing")
	}
	result, err := s.DryasFindBySpecialization(ctx, &pb.GetSymbolRequest{ModelHash: h, SymbolId: "D::A"})
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, e := range result.Elements {
		seen[e.Id] = true
	}
	if !seen["D::B"] || !seen["D::C"] || seen["D::Other"] || seen["D::A"] {
		t.Fatalf("wrong closure: %v", seen)
	}
	for _, method := range []func(context.Context, *pb.GetSymbolRequest) (*pb.QueryResponse, error){s.DryasDescribeInherited, s.DryasFindBySpecialization} {
		if _, err := method(ctx, &pb.GetSymbolRequest{ModelHash: h, SymbolId: "D::Missing"}); connect.CodeOf(err) != connect.CodeNotFound {
			t.Fatalf("missing name: %v", err)
		}
		if _, err := method(ctx, &pb.GetSymbolRequest{ModelHash: "absent", SymbolId: "D::A"}); connect.CodeOf(err) != connect.CodeNotFound {
			t.Fatalf("missing model: %v", err)
		}
	}
}
