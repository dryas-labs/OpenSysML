// SPDX-License-Identifier: Apache-2.0
// Added by DRYAS maintainers for the downstream OpenSysML implementation.
package grpc

import (
	"connectrpc.com/connect"
	"context"
	pb "github.com/Open-MBEE/OpenSysML/api/proto"
	"testing"
)

func TestDryasImplicitNativeProvenance(t *testing.T) {
	service := mustNewService(t, 4)
	handle := mustParse(t, service, `package P {
        private import Metaobjects::SemanticMetadata;
        requirement def SafetyRequirement;
        requirement base : SafetyRequirement[*] nonunique;
        metadata def <safety> Tag :> SemanticMetadata { :>> baseType = base meta SysML::Usage; }
        alias secure for Tag;
        #secure requirement tagged;
        requirement body { @Tag; }
        requirement plain;
        part def Component;
    }`)
	query := func(name string) *pb.QueryResponse {
		result, err := service.DryasDescribeProvenance(context.Background(), &pb.GetSymbolRequest{ModelHash: handle, SymbolId: name})
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	found := false
	for _, e := range query("P::Component").Elements {
		if e.Type == "DryasGeneralization" && e.Properties["target"] == "Parts::Part" {
			found = e.Properties["implicit"] == "true" && e.Properties["kindImplicit"] == "true"
		}
	}
	if !found {
		t.Fatal("native kind-implicit Part relation missing")
	}
	for _, name := range []string{"P::tagged", "P::body"} {
		relation, annotation := false, false
		for _, e := range query(name).Elements {
			if e.Type == "DryasGeneralization" && e.Properties["target"] == "P::base" {
				relation = e.Properties["semanticMetadataImplicit"] == "true"
			}
			if e.Type == "DryasMetadataAnnotation" {
				annotation = e.Properties["type"] == "P::Tag" && e.Properties["semantic"] == "true"
				if name == "P::tagged" && e.Properties["keyword"] != "secure" {
					t.Fatal("alias spelling lost")
				}
				if name == "P::body" && e.Properties["keyword"] != "" {
					t.Fatal("invented prefix keyword for body annotation")
				}
			}
		}
		if !relation || !annotation {
			t.Fatalf("missing native facts for %s", name)
		}
	}
	for _, e := range query("P::plain").Elements {
		if e.Type == "DryasMetadataAnnotation" || e.Properties["target"] == "P::base" {
			t.Fatal("false semantic metadata match")
		}
	}
	broken := mustParse(t, service, `package Broken { #Unknown requirement r; }`)
	_, err := service.DryasDescribeProvenance(context.Background(), &pb.GetSymbolRequest{ModelHash: broken, SymbolId: "Broken::r"})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("provisional relation must be unanswered: %v", err)
	}
}
