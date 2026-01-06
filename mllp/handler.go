package mllp

import (
	"context"

	"github.com/dshills/golevel7/hl7"
)

// Handler defines the interface for handling incoming HL7 messages.
//
// Implementations process the incoming message and return a response message
// (typically an ACK) or an error. The context can be used for cancellation
// and deadline propagation.
//
// Example implementation:
//
//	type MyHandler struct {
//	    db *sql.DB
//	}
//
//	func (h *MyHandler) HandleMessage(ctx context.Context, msg hl7.Message) (hl7.Message, error) {
//	    msgType := msg.Type()
//	    switch msgType {
//	    case "ADT^A01":
//	        return h.handleAdmission(ctx, msg)
//	    case "ORU^R01":
//	        return h.handleLabResult(ctx, msg)
//	    default:
//	        return createNAK(msg, "Unsupported message type")
//	    }
//	}
type Handler interface {
	// HandleMessage processes an incoming HL7 message and returns a response.
	// The response is typically an ACK (acknowledgment) message.
	//
	// If the handler returns an error, the server will send a NAK (negative
	// acknowledgment) to the client if possible.
	//
	// The context will be canceled if the client disconnects or the server
	// is shutting down.
	HandleMessage(ctx context.Context, msg hl7.Message) (hl7.Message, error)
}

// HandlerFunc is an adapter type that allows ordinary functions to be used
// as message handlers.
//
// If f is a function with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
//
// Example usage:
//
//	handler := mllp.HandlerFunc(func(ctx context.Context, msg hl7.Message) (hl7.Message, error) {
//	    // Process message and return ACK
//	    return createACK(msg), nil
//	})
//
//	server := mllp.NewServer(mllp.WithHandler(handler))
type HandlerFunc func(ctx context.Context, msg hl7.Message) (hl7.Message, error)

// HandleMessage calls f(ctx, msg).
// This implements the Handler interface for HandlerFunc.
func (f HandlerFunc) HandleMessage(ctx context.Context, msg hl7.Message) (hl7.Message, error) {
	return f(ctx, msg)
}

// Ensure HandlerFunc implements Handler at compile time.
var _ Handler = HandlerFunc(nil)

// MiddlewareFunc defines a function that wraps a Handler to add behavior.
// Middleware can be used for logging, authentication, metrics, etc.
//
// Example middleware:
//
//	func LoggingMiddleware(logger *log.Logger) MiddlewareFunc {
//	    return func(next Handler) Handler {
//	        return HandlerFunc(func(ctx context.Context, msg hl7.Message) (hl7.Message, error) {
//	            start := time.Now()
//	            logger.Printf("Received message type: %s", msg.Type())
//	            resp, err := next.HandleMessage(ctx, msg)
//	            logger.Printf("Processed in %v", time.Since(start))
//	            return resp, err
//	        })
//	    }
//	}
type MiddlewareFunc func(Handler) Handler

// Chain applies middleware functions to a handler in the order provided.
// The first middleware in the slice will be the outermost wrapper.
//
// Example:
//
//	handler := mllp.Chain(
//	    baseHandler,
//	    LoggingMiddleware(logger),
//	    MetricsMiddleware(metrics),
//	    AuthMiddleware(auth),
//	)
func Chain(h Handler, middleware ...MiddlewareFunc) Handler {
	// Apply middleware in reverse order so the first middleware
	// in the slice becomes the outermost wrapper
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}
