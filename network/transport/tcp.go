/*
 *    Copyright 2018 Insolar
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package transport

import (
	"context"
	"io"
	"net"

	"github.com/pkg/errors"

	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/log"
	"github.com/insolar/insolar/network/transport/pool"
	"github.com/insolar/insolar/network/transport/relay"
	"github.com/insolar/insolar/network/utils"
)

type tcpTransport struct {
	baseTransport

	pool     pool.ConnectionPool
	listener net.Listener
	addr     string
}

func newTCPTransport(addr string, proxy relay.Proxy, publicAddress string) (*tcpTransport, error) {
	transport := &tcpTransport{
		baseTransport: newBaseTransport(proxy, publicAddress),
		addr:          addr,
		pool:          pool.NewConnectionPool(&tcpConnectionFactory{}),
	}

	transport.sendFunc = transport.send

	return transport, nil
}

func (t *tcpTransport) send(address string, data []byte) error {
	ctx := context.Background()
	logger := inslogger.FromContext(ctx)

	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return errors.Wrap(err, "[ send ] Failed to resolve net address")
	}

	conn, err := t.pool.GetConnection(ctx, addr)
	if err != nil {
		return errors.Wrap(err, "[ send ] Failed to get connection")
	}

	logger.Debug("[ send ] len = ", len(data))

	_, err = conn.Write(data)

	if err != nil {
		// All this to check is error EPIPE
		// if netErr, ok := err.(*net.OpError); ok {
		// 	switch realNetErr := netErr.Err.(type) {
		// 	case *os.SyscallError:
		// 		if realNetErr.Err == syscall.EPIPE {
		t.pool.CloseConnection(ctx, addr)
		conn, err = t.pool.GetConnection(ctx, addr)
		if err != nil {
			return errors.Wrap(err, "[ send ] Failed to get connection")
		}
		_, err = conn.Write(data)
		// 		}
		// 	}
		// }
	}

	return errors.Wrap(err, "[ send ] Failed to write data")
}

func (t *tcpTransport) prepareListen() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.disconnectStarted = make(chan bool, 1)
	t.disconnectFinished = make(chan bool, 1)
	listener, err := net.Listen("tcp", t.addr)
	if err != nil {
		return err
	}

	t.listener = listener

	return nil
}

// Start starts networking.
func (t *tcpTransport) Listen(ctx context.Context, started chan struct{}) error {
	logger := inslogger.FromContext(ctx)
	logger.Info("[ Listen ] Start TCP transport")

	if err := t.prepareListen(); err != nil {
		logger.Info("[ Listen ] Failed to prepare TCP transport")
		return err
	}

	started <- struct{}{}
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			<-t.disconnectFinished
			logger.Error("[ Listen ] Failed to accept connection: ", err.Error())
			return errors.Wrap(err, "[ Listen ] Failed to accept connection")
		}

		logger.Debugf("[ Listen ] Accepted new connection from %s", conn.RemoteAddr())

		go t.handleAcceptedConnection(conn)
	}
}

// Stop stops networking.
func (t *tcpTransport) Stop() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	log.Info("[ Stop ] Stop TCP transport")
	t.prepareDisconnect()

	utils.CloseVerbose(t.listener)
	t.pool.Reset()
}

func (t *tcpTransport) handleAcceptedConnection(conn net.Conn) {
	defer utils.CloseVerbose(conn)

	for {
		msg, err := t.serializer.DeserializePacket(conn)

		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				log.Warn("[ handleAcceptedConnection ] Connection closed by peer")
				return
			}

			log.Error("[ handleAcceptedConnection ] Failed to deserialize packet: ", err.Error())
		} else {
			ctx, logger := inslogger.WithTraceField(context.Background(), msg.TraceID)
			logger.Debug("[ handleAcceptedConnection ] Handling packet: ", msg.RequestID)

			go t.packetHandler.Handle(ctx, msg)
		}
	}
}

type tcpConnectionFactory struct{}

func (*tcpConnectionFactory) CreateConnection(ctx context.Context, address net.Addr) (net.Conn, error) {
	logger := inslogger.FromContext(ctx)

	tcpAddress, ok := address.(*net.TCPAddr)
	if !ok {
		return nil, errors.New("[ createConnection ] Failed to get tcp address")
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddress)
	if err != nil {
		logger.Errorf("[ createConnection ] Failed to open connection to %s: %s", address, err.Error())
		return nil, errors.Wrap(err, "[ createConnection ] Failed to open connection")
	}

	err = conn.SetKeepAlive(true)
	if err != nil {
		logger.Error("[ createConnection ] Failed to set keep alive")
	}

	err = conn.SetNoDelay(true)
	if err != nil {
		logger.Errorln("[ createConnection ] Failed to set connection no delay: ", err.Error())
	}

	return conn, nil
}
