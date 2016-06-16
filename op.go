package mongotape

import (
	"fmt"
	"io"
	"time"

	"github.com/10gen/llmgo"
)

// ErrNotMsg is returned if a provided buffer is too small to contain a Mongo message
var ErrNotMsg = fmt.Errorf("buffer is too small to be a Mongo message")

type OpMetadata struct {
	// Op represents the actual operation being performed accounting for write commands, so
	// this may be "insert" or "update" even when the wire protocol message was OP_QUERY.
	Op string

	// Namespace against which the operation executes. If not applicable, will be blank.
	Ns string

	// Command name is the name of the command, when Op is "command" (otherwise will be blank.)
	// For example, this might be "getLastError" or "serverStatus".
	Command string

	// Data contains the payload of the operation.
	// For queries: the query selector, limit and sort, etc.
	// For inserts: the document(s) to be inserted.
	// For updates: the query selector, modifiers, and upsert/multi flags.
	// For removes: the query selector for removes.
	// For commands: the full set of parameters for the command.
	// For killcursors: the list of cursorId's to be killed.
	// For getmores: the cursorId for the getmore batch.
	Data interface{}
}

// Op is a Mongo operation
type Op interface {
	OpCode() OpCode
	FromReader(io.Reader) error
	Execute(*mgo.Session) (replyContainer, error)
	Equals(Op) bool
	Meta() OpMetadata
	Abbreviated(int) string
}

type replyContainer struct {
	*CommandReplyOp
	*ReplyOp
	Latency time.Duration
}

// ErrUnknownOpcode is an error that represents an unrecognized opcode.
type ErrUnknownOpcode int

func (e ErrUnknownOpcode) Error() string {
	return fmt.Sprintf("Unknown opcode %d", e)
}

//IsDriverOp checks if an operation is one of the types generated by the driver
//such as 'ismaster', or 'getnonce'. It takes an Op that has already been
//unmarshalled using its 'FromReader' method and checks if it is a command matching
//the ones the driver generates.
func IsDriverOp(op Op) bool {
	query, ok := op.(*QueryOp)

	if !ok {
		return false
	}

	opType, commandType := extractOpType(query.QueryOp.Query)
	if opType != "command" {
		return false
	}
	switch commandType {
	case "isMaster", "ismaster":
		return true
	case "getnonce":
		return true
	case "ping":
		return true
	case "saslStart":
		return true
	case "saslContinue":
		return true
	default:
		return false
	}
}
