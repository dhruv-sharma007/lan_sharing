package transfer

// MessageType identifies the kind of transfer control message.
type MessageType string

const (
	TypeTransferOffer    MessageType = "transfer_offer"
	TypeTransferAccept   MessageType = "transfer_accept"
	TypeTransferComplete MessageType = "transfer_complete"
	TypeTransferError    MessageType = "transfer_error"
)

// ControlMessage is a wrapper for all control messages.
type ControlMessage struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

// TransferOffer is sent by the sender to propose a file transfer.
type TransferOffer struct {
	ProtocolVersion int    `json:"protocol_version"`
	TransferID      string `json:"transfer_id"`
	SenderNodeID    uint64 `json:"sender_node_id"`
	Filename        string `json:"filename"`
	RelativePath    string `json:"relative_path"`
	Size            int64  `json:"size"`
	ModTime         int64  `json:"modified_time"`
}

// TransferAccept is sent by the receiver to accept a file transfer.
type TransferAccept struct {
	TransferID string `json:"transfer_id"`
	Token      string `json:"token"`
}

// TransferComplete is sent by the receiver when the file is fully verified and saved.
type TransferComplete struct {
	TransferID string `json:"transfer_id"`
}

// TransferError is sent by either party if a transfer fails.
type TransferError struct {
	TransferID string `json:"transfer_id"`
	Message    string `json:"message"`
}
