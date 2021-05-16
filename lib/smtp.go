package lib

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime"
	"net/mail"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(STMP, `

declare namespace smtp {
    export function newMessage(): Message

    export function send(
        msg: Message,
        user: string,
        password: string,
        host: string,
        port: number,
        insecureSkipVerify?: boolean): void

    export interface Message {
        from: string
        fromName: string
        to: string[]
        cc: string[]
        bcc: string[]
        replyTo: string
        subject: string
        body: string
        html: boolean
        string(): string
        attach(fileName: string, data: byte[], inline: boolean): void
    }
}


`)
}

var STMP = []dune.NativeFunction{
	{
		Name:      "smtp.newMessage",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewObject(&SmtMessage{}), nil
		},
	},
	{
		Name:        "smtp.send",
		Arguments:   -1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.Object, dune.String, dune.String, dune.String, dune.Int, dune.Bool); err != nil {
				return dune.NullValue, err
			}

			msg, ok := args[0].ToObject().(*SmtMessage)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected a mail message, got %s", args[0].TypeName())
			}

			var err error
			var user, smtPasswd, host string
			var port int
			var skipVerify bool

			switch len(args) {
			case 5:
				user = args[1].String()
				smtPasswd = args[2].String()
				host = args[3].String()
				port = int(args[4].ToInt())

			case 6:
				user = args[1].String()
				smtPasswd = args[2].String()
				host = args[3].String()
				port = int(args[4].ToInt())
				skipVerify = args[5].ToBool()
			default:
				return dune.NullValue, fmt.Errorf("expected 4 or 5 params, got %d", len(args))
			}

			err = msg.Send(user, smtPasswd, host, port, skipVerify)
			return dune.NullValue, err
		},
	},
}

// Message represents a smtp message.
type SmtMessage struct {
	From        mail.Address
	To          dune.Value
	Cc          dune.Value
	Bcc         dune.Value
	ReplyTo     string
	Subject     string
	Body        string
	Html        bool
	Attachments map[string]*attachment
}

func (SmtMessage) Type() string {
	return "smtp.Message"
}

func (m *SmtMessage) GetField(key string, vm *dune.VM) (dune.Value, error) {
	switch key {
	case "from":
		return dune.NewString(m.From.Address), nil
	case "fromName":
		return dune.NewString(m.From.Name), nil
	case "to":
		if m.To.Type == dune.Null {
			m.To = dune.NewArray(0)
		}
		return m.To, nil
	case "cc":
		if m.Cc.Type == dune.Null {
			m.Cc = dune.NewArray(0)
		}
		return m.Cc, nil
	case "bcc":
		if m.Bcc.Type == dune.Null {
			m.Bcc = dune.NewArray(0)
		}
		return m.Bcc, nil
	case "replyTo":
		return dune.NewString(m.ReplyTo), nil
	case "subject":
		return dune.NewString(m.Subject), nil
	case "body":
		return dune.NewString(m.Body), nil
	case "html":
		return dune.NewBool(m.Html), nil
	}
	return dune.UndefinedValue, nil
}

func (m *SmtMessage) SetField(key string, v dune.Value, vm *dune.VM) error {
	switch key {
	case "from":
		if v.Type != dune.String {
			return ErrInvalidType
		}
		m.From.Address = v.String()
		return nil
	case "fromName":
		if v.Type != dune.String {
			return ErrInvalidType
		}

		m.From.Name = v.String()
		return nil
	case "to":
		if v.Type != dune.Array {
			return ErrInvalidType
		}
		m.To = v
		return nil
	case "cc":
		if v.Type != dune.Array {
			return ErrInvalidType
		}
		m.Cc = v
		return nil
	case "bcc":
		if v.Type != dune.Array {
			return ErrInvalidType
		}
		m.Bcc = v
		return nil
	case "replyTo":
		if v.Type != dune.String {
			return ErrInvalidType
		}
		m.ReplyTo = v.String()
		return nil
	case "subject":
		if v.Type != dune.String {
			return ErrInvalidType
		}
		m.Subject = v.String()
		return nil
	case "body":
		if v.Type != dune.String {
			return ErrInvalidType
		}
		m.Body = v.String()
		return nil
	case "html":
		if v.Type != dune.Bool {
			return ErrInvalidType
		}
		m.Html = v.ToBool()
		return nil
	}

	return ErrReadOnlyOrUndefined
}

func (m *SmtMessage) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "attach":
		return m.attachData
	case "string":
		return m.string
	}
	return nil
}

func (m *SmtMessage) string(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgRange(args, 0, 0); err != nil {
		return dune.NullValue, err
	}
	return dune.NewString(string(m.Bytes())), nil
}

func (m *SmtMessage) attachData(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgRange(args, 2, 3); err != nil {
		return dune.NullValue, err
	}

	var fileName string
	var data []byte
	var inline bool

	a := args[0]
	switch a.Type {
	case dune.String:
		fileName = a.String()
	default:
		return dune.NullValue, fmt.Errorf("invalid argument 1 type: %v", a)
	}

	b := args[1]
	switch b.Type {
	case dune.Bytes, dune.String:
		data = b.ToBytes()
	default:
		return dune.NullValue, fmt.Errorf("invalid argument 2 type: %v", b)
	}

	if len(args) == 3 {
		c := args[2]
		switch c.Type {
		case dune.Bool:
			inline = c.ToBool()
		default:
			return dune.NullValue, fmt.Errorf("invalid argument 2 type: %v", c)
		}
	}

	err := m.AttachBuffer(fileName, data, inline)
	return dune.NullValue, err
}

func (m *SmtMessage) ContentType() string {
	if m.Html {
		return "text/html"
	}
	return "text/plain"
}

