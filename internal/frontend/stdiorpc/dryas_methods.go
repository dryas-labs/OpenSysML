// SPDX-License-Identifier: Apache-2.0
// Added by DRYAS maintainers for the downstream OpenSysML implementation.
// DRYAS experimental methods use existing protobuf envelopes over stdio only.
package stdiorpc

import (
	"context"
	pb "github.com/Open-MBEE/OpenSysML/api/proto"
)

type dryasDerived interface {
	DryasDescribeInherited(context.Context, *pb.GetSymbolRequest) (*pb.QueryResponse, error)
	DryasFindBySpecialization(context.Context, *pb.GetSymbolRequest) (*pb.QueryResponse, error)
}

func registerDryas(s *Server) {
	impl, ok := s.impl.(dryasDerived)
	if !ok {
		return
	}
	methods := map[string]func(context.Context, *pb.GetSymbolRequest) (*pb.QueryResponse, error){
		"DryasDescribeInherited":    impl.DryasDescribeInherited,
		"DryasFindBySpecialization": impl.DryasFindBySpecialization,
	}
	for name, call := range methods {
		s.methods[name] = func(ctx context.Context, dec func(any) error) (any, error) {
			req := &pb.GetSymbolRequest{}
			if err := dec(req); err != nil {
				return nil, err
			}
			return call(ctx, req)
		}
	}
}
