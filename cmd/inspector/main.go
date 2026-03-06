package main

import (
	"fmt"
	"time"

	"github.com/apex/monitor/pkg/vault"
)

func main() {
	key := []byte("this-is-a-32-byte-secret-key-!!!")
	v, err := vault.New("apex.db", key)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer v.Close()

	reports, _ := v.FetchAll()

	fmt.Println("=== Apex Vault Inspector ===")
	fmt.Printf("Total Records: %d\n\n", len(reports))

	for _, r := range reports {
		t := time.Unix(r.Timestamp, 0).Format("2006-01-02 15:04:05")
		fmt.Printf("[%s] %s\n", t, r.Message)
		fmt.Printf("   OS: %s | Mem: %d KB\n", r.Context.Os, r.Context.TotalMemory/1024)
		fmt.Println("-------------------------------------------")
	}
}