func (m *SmtMessage) Send(user, password, host string, port int, insecureSkipVerify bool) error {
	auth := LoginAuth(user, password)
	address := host + ":" + strconv.Itoa(port)

	return SendMail(address, auth, m.From.Address, m.AllRecipients(), m.Bytes(), insecureSkipVerify)
}

// AttachBuffer attaches a binary attachment.
func (m *SmtMessage) AttachBuffer(filename string, buf []byte, inline bool) error {
	if m.Attachments == nil {
		m.Attachments = make(map[string]*attachment)
	}

	m.Attachments[filename] = &attachment{
		Filename: filename,
		Data:     buf,
		Inline:   inline,
	}
	return nil
}

// Attachment represents an email attachment.
type attachment struct {
	Filename string
	Data     []byte
	Inline   bool
}

func (m *SmtMessage) ToList() []string {
	if m.To.Type == dune.Null {
		return []string{}
	}

	a := m.To.ToArray()
	dirs := make([]string, len(a))
	for i, v := range a {
		dirs[i] = v.String()
	}
	return dirs
}

func (m *SmtMessage) CcList() []string {
	if m.Cc.Type == dune.Null {
		return []string{}
	}

	a := m.Cc.ToArray()
	dirs := make([]string, len(a))
	for i, v := range a {
		dirs[i] = v.String()
	}
	return dirs
}

func (m *SmtMessage) BccList() []string {
	if m.Bcc.Type == dune.Null {
		return []string{}
	}

	a := m.Bcc.ToArray()
	dirs := make([]string, len(a))
	for i, v := range a {
		dirs[i] = v.String()
	}
	return dirs
}

// Tolist returns all the recipients of the email
func (m *SmtMessage) AllRecipients() []string {
	dirs := m.ToList()
	dirs = append(dirs, m.CcList()...)
	dirs = append(dirs, m.BccList()...)
	return dirs
}

// Bytes returns the mail data
func (m *SmtMessage) Bytes() []byte {
	buf := bytes.NewBuffer(nil)

	buf.WriteString("From: " + m.From.String() + "\r\n")

	buf.WriteString("Date: " + time.Now().UTC().Format(time.RFC1123Z) + "\r\n")

	buf.WriteString("To: " + strings.Join(m.ToList(), ",") + "\r\n")

	cc := m.CcList()
	if len(cc) > 0 {
		buf.WriteString("Cc: " + strings.Join(cc, ",") + "\r\n")
	}

	bcc := m.BccList()
	if len(bcc) > 0 {
		buf.WriteString("Bcc: " + strings.Join(bcc, ",") + "\r\n")
	}

	//fix  Encode
	var coder = base64.StdEncoding
	var subject = "=?UTF-8?B?" + coder.EncodeToString([]byte(m.Subject)) + "?="
	buf.WriteString("Subject: " + subject + "\r\n")

	if len(m.ReplyTo) > 0 {
		buf.WriteString("Reply-To: " + m.ReplyTo + "\r\n")
	}

	buf.WriteString("MIME-Version: 1.0\r\n")

	boundary := "f46d043c813270fc6b04c2d223da"

	if len(m.Attachments) > 0 {
		buf.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\r\n")
		buf.WriteString("\r\n--" + boundary + "\r\n")
	}

	buf.WriteString(fmt.Sprintf("Content-Type: %s; charset=utf-8\r\n\r\n", m.ContentType()))
	buf.WriteString(m.Body)
	buf.WriteString("\r\n")

	if len(m.Attachments) > 0 {
		for _, attachment := range m.Attachments {
			buf.WriteString("\r\n\r\n--" + boundary + "\r\n")

			if attachment.Inline {
				buf.WriteString("Content-Type: message/rfc822\r\n")
				buf.WriteString("Content-Disposition: inline; filename=\"" + attachment.Filename + "\"\r\n\r\n")

				buf.Write(attachment.Data)
			} else {
				ext := filepath.Ext(attachment.Filename)
				mimetype := mime.TypeByExtension(ext)
				if mimetype != "" {
					mime := fmt.Sprintf("Content-Type: %s\r\n", mimetype)
					buf.WriteString(mime)
				} else {
					buf.WriteString("Content-Type: application/octet-stream\r\n")
				}
				buf.WriteString("Content-Transfer-Encoding: base64\r\n")

				buf.WriteString("Content-Disposition: attachment; filename=\"=?UTF-8?B?")
				buf.WriteString(coder.EncodeToString([]byte(attachment.Filename)))
				buf.WriteString("?=\"\r\n\r\n")

				b := make([]byte, base64.StdEncoding.EncodedLen(len(attachment.Data)))
				base64.StdEncoding.Encode(b, attachment.Data)

				// write base64 content in lines of up to 76 chars
				for i, l := 0, len(b); i < l; i++ {
					buf.WriteByte(b[i])
					if (i+1)%76 == 0 {
						buf.WriteString("\r\n")
					}
				}
			}

			buf.WriteString("\r\n--" + boundary)
		}

		buf.WriteString("--")
	}

	return buf.Bytes()
}

type loginAuth struct {
	username, password string
}

// loginAuth returns an Auth that implements the LOGIN authentication
// mechanism as defined in RFC 4616.
func LoginAuth(username, password string) Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *ServerInfo) (string, []byte, error) {
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	command := string(fromServer)
	command = strings.TrimSpace(command)
	command = strings.TrimSuffix(command, ":")
	command = strings.ToLower(command)

	if more {
		if command == "username" {
			return []byte(a.username), nil
		} else if command == "password" {
			return []byte(a.password), nil
		} else {
			// We've already sent everything.
			return nil, fmt.Errorf("unexpected server challenge: %s", command)
		}
	}
	return nil, nil
}
