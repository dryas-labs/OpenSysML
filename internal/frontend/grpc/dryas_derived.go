// DRYAS experimental API overlay for OpenSysML v0.9.2. Not an upstream API.
package grpc

import (
	"connectrpc.com/connect"
	"context"
	"encoding/json"
	pb "github.com/Open-MBEE/OpenSysML/api/proto"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
)

// dryasTarget refuses ambiguous identities rather than copying GetSymbol's first-match behavior.
func (s *Service) dryasTarget(req *pb.GetSymbolRequest) (*CachedModel, *symbols.Symbol, error) {
	cached, ok := s.cache.Get(req.ModelHash)
	if !ok {
		return nil, nil, statusError(connect.CodeNotFound, "model not found")
	}
	syms := lookupNamed(cached.Index, req.SymbolId)
	if len(syms) == 0 {
		return nil, nil, statusError(connect.CodeNotFound, "symbol not found")
	}
	if len(syms) != 1 {
		return nil, nil, statusError(connect.CodeFailedPrecondition, "ambiguous symbol identity")
	}
	return cached, syms[0], nil
}

// DryasDescribeInherited exposes the engine's effective members and their native owner identities.
// All library features are retained. No adapter-defined inheritance or masking rules are used.
func (s *Service) DryasDescribeInherited(ctx context.Context, req *pb.GetSymbolRequest) (*pb.QueryResponse, error) {
	cached, target, err := s.dryasTarget(req)
	if err != nil {
		return nil, err
	}
	sc := cached.SymbolContext()
	defer sc.Lock()()
	members := sc.Semantics.MembersOf(target)
	if !sc.Semantics.MemberSourcesStable(target) {
		return nil, statusError(connect.CodeFailedPrecondition, "native member sources are provisional")
	}
	result := &pb.QueryResponse{}
	for _, member := range members {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !member.IsFeature() {
			continue
		}
		id := sc.Index.GetFQN(member)
		if member.Name == "" || id == "" {
			return nil, statusError(connect.CodeFailedPrecondition, "unnamed feature cannot be represented by this experimental identity contract")
		}
		if member.OwnerScope == nil || member.OwnerScope.Owner() == nil {
			return nil, statusError(connect.CodeFailedPrecondition, "native declaring owner unavailable")
		}
		owner := member.OwnerScope.Owner()
		props := map[string]string{"name": member.Name, "declaringOwner": sc.Index.GetFQN(owner)}
		if !symbols.SameElement(owner, target) {
			props["inheritedFrom"] = props["declaringOwner"]
		}
		var types []string
		for _, typ := range sc.Semantics.FeatureTypeSet(member) {
			types = append(types, sc.Index.GetFQN(typ))
		}
		encoded, err := json.Marshal(types)
		if err != nil {
			return nil, err
		}
		props["types"] = string(encoded)
		if len(types) == 1 {
			props["type"] = types[0]
		}
		result.Elements = append(result.Elements, &pb.QueryResultElement{Id: id, Type: member.Kind.String(), Properties: props})
	}
	return result, nil
}

// DryasFindBySpecialization asks the engine for each project's transitive supertypes.
// It excludes the requested element itself; it includes usages when the engine says they specialize it.
func (s *Service) DryasFindBySpecialization(ctx context.Context, req *pb.GetSymbolRequest) (*pb.QueryResponse, error) {
	cached, target, err := s.dryasTarget(req)
	if err != nil {
		return nil, err
	}
	sc := cached.SymbolContext()
	defer sc.Lock()()
	evaluator := &queryEval{sc: sc, cached: cached}
	candidates, err := evaluator.candidates(cached, nil)
	if err != nil {
		return nil, err
	}
	result := &pb.QueryResponse{}
	emitted := make(map[string]bool)
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if symbols.SameElement(candidate, target) {
			continue
		}
		supers := sc.Semantics.AllSupertypes(candidate)
		if sc.Semantics.SupertypesProvisional(candidate) {
			return nil, statusError(connect.CodeFailedPrecondition, "native supertypes are provisional")
		}
		for _, super := range supers {
			if !symbols.SameElement(super, target) {
				continue
			}
			id := sc.Index.GetFQN(candidate)
			if candidate.Name == "" || id == "" {
				return nil, statusError(connect.CodeFailedPrecondition, "unnamed specialization cannot be represented")
			}
			if emitted[id] {
				return nil, statusError(connect.CodeFailedPrecondition, "ambiguous result identity")
			}
			emitted[id] = true
			result.Elements = append(result.Elements, &pb.QueryResultElement{Id: id, Type: candidate.Kind.String(), Properties: map[string]string{"qualifiedName": id}})
			break
		}
	}
	return result, nil
}
