// Package buildinfo хранит метаданные сборки и выводит их при старте.
package buildinfo

import "fmt"

var (
	BuildVersion = "N/A"
	BuildDate    = "N/A"
	BuildCommit  = "N/A"
)

// Print выводит версию, дату и коммит сборки в stdout.
func Print() {
	fmt.Println("Build version:", BuildVersion)
	fmt.Println("Build date:", BuildDate)
	fmt.Println("Build commit:", BuildCommit)
}
