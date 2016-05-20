package runtime

import (
	"net/http"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

const metadataHeaderPrefix = "Grpc-Metadata-"
const metadataTrailerPrefix = "Grpc-Trailer-"
const corsHeaderPrefix = "access-control-"
const csrfTokenHeader = "X-Phabricator-Csrf"
/*
AnnotateContext adds context information such as metadata from the request.

If there are no metadata headers in the request, then the context returned
will be the same context.
*/
func AnnotateContext(ctx context.Context, req *http.Request) context.Context {
	var pairs []string
	for key, vals := range req.Header {
		for _, val := range vals {
			if key == "Authorization" {
				pairs = append(pairs, key, val)
				continue
			}
			if strings.EqualFold(key, csrfTokenHeader) {
				pairs = append(pairs, key, val)
				continue
			}
			if strings.EqualFold(key, corsHeaderPrefix) {
				pairs = append(pairs, key, val)
				continue
			}
			if strings.HasPrefix(key, metadataHeaderPrefix) {
				pairs = append(pairs, key[len(metadataHeaderPrefix):], val)
			}
		}
	}

	// append other cookies

	// adding extra headers to metadata
	pairs = append(pairs,
		"http-request-method", req.Method,
		"http-request-endpoint", req.RequestURI,
		"http-request-host", req.Host,
		"http-userAgent", req.UserAgent(),
	)

	for _, cookie := range req.Cookies() {
		pairs = append(pairs, "http-request-cookie-" + cookie.Name, cookie.Value)
	}
	if len(pairs) == 0 {
		return ctx
	}
	return metadata.NewContext(ctx, metadata.Pairs(pairs...))
}

// ServerMetadata consists of metadata sent from gRPC server.
type ServerMetadata struct {
	HeaderMD  metadata.MD
	TrailerMD metadata.MD
}

type serverMetadataKey struct{}

// NewServerMetadataContext creates a new context with ServerMetadata
func NewServerMetadataContext(ctx context.Context, md ServerMetadata) context.Context {
	return context.WithValue(ctx, serverMetadataKey{}, md)
}

// ServerMetadataFromContext returns the ServerMetadata in ctx
func ServerMetadataFromContext(ctx context.Context) (md ServerMetadata, ok bool) {
	md, ok = ctx.Value(serverMetadataKey{}).(ServerMetadata)
	return
}
