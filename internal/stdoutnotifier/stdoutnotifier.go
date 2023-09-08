// Package stdoutnotifier implements Notifier interface with
// standard output (stdout) printer.
package stdoutnotifier

import "fmt"

type StdOut struct{}

// Notify prints message to stdout.
func (t *StdOut) Notify(_, message string) error {
	_, err := fmt.Println(message)

	return err
}
