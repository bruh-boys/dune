package lib

import (
	"encoding/json"
	"fmt"

	"github.com/dunelang/dune"

	"github.com/gorilla/websocket"
)

func init() {
	dune.RegisterLib(WebSocket, `

declare namespace websocket {
    export function upgrade(r: http.Request): WebsocketConnection

    export interface WebsocketConnection {
        guid: string
        write(v: any): number | void
        writeJSON(v: any): void
        writeText(text: string | byte[]): void
        readMessage(): WebSocketMessage
        close(): void
    }

    export interface WebSocketMessage {
        type: WebsocketType
        message: string
    }

    export enum WebsocketType {
        text = 1,
        binary = 2,
        close = 8,
        ping = 9,
        pong = 10
    }
}

`)
}

var upgrader = websocket.Upgrader{} // websockets: use default options

var WebSocket = []dune.NativeFunction{
	{
		Name:      "websocket.upgrade",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if !vm.HasPermission("networking") {
				return dune.NullValue, ErrUnauthorized
			}
			if err := ValidateArgs(args, dune.Object); err != nil {
				return dune.NullValue, err
			}
			r, ok := args[0].ToObject().(*request)
			if !ok {
				return dune.NullValue, fmt.Errorf("invalid Request, got %s", args[1].TypeName())
			}

			c, err := upgrader.Upgrade(r.writer, r.request, nil)
			if err != nil {
				return dune.NullValue, err
			}

			// Maximum message size allowed from peer.
			c.SetReadLimit(8192)

			return dune.NewObject(newWebsocketConn(c, vm)), nil
		},
	},
}

func newWebsocketConn(con *websocket.Conn, vm *dune.VM) *websocketConn {
	f := &websocketConn{con: con}
	vm.SetGlobalFinalizer(f)
	return f
}

type websocketConn struct {
	guid string
	con  *websocket.Conn
}

func (c *websocketConn) Type() string {
	return "http.WebsocketConnection"
}

func (c *websocketConn) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "guid":
		return dune.NewString(c.guid), nil
	}
	return dune.UndefinedValue, nil
}

func (c *websocketConn) SetProperty(name string, v dune.Value, vm *dune.VM) error {
	switch name {
	case "guid":
		if v.Type != dune.String {
			return ErrInvalidType
		}
		c.guid = v.ToString()
		return nil
	}
	return ErrReadOnlyOrUndefined
}

func (c *websocketConn) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "write":
		return c.write
	case "writeText":
		return c.writeTextMessage
	case "writeJSON":
		return c.writeJSON
	case "readMessage":
		return c.readMessage
	case "close":
		return c.close
	}
	return nil
}

func (c *websocketConn) Close() error {
	ws := c.con

	// // Time allowed to write a message to the peer.
	// writeWait := 5 * time.Second
	// ws.SetWriteDeadline(time.Now().Add(writeWait))
	// ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	// time.Sleep(writeWait)

	return ws.Close()
}

func (c *websocketConn) Write(b []byte) (n int, err error) {
	err = c.con.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *websocketConn) write(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expecting 1 parameter, got %d", len(args))
	}

	v := args[0]
	var b []byte

	switch v.Type {
	case dune.String, dune.Bytes:
		b = v.ToBytes()
	default:
		return dune.NullValue, ErrInvalidType
	}

	n, err := c.Write(b)

	return dune.NewInt(n), err
}

func (c *websocketConn) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expecting no parameters, got %d", len(args))
	}

	err := c.Close()
	return dune.NullValue, err
}

func (c *websocketConn) readMessage(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expecting no parameters, got %d", len(args))
	}

	mType, msg, err := c.con.ReadMessage()
	if err != nil {
		return dune.NullValue, err
	}

	result := make(map[dune.Value]dune.Value)
	result[dune.NewString("message")] = dune.NewBytes(msg)
	result[dune.NewString("type")] = dune.NewInt(mType)
	return dune.NewMapValues(result), nil
}

func (c *websocketConn) writeJSON(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expecting 1 parameter, got %d", len(args))
	}

	v := args[0].Export(0)

	b, err := json.Marshal(v)

	if err != nil {
		return dune.NullValue, err
	}

	if err := vm.AddAllocations(len(b)); err != nil {
		return dune.NullValue, err
	}

	err = c.con.WriteMessage(websocket.TextMessage, b)
	return dune.NullValue, err
}

func (c *websocketConn) writeTextMessage(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expecting 1 parameter, got %d", len(args))
	}

	var b []byte
	a := args[0]

	switch a.Type {
	case dune.String, dune.Bytes:
		b = a.ToBytes()
	default:
		return dune.NullValue, fmt.Errorf("invalid parameter type: %s", a.TypeName())
	}

	err := c.con.WriteMessage(websocket.TextMessage, b)
	return dune.NullValue, err
}
