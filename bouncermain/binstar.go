package bouncermain

import (
	"errors"
	"time"
)

type PeerRequest struct {
	Event uint
}

const (
	NULL = iota
	PRIMARY
	SECONDARY
	MASTER
	SLAVE
	CONFIG_ERROR
	RUNTIME_ERROR
	PASS
	REPLY
)

const (
	PEER_NULL = iota
	PEER_PRIMARY
	PEER_SECONDARY
	PEER_MASTER
	PEER_SLAVE
	PEER_TIMEOUT
	CLIENT_REQUEST
)

const (
	NODE_INACTIVE = iota
	NODE_OK
	NODE_WAIT_TIMEOUT
)

var BinstarStates = map[uint]string{
	PRIMARY:       "PRIMARY",
	SECONDARY:     "SECONDARY",
	MASTER:        "MASTER",
	SLAVE:         "SLAVE",
	CONFIG_ERROR:  "CONFIG_ERROR",
	RUNTIME_ERROR: "RUNTIME_ERROR",
	PASS:          "PASS",
	REPLY:         "REPLY",
}

var BinstarEvents = map[uint]string{
	PEER_PRIMARY:   "PEER_PRIMARY",
	PEER_SECONDARY: "PEER_SECONDARY",
	PEER_MASTER:    "PEER_MASTER",
	PEER_SLAVE:     "PEER_SLAVE",
	PEER_TIMEOUT:   "PEER_TIMEOUT",
	CLIENT_REQUEST: "CLIENT_REQUEST",
}

var BinstarTransitions = map[uint]map[uint]uint{
	CLIENT_REQUEST: {
		SLAVE:     PASS,
		MASTER:    REPLY,
		PRIMARY:   PASS,
		SECONDARY: PASS,
	},

	PEER_TIMEOUT: {
		SLAVE:     MASTER,
		MASTER:    PASS,
		PRIMARY:   MASTER,
		SECONDARY: MASTER,
	},

	PEER_SECONDARY: {
		SLAVE:     PASS,
		MASTER:    PASS,
		PRIMARY:   MASTER,
		SECONDARY: CONFIG_ERROR,
	},

	PEER_MASTER: {
		SLAVE:     PASS,
		MASTER:    RUNTIME_ERROR,
		PRIMARY:   SLAVE,
		SECONDARY: SLAVE,
	},

	PEER_PRIMARY: {
		SLAVE:     PASS,
		MASTER:    PASS,
		PRIMARY:   CONFIG_ERROR,
		SECONDARY: PASS,
	},

	PEER_SLAVE: {
		SLAVE:     RUNTIME_ERROR,
		MASTER:    PASS,
		PRIMARY:   MASTER,
		SECONDARY: MASTER,
	},
}

var (
	ErrBinstarConfigError  = errors.New("config error")
	ErrBinstarRuntimeError = errors.New("runtime error")
)

// ffjson: skip
type BinaryStar struct {
	State         uint
	Active        bool
	LastReceived  time.Time
	LastPublished time.Time
	LastPeerState uint
	Timeout       time.Duration
	hbTicker      *time.Ticker
}

func NewBinaryStar(initialState uint, timeout uint) (binstar *BinaryStar) {
	binstar = &BinaryStar{
		LastReceived: time.Now(),
		Timeout:      time.Duration(timeout) * time.Second,
		State:        initialState,
		hbTicker:     time.NewTicker(time.Second),
	}

	logger.Infof("Initialized as %v", BinstarStates[initialState])

	go binstar.monitorHeartbeat()

	return binstar
}

func (binstar *BinaryStar) update(peerState uint) error {
	action := BinstarTransitions[peerState][binstar.State]

	stateName := BinstarStates[binstar.State]
	eventName := BinstarEvents[peerState]

	logger.Infof("I am %s, peer reported %s.", stateName, eventName)

	switch action {
	case MASTER:
		logger.Infof("Ready as MASTER.")
		binstar.State = MASTER
	case SLAVE:
		logger.Infof("Ready as SLAVE.")
		binstar.State = SLAVE
	case CONFIG_ERROR:
		logger.Errorf("Config error. Only one %s instance is allowed.", stateName)
		return ErrBinstarConfigError
	case RUNTIME_ERROR:
		logger.Errorf("Runtime error. Only one %s instance is allowed at any time", stateName)
		return ErrBinstarRuntimeError
	case PASS:
		return nil
	case REPLY:
		return nil
	default:
		logger.Errorf("Invalid state received: %s", eventName)
	}

	binstar.Active = binstar.State == MASTER
	return nil
}

func (binstar *BinaryStar) Receive(peerState uint) error {
	binstar.LastReceived = time.Now()

	if peerState != binstar.LastPeerState {
		err := binstar.update(peerState)
		if err != nil {
			return err
		}
		binstar.LastPeerState = peerState
	}
	return nil
}

func (binstar *BinaryStar) monitorHeartbeat() {

	for _ = range binstar.hbTicker.C {
		if binstar.LastPeerState == PEER_TIMEOUT {
			continue
		}

		if time.Since(binstar.LastReceived) > time.Duration(binstar.Timeout) {
			binstar.Receive(PEER_TIMEOUT)
		}

	}

}
