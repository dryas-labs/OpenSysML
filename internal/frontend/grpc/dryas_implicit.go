// SPDX-License-Identifier: Apache-2.0
// Added by DRYAS maintainers for the downstream OpenSysML implementation.
// DRYAS experimental provenance export; all relations are computed by OpenSysML.
package grpc

import (
	"connectrpc.com/connect"
	"context"
	pb "github.com/Open-MBEE/OpenSysML/api/proto"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/semantics"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
	"strconv"
)

func (s *Service) DryasDescribeProvenance(ctx context.Context, req *pb.GetSymbolRequest) (*pb.QueryResponse, error) {
	cached, target, err := s.dryasTarget(req)
	if err != nil {
		return nil, err
	}
	sc := cached.SymbolContext()
	defer sc.Lock()()
	direct := sc.Semantics.DirectSupertypes(target)
	kindImplicit := sc.Semantics.ImplicitGenerals(target)
	semanticImplicit, complete := sc.Semantics.DryasSemanticMetadataBases(target)
	if !complete || sc.Semantics.SupertypesProvisional(target) {
		return nil, statusError(connect.CodeFailedPrecondition, "native implicit relations are provisional")
	}
	contains := func(list []*symbols.Symbol, value *symbols.Symbol) bool {
		for _, candidate := range list {
			if symbols.SameElement(candidate, value) {
				return true
			}
		}
		return false
	}
	result := &pb.QueryResponse{}
	declared := sc.specializationsOf(target)
	for _, general := range direct {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		id := sc.Index.GetFQN(general)
		if general.Name == "" || id == "" {
			return nil, statusError(connect.CodeFailedPrecondition, "unnamed native generalization")
		}
		props := map[string]string{"target": id, "kindImplicit": strconv.FormatBool(contains(kindImplicit, general)), "semanticMetadataImplicit": strconv.FormatBool(contains(semanticImplicit, general))}
		if contains(kindImplicit, general) || contains(semanticImplicit, general) {
			props["implicit"] = "true"
		}
		for _, edge := range declared {
			if edge.TargetId == id {
				props["declaredKind"] = edge.Kind
				if _, exists := props["implicit"]; !exists {
					props["implicit"] = "false"
				}
			}
		}
		result.Elements = append(result.Elements, &pb.QueryResultElement{Id: id, Type: "DryasGeneralization", Properties: props})
	}
	// Match node identity to the engine's own prefix/body classification. The keyword
	// is the spelling used at this annotation, never the metadata definition's short name.
	written := semantics.MetadataAnnotationsWritten(target.Decl)
	ordinals := make(map[string]int)
	for _, annotation := range sc.Semantics.ElementMetadataOf(target) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		typ := sc.Index.GetFQN(annotation.Type)
		if typ == "" {
			return nil, statusError(connect.CodeFailedPrecondition, "unnamed annotation type")
		}
		ordinals[typ]++
		props := map[string]string{"type": typ, "semantic": strconv.FormatBool(sc.Semantics.DryasIsSemanticMetadata(annotation.Type)),
			"typeOrdinal": strconv.Itoa(ordinals[typ]), "document": annotation.Doc, "offset": strconv.Itoa(annotation.Node.Span().Offset), "prefix": "false", "about": strconv.FormatBool(annotation.About)}
		for _, source := range written {
			if source.Node == annotation.Node && source.Prefix {
				props["prefix"] = "true"
				if source.Node.Type != nil {
					props["keyword"] = source.Node.Type.Text()
				}
			}
		}
		result.Elements = append(result.Elements, &pb.QueryResultElement{Id: typ + "#" + strconv.Itoa(ordinals[typ]), Type: "DryasMetadataAnnotation", Properties: props})
	}
	return result, nil
}
