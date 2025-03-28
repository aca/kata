package main

import (
	"context"
	"log"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
)

var dbusConn *dbus.Conn

func main() {
	ctx := context.Background()
	var err error
	dbusConn, err = dbus.NewUserConnectionContext(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to systemd: %v", err)
	}
	defer dbusConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	st, err := dbusConn.ListUnitsByNamesContext(ctx, []string{"xd-*"})
	if err != nil {
		panic(err)
	}

	log.Println(len(st))

	for _, u := range st {
		log.Printf("Unit: %s, ActiveState: %s, SubState: %s", u.Name, u.ActiveState, u.SubState)
	}

	// properties, err := dbusConn.GetUnitPropertiesContext(ctx, serviceName)
	// if err != nil {
	// 	return err.Error()
	// }
}
