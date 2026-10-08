// SPDX-License-Identifier: Apache-2.0
// Added by DRYAS maintainers for the downstream OpenSysML implementation.
package stdiorpc

import (
	"context"
	pb "github.com/Open-MBEE/OpenSysML/api/proto"
)

func registerDryasImplicit(s *Server) {
	impl, ok := s.impl.(interface {
		DryasDescribeProvenance(context.Context, *pb.GetSymbolRequest) (*pb.QueryResponse, error)
	})
	if !ok {
		return
	}
	s.methods["DryasDescribeProvenance"] = func(ctx context.Context, dec func(any) error) (any, error) {
		req := &pb.GetSymbolRequest{}
		if err := dec(req); err != nil {
			return nil, err
		}
		return impl.DryasDescribeProvenance(ctx, req)
	}
}
