package main

import (
	"errors"
	"fmt"
	"lesson5/student/practice"
)

func printMessages(messages []string) {
	for _, message := range messages {
		fmt.Printf("%s ", message)
		fmt.Println()
	}
}

func main() {
	var email = &practice.EmailNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "email"},
		Address:      "iluxasxemotexnika@dvenashka.nasledie",
	}

	var push = &practice.PushNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "push"},
		DeviceID:     "1267",
	}

	var notifiers = []practice.Notifier{email, push}

	fmt.Println("всё хорошо:")

	var err = practice.Dispatch(notifiers, "Заказ готов")

	if err == nil {
		printMessages(email.Messages)
		printMessages(push.Messages)
	}

	fmt.Println()
	fmt.Println("пустое сообщение:")

	err = practice.Dispatch(notifiers, "")
	if errors.Is(err, practice.ErrEmptyMessage) {
		fmt.Println(err)
		printMessages(email.Messages)
		printMessages(push.Messages)
	}

	fmt.Println()
	fmt.Println("push без device id:")

	notifiers[1] = &practice.PushNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "push"},
	}

	err = practice.Dispatch(notifiers, "yalubludvenashku")
	var sendError *practice.SendError
	if errors.As(err, &sendError) {
		fmt.Println(err)
	}

	var brokenAndValidPushSlice = []practice.Notifier{&practice.BrokenNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "broken"},
	}, push}

	fmt.Println()
	fmt.Println("brokenNotifier и исправный push:")

	var errs = practice.SafeDispatch(brokenAndValidPushSlice, "schemotechique")
	if errs != nil {
		for _, e := range errs {
			fmt.Println(e)
		}
	}
	printMessages(push.Messages)
}
