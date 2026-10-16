package practice

type Notifier interface {
	Send(message string) error
}

type BaseNotifier struct {
	Channel string
}

func (b BaseNotifier) Label(message string) string {
	return "[" + b.Channel + "] " + message
}

type EmailNotifier struct {
	BaseNotifier
	Address  string
	Messages []string
}

type PushNotifier struct {
	BaseNotifier
	DeviceID string
	Messages []string
}

func (n *EmailNotifier) Send(message string) error {
	n.Messages = append(n.Messages, n.Label(message))
	return nil
}

func (n *PushNotifier) Send(message string) error {
	n.Messages = append(n.Messages, n.Label(message))
	return nil
}

func Dispatch(all []Notifier, message string) error {
	for _, n := range all {
		err := n.Send(message)
		if err != nil {
			return err
		}
	}
	return nil
}

func EmailAddress(n Notifier) (string, bool) {
	if email, ok := n.(*EmailNotifier); ok {
		return email.Address, true
	}
	return "", false
}

func Describe(n Notifier) string {
	switch n.(type) {
	case *EmailNotifier:
		return "email"
	case *PushNotifier:
		return "push"
	}
	return "unknown"
}
